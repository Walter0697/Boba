package ui

import (
	"fmt"
	"strings"
	
	"boba/internal/config"
	"boba/internal/github"
	"boba/internal/installer"
	"boba/internal/parser"
)

// InitialModel creates and initializes the menu with enhanced startup flow
func InitialModel() MenuModel {
	configManager := config.NewConfigManager()
	
	// Load existing configuration
	configManager.LoadConfig()
	configManager.LoadCredentials()
	
	// Initialize system installer
	systemInstaller, err := installer.NewSystemInstaller()
	if err != nil {
		// Log error but don't fail initialization
		fmt.Printf("Warning: Failed to initialize system installer: %v\n", err)
	}

	model := MenuModel{
		currentMenu:   MainMenu,
		menuStack:     []MenuType{},
		configManager: configManager,
		systemInstaller: systemInstaller,
		choices: []string{
			"Install Everything",
			"Update Everything",
			"List of Available Tools", 
			"Setup Environment",
			"Installation Configuration",
		},
		selected: make(map[int]struct{}),
		toolInstallStatus: make(map[string]bool),
		isLoading: false,
		loadingMessage: "",
		installationInProgress: false,
		installationResults: []InstallationResult{},
		showingResults: false,
		installEverythingMode: false,
		pendingEnvironments: []parser.Environment{},
		authError: "",
	}
	
	// Perform initial setup validation
	model = performInitialSetup(model)
	
	return model
}

// performInitialSetup handles the initial configuration and validation
func performInitialSetup(model MenuModel) MenuModel {
	credentials := model.configManager.GetCredentials()
	config := model.configManager.GetConfig()
	
	// Check if this is the first run
	if !model.configManager.IsConfigured() {
		model.authError = "Welcome to BOBA! Please configure your GitHub repository in 'Installation Configuration' to get started."
		return model
	}
	
	// Validate existing configuration
	if credentials.GitHubToken == "" {
		model.authError = "GitHub authentication required. Please set up your token in 'Installation Configuration'."
		return model
	}
	
	// Initialize GitHub integration
	model = initializeGitHubIntegration(model, credentials, config)
	
	return model
}

// initializeGitHubIntegration sets up GitHub client and related components
func initializeGitHubIntegration(model MenuModel, credentials config.Credentials, config config.Config) MenuModel {
	// Use configured repository URL or default to "boba-config"
	repoURL := config.RepositoryURL
	if repoURL == "" {
		repoURL = "boba-config"
	}
	
	// If repository URL doesn't contain "/", try to get username and prepend it
	if !strings.Contains(repoURL, "/") {
		model = resolveRepositoryURL(model, credentials.GitHubToken, repoURL)
		if model.authError != "" {
			return model
		}
		// Get updated config after URL resolution
		config = model.configManager.GetConfig()
		repoURL = config.RepositoryURL
	}
	
	// Parse repository URL to get owner and repo
	owner, repo, err := github.ParseRepositoryURL(repoURL)
	if err != nil {
		model.authError = fmt.Sprintf("Invalid repository URL format: %v", err)
		return model
	}
	
	// Initialize GitHub client
	model.githubClient = github.NewGitHubClient(credentials.GitHubToken, owner, repo)
	
	// Test connection and initialize components if successful
	if err := model.githubClient.TestConnection(); err != nil {
		model.authError = fmt.Sprintf("Repository access failed: %v\nPlease check your token and repository settings.", err)
		return model
	}
	
	// Initialize parser, installation engine, and dependency resolver
	model.repoParser = parser.NewRepositoryParser(model.githubClient)
	model.installEngine = installer.NewInstallationEngine(model.githubClient)
	model.dependencyResolver = installer.NewDependencyResolver()
	
	return model
}

// resolveRepositoryURL attempts to resolve a short repository name to full URL
func resolveRepositoryURL(model MenuModel, token, repoName string) MenuModel {
	// Create a temporary client to get the username
	tempClient := github.NewGitHubClient(token, "", "")
	authResult, err := tempClient.ValidateToken()
	
	if err != nil {
		model.authError = fmt.Sprintf("Token validation failed: %v", err)
		return model
	}
	
	if !authResult.Success {
		model.authError = fmt.Sprintf("Invalid GitHub token: %v", authResult.Error)
		return model
	}
	
	if authResult.User == nil || authResult.User.Login == nil {
		model.authError = "Unable to determine GitHub username from token"
		return model
	}
	
	username := *authResult.User.Login
	fullRepoURL := fmt.Sprintf("%s/%s", username, repoName)
	
	// Update the config with the corrected repository URL
	if err := model.configManager.SetRepositoryURL(fullRepoURL); err != nil {
		model.authError = fmt.Sprintf("Failed to save repository URL: %v", err)
		return model
	}
	
	return model
}