package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// InstalledTool represents a tool that has been installed
type InstalledTool struct {
	Name            string    `json:"name"`
	Version         string    `json:"version"`
	InstallDate     time.Time `json:"install_date"`
	LastUpdateDate  time.Time `json:"last_update_date,omitempty"`
	InstallMethod   string    `json:"install_method"` // "auto" or "manual"
}

// Config represents the main configuration structure
type Config struct {
	RepositoryURL        string                    `json:"repository_url"`
	ToolOverrides        map[string]bool           `json:"tool_overrides"`
	EnvironmentOverrides map[string]bool           `json:"environment_overrides"`
	InstalledTools       map[string]InstalledTool  `json:"installed_tools"`
	LastSync             time.Time                 `json:"last_sync"`
}

// Credentials stores sensitive authentication information separately
type Credentials struct {
	GitHubToken string `json:"github_token"`
}

// ConfigManager handles configuration file operations
type ConfigManager struct {
	configDir     string
	configPath    string
	credPath      string
	config        *Config
	credentials   *Credentials
}

// NewConfigManager creates a new configuration manager
func NewConfigManager() *ConfigManager {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home directory is not accessible
		homeDir = "."
	}
	
	// Docker container detection and permission handling
	configDir := getConfigDir(homeDir)
	
	return &ConfigManager{
		configDir:   configDir,
		configPath:  filepath.Join(configDir, "config.json"),
		credPath:    filepath.Join(configDir, "credentials.json"),
		config:      &Config{},
		credentials: &Credentials{},
	}
}

