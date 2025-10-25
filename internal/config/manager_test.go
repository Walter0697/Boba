package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewConfigManager(t *testing.T) {
	cm := NewConfigManager()
	
	if cm == nil {
		t.Fatal("NewConfigManager returned nil")
	}
	
	if cm.configDir == "" {
		t.Error("configDir should not be empty")
	}
	
	if cm.configPath == "" {
		t.Error("configPath should not be empty")
	}
	
	if cm.credPath == "" {
		t.Error("credPath should not be empty")
	}
}

func TestInitConfigDir(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	cm := &ConfigManager{
		configDir:  filepath.Join(tempDir, ".boba"),
		configPath: filepath.Join(tempDir, ".boba", "config.json"),
		credPath:   filepath.Join(tempDir, ".boba", "credentials.json"),
		config:     &Config{},
		credentials: &Credentials{},
	}
	
	err := cm.InitConfigDir()
	if err != nil {
		t.Fatalf("InitConfigDir failed: %v", err)
	}
	
	// Check if config directory was created
	if _, err := os.Stat(cm.configDir); os.IsNotExist(err) {
		t.Error("Config directory was not created")
	}
	
	// Check if cache directory was created
	cacheDir := filepath.Join(cm.configDir, "cache")
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		t.Error("Cache directory was not created")
	}
}

func TestLoadConfigNewFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	cm := &ConfigManager{
		configDir:  filepath.Join(tempDir, ".boba"),
		configPath: filepath.Join(tempDir, ".boba", "config.json"),
		credPath:   filepath.Join(tempDir, ".boba", "credentials.json"),
		config:     &Config{},
		credentials: &Credentials{},
	}
	
	err := cm.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	
	// Check if default config was created
	if cm.config.ToolOverrides == nil {
		t.Error("ToolOverrides should be initialized")
	}
	
	// Check if config file was created
	if _, err := os.Stat(cm.configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}
}

func TestLoadConfigExistingFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".boba")
	configPath := filepath.Join(configDir, "config.json")
	
	// Create config directory
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}
	
	// Create a test config file
	testConfig := Config{
		RepositoryURL: "https://github.com/test/repo",
		ToolOverrides: map[string]bool{
			"git":    true,
			"docker": false,
		},
		LastSync: time.Now(),
	}
	
	data, err := json.MarshalIndent(testConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test config: %v", err)
	}
	
	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}
	
	// Test loading the config
	cm := &ConfigManager{
		configDir:  configDir,
		configPath: configPath,
		credPath:   filepath.Join(configDir, "credentials.json"),
		config:     &Config{},
		credentials: &Credentials{},
	}
	
	err = cm.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	
	// Verify loaded config
	if cm.config.RepositoryURL != testConfig.RepositoryURL {
		t.Errorf("Expected RepositoryURL %s, got %s", testConfig.RepositoryURL, cm.config.RepositoryURL)
	}
	
	if len(cm.config.ToolOverrides) != 2 {
		t.Errorf("Expected 2 tool overrides, got %d", len(cm.config.ToolOverrides))
	}
	
	if cm.config.ToolOverrides["git"] != true {
		t.Error("Expected git override to be true")
	}
	
	if cm.config.ToolOverrides["docker"] != false {
		t.Error("Expected docker override to be false")
	}
}

func TestSaveConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	cm := &ConfigManager{
		configDir:  filepath.Join(tempDir, ".boba"),
		configPath: filepath.Join(tempDir, ".boba", "config.json"),
		credPath:   filepath.Join(tempDir, ".boba", "credentials.json"),
		config: &Config{
			RepositoryURL: "https://github.com/test/repo",
			ToolOverrides: map[string]bool{
				"git": true,
			},
			LastSync: time.Now(),
		},
		credentials: &Credentials{},
	}
	
	err := cm.SaveConfig()
	if err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}
	
	// Verify file was created
	if _, err := os.Stat(cm.configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}
	
	// Verify file content
	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}
	
	var savedConfig Config
	err = json.Unmarshal(data, &savedConfig)
	if err != nil {
		t.Fatalf("Failed to unmarshal saved config: %v", err)
	}
	
	if savedConfig.RepositoryURL != cm.config.RepositoryURL {
		t.Error("Saved config does not match original")
	}
}

func TestLoadCredentials(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	cm := &ConfigManager{
		configDir:  filepath.Join(tempDir, ".boba"),
		configPath: filepath.Join(tempDir, ".boba", "config.json"),
		credPath:   filepath.Join(tempDir, ".boba", "credentials.json"),
		config:     &Config{},
		credentials: &Credentials{},
	}
	
	// Test loading non-existent credentials
	err := cm.LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials failed: %v", err)
	}
	
	// Should create empty credentials
	if cm.credentials.GitHubToken != "" {
		t.Error("Expected empty GitHub token for new credentials")
	}
}

func TestSaveCredentials(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	cm := &ConfigManager{
		configDir:  filepath.Join(tempDir, ".boba"),
		configPath: filepath.Join(tempDir, ".boba", "config.json"),
		credPath:   filepath.Join(tempDir, ".boba", "credentials.json"),
		config:     &Config{},
		credentials: &Credentials{
			GitHubToken: "test-token-123",
		},
	}
	
	err := cm.SaveCredentials()
	if err != nil {
		t.Fatalf("SaveCredentials failed: %v", err)
	}
	
	// Verify file was created with correct permissions
	info, err := os.Stat(cm.credPath)
	if err != nil {
		t.Fatalf("Credentials file was not created: %v", err)
	}
	
	// Check file permissions (should be 0600)
	if info.Mode().Perm() != 0600 {
		t.Errorf("Expected file permissions 0600, got %o", info.Mode().Perm())
	}
}

func TestToolOverrides(t *testing.T) {
	tempDir := t.TempDir()
	
	cm := &ConfigManager{
		configDir:  filepath.Join(tempDir, ".boba"),
		configPath: filepath.Join(tempDir, ".boba", "config.json"),
		credPath:   filepath.Join(tempDir, ".boba", "credentials.json"),
		config: &Config{
			ToolOverrides: make(map[string]bool),
		},
		credentials: &Credentials{},
	}
	
	// Test getting non-existent override
	enabled, exists := cm.GetToolOverride("nonexistent")
	if exists {
		t.Error("Expected override to not exist")
	}
	if enabled {
		t.Error("Expected override to be false")
	}
	
	// Test setting override
	err := cm.SetToolOverride("git", true)
	if err != nil {
		t.Fatalf("SetToolOverride failed: %v", err)
	}
	
	enabled, exists = cm.GetToolOverride("git")
	if !exists {
		t.Error("Expected override to exist")
	}
	if !enabled {
		t.Error("Expected override to be true")
	}
	
	// Test removing override
	err = cm.RemoveToolOverride("git")
	if err != nil {
		t.Fatalf("RemoveToolOverride failed: %v", err)
	}
	
	enabled, exists = cm.GetToolOverride("git")
	if exists {
		t.Error("Expected override to not exist after removal")
	}
}

func TestSetRepositoryURL(t *testing.T) {
	tempDir := t.TempDir()
	
	cm := &ConfigManager{
		configDir:  filepath.Join(tempDir, ".boba"),
		configPath: filepath.Join(tempDir, ".boba", "config.json"),
		credPath:   filepath.Join(tempDir, ".boba", "credentials.json"),
		config:     nil, // Test with nil config
		credentials: &Credentials{},
	}
	
	testURL := "https://github.com/test/repo"
	err := cm.SetRepositoryURL(testURL)
	if err != nil {
		t.Fatalf("SetRepositoryURL failed: %v", err)
	}
	
	if cm.config.RepositoryURL != testURL {
		t.Errorf("Expected RepositoryURL %s, got %s", testURL, cm.config.RepositoryURL)
	}
}

func TestSetGitHubToken(t *testing.T) {
	tempDir := t.TempDir()
	
	cm := &ConfigManager{
		configDir:  filepath.Join(tempDir, ".boba"),
		configPath: filepath.Join(tempDir, ".boba", "config.json"),
		credPath:   filepath.Join(tempDir, ".boba", "credentials.json"),
		config:     &Config{},
		credentials: nil, // Test with nil credentials
	}
	
	testToken := "ghp_test123456789"
	err := cm.SetGitHubToken(testToken)
	if err != nil {
		t.Fatalf("SetGitHubToken failed: %v", err)
	}
	
	if cm.credentials.GitHubToken != testToken {
		t.Errorf("Expected GitHubToken %s, got %s", testToken, cm.credentials.GitHubToken)
	}
}

func TestValidateConfig(t *testing.T) {
	// Test with nil config
	cm := &ConfigManager{}
	err := cm.ValidateConfig()
	if err == nil {
		t.Error("Expected validation to fail with nil config")
	}
	
	// Test with valid config
	cm.config = &Config{
		RepositoryURL: "https://github.com/test/repo",
		ToolOverrides: map[string]bool{},
	}
	
	err = cm.ValidateConfig()
	if err != nil {
		t.Fatalf("ValidateConfig failed: %v", err)
	}
	
	// Test with invalid repository URL
	cm.config.RepositoryURL = "short"
	err = cm.ValidateConfig()
	if err == nil {
		t.Error("Expected validation to fail with short repository URL")
	}
	
	// Test with nil ToolOverrides (should be fixed by validation)
	cm.config.RepositoryURL = "https://github.com/test/repo"
	cm.config.ToolOverrides = nil
	
	err = cm.ValidateConfig()
	if err != nil {
		t.Fatalf("ValidateConfig failed: %v", err)
	}
	
	if cm.config.ToolOverrides == nil {
		t.Error("Expected ToolOverrides to be initialized by validation")
	}
}

func TestUpdateLastSync(t *testing.T) {
	tempDir := t.TempDir()
	
	cm := &ConfigManager{
		configDir:  filepath.Join(tempDir, ".boba"),
		configPath: filepath.Join(tempDir, ".boba", "config.json"),
		credPath:   filepath.Join(tempDir, ".boba", "credentials.json"),
		config:     nil, // Test with nil config
		credentials: &Credentials{},
	}
	
	before := time.Now()
	err := cm.UpdateLastSync()
	after := time.Now()
	
	if err != nil {
		t.Fatalf("UpdateLastSync failed: %v", err)
	}
	
	if cm.config.LastSync.Before(before) || cm.config.LastSync.After(after) {
		t.Error("LastSync timestamp is not within expected range")
	}
}

func TestGetConfig(t *testing.T) {
	cm := &ConfigManager{
		config: &Config{
			RepositoryURL: "https://github.com/test/repo",
			ToolOverrides: map[string]bool{
				"git": true,
			},
		},
	}
	
	config := cm.GetConfig()
	
	// Verify we get a copy, not the original
	config.RepositoryURL = "modified"
	if cm.config.RepositoryURL == "modified" {
		t.Error("GetConfig should return a copy, not the original")
	}
	
	// Test with nil config
	cm.config = nil
	config = cm.GetConfig()
	if config.ToolOverrides == nil {
		t.Error("GetConfig should initialize ToolOverrides even with nil config")
	}
}

func TestGetCredentials(t *testing.T) {
	cm := &ConfigManager{
		credentials: &Credentials{
			GitHubToken: "test-token",
		},
	}
	
	creds := cm.GetCredentials()
	
	// Verify we get a copy, not the original
	creds.GitHubToken = "modified"
	if cm.credentials.GitHubToken == "modified" {
		t.Error("GetCredentials should return a copy, not the original")
	}
	
	// Test with nil credentials
	cm.credentials = nil
	creds = cm.GetCredentials()
	if creds.GitHubToken != "" {
		t.Error("GetCredentials should return empty credentials when nil")
	}
}

func TestGetEnabledTools(t *testing.T) {
	tempDir := t.TempDir()
	
	cm := &ConfigManager{
		configDir:  filepath.Join(tempDir, ".boba"),
		configPath: filepath.Join(tempDir, ".boba", "config.json"),
		credPath:   filepath.Join(tempDir, ".boba", "credentials.json"),
		config: &Config{
			ToolOverrides: map[string]bool{
				"docker": false,
				"git":    true,
			},
		},
		credentials: &Credentials{},
	}
	
	allTools := []string{"git", "docker", "kubectl", "terraform"}
	enabledTools := cm.GetEnabledTools(allTools)
	
	// Should include git (explicitly enabled), kubectl and terraform (no override = enabled by default)
	// Should exclude docker (explicitly disabled)
	expectedEnabled := []string{"git", "kubectl", "terraform"}
	
	if len(enabledTools) != len(expectedEnabled) {
		t.Errorf("Expected %d enabled tools, got %d", len(expectedEnabled), len(enabledTools))
	}
	
	for _, expected := range expectedEnabled {
		found := false
		for _, enabled := range enabledTools {
			if enabled == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected tool %s to be enabled", expected)
		}
	}
	
	// Verify docker is not in the enabled list
	for _, enabled := range enabledTools {
		if enabled == "docker" {
			t.Error("Docker should not be enabled")
		}
	}
}

func TestIsConfigured(t *testing.T) {
	cm := &ConfigManager{
		config: &Config{
			RepositoryURL: "",
		},
	}
	
	if cm.IsConfigured() {
		t.Error("Expected IsConfigured to return false with empty repository URL")
	}
	
	cm.config.RepositoryURL = "https://github.com/test/repo"
	if !cm.IsConfigured() {
		t.Error("Expected IsConfigured to return true with repository URL set")
	}
}

func TestHasGitHubToken(t *testing.T) {
	cm := &ConfigManager{
		credentials: &Credentials{
			GitHubToken: "",
		},
	}
	
	if cm.HasGitHubToken() {
		t.Error("Expected HasGitHubToken to return false with empty token")
	}
	
	cm.credentials.GitHubToken = "ghp_test123"
	if !cm.HasGitHubToken() {
		t.Error("Expected HasGitHubToken to return true with token set")
	}
}