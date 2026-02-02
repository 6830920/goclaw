// Package main provides the OpenClaw-Go server with HTTP API
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"openclaw-go/internal/chat"
	"openclaw-go/internal/config"
	"openclaw-go/internal/memory"
	"openclaw-go/internal/vector"
)

// Version info
const Version = "0.1.0"

// API Response types
type APIResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func main() {
	fmt.Printf("OpenClaw-Go Server v%s\n", Version)
	fmt.Println("======================\n")

	// Load configuration
	cfg := loadConfig()

	// Initialize components
	embedder := initEmbedder(cfg)
	
	memoryStore := memory.NewMemoryStore(memory.MemoryConfig{
		ShortTermMax:   50,
		WorkingMax:     10,
		SimilarityCut:  0.7,
	})
	
	chatManager := chat.NewChatManager(100)
	
	var vectorStore vector.VectorStore = vector.NewInMemoryStore(embedder)

	// Use port 18888 to avoid conflicts with original OpenClaw
	port := "18888"
	fmt.Printf("Starting OpenClaw-Go server on port %s\n", port)
	
	// API Routes
	http.HandleFunc("/api/chat", handleChat(embedder, memoryStore, chatManager, vectorStore, cfg))
	http.HandleFunc("/api/memory/search", handleMemorySearch(embedder, memoryStore))
	http.HandleFunc("/api/memory/stats", handleMemoryStats(memoryStore))
	http.HandleFunc("/api/sessions", handleSessions(chatManager))
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(APIResponse{Status: "ok", Message: "OpenClaw-Go is running"})
	})
	
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(APIResponse{
			Status:  "ok",
			Message: "OpenClaw-Go API Server v" + Version,
			Data: map[string]interface{}{
				"endpoints": []string{
					"/api/chat",
					"/api/memory/search",
					"/api/memory/stats",
					"/api/sessions",
					"/health",
				},
				"port": port,
			},
		})
	})

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func loadConfig() *config.Config {
	cfg := config.NewDefaultConfig()
	
	// Override default port to avoid conflicts with original OpenClaw
	cfg.Gateway.Port = 18888
	
	// Try to load from file
	if _, err := os.Stat("config.json"); err == nil {
		loadedCfg, err := config.LoadConfig("config.json")
		if err == nil {
			cfg = loadedCfg
			fmt.Println("Loaded configuration from config.json")
		}
	}
	
	return cfg
}

func initEmbedder(cfg *config.Config) vector.Embedder {
	// Check if Ollama is available
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	req, _ := http.NewRequestWithContext(ctx, "GET", "http://localhost:11434/api/version", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Note: Ollama not detected, embedding features will be limited")
		fmt.Println("Run 'ollama serve' to enable local embeddings")
		return nil
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusOK {
		fmt.Println("Connected to Ollama for embeddings")
		return vector.NewOllamaEmbedder("", "")
	}
	
	return nil
}

func handleChat(embedder vector.Embedder, memStore *memory.MemoryStore, chatMgr *chat.ChatManager, vectorStore vector.VectorStore, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Message    string `json:"message"`
			SessionID  string `json:"sessionId,omitempty"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		sessionID := req.SessionID
		if sessionID == "" {
			sessionID = fmt.Sprintf("api_session_%d", time.Now().Unix())
			chatMgr.CreateSession(sessionID, cfg.Agent.Model)
		}

		// Add user message
		chatMgr.AddMessage(sessionID, "user", req.Message)

		// Get context from memory
		var contextText string
		if embedder != nil {
			ctx := context.Background()
			embedding, _ := embedder.Embed(ctx, req.Message)
			contextText, _ = memStore.GetContext(ctx, req.Message, embedding, 500)
		}

		// Generate response
		response := generateResponse(req.Message, contextText, chatMgr, sessionID)

		// Add assistant message
		chatMgr.AddMessage(sessionID, "assistant", response)

		// Add to short-term memory
		memStore.AddShortTerm(req.Message, map[string]interface{}{
			"session": sessionID,
			"source":  "api",
		})

		// Get updated messages
		messages, _ := chatMgr.GetMessages(sessionID)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(APIResponse{
			Status: "ok",
			Data: map[string]interface{}{
				"sessionId": sessionID,
				"response":  response,
				"messages":  messages,
			},
		})
	}
}

func handleMemorySearch(embedder vector.Embedder, memStore *memory.MemoryStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Query string `json:"query"`
			Limit int    `json:"limit,omitempty"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if embedder == nil {
			http.Error(w, "No embedder available", http.StatusServiceUnavailable)
			return
		}

		ctx := context.Background()
		embedding, err := embedder.Embed(ctx, req.Query)
		if err != nil {
			http.Error(w, "Failed to generate embedding", http.StatusInternalServerError)
			return
		}

		limit := req.Limit
		if limit <= 0 {
			limit = 5
		}

		results, err := memStore.Search(ctx, req.Query, embedding, limit)
		if err != nil {
			http.Error(w, "Search failed", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(APIResponse{
			Status: "ok",
			Data:   results,
		})
	}
}

func handleMemoryStats(memStore *memory.MemoryStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats := memStore.Stats()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(APIResponse{
			Status: "ok",
			Data:   stats,
		})
	}
}

func handleSessions(chatMgr *chat.ChatManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		sessions := chatMgr.ListSessions()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(APIResponse{
			Status: "ok",
			Data: map[string]interface{}{
				"sessions":     sessions,
				"sessionCount": len(sessions),
			},
		})
	}
}

func generateResponse(input, contextText string, chatMgr *chat.ChatManager, sessionID string) string {
	// Get conversation history
	messages, _ := chatMgr.GetMessages(sessionID)
	
	// Build prompt
	prompt := buildPrompt(input, contextText, messages)
	
	// Call Claude Code CLI if available
	response := callClaudeCode(prompt)
	
	return response
}

func buildPrompt(input, contextText string, messages []chat.Message) string {
	var sb strings.Builder
	
	sb.WriteString("You are OpenClaw-Go, a personal AI assistant.\n\n")
	
	if contextText != "" {
		sb.WriteString("Context from memory:\n")
		sb.WriteString(contextText)
		sb.WriteString("\n\n")
	}
	
	sb.WriteString("Conversation:\n")
	for _, msg := range messages {
		if msg.Role == "system" {
			continue
		}
		sb.WriteString(fmt.Sprintf("%s: %s\n", msg.Role, msg.Content))
	}
	
	sb.WriteString("\nProvide a helpful, concise response.\n")
	
	return sb.String()
}

func callClaudeCode(prompt string) string {
	// Try to use Claude Code CLI
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, "claude-code", "--print", "--no-stream")
	cmd.Stdin = strings.NewReader(prompt)
	
	output, err := cmd.Output()
	if err != nil {
		// Fallback to simple response
		return generateSimpleResponse(prompt)
	}
	
	return strings.TrimSpace(string(output))
}

func generateSimpleResponse(prompt string) string {
	// Simple fallback response
	promptLower := strings.ToLower(prompt)
	
	if strings.Contains(promptLower, "hello") || strings.Contains(promptLower, "hi") {
		return "Hello! I'm OpenClaw-Go. How can I help you today?"
	}
	
	if strings.Contains(promptLower, "time") {
		return fmt.Sprintf("The current time is %s", time.Now().Format("3:04 PM"))
	}
	
	return "I understand you're saying: \"" + prompt + "\"\n\nI'm OpenClaw-Go API server running on port 18888."
}