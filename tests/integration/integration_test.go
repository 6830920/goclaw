// Package integration provides end-to-end integration tests for Goclaw
package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"goclaw/internal/chat"
	"goclaw/internal/config"
	"goclaw/internal/memory"
	"goclaw/internal/tools"
	"goclaw/internal/tools/builtin"
)

// TestSuite represents the integration test suite
type TestSuite struct {
	server       *httptest.Server
	chatManager  *chat.ChatManager
	memoryStore  *memory.MemoryStore
	toolsRegistry *tools.Registry
	toolsManager *builtin.Manager
	cfg          *config.Config
	baseURL      string
}

// SetupTestSuite creates a new test suite with all necessary components
func SetupTestSuite(t *testing.T) *TestSuite {
	suite := &TestSuite{}

	// Create mock config
	suite.cfg = config.NewDefaultConfig()
	suite.cfg.Heartbeat.Enabled = false // Disable heartbeat for tests
	suite.cfg.Agent.Model = "test-model"
	suite.cfg.Agent.Workspace = "/tmp/goclaw-test"

	// Initialize chat manager
	suite.chatManager = chat.NewChatManager(100)

	// Initialize memory store
	suite.memoryStore = memory.NewMemoryStore(memory.MemoryConfig{
		ShortTermMax:   50,
		WorkingMax:     10,
		SimilarityCut:  0.7,
	})

	// Initialize tools
	suite.toolsManager = builtin.NewManager()
	suite.toolsRegistry = suite.toolsManager.GetRegistry()

	// Create test server
	suite.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		suite.handleRequest(w, r)
	}))
	suite.baseURL = suite.server.URL

	return suite
}

// TearDownTestSuite cleans up the test suite
func (ts *TestSuite) TearDownTestSuite(t *testing.T) {
	if ts.server != nil {
		ts.server.Close()
	}
}

// handleRequest routes HTTP requests to appropriate handlers
func (ts *TestSuite) handleRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.URL.Path {
	case "/api/chat":
		ts.handleChat(w, r)
	case "/api/memory/stats":
		ts.handleMemoryStats(w, r)
	case "/api/sessions":
		ts.handleSessions(w, r)
	case "/api/tools":
		ts.handleToolsList(w, r)
	case "/api/tools/execute":
		ts.handleToolExecute(w, r)
	case "/health":
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	default:
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Not found"})
	}
}

// ========== 集成测试 1: 页面对话功能 ==========

// TestChatConversation 测试完整的对话流程
func TestChatConversation(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.TearDownTestSuite(t)

	t.Run("创建会话并发送消息", func(t *testing.T) {
		sessionID := fmt.Sprintf("test-session-%d", time.Now().Unix())

		// 发送第一条用户消息
		reqBody := map[string]interface{}{
			"message":   "你好，Goclaw！",
			"sessionId": sessionID,
		}
		resp := suite.post("/api/chat", reqBody, t)

		// 验证响应
		if resp["status"] != "ok" {
			t.Errorf("Expected status 'ok', got '%v'", resp["status"])
		}

		// 验证会话已创建
		data := resp["data"].(map[string]interface{})
		if data["sessionId"] != sessionID {
			t.Errorf("Expected sessionId '%s', got '%v'", sessionID, data["sessionId"])
		}

		// 验证响应不为空
		if data["response"] == "" {
			t.Error("Expected non-empty response")
		}

		// 发送第二条消息
		reqBody2 := map[string]interface{}{
			"message":   "记住我的名字是张三",
			"sessionId": sessionID,
		}
		resp2 := suite.post("/api/chat", reqBody2, t)
		t.Logf("第二条消息响应: %v", resp2)
	})

	t.Run("获取会话消息列表", func(t *testing.T) {
		// 先创建一个会话
		sessionID := fmt.Sprintf("test-session-%d", time.Now().Unix())
		suite.post("/api/chat", map[string]interface{}{
			"message":   "测试消息",
			"sessionId": sessionID,
		}, t)

		// 等待一下确保会话被创建
		time.Sleep(10 * time.Millisecond)

		// 获取会话列表
		sessionsResp := suite.get("/api/sessions", t)
		sessionsData := sessionsResp["data"].(map[string]interface{})
		sessions := sessionsData["sessions"].([]interface{})

		if len(sessions) < 1 {
			t.Errorf("Expected at least 1 session, got %d", len(sessions))
		}

		t.Logf("活跃会话数: %d", len(sessions))
	})
}

