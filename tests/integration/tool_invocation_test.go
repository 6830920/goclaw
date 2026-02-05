// Package integration provides end-to-end integration tests for Goclaw
package integration

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// ========== 工具调用功能集成测试 ==========

// TestToolInvocation_展示文件前三行 测试用户通过对话调用工具读取文件
func TestToolInvocation_展示文件前三行(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.TearDownTestSuite(t)

	t.Run("用户请求读取文件前三行", func(t *testing.T) {
		// 1. 准备测试文件
		testFile := "/tmp/test-read-lines.txt"
		testContent := "第一行内容\n第二行内容\n第三行内容\n第四行内容\n第五行内容"
		
		// 使用write工具创建测试文件
		writeReq := map[string]interface{}{
			"tool": "write",
			"params": map[string]interface{}{
				"path":    testFile,
				"content": testContent,
			},
		}
		suite.post("/api/tools/execute", writeReq, t)
		
		// 2. 用户发起对话请求
		sessionID := fmt.Sprintf("tool-test-%d", time.Now().Unix())
		chatReq := map[string]interface{}{
			"message":   fmt.Sprintf("给我展示%s这个文件的前三行", testFile),
			"sessionId": sessionID,
		}
		chatResp := suite.post("/api/chat", chatReq, t)
		
		t.Logf("聊天响应: %v", chatResp)
		
		// 3. 验证响应
		if chatResp["status"] != "ok" {
			t.Errorf("Expected status 'ok', got '%v'", chatResp["status"])
		}
		
		data := chatResp["data"].(map[string]interface{})
		response := data["response"].(string)
		
		// 4. 验证响应中包含前三行内容
		expectedLines := []string{"第一行内容", "第二行内容", "第三行内容"}
		for _, expectedLine := range expectedLines {
			if !strings.Contains(response, expectedLine) {
				t.Errorf("响应应该包含 '%s'，实际响应: %s", expectedLine, response)
			}
		}
		
		// 5. 验证响应中不包含第四、五行
		unexpectedLines := []string{"第四行内容", "第五行内容"}
		for _, unexpectedLine := range unexpectedLines {
			if strings.Contains(response, unexpectedLine) {
				t.Errorf("响应不应该包含 '%s'，实际响应: %s", unexpectedLine, response)
			}
		}
	})
	
	t.Run("用户使用自然语言请求读取文件", func(t *testing.T) {
		// 1. 准备测试文件
		testFile := "/tmp/test-natural-lang.txt"
		testContent := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5"
		
		suite.post("/api/tools/execute", map[string]interface{}{
			"tool": "write",
			"params": map[string]interface{}{
				"path":    testFile,
				"content": testContent,
			},
		}, t)
		
		// 2. 用户使用不同的自然语言表达
		sessionID := fmt.Sprintf("natural-%d", time.Now().Unix())
		variations := []string{
			fmt.Sprintf("读取%s的前3行", testFile),
			fmt.Sprintf("帮我看看%s的开头几行", testFile),
			fmt.Sprintf("显示%s文件的第一部分，只要前三行", testFile),
		}
		
		for i, userMessage := range variations {
			t.Logf("测试变体 %d: %s", i+1, userMessage)
			
			chatResp := suite.post("/api/chat", map[string]interface{}{
				"message":   userMessage,
				"sessionId": sessionID,
			}, t)
			
			if chatResp["status"] != "ok" {
				t.Errorf("变体 %d 失败: %v", i+1, chatResp)
			}
			
			data := chatResp["data"].(map[string]interface{})
			response := data["response"].(string)
			
			// 至少应该包含文件内容的某些部分
			if !strings.Contains(response, "Line") {
				t.Errorf("变体 %d 的响应应该包含文件内容，实际: %s", i+1, response)
			}
		}
	})
}

