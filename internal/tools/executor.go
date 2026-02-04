package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Executor handles tool execution
type Executor struct {
	registry *Registry
	timeout  time.Duration
}

// NewExecutor creates a new tool executor
func NewExecutor(registry *Registry) *Executor {
	return &Executor{
		registry: registry,
		timeout:  30 * time.Second, // Default timeout
	}
}

// SetTimeout sets the default timeout for tool execution
func (e *Executor) SetTimeout(timeout time.Duration) {
	e.timeout = timeout
}

// Execute executes a tool call
func (e *Executor) Execute(ctx context.Context, toolName string, params map[string]interface{}) (*ToolResult, error) {
	// Get tool from registry
	tool, err := e.registry.Get(toolName)
	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	// Validate parameters
	if err := tool.Validate(params); err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("parameter validation failed: %v", err),
		}, err
	}

	// Create context with timeout if not already set
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, e.timeout)
		defer cancel()
	}

	// Execute the tool
	resultChan := make(chan interface{}, 1)
	errChan := make(chan error, 1)

	go func() {
		result, err := tool.Execute(ctx, params)
		if err != nil {
			errChan <- err
			return
		}
		resultChan <- result
	}()

	// Wait for completion or timeout
	select {
	case result := <-resultChan:
		return &ToolResult{
			Success: true,
			Data:    result,
		}, nil
	case err := <-errChan:
		return &ToolResult{
			Success: false,
			Error:   err.Error(),
		}, err
	case <-ctx.Done():
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("tool execution timed out: %v", ctx.Err()),
		}, ctx.Err()
	}
}

// ExecuteMultiple executes multiple tool calls in sequence
func (e *Executor) ExecuteMultiple(ctx context.Context, calls []ToolCall) []ToolResult {
	results := make([]ToolResult, len(calls))

	for i, call := range calls {
		result, err := e.Execute(ctx, call.Name, call.Params)
		if err != nil {
			results[i] = *result
		} else {
			results[i] = *result
		}
	}

	return results
}

// ParseToolCall parses AI response to extract tool call information
// Supports multiple formats:
// 1. JSON format: {"tool": "tool_name", "params": {...}}
// 2. Natural language: "Use tool_name with params: ..."
func (e *Executor) ParseToolCall(aiResponse string) (*ToolCall, error) {
	// Try JSON format first
	if e.IsJSONToolCall(aiResponse) {
		return e.parseJSONToolCall(aiResponse)
	}

	// Try natural language format
	return e.parseNaturalLanguageCall(aiResponse)
}

// IsJSONToolCall checks if response contains JSON tool call
func (e *Executor) IsJSONToolCall(response string) bool {
	// Look for JSON-like structure
	jsonPattern := regexp.MustCompile(`\{[^}]*"tool"[\s]*:[\s]*"[^"]+"`)
	return jsonPattern.MatchString(response)
}

// parseJSONToolCall extracts tool call from JSON format
func (e *Executor) parseJSONToolCall(response string) (*ToolCall, error) {
	// Extract JSON object
	jsonStart := strings.Index(response, "{")
	if jsonStart == -1 {
		return nil, fmt.Errorf("no JSON object found")
	}

	jsonEnd := strings.LastIndex(response, "}")
	if jsonEnd == -1 {
		return nil, fmt.Errorf("invalid JSON object")
	}

	jsonStr := response[jsonStart : jsonEnd+1]

	// Parse JSON
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	// Extract tool name
	toolName, ok := data["tool"].(string)
	if !ok {
		// Try alternative key names
		if name, ok := data["name"].(string); ok {
			toolName = name
		} else {
			return nil, fmt.Errorf("tool name not found in JSON")
		}
	}

	// Extract parameters
	params, ok := data["params"].(map[string]interface{})
	if !ok {
		// Try alternative key names
		if p, ok := data["parameters"].(map[string]interface{}); ok {
			params = p
		} else {
			params = make(map[string]interface{})
		}
	}

	return &ToolCall{
		Name:   toolName,
		Params: params,
	}, nil
}

// parseNaturalLanguageCall extracts tool call from natural language
func (e *Executor) parseNaturalLanguageCall(response string) (*ToolCall, error) {
	lowerResponse := strings.ToLower(response)

	// Look for tool name patterns
	tools := e.registry.List()

	// Try to find any tool name in the response
	for _, tool := range tools {
		if strings.Contains(lowerResponse, strings.ToLower(tool.Name)) {
			// Extract parameters
			params := make(map[string]interface{})

			// Simple extraction: look for "param: value" patterns
			for paramName := range tool.Parameters {
				pattern := fmt.Sprintf(`%s[\s:]+([^\n]+)`, paramName)
				re := regexp.MustCompile(pattern)
				matches := re.FindStringSubmatch(response)
				if len(matches) > 1 {
					// Try to parse as different types
					value := strings.TrimSpace(matches[1])
					params[paramName] = e.parseParameterValue(value)
				}
			}

			return &ToolCall{
				Name:   tool.Name,
				Params: params,
			}, nil
		}
	}

	return nil, fmt.Errorf("no tool call found in response")
}

// parseParameterValue tries to parse a string value into appropriate type
func (e *Executor) parseParameterValue(value string) interface{} {
	value = strings.TrimSpace(value)

	// Remove quotes if present
	if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
		return strings.Trim(value, "\"")
	}

	if strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") {
		return strings.Trim(value, "'")
	}

	// Try boolean
	if strings.EqualFold(value, "true") {
		return true
	}
	if strings.EqualFold(value, "false") {
		return false
	}

	// Try to parse as number (very basic)
	// In production, you'd want more sophisticated parsing
	// For now, return as string and let the tool handle conversion
	return value
}

// FormatToolCall formats a tool call for display/logging
func (e *Executor) FormatToolCall(call *ToolCall) string {
	paramsJSON, _ := json.MarshalIndent(call.Params, "  ", "  ")
	return fmt.Sprintf("Tool: %s\nParams: %s", call.Name, string(paramsJSON))
}

// FormatToolResult formats a tool result for display/logging
func (e *Executor) FormatToolResult(result *ToolResult) string {
	if result.Success {
		dataJSON, _ := json.MarshalIndent(result.Data, "  ", "  ")
		return fmt.Sprintf("Success: true\nData: %s", string(dataJSON))
	}

	return fmt.Sprintf("Success: false\nError: %s", result.Error)
}
