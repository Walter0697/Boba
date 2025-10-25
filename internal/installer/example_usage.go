package installer

import (
	"fmt"
	"log"

	"boba/internal/github"
	"boba/internal/parser"
)

// ExampleUsage demonstrates how to use the InstallationEngine
func ExampleUsage() {
	// This is an example of how the InstallationEngine would be used
	// in the main application once task 8 (integration) is implemented
	
	// 1. Create GitHub client (this would come from authentication flow)
	githubClient := github.NewGitHubClient("your-token", "your-username", "boba-config")
	
	// 2. Create installation engine
	engine := NewInstallationEngine(githubClient)
	defer engine.Cleanup() // Always cleanup temporary files
	
	// 3. Create a sample tool (this would come from repository parser)
	tool := parser.Tool{
		Name:            "example-tool",
		Description:     "An example CLI tool",
		FolderName:      "example-tool",
		InstallScript:   "tools/example-tool/install.sh",
		UninstallScript: "tools/example-tool/uninstall.sh",
		AutoInstall:     true,
	}
	
	// 4. Check if tool is already installed
	if engine.IsToolInstalled(tool) {
		fmt.Printf("Tool %s is already installed\n", tool.Name)
		
		// Verify installation
		isVerified, message := engine.VerifyInstallation(tool)
		fmt.Printf("Verification: %t - %s\n", isVerified, message)
		return
	}
	
	// 5. Install the tool
	fmt.Printf("Installing tool: %s\n", tool.Name)
	result, err := engine.InstallTool(tool)
	if err != nil {
		log.Printf("Installation failed: %v", err)
		return
	}
	
	// 6. Check installation result
	if result.Success {
		fmt.Printf("Installation successful! Duration: %v\n", result.Duration)
		fmt.Printf("Output:\n%s\n", result.Output)
		
		// 7. Verify the installation
		isVerified, message := engine.VerifyInstallation(tool)
		fmt.Printf("Verification: %t - %s\n", isVerified, message)
	} else {
		fmt.Printf("Installation failed with exit code %d\n", result.ExitCode)
		fmt.Printf("Error: %v\n", result.Error)
		fmt.Printf("Output:\n%s\n", result.Output)
	}
}

// ExampleUninstallUsage demonstrates how to uninstall a tool
func ExampleUninstallUsage() {
	// This shows how to uninstall a tool
	
	githubClient := github.NewGitHubClient("your-token", "your-username", "boba-config")
	engine := NewInstallationEngine(githubClient)
	defer engine.Cleanup()
	
	tool := parser.Tool{
		Name:            "example-tool",
		FolderName:      "example-tool",
		UninstallScript: "tools/example-tool/uninstall.sh",
	}
	
	// Check if tool is installed before uninstalling
	if !engine.IsToolInstalled(tool) {
		fmt.Printf("Tool %s is not installed\n", tool.Name)
		return
	}
	
	// Uninstall the tool
	fmt.Printf("Uninstalling tool: %s\n", tool.Name)
	result, err := engine.UninstallTool(tool)
	if err != nil {
		log.Printf("Uninstallation failed: %v", err)
		return
	}
	
	if result.Success {
		fmt.Printf("Uninstallation successful! Duration: %v\n", result.Duration)
		fmt.Printf("Output:\n%s\n", result.Output)
		
		// Verify the tool is no longer installed
		if !engine.IsToolInstalled(tool) {
			fmt.Printf("Tool %s successfully removed\n", tool.Name)
		} else {
			fmt.Printf("Warning: Tool %s may still be installed\n", tool.Name)
		}
	} else {
		fmt.Printf("Uninstallation failed with exit code %d\n", result.ExitCode)
		fmt.Printf("Error: %v\n", result.Error)
		fmt.Printf("Output:\n%s\n", result.Output)
	}
}

// ExampleBatchInstallation demonstrates installing multiple tools
func ExampleBatchInstallation() {
	githubClient := github.NewGitHubClient("your-token", "your-username", "boba-config")
	engine := NewInstallationEngine(githubClient)
	defer engine.Cleanup()
	
	// Sample tools (these would come from repository parser)
	tools := []parser.Tool{
		{
			Name:          "git",
			FolderName:    "git",
			InstallScript: "tools/git/install.sh",
			AutoInstall:   true,
		},
		{
			Name:          "docker",
			FolderName:    "docker",
			InstallScript: "tools/docker/install.sh",
			AutoInstall:   true,
		},
		{
			Name:          "kubectl",
			FolderName:    "kubectl",
			InstallScript: "tools/kubectl/install.sh",
			AutoInstall:   false, // Manual install only
		},
	}
	
	fmt.Printf("Installing %d tools...\n", len(tools))
	
	for _, tool := range tools {
		fmt.Printf("\n--- Processing tool: %s ---\n", tool.Name)
		
		// Skip if already installed
		if engine.IsToolInstalled(tool) {
			fmt.Printf("✓ %s is already installed\n", tool.Name)
			continue
		}
		
		// Install the tool
		result, err := engine.InstallTool(tool)
		if err != nil {
			fmt.Printf("✗ Failed to install %s: %v\n", tool.Name, err)
			continue
		}
		
		if result.Success {
			fmt.Printf("✓ Successfully installed %s (took %v)\n", tool.Name, result.Duration)
			
			// Verify installation
			isVerified, message := engine.VerifyInstallation(tool)
			if isVerified {
				fmt.Printf("✓ Verification passed: %s\n", message)
			} else {
				fmt.Printf("⚠ Verification failed: %s\n", message)
			}
		} else {
			fmt.Printf("✗ Installation failed for %s (exit code %d)\n", tool.Name, result.ExitCode)
			if result.Output != "" {
				fmt.Printf("Output: %s\n", result.Output)
			}
		}
	}
	
	fmt.Println("\nBatch installation complete!")
}