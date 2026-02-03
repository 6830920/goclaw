package main

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"goclaw/internal/chat"
	"goclaw/internal/config"
	"goclaw/internal/memory"
	"goclaw/internal/vector"
)

// Version info
const Version = "0.1.0"

func main() {
	fmt.Printf("Goclaw v%s\n", Version)
	fmt.Println("==================")

	// Load configuration
	cfg := loadConfig()

	// Initialize components
	embedder := initEmbedder(cfg)

	memoryStore := memory.NewMemoryStore(memory.MemoryConfig{
		ShortTermMax:  50,
		WorkingMax:    10,
		SimilarityCut: 0.7,
	})

	chatManager := chat.NewChatManager(100)

	var vectorStore vector.VectorStore = vector.NewInMemoryStore(embedder)

	// Start CLI
	runCLI(embedder, memoryStore, chatManager, vectorStore, cfg)
}

func loadConfig() *config.Config {
	cfg := config.NewDefaultConfig()

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

func runCLI(embedder vector.Embedder, memStore *memory.MemoryStore, chatMgr *chat.ChatManager, vectorStore vector.VectorStore, cfg *config.Config) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\nGoclaw CLI")
	fmt.Println("===============")
	fmt.Println("Commands:")
	fmt.Println("  /new           - Start new session")
	fmt.Println("  /quit          - Exit")
	fmt.Println("  /remember <x>  - Save to memory")
	fmt.Println("  /recall <x>    - Search memory")
	fmt.Println("  /stats         - Show memory stats")
	fmt.Println("  /help          - Show this help")
	fmt.Println("")
	fmt.Println("Just type to chat!")

	sessionID := "default"

	// Create default session
	chatMgr.CreateSession(sessionID, cfg.Agent.Model)

	for {
		fmt.Print("You: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Handle commands
		if strings.HasPrefix(input, "/") {
			if err := handleCommand(input, embedder, memStore, chatMgr, vectorStore, &sessionID); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
			continue
		}

		// Regular message
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

		fmt.Printf("Assistant: %s\n", response)

		chatMgr.AddMessage(sessionID, "assistant", response)

		// Add to short-term memory
		memStore.AddShortTerm(input, map[string]interface{}{
			"session": sessionID,
		})
	}
}

func handleCommand(cmd string, embedder vector.Embedder, memStore *memory.MemoryStore, chatMgr *chat.ChatManager, vectorStore vector.VectorStore, sessionID *string) error {
	parts := strings.SplitN(cmd, " ", 2)
	command := strings.ToLower(parts[0])

	switch command {
	case "/new":
		*sessionID = fmt.Sprintf("session_%d", time.Now().Unix())
		chatMgr.CreateSession(*sessionID, "")
		fmt.Printf("Started new session: %s\n", *sessionID)

	case "/quit":
		fmt.Println("Goodbye!")
		os.Exit(0)

	case "/remember":
		if len(parts) < 2 {
			return fmt.Errorf("usage: /remember <text>")
		}
		content := parts[1]
		memStore.AddShortTerm(content, map[string]interface{}{
			"type": "manual",
		})
		fmt.Println("Remembered!")

	case "/recall":
		if len(parts) < 2 {
			return fmt.Errorf("usage: /recall <query>")
		}
		query := parts[1]

		if embedder == nil {
			return fmt.Errorf("no embedder available, cannot search")
		}

		ctx := context.Background()
		embedding, err := embedder.Embed(ctx, query)
		if err != nil {
			return err
		}

		results, err := memStore.Search(ctx, query, embedding, 5)
		if err != nil {
			return err
		}

		fmt.Println("\nMemory Search Results:")
		for _, r := range results {
			fmt.Printf("  [%.2f] %s\n", r.Score, r.Entry.Content)
		}
		if len(results) == 0 {
			fmt.Println("  No memories found")
		}

	case "/stats":
		stats := memStore.Stats()
		fmt.Printf("\nMemory Stats:")
		fmt.Printf("  Short-term: %d\n", stats.ShortTermCount)
		fmt.Printf("  Long-term:  %d\n", stats.LongTermCount)
		fmt.Printf("  Working:    %d\n", stats.WorkingCount)

	case "/help":
		fmt.Println("\nCommands:")
		fmt.Println("  /new           - Start new session")
		fmt.Println("  /quit          - Exit")
		fmt.Println("  /remember <x>  - Save to memory")
		fmt.Println("  /recall <x>    - Search memory")
		fmt.Println("  /stats         - Show memory stats")
		fmt.Println("  /help          - Show this help")

	default:
		return fmt.Errorf("unknown command: %s", command)
	}

	return nil
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

	sb.WriteString("You are Goclaw, a personal AI assistant.\n\n")

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
		return "Hello! I'm Goclaw. How can I help you today?"
	}

	if strings.Contains(promptLower, "time") {
		return fmt.Sprintf("The current time is %s", time.Now().Format("3:04 PM"))
	}

	return "I understand you're saying: \"" + prompt + "\"\n\nAs a simple assistant, I can remember things, search my memory, and have basic conversations. Try /help for available commands!"
}
