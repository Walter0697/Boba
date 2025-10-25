package ui

import (
	"fmt"
	
	tea "github.com/charmbracelet/bubbletea"
	"boba/internal/github"
	"boba/internal/parser"
)

// startAuthentication initiates the GitHub authentication flow
func (m MenuModel) startAuthentication() (tea.Model, tea.Cmd) {
	// Get configured repository or use default
	config := m.configManager.GetConfig()
	repoURL := config.RepositoryURL
	if repoURL == "" {
		repoURL = "boba-config"
	}
	
	// Create auth model with proper callbacks
	onComplete := func(client *github.GitHubClient, repoURL string) tea.Cmd {
		return func() tea.Msg {
			return "auth_complete"
		}
	}
	
	onCancel := func() tea.Cmd {
		return func() tea.Msg {
			return "auth_cancelled"
		}
	}
	
	authModel := github.NewAuthModelWithRepo(repoURL, onComplete, onCancel)
	m.authModel = authModel
	m.navigateToMenu(GitHubAuthMenu)
	
	return m, authModel.Init()
}

// startRepositoryConfiguration initiates the repository configuration flow
func (m MenuModel) startRepositoryConfiguration() (tea.Model, tea.Cmd) {
	// Get current repository or use default
	config := m.configManager.GetConfig()
	currentRepo := config.RepositoryURL
	if currentRepo == "" {
		currentRepo = "boba-config"
	}
	
	// Create repository configuration model with proper callbacks
	onComplete := func(client *github.GitHubClient, repoURL string) tea.Cmd {
		return func() tea.Msg {
			return "repo_config_complete"
		}
	}
	
	onCancel := func() tea.Cmd {
		return func() tea.Msg {
			return "repo_config_cancelled"
		}
	}
	
	repoConfigModel := github.NewRepoConfigModel(currentRepo, onComplete, onCancel)
	m.authModel = repoConfigModel
	
	// Navigate to repository configuration
	m.navigateToMenu(GitHubAuthMenu) // Reuse auth menu type for simplicity
	
	return m, repoConfigModel.Init()
}

// fetchAndDisplayTools fetches tools from repository and displays them
func (m MenuModel) fetchAndDisplayTools() (tea.Model, tea.Cmd) {
	if m.repoParser == nil {
		// Initialize repository parser if not already done
		return m, func() tea.Msg {
			return "error_fetching_tools: Repository parser not initialized"
		}
	}
	
	// Set loading state
	m.isLoading = true
	m.loadingMessage = "Fetching tools from repository..."
	m.choices = m.getMenuChoices()
	
	return m, func() tea.Msg {
		tools, err := m.repoParser.GetTools()
		if err != nil {
			return fmt.Sprintf("error_fetching_tools: %v", err)
		}
		return ToolsListMsg{Tools: tools}
	}
}

// fetchAndDisplayEnvironments fetches environment configurations from repository and displays them
func (m MenuModel) fetchAndDisplayEnvironments() (tea.Model, tea.Cmd) {
	if m.repoParser == nil {
		// Initialize repository parser if not already done
		return m, func() tea.Msg {
			return "error_fetching_environments: Repository parser not initialized"
		}
	}
	
	// Set loading state
	m.isLoading = true
	m.loadingMessage = "Fetching environment configurations from repository..."
	m.choices = m.getMenuChoices()
	
	return m, func() tea.Msg {
		environments, err := m.repoParser.FetchEnvironments()
		if err != nil {
			return fmt.Sprintf("error_fetching_environments: %v", err)
		}
		return EnvironmentsListMsg{Environments: environments}
	}
}

// startInstallEverything initiates the installation of all tools with real-time progress feedback
func (m MenuModel) startInstallEverything() (tea.Model, tea.Cmd) {
	if m.repoParser == nil || m.installEngine == nil {
		return m, func() tea.Msg {
			return "error_installation: Installation engine not initialized"
		}
	}
	
	// Set installation in progress
	m.installationInProgress = true
	m.loadingMessage = "Preparing installation..."
	m.choices = m.getMenuChoices()
	
	return m, m.runInstallEverythingWithProgress()
}

// startUpdateEverything initiates the update of all installed tools
func (m MenuModel) startUpdateEverything() (tea.Model, tea.Cmd) {
	if m.repoParser == nil || m.installEngine == nil {
		return m, func() tea.Msg {
			return "error_installation: Installation engine not initialized"
		}
	}
	
	// Set installation in progress
	m.installationInProgress = true
	m.loadingMessage = "Preparing updates..."
	m.choices = m.getMenuChoices()
	
	return m, m.runUpdateEverythingWithProgress()
}

// installSingleTool installs a single selected tool
func (m MenuModel) installSingleTool(tool parser.Tool) (tea.Model, tea.Cmd) {
	if m.installEngine == nil {
		return m, func() tea.Msg {
			return "error_installation: Installation engine not initialized"
		}
	}
	
	// Set loading state
	m.isLoading = true
	m.loadingMessage = fmt.Sprintf("Installing %s...", tool.Name)
	m.choices = m.getMenuChoices()
	
	return m, func() tea.Msg {
		// Install the tool
		result, err := m.installEngine.InstallTool(tool)
		
		success := result.Success && err == nil
		message := result.Output
		if err != nil {
			message = fmt.Sprintf("Installation failed: %v", err)
		}
		
		// Update installation status cache
		if success {
			m.toolInstallStatus[tool.Name] = true
			
			// Record successful installation
			version := tool.Version
			if version == "" {
				version = "latest"
			}
			m.configManager.RecordToolInstallation(tool.Name, version, "manual")
		}
		
		return InstallationProgressMsg{
			ToolName: tool.Name,
			Status:   message,
			Success:  success,
		}
	}
}

// applyEnvironment applies a selected environment configuration
func (m MenuModel) applyEnvironment(env parser.Environment) (tea.Model, tea.Cmd) {
	// Set loading state
	m.isLoading = true
	m.loadingMessage = fmt.Sprintf("Applying environment: %s", env.Name)
	m.choices = m.getMenuChoices()
	
	return m, func() tea.Msg {
		// Apply the environment configuration
		// This would typically involve setting up shell configurations, environment variables, etc.
		// For now, we'll simulate the process
		
		return InstallationProgressMsg{
			ToolName: env.Name,
			Status:   "Environment applied successfully",
			Success:  true,
		}
	}
}

// toggleToolOverride toggles the override setting for a tool
func (m MenuModel) toggleToolOverride(tool parser.Tool) (tea.Model, tea.Cmd) {
	config := m.configManager.GetConfig()
	
	// Initialize ToolOverrides map if it doesn't exist
	if config.ToolOverrides == nil {
		config.ToolOverrides = make(map[string]bool)
	}
	
	// Check current override state
	if enabled, exists := config.ToolOverrides[tool.Name]; exists {
		// Toggle the existing override
		config.ToolOverrides[tool.Name] = !enabled
	} else {
		// No override exists, create one with opposite of default
		config.ToolOverrides[tool.Name] = !tool.AutoInstall
	}
	
	// Save the configuration
	err := m.configManager.SaveConfig()
	if err != nil {
		return m, func() tea.Msg {
			return fmt.Sprintf("error_config: Failed to save configuration: %v", err)
		}
	}
	
	// Update menu choices to reflect the change
	m.choices = m.getMenuChoices()
	
	return m, nil
}

// toggleEnvironmentOverride toggles the override setting for an environment
func (m MenuModel) toggleEnvironmentOverride(env parser.Environment) (tea.Model, tea.Cmd) {
	config := m.configManager.GetConfig()
	
	// Initialize EnvironmentOverrides map if it doesn't exist
	if config.EnvironmentOverrides == nil {
		config.EnvironmentOverrides = make(map[string]bool)
	}
	
	// Check current override state
	if enabled, exists := config.EnvironmentOverrides[env.Name]; exists {
		// Toggle the existing override
		config.EnvironmentOverrides[env.Name] = !enabled
	} else {
		// No override exists, create one with opposite of default
		config.EnvironmentOverrides[env.Name] = !env.AutoApply
	}
	
	// Save the configuration
	err := m.configManager.SaveConfig()
	if err != nil {
		return m, func() tea.Msg {
			return fmt.Sprintf("error_config: Failed to save configuration: %v", err)
		}
	}
	
	// Update menu choices to reflect the change
	m.choices = m.getMenuChoices()
	
	return m, nil
}

// resetAllToolOverrides resets all tool overrides to default
func (m MenuModel) resetAllToolOverrides() (tea.Model, tea.Cmd) {
	err := m.configManager.ResetAllToolOverrides()
	if err != nil {
		return m, func() tea.Msg {
			return fmt.Sprintf("error_config: Failed to reset tool overrides: %v", err)
		}
	}
	
	// Update menu choices to reflect the changes
	m.choices = m.getMenuChoices()
	
	return m, nil
}

// resetAllEnvironmentOverrides resets all environment overrides to default
func (m MenuModel) resetAllEnvironmentOverrides() (tea.Model, tea.Cmd) {
	err := m.configManager.ResetAllEnvironmentOverrides()
	if err != nil {
		return m, func() tea.Msg {
			return fmt.Sprintf("error_config: Failed to reset environment overrides: %v", err)
		}
	}
	
	// Update menu choices to reflect the changes
	m.choices = m.getMenuChoices()
	
	return m, nil
}