package builtin

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"

	"goclaw/internal/tools"
)

// ExecTool executes shell commands
func ExecTool() *tools.Tool {
	return &tools.Tool{
		Name:        "exec",
		Description: "Execute shell commands. Returns command output (stdout and stderr) and exit code. Use for system operations, running scripts, or any CLI interaction.",
		Parameters: map[string]tools.Parameter{
			"command": {
				Type:        "string",
				Description: "Shell command to execute",
				Required:    true,
			},
			"timeout": {
				Type:        "number",
				Description: "Timeout in seconds (optional)",
				Required:    false,
				Default:     30,
			},
			"workdir": {
				Type:        "string",
				Description: "Working directory (optional)",
				Required:    false,
			},
		},
		Execute: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			// Extract parameters
			command, ok := params["command"].(string)
			if !ok {
				return nil, fmt.Errorf("command parameter is required and must be a string")
			}

			// Get optional timeout
			timeout := 30 * time.Second
			if timeoutVal, exists := params["timeout"]; exists {
				switch v := timeoutVal.(type) {
				case float64:
					timeout = time.Duration(v) * time.Second
				case int:
					timeout = time.Duration(v) * time.Second
				case int64:
					timeout = time.Duration(v) * time.Second
				}
			}

			// Create context with timeout if not already set
			if _, hasDeadline := ctx.Deadline(); !hasDeadline {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, timeout)
				defer cancel()
			}

			// Create command
			cmd := exec.CommandContext(ctx, "sh", "-c", command)

			// Set working directory if provided
			if workdir, exists := params["workdir"]; exists {
				if dir, ok := workdir.(string); ok && dir != "" {
					cmd.Dir = dir
				}
			}

			// Capture output
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			// Execute command
			startTime := time.Now()
			err := cmd.Run()
			duration := time.Since(startTime)

			// Determine exit code
			exitCode := 0
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					exitCode = exitErr.ExitCode()
				} else {
					exitCode = -1
				}
			}

			return map[string]interface{}{
				"command":   command,
				"exitCode":  exitCode,
				"stdout":    stdout.String(),
				"stderr":    stderr.String(),
				"duration":  duration.String(),
				"timedOut":  ctx.Err() == context.DeadlineExceeded,
				"workdir":   cmd.Dir,
			}, nil
		},
	}
}
