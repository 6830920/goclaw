package tools

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// Registry manages a collection of tools
type Registry struct {
	tools map[string]*Tool
	mu    sync.RWMutex
}

// NewRegistry creates a new tool registry
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]*Tool),
	}
}

// Register adds a tool to the registry
func (r *Registry) Register(tool *Tool) error {
	if tool == nil {
		return fmt.Errorf("tool cannot be nil")
	}

	if tool.Name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}

	if tool.Execute == nil {
		return fmt.Errorf("tool execute function cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if tool already exists
	if _, exists := r.tools[tool.Name]; exists {
		return fmt.Errorf("tool '%s' is already registered", tool.Name)
	}

	r.tools[tool.Name] = tool
	return nil
}

// Unregister removes a tool from the registry
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tools[name]; !exists {
		return fmt.Errorf("tool '%s' is not registered", name)
	}

	delete(r.tools, name)
	return nil
}

// Get retrieves a tool by name
func (r *Registry) Get(name string) (*Tool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, exists := r.tools[name]
	if !exists {
		return nil, fmt.Errorf("tool '%s' not found", name)
	}

	return tool, nil
}

// Exists checks if a tool is registered
func (r *Registry) Exists(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.tools[name]
	return exists
}

// List returns all registered tools
func (r *Registry) List() []*Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]*Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}

	return tools
}

// Count returns the number of registered tools
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.tools)
}

// Clear removes all tools from the registry
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tools = make(map[string]*Tool)
}

// ToMarkdown formats all tools as Markdown for AI consumption
func (r *Registry) ToMarkdown() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.tools) == 0 {
		return "# Available Tools\n\nNo tools are available.\n"
	}

	var sb strings.Builder

	sb.WriteString("# Available Tools\n\n")
	sb.WriteString("You have access to the following tools. Use them when they can help with the user's request:\n\n")

	for _, tool := range r.tools {
		sb.WriteString(tool.ToMarkdown())
		sb.WriteString("---\n\n")
	}

	return sb.String()
}

// ToJSON formats all tools as JSON
func (r *Registry) ToJSON() (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]map[string]interface{}, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
			"parameters":  tool.Parameters,
		})
	}

	jsonBytes, err := json.MarshalIndent(tools, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// FormatForAI returns a formatted string suitable for inclusion in AI prompts
func (r *Registry) FormatForAI() string {
	return r.ToMarkdown()
}

// GetParameterNames returns all parameter names for a tool
func (r *Registry) GetParameterNames(toolName string) ([]string, error) {
	tool, err := r.Get(toolName)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(tool.Parameters))
	for name := range tool.Parameters {
		names = append(names, name)
	}

	return names, nil
}
