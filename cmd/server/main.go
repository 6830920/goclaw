// Package main provides the OpenClaw-Go server with WebSocket gateway
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"openclaw-go/internal/chat"
	"openclaw-go/internal/config"
	"openclaw-go/internal/memory"
	"openclaw-go/internal/vector"
)

// Version info
const Version = "0.1.0"

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin during development
		// In production, restrict this to specific origins
		return true
	},
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

	// Use a different port to avoid conflicts with original OpenClaw
	port := "18888" // Changed from default 18789 to avoid conflicts
	fmt.Printf("Starting OpenClaw-Go server on port %s\n", port)
	
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(w, r, embedder, memoryStore, chatManager, vectorStore, cfg)
	})
	
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OpenClaw-Go Server v%s\n", Version)
		fmt.Fprintf(w, "WebSocket endpoint available at /ws\n")
	})

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func loadConfig() *config.Config {
	cfg := config.NewDefaultConfig()
	
	// Override default port to avoid conflicts with original OpenClaw
	cfg.Gateway.Port = 18790
	
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

func handleWebSocket(w http.ResponseWriter, r *http.Request, embedder vector.Embedder, memStore *memory.MemoryStore, chatMgr *chat.ChatManager, vectorStore vector.VectorStore, cfg *config.Config) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()
	
	sessionID := fmt.Sprintf("ws_session_%d", time.Now().Unix())
	chatMgr.CreateSession(sessionID, cfg.Agent.Model)
	
	fmt.Printf("New WebSocket connection established: %s\n", sessionID)
	
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}
		
		log.Printf("Received message from %s: %s", sessionID, string(message))
		
		// Parse the message - expect JSON format with action and content
		input := string(message)
		
		// Add to chat session
		chatMgr.AddMessage(sessionID, "user", input)
		
		// Get context from memory
		var contextText string
		if embedder != nil {
			ctx := context.Background()
			embedding, _ := embedder.Embed(ctx, input)
			contextText, _ = memStore.GetContext(ctx, input, embedding, 500)
		}
		
		// Generate response
		response := generateResponse(input, contextText, chatMgr, sessionID)
		
		// Add to chat session
		chatMgr.AddMessage(sessionID, "assistant", response)
		
		// Add to short-term memory
		memStore.AddShortTerm(input, map[string]interface{}{
			"session": sessionID,
			"source":  "websocket",
		})
		
		// Send response back through WebSocket
		err = conn.WriteMessage(messageType, []byte(response))
		if err != nil {
			log.Printf("WebSocket write error: %v", err)
			break
		}
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
		return "Hello! I'm OpenClaw-Go Server. How can I help you today?"
	}
	
	if strings.Contains(promptLower, "time") {
		return fmt.Sprintf("The current time is %s", time.Now().Format("3:04 PM"))
	}
	
	return "I understand you're saying: \"" + prompt + "\"\n\nI'm OpenClaw-Go Server with WebSocket support running on port 18790."
}