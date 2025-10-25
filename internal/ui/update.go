package ui

import (
	"fmt"
	"strings"
	
	tea "github.com/charmbracelet/bubbletea"
	"boba/internal/installer"
	"boba/internal/parser"
)

// Update handles user input and updates the model
func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// First handle authentication completion messages regardless of current menu
	if strMsg, ok := msg.(string); ok {
		switch strMsg {
		case "auth_complete":
			// Authentication successful, save credentials and initialize components
			if m.authModel != nil && m.authModel.GetClient() != nil {
				client := m.authModel.GetClient()
				token := m.authModel.GetToken()
				repoURL := m.authModel.GetRepoURL()
				
				// Save the GitHub token to credentials
				if err := m.configManager.SetGitHubToken(token); err != nil {
					// Handle error but don't fail completely
					fmt.Printf("Warning: Failed to save GitHub token: %v\n", err)
				}
				
				// Save the repository URL to config
				if err := m.configManager.SetRepositoryURL(repoURL); err != nil {
					// Handle error but don't fail completely
					fmt.Printf("Warning: Failed to save repository URL: %v\n", err)
				}
				
				// Set the GitHub client and initialize components
				m.githubClient = client
				m.repoParser = parser.NewRepositoryParser(m.githubClient)
				m.installEngine = installer.NewInstallationEngine(m.githubClient)
			}
			m.currentMenu = MainMenu
			m.menuStack = []MenuType{} // Clear the stack
			m.choices = m.getMenuChoices()
			m.cursor = 0
			return m, nil
		case "auth_cancelled":
			// Authentication cancelled, go back to main menu
			m.currentMenu = MainMenu
			m.menuStack = []MenuType{} // Clear the stack
			m.choices = m.getMenuChoices()
			m.cursor = 0
			return m, nil
		case "repo_config_complete":
			// Repository configuration complete, go back to repository config menu
			m.currentMenu = RepositoryConfigMenu
			m.menuStack = []MenuType{ConfigurationMenu} // Set proper navigation stack
			m.choices = m.getMenuChoices()
			m.cursor = 0
			return m, nil
		case "repo_config_cancelled":
			// Repository configuration cancelled, go back to repository config menu
			m.currentMenu = RepositoryConfigMenu
			m.menuStack = []MenuType{ConfigurationMenu} // Set proper navigation stack
			m.choices = m.getMenuChoices()
			m.cursor = 0
			return m, nil
		}
	}

	// Handle tools list message
	if toolsMsg, ok := msg.(ToolsListMsg); ok {
		m.availableTools = toolsMsg.Tools
		m.isLoading = false
		m.loadingMessage = ""
		
		// Cache installation status for all tools to avoid repeated system calls
		m.toolInstallStatus = make(map[string]bool)
		if m.installEngine != nil {
			for _, tool := range toolsMsg.Tools {
				m.toolInstallStatus[tool.Name] = m.installEngine.IsToolInstalled(tool)
			}
		}
		
		// Update the menu choices to show the tools
		m.choices = m.getMenuChoices()
		return m, nil
	}

	// Handle environments list message
	if envMsg, ok := msg.(EnvironmentsListMsg); ok {
		m.availableEnvironments = envMsg.Environments
		m.isLoading = false
		m.loadingMessage = ""
		
		// Update the menu choices to show the environments
		m.choices = m.getMenuChoices()
		return m, nil
	}

	// Handle installation progress messages
	if progressMsg, ok := msg.(InstallationProgressMsg); ok {
		// Update installation status cache if installation was successful
		if progressMsg.Success {
			m.toolInstallStatus[progressMsg.ToolName] = true
			
			// Record successful installation for single tool installs
			// Find the tool to get its version
			for _, tool := range m.availableTools {
				if tool.Name == progressMsg.ToolName {
					version := tool.Version
					if version == "" {
						version = "latest"
					}
					m.configManager.RecordToolInstallation(tool.Name, version, "manual")
					break
				}
			}
		}
		
		// Create a result to display
		result := InstallationResult{
			ToolName: progressMsg.ToolName,
			Success:  progressMsg.Success,
			Message:  progressMsg.Status,
		}
		
		// Show results screen instead of immediately returning to menu
		m.isLoading = false
		m.installationInProgress = false
		m.showingResults = true
		m.installationResults = []InstallationResult{result}
		m.loadingMessage = "" // Clear loading message
		m.choices = m.getMenuChoices()
		return m, nil
	}

	// Handle install everything phase messages
	if phaseMsg, ok := msg.(InstallEverythingPhaseMsg); ok {
		if phaseMsg.Phase == "tools" {
			// Set Install Everything mode and store pending environments
			m.installEverythingMode = true
			m.pendingEnvironments = phaseMsg.Environments
			
			if len(phaseMsg.Tools) > 0 {
				// Start installing tools
				m.loadingMessage = "Installing tools..."
				return m, m.installNextTool(phaseMsg.Tools, 0, []InstallationResult{})
			} else {
				// No tools to install, move to environments phase
				return m, func() tea.Msg {
					return InstallEverythingPhaseMsg{
						Phase:        "environments",
						Tools:        phaseMsg.Tools,
						Environments: phaseMsg.Environments,
					}
				}
			}
		} else if phaseMsg.Phase == "environments" {
			if len(phaseMsg.Environments) > 0 {
				// Start applying environments
				m.loadingMessage = "Applying environment configurations..."
				return m, m.applyNextEnvironment(phaseMsg.Environments, 0, []EnvironmentApplicationResult{})
			} else {
				// No environments to apply, complete the process
				m.installEverythingMode = false
				return m, func() tea.Msg {
					return InstallationCompleteMsg{
						Results: []InstallationResult{{
							ToolName: "Complete",
							Success:  true,
							Message:  "Installation completed successfully",
						}},
					}
				}
			}
		}
	}

	// Handle installation start messages
	if startMsg, ok := msg.(InstallationStartMsg); ok {
		// Start installing the first tool
		return m, m.installNextTool(startMsg.Tools, startMsg.CurrentIndex, startMsg.Results)
	}

	// Handle installation next messages
	if nextMsg, ok := msg.(InstallationNextMsg); ok {
		// Update the accumulated results
		m.installationResults = nextMsg.Results
		
		// Update progress display before continuing
		if nextMsg.CurrentIndex < len(nextMsg.Tools) {
			currentTool := nextMsg.Tools[nextMsg.CurrentIndex]
			m.loadingMessage = fmt.Sprintf("Installing %s (%d/%d)", currentTool.Name, nextMsg.CurrentIndex+1, len(nextMsg.Tools))
			m.choices = m.getMenuChoices()
		}
		// Continue with the next tool
		return m, m.installNextTool(nextMsg.Tools, nextMsg.CurrentIndex, nextMsg.Results)
	}

	// Handle environment application next messages
	if nextMsg, ok := msg.(EnvironmentApplicationNextMsg); ok {
		// Update progress display before continuing
		if nextMsg.CurrentIndex < len(nextMsg.Environments) {
			currentEnv := nextMsg.Environments[nextMsg.CurrentIndex]
			m.loadingMessage = fmt.Sprintf("Applying %s (%d/%d)", currentEnv.Name, nextMsg.CurrentIndex+1, len(nextMsg.Environments))
			m.choices = m.getMenuChoices()
		}
		// Continue with the next environment
		return m, m.applyNextEnvironment(nextMsg.Environments, nextMsg.CurrentIndex, nextMsg.Results)
	}

	// Handle system installation completion messages
	if sysCompleteMsg, ok := msg.(SystemInstallationCompleteMsg); ok {
		m.isLoading = false
		m.systemInstallResult = sysCompleteMsg.Result
		
		if sysCompleteMsg.Result.Success {
			m.loadingMessage = "System installation completed successfully"
		} else {
			m.loadingMessage = "System installation failed"
		}
		
		m.choices = m.getMenuChoices()
		return m, nil
	}

	// Handle installation completion messages
	if completeMsg, ok := msg.(InstallationCompleteMsg); ok {
		// Check if we're in Install Everything mode and have pending environments
		if m.installEverythingMode && len(m.pendingEnvironments) > 0 {
			// Store tool results and move to environments phase
			m.installationResults = completeMsg.Results
			
			// Update installation status cache based on results
			for _, result := range completeMsg.Results {
				if result.Success {
					m.toolInstallStatus[result.ToolName] = true
				}
			}
			
			// Move to environments phase
			return m, func() tea.Msg {
				return InstallEverythingPhaseMsg{
					Phase:        "environments",
					Tools:        []parser.Tool{}, // Tools already processed
					Environments: m.pendingEnvironments,
				}
			}
		}
		
		// Normal completion (not Install Everything mode or no pending environments)
		m.installationInProgress = false
		m.installationResults = completeMsg.Results
		m.isLoading = false
		m.showingResults = true // Show results screen instead of immediately returning to menu
		m.loadingMessage = "" // Clear loading message
		m.installEverythingMode = false
		m.pendingEnvironments = nil
		
		// Update installation status cache based on results
		for _, result := range completeMsg.Results {
			if result.Success {
				m.toolInstallStatus[result.ToolName] = true
			}
		}
		
		m.choices = m.getMenuChoices()
		return m, nil
	}

	// Handle error messages
	if errMsg, ok := msg.(string); ok && strings.HasPrefix(errMsg, "error_fetching_tools:") {
		// Show error and reset loading state
		errorText := strings.TrimPrefix(errMsg, "error_fetching_tools: ")
		m.isLoading = false
		m.loadingMessage = fmt.Sprintf("Error: %s", errorText)
		m.choices = m.getMenuChoices()
		return m, nil
	}

	// Handle environment fetching error messages
	if errMsg, ok := msg.(string); ok && strings.HasPrefix(errMsg, "error_fetching_environments:") {
		// Show error and reset loading state
		errorText := strings.TrimPrefix(errMsg, "error_fetching_environments: ")
		m.isLoading = false
		m.loadingMessage = fmt.Sprintf("Error: %s", errorText)
		m.choices = m.getMenuChoices()
		return m, nil
	}

	// Handle installation error messages
	if errMsg, ok := msg.(string); ok && strings.HasPrefix(errMsg, "error_installation:") {
		// Show error and reset installation state
		errorText := strings.TrimPrefix(errMsg, "error_installation: ")
		m.installationInProgress = false
		m.isLoading = false
		m.loadingMessage = fmt.Sprintf("Installation Error: %s", errorText)
		m.choices = m.getMenuChoices()
		return m, nil
	}

	// Handle authentication error messages
	if errMsg, ok := msg.(string); ok && strings.HasPrefix(errMsg, "error_auth:") {
		// Show error and go back to main menu
		errorText := strings.TrimPrefix(errMsg, "error_auth: ")
		m.currentMenu = MainMenu
		m.menuStack = []MenuType{} // Clear the stack
		m.loadingMessage = fmt.Sprintf("Auth Error: %s", errorText)
		m.choices = m.getMenuChoices()
		m.cursor = 0
		return m, nil
	}

	// Handle authentication model updates when in auth mode
	if m.currentMenu == GitHubAuthMenu && m.authModel != nil {
		updatedAuthModel, cmd := m.authModel.Update(msg)
		m.authModel = updatedAuthModel
		
		// Return the command from auth model (which might generate auth_complete/auth_cancelled messages)
		if cmd != nil {
			return m, cmd
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle results screen - any key returns to menu
		if m.showingResults {
			m.showingResults = false
			m.installationResults = []InstallationResult{}
			m.choices = m.getMenuChoices()
			return m, nil
		}
		
		switch msg.String() {
		case "ctrl+c":
			// If installation is in progress, ask for confirmation
			if m.installationInProgress {
				// TODO: Add confirmation dialog for interrupting installation
				// For now, allow immediate exit but could be enhanced
				if m.installEngine != nil {
					m.installEngine.Cleanup()
				}
			}
			return m, tea.Quit
		case "q":
			// Only quit from main menu, otherwise go back
			if m.currentMenu == MainMenu {
				return m, tea.Quit
			} else {
				m.navigateBack()
			}
		case "esc", "b":
			// Navigate back to previous menu
			if m.currentMenu == MainMenu {
				return m, tea.Quit
			} else {
				m.navigateBack()
			}
		case "up", "k":
			currentChoices := m.getMenuChoices()
			if m.cursor > 0 {
				m.cursor--
			} else {
				// Wrap to bottom
				m.cursor = len(currentChoices) - 1
			}
		case "down", "j":
			currentChoices := m.getMenuChoices()
			if m.cursor < len(currentChoices)-1 {
				m.cursor++
			} else {
				// Wrap to top
				m.cursor = 0
			}
		case "enter", " ":
			return m.handleMenuSelection()
		}
	}
	return m, nil
}