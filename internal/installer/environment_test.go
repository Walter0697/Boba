package installer

import (
	"fmt"
	"strings"
	"testing"
	"boba/internal/parser"
)

// MockGitHubClientForEnvironment is a mock GitHub client for environment testing
type MockGitHubClientForEnvironment struct {
	setupScript   []byte
	restoreScript []byte
	shouldError   bool
}

func (m *MockGitHubClientForEnvironment) GetRepositoryContents(path string) ([]byte, error) {
	if m.shouldError {
		return nil, fmt.Errorf("mock error")
	}
	
	if strings.Contains(path, "setup.sh") {
		if m.setupScript != nil {
			return m.setupScript, nil
		}
		return []byte("#!/bin/bash\necho 'Setting up environment'\nexit 0"), nil
	}
	
	if strings.Contains(path, "restore.sh") {
		if m.restoreScript != nil {
			return m.restoreScript, nil
		}
		return []byte("#!/bin/bash\necho 'Restoring environment'\nexit 0"), nil
	}
	
	return nil, fmt.Errorf("file not found: %s", path)
}

func TestApplyEnvironment(t *testing.T) {
	mockClient := &MockGitHubClientForEnvironment{}
	engine := NewInstallationEngine(mockClient)
	
	env := parser.Environment{
		Name:        "test-env",
		Description: "Test environment",
		FolderName:  "test-env",
		SetupScript: "environments/test-env/setup.sh",
	}
	
	result, err := engine.ApplyEnvironment(env)
	
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	if !result.Success {
		t.Errorf("Expected successful environment application, got failure: %s", result.Output)
	}
	
	if result.Duration == 0 {
		t.Error("Expected non-zero duration")
	}
}

func TestApplyEnvironmentWithError(t *testing.T) {
	mockClient := &MockGitHubClientForEnvironment{shouldError: true}
	engine := NewInstallationEngine(mockClient)
	
	env := parser.Environment{
		Name:        "test-env",
		FolderName:  "test-env",
		SetupScript: "environments/test-env/setup.sh",
	}
	
	result, err := engine.ApplyEnvironment(env)
	
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	
	if result.Success {
		t.Error("Expected failed environment application")
	}
}

func TestRestoreEnvironment(t *testing.T) {
	mockClient := &MockGitHubClientForEnvironment{}
	engine := NewInstallationEngine(mockClient)
	
	env := parser.Environment{
		Name:          "test-env",
		Description:   "Test environment",
		FolderName:    "test-env",
		RestoreScript: "environments/test-env/restore.sh",
	}
	
	result, err := engine.RestoreEnvironment(env)
	
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	if !result.Success {
		t.Errorf("Expected successful environment restoration, got failure: %s", result.Output)
	}
}

func TestApplyEnvironmentMissingScript(t *testing.T) {
	mockClient := &MockGitHubClientForEnvironment{shouldError: true}
	engine := NewInstallationEngine(mockClient)
	
	env := parser.Environment{
		Name:        "test-env",
		FolderName:  "test-env",
		SetupScript: "environments/test-env/nonexistent.sh",
	}
	
	result, err := engine.ApplyEnvironment(env)
	
	if err == nil {
		t.Fatal("Expected error for missing script, got nil")
	}
	
	if result.Success {
		t.Error("Expected failed environment application for missing script")
	}
}

func TestIsEnvironmentApplied(t *testing.T) {
	mockClient := &MockGitHubClientForEnvironment{}
	engine := NewInstallationEngine(mockClient)
	
	env := parser.Environment{
		Name:       "test-env",
		FolderName: "test-env",
	}
	
	// Currently always returns false (placeholder implementation)
	isApplied := engine.IsEnvironmentApplied(env)
	
	if isApplied {
		t.Error("Expected environment not to be applied (placeholder implementation)")
	}
}

func TestVerifyEnvironmentApplication(t *testing.T) {
	mockClient := &MockGitHubClientForEnvironment{}
	engine := NewInstallationEngine(mockClient)
	
	env := parser.Environment{
		Name:       "test-env",
		FolderName: "test-env",
	}
	
	// Currently always returns true (placeholder implementation)
	success, message := engine.VerifyEnvironmentApplication(env)
	
	if !success {
		t.Error("Expected environment verification to succeed (placeholder implementation)")
	}
	
	if message == "" {
		t.Error("Expected non-empty verification message")
	}
}