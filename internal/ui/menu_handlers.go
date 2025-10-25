package ui

import (
	"strings"
	
	tea "github.com/charmbracelet/bubbletea"
)

// handleMenuSelection handles menu item selection
func (m MenuModel) handleMenuSelection() (tea.Model, tea.Cmd) {
	switch m.currentMenu {
	case MainMenu:
		return m.handleMainMenuSelection()
	case InstallEverythingMenu, ToolsListMenu, EnvironmentMenu, ToolOverrideMenu, EnvironmentOverrideMenu:
		return m.handleComplexMenuSelection()
	case ConfigurationMenu:
		return m.handleConfigurationMenuSelection()
	case RepositoryConfigMenu:
		return m.handleRepositoryConfigMenuSelection()
	case SystemInstallMenu:
		return m.handleSystemInstallMenuSelection()
	}
	return m, nil
}

func (m MenuModel) handleMainMenuSelection() (tea.Model, tea.Cmd) {
	if !m.isGitHubAuthenticated() {
		// When not authenticated, handle the extra auth option
		switch m.cursor {
		case 0:
			// Install Everything - navigate to submenu
			m.navigateToMenu(InstallEverythingMenu)
		case 1:
			// List of Available Tools - navigate to submenu
			m.navigateToMenu(ToolsListMenu)
		case 2:
			// Setup Environment
			m.navigateToMenu(EnvironmentMenu)
		case 3:
			// Installation Configuration
			m.navigateToMenu(ConfigurationMenu)
		case 4:
			// Install BOBA to System
			m.navigateToMenu(SystemInstallMenu)
		case 5:
			// GitHub Authentication
			return m.startAuthentication()
		}
	} else {
		// When authenticated, normal menu
		switch m.cursor {
		case 0:
			// Install Everything - navigate to submenu
			m.navigateToMenu(InstallEverythingMenu)
		case 1:
			// List of Available Tools - navigate to submenu and auto-fetch
			m.navigateToMenu(ToolsListMenu)
			return m.fetchAndDisplayTools()
		case 2:
			// Setup Environment - navigate to submenu and auto-fetch
			m.navigateToMenu(EnvironmentMenu)
			return m.fetchAndDisplayEnvironments()
		case 3:
			// Installation Configuration
			m.navigateToMenu(ConfigurationMenu)
		case 4:
			// Install BOBA to System
			m.navigateToMenu(SystemInstallMenu)
		}
	}
	return m, nil
}

func (m MenuModel) handleConfigurationMenuSelection() (tea.Model, tea.Cmd) {
	currentChoices := m.getMenuChoices()
	if m.cursor == len(currentChoices)-1 {
		// Back option
		m.navigateBack()
	} else {
		switch m.cursor {
		case 0:
			// Repository Configuration
			m.navigateToMenu(RepositoryConfigMenu)
		case 1:
			// Tool Override Management - navigate to submenu and auto-fetch
			m.navigateToMenu(ToolOverrideMenu)
			if m.isGitHubAuthenticated() {
				return m.fetchAndDisplayTools()
			}
		case 2:
			// Environment Override Management - navigate to submenu and auto-fetch
			m.navigateToMenu(EnvironmentOverrideMenu)
			if m.isGitHubAuthenticated() {
				return m.fetchAndDisplayEnvironments()
			}
		}
	}
	return m, nil
}

func (m MenuModel) handleRepositoryConfigMenuSelection() (tea.Model, tea.Cmd) {
	currentChoices := m.getMenuChoices()
	if m.cursor == len(currentChoices)-1 {
		// Back option
		m.navigateBack()
	} else {
		switch m.cursor {
		case 0:
			// Current Repository - just display, no action
			break
		case 1:
			// Change Repository Name
			return m.startRepositoryConfiguration()
		case 2:
			// Reset to Default
			m.configManager.SetRepositoryURL("boba-config")
			// Refresh the menu to show updated repository
			m.choices = m.getMenuChoices()
		}
	}
	return m, nil
}

func (m MenuModel) handleComplexMenuSelection() (tea.Model, tea.Cmd) {
	currentChoices := m.getMenuChoices()
	if m.cursor == len(currentChoices)-1 {
		// Back option
		m.navigateBack()
	} else if m.cursor == 0 {
		// First option - either authenticate or perform action
		if !m.isGitHubAuthenticated() {
			return m.startAuthentication()
		}
		// Handle different menu actions for authenticated users
		return m.handleFirstOptionForAuthenticatedUser()
	} else if m.cursor == 1 && !m.isGitHubAuthenticated() {
		// Second option when not authenticated - also trigger auth
		return m.startAuthentication()
	} else {
		// Handle other menu-specific options
		return m.handleMenuSpecificOptions()
	}
	return m, nil
}

