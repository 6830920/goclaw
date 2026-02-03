// Package config provides utilities for loading configuration from various sources
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

// LoadGlobalConfig attempts to load configuration from the global openclaw config
func LoadGlobalConfig() (*Config, error) {
	// Get user home directory
	usr, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("could not get user home directory: %w", err)
	}

	// Look for config in ~/.openclaw/openclaw.json
	globalConfigPath := filepath.Join(usr.HomeDir, ".openclaw", "openclaw.json")

	if _, err := os.Stat(globalConfigPath); err == nil {
		// Found global config, load it
		data, err := os.ReadFile(globalConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read global config: %w", err)
		}

		var config Config
		if err := json.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse global config: %w", err)
		}

		return &config, nil
	}

	return nil, fmt.Errorf("global config file not found: %s", globalConfigPath)
}

// MergeConfigs merges a global config with a local config, with local values taking precedence
func MergeConfigs(global, local *Config) *Config {
	if global == nil {
		return local
	}
	if local == nil {
		return global
	}

	// Create a new config based on global, then override with local values
	merged := *global

	// Override with local agent settings
	if local.Agent.Model != "" {
		merged.Agent.Model = local.Agent.Model
	}
	if local.Agent.Workspace != "" {
		merged.Agent.Workspace = local.Agent.Workspace
	}

	// Override with local gateway settings
	if local.Gateway.Port != 0 {
		merged.Gateway.Port = local.Gateway.Port
	}
	if local.Gateway.Bind != "" {
		merged.Gateway.Bind = local.Gateway.Bind
	}

	// Override with local Zhipu settings
	if local.Zhipu.ApiKey != "" {
		merged.Zhipu.ApiKey = local.Zhipu.ApiKey
	}
	if local.Zhipu.Model != "" {
		merged.Zhipu.Model = local.Zhipu.Model
	}
	if local.Zhipu.BaseURL != "" {
		merged.Zhipu.BaseURL = local.Zhipu.BaseURL
	}

	// For maps, merge them together (local takes precedence)
	if merged.Models == nil {
		merged.Models = make(map[string]interface{})
	}
	for k, v := range local.Models {
		merged.Models[k] = v
	}

	if merged.Channels == nil {
		merged.Channels = make(map[string]interface{})
	}
	for k, v := range local.Channels {
		merged.Channels[k] = v
	}

	return &merged
}