// TestToolInvocation_工具执行验证 测试工具是否被正确执行
func TestToolInvocation_工具执行验证(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.TearDownTestSuite(t)
	
	t.Run("验证read工具被调用", func(t *testing.T) {
		testFile := "/tmp/verify-read.txt"
		testContent := "测试read工具调用"
		
		// 创建文件
		suite.post("/api/tools/execute", map[string]interface{}{
			"tool": "write",
			"params": map[string]interface{}{
				"path":    testFile,
				"content": testContent,
			},
		}, t)
		
		// 检查工具列表
		toolsResp := suite.get("/api/tools", t)
		toolsData := toolsResp["data"].(map[string]interface{})
		toolsList := toolsData["tools"].([]interface{})
		
		// 验证read工具存在
		hasReadTool := false
		for _, tool := range toolsList {
			if toolMap, ok := tool.(map[string]interface{}); ok {
				if toolMap["name"] == "read" {
					hasReadTool = true
					t.Logf("read工具参数: %v", toolMap["parameters"])
					break
				}
			}
		}
		
		if !hasReadTool {
			t.Error("read工具应该存在")
		}
		
		// 调用read工具
		readResp := suite.post("/api/tools/execute", map[string]interface{}{
			"tool": "read",
			"params": map[string]interface{}{
				"path": testFile,
			},
		}, t)
		
		t.Logf("read工具响应: %v", readResp)
		
		readData := readResp["data"].(map[string]interface{})
		if readData["success"] != true {
			t.Errorf("read工具应该成功，实际: %v", readData)
		}
		
		// 验证读取的内容
		result := readData["data"].(map[string]interface{})
		if content, ok := result["content"].(string); ok {
			if !strings.Contains(content, testContent) {
				t.Errorf("读取的内容不正确，期望 '%s'，实际 '%s'", testContent, content)
			}
		}
	})
}

// TestToolInvocation_错误处理 测试工具调用失败时的处理
func TestToolInvocation_错误处理(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.TearDownTestSuite(t)
	
	t.Run("读取不存在的文件", func(t *testing.T) {
		nonExistentFile := "/tmp/non-existent-file-12345.txt"
		
		sessionID := fmt.Sprintf("error-test-%d", time.Now().Unix())
		chatResp := suite.post("/api/chat", map[string]interface{}{
			"message":   fmt.Sprintf("读取%s的前三行", nonExistentFile),
			"sessionId": sessionID,
		}, t)
		
		t.Logf("错误处理响应: %v", chatResp)
		
		// 验证响应状态（可能是ok，但包含错误信息）
		if chatResp["status"] != "ok" {
			t.Logf("注意：错误情况下返回status: %v", chatResp["status"])
		}
		
		data := chatResp["data"].(map[string]interface{})
		response := data["response"].(string)
		
		// 响应应该包含错误提示
		if strings.Contains(response, "成功") || strings.Contains(response, "完成") {
			t.Error("对于不存在的文件，响应不应该暗示成功")
		}
	})
	
	t.Run("缺少必需的参数", func(t *testing.T) {
		// 请求调用read工具但不提供path参数
		resp := suite.post("/api/tools/execute", map[string]interface{}{
			"tool": "read",
			"params": map[string]interface{}{},
		}, t)
		
		t.Logf("参数缺失响应: %v", resp)
		
		// 应该返回错误
		data := resp["data"].(map[string]interface{})
		if data["success"] == true {
			t.Error("缺少path参数时，read工具应该失败")
		}
	})
}

// TestToolInvocation_多行读取 测试不同行数的读取
func TestToolInvocation_多行读取(t *testing.T) {
	suite := SetupTestSuite(t)
	defer suite.TearDownTestSuite(t)
	
	t.Run("读取前5行", func(t *testing.T) {
		testFile := "/tmp/test-five-lines.txt"
		testContent := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6\nLine 7\nLine 8"
		
		suite.post("/api/tools/execute", map[string]interface{}{
			"tool": "write",
			"params": map[string]interface{}{
				"path":    testFile,
				"content": testContent,
			},
		}, t)
		
		// 直接调用read工具
		readResp := suite.post("/api/tools/execute", map[string]interface{}{
			"tool": "read",
			"params": map[string]interface{}{
				"path": testFile,
			},
		}, t)
		
		readData := readResp["data"].(map[string]interface{})
		if readData["success"] != true {
			t.Errorf("read工具执行失败: %v", readData)
		}
		
		result := readData["data"].(map[string]interface{})
		content := result["content"].(string)
		
		// 验证文件内容被完整读取
		if !strings.Contains(content, "Line 8") {
			t.Errorf("应该读取完整的文件内容")
		}
	})
}
