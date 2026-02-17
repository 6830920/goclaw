// Package integration provides HTTP handlers for integration tests
package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"goclaw/internal/chat"
	"goclaw/internal/tools"
)

// handleChat handles chat API requests
func (ts *TestSuite) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Message   string `json:"message"`
		SessionID string `json:"sessionId,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = fmt.Sprintf("api_session_%d", time.Now().UnixNano())
		ts.chatManager.CreateSession(sessionID, ts.cfg.Agent.Model)
	}

	// 添加用户消息
	ts.chatManager.AddMessage(sessionID, "user", req.Message)

	// 生成响应（模拟AI的工具调用）
	response := ts.generateTestResponse(req.Message, sessionID)

	// 添加助手消息
	ts.chatManager.AddMessage(sessionID, "assistant", response)

	// 添加到短期记忆
	ts.memoryStore.AddShortTerm(req.Message, map[string]interface{}{
		"session": sessionID,
		"source":  "integration-test",
	})

	// 获取更新后的消息
	messages, err := ts.chatManager.GetMessages(sessionID)
	if err != nil {
		messages = []chat.Message{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"data": map[string]interface{}{
			"sessionId": sessionID,
			"response":  response,
			"messages":  messages,
		},
	})
}

// generateTestResponse generates a test response (simulating AI tool calls)
func (ts *TestSuite) generateTestResponse(input, sessionID string) string {
	inputLower := strings.ToLower(input)

	// 模拟AI理解用户意图并返回工具调用
	// 参考 OpenClaw 的实现：AI 返回结构化的 toolCall 对象
	// 在测试环境中，我们简化处理，直接检测意图并执行工具

	// 检测"读取文件前N行"的意图
	if ts.containsFileReadIntent(inputLower) && strings.Contains(input, "/") {
		// 提取文件路径
		filePath := ts.extractFilePath(input)
		if filePath != "" {
			// 根据请求确定行数
			lineCount := ts.extractLineCount(input)
			// 直接执行工具并返回结果
			return ts.executeToolAndFormatResult(filePath, lineCount)
		}
	}

	// 简单的测试响应逻辑
	if strings.Contains(inputLower, "你好") || strings.Contains(inputLower, "hello") {
		return "你好！我是Goclaw，很高兴为你服务！"
	}

	if strings.Contains(inputLower, "名字") || strings.Contains(inputLower, "我是谁") {
		return "你是我的主人，我是Goclaw AI助手！"
	}

	if strings.Contains(inputLower, "记得") || strings.Contains(inputLower, "记住") {
		return "好的，我会记住这个信息。"
	}

	if strings.Contains(inputLower, "喜欢") || strings.Contains(inputLower, "喜欢什么") {
		return "根据我的记忆，你有很多兴趣爱好！"
	}

	// 默认响应
	return fmt.Sprintf("我收到了你的消息：%s\n这是测试环境下的模拟响应。", input)
}

// containsFileReadIntent checks if the input contains file reading intent
func (ts *TestSuite) containsFileReadIntent(inputLower string) bool {
	// 检查是否包含读取文件的关键词组合
	reads := []string{"展示", "显示", "读取", "查看", "看看", "读", "打开"}
	lines := []string{"前", "前几", "开头", "第一", "头", "几行", "行", "内容"}

	// 检查是否存在读取关键词
	hasReadKeyword := false
	for _, read := range reads {
		if strings.Contains(inputLower, strings.ToLower(read)) {
			hasReadKeyword = true
			break
		}
	}

	if !hasReadKeyword {
		return false
	}

	// 检查是否存在行数相关关键词
	for _, line := range lines {
		if strings.Contains(inputLower, strings.ToLower(line)) {
			return true
		}
	}

	return false
}

// extractLineCount extracts the number of lines to read from input
func (ts *TestSuite) extractLineCount(input string) int {
	// 默认3行
	defaultLines := 3

	// 检查是否明确指定了行数
	if strings.Contains(input, "前1行") || strings.Contains(input, "第一行") {
		return 1
	} else if strings.Contains(input, "前2行") || strings.Contains(input, "前两行") {
		return 2
	} else if strings.Contains(input, "前3行") || strings.Contains(input, "前三行") {
		return 3
	} else if strings.Contains(input, "前4行") || strings.Contains(input, "前四行") {
		return 4
	} else if strings.Contains(input, "前5行") || strings.Contains(input, "前五行") {
		return 5
	} else if strings.Contains(input, "开头几行") || strings.Contains(input, "前几行") {
		return 3 // 默认3行
	}

	return defaultLines
}

// extractFilePath extracts file path from user input
// extractFilePath extracts file path from user input
func (ts *TestSuite) extractFilePath(input string) string {
	// 查找 / 开头的路径
	startIdx := strings.Index(input, "/")
	if startIdx == -1 {
		return ""
	}

	// 高优先级：特定模式，按优先级排序
	// 1. "这个文件" 模式 (最高优先级)
	if idx := strings.Index(input[startIdx:], "这个文件"); idx != -1 {
		return strings.TrimSpace(input[startIdx : startIdx+idx])
	}
	
	// 2. "文件的前" 模式 (在"文件"处结束)
	if idx := strings.Index(input[startIdx:], "文件的前"); idx != -1 {
		return strings.TrimSpace(input[startIdx : startIdx+idx])
	}
	
	// 3. "文件的开头几行" 模式 (在"文件"处结束)
	if idx := strings.Index(input[startIdx:], "文件的开头几行"); idx != -1 {
		return strings.TrimSpace(input[startIdx : startIdx+idx])
	}
	
	// 4. "文件的第一部分" 模式 (在"文件"处结束)
	if idx := strings.Index(input[startIdx:], "文件的第一部分"); idx != -1 {
		return strings.TrimSpace(input[startIdx : startIdx+idx])
	}
	
	// 5. "，只要" 模式
	if idx := strings.Index(input[startIdx:], "，只要"); idx != -1 {
		return strings.TrimSpace(input[startIdx : startIdx+idx])
	}
	
	// 6. "只要" 模式
	if idx := strings.Index(input[startIdx:], "只要"); idx != -1 {
		return strings.TrimSpace(input[startIdx : startIdx+idx])
	}
	
	// 7. "的开头几行" 模式
	if idx := strings.Index(input[startIdx:], "的开头几行"); idx != -1 {
		return strings.TrimSpace(input[startIdx : startIdx+idx])
	}
	
	// 8. "的前" 模式（在文件扩展名后）
	if idx := strings.Index(input[startIdx:], "的前"); idx != -1 {
		// 检查是否在文件扩展名后
		pathPart := input[startIdx : startIdx+idx]
		// 查找文件扩展名
		for i := len(pathPart) - 1; i >= 0; i-- {
			if pathPart[i] == '/' {
				// 路径分隔符，返回完整路径
				return strings.TrimSpace(pathPart)
			}
			if pathPart[i] == '.' && i+2 < len(pathPart) {
				// 文件扩展名，返回完整路径
				return strings.TrimSpace(pathPart)
			}
		}
	}

	// 默认：返回从 / 开始到结束
	return strings.TrimSpace(input[startIdx:])
}
// executeToolAndFormatResult executes a tool and formats the result
func (ts *TestSuite) executeToolAndFormatResult(filePath string, lineCount int) string {
	// 创建执行器
	executor := tools.NewExecutor(ts.toolsRegistry)
	result, err := executor.Execute(context.Background(), "read", map[string]interface{}{
		"path": filePath,
	})

	if err != nil {
		return fmt.Sprintf("工具调用失败：%s", err.Error())
	}

	if !result.Success {
		return "工具执行失败"
	}

	// 根据read工具的返回格式解析结果
	// read工具返回 map[string]interface{} 包含 "content" 字段
	dataMap, ok := result.Data.(map[string]interface{})
	if !ok {
		return "无法解析工具结果"
	}

	// 获取内容
	content, ok := dataMap["content"].(string)
	if !ok {
		return "无法获取文件内容"
	}

	// 按行分割并取前N行
	lines := strings.Split(content, "\n")
	if len(lines) > lineCount {
		lines = lines[:lineCount]
	}

	// 格式化输出
	output := fmt.Sprintf("已读取文件：%s\n\n前%d行内容：\n", filePath, lineCount)
	for i, line := range lines {
		output += fmt.Sprintf("%d. %s\n", i+1, line)
	}

	return output
}

// handleMemoryStats handles memory stats API requests
func (ts *TestSuite) handleMemoryStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := ts.memoryStore.Stats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"data":   stats,
	})
}

// handleSessions handles sessions API requests
func (ts *TestSuite) handleSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessions := ts.chatManager.ListSessions()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"data": map[string]interface{}{
			"sessions":     sessions,
			"sessionCount": len(sessions),
		},
	})
}

// handleToolsList handles tools list API requests
func (ts *TestSuite) handleToolsList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tools := ts.toolsRegistry.List()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"data": map[string]interface{}{
			"count": len(tools),
			"tools": tools,
		},
	})
}

// handleToolExecute handles tool execution API requests
func (ts *TestSuite) handleToolExecute(w http.ResponseWriter, r *http.Request) {
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

	// 创建执行器
	executor := tools.NewExecutor(ts.toolsRegistry)
	result, err := executor.Execute(r.Context(), req.ToolName, req.Params)

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": err.Error(),
			"data":    result,
		})
	} else {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"data":   result,
		})
	}
}
