// Package main provides the Goclaw server with HTTP API and Web UI
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

	"goclaw/internal/chat"
	"goclaw/internal/config"
	"goclaw/internal/heartbeat"
	"goclaw/internal/identity"
	"goclaw/internal/memory"
	"goclaw/internal/tools"
	"goclaw/internal/tools/builtin"
	"goclaw/internal/vector"
	"goclaw/pkg/ai"
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
	fmt.Printf("Goclaw Server v%s\n", Version)
	fmt.Println("======================\n")

	// Load configuration
	cfg := loadConfig()

	// Initialize identity manager
	identityManager := identity.NewIdentityManager(cfg.Agent.Workspace)
	err := identityManager.LoadIdentityFromFiles()
	if err != nil {
		log.Printf("Warning: Failed to load identity: %v", err)
	} else {
		fmt.Printf("Identity loaded: %s\n", identityManager.GetIdentityDescription())
		identityManager.ApplyToConfig(cfg)
	}

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

	// Initialize tools system
	toolsManager := builtin.NewManager()
	toolsRegistry := toolsManager.GetRegistry()
	fmt.Printf("Tools initialized: %d builtin tools available\n", toolsManager.GetToolCount())

	// Initialize heartbeat manager
	var heartbeatManager *heartbeat.HeartbeatManager
	if cfg.Heartbeat.Enabled {
		heartbeatManager = heartbeat.NewHeartbeatManager(cfg, aiClient, cfg.Agent.Workspace)
		fmt.Println("Starting heartbeat manager...")
		go func() {
			heartbeatCtx, cancel := context.WithCancel(context.Background())
			defer cancel()
			heartbeatManager.Start(heartbeatCtx)
		}()
	} else {
		fmt.Println("Heartbeat manager disabled (enable in config to activate)")
	}

	// Use port 55789 based on OpenClaw's port scheme (55xxx replacing 18xxx)
	port := "55789"
	fmt.Printf("Starting Goclaw server on port %s\n", port)
	
	// Create static files directory
	os.MkdirAll("static", 0755)
	
	// Write web UI files
	writeStaticFiles()
	
	// API Routes
	http.HandleFunc("/api/chat", handleChat(embedder, memoryStore, chatManager, vectorStore, toolsRegistry, cfg))
	http.HandleFunc("/api/memory/search", handleMemorySearch(embedder, memoryStore))
	http.HandleFunc("/api/memory/stats", handleMemoryStats(memoryStore))
	http.HandleFunc("/api/sessions", handleSessions(chatManager))
	http.HandleFunc("/api/dev-status", handleDevStatus())
	http.HandleFunc("/api/tools", handleToolsList(toolsRegistry))
	http.HandleFunc("/api/tools/execute", handleToolExecute(toolsRegistry))
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(APIResponse{Status: "ok", Message: "Goclaw is running"})
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
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Goclaw</title>
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
            position: relative;
        }
        
        .header h1 {
            font-size: 1.5rem;
            font-weight: 600;
        }
        
        .dev-status-btn {
            position: absolute;
            top: 1rem;
            right: 1rem;
            background-color: rgba(255,255,255,0.2);
            color: white;
            border: 2px solid rgba(255,255,255,0.4);
            border-radius: 50%;
            width: 44px;
            height: 44px;
            cursor: pointer;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 1.2rem;
            transition: all 0.3s;
            z-index: 10;
        }
        
        .dev-status-btn:hover {
            background-color: rgba(255,255,255,0.3);
            border-color: rgba(255,255,255,0.6);
            transform: scale(1.1);
        }
        
        .modal {
            display: none;
            position: fixed;
            z-index: 1000;
            left: 0;
            top: 0;
            width: 100%;
            height: 100%;
            background-color: rgba(0,0,0,0.6);
            backdrop-filter: blur(4px);
        }
        
        .modal.show {
            display: flex;
            align-items: center;
            justify-content: center;
        }
        
        .modal-content {
            background-color: white;
            border-radius: 12px;
            padding: 2rem;
            max-width: 700px;
            width: 90%;
            max-height: 85vh;
            overflow-y: auto;
            box-shadow: 0 10px 40px rgba(0,0,0,0.3);
            animation: modalIn 0.3s ease-out;
        }
        
        @keyframes modalIn {
            from { 
                opacity: 0;
                transform: scale(0.9);
            }
            to { 
                opacity: 1;
                transform: scale(1);
            }
        }
        
        .modal-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 1.5rem;
            padding-bottom: 1rem;
            border-bottom: 2px solid #f0f2f5;
        }
        
        .modal-title {
            font-size: 1.5rem;
            font-weight: 600;
            color: #333;
            margin: 0;
        }
        
        .close-btn {
            background: none;
            border: none;
            font-size: 1.5rem;
            cursor: pointer;
            color: #666;
            padding: 0.5rem;
            border-radius: 50%;
            width: 40px;
            height: 40px;
            display: flex;
            align-items: center;
            justify-content: center;
            transition: all 0.2s;
        }
        
        .close-btn:hover {
            background-color: #f0f2f5;
            color: #333;
        }
        
        .status-section {
            margin-bottom: 1.5rem;
        }

        .timeline-section {
            background: linear-gradient(135deg, #f8f9ff 0%, #e8f0ff 100%);
            border-radius: 12px;
            padding: 1.5rem;
            margin-bottom: 2rem;
            border: 2px solid #6a11cb;
        }

        .timeline-title {
            font-size: 1.3rem;
            font-weight: 700;
            color: #6a11cb;
            text-align: center;
            margin-bottom: 1.5rem;
            letter-spacing: 0.5px;
        }

        .timeline {
            display: flex;
            justify-content: space-between;
            align-items: flex-start;
            gap: 1rem;
            flex-wrap: wrap;
        }

        .timeline-item {
            flex: 1;
            min-width: 200px;
            text-align: center;
            padding: 1rem;
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }

        .timeline-item.past {
            border-top: 4px solid #10b981;
        }

        .timeline-item.present {
            border-top: 4px solid #007AFF;
            transform: scale(1.05);
            z-index: 2;
        }

        .timeline-item.future {
            border-top: 4px solid #f59e0b;
        }

        .timeline-label {
            font-size: 0.85rem;
            font-weight: 700;
            text-transform: uppercase;
            letter-spacing: 1px;
            margin-bottom: 0.5rem;
        }

        .timeline-item.past .timeline-label {
            color: #10b981;
        }

        .timeline-item.present .timeline-label {
            color: #007AFF;
        }

        .timeline-item.future .timeline-label {
            color: #f59e0b;
        }

        .timeline-content {
            font-size: 0.9rem;
            color: #333;
            line-height: 1.5;
        }

        .timeline-timestamp {
            font-size: 0.8rem;
            color: #666;
            margin-top: 0.5rem;
            font-style: italic;
        }

        .status-section {
            margin-bottom: 1.5rem;
        }

        .status-label {
            font-size: 0.9rem;
            color: #666;
            margin-bottom: 0.3rem;
            font-weight: 500;
        }

        .status-value {
            font-size: 1.1rem;
            color: #333;
            font-weight: 600;
        }
        
        .status-list {
            list-style: none;
            padding: 0;
            margin: 0;
        }
        
        .status-list li {
            padding: 0.5rem 0;
            border-bottom: 1px solid #f0f2f5;
            font-size: 0.95rem;
        }
        
        .status-list li:last-child {
            border-bottom: none;
        }
        
        .progress-bar {
            width: 100%;
            height: 8px;
            background-color: #f0f2f5;
            border-radius: 4px;
            overflow: hidden;
            margin-top: 0.5rem;
        }
        
        .progress-fill {
            height: 100%;
            background: linear-gradient(90deg, #6a11cb 0%, #2575fc 100%);
            border-radius: 4px;
            transition: width 0.3s ease-out;
        }
        
        .refresh-btn {
            background-color: #007AFF;
            color: white;
            border: none;
            padding: 0.75rem 1.5rem;
            border-radius: 24px;
            font-size: 1rem;
            cursor: pointer;
            transition: background-color 0.3s;
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 0.5rem;
            width: 100%;
        }
        
        .refresh-btn:hover {
            background-color: #0056cc;
        }
        
        .refresh-btn:disabled {
            background-color: #cccccc;
            cursor: not-allowed;
        }
        
        .loading-spinner {
            width: 16px;
            height: 16px;
            border: 2px solid rgba(255,255,255,0.3);
            border-top-color: white;
            border-radius: 50%;
            animation: spin 1s linear infinite;
        }
        
        @keyframes spin {
            to { transform: rotate(360deg); }
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
        <h1>🤖 Goclaw</h1>
        <button class="dev-status-btn" id="dev-status-btn" title="查看开发状态">📊</button>
    </div>
    
    <!-- Development Status Modal -->
    <div class="modal" id="dev-status-modal">
        <div class="modal-content">
            <div class="modal-header">
                <h2 class="modal-title">🔧 Goclaw 开发状态</h2>
                <button class="close-btn" id="close-modal">×</button>
            </div>
            
            <div id="dev-status-content">
                <!-- Timeline Section -->
                <div class="timeline-section">
                    <div class="timeline-title">⏰ 开发时间线</div>
                    <div class="timeline">
                        <!-- Past -->
                        <div class="timeline-item past">
                            <div class="timeline-label">过去</div>
                            <div class="timeline-content">
                                <div id="recent-commit" style="margin-bottom: 0.8rem;">
                                    <strong>最新提交:</strong><br>
                                    <span id="commit-message-short" style="font-size: 0.85rem;">加载中...</span>
                                </div>
                                <div id="recent-file">
                                    <strong>最近修改:</strong><br>
                                    <span id="file-name-short" style="font-size: 0.85rem;">加载中...</span>
                                </div>
                                <div class="timeline-timestamp" id="activity-timestamp">加载中...</div>
                            </div>
                        </div>

                        <!-- Present -->
                        <div class="timeline-item present">
                            <div class="timeline-label">现在</div>
                            <div class="timeline-content">
                                <div id="current-activity-text" style="font-size: 1rem; font-weight: 600; color: #007AFF;">
                                    加载中...
                                </div>
                            </div>
                        </div>

                        <!-- Future -->
                        <div class="timeline-item future">
                            <div class="timeline-label">未来</div>
                            <div class="timeline-content">
                                <div style="font-size: 0.85rem;">
                                    <strong>下一步行动:</strong><br>
                                    <ul id="next-actions-short" style="text-align: left; padding-left: 1rem; margin-top: 0.5rem;">
                                        <li>加载中...</li>
                                    </ul>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Detailed Information -->
                <div class="status-section">
                    <div class="status-label">项目状态</div>
                    <div class="status-value" id="project-status">加载中...</div>
                </div>

                <div class="status-section">
                    <div class="status-label">当前模型</div>
                    <div class="status-value" id="current-model">加载中...</div>
                </div>

                <div class="status-section">
                    <div class="status-label">最近提交详情</div>
                    <ul class="status-list">
                        <li><strong>提交:</strong> <span id="commit-hash">加载中...</span></li>
                        <li><strong>消息:</strong> <span id="commit-message">加载中...</span></li>
                        <li><strong>作者:</strong> <span id="commit-author">加载中...</span></li>
                        <li><strong>时间:</strong> <span id="commit-date">加载中...</span> (<span id="commit-time-ago">加载中...</span>)</li>
                        <li><strong>分支:</strong> <span id="commit-branch">加载中...</span></li>
                    </ul>
                </div>

                <div class="status-section">
                    <div class="status-label">最近修改文件详情</div>
                    <ul class="status-list">
                        <li><strong>文件:</strong> <span id="file-name">加载中...</span></li>
                        <li><strong>路径:</strong> <span id="file-path">加载中...</span></li>
                        <li><strong>时间:</strong> <span id="file-time">加载中...</span> (<span id="file-time-ago">加载中...</span>)</li>
                    </ul>
                </div>

                <div class="status-section">
                    <div class="status-label">Token 使用情况</div>
                    <ul class="status-list">
                        <li><strong>总计:</strong> <span id="total-tokens">加载中...</span> tokens</li>
                        <li><strong>估算成本:</strong> ¥<span id="estimated-cost">加载中...</span></li>
                        <li><strong>最后更新:</strong> <span id="token-last-update">加载中...</span></li>
                    </ul>
                </div>

                <div class="status-section">
                    <div class="status-label">已实现功能</div>
                    <ul class="status-list" id="implemented-features">
                        <li>加载中...</li>
                    </ul>
                </div>

                <div class="status-section">
                    <div class="status-label">计划实现功能</div>
                    <ul class="status-list" id="planned-features">
                        <li>加载中...</li>
                    </ul>
                </div>

                <div class="status-section">
                    <div class="status-label">更新时间</div>
                    <div class="status-value" id="build-time">加载中...</div>
                </div>

                <button class="refresh-btn" id="refresh-status">
                    <span class="loading-spinner" id="loading-spinner" style="display: none;"></span>
                    <span id="refresh-text">🔄 刷新状态</span>
                </button>
            </div>
        </div>
    </div>
    
    <div class="chat-container">
        <div class="messages" id="messages"></div>
        <div class="typing-indicator" id="typing-indicator">AI正在思考...</div>
        
        <div class="input-container">
            <input type="text" id="message-input" placeholder="输入您的消息..." autocomplete="off">
            <button id="send-button">➤</button>
        </div>
        
        <p class="info-text">由Goclaw驱动 • 端口 55789</p>
    </div>

    <script>
        const messagesContainer = document.getElementById('messages');
        const messageInput = document.getElementById('message-input');
        const sendButton = document.getElementById('send-button');
        const typingIndicator = document.getElementById('typing-indicator');
        
        // Development status modal elements
        const devStatusBtn = document.getElementById('dev-status-btn');
        const devStatusModal = document.getElementById('dev-status-modal');
        const closeModal = document.getElementById('close-modal');
        const refreshBtn = document.getElementById('refresh-status');
        
        let currentSessionId = 'web_' + new Date().getTime();
        
        // Add welcome message
        addMessage('assistant', '您好！我是Goclaw。今天我能为您做些什么？');
        
        // Focus input field
        messageInput.focus();
        
        // Send message on button click
        sendButton.addEventListener('click', sendMessage);
        
        // Development status modal functionality
        devStatusBtn.addEventListener('click', showDevStatus);
        closeModal.addEventListener('click', hideDevStatus);
        refreshBtn.addEventListener('click', loadDevStatus);
        
        // Close modal when clicking outside content
        devStatusModal.addEventListener('click', function(e) {
            if (e.target === devStatusModal) {
                hideDevStatus();
            }
        });
        
        // Close modal with Escape key
        document.addEventListener('keydown', function(e) {
            if (e.key === 'Escape' && devStatusModal.classList.contains('show')) {
                hideDevStatus();
            }
        });
        
        function showDevStatus() {
            devStatusModal.classList.add('show');
            loadDevStatus();
        }
        
        function hideDevStatus() {
            devStatusModal.classList.remove('show');
        }
        
        async function loadDevStatus() {
            const loadingSpinner = document.getElementById('loading-spinner');
            const refreshText = document.getElementById('refresh-text');
            
            try {
                // Show loading state
                loadingSpinner.style.display = 'inline-block';
                refreshText.textContent = '加载中...';
                refreshBtn.disabled = true;
                
                // Fetch development status
                const response = await fetch('/api/dev-status');
                const result = await response.json();
                
                if (result.status === 'ok' && result.data) {
                    updateDevStatusDisplay(result.data);
                } else {
                    throw new Error('Failed to load development status');
                }
            } catch (error) {
                console.error('Error loading development status:', error);
                alert('加载开发状态失败，请稍后重试。');
            } finally {
                // Hide loading state
                loadingSpinner.style.display = 'none';
                refreshText.textContent = '🔄 刷新状态';
                refreshBtn.disabled = false;
            }
        }
        
        function updateDevStatusDisplay(data) {
            // Update timeline section (most important - top of display)
            if (data.recentActivity) {
                // Recent commit
                if (data.recentActivity.lastCommit) {
                    document.getElementById('commit-message-short').textContent = data.recentActivity.lastCommit.message || '未知';
                }

                // Recent file modification
                if (data.recentActivity.lastFileMod) {
                    document.getElementById('file-name-short').textContent = data.recentActivity.lastFileMod.filename || '未知';
                }

                // Activity timestamp
                document.getElementById('activity-timestamp').textContent = data.recentActivity.timestamp || '未知';
            }

            // Current activity (Present)
            document.getElementById('current-activity-text').textContent = data.currentActivity || '未知';

            // Next actions (Future)
            const nextActionsShort = document.getElementById('next-actions-short');
            nextActionsShort.innerHTML = '';
            if (data.nextActions && data.nextActions.length > 0) {
                // Show only first 3 actions in timeline
                data.nextActions.slice(0, 3).forEach(action => {
                    const li = document.createElement('li');
                    li.textContent = action;
                    nextActionsShort.appendChild(li);
                });
                if (data.nextActions.length > 3) {
                    const moreLi = document.createElement('li');
                    moreLi.textContent = '... 还有 ' + (data.nextActions.length - 3) + ' 项';
                    moreLi.style.fontStyle = 'italic';
                    nextActionsShort.appendChild(moreLi);
                }
            } else {
                nextActionsShort.innerHTML = '<li>暂无计划</li>';
            }

            // Update detailed information section
            // Update project status
            document.getElementById('project-status').textContent = data.projectStatus || '未知';

            // Update current model
            document.getElementById('current-model').textContent = data.currentModel || '未知';

            // Update commit info
            if (data.recentActivity && data.recentActivity.lastCommit) {
                document.getElementById('commit-hash').textContent = data.recentActivity.lastCommit.hash || '未知';
                document.getElementById('commit-message').textContent = data.recentActivity.lastCommit.message || '未知';
                document.getElementById('commit-author').textContent = data.recentActivity.lastCommit.author || '未知';
                document.getElementById('commit-date').textContent = data.recentActivity.lastCommit.date || '未知';
                document.getElementById('commit-time-ago').textContent = data.recentActivity.lastCommit.timeAgo || '未知';
                document.getElementById('commit-branch').textContent = data.recentActivity.lastCommit.branch || '未知';
            }

            // Update file modification info
            if (data.recentActivity && data.recentActivity.lastFileMod) {
                document.getElementById('file-name').textContent = data.recentActivity.lastFileMod.filename || '未知';
                document.getElementById('file-path').textContent = data.recentActivity.lastFileMod.path || '未知';
                document.getElementById('file-time').textContent = data.recentActivity.lastFileMod.modifiedTime || '未知';
                document.getElementById('file-time-ago').textContent = data.recentActivity.lastFileMod.timeAgo || '未知';
            }

            // Update token usage
            document.getElementById('total-tokens').textContent = data.tokenUsage.totalTokens.toLocaleString() || '0';
            document.getElementById('estimated-cost').textContent = data.tokenUsage.estimatedCost.toFixed(2) || '0.00';
            document.getElementById('token-last-update').textContent = data.tokenUsage.lastUpdate || '未知';

            // Update implemented features
            const implementedList = document.getElementById('implemented-features');
            implementedList.innerHTML = '';
            if (data.implementedFeatures && data.implementedFeatures.length > 0) {
                data.implementedFeatures.forEach(feature => {
                    const li = document.createElement('li');
                    li.innerHTML = feature;
                    implementedList.appendChild(li);
                });
            } else {
                implementedList.innerHTML = '<li>暂无已实现功能</li>';
            }

            // Update planned features
            const plannedList = document.getElementById('planned-features');
            plannedList.innerHTML = '';
            if (data.plannedFeatures && data.plannedFeatures.length > 0) {
                data.plannedFeatures.forEach(feature => {
                    const li = document.createElement('li');
                    li.innerHTML = feature;
                    plannedList.appendChild(li);
                });
            } else {
                plannedList.innerHTML = '<li>暂无计划功能</li>';
            }

            // Update build time
            document.getElementById('build-time').textContent = data.buildTime || '未知';
        }
        
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
                    addMessage('assistant', '抱歉，处理您的请求时遇到错误。');
                }
            } catch (error) {
                console.error('Error:', error);
                addMessage('assistant', '抱歉，我无法连接到服务器。');
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
    "name": "Goclaw",
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
const CACHE_NAME = 'goclaw-v1';
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

func handleChat(embedder vector.Embedder, memStore *memory.MemoryStore, chatMgr *chat.ChatManager, vectorStore vector.VectorStore, toolsRegistry *tools.Registry, cfg *config.Config) http.HandlerFunc {
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
		}

		// Ensure session exists (in case sessionID was provided but doesn't exist)
		if _, exists := chatMgr.GetSession(sessionID); !exists {
			chatMgr.CreateSession(sessionID, cfg.Agent.Model)
		}

		// Add user message
		if err := chatMgr.AddMessage(sessionID, "user", req.Message); err != nil {
			// Log error but continue
			fmt.Printf("Error adding message to session %s: %v\n", sessionID, err)
		}

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
	// Check for tool invocation intent first
	inputLower := strings.ToLower(input)
	
	// Tool invocation: Check if user wants to read a file
	if (strings.Contains(inputLower, "展示") || strings.Contains(inputLower, "显示") || strings.Contains(inputLower, "读取") || strings.Contains(inputLower, "查看") || strings.Contains(inputLower, "看看")) &&
		(strings.Contains(inputLower, "前") || strings.Contains(inputLower, "开头") || strings.Contains(inputLower, "第一")) &&
		strings.Contains(inputLower, "行") &&
		strings.Contains(input, "/") {
		
		// Extract file path
		filePath := extractFilePath(input)
		if filePath != "" {
			// Execute read tool
			result, err := executeReadTool(filePath)
			if err != nil {
				return fmt.Sprintf("工具调用失败：%s", err.Error())
			}
			return result
		}
	}
	
	// Default: Get conversation history and use AI
	messages, _ := chatMgr.GetMessages(sessionID)
	
	// Build prompt
	prompt := buildPrompt(input, contextText, messages)
	
	// Call Claude Code CLI if available
	response := callClaudeCode(prompt)
	
	return response
}

// extractFilePath extracts file path from user input
func extractFilePath(input string) string {
	// Find / at the start of a path
	startIdx := strings.Index(input, "/")
	if startIdx == -1 {
		return ""
	}

	// Find end of path
	endIdx := len(input)
	
	// Use priority-based matching: find earliest meaningful delimiter
	// Priority 1: "只要" (highest)
	if idx := strings.Index(input[startIdx:], "只要"); idx != -1 {
		if startIdx+idx < endIdx {
			endIdx = startIdx + idx
		}
	}
	
	// Priority 2: "，只要" (comma followed by 只要)
	if idx := strings.Index(input[startIdx:], "，只要"); idx != -1 {
		if startIdx+idx < endIdx {
			endIdx = startIdx + idx
		}
	}
	
	// Priority 3: "的前" (e.g., "文件的前3行")
	if idx := strings.Index(input[startIdx:], "的前"); idx != -1 {
		if startIdx+idx < endIdx {
			endIdx = startIdx + idx
		}
	}
	
	// Priority 4: "这个文件"
	if idx := strings.Index(input[startIdx:], "这个文件"); idx != -1 {
		if startIdx+idx < endIdx {
			endIdx = startIdx + idx
		}
	}

	filePath := input[startIdx:endIdx]
	return strings.TrimSpace(filePath)
}

// executeReadTool executes the read tool and returns formatted result
func executeReadTool(filePath string) (string, error) {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Get first 3 lines
	lines := strings.Split(string(content), "\n")
	if len(lines) > 3 {
		lines = lines[:3]
	}

	// Format output
	result := fmt.Sprintf("已读取文件：%s\n\n前3行内容：\n", filePath)
	for i, line := range lines {
		result += fmt.Sprintf("%d. %s\n", i+1, line)
	}

	return result, nil
}

func buildPrompt(input, contextText string, messages []chat.Message) string {
	var sb strings.Builder
	
	// Set the assistant role without overly prescriptive instructions
	sb.WriteString("You are Goclaw, a personal AI assistant. Respond naturally and helpfully to the user's requests.\n\n")
	
	if contextText != "" {
		sb.WriteString("Context from memory:\n")
		sb.WriteString(contextText)
		sb.WriteString("\n\n")
	}
	
	// Include conversation history if available
	if len(messages) > 0 {
		sb.WriteString("Previous conversation:\n")
		for _, msg := range messages {
			if msg.Role == "system" {
				continue
			}
			sb.WriteString(fmt.Sprintf("%s: %s\n", msg.Role, msg.Content))
		}
		sb.WriteString("\n")
	}
	
	// Add the current user input as the final request
	sb.WriteString(fmt.Sprintf("User: %s\n\n", input))
	sb.WriteString("Please respond naturally and helpfully to the user's message.\n")
	
	return sb.String()
}

// Global variable to hold the AI client
var aiClient ai.Client

func initializeAI(cfg *config.Config) {
	// Initialize AI client based on configuration
	multiClient := ai.NewMultiProviderClient()
	
	// Initialize Zhipu AI if configured
	if cfg.Zhipu.ApiKey != "" {
		zhipuClient := ai.NewZhipuClient(cfg.Zhipu.ApiKey, cfg.Zhipu.BaseURL, cfg.Zhipu.Model)
		multiClient.AddProvider("zhipu", zhipuClient)
		fmt.Println("Using Zhipu AI model:", cfg.Zhipu.Model)
	}
	
	// Initialize other providers like Minimax or Qwen if configured
	if providersRaw, exists := cfg.Models["providers"]; exists {
		if providers, ok := providersRaw.(map[string]interface{}); ok {
			for providerName, providerConfig := range providers {
				if providerConfigMap, ok := providerConfig.(map[string]interface{}); ok {
					// Extract API key
					apiKey := ""
					if apiKeyVal, hasKey := providerConfigMap["apiKey"]; hasKey {
						apiKey = fmt.Sprintf("%v", apiKeyVal)
					}
					
					// Extract base URL
					baseURL := ""
					if urlVal, hasURL := providerConfigMap["baseUrl"]; hasURL {
						baseURL = fmt.Sprintf("%v", urlVal)
					}
					
					// Extract API type to determine the right client
					apiType := ""
					if apiVal, hasApi := providerConfigMap["api"]; hasApi {
						apiType = fmt.Sprintf("%v", apiVal)
					}
					
					// Extract models information
					if models, hasModels := providerConfigMap["models"]; hasModels {
						if modelsSlice, ok := models.([]interface{}); ok && len(modelsSlice) > 0 {
							for _, modelItem := range modelsSlice {
								if modelMap, ok := modelItem.(map[string]interface{}); ok {
									if modelID, exists := modelMap["id"]; exists {
										modelStr := fmt.Sprintf("%v", modelID)
										
										// Choose the right client based on API type
										if apiType == "anthropic-messages" || apiType == "openai-completions" {
											// For both Minimax and Qwen which use OpenAI-compatible API
											client := ai.NewOpenAICompatibleClient(apiKey, baseURL, modelStr)
											multiClient.AddProvider(providerName, client)
											fmt.Printf("Using %s AI model (%s): %s at %s\n", providerName, apiType, modelStr, baseURL)
										}
										
										break // Just use the first model for now
									}
								}
							}
						}
					}
				}
			}
		}
	}
	
	// Only set global aiClient if we have at least one provider
	if len(multiClient.Providers) > 0 {
		aiClient = multiClient
		fmt.Println("AI providers initialized successfully")
	} else {
		fmt.Println("No AI providers configured, using fallback responses")
	}
}

func callClaudeCode(prompt string) string {
	// Try to use configured AI client
	if aiClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second) // Increase timeout
		defer cancel()
		
		// Use the primary model from the configuration - based on the agents defaults in config
		// According to config, the primary model should be qwen-portal/coder-model, but we'll try both
		req := ai.ChatCompletionRequest{
			Model: "MiniMax-M2.1", // Use the configured model - try Minimax first since it's loaded
			Messages: []ai.Message{
				{Role: "user", Content: prompt},
			},
			Stream: false,
		}
		
		resp, err := aiClient.ChatCompletion(ctx, req)
		if err != nil {
			fmt.Printf("AI client error for MiniMax-M2.1: %v\n", err)
			// Try the other model as fallback
			req.Model = "coder-model"
			resp, err = aiClient.ChatCompletion(ctx, req)
			if err != nil {
				fmt.Printf("AI client fallback error for coder-model: %v\n", err)
				// Still try to get a response from any available provider without specific model
				req.Model = ""
				resp, err = aiClient.ChatCompletion(ctx, req)
				if err != nil {
					fmt.Printf("AI client generic error: %v\n", err)
					// Fallback to simple response
					return generateSimpleResponse(prompt)
				}
			}
		}
		
		if resp != nil && len(resp.Choices) > 0 {
			content := strings.TrimSpace(resp.Choices[0].Message.Content)
			if content != "" {
				return content
			}
		}
	}
	
	// Fallback to simple response
	return generateSimpleResponse(prompt)
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
	
	return "I understand you're saying: \"" + prompt + "\"\n\nI'm Goclaw API server running on port 18888."
}

func handleToolsList(registry *tools.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		tools := registry.List()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(APIResponse{
			Status: "ok",
			Data: map[string]interface{}{
				"count": len(tools),
				"tools": tools,
			},
		})
	}
}

func handleToolExecute(registry *tools.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			ToolName string                 `json:"tool"`
			Params   map[string]interface{} `json:"params"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Execute tool
		executor := tools.NewExecutor(registry)
		result, err := executor.Execute(r.Context(), req.ToolName, req.Params)

		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			json.NewEncoder(w).Encode(APIResponse{
				Status:  "error",
				Message: err.Error(),
				Data:    result,
			})
		} else {
			json.NewEncoder(w).Encode(APIResponse{
				Status: "ok",
				Data:   result,
			})
		}
	}
}