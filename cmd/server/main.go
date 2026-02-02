// Package main provides the OpenClaw-Go server with HTTP API and Web UI
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"openclaw-go/internal/chat"
	"openclaw-go/internal/config"
	"openclaw-go/internal/memory"
	"openclaw-go/internal/vector"
	"openclaw-go/pkg/ai"
	"openclaw-go/pkg/utils"
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
	var embedder vector.Embedder
	// Check if any AI provider is configured
	hasAIProvider := cfg.Zhipu.ApiKey != "" || 
		(cfg.Models["providers"] != nil && len(cfg.Models["providers"].(map[string]interface{})) > 0)
	
	if hasAIProvider {
		// AI provider is configured, skip Ollama embedder
		fmt.Println("AI provider configured - skipping Ollama embedder initialization")
		embedder = nil
	} else {
		// Only try to initialize Ollama embedder if no other AI provider is configured
		embedder = initEmbedder(cfg)
	}
	
	memoryStore := memory.NewMemoryStore(memory.MemoryConfig{
		ShortTermMax:   50,
		WorkingMax:     10,
		SimilarityCut:  0.7,
	})
	
	chatManager := chat.NewChatManager(100)
	
	var vectorStore vector.VectorStore
	if embedder != nil {
		vectorStore = vector.NewInMemoryStore(embedder)
		fmt.Println("Vector store initialized with embedder")
	} else {
		// Create a minimal vector store without embedding capabilities
		vectorStore = vector.NewInMemoryStore(nil)
		fmt.Println("Vector store initialized without embedder (limited functionality)")
	}

	// Initialize AI client
	initializeAI(cfg)

	// Use port 18890 to avoid conflicts
	port := "18890"
	fmt.Printf("Starting OpenClaw-Go server on port %s\n", port)
	
	// Create static files directory
	os.MkdirAll("static", 0755)
	
	// Write web UI files
	writeStaticFiles()
	
	// API Routes
	http.HandleFunc("/api/chat", handleChat(embedder, memoryStore, chatManager, vectorStore, cfg))
	http.HandleFunc("/api/memory/search", handleMemorySearch(embedder, memoryStore))
	http.HandleFunc("/api/memory/stats", handleMemoryStats(memoryStore))
	http.HandleFunc("/api/sessions", handleSessions(chatManager))
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(APIResponse{Status: "ok", Message: "OpenClaw-Go is running"})
	})
	
	// Static file handlers
	fs := http.FileServer(http.Dir("./static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/index.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Serve index.html for root path to support SPA
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "./static/index.html")
		} else {
			http.ServeFile(w, r, "./static/index.html")
		}
	})

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// writeStaticFiles creates the necessary static files for the web UI
func writeStaticFiles() {
	// Create index.html
	indexHTML := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>OpenClaw-Go</title>
    <link rel="manifest" href="/static/manifest.json">
    <link rel="icon" type="image/x-icon" href="data:image/svg+xml,<svg xmlns=%22http://www.w3.org/2000/svg%22 viewBox=%220 0 100 100%22><text y=%22.9em%22 font-size=%2290%22>🤖</text></svg>">
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background-color: #f5f7fb;
            color: #333;
            line-height: 1.6;
            height: 100vh;
            display: flex;
            flex-direction: column;
        }
        
        .header {
            background: linear-gradient(135deg, #6a11cb 0%, #2575fc 100%);
            color: white;
            padding: 1rem;
            text-align: center;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        
        .header h1 {
            font-size: 1.5rem;
            font-weight: 600;
        }
        
        .chat-container {
            flex: 1;
            display: flex;
            flex-direction: column;
            max-width: 800px;
            width: 100%;
            margin: 0 auto;
            padding: 1rem;
            overflow: hidden;
        }
        
        .messages {
            flex: 1;
            overflow-y: auto;
            padding: 1rem 0;
            display: flex;
            flex-direction: column;
            gap: 1rem;
        }
        
        .message {
            max-width: 80%;
            padding: 0.75rem 1rem;
            border-radius: 18px;
            position: relative;
            animation: fadeIn 0.3s ease-out;
        }
        
        @keyframes fadeIn {
            from { opacity: 0; transform: translateY(10px); }
            to { opacity: 1; transform: translateY(0); }
        }
        
        .user-message {
            align-self: flex-end;
            background-color: #007AFF;
            color: white;
            border-bottom-right-radius: 4px;
        }
        
        .assistant-message {
            align-self: flex-start;
            background-color: #f0f2f5;
            color: #333;
            border-bottom-left-radius: 4px;
        }
        
        .input-container {
            display: flex;
            padding: 1rem 0;
            gap: 0.5rem;
        }
        
        #message-input {
            flex: 1;
            padding: 0.75rem 1rem;
            border: 1px solid #ddd;
            border-radius: 24px;
            font-size: 1rem;
            outline: none;
            transition: border-color 0.3s;
        }
        
        #message-input:focus {
            border-color: #007AFF;
            box-shadow: 0 0 0 2px rgba(0, 122, 255, 0.2);
        }
        
        #send-button {
            background-color: #007AFF;
            color: white;
            border: none;
            border-radius: 50%;
            width: 48px;
            height: 48px;
            cursor: pointer;
            display: flex;
            align-items: center;
            justify-content: center;
            transition: background-color 0.3s;
        }
        
        #send-button:hover {
            background-color: #0056cc;
        }
        
        #send-button:disabled {
            background-color: #cccccc;
            cursor: not-allowed;
        }
        
        .typing-indicator {
            align-self: flex-start;
            background-color: #f0f2f5;
            color: #333;
            padding: 0.75rem 1rem;
            border-radius: 18px;
            font-style: italic;
            display: none;
        }
        
        .info-text {
            text-align: center;
            color: #666;
            font-size: 0.9rem;
            margin-top: 1rem;
        }
        
        @media (max-width: 768px) {
            .chat-container {
                padding: 0.5rem;
            }
            
            .message {
                max-width: 90%;
            }
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>🤖 OpenClaw-Go</h1>
    </div>
    
    <div class="chat-container">
        <div class="messages" id="messages"></div>
        <div class="typing-indicator" id="typing-indicator">Assistant is typing...</div>
        
        <div class="input-container">
            <input type="text" id="message-input" placeholder="Type your message..." autocomplete="off">
            <button id="send-button">➤</button>
        </div>
        
        <p class="info-text">Powered by OpenClaw-Go • Port 18888</p>
    </div>

    <script>
        const messagesContainer = document.getElementById('messages');
        const messageInput = document.getElementById('message-input');
        const sendButton = document.getElementById('send-button');
        const typingIndicator = document.getElementById('typing-indicator');
        
        let currentSessionId = 'web_' + new Date().getTime();
        
        // Add welcome message
        addMessage('assistant', 'Hello! I\'m OpenClaw-Go. How can I help you today?');
        
        // Focus input field
        messageInput.focus();
        
        // Send message on button click
        sendButton.addEventListener('click', sendMessage);
        
        // Send message on Enter key (but allow Shift+Enter for new line)
        messageInput.addEventListener('keydown', function(e) {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                sendMessage();
            }
        });
        
        async function sendMessage() {
            const message = messageInput.value.trim();
            if (!message) return;
            
            // Add user message to UI
            addMessage('user', message);
            messageInput.value = '';
            
            // Show typing indicator
            typingIndicator.style.display = 'block';
            scrollToBottom();
            
            try {
                // Send message to API
                const response = await fetch('/api/chat', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({
                        message: message,
                        sessionId: currentSessionId
                    })
                });
                
                const data = await response.json();
                
                if (data.status === 'ok') {
                    // Add assistant response to UI
                    addMessage('assistant', data.data.response);
                } else {
                    addMessage('assistant', 'Sorry, I encountered an error processing your request.');
                }
            } catch (error) {
                console.error('Error:', error);
                addMessage('assistant', 'Sorry, I\'m having trouble connecting to the server.');
            } finally {
                // Hide typing indicator
                typingIndicator.style.display = 'none';
            }
        }
        
        function addMessage(sender, text) {
            const messageDiv = document.createElement('div');
            messageDiv.classList.add('message');
            messageDiv.classList.add(sender + '-message');
            messageDiv.textContent = text;
            messagesContainer.appendChild(messageDiv);
            
            scrollToBottom();
        }
        
        function scrollToBottom() {
            messagesContainer.scrollTop = messagesContainer.scrollHeight;
        }
        
        // Service Worker registration for PWA functionality
        if ('serviceWorker' in navigator) {
            window.addEventListener('load', () => {
                navigator.serviceWorker.register('/static/sw.js')
                    .then(registration => {
                        console.log('SW registered: ', registration);
                    })
                    .catch(registrationError => {
                        console.log('SW registration failed: ', registrationError);
                    });
            });
        }
    </script>
