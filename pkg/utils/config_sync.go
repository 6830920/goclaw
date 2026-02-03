// Package utils provides utility functions for Goclaw
package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

// SyncGlobalConfig copies configuration from ~/.openclaw/openclaw.json to local config
func SyncGlobalConfig(localConfigPath string) error {
	// Get user home directory
	usr, err := user.Current()
	if err != nil {
		return fmt.Errorf("could not get user home directory: %w", err)
	}

	// Source: ~/.openclaw/openclaw.json
	globalConfigPath := filepath.Join(usr.HomeDir, ".openclaw", "openclaw.json")

	// Check if global config exists
	if _, err := os.Stat(globalConfigPath); os.IsNotExist(err) {
		return fmt.Errorf("global config file does not exist: %s", globalConfigPath)
	}

	// Read global config
	globalData, err := os.ReadFile(globalConfigPath)
	if err != nil {
		return fmt.Errorf("failed to read global config: %w", err)
	}

	// Parse global config
	var globalConfig map[string]interface{}
	if err := json.Unmarshal(globalData, &globalConfig); err != nil {
		return fmt.Errorf("failed to parse global config: %w", err)
	}

	// Check if local config exists, if not create it
	var localConfig map[string]interface{}
	if _, err := os.Stat(localConfigPath); err == nil {
		// Read existing local config
		localData, err := os.ReadFile(localConfigPath)
		if err != nil {
			return fmt.Errorf("failed to read local config: %w", err)
		}

		if err := json.Unmarshal(localData, &localConfig); err != nil {
			return fmt.Errorf("failed to parse local config: %w", err)
		}
	} else {
		// Create new local config
		localConfig = make(map[string]interface{})
	}

	// Copy Zhipu configuration from global to local
	if zhipuConfig, ok := globalConfig["zhipu"]; ok {
		localConfig["zhipu"] = zhipuConfig
		fmt.Println("Copied Zhipu configuration from global config")
	}

	// Copy other relevant configurations
	if agentConfig, ok := globalConfig["agent"]; ok {
		localConfig["agent"] = agentConfig
	}
	if modelsConfig, ok := globalConfig["models"]; ok {
		localConfig["models"] = modelsConfig
	}

	// Write updated local config
	localData, err := json.MarshalIndent(localConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal local config: %w", err)
	}

	if err := os.WriteFile(localConfigPath, localData, 0644); err != nil {
		return fmt.Errorf("failed to write local config: %w", err)
	}

	fmt.Printf("Successfully synchronized config to %s\n", localConfigPath)
	return nil
}