func (m MenuModel) handleFirstOptionForAuthenticatedUser() (tea.Model, tea.Cmd) {
	if m.currentMenu == ToolsListMenu {
		if m.isLoading {
			// Don't allow action while loading
			return m, nil
		}
		// Since we auto-fetch, cursor 0 should be first tool or error retry
		if len(m.availableTools) > 0 {
			// First tool selection
			selectedTool := m.availableTools[0]
			return m.installSingleTool(selectedTool)
		} else if m.loadingMessage != "" {
			// Retry fetching tools on error
			m.loadingMessage = "" // Clear error message
			return m.fetchAndDisplayTools()
		}
		return m, nil
	} else if m.currentMenu == EnvironmentMenu {
		if m.isLoading {
			// Don't allow action while loading
			return m, nil
		}
		// Since we auto-fetch, cursor 0 should be first environment or error retry
		if len(m.availableEnvironments) > 0 {
			// First environment selection
			selectedEnv := m.availableEnvironments[0]
			return m.applyEnvironment(selectedEnv)
		} else if m.loadingMessage != "" {
			// Retry fetching environments on error
			m.loadingMessage = "" // Clear error message
			return m.fetchAndDisplayEnvironments()
		}
		return m, nil
	} else if m.currentMenu == ToolOverrideMenu {
		if m.isLoading {
			// Don't allow action while loading
			return m, nil
		}
		// Since we auto-fetch, cursor 0 should be first tool or error retry
		if len(m.availableTools) > 0 {
			// First tool override toggle
			selectedTool := m.availableTools[0]
			return m.toggleToolOverride(selectedTool)
		} else if m.loadingMessage != "" {
			// Retry fetching tools on error
			m.loadingMessage = "" // Clear error message
			return m.fetchAndDisplayTools()
		}
		return m, nil
	} else if m.currentMenu == EnvironmentOverrideMenu {
		if m.isLoading {
			// Don't allow action while loading
			return m, nil
		}
		// Since we auto-fetch, cursor 0 should be first environment or error retry
		if len(m.availableEnvironments) > 0 {
			// First environment override toggle
			selectedEnv := m.availableEnvironments[0]
			return m.toggleEnvironmentOverride(selectedEnv)
		} else if m.loadingMessage != "" {
			// Retry fetching environments on error
			m.loadingMessage = "" // Clear error message
			return m.fetchAndDisplayEnvironments()
		}
		return m, nil
	} else if m.currentMenu == InstallEverythingMenu {
		if m.installationInProgress {
			// Don't allow action while installation is in progress
			return m, nil
		}
		return m.startInstallEverything()
	}
	return m, nil
}

func (m MenuModel) handleMenuSpecificOptions() (tea.Model, tea.Cmd) {
	if m.currentMenu == InstallEverythingMenu && m.isGitHubAuthenticated() {
		return m.handleInstallEverythingOptions()
	} else if m.currentMenu == ToolsListMenu && m.isGitHubAuthenticated() {
		return m.handleToolsListOptions()
	} else if m.currentMenu == EnvironmentMenu && m.isGitHubAuthenticated() {
		return m.handleEnvironmentOptions()
	} else if m.currentMenu == ToolOverrideMenu && m.isGitHubAuthenticated() {
		return m.handleToolOverrideOptions()
	} else if m.currentMenu == EnvironmentOverrideMenu && m.isGitHubAuthenticated() {
		return m.handleEnvironmentOverrideOptions()
	}
	return m, nil
}

func (m MenuModel) handleInstallEverythingOptions() (tea.Model, tea.Cmd) {
	// Handle InstallEverythingMenu options
	currentChoices := m.getMenuChoices()
	if len(m.installationResults) > 0 {
		// When results are shown, check for options
		if m.cursor == len(currentChoices)-3 { // "Run Installation Again"
			m.installationResults = []InstallationResult{} // Clear previous results
			return m.startInstallEverything()
		} else if m.cursor == len(currentChoices)-2 { // "Update Everything"
			m.installationResults = []InstallationResult{} // Clear previous results
			return m.startUpdateEverything()
		}
	} else {
		// When no results are shown, check for Update Everything option
		if m.cursor == 1 { // "Update Everything"
			return m.startUpdateEverything()
		}
	}
	return m, nil
}

func (m MenuModel) handleToolsListOptions() (tea.Model, tea.Cmd) {
	// Handle other options in tools list menu
	currentChoices := m.getMenuChoices()
	if len(m.availableTools) > 0 {
		// When tools are loaded, check for refresh option
		if m.cursor == len(currentChoices)-2 { // "Refresh Tools List"
			// Clear the cache when refreshing
			m.toolInstallStatus = make(map[string]bool)
			return m.fetchAndDisplayTools()
		} else if m.cursor < len(m.availableTools) {
			// Individual tool selection
			selectedTool := m.availableTools[m.cursor]
			return m.installSingleTool(selectedTool)
		}
	} else if m.loadingMessage != "" && m.cursor == 1 { // "Retry Fetching Tools"
		m.loadingMessage = "" // Clear error message
		return m.fetchAndDisplayTools()
	}
	return m, nil
}

func (m MenuModel) handleEnvironmentOptions() (tea.Model, tea.Cmd) {
	// Handle other options in environment menu
	currentChoices := m.getMenuChoices()
	if len(m.availableEnvironments) > 0 {
		// When environments are loaded, check for refresh option
		if m.cursor == len(currentChoices)-2 { // "Refresh Environments List"
			return m.fetchAndDisplayEnvironments()
		} else if m.cursor < len(m.availableEnvironments) {
			// Individual environment selection
			selectedEnv := m.availableEnvironments[m.cursor]
			return m.applyEnvironment(selectedEnv)
		}
	} else if m.loadingMessage != "" && m.cursor == 1 { // "Retry Fetching Environments"
		m.loadingMessage = "" // Clear error message
		return m.fetchAndDisplayEnvironments()
	}
	return m, nil
}

func (m MenuModel) handleToolOverrideOptions() (tea.Model, tea.Cmd) {
	// Handle tool override menu options
	currentChoices := m.getMenuChoices()
	if len(m.availableTools) > 0 {
		// When tools are loaded, check for special options
		if m.cursor == len(currentChoices)-3 { // "Refresh Tools List"
			return m.fetchAndDisplayTools()
		} else if m.cursor == len(currentChoices)-2 { // "Reset All to Default"
			return m.resetAllToolOverrides()
		} else if m.cursor < len(m.availableTools) {
			// Individual tool override toggle
			selectedTool := m.availableTools[m.cursor]
			return m.toggleToolOverride(selectedTool)
		}
	} else if m.loadingMessage != "" && m.cursor == 1 { // "Retry Fetching Tools"
		m.loadingMessage = "" // Clear error message
		return m.fetchAndDisplayTools()
	}
	return m, nil
}

func (m MenuModel) handleEnvironmentOverrideOptions() (tea.Model, tea.Cmd) {
	// Handle environment override menu options
	currentChoices := m.getMenuChoices()
	if len(m.availableEnvironments) > 0 {
		// When environments are loaded, check for special options
		if m.cursor == len(currentChoices)-3 { // "Refresh Environments List"
			return m.fetchAndDisplayEnvironments()
		} else if m.cursor == len(currentChoices)-2 { // "Reset All to Default"
			return m.resetAllEnvironmentOverrides()
		} else if m.cursor < len(m.availableEnvironments) {
			// Individual environment override toggle
			selectedEnv := m.availableEnvironments[m.cursor]
			return m.toggleEnvironmentOverride(selectedEnv)
		}
	} else if m.loadingMessage != "" && m.cursor == 1 { // "Retry Fetching Environments"
		m.loadingMessage = "" // Clear error message
		return m.fetchAndDisplayEnvironments()
	}
	return m, nil
}

func (m MenuModel) handleSystemInstallMenuSelection() (tea.Model, tea.Cmd) {
	currentChoices := m.getMenuChoices()
	if m.cursor == len(currentChoices)-1 {
		// Back option
		m.navigateBack()
		return m, nil
	}
	
	if m.systemInstaller == nil {
		// System installer not available
		return m, nil
	}
	
	if m.isLoading {
		// Don't allow actions while loading
		return m, nil
	}
	
	if m.systemInstallResult != nil {
		// Handle post-installation options
		if m.systemInstallResult.Success {
			// Successful installation - check for uninstall option
			if strings.Contains(currentChoices[m.cursor], "Uninstall from System") {
				return m.startSystemUninstallation()
			}
		} else {
			// Failed installation - check for retry option
			if strings.Contains(currentChoices[m.cursor], "Retry Installation") {
				m.systemInstallResult = nil // Clear previous result
				return m.startSystemInstallation()
			}
		}
		return m, nil
	}
	
	// Handle installation options
	info := m.systemInstaller.GetInstallationInfo()
	isInstalled := info["is_installed"].(bool)
	
	if isInstalled {
		// Already installed - handle reinstall or uninstall
		switch {
		case strings.Contains(currentChoices[m.cursor], "Reinstall BOBA"):
			return m.startSystemInstallation()
		case strings.Contains(currentChoices[m.cursor], "Uninstall from System"):
			return m.startSystemUninstallation()
		case strings.Contains(currentChoices[m.cursor], "View Installation Details"):
			return m.showSystemInstallationDetails()
		}
	} else {
		// Not installed - handle installation
		switch {
		case strings.Contains(currentChoices[m.cursor], "Start System Installation"):
			return m.startSystemInstallation()
		case strings.Contains(currentChoices[m.cursor], "View Installation Details"):
			return m.showSystemInstallationDetails()
		}
	}
	
	return m, nil
}