// Package tools implements various utility tools for Goclaw
// Similar to the tools available in the original OpenClaw
package tools

import (
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// FileReader provides functionality to read files
type FileReader struct{}

// ReadFile reads the content of a file
func (fr *FileReader) ReadFile(filePath string) (string, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	
	return string(content), nil
}

// FileWriter provides functionality to write files
type FileWriter struct{}

// WriteFile writes content to a file
func (fw *FileWriter) WriteFile(filePath, content string) error {
	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	return ioutil.WriteFile(filePath, []byte(content), 0644)
}

// Executor provides functionality to execute commands
type Executor struct{}

// Execute executes a shell command
func (ex *Executor) Execute(ctx context.Context, command string) (string, error) {
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", command)
	output, err := cmd.CombinedOutput()
	
	return string(output), err
}

// FileEditor provides functionality to edit files
type FileEditor struct{}

// ReplaceInFile replaces old text with new text in a file
func (fe *FileEditor) ReplaceInFile(filePath, oldText, newText string) error {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	
	newContent := strings.ReplaceAll(string(content), oldText, newText)
	return ioutil.WriteFile(filePath, []byte(newContent), 0644)
}

// FileSystem provides general filesystem operations
type FileSystem struct{}

// ListFiles lists files in a directory
func (fs *FileSystem) ListFiles(dirPath string) ([]string, error) {
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}
	
	var fileNames []string
	for _, file := range files {
		fileNames = append(fileNames, file.Name())
	}
	
	return fileNames, nil
}

// FileExists checks if a file exists
func (fs *FileSystem) FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}