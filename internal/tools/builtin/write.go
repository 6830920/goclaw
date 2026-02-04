package builtin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"goclaw/internal/tools"
)

// WriteTool writes content to a file
func WriteTool() *tools.Tool {
	return &tools.Tool{
		Name:        "write",
		Description: "Write content to a file. Creates the file if it doesn't exist, overwrites if it does. Automatically creates parent directories.",
		Parameters: map[string]tools.Parameter{
			"path": {
				Type:        "string",
				Description: "Path to the file to write (relative or absolute)",
				Required:    true,
			},
			"content": {
				Type:        "string",
				Description: "Content to write to the file",
				Required:    true,
			},
		},
		Execute: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			// Extract parameters
			path, ok := params["path"].(string)
			if !ok {
				return nil, fmt.Errorf("path parameter is required and must be a string")
			}

			content, ok := params["content"].(string)
			if !ok {
				return nil, fmt.Errorf("content parameter is required and must be a string")
			}

			// Create parent directories if needed
			dir := filepath.Dir(path)
			if dir != "." && dir != "" {
				if err := os.MkdirAll(dir, 0755); err != nil {
					return nil, fmt.Errorf("failed to create parent directories: %w", err)
				}
			}

			// Write file
			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				return nil, fmt.Errorf("failed to write file: %w", err)
			}

			return map[string]interface{}{
				"path":     path,
				"bytes":    len(content),
				"lines":    len(strings.Split(content, "\n")),
				"status":   "written",
			}, nil
		},
	}
}