// InitConfigDir creates the configuration directory if it doesn't exist
func (cm *ConfigManager) InitConfigDir() error {
	if err := os.MkdirAll(cm.configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	// Create cache subdirectory for future use
	cacheDir := filepath.Join(cm.configDir, "cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}
	
	return nil
}

// LoadConfig loads the configuration from the config file
func (cm *ConfigManager) LoadConfig() error {
	// Initialize config directory if it doesn't exist
	if err := cm.InitConfigDir(); err != nil {
		return err
	}
	
	// Check if config file exists
	if _, err := os.Stat(cm.configPath); os.IsNotExist(err) {
		// Create default config if file doesn't exist
		cm.config = &Config{
			RepositoryURL:        "",
			ToolOverrides:        make(map[string]bool),
			EnvironmentOverrides: make(map[string]bool),
			InstalledTools:       make(map[string]InstalledTool),
			LastSync:             time.Time{},
		}
		return cm.SaveConfig()
	}
	
	// Read and parse config file
	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	
	if err := json.Unmarshal(data, cm.config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}
	
	// Initialize ToolOverrides map if it's nil
	if cm.config.ToolOverrides == nil {
		cm.config.ToolOverrides = make(map[string]bool)
	}
	
	// Initialize EnvironmentOverrides map if it's nil
	if cm.config.EnvironmentOverrides == nil {
		cm.config.EnvironmentOverrides = make(map[string]bool)
	}
	
	// Initialize InstalledTools map if it's nil
	if cm.config.InstalledTools == nil {
		cm.config.InstalledTools = make(map[string]InstalledTool)
	}
	
	return nil
}

// SaveConfig saves the current configuration to the config file
func (cm *ConfigManager) SaveConfig() error {
	if err := cm.InitConfigDir(); err != nil {
		return err
	}
	
	data, err := json.MarshalIndent(cm.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	if err := os.WriteFile(cm.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	return nil
}

// LoadCredentials loads credentials from the credentials file
func (cm *ConfigManager) LoadCredentials() error {
	// Initialize config directory if it doesn't exist
	if err := cm.InitConfigDir(); err != nil {
		return err
	}
	
	// Check if credentials file exists
	if _, err := os.Stat(cm.credPath); os.IsNotExist(err) {
		// Create empty credentials if file doesn't exist
		cm.credentials = &Credentials{}
		return nil
	}
	
	// Read and parse credentials file
	data, err := os.ReadFile(cm.credPath)
	if err != nil {
		return fmt.Errorf("failed to read credentials file: %w", err)
	}
	
	if err := json.Unmarshal(data, cm.credentials); err != nil {
		return fmt.Errorf("failed to parse credentials file: %w", err)
	}
	
	return nil
}

// SaveCredentials saves credentials to the credentials file with restricted permissions
func (cm *ConfigManager) SaveCredentials() error {
	if err := cm.InitConfigDir(); err != nil {
		return err
	}
	
	data, err := json.MarshalIndent(cm.credentials, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}
	
	// Debug: Print where we're trying to save credentials
	fmt.Printf("DEBUG: Saving credentials to: %s\n", cm.credPath)
	
	// Write with restricted permissions (600) for security
	if err := os.WriteFile(cm.credPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write credentials file: %w", err)
	}
	
	fmt.Printf("DEBUG: Credentials saved successfully\n")
	return nil
}

// GetConfig returns a copy of the current configuration
func (cm *ConfigManager) GetConfig() Config {
	if cm.config == nil {
		return Config{
			ToolOverrides: make(map[string]bool),
		}
	}
	
	// Return a copy to prevent external modification
	configCopy := *cm.config
	if configCopy.ToolOverrides == nil {
		configCopy.ToolOverrides = make(map[string]bool)
	}
	
	return configCopy
}

// SetRepositoryURL sets the repository URL in the configuration
func (cm *ConfigManager) SetRepositoryURL(url string) error {
	if cm.config == nil {
		cm.config = &Config{
			ToolOverrides: make(map[string]bool),
		}
	}
	
	cm.config.RepositoryURL = url
	return cm.SaveConfig()
}

// GetToolOverride returns the override setting for a specific tool
// Returns (enabled, exists) where exists indicates if an override is set
func (cm *ConfigManager) GetToolOverride(toolName string) (bool, bool) {
	if cm.config == nil || cm.config.ToolOverrides == nil {
		return false, false
	}
	
	enabled, exists := cm.config.ToolOverrides[toolName]
	return enabled, exists
}

// SetToolOverride sets the override setting for a specific tool
func (cm *ConfigManager) SetToolOverride(toolName string, enabled bool) error {
	if cm.config == nil {
		cm.config = &Config{
			ToolOverrides: make(map[string]bool),
		}
	}
	
	if cm.config.ToolOverrides == nil {
		cm.config.ToolOverrides = make(map[string]bool)
	}
	
	cm.config.ToolOverrides[toolName] = enabled
	return cm.SaveConfig()
}

// RemoveToolOverride removes the override setting for a specific tool
func (cm *ConfigManager) RemoveToolOverride(toolName string) error {
	if cm.config == nil || cm.config.ToolOverrides == nil {
		return nil // Nothing to remove
	}
	
	delete(cm.config.ToolOverrides, toolName)
	return cm.SaveConfig()
}

// UpdateLastSync updates the last synchronization timestamp
func (cm *ConfigManager) UpdateLastSync() error {
	if cm.config == nil {
		cm.config = &Config{
			ToolOverrides: make(map[string]bool),
		}
	}
	
	cm.config.LastSync = time.Now()
	return cm.SaveConfig()
}

// GetCredentials returns a copy of the current credentials
func (cm *ConfigManager) GetCredentials() Credentials {
	if cm.credentials == nil {
		return Credentials{}
	}
	
	// Return a copy to prevent external modification
	return *cm.credentials
}

// SetGitHubToken sets the GitHub token in credentials
func (cm *ConfigManager) SetGitHubToken(token string) error {
	if cm.credentials == nil {
		cm.credentials = &Credentials{}
	}
	
	cm.credentials.GitHubToken = token
	return cm.SaveCredentials()
}

// ValidateConfig validates the current configuration
func (cm *ConfigManager) ValidateConfig() error {
	if cm.config == nil {
		return fmt.Errorf("configuration is not loaded")
	}
	
	// Validate repository URL format if it's set
	if cm.config.RepositoryURL != "" {
		// Basic validation - check if it looks like a GitHub URL
		if len(cm.config.RepositoryURL) < 10 {
			return fmt.Errorf("repository URL appears to be too short")
		}
	}
	
	// Validate tool overrides
	if cm.config.ToolOverrides == nil {
		cm.config.ToolOverrides = make(map[string]bool)
	}
	
	return nil
}

// GetConfigDir returns the configuration directory path
func (cm *ConfigManager) GetConfigDir() string {
	return cm.configDir
}

// GetConfigPath returns the configuration file path
func (cm *ConfigManager) GetConfigPath() string {
	return cm.configPath
}

// GetCredentialsPath returns the credentials file path
func (cm *ConfigManager) GetCredentialsPath() string {
	return cm.credPath
}

// RecordToolInstallation records that a tool has been installed
func (cm *ConfigManager) RecordToolInstallation(name, version, method string) error {
	if cm.config == nil {
		cm.config = &Config{
			ToolOverrides:  make(map[string]bool),
			InstalledTools: make(map[string]InstalledTool),
		}
	}
	
	if cm.config.InstalledTools == nil {
		cm.config.InstalledTools = make(map[string]InstalledTool)
	}
	
	now := time.Now()
	
	// Check if tool is already recorded
	if existing, exists := cm.config.InstalledTools[name]; exists {
		// Update existing record
		existing.Version = version
		existing.LastUpdateDate = now
		existing.InstallMethod = method
		cm.config.InstalledTools[name] = existing
	} else {
		// Create new record
		cm.config.InstalledTools[name] = InstalledTool{
			Name:          name,
			Version:       version,
			InstallDate:   now,
			InstallMethod: method,
		}
	}
	
	return cm.SaveConfig()
}

// GetInstalledTool returns information about an installed tool
func (cm *ConfigManager) GetInstalledTool(name string) (InstalledTool, bool) {
	if cm.config == nil || cm.config.InstalledTools == nil {
		return InstalledTool{}, false
	}
	
	tool, exists := cm.config.InstalledTools[name]
	return tool, exists
}

// GetAllInstalledTools returns all installed tools
func (cm *ConfigManager) GetAllInstalledTools() map[string]InstalledTool {
	if cm.config == nil || cm.config.InstalledTools == nil {
		return make(map[string]InstalledTool)
	}
	
	// Return a copy to prevent external modification
	result := make(map[string]InstalledTool)
	for k, v := range cm.config.InstalledTools {
		result[k] = v
	}
	return result
}

// RemoveInstalledTool removes a tool from the installation history
func (cm *ConfigManager) RemoveInstalledTool(name string) error {
	if cm.config == nil || cm.config.InstalledTools == nil {
		return nil // Nothing to remove
	}
	
	delete(cm.config.InstalledTools, name)
	return cm.SaveConfig()
}

// GetEnvironmentOverride returns the override setting for a specific environment
// Returns (enabled, exists) where exists indicates if an override is set
func (cm *ConfigManager) GetEnvironmentOverride(envName string) (bool, bool) {
	if cm.config == nil || cm.config.EnvironmentOverrides == nil {
		return false, false
	}
	
	enabled, exists := cm.config.EnvironmentOverrides[envName]
	return enabled, exists
}

// SetEnvironmentOverride sets the override setting for a specific environment
func (cm *ConfigManager) SetEnvironmentOverride(envName string, enabled bool) error {
	if cm.config == nil {
		cm.config = &Config{
			ToolOverrides:        make(map[string]bool),
			EnvironmentOverrides: make(map[string]bool),
			InstalledTools:       make(map[string]InstalledTool),
		}
	}
	
	if cm.config.EnvironmentOverrides == nil {
		cm.config.EnvironmentOverrides = make(map[string]bool)
	}
	
	cm.config.EnvironmentOverrides[envName] = enabled
	return cm.SaveConfig()
}

// RemoveEnvironmentOverride removes the override setting for a specific environment
func (cm *ConfigManager) RemoveEnvironmentOverride(envName string) error {
	if cm.config == nil || cm.config.EnvironmentOverrides == nil {
		return nil // Nothing to remove
	}
	
	delete(cm.config.EnvironmentOverrides, envName)
	return cm.SaveConfig()
}

// ResetAllToolOverrides removes all tool overrides, returning to defaults
func (cm *ConfigManager) ResetAllToolOverrides() error {
	if cm.config == nil {
		return nil
	}
	
	cm.config.ToolOverrides = make(map[string]bool)
	return cm.SaveConfig()
}

// ResetAllEnvironmentOverrides removes all environment overrides, returning to defaults
func (cm *ConfigManager) ResetAllEnvironmentOverrides() error {
	if cm.config == nil {
		return nil
	}
	
	cm.config.EnvironmentOverrides = make(map[string]bool)
	return cm.SaveConfig()
}

// getConfigDir determines the best config directory based on environment
func getConfigDir(homeDir string) string {
	// Check if we're in a Docker container
	if isDockerContainer() {
		// In Docker, try /tmp first as it's always writable
		tmpConfig := "/tmp/.boba"
		if canCreateDir(tmpConfig) {
			return tmpConfig
		}
	}
	
	// Try the normal home directory
	normalConfig := filepath.Join(homeDir, ".boba")
	if canCreateDir(normalConfig) {
		return normalConfig
	}
	
	// Fallback to current directory
	currentConfig := ".boba"
	if canCreateDir(currentConfig) {
		return currentConfig
	}
	
	// Last resort - use /tmp
	return "/tmp/.boba"
}

// isDockerContainer detects if we're running inside a Docker container
func isDockerContainer() bool {
	// Check for .dockerenv file (most reliable)
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	
	// Check cgroup for docker
	if data, err := os.ReadFile("/proc/1/cgroup"); err == nil {
		content := string(data)
		if strings.Contains(content, "docker") || strings.Contains(content, "containerd") {
			return true
		}
	}
	
	// Check if we're running as PID 1 (common in containers)
	if os.Getpid() == 1 {
		return true
	}
	
	return false
}

// canCreateDir tests if we can create a directory at the given path
func canCreateDir(path string) bool {
	// Try to create the directory
	if err := os.MkdirAll(path, 0755); err != nil {
		return false
	}
	
	// Try to create a test file to verify write permissions
	testFile := filepath.Join(path, ".test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return false
	}
	
	// Clean up test file
	os.Remove(testFile)
	return true
}