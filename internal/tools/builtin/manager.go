package builtin

import (
	"goclaw/internal/tools"
)

// Manager manages all builtin tools
type Manager struct {
	registry *tools.Registry
}

// NewManager creates a new builtin tools manager
func NewManager() *Manager {
	registry := tools.NewRegistry()
	manager := &Manager{
		registry: registry,
	}

	// Register all builtin tools
	manager.registerBuiltinTools()

	return manager
}

// GetRegistry returns the underlying tool registry
func (m *Manager) GetRegistry() *tools.Registry {
	return m.registry
}

// registerBuiltinTools registers all builtin tools
func (m *Manager) registerBuiltinTools() {
	// File operations
	m.registry.Register(ReadTool())
	m.registry.Register(WriteTool())

	// System operations
	m.registry.Register(ExecTool())

	// Note: More tools will be added here as they are implemented:
	// - web_search
	// - web_fetch
	// - memory_search
	// - browser control
	// - messaging
	// - etc.
}

// GetAllTools returns all builtin tools
func (m *Manager) GetAllTools() []*tools.Tool {
	return m.registry.List()
}

// GetToolCount returns the number of builtin tools
func (m *Manager) GetToolCount() int {
	return m.registry.Count()
}
