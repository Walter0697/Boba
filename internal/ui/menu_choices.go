package ui

import (
	"fmt"
)

// getMenuChoices returns the choices for the current menu
func (m MenuModel) getMenuChoices() []string {
	switch m.currentMenu {
	case MainMenu:
		if !m.isGitHubAuthenticated() {
			return []string{
				"Install Everything",
				"List of Available Tools", 
				"Setup Environment",
				"Installation Configuration",
				"🔧 Install BOBA to System",
				"🔐 GitHub Authentication",
			}
		} else {
			return []string{
				"Install Everything",
				"List of Available Tools", 
				"Setup Environment",
				"Installation Configuration",
				"🔧 Install BOBA to System",
			}
		}
	case InstallEverythingMenu:
		return m.getInstallEverythingChoices()
	case ToolsListMenu:
		return m.getToolsListChoices()
	case EnvironmentMenu:
		return m.getEnvironmentChoices()
	case ConfigurationMenu:
		return []string{
			"Repository Configuration",
			"Tool Override Management",
			"Environment Override Management",
			"← Back to Main Menu",
		}
	case RepositoryConfigMenu:
		return m.getRepositoryConfigChoices()
	case ToolOverrideMenu:
		return m.getToolOverrideChoices()
	case EnvironmentOverrideMenu:
		return m.getEnvironmentOverrideChoices()
	case GitHubAuthMenu:
		return []string{} // Auth model handles its own display
	case SystemInstallMenu:
		return m.getSystemInstallChoices()
	default:
		return []string{"← Back to Main Menu"}
	}
}

func (m MenuModel) getInstallEverythingChoices() []string {
	if m.isGitHubAuthenticated() {
		if m.installationInProgress {
			var progressInfo []string
			progressInfo = append(progressInfo, "🔄 " + m.loadingMessage)
			progressInfo = append(progressInfo, "Installation in progress...")
			
			// Show recent results if any
			if len(m.installationResults) > 0 {
				progressInfo = append(progressInfo, "")
				progressInfo = append(progressInfo, "📊 Recent Results:")
				// Show last few results
				start := len(m.installationResults) - 3
				if start < 0 {
					start = 0
				}
				for i := start; i < len(m.installationResults); i++ {
					result := m.installationResults[i]
					status := "✅"
					if !result.Success {
						status = "❌"
					}
					progressInfo = append(progressInfo, fmt.Sprintf("  %s %s", status, result.ToolName))
				}
			}
			
			progressInfo = append(progressInfo, "")
			progressInfo = append(progressInfo, "Press Ctrl+C to cancel")
			progressInfo = append(progressInfo, "← Back to Main Menu")
			return progressInfo
		} else if len(m.installationResults) > 0 {
			// Show installation results
			var choices []string
			choices = append(choices, "📊 Installation Results:")
			for _, result := range m.installationResults {
				status := "✅"
				if !result.Success {
					status = "❌"
				}
				choices = append(choices, fmt.Sprintf("  %s %s", status, result.ToolName))
			}
			choices = append(choices, "🔄 Run Installation Again")
			choices = append(choices, "🔄 Update Everything")
			choices = append(choices, "← Back to Main Menu")
			return choices
		} else {
			// Show what will be installed based on current configuration
			var description string
			config := m.configManager.GetConfig()
			if len(config.ToolOverrides) > 0 {
				enabledCount := 0
				for _, enabled := range config.ToolOverrides {
					if enabled {
						enabledCount++
					}
				}
				description = fmt.Sprintf("Will install tools based on your configuration (%d overrides set)", len(config.ToolOverrides))
			} else {
				description = "Will install all auto-install tools from your repository"
			}
			
			return []string{
				"🚀 Start Installation Process",
				"🔄 Update Everything",
				description,
				"← Back to Main Menu",
			}
		}
	} else {
		return []string{
			"🔐 Click to Authenticate with GitHub",
			m.getAuthenticationMessage(),
			"← Back to Main Menu",
		}
	}
}

func (m MenuModel) getToolsListChoices() []string {
	if m.isGitHubAuthenticated() {
		if m.isLoading {
			// Show loading state
			return []string{
				"🔄 " + m.loadingMessage,
				"Please wait...",
				"← Back to Main Menu",
			}
		} else if len(m.availableTools) > 0 {
			// Show the actual tools with installation status
			var choices []string
			for _, tool := range m.availableTools {
				var statusIcon string
				var autoIcon string
				
				// Check installation status from cache
				if installed, exists := m.toolInstallStatus[tool.Name]; exists && installed {
					statusIcon = "✅"
				} else {
					statusIcon = "⬜"
				}
				
				// Check auto-install setting
				if tool.AutoInstall {
					autoIcon = "⚡" // Auto-install tools get a lightning bolt
				} else {
					autoIcon = "🔧" // Manual-install tools get a wrench
				}
				
				toolDisplay := fmt.Sprintf("%s %s %s - %s", statusIcon, autoIcon, tool.Name, tool.Description)
				choices = append(choices, toolDisplay)
			}
			choices = append(choices, "🔄 Refresh Tools List")
			choices = append(choices, "← Back to Main Menu")
			return choices
		} else if m.loadingMessage != "" {
			// Show error state
			return []string{
				"❌ " + m.loadingMessage,
				"📋 Retry Fetching Tools",
				"← Back to Main Menu",
			}
		} else {
			// This case should not happen since we auto-fetch, but keep as fallback
			return []string{
				"🔄 Loading tools...",
				"Please wait while we fetch your tools",
				"← Back to Main Menu",
			}
		}
	} else {
		return []string{
			"🔐 Click to Authenticate with GitHub",
			m.getAuthenticationMessage(),
			"← Back to Main Menu",
		}
	}
}

func (m MenuModel) getEnvironmentChoices() []string {
	if m.isGitHubAuthenticated() {
		if m.isLoading {
			// Show loading state
			return []string{
				"🔄 " + m.loadingMessage,
				"Please wait...",
				"← Back to Main Menu",
			}
		} else if len(m.availableEnvironments) > 0 {
			// Show the actual environments with status
			var choices []string
			for _, env := range m.availableEnvironments {
				var autoIcon string
				
				// Check auto-apply setting
				if env.AutoApply {
					autoIcon = "⚡" // Auto-apply environments get a lightning bolt
				} else {
					autoIcon = "🔧" // Manual-apply environments get a wrench
				}
				
				shellIcon := "🐚" // Default shell icon
				if env.Shell != "" {
					switch env.Shell {
					case "zsh":
						shellIcon = "🦓"
					case "bash":
						shellIcon = "🐚"
					case "fish":
						shellIcon = "🐟"
					default:
						shellIcon = "💻"
					}
				}
				
				envDisplay := fmt.Sprintf("%s %s %s - %s", shellIcon, autoIcon, env.Name, env.Description)
				choices = append(choices, envDisplay)
			}
			choices = append(choices, "🔄 Refresh Environments List")
			choices = append(choices, "← Back to Main Menu")
			return choices
		} else if m.loadingMessage != "" {
			// Show error state
			return []string{
				"❌ " + m.loadingMessage,
				"📋 Retry Fetching Environments",
				"← Back to Main Menu",
			}
		} else {
			// This case should not happen since we auto-fetch, but keep as fallback
			return []string{
				"🔄 Loading environments...",
				"Please wait while we fetch your environment configurations",
				"← Back to Main Menu",
			}
		}
	} else {
		return []string{
			"🔐 Click to Authenticate with GitHub",
			m.getAuthenticationMessage(),
			"← Back to Main Menu",
		}
	}
}

func (m MenuModel) getRepositoryConfigChoices() []string {
	config := m.configManager.GetConfig()
	currentRepo := config.RepositoryURL
	if currentRepo == "" {
		currentRepo = "boba-config (default)"
	}
	return []string{
		fmt.Sprintf("Current Repository: %s", currentRepo),
		"Change Repository Name",
		"Reset to Default (boba-config)",
		"← Back to Configuration Menu",
	}
}

func (m MenuModel) getToolOverrideChoices() []string {
	if m.isGitHubAuthenticated() {
		if m.isLoading {
			// Show loading state
			return []string{
				"🔄 " + m.loadingMessage,
				"Please wait...",
				"← Back to Configuration Menu",
			}
		} else if len(m.availableTools) > 0 {
			// Show tools with override toggles
			var choices []string
			config := m.configManager.GetConfig()
			
			for _, tool := range m.availableTools {
				var statusIcon string
				var overrideStatus string
				
				// Check if there's an override for this tool
				if enabled, exists := config.ToolOverrides[tool.Name]; exists {
					if enabled {
						statusIcon = "✅"
						overrideStatus = "Enabled (Override)"
					} else {
						statusIcon = "❌"
						overrideStatus = "Disabled (Override)"
					}
				} else {
					// No override, use default auto_install setting
					if tool.AutoInstall {
						statusIcon = "⚡"
						overrideStatus = "Auto-install (Default)"
					} else {
						statusIcon = "🔧"
						overrideStatus = "Manual-install (Default)"
					}
				}
				
				toolDisplay := fmt.Sprintf("%s %s - %s", statusIcon, tool.Name, overrideStatus)
				choices = append(choices, toolDisplay)
			}
			choices = append(choices, "🔄 Refresh Tools List")
			choices = append(choices, "🔄 Reset All to Default")
			choices = append(choices, "← Back to Configuration Menu")
			return choices
		} else if m.loadingMessage != "" {
			// Show error state
			return []string{
				"❌ " + m.loadingMessage,
				"📋 Retry Fetching Tools",
				"← Back to Configuration Menu",
			}
		} else {
			// Auto-fetch tools when entering this menu
			return []string{
				"🔄 Loading tools...",
				"Please wait while we fetch your tools",
				"← Back to Configuration Menu",
			}
		}
	} else {
		return []string{
			"🔐 Click to Authenticate with GitHub",
			m.getAuthenticationMessage(),
			"← Back to Configuration Menu",
		}
	}
}

func (m MenuModel) getEnvironmentOverrideChoices() []string {
	if m.isGitHubAuthenticated() {
		if m.isLoading {
			// Show loading state
			return []string{
				"🔄 " + m.loadingMessage,
				"Please wait...",
				"← Back to Configuration Menu",
			}
		} else if len(m.availableEnvironments) > 0 {
			// Show environments with override toggles
			var choices []string
			config := m.configManager.GetConfig()
			
			for _, env := range m.availableEnvironments {
				var statusIcon string
				var overrideStatus string
				
				// Check if there's an override for this environment
				if enabled, exists := config.EnvironmentOverrides[env.Name]; exists {
					if enabled {
						statusIcon = "✅"
						overrideStatus = "Enabled (Override)"
					} else {
						statusIcon = "❌"
						overrideStatus = "Disabled (Override)"
					}
				} else {
					// No override, use default auto_apply setting
					if env.AutoApply {
						statusIcon = "⚡"
						overrideStatus = "Auto-apply (Default)"
					} else {
						statusIcon = "🔧"
						overrideStatus = "Manual-apply (Default)"
					}
				}
				
				// Add shell icon
				shellIcon := "🐚"
				switch env.Shell {
				case "zsh":
					shellIcon = "🦓"
				case "bash":
					shellIcon = "🐚"
				case "fish":
					shellIcon = "🐟"
				default:
					shellIcon = "💻"
				}
				
				envDisplay := fmt.Sprintf("%s %s %s - %s", statusIcon, shellIcon, env.Name, overrideStatus)
				choices = append(choices, envDisplay)
			}
			choices = append(choices, "🔄 Refresh Environments List")
			choices = append(choices, "🔄 Reset All to Default")
			choices = append(choices, "← Back to Configuration Menu")
			return choices
		} else if m.loadingMessage != "" {
			// Show error state
			return []string{
				"❌ " + m.loadingMessage,
				"📋 Retry Fetching Environments",
				"← Back to Configuration Menu",
			}
		} else {
			// Auto-fetch environments when entering this menu
			return []string{
				"🔄 Loading environments...",
				"Please wait while we fetch your environments",
				"← Back to Configuration Menu",
			}
		}
	} else {
		return []string{
			"🔐 Click to Authenticate with GitHub",
			m.getAuthenticationMessage(),
			"← Back to Configuration Menu",
		}
	}
}

// isGitHubAuthenticated checks if GitHub authentication is available
func (m *MenuModel) isGitHubAuthenticated() bool {
	if m.configManager == nil {
		return false
	}
	
	// If there's an authentication error, we're not authenticated
	if m.authError != "" {
		return false
	}
	
	credentials := m.configManager.GetCredentials()
	config := m.configManager.GetConfig()
	
	// If no repository URL is set, use the default "boba-config"
	repoURL := config.RepositoryURL
	if repoURL == "" {
		repoURL = "boba-config"
	}
	
	return credentials.GitHubToken != "" && repoURL != "" && m.githubClient != nil
}

// getAuthenticationMessage returns the appropriate authentication message
func (m *MenuModel) getAuthenticationMessage() string {
	if m.authError != "" {
		return fmt.Sprintf("❌ Authentication Error: %s", m.authError)
	}
	return "Authentication required to access your repository"
}

func (m MenuModel) getSystemInstallChoices() []string {
	if m.systemInstaller == nil {
		return []string{
			"❌ System installer not available",
			"System installation requires proper initialization",
			"← Back to Main Menu",
		}
	}
	
	if m.isLoading {
		return []string{
			"🔄 " + m.loadingMessage,
			"Please wait...",
			"← Back to Main Menu",
		}
	}
	
	if m.systemInstallResult != nil {
		// Show installation result
		var choices []string
		if m.systemInstallResult.Success {
			choices = append(choices, "✅ System Installation Successful!")
			choices = append(choices, fmt.Sprintf("📍 Binary installed to: %s", m.systemInstaller.GetInstallationInfo()["install_path"]))
			if m.systemInstallResult.ZshrcModified {
				choices = append(choices, "🐚 Shell integration configured")
			}
			choices = append(choices, "")
			choices = append(choices, m.systemInstallResult.Message)
			choices = append(choices, "")
			choices = append(choices, "🗑️ Uninstall from System")
		} else {
			choices = append(choices, "❌ System Installation Failed")
			if m.systemInstallResult.Error != nil {
				choices = append(choices, fmt.Sprintf("Error: %s", m.systemInstallResult.Error.Error()))
			}
			choices = append(choices, "")
			choices = append(choices, "🔄 Retry Installation")
		}
		choices = append(choices, "← Back to Main Menu")
		return choices
	}
	
	// Show installation options
	info := m.systemInstaller.GetInstallationInfo()
	isInstalled := info["is_installed"].(bool)
	requiresSudo := info["requires_sudo"].(bool)
	
	var choices []string
	
	if isInstalled {
		choices = append(choices, "✅ BOBA is already installed system-wide")
		choices = append(choices, fmt.Sprintf("📍 Location: %s", info["install_path"]))
		choices = append(choices, "")
		choices = append(choices, "🔄 Reinstall BOBA")
		choices = append(choices, "🗑️ Uninstall from System")
	} else {
		choices = append(choices, "🔧 Install BOBA to System")
		choices = append(choices, fmt.Sprintf("📍 Will install to: %s", info["install_path"]))
		if requiresSudo {
			choices = append(choices, "⚠️  Requires sudo privileges")
		}
		choices = append(choices, "🐚 Will configure zsh shell integration")
		choices = append(choices, "")
		choices = append(choices, "▶️ Start System Installation")
	}
	
	choices = append(choices, "ℹ️  View Installation Details")
	choices = append(choices, "← Back to Main Menu")
	
	return choices
}