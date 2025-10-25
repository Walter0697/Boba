package config

import (
	"fmt"
	"log"
)

// ExampleUsage demonstrates how to use the ConfigManager
func ExampleUsage() {
	// Create a new configuration manager
	cm := NewConfigManager()
	
	// Load existing configuration or create default
	if err := cm.LoadConfig(); err != nil {
		log.Printf("Error loading config: %v", err)
		return
	}
	
	// Load credentials
	if err := cm.LoadCredentials(); err != nil {
		log.Printf("Error loading credentials: %v", err)
		return
	}
	
	// Example: Set repository URL
	if err := cm.SetRepositoryURL("https://github.com/username/dev-setup"); err != nil {
		log.Printf("Error setting repository URL: %v", err)
		return
	}
	
	// Example: Configure tool overrides
	if err := cm.SetToolOverride("docker", false); err != nil {
		log.Printf("Error setting tool override: %v", err)
		return
	}
	
	if err := cm.SetToolOverride("git", true); err != nil {
		log.Printf("Error setting tool override: %v", err)
		return
	}
	
	// Example: Set GitHub token
	if err := cm.SetGitHubToken("ghp_example_token_123"); err != nil {
		log.Printf("Error setting GitHub token: %v", err)
		return
	}
	
	// Example: Check tool overrides
	if enabled, exists := cm.GetToolOverride("docker"); exists {
		fmt.Printf("Docker override: %t\n", enabled)
	} else {
		fmt.Println("No override set for docker")
	}
	
	// Example: Get current configuration
	config := cm.GetConfig()
	fmt.Printf("Repository URL: %s\n", config.RepositoryURL)
	fmt.Printf("Tool overrides: %v\n", config.ToolOverrides)
	
	// Example: Validate configuration
	if err := cm.ValidateConfig(); err != nil {
		log.Printf("Configuration validation failed: %v", err)
		return
	}
	
	// Example: Update last sync timestamp
	if err := cm.UpdateLastSync(); err != nil {
		log.Printf("Error updating last sync: %v", err)
		return
	}
	
	fmt.Println("Configuration management example completed successfully!")
}

// GetEnabledTools is a helper function that would be used by other components
// to filter tools based on local overrides
func (cm *ConfigManager) GetEnabledTools(allTools []string) []string {
	var enabledTools []string
	
	for _, tool := range allTools {
		if enabled, exists := cm.GetToolOverride(tool); exists {
			// If override exists, use the override value
			if enabled {
				enabledTools = append(enabledTools, tool)
			}
		} else {
			// If no override exists, include the tool by default
			enabledTools = append(enabledTools, tool)
		}
	}
	
	return enabledTools
}

// IsConfigured checks if the basic configuration is set up
func (cm *ConfigManager) IsConfigured() bool {
	config := cm.GetConfig()
	return config.RepositoryURL != ""
}

// HasGitHubToken checks if a GitHub token is configured
func (cm *ConfigManager) HasGitHubToken() bool {
	creds := cm.GetCredentials()
	return creds.GitHubToken != ""
}