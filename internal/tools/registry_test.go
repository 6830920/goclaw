package tools

import (
	"context"
	"testing"
	"time"
)

func TestToolValidate(t *testing.T) {
	tool := &Tool{
		Name:        "test_tool",
		Description: "A test tool",
		Parameters: map[string]Parameter{
			"required_param": {
				Type:     "string",
				Required: true,
			},
			"optional_param": {
				Type:     "number",
				Required: false,
				Default:  42,
			},
		},
	}

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name: "valid params",
			params: map[string]interface{}{
				"required_param": "test",
				"optional_param": 123,
			},
			wantError: false,
		},
		{
			name: "missing required param",
			params: map[string]interface{}{
				"optional_param": 123,
			},
			wantError: true,
		},
		{
			name: "only required param",
			params: map[string]interface{}{
				"required_param": "test",
			},
			wantError: false,
		},
		{
			name:      "empty params",
			params:    map[string]interface{}{},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tool.Validate(tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestRegistry(t *testing.T) {
	registry := NewRegistry()

	// Create a test tool
	tool := &Tool{
		Name:        "test_tool",
		Description: "A test tool",
		Parameters:  map[string]Parameter{},
		Execute: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return "test result", nil
		},
	}

	// Test registration
	err := registry.Register(tool)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	// Test duplicate registration
	err = registry.Register(tool)
	if err == nil {
		t.Error("Register() should return error for duplicate tool")
	}

	// Test Get
	retrieved, err := registry.Get("test_tool")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if retrieved.Name != tool.Name {
		t.Errorf("Get() got name = %v, want %v", retrieved.Name, tool.Name)
	}

	// Test Exists
	if !registry.Exists("test_tool") {
		t.Error("Exists() should return true for existing tool")
	}
	if registry.Exists("nonexistent") {
		t.Error("Exists() should return false for nonexistent tool")
	}

	// Test List
	tools := registry.List()
	if len(tools) != 1 {
		t.Errorf("List() got count = %v, want %v", len(tools), 1)
	}

	// Test Count
	if registry.Count() != 1 {
		t.Errorf("Count() = %v, want %v", registry.Count(), 1)
	}

	// Test Unregister
	err = registry.Unregister("test_tool")
	if err != nil {
		t.Errorf("Unregister() error = %v", err)
	}
	if registry.Exists("test_tool") {
		t.Error("Tool should not exist after Unregister()")
	}
}

func TestExecutor(t *testing.T) {
	registry := NewRegistry()

	// Create a test tool
	tool := &Tool{
		Name:        "test_tool",
		Description: "A test tool",
		Parameters: map[string]Parameter{
			"value": {
				Type:     "string",
				Required: true,
			},
		},
		Execute: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return params["value"], nil
		},
	}

	registry.Register(tool)
	executor := NewExecutor(registry)

	t.Run("successful execution", func(t *testing.T) {
		result, err := executor.Execute(context.Background(), "test_tool", map[string]interface{}{
			"value": "test",
		})

		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if !result.Success {
			t.Error("Execute() result.Success = false, want true")
		}
		if result.Data != "test" {
			t.Errorf("Execute() result.Data = %v, want test", result.Data)
		}
	})

	t.Run("missing parameter", func(t *testing.T) {
		result, err := executor.Execute(context.Background(), "test_tool", map[string]interface{}{})

		if err == nil {
			t.Error("Execute() should return error for missing parameter")
		}
		if result.Success {
			t.Error("Execute() result.Success = true, want false")
		}
	})

	t.Run("tool not found", func(t *testing.T) {
		_, err := executor.Execute(context.Background(), "nonexistent", map[string]interface{}{})

		if err == nil {
			t.Error("Execute() should return error for nonexistent tool")
		}
	})

	t.Run("execution timeout", func(t *testing.T) {
		// Create a slow tool
		slowTool := &Tool{
			Name:        "slow_tool",
			Description: "A slow test tool",
			Parameters:  map[string]Parameter{},
			Execute: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				time.Sleep(5 * time.Second)
				return "done", nil
			},
		}

		registry.Register(slowTool)
		executor.SetTimeout(100 * time.Millisecond)

		result, err := executor.Execute(context.Background(), "slow_tool", map[string]interface{}{})

		if err == nil {
			t.Error("Execute() should return error for timeout")
		}
		if result.Success {
			t.Error("Execute() result.Success = true, want false")
		}
	})
}

func TestParseToolCall(t *testing.T) {
	registry := NewRegistry()

	// Register test tools
	registry.Register(&Tool{
		Name:        "read",
		Description: "Read a file",
		Parameters: map[string]Parameter{
			"path": {
				Type:     "string",
				Required: true,
			},
		},
		Execute: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return nil, nil
		},
	})

	executor := NewExecutor(registry)

	t.Run("JSON format", func(t *testing.T) {
		response := `{"tool": "read", "params": {"path": "/tmp/test.txt"}}`
		call, err := executor.ParseToolCall(response)

		if err != nil {
			t.Fatalf("ParseToolCall() error = %v", err)
		}
		if call.Name != "read" {
			t.Errorf("ParseToolCall() name = %v, want read", call.Name)
		}
		if call.Params["path"] != "/tmp/test.txt" {
			t.Errorf("ParseToolCall() path = %v, want /tmp/test.txt", call.Params["path"])
		}
	})

	t.Run("natural language format", func(t *testing.T) {
		response := "Use read tool with path /tmp/test.txt"
		call, err := executor.ParseToolCall(response)

		if err != nil {
			t.Fatalf("ParseToolCall() error = %v", err)
		}
		if call.Name != "read" {
			t.Errorf("ParseToolCall() name = %v, want read", call.Name)
		}
		// Note: Natural language parsing is simple and may not extract all params
	})

	t.Run("invalid format", func(t *testing.T) {
		response := "This is just a regular response without tool calls"
		_, err := executor.ParseToolCall(response)

		if err == nil {
			t.Error("ParseToolCall() should return error for invalid format")
		}
	})
}