</body>
</html>`

	staticDir := "static"
	os.MkdirAll(staticDir, 0755)
	
	// Write index.html
	err := os.WriteFile(staticDir+"/index.html", []byte(indexHTML), 0644)
	if err != nil {
		log.Printf("Error writing index.html: %v", err)
	}
	
	// Create manifest.json for PWA
	manifestJSON := `{
    "name": "OpenClaw-Go",
    "short_name": "OC-Go",
    "description": "Personal AI Assistant",
    "start_url": "/",
    "display": "standalone",
    "background_color": "#f5f7fb",
    "theme_color": "#6a11cb",
    "icons": [
        {
            "src": "data:image/svg+xml,<svg xmlns=%22http://www.w3.org/2000/svg%22 viewBox=%220 0 100 100%22><text y=%22.9em%22 font-size=%2290%22>🤖</text></svg>",
            "sizes": "192x192",
            "type": "image/svg+xml"
        }
    ]
}`
	
	err = os.WriteFile(staticDir+"/manifest.json", []byte(manifestJSON), 0644)
	if err != nil {
		log.Printf("Error writing manifest.json: %v", err)
	}
	
	// Create service worker for PWA
	swJS := `// Simple service worker for caching
const CACHE_NAME = 'openclaw-go-v1';
const urlsToCache = [
  '/',
  '/static/index.html',
];