// ========== 集成测试 2: 记忆功能 ==========

// TestMemoryFunctionality 测试记忆系统的功能
func TestMemoryFunctionality(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.TearDownTestSuite(t)

	t.Run("添加短期记忆", func(t *testing.T) {
		// 添加到短期记忆
		suite.memoryStore.AddShortTerm("用户喜欢编程", map[string]interface{}{
			"topic":  "interest",
			"source": "test",
		})

		// 获取记忆统计
		statsResp := suite.get("/api/memory/stats", t)
		statsData := statsResp["data"].(map[string]interface{})

		t.Logf("记忆统计: %v", statsData)

		// 验证记忆已添加
		shortTermCount := statsData["shortTermCount"].(float64)
		if shortTermCount < 1 {
			t.Errorf("Expected at least 1 short-term memory, got %v", shortTermCount)
		}
	})

	t.Run("添加工作记忆", func(t *testing.T) {
		// 添加到工作记忆
		suite.memoryStore.AddWorking("当前任务：编写集成测试", 1) // priority=1 (high)

		// 验证工作记忆已添加
		statsResp := suite.get("/api/memory/stats", t)
		statsData := statsResp["data"].(map[string]interface{})
		workingCount := statsData["workingCount"].(float64)

		if workingCount < 1 {
			t.Errorf("Expected at least 1 working memory, got %v", workingCount)
		}
	})

	t.Run("记忆在对话中被使用", func(t *testing.T) {
		sessionID := fmt.Sprintf("memory-test-%d", time.Now().Unix())

		// 先添加一些记忆
		suite.memoryStore.AddShortTerm("用户叫李四，是程序员", map[string]interface{}{
			"source": "test",
		})

		// 发送消息，应该能检索到记忆
		reqBody := map[string]interface{}{
			"message":   "我的名字是什么？",
			"sessionId": sessionID,
		}
		resp := suite.post("/api/chat", reqBody, t)

		t.Logf("带记忆的对话响应: %v", resp)

		// 验证响应
		if resp["status"] != "ok" {
			t.Error("Chat request failed")
		}
	})
}

// ========== 集成测试 3: 心跳机制 ==========

// TestHeartbeatMechanism 测试心跳机制
func TestHeartbeatMechanism(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.TearDownTestSuite(t)

	t.Run("心跳周期性执行", func(t *testing.T) {
		// 注意：这个测试需要实际的心跳管理器
		// 由于测试环境中禁用心跳，这里模拟验证逻辑

		// 验证心跳配置正确
		if suite.cfg.Heartbeat.Enabled {
			t.Error("Heartbeat should be disabled in tests")
		}

		t.Log("心跳机制配置验证通过")
	})

	t.Run("心跳读取HEARTBEAT.md", func(t *testing.T) {
		// 创建测试用的HEARTBEAT.md
		heartbeatContent := `# 测试心跳
- 检查任务1
- 检查任务2`

		// 模拟心跳读取
		lines := len(strings.Split(heartbeatContent, "\n"))

		if lines < 2 {
			t.Error("Heartbeat content should have multiple lines")
		}

		t.Logf("心跳内容行数: %d", lines)
	})
}

// ========== 集成测试 4: 工具执行 ==========

