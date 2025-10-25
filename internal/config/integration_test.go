package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestConfigManagerIntegration tests the complete workflow of the ConfigManager
func TestConfigManagerIntegration(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Override the home directory for testing
	originalHome := os.Getenv("HOME")
	defer func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		}
	}()
	os.Setenv("HOME", tempDir)
	
	// Create a new ConfigManager (should use the temp home directory)
	cm := NewConfigManager()
	
	// Test initial setup
	err := cm.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load initial config: %v", err)
	}
	
	err = cm.LoadCredentials()
	if err != nil {
		t.Fatalf("Failed to load initial credentials: %v", err)
	}
	
	// Verify initial state
	if cm.IsConfigured() {
		t.Error("Expected configuration to be empty initially")
	}
	
	if cm.HasGitHubToken() {
		t.Error("Expected no GitHub token initially")
	}
	
	// Configure the application
	testRepoURL := "https://github.com/testuser/dev-setup"
	err = cm.SetRepositoryURL(testRepoURL)
	if err != nil {
		t.Fatalf("Failed to set repository URL: %v", err)
	}
	
	testToken := "ghp_test_token_123456789"
	err = cm.SetGitHubToken(testToken)
	if err != nil {
		t.Fatalf("Failed to set GitHub token: %v", err)
	}
	
	// Set some tool overrides
	err = cm.SetToolOverride("docker", false)
	if err != nil {
		t.Fatalf("Failed to set docker override: %v", err)
	}
	
	err = cm.SetToolOverride("git", true)
	if err != nil {
		t.Fatalf("Failed to set git override: %v", err)
	}
	
	err = cm.SetToolOverride("kubectl", true)
	if err != nil {
		t.Fatalf("Failed to set kubectl override: %v", err)
	}
	
	// Verify configuration state
	if !cm.IsConfigured() {
		t.Error("Expected configuration to be set")
	}
	
	if !cm.HasGitHubToken() {
		t.Error("Expected GitHub token to be set")
	}
	
	// Test tool filtering
	allTools := []string{"git", "docker", "kubectl", "terraform", "aws-cli"}
	enabledTools := cm.GetEnabledTools(allTools)
	
	// Should include: git (enabled), kubectl (enabled), terraform (default), aws-cli (default)
	// Should exclude: docker (disabled)
	expectedCount := 4
	if len(enabledTools) != expectedCount {
		t.Errorf("Expected %d enabled tools, got %d: %v", expectedCount, len(enabledTools), enabledTools)
	}
	
	// Verify docker is not in enabled tools
	for _, tool := range enabledTools {
		if tool == "docker" {
			t.Error("Docker should not be in enabled tools")
		}
	}
	
	// Verify git is in enabled tools
	gitFound := false
	for _, tool := range enabledTools {
		if tool == "git" {
			gitFound = true
			break
		}
	}
	if !gitFound {
		t.Error("Git should be in enabled tools")
	}
	
	// Test persistence by creating a new ConfigManager instance
	cm2 := NewConfigManager()
	
	err = cm2.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config with second instance: %v", err)
	}
	
	err = cm2.LoadCredentials()
	if err != nil {
		t.Fatalf("Failed to load credentials with second instance: %v", err)
	}
	
	// Verify persistence
	config := cm2.GetConfig()
	if config.RepositoryURL != testRepoURL {
		t.Errorf("Repository URL not persisted: expected %s, got %s", testRepoURL, config.RepositoryURL)
	}
	
	creds := cm2.GetCredentials()
	if creds.GitHubToken != testToken {
		t.Errorf("GitHub token not persisted: expected %s, got %s", testToken, creds.GitHubToken)
	}
	
	// Verify tool overrides persistence
	if enabled, exists := cm2.GetToolOverride("docker"); !exists || enabled {
		t.Error("Docker override not persisted correctly")
	}
	
	if enabled, exists := cm2.GetToolOverride("git"); !exists || !enabled {
		t.Error("Git override not persisted correctly")
	}
	
	// Test validation
	err = cm2.ValidateConfig()
	if err != nil {
		t.Fatalf("Configuration validation failed: %v", err)
	}
	
	// Verify file structure was created correctly
	expectedFiles := []string{
		filepath.Join(cm.GetConfigDir(), "config.json"),
		filepath.Join(cm.GetConfigDir(), "credentials.json"),
		filepath.Join(cm.GetConfigDir(), "cache"),
	}
	
	for _, expectedPath := range expectedFiles {
		if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
			t.Errorf("Expected file/directory not created: %s", expectedPath)
		}
	}
	
	// Test credentials file permissions
	credFile := filepath.Join(cm.GetConfigDir(), "credentials.json")
	info, err := os.Stat(credFile)
	if err != nil {
		t.Fatalf("Failed to stat credentials file: %v", err)
	}
	
	if info.Mode().Perm() != 0600 {
		t.Errorf("Credentials file has incorrect permissions: expected 0600, got %o", info.Mode().Perm())
	}
}