self.addEventListener('install', event => {
  event.waitUntil(
    caches.open(CACHE_NAME)
      .then(cache => cache.addAll(urlsToCache))
  );
});

self.addEventListener('fetch', event => {
  event.respondWith(
    caches.match(event.request)
      .then(response => response || fetch(event.request))
  );
});`

	err = os.WriteFile(staticDir+"/sw.js", []byte(swJS), 0644)
	if err != nil {
		log.Printf("Error writing sw.js: %v", err)
	}
}

func loadConfig() *config.Config {
	cfg := config.NewDefaultConfig()
	
	// Override default port to avoid conflicts with original OpenClaw
	cfg.Gateway.Port = 18890
	
	// Try to load local config (config.json) first
	if _, err := os.Stat("config.json"); err == nil {
		localCfg, err := config.LoadConfig("config.json")
		if err == nil {
			fmt.Println("Loaded local configuration from config.json")
			// Use local config
			cfg = localCfg
		}
	}
	
	// Then try to load global config (~/.openclaw/openclaw.json), which takes precedence
	globalCfg, err := config.LoadGlobalConfig()
	if err != nil {
		fmt.Printf("No global config found: %v\n", err)
	} else {
		fmt.Println("Loaded global configuration from ~/.openclaw/openclaw.json")
		// Merge global config with local/default, with global taking precedence
		cfg = config.MergeConfigs(globalCfg, cfg)
	}
	
	return cfg
}

func initEmbedder(cfg *config.Config) vector.Embedder {
	// Only check for Ollama if no Zhipu AI is configured
	if cfg.Zhipu.ApiKey != "" {
		fmt.Println("Zhipu AI configured - skipping Ollama embedder initialization")
		return nil
	}
	
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

// Global variable to hold the AI client
var aiClient ai.Client

func initializeAI(cfg *config.Config) {
	// Initialize AI client based on configuration
	if cfg.Zhipu.ApiKey != "" {
		aiClient = ai.NewZhipuClient(cfg.Zhipu.ApiKey, cfg.Zhipu.BaseURL, cfg.Zhipu.Model)
		fmt.Println("Using Zhipu AI model:", cfg.Zhipu.Model)
	} else if cfg.Models["providers"] != nil {
		// Check for other providers like Minimax or Qwen
		providers := cfg.Models["providers"].(map[string]interface{})
		if len(providers) > 0 {
			// For now, we'll just detect that a provider exists
			fmt.Println("AI provider configured (Minimax/Qwen or other)")
			// In the future, we can add specific implementations for these providers
		} else {
			fmt.Println("No AI provider configured, using fallback responses")
		}
	} else {
		fmt.Println("No AI provider configured, using fallback responses")
	}
}

func callClaudeCode(prompt string) string {
	// Try to use configured AI client
	if aiClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		
		req := ai.ChatCompletionRequest{
			Model: "",
			Messages: []ai.Message{
				{Role: "user", Content: prompt},
			},
			Stream: false,
		}
		
		resp, err := aiClient.ChatCompletion(ctx, req)
		if err != nil {
			fmt.Printf("AI client error: %v\n", err)
			// Fallback to simple response
			return generateSimpleResponse(prompt)
		}
		
		if len(resp.Choices) > 0 {
			return strings.TrimSpace(resp.Choices[0].Message.Content)
		}
	}
	
	// Fallback to simple response
	return generateSimpleResponse(prompt)
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