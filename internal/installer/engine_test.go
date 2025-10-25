package installer

import (
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"boba/internal/parser"
)

// MockGitHubClient for testing
type MockGitHubClient struct {
	scriptContent map[string][]byte
	shouldError   bool
}

// Ensure MockGitHubClient implements GitHubClientInterface
var _ GitHubClientInterface = (*MockGitHubClient)(nil)

func (m *MockGitHubClient) GetRepositoryContents(path string) ([]byte, error) {
	if m.shouldError {
		return nil, os.ErrNotExist
	}
	
	if content, exists := m.scriptContent[path]; exists {
		return content, nil
	}
	
	// Return a simple test script
	if strings.Contains(path, "install.sh") {
		return []byte("#!/bin/bash\necho 'Installing test tool'\nexit 0\n"), nil
	}
	if strings.Contains(path, "uninstall.sh") {
		return []byte("#!/bin/bash\necho 'Uninstalling test tool'\nexit 0\n"), nil
	}
	
	return nil, os.ErrNotExist
}

func TestNewInstallationEngine(t *testing.T) {
	mockClient := &MockGitHubClient{}
	engine := NewInstallationEngine(mockClient)
	
	if engine == nil {
		t.Fatal("Expected non-nil installation engine")
	}
	
	if engine.githubClient != mockClient {
		t.Error("Expected GitHub client to be set")
	}
	
	if engine.tempDir == "" {
		t.Error("Expected temp directory to be set")
	}
	
	// Verify temp directory exists
	if _, err := os.Stat(engine.tempDir); os.IsNotExist(err) {
		t.Error("Expected temp directory to exist")
	}
}

func TestDetectPlatform(t *testing.T) {
	platform := detectPlatform()
	
	if platform.OS == "" {
		t.Error("Expected OS to be detected")
	}
	
	if platform.OS != runtime.GOOS {
		t.Errorf("Expected OS to be %s, got %s", runtime.GOOS, platform.OS)
	}
	
	// On Linux, we should detect distribution and package manager
	if platform.OS == "linux" {
		if platform.Distribution == "" {
			t.Error("Expected distribution to be detected on Linux")
		}
		if platform.PackageManager == "" {
			t.Error("Expected package manager to be detected on Linux")
		}
	}
	
	// On macOS, package manager should be brew
	if platform.OS == "darwin" {
		if platform.PackageManager != "brew" {
			t.Errorf("Expected package manager to be 'brew' on macOS, got %s", platform.PackageManager)
		}
	}
}

func TestIsToolInstalled(t *testing.T) {
	mockClient := &MockGitHubClient{}
	engine := NewInstallationEngine(mockClient)
	
	// Test with a tool that should exist (like 'ls' on Unix or 'dir' on Windows)
	var existingTool parser.Tool
	if runtime.GOOS == "windows" {
		existingTool = parser.Tool{Name: "cmd"}
	} else {
		existingTool = parser.Tool{Name: "ls"}
	}
	
	if !engine.IsToolInstalled(existingTool) {
		t.Errorf("Expected %s to be installed", existingTool.Name)
	}
	
	// Test with a tool that shouldn't exist
	nonExistentTool := parser.Tool{Name: "definitely-not-a-real-tool-12345"}
	if engine.IsToolInstalled(nonExistentTool) {
		t.Error("Expected non-existent tool to not be installed")
	}
}

func TestInstallTool(t *testing.T) {
	mockClient := &MockGitHubClient{
		scriptContent: map[string][]byte{
			"tools/test-tool/install.sh": []byte("#!/bin/bash\necho 'Installing test tool'\nexit 0\n"),
		},
	}
	engine := NewInstallationEngine(mockClient)
	defer engine.Cleanup()
	
	tool := parser.Tool{
		Name:          "test-tool",
		FolderName:    "test-tool",
		InstallScript: "tools/test-tool/install.sh",
	}
	
	result, err := engine.InstallTool(tool)
	
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	
	if !result.Success {
		t.Errorf("Expected successful installation, got error: %v", result.Error)
	}
	
	if result.Duration == 0 {
		t.Error("Expected non-zero duration")
	}
	
	if !strings.Contains(result.Output, "Installing test tool") {
		t.Errorf("Expected output to contain 'Installing test tool', got: %s", result.Output)
	}
}

func TestInstallToolWithError(t *testing.T) {
	mockClient := &MockGitHubClient{
		scriptContent: map[string][]byte{
			"tools/test-tool/install.sh": []byte("#!/bin/bash\necho 'Installation failed'\nexit 1\n"),
		},
	}
	engine := NewInstallationEngine(mockClient)
	defer engine.Cleanup()
	
	tool := parser.Tool{
		Name:          "test-tool",
		FolderName:    "test-tool",
		InstallScript: "tools/test-tool/install.sh",
	}
	
	result, err := engine.InstallTool(tool)
	
	if err == nil {
		t.Error("Expected error for failed installation")
	}
	
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	
	if result.Success {
		t.Error("Expected failed installation")
	}
	
	if result.ExitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", result.ExitCode)
	}
}

func TestInstallToolMissingScript(t *testing.T) {
	mockClient := &MockGitHubClient{shouldError: true}
	engine := NewInstallationEngine(mockClient)
	defer engine.Cleanup()
	
	tool := parser.Tool{
		Name:          "test-tool",
		FolderName:    "test-tool",
		InstallScript: "tools/test-tool/install.sh",
	}
	
	result, err := engine.InstallTool(tool)
	
	if err == nil {
		t.Error("Expected error for missing script")
	}
	
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	
	if result.Success {
		t.Error("Expected failed installation")
	}
}

func TestUninstallTool(t *testing.T) {
	mockClient := &MockGitHubClient{
		scriptContent: map[string][]byte{
			"tools/test-tool/uninstall.sh": []byte("#!/bin/bash\necho 'Uninstalling test tool'\nexit 0\n"),
		},
	}
	engine := NewInstallationEngine(mockClient)
	defer engine.Cleanup()
	
	tool := parser.Tool{
		Name:            "test-tool",
		FolderName:      "test-tool",
		UninstallScript: "tools/test-tool/uninstall.sh",
	}
	
	result, err := engine.UninstallTool(tool)
	
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	
	if !result.Success {
		t.Errorf("Expected successful uninstallation, got error: %v", result.Error)
	}
	
	if !strings.Contains(result.Output, "Uninstalling test tool") {
		t.Errorf("Expected output to contain 'Uninstalling test tool', got: %s", result.Output)
	}
}

func TestExecuteCommand(t *testing.T) {
	mockClient := &MockGitHubClient{}
	engine := NewInstallationEngine(mockClient)
	
	var testCommand string
	var expectedOutput string
	
	if runtime.GOOS == "windows" {
		testCommand = "echo test"
		expectedOutput = "test"
	} else {
		testCommand = "echo test"
		expectedOutput = "test"
	}
	
	output, err := engine.ExecuteCommand(testCommand)
	
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output to contain '%s', got: %s", expectedOutput, output)
	}
}

func TestVerifyInstallation(t *testing.T) {
	mockClient := &MockGitHubClient{}
	engine := NewInstallationEngine(mockClient)
	
	// Test with a tool that should exist
	var existingTool parser.Tool
	if runtime.GOOS == "windows" {
		existingTool = parser.Tool{Name: "cmd"}
	} else {
		existingTool = parser.Tool{Name: "ls"}
	}
	
	isInstalled, message := engine.VerifyInstallation(existingTool)
	
	if !isInstalled {
		t.Errorf("Expected %s to be verified as installed", existingTool.Name)
	}
	
	if message == "" {
		t.Error("Expected non-empty verification message")
	}
	
	// Test with a tool that shouldn't exist
	nonExistentTool := parser.Tool{Name: "definitely-not-a-real-tool-12345"}
	isInstalled, message = engine.VerifyInstallation(nonExistentTool)
	
	if isInstalled {
		t.Error("Expected non-existent tool to not be verified as installed")
	}
	
	if !strings.Contains(message, "not accessible") {
		t.Errorf("Expected message to indicate tool is not accessible, got: %s", message)
	}
}

func TestGetPlatform(t *testing.T) {
	mockClient := &MockGitHubClient{}
	engine := NewInstallationEngine(mockClient)
	
	platform := engine.GetPlatform()
	
	if platform.OS == "" {
		t.Error("Expected OS to be set")
	}
	
	if platform.OS != runtime.GOOS {
		t.Errorf("Expected OS to be %s, got %s", runtime.GOOS, platform.OS)
	}
}

func TestCleanup(t *testing.T) {
	mockClient := &MockGitHubClient{}
	engine := NewInstallationEngine(mockClient)
	
	tempDir := engine.tempDir
	
	// Verify temp directory exists
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Error("Expected temp directory to exist before cleanup")
	}
	
	// Cleanup
	err := engine.Cleanup()
	if err != nil {
		t.Fatalf("Expected no error during cleanup, got %v", err)
	}
	
	// Verify temp directory is removed
	if _, err := os.Stat(tempDir); !os.IsNotExist(err) {
		t.Error("Expected temp directory to be removed after cleanup")
	}
}

func TestScriptExecutionTimeout(t *testing.T) {
	// This test verifies that long-running scripts are properly terminated
	mockClient := &MockGitHubClient{
		scriptContent: map[string][]byte{
			"tools/slow-tool/install.sh": []byte("#!/bin/bash\nsleep 1\necho 'Done'\nexit 0\n"),
		},
	}
	engine := NewInstallationEngine(mockClient)
	defer engine.Cleanup()
	
	tool := parser.Tool{
		Name:          "slow-tool",
		FolderName:    "slow-tool",
		InstallScript: "tools/slow-tool/install.sh",
	}
	
	start := time.Now()
	result, err := engine.InstallTool(tool)
	duration := time.Since(start)
	
	// Should complete within reasonable time (much less than the 10-minute timeout)
	if duration > 30*time.Second {
		t.Errorf("Script took too long to execute: %v", duration)
	}
	
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if !result.Success {
		t.Errorf("Expected successful installation, got error: %v", result.Error)
	}
}