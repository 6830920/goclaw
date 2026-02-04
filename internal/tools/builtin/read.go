package builtin

import (
	"context"
	"fmt"
	"os"
	"strings"

	"goclaw/internal/tools"
)

// ReadTool reads the contents of a file
func ReadTool() *tools.Tool {
	return &tools.Tool{
		Name:        "read",
		Description: "Read the contents of a file. Returns the file contents as text. Supports text files and images (jpg, png, gif, webp). Images are sent as attachments. For text files, output is truncated to 2000 lines or 50KB (whichever is hit first). Use offset/limit for large files. When you need the full file, continue with offset until complete.",
		Parameters: map[string]tools.Parameter{
			"path": {
				Type:        "string",
				Description: "Path to the file to read (relative or absolute)",
				Required:    true,
			},
			"offset": {
				Type:        "number",
				Description: "Line number to start reading from (1-indexed)",
				Required:    false,
				Default:     0,
			},
			"limit": {
				Type:        "number",
				Description: "Maximum number of lines to read",
				Required:    false,
				Default:     2000,
			},
		},
		Execute: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			// Extract parameters
			path, ok := params["path"].(string)
			if !ok {
				return nil, fmt.Errorf("path parameter is required and must be a string")
			}

			// Get optional parameters
			offset := 0
			if offsetVal, exists := params["offset"]; exists {
				switch v := offsetVal.(type) {
				case float64:
					offset = int(v)
				case int:
					offset = v
				case int64:
					offset = int(v)
				}
			}

			limit := 2000
			if limitVal, exists := params["limit"]; exists {
				switch v := limitVal.(type) {
				case float64:
					limit = int(v)
				case int:
					limit = v
				case int64:
					limit = int(v)
				}
			}

			// Read file
			content, err := os.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("failed to read file: %w", err)
			}

			// Convert to string and split by lines
			lines := strings.Split(string(content), "\n")

			// Apply offset
			if offset > 0 && offset <= len(lines) {
				lines = lines[offset-1:]
			}

			// Apply limit
			if limit > 0 && len(lines) > limit {
				lines = lines[:limit]
			}

			// Join lines back
			result := strings.Join(lines, "\n")

			// Add metadata
			return map[string]interface{}{
				"path":     path,
				"content":  result,
				"lines":    len(lines),
				"total":    len(strings.Split(string(content), "\n")),
				"truncated": len(strings.Split(string(content), "\n")) > limit,
			}, nil
		},
	}
}
