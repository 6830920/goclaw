package tools

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

// Executor 执行器接口
type Executor interface {
	ExecuteCommand(ctx context.Context, command string, args []string) (ExecutionResult, error)
	ReadFile(filename string) (string, error)
	WriteFile(filename, content string) error
	AppendToFile(filename, content string) error
	FileExists(filename string) bool
}

// ExecutionResult 执行结果
type ExecutionResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Duration time.Duration
	Error    error
}

// SystemExecutor 系统执行器
type SystemExecutor struct {
	Timeout time.Duration
}

// NewSystemExecutor 创建系统执行器
func NewSystemExecutor(timeout time.Duration) *SystemExecutor {
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	return &SystemExecutor{
		Timeout: timeout,
	}
}

// ExecuteCommand 执行命令
func (se *SystemExecutor) ExecuteCommand(ctx context.Context, command string, args []string) (ExecutionResult, error) {
	start := time.Now()
	
	ctx, cancel := context.WithTimeout(ctx, se.Timeout)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, command, args...)
	
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	
	err := cmd.Run()
	
	result := ExecutionResult{
		Stdout:   stdoutBuf.String(),
		Stderr:   stderrBuf.String(),
		Duration: time.Since(start),
	}
	
	if err != nil {
		result.Error = err
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			result.ExitCode = 1
		}
	} else {
		result.ExitCode = 0
	}
	
	return result, nil
}

// ReadFile 读取文件
func (se *SystemExecutor) ReadFile(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WriteFile 写入文件
func (se *SystemExecutor) WriteFile(filename, content string) error {
	return os.WriteFile(filename, []byte(content), 0644)
}

// AppendToFile 追加到文件
func (se *SystemExecutor) AppendToFile(filename, content string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	
	_, err = file.WriteString(content)
	return err
}

// FileExists 检查文件是否存在
func (se *SystemExecutor) FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// StreamCommand 流式执行命令
func (se *SystemExecutor) StreamCommand(ctx context.Context, command string, args []string, outputChan chan<- string) error {
	ctx, cancel := context.WithTimeout(ctx, se.Timeout)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, command, args...)
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	
	if err := cmd.Start(); err != nil {
		return err
	}
	
	// 读取输出
	go func() {
		reader := io.MultiReader(stdout, stderr)
		buffer := make([]byte, 1024)
		
		for {
			select {
			case <-ctx.Done():
				return
			default:
				n, err := reader.Read(buffer)
				if n > 0 {
					outputChan <- string(buffer[:n])
				}
				if err != nil {
					if err != io.EOF {
						outputChan <- fmt.Sprintf("Error reading output: %v", err)
					}
					return
				}
			}
		}
	}()
	
	return cmd.Wait()
}