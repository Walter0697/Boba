package installer

import (
	"os"
	"testing"

	"boba/internal/parser"
)

// TestInstallationEngineIntegration tests the installation engine with more realistic scenarios
func TestInstallationEngineIntegration(t *testing.T) {
	// Skip integration tests in CI or when SKIP_INTEGRATION is set
	if os.Getenv("SKIP_INTEGRATION") != "" {
		t.Skip("Skipping integration test")
	}
	
	// Create a more comprehensive mock client
	mockClient := &MockGitHubClient{
		scriptContent: map[string][]byte{
			"tools/test-tool/install.sh": []byte(`#!/bin/bash
echo "Starting installation of test-tool"
echo "Platform: $BOBA_PLATFORM"
echo "Package Manager: $BOBA_PACKAGE_MANAGER"
echo "Tool Name: $BOBA_TOOL_NAME"

# Simulate some installation work
sleep 0.1
echo "Downloading test-tool..."
sleep 0.1
echo "Installing test-tool..."
sleep 0.1
echo "Configuring test-tool..."

echo "Installation completed successfully"
exit 0
`),
			"tools/test-tool/uninstall.sh": []byte(`#!/bin/bash
echo "Starting uninstallation of test-tool"
echo "Platform: $BOBA_PLATFORM"
echo "Tool Name: $BOBA_TOOL_NAME"

# Simulate some uninstallation work
sleep 0.1
echo "Removing test-tool files..."
sleep 0.1
echo "Cleaning up configuration..."

echo "Uninstallation completed successfully"
exit 0
`),
			"tools/failing-tool/install.sh": []byte(`#!/bin/bash
echo "Starting installation of failing-tool"
echo "This tool will fail to install"
echo "Simulating error condition..."
echo "ERROR: Installation failed!" >&2
exit 1
`),
		},
	}
	
	engine := NewInstallationEngine(mockClient)
	defer engine.Cleanup()
	
	t.Run("SuccessfulInstallation", func(t *testing.T) {
		tool := parser.Tool{
			Name:          "test-tool",
			Description:   "A test tool for integration testing",
			FolderName:    "test-tool",
			InstallScript: "tools/test-tool/install.sh",
			AutoInstall:   true,
		}
		
		result, err := engine.InstallTool(tool)
		
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		
		if !result.Success {
			t.Errorf("Expected successful installation, got error: %v", result.Error)
		}
		
		// Check that environment variables were passed correctly
		if !containsString(result.Output, "Platform:") {
			t.Error("Expected platform information in output")
		}
		
		if !containsString(result.Output, "Tool Name: test-tool") {
			t.Error("Expected tool name in output")
		}
		
		if !containsString(result.Output, "Installation completed successfully") {
			t.Error("Expected success message in output")
		}
		
		// Check that both stdout and stderr are captured
		if !containsString(result.Output, "STDOUT:") {
			t.Error("Expected stdout to be captured")
		}
	})
	
	t.Run("SuccessfulUninstallation", func(t *testing.T) {
		tool := parser.Tool{
			Name:            "test-tool",
			FolderName:      "test-tool",
			UninstallScript: "tools/test-tool/uninstall.sh",
		}
		
		result, err := engine.UninstallTool(tool)
		
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		
		if !result.Success {
			t.Errorf("Expected successful uninstallation, got error: %v", result.Error)
		}
		
		if !containsString(result.Output, "Uninstallation completed successfully") {
			t.Error("Expected success message in output")
		}
	})
	
	t.Run("FailedInstallation", func(t *testing.T) {
		tool := parser.Tool{
			Name:          "failing-tool",
			FolderName:    "failing-tool",
			InstallScript: "tools/failing-tool/install.sh",
		}
		
		result, err := engine.InstallTool(tool)
		
		if err == nil {
			t.Error("Expected error for failed installation")
		}
		
		if result.Success {
			t.Error("Expected failed installation")
		}
		
		if result.ExitCode != 1 {
			t.Errorf("Expected exit code 1, got %d", result.ExitCode)
		}
		
		// Check that stderr is captured
		if !containsString(result.Output, "ERROR: Installation failed!") {
			t.Error("Expected error message in output")
		}
		
		if !containsString(result.Output, "STDERR:") {
			t.Error("Expected stderr to be captured")
		}
	})
	
	t.Run("PlatformDetection", func(t *testing.T) {
		platform := engine.GetPlatform()
		
		if platform.OS == "" {
			t.Error("Expected OS to be detected")
		}
		
		// Verify platform-specific detection
		switch platform.OS {
		case "linux":
			if platform.Distribution == "" {
				t.Error("Expected Linux distribution to be detected")
			}
			if platform.PackageManager == "" {
				t.Error("Expected package manager to be detected on Linux")
			}
		case "darwin":
			if platform.PackageManager != "brew" {
				t.Errorf("Expected package manager to be 'brew' on macOS, got %s", platform.PackageManager)
			}
		case "windows":
			// Windows doesn't have a standard package manager detection
		default:
			t.Logf("Running on platform: %s", platform.OS)
		}
	})
	
	t.Run("SecurityMeasures", func(t *testing.T) {
		// Test that the engine properly handles script execution with security measures
		tool := parser.Tool{
			Name:          "test-tool",
			FolderName:    "test-tool",
			InstallScript: "tools/test-tool/install.sh",
		}
		
		result, err := engine.InstallTool(tool)
		
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		
		// Verify that the script was executed with proper environment variables
		// The script echoes these values, so we should see them in the output
		if !containsString(result.Output, "Tool Name: test-tool") {
			t.Error("Expected tool name environment variable to be passed to script")
		}
		
		if !containsString(result.Output, "Platform: " + engine.platform.OS) {
			t.Error("Expected platform environment variable to be passed to script")
		}
		
		// Verify that the script execution completed within reasonable time
		if result.Duration.Seconds() > 30 {
			t.Errorf("Script execution took too long: %v", result.Duration)
		}
	})
}

// Helper function to check if a string contains a substring
func containsString(haystack, needle string) bool {
	return len(haystack) >= len(needle) && 
		   (haystack == needle || 
		    haystack[:len(needle)] == needle || 
		    haystack[len(haystack)-len(needle):] == needle ||
		    findSubstring(haystack, needle))
}

func findSubstring(haystack, needle string) bool {
	for i := 0; i <= len(haystack)-len(needle); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}