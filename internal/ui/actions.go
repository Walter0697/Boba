package ui

import (
	"fmt"
	"strings"
	
	tea "github.com/charmbracelet/bubbletea"
	"boba/internal/github"
	"boba/internal/installer"
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

// installSingleTool installs a single selected tool with dependency resolution
func (m MenuModel) installSingleTool(tool parser.Tool) (tea.Model, tea.Cmd) {
	if m.installEngine == nil || m.dependencyResolver == nil {
		return m, func() tea.Msg {
			return "error_installation: Installation engine not initialized"
		}
	}
	
	// Set loading state
	m.isLoading = true
	m.loadingMessage = fmt.Sprintf("Resolving dependencies for %s...", tool.Name)
	m.choices = m.getMenuChoices()
	
	return m, func() tea.Msg {
		// Get all available tools to resolve dependencies
		allTools, err := m.repoParser.GetTools()
		if err != nil {
			return InstallationProgressMsg{
				ToolName: tool.Name,
				Status:   fmt.Sprintf("Failed to fetch tools for dependency resolution: %v", err),
				Success:  false,
			}
		}
		
		// Find dependencies for this tool
		var toolsToInstall []parser.Tool
		toolMap := make(map[string]parser.Tool)
		for _, t := range allTools {
			toolMap[t.Name] = t
		}
		
		// Collect the tool and its dependencies
		var collectDependencies func(toolName string, visited map[string]bool) error
		collectDependencies = func(toolName string, visited map[string]bool) error {
			if visited[toolName] {
				return nil // Already processed
			}
			
			currentTool, exists := toolMap[toolName]
			if !exists {
				return fmt.Errorf("dependency not found: %s", toolName)
			}
			
			visited[toolName] = true
			
			// First, collect dependencies
			for _, dep := range currentTool.Dependencies {
				if err := collectDependencies(dep, visited); err != nil {
					return err
				}
			}
			
			// Then add the tool itself if not already installed
			if !m.installEngine.IsToolInstalled(currentTool) {
				toolsToInstall = append(toolsToInstall, currentTool)
			}
			
			return nil
		}
		
		visited := make(map[string]bool)
		if err := collectDependencies(tool.Name, visited); err != nil {
			return InstallationProgressMsg{
				ToolName: tool.Name,
				Status:   fmt.Sprintf("Dependency resolution failed: %v", err),
				Success:  false,
			}
		}
		
		// If no tools need to be installed, the tool is already installed
		if len(toolsToInstall) == 0 {
			return InstallationProgressMsg{
				ToolName: tool.Name,
				Status:   fmt.Sprintf("%s is already installed", tool.Name),
				Success:  true,
			}
		}
		
		// Install tools in dependency order
		var results []string
		for _, toolToInstall := range toolsToInstall {
			result, err := m.installEngine.InstallTool(toolToInstall)
			
			success := result.Success && err == nil
			if success {
				m.toolInstallStatus[toolToInstall.Name] = true
				
				// Record successful installation
				version := toolToInstall.Version
				if version == "" {
					version = "latest"
				}
				m.configManager.RecordToolInstallation(toolToInstall.Name, version, "manual")
				results = append(results, fmt.Sprintf("✓ %s installed successfully", toolToInstall.Name))
			} else {
				message := result.Output
				if err != nil {
					message = fmt.Sprintf("Installation failed: %v", err)
				}
				results = append(results, fmt.Sprintf("✗ %s failed: %s", toolToInstall.Name, message))
				
				// If a dependency fails, stop the installation
				return InstallationProgressMsg{
					ToolName: tool.Name,
					Status:   fmt.Sprintf("Installation failed due to dependency failure:\n%s", strings.Join(results, "\n")),
					Success:  false,
				}
			}
		}
		
		return InstallationProgressMsg{
			ToolName: tool.Name,
			Status:   fmt.Sprintf("Installation completed:\n%s", strings.Join(results, "\n")),
			Success:  true,
		}
	}
}

// applyEnvironment applies a selected environment configuration with dependency resolution
func (m MenuModel) applyEnvironment(env parser.Environment) (tea.Model, tea.Cmd) {
	if m.installEngine == nil || m.dependencyResolver == nil {
		return m, func() tea.Msg {
			return InstallationProgressMsg{
				ToolName: env.Name,
				Status:   "Installation engine not initialized",
				Success:  false,
			}
		}
	}
	
	// Set loading state
	m.isLoading = true
	m.loadingMessage = fmt.Sprintf("Resolving dependencies for environment: %s", env.Name)
	m.choices = m.getMenuChoices()
	
	return m, func() tea.Msg {
		// Get all available environments to resolve dependencies
		allEnvironments, err := m.repoParser.FetchEnvironments()
		if err != nil {
			return InstallationProgressMsg{
				ToolName: env.Name,
				Status:   fmt.Sprintf("Failed to fetch environments for dependency resolution: %v", err),
				Success:  false,
			}
		}
		
		// Find dependencies for this environment
		var environmentsToApply []parser.Environment
		envMap := make(map[string]parser.Environment)
		for _, e := range allEnvironments {
			envMap[e.Name] = e
		}
		
		// Collect the environment and its dependencies
		var collectDependencies func(envName string, visited map[string]bool) error
		collectDependencies = func(envName string, visited map[string]bool) error {
			if visited[envName] {
				return nil // Already processed
			}
			
			currentEnv, exists := envMap[envName]
			if !exists {
				return fmt.Errorf("environment dependency not found: %s", envName)
			}
			
			visited[envName] = true
			
			// First, collect dependencies
			for _, dep := range currentEnv.Dependencies {
				if err := collectDependencies(dep, visited); err != nil {
					return err
				}
			}
			
			// Then add the environment itself if not already applied
			if !m.installEngine.IsEnvironmentApplied(currentEnv) {
				environmentsToApply = append(environmentsToApply, currentEnv)
			}
			
			return nil
		}
		
		visited := make(map[string]bool)
		if err := collectDependencies(env.Name, visited); err != nil {
			return InstallationProgressMsg{
				ToolName: env.Name,
				Status:   fmt.Sprintf("Environment dependency resolution failed: %v", err),
				Success:  false,
			}
		}
		
		// If no environments need to be applied, the environment is already applied
		if len(environmentsToApply) == 0 {
			return InstallationProgressMsg{
				ToolName: env.Name,
				Status:   fmt.Sprintf("Environment '%s' is already applied", env.Name),
				Success:  true,
			}
		}
		
		// Apply environments in dependency order
		var results []string
		for _, envToApply := range environmentsToApply {
			result, err := m.installEngine.ApplyEnvironment(envToApply)
			
			success := result.Success && err == nil
			if success {
				// Record successful application
				results = append(results, fmt.Sprintf("✓ %s applied successfully", envToApply.Name))
			} else {
				message := result.Output
				if err != nil {
					message = fmt.Sprintf("Application failed: %v", err)
				}
				results = append(results, fmt.Sprintf("✗ %s failed: %s", envToApply.Name, message))
				
				// If a dependency fails, stop the application
				return InstallationProgressMsg{
					ToolName: env.Name,
					Status:   fmt.Sprintf("Environment application failed due to dependency failure:\n%s", strings.Join(results, "\n")),
					Success:  false,
				}
			}
		}
		
		return InstallationProgressMsg{
			ToolName: env.Name,
			Status:   fmt.Sprintf("Environment application completed:\n%s", strings.Join(results, "\n")),
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

// startSystemInstallation initiates the system installation process
func (m MenuModel) startSystemInstallation() (tea.Model, tea.Cmd) {
	if m.systemInstaller == nil {
		return m, func() tea.Msg {
			return SystemInstallationCompleteMsg{
				Result: &installer.SystemInstallationResult{
					Success: false,
					Error:   fmt.Errorf("system installer not available"),
					Message: "System installer could not be initialized",
				},
			}
		}
	}
	
	// Set loading state
	m.isLoading = true
	m.loadingMessage = "Installing BOBA to system..."
	m.systemInstallResult = nil // Clear any previous result
	m.choices = m.getMenuChoices()
	
	return m, func() tea.Msg {
		result, err := m.systemInstaller.InstallToSystem()
		if err != nil && result == nil {
			result = &installer.SystemInstallationResult{
				Success: false,
				Error:   err,
				Message: "System installation failed",
			}
		}
		return SystemInstallationCompleteMsg{Result: result}
	}
}

// startSystemUninstallation initiates the system uninstallation process
func (m MenuModel) startSystemUninstallation() (tea.Model, tea.Cmd) {
	if m.systemInstaller == nil {
		return m, func() tea.Msg {
			return SystemInstallationCompleteMsg{
				Result: &installer.SystemInstallationResult{
					Success: false,
					Error:   fmt.Errorf("system installer not available"),
					Message: "System installer could not be initialized",
				},
			}
		}
	}
	
	// Set loading state
	m.isLoading = true
	m.loadingMessage = "Uninstalling BOBA from system..."
	m.systemInstallResult = nil // Clear any previous result
	m.choices = m.getMenuChoices()
	
	return m, func() tea.Msg {
		result, err := m.systemInstaller.UninstallFromSystem()
		if err != nil && result == nil {
			result = &installer.SystemInstallationResult{
				Success: false,
				Error:   err,
				Message: "System uninstallation failed",
			}
		}
		return SystemInstallationCompleteMsg{Result: result}
	}
}

// showSystemInstallationDetails shows detailed information about system installation
func (m MenuModel) showSystemInstallationDetails() (tea.Model, tea.Cmd) {
	if m.systemInstaller == nil {
		return m, nil
	}
	
	// For now, just refresh the menu to show current status
	// In a more advanced implementation, this could show a detailed view
	m.choices = m.getMenuChoices()
	return m, nil
}