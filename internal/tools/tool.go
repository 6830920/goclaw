package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Tool represents a callable tool that AI can use
type Tool struct {
	Name        string                 // Tool name (unique identifier)
	Description string                 // Tool description for AI
	Parameters  map[string]Parameter   // Parameter definitions
	Execute     ToolExecuteFunc        // Execution function
}

// Parameter defines a tool parameter
type Parameter struct {
	Type        string      // Parameter type: string, number, boolean, array, object
	Description string      // Parameter description
	Required    bool        // Whether the parameter is required
	Default     interface{} // Default value
}

// ToolExecuteFunc is the function signature for tool execution
type ToolExecuteFunc func(ctx context.Context, params map[string]interface{}) (interface{}, error)

// ToolCall represents a single tool call request
type ToolCall struct {
	Name   string                 `json:"name"`
	Params map[string]interface{} `json:"params"`
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Validate validates parameters against the tool's parameter definitions
func (t *Tool) Validate(params map[string]interface{}) error {
	// Check required parameters
	for paramName, paramDef := range t.Parameters {
		if paramDef.Required {
			if _, exists := params[paramName]; !exists {
				return fmt.Errorf("missing required parameter: %s", paramName)
			}
		}
	}

	// Type validation
	for paramName, paramValue := range params {
		if paramDef, exists := t.Parameters[paramName]; exists {
			if err := validateType(paramName, paramValue, paramDef.Type); err != nil {
				return err
			}
		}
	}

	return nil
}

// validateType checks if a value matches the expected type
func validateType(paramName string, value interface{}, expectedType string) error {
	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("parameter %s must be a string, got %T", paramName, value)
		}
	case "number":
		// Check for int, float32, float64
		switch value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			// OK
		default:
			return fmt.Errorf("parameter %s must be a number, got %T", paramName, value)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("parameter %s must be a boolean, got %T", paramName, value)
		}
	case "array":
		if _, ok := value.([]interface{}); !ok {
			return fmt.Errorf("parameter %s must be an array, got %T", paramName, value)
		}
	case "object":
		if _, ok := value.(map[string]interface{}); !ok {
			return fmt.Errorf("parameter %s must be an object, got %T", paramName, value)
		}
	default:
		return fmt.Errorf("unknown parameter type: %s", expectedType)
	}

	return nil
}

// ToJSON converts the tool to JSON representation
func (t *Tool) ToJSON() (string, error) {
	data := map[string]interface{}{
		"name":        t.Name,
		"description": t.Description,
		"parameters":  t.Parameters,
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// ToMarkdown converts the tool to Markdown representation for AI
func (t *Tool) ToMarkdown() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("## Tool: %s\n\n", t.Name))
	sb.WriteString(fmt.Sprintf("**Description:** %s\n\n", t.Description))
	sb.WriteString("**Parameters:**\n\n")

	if len(t.Parameters) == 0 {
		sb.WriteString("No parameters required.\n\n")
	} else {
		sb.WriteString("| Parameter | Type | Required | Description |\n")
		sb.WriteString("|-----------|------|----------|-------------|\n")

		for name, param := range t.Parameters {
			required := "No"
			if param.Required {
				required = "Yes"
			}

			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
				name, param.Type, required, param.Description))
		}

		sb.WriteString("\n")
	}

	return sb.String()
}