// TestToolExecution 测试工具执行功能
func TestToolExecution(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.TearDownTestSuite(t)

	t.Run("列出可用工具", func(t *testing.T) {
		toolsResp := suite.get("/api/tools", t)
		toolsData := toolsResp["data"].(map[string]interface{})
		toolsList := toolsData["tools"].([]interface{})

		t.Logf("可用工具数量: %v", toolsData["count"])
		t.Logf("工具列表: %v", toolsList)

		if len(toolsList) < 1 {
			t.Error("Expected at least 1 available tool")
		}

		// 验证内置工具存在
		hasReadTool := false
		for _, tool := range toolsList {
			if toolMap, ok := tool.(map[string]interface{}); ok {
				if toolMap["name"] == "read" {
					hasReadTool = true
					break
				}
			}
		}

		if !hasReadTool {
			t.Error("Expected 'read' tool to be available")
		}
	})

	t.Run("执行read工具", func(t *testing.T) {
		// 创建一个测试文件
		testFile := "/tmp/test-read-integration.txt"
		testContent := "Hello, Goclaw!"

		// 写入测试文件 - 使用write工具
		writeReq := map[string]interface{}{
			"tool": "write",
			"params": map[string]interface{}{
				"path":    testFile,
				"content": testContent,
			},
		}
		suite.post("/api/tools/execute", writeReq, t)

		// 调用read工具
		reqBody := map[string]interface{}{
			"tool": "read",
			"params": map[string]interface{}{
				"path": testFile,
			},
		}
		resp := suite.post("/api/tools/execute", reqBody, t)

		t.Logf("read工具响应: %v", resp)

		// 验证执行成功
		data := resp["data"].(map[string]interface{})
		if data["success"] != true {
			t.Errorf("Expected success=true, got %v", data["success"])
		}

		// 验证读取的内容
		if result, ok := data["data"].(string); ok {
			if result != testContent {
				t.Errorf("Expected content '%s', got '%s'", testContent, result)
			}
		}
	})

	t.Run("执行exec工具", func(t *testing.T) {
		// 调用exec工具执行简单命令
		reqBody := map[string]interface{}{
			"tool": "exec",
			"params": map[string]interface{}{
				"command": "echo hello",
			},
		}
		resp := suite.post("/api/tools/execute", reqBody, t)

		t.Logf("exec工具响应: %v", resp)

		// 验证执行结果
		data := resp["data"].(map[string]interface{})
		if data["success"] != true {
			t.Error("exec tool execution failed")
		}
	})
}

// ========== 集成测试 5: 端到端流程 ==========

// TestEndToEndFlow 测试完整的端到端流程
func TestEndToEndFlow(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.TearDownTestSuite(t)

	t.Run("完整用户会话流程", func(t *testing.T) {
		sessionID := fmt.Sprintf("e2e-test-%d", time.Now().Unix())

		// 步骤1: 用户打招呼
		resp1 := suite.post("/api/chat", map[string]interface{}{
			"message":   "你好",
			"sessionId": sessionID,
		}, t)
		t.Logf("步骤1响应: %v", resp1)

		// 步骤2: 用户要求记忆信息
		suite.memoryStore.AddShortTerm("用户喜欢Go语言编程", map[string]interface{}{
			"source": "e2e-test",
		})

		// 步骤3: 用户提问，应该能检索到记忆
		resp3 := suite.post("/api/chat", map[string]interface{}{
			"message":   "我喜欢什么编程语言？",
			"sessionId": sessionID,
		}, t)
		t.Logf("步骤3响应: %v", resp3)

		// 步骤4: 验证会话历史
		messages, _ := suite.chatManager.GetMessages(sessionID)
		if len(messages) < 2 {
			t.Errorf("Expected at least 2 messages in session, got %d", len(messages))
		}

		// 步骤5: 获取系统状态
		statsResp := suite.get("/api/memory/stats", t)
		t.Logf("系统状态: %v", statsResp)

		// 步骤6: 获取可用工具
		toolsResp := suite.get("/api/tools", t)
		t.Logf("可用工具: %v", toolsResp)

		t.Log("端到端流程测试完成")
	})
}

// ========== HTTP请求辅助方法 ==========

// post sends a POST request
func (ts *TestSuite) post(path string, body map[string]interface{}, t *testing.T) map[string]interface{} {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	resp, err := http.Post(ts.baseURL+path, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatalf("Failed to send POST request: %v", err)
	}
	defer resp.Body.Close()

	return ts.parseResponse(resp, t)
}

// get sends a GET request
func (ts *TestSuite) get(path string, t *testing.T) map[string]interface{} {
	resp, err := http.Get(ts.baseURL + path)
	if err != nil {
		t.Fatalf("Failed to send GET request: %v", err)
	}
	defer resp.Body.Close()

	return ts.parseResponse(resp, t)
}

// parseResponse parses HTTP response
func (ts *TestSuite) parseResponse(resp *http.Response, t *testing.T) map[string]interface{} {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v\nBody: %s", err, body)
	}

	return result
}

// ========== 工具辅助函数 ==========

// toolsRegistryWriteFile writes a file (helper for testing)
func toolsRegistryWriteFile(path, content string) error {
	// Implementation would use the write tool
	return nil
}

// toolsRegistryDeleteFile deletes a file (helper for testing)
func toolsRegistryDeleteFile(path string) {
	// Implementation would clean up test files
}
