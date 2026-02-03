// Package main provides a one-time configuration copy utility
// This tool copies configuration from ~/.openclaw/openclaw.json to the local config.json
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

func main() {
	fmt.Println("Goclaw Configuration Copy Tool")
	fmt.Println("=================================")

	// Get user home directory
	usr, err := user.Current()
	if err != nil {
		fmt.Printf("Error: Could not get user home directory: %v\n", err)
		os.Exit(1)
	}

	// Source: ~/.openclaw/openclaw.json
	sourcePath := filepath.Join(usr.HomeDir, ".openclaw", "openclaw.json")
	// Destination: ./config.json
	destPath := "config.json"

	// Check if source config exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		fmt.Printf("Error: Source config file does not exist: %s\n", sourcePath)
		os.Exit(1)
	}

	// Read source config
	sourceData, err := os.ReadFile(sourcePath)
	if err != nil {
		fmt.Printf("Error: Failed to read source config: %v\n", err)
		os.Exit(1)
	}

	// Parse source config to ensure it's valid
	var sourceConfig map[string]interface{}
	if err := json.Unmarshal(sourceData, &sourceConfig); err != nil {
		fmt.Printf("Error: Failed to parse source config: %v\n", err)
		os.Exit(1)
	}

	// Check if destination config already exists
	if _, err := os.Stat(destPath); err == nil {
		// Destination exists, ask for confirmation
		fmt.Printf("Warning: %s already exists. Overwrite? (y/N): ", destPath)
		var response string
		fmt.Scanf("%s", &response)
		if response != "y" && response != "Y" {
			fmt.Println("Operation cancelled.")
			os.Exit(0)
		}
	}

	// Write source config to destination
	if err := os.WriteFile(destPath, sourceData, 0644); err != nil {
		fmt.Printf("Error: Failed to write destination config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Success: Copied configuration from %s to %s\n", sourcePath, destPath)

	// Check for model providers in the copied config
	if providers, ok := sourceConfig["models"].(map[string]interface{})["providers"]; ok {
		if providerMap, ok := providers.(map[string]interface{}); ok {
			if len(providerMap) > 0 {
				fmt.Println("Detected model providers:")
				for provider := range providerMap {
					fmt.Printf("  - %s\n", provider)
				}
			}
		}
	}
}
