package main

import (
	"fmt"

	"boba/internal/installer"
	"boba/internal/parser"
)

// MockGitHubClient for testing without real GitHub API calls
type MockGitHubClient struct{}

func (m *MockGitHubClient) GetRepositoryContents(path string) ([]byte, error) {
	// Return different scripts based on the path
	switch path {
	case "tools/test-tool/install.sh":
		return []byte(`#!/bin/bash
echo "🚀 Starting installation of $BOBA_TOOL_NAME"
echo "📋 Platform: $BOBA_PLATFORM"
echo "📦 Package Manager: $BOBA_PACKAGE_MANAGER"
echo ""
echo "⏳ Downloading test-tool..."
sleep 1
echo "🔧 Installing test-tool..."
sleep 1
echo "⚙️  Configuring test-tool..."
sleep 1
echo ""
echo "✅ Installation of $BOBA_TOOL_NAME completed successfully!"
exit 0
`), nil
	case "tools/test-tool/uninstall.sh":
		return []byte(`#!/bin/bash
echo "🗑️  Starting uninstallation of $BOBA_TOOL_NAME"
echo "📋 Platform: $BOBA_PLATFORM"
echo ""
echo "🧹 Removing test-tool files..."
sleep 1
echo "🧽 Cleaning up configuration..."
sleep 1
echo ""
echo "✅ Uninstallation of $BOBA_TOOL_NAME completed successfully!"
exit 0
`), nil
	case "tools/failing-tool/install.sh":
		return []byte(`#!/bin/bash
echo "❌ Starting installation of failing-tool"
echo "⚠️  This tool will fail to install"
echo "🔍 Simulating error condition..."
echo "ERROR: Installation failed - dependency not found!" >&2
exit 1
`), nil
	default:
		return nil, fmt.Errorf("script not found: %s", path)
	}
}

func main() {
	fmt.Println("🧪 Testing BOBA Installation Engine")
	fmt.Println("===================================")
	
	// Create mock GitHub client and installation engine
	mockClient := &MockGitHubClient{}
	engine := installer.NewInstallationEngine(mockClient)
	defer engine.Cleanup()
	
	// Display platform information
	platform := engine.GetPlatform()
	fmt.Printf("🖥️  Detected Platform: %s\n", platform.OS)
	if platform.Distribution != "" {
		fmt.Printf("🐧 Distribution: %s\n", platform.Distribution)
	}
	if platform.PackageManager != "" {
		fmt.Printf("📦 Package Manager: %s\n", platform.PackageManager)
	}
	fmt.Println()
	
	// Test 1: Check if a real tool is installed
	fmt.Println("🔍 Test 1: Checking if 'ls' command is installed...")
	realTool := parser.Tool{Name: "ls"}
	if engine.IsToolInstalled(realTool) {
		fmt.Println("✅ 'ls' command is installed")
		verified, message := engine.VerifyInstallation(realTool)
		if verified {
			fmt.Printf("✅ Verification: %s\n", message)
		}
	} else {
		fmt.Println("❌ 'ls' command not found")
	}
	fmt.Println()
	
	// Test 2: Install a mock tool successfully
	fmt.Println("🔧 Test 2: Installing test-tool...")
	testTool := parser.Tool{
		Name:          "test-tool",
		Description:   "A test tool for demonstration",
		FolderName:    "test-tool",
		InstallScript: "tools/test-tool/install.sh",
		AutoInstall:   true,
	}
	
	result, err := engine.InstallTool(testTool)
	if err != nil {
		fmt.Printf("❌ Installation failed: %v\n", err)
	} else if result.Success {
		fmt.Printf("✅ Installation successful! (took %v)\n", result.Duration)
		fmt.Println("📄 Output:")
		fmt.Println(result.Output)
	} else {
		fmt.Printf("❌ Installation failed with exit code %d\n", result.ExitCode)
		fmt.Printf("📄 Output:\n%s\n", result.Output)
	}
	fmt.Println()
	
	// Test 3: Uninstall the tool
	fmt.Println("🗑️  Test 3: Uninstalling test-tool...")
	testTool.UninstallScript = "tools/test-tool/uninstall.sh"
	
	result, err = engine.UninstallTool(testTool)
	if err != nil {
		fmt.Printf("❌ Uninstallation failed: %v\n", err)
	} else if result.Success {
		fmt.Printf("✅ Uninstallation successful! (took %v)\n", result.Duration)
		fmt.Println("📄 Output:")
		fmt.Println(result.Output)
	} else {
		fmt.Printf("❌ Uninstallation failed with exit code %d\n", result.ExitCode)
		fmt.Printf("📄 Output:\n%s\n", result.Output)
	}
	fmt.Println()
	
	// Test 4: Try to install a failing tool
	fmt.Println("💥 Test 4: Installing failing-tool (should fail)...")
	failingTool := parser.Tool{
		Name:          "failing-tool",
		FolderName:    "failing-tool",
		InstallScript: "tools/failing-tool/install.sh",
	}
	
	result, err = engine.InstallTool(failingTool)
	if err != nil {
		fmt.Printf("❌ Installation failed as expected: %v\n", err)
		fmt.Printf("📄 Exit Code: %d\n", result.ExitCode)
		fmt.Println("📄 Output:")
		fmt.Println(result.Output)
	} else {
		fmt.Println("⚠️  Expected installation to fail, but it succeeded")
	}
	fmt.Println()
	
	// Test 5: Execute a simple command
	fmt.Println("⚡ Test 5: Executing simple command...")
	output, err := engine.ExecuteCommand("echo 'Hello from BOBA installer!'")
	if err != nil {
		fmt.Printf("❌ Command failed: %v\n", err)
	} else {
		fmt.Printf("✅ Command output: %s\n", output)
	}
	
	fmt.Println("🎉 All tests completed!")
}