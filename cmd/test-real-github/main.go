package main

import (
	"fmt"
	"os"

	"boba/internal/github"
	"boba/internal/installer"
	"boba/internal/parser"
)

func main() {
	// This requires a real GitHub token and repository
	// Set these environment variables:
	// export GITHUB_TOKEN="your_token_here"
	// export GITHUB_REPO="username/repo-name"
	
	token := os.Getenv("GITHUB_TOKEN")
	repo := os.Getenv("GITHUB_REPO")
	
	if token == "" || repo == "" {
		fmt.Println("❌ Please set GITHUB_TOKEN and GITHUB_REPO environment variables")
		fmt.Println("Example:")
		fmt.Println("  export GITHUB_TOKEN=\"ghp_your_token_here\"")
		fmt.Println("  export GITHUB_REPO=\"username/boba-config\"")
		os.Exit(1)
	}
	
	// Parse repository URL
	owner, repoName, err := github.ParseRepositoryURL(repo)
	if err != nil {
		fmt.Printf("❌ Invalid repository URL: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("🔗 Testing with repository: %s/%s\n", owner, repoName)
	
	// Create GitHub client
	githubClient := github.NewGitHubClient(token, owner, repoName)
	
	// Test connection
	if err := githubClient.TestConnection(); err != nil {
		fmt.Printf("❌ GitHub connection failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ GitHub connection successful")
	
	// Create installation engine
	engine := installer.NewInstallationEngine(githubClient)
	defer engine.Cleanup()
	
	// Example: Try to install a tool from your repository
	// This assumes you have a tool in tools/example-tool/ directory
	tool := parser.Tool{
		Name:          "example-tool",
		FolderName:    "example-tool",
		InstallScript: "tools/example-tool/install.sh",
	}
	
	fmt.Printf("🔧 Attempting to install %s...\n", tool.Name)
	result, err := engine.InstallTool(tool)
	
	if err != nil {
		fmt.Printf("❌ Installation failed: %v\n", err)
		if result != nil {
			fmt.Printf("📄 Output:\n%s\n", result.Output)
		}
	} else if result.Success {
		fmt.Printf("✅ Installation successful! (took %v)\n", result.Duration)
		fmt.Printf("📄 Output:\n%s\n", result.Output)
		
		// Verify installation
		verified, message := engine.VerifyInstallation(tool)
		fmt.Printf("🔍 Verification: %t - %s\n", verified, message)
	} else {
		fmt.Printf("❌ Installation failed with exit code %d\n", result.ExitCode)
		fmt.Printf("📄 Output:\n%s\n", result.Output)
	}
}