// Package config handles configuration for OpenClaw-Go
package config

import (
	"encoding/json"
	"os"
)

// Config represents the main configuration
type Config struct {
	Agent    AgentConfig            `json:"agent,omitempty"`
	Channels map[string]interface{} `json:"channels,omitempty"`
	Gateway  GatewayConfig          `json:"gateway,omitempty"`
	Models   map[string]interface{} `json:"models,omitempty"`
}

// AgentConfig holds agent-specific configuration
type AgentConfig struct {
	Model     string                 `json:"model,omitempty"`
	Workspace string                 `json:"workspace,omitempty"`
	Sandbox   SandboxConfig          `json:"sandbox,omitempty"`
	Defaults  AgentDefaults          `json:"defaults,omitempty"`
}

// AgentDefaults holds default agent settings
type AgentDefaults struct {
	ImageModel string `json:"imageModel,omitempty"`
	Workspace  string `json:"workspace,omitempty"`
}

// GatewayConfig holds gateway configuration
type GatewayConfig struct {
	Port         int                    `json:"port,omitempty"`
	Bind         string                 `json:"bind,omitempty"`
	Tailscale    TailscaleConfig        `json:"tailscale,omitempty"`
	Auth         AuthConfig             `json:"auth,omitempty"`
	Credentials  map[string]interface{} `json:"credentials,omitempty"`
}

// TailscaleConfig holds Tailscale-related configuration
type TailscaleConfig struct {
	Mode   string `json:"mode,omitempty"` // "off", "serve", "funnel"
	Reset  bool   `json:"resetOnExit,omitempty"`
	Domain string `json:"domain,omitempty"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Mode         string   `json:"mode,omitempty"` // "off", "password", "oauth"
	Password     string   `json:"password,omitempty"`
	AllowTailscale bool   `json:"allowTailscale,omitempty"`
	Users        []string `json:"users,omitempty"`
}

// SandboxConfig holds sandbox configuration
type SandboxConfig struct {
	Mode    string   `json:"mode,omitempty"` // "off", "non-main", "all"
	Allow   []string `json:"allow,omitempty"`
	Deny    []string `json:"deny,omitempty"`
	Timeout int      `json:"timeout,omitempty"`
}

// LoadConfig loads configuration from a JSON file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveConfig saves configuration to a JSON file
func SaveConfig(config *Config, path string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// NewDefaultConfig creates a new default configuration
func NewDefaultConfig() *Config {
	return &Config{
		Agent: AgentConfig{
			Model: "anthropic/claude-opus-4-5",
			Defaults: AgentDefaults{
				ImageModel: "openai/gpt-4-vision-preview",
				Workspace:  "~/.openclaw/workspace",
			},
		},
		Channels: make(map[string]interface{}),
		Gateway: GatewayConfig{
			Port: 18789,
			Bind: "127.0.0.1",
		},
		Models: make(map[string]interface{}),
	}
}