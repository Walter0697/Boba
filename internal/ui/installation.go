package ui

import (
	"fmt"
	
	tea "github.com/charmbracelet/bubbletea"
	"boba/internal/parser"
)

// runInstallEverythingWithProgress runs the installation process with real-time progress updates
func (m MenuModel) runInstallEverythingWithProgress() tea.Cmd {
	return func() tea.Msg {
		// Fetch tools from repository with retry logic
		tools, err := m.repoParser.GetTools()
		if err != nil {
			return fmt.Sprintf("error_installation: Failed to fetch tools: %v", err)
		}
		
		// Fetch environments from repository
		environments, err := m.repoParser.FetchEnvironments()
		if err != nil {
			return fmt.Sprintf("error_installation: Failed to fetch environments: %v", err)
		}
		
		// Filter tools and environments based on configuration
		config := m.configManager.GetConfig()
		
		var toolsToInstall []parser.Tool
		for _, tool := range tools {
			shouldInstall := tool.AutoInstall
			
			// Check for override
			if override, exists := config.ToolOverrides[tool.Name]; exists {
				shouldInstall = override
			}
			
			if shouldInstall {
				toolsToInstall = append(toolsToInstall, tool)
			}
		}
		
		var environmentsToApply []parser.Environment
		for _, env := range environments {
			shouldApply := env.AutoApply
			
			// Check for override
			if override, exists := config.EnvironmentOverrides[env.Name]; exists {
				shouldApply = override
			}
			
			if shouldApply {
				environmentsToApply = append(environmentsToApply, env)
			}
		}
		
		// Resolve dependencies and get installation order
		resolver := m.dependencyResolver
		orderedTools, orderedEnvironments, err := resolver.GetInstallationOrder(toolsToInstall, environmentsToApply)
		if err != nil {
			return fmt.Sprintf("error_installation: Failed to resolve dependencies: %v", err)
		}
		
		// Start with tools phase
		return InstallEverythingPhaseMsg{
			Phase:        "tools",
			Tools:        orderedTools,
			Environments: orderedEnvironments,
		}
	}
}

// runUpdateEverythingWithProgress runs the update process for installed tools
func (m MenuModel) runUpdateEverythingWithProgress() tea.Cmd {
	return func() tea.Msg {
		// Get list of installed tools
		tools, err := m.repoParser.GetTools()
		if err != nil {
			return fmt.Sprintf("error_installation: Failed to fetch tools: %v", err)
		}
		
		// Filter to only installed tools
		var installedTools []parser.Tool
		for _, tool := range tools {
			if m.installEngine.IsToolInstalled(tool) {
				installedTools = append(installedTools, tool)
			}
		}
		
		if len(installedTools) == 0 {
			return InstallationCompleteMsg{
				Results: []InstallationResult{{
					ToolName: "Update",
					Success:  true,
					Message:  "No installed tools found to update",
				}},
			}
		}
		
		// Start updating tools
		return InstallationStartMsg{
			Tools:        installedTools,
			CurrentIndex: 0,
			Results:      []InstallationResult{},
		}
	}
}

// installNextTool installs the next tool in the sequence with progress feedback
func (m MenuModel) installNextTool(tools []parser.Tool, currentIndex int, results []InstallationResult) tea.Cmd {
	if currentIndex >= len(tools) {
		// All tools processed, check if we need to move to environments phase
		if m.installEverythingMode && len(m.pendingEnvironments) > 0 {
			// Move to environments phase
			return func() tea.Msg {
				return InstallEverythingPhaseMsg{
					Phase:        "environments",
					Tools:        tools,
					Environments: m.pendingEnvironments,
				}
			}
		}
		
		// Complete the installation
		return func() tea.Msg {
			return InstallationCompleteMsg{Results: results}
		}
	}
	
	return func() tea.Msg {
		currentTool := tools[currentIndex]
		
		// Install the tool
		result, err := m.installEngine.InstallTool(currentTool)
		
		success := result.Success && err == nil
		message := result.Output
		if err != nil {
			message = fmt.Sprintf("Installation failed: %v", err)
		}
		
		// Record successful installation
		if success {
			version := currentTool.Version
			if version == "" {
				version = "latest"
			}
			m.configManager.RecordToolInstallation(currentTool.Name, version, "auto")
		}
		
		// Add result to the list
		newResults := append(results, InstallationResult{
			ToolName: currentTool.Name,
			Success:  success,
			Message:  message,
			Error:    err,
		})
		
		// Continue with next tool
		return InstallationNextMsg{
			Tools:        tools,
			CurrentIndex: currentIndex + 1,
			Results:      newResults,
		}
	}
}

// applyNextEnvironment applies the next environment in the sequence with progress feedback
func (m MenuModel) applyNextEnvironment(environments []parser.Environment, currentIndex int, results []EnvironmentApplicationResult) tea.Cmd {
	if currentIndex >= len(environments) {
		// All environments processed, convert results to InstallationResult format and complete
		var installResults []InstallationResult
		for _, result := range results {
			installResults = append(installResults, InstallationResult{
				ToolName: result.EnvironmentName,
				Success:  result.Success,
				Message:  result.Message,
				Error:    result.Error,
			})
		}
		
		return func() tea.Msg {
			return InstallationCompleteMsg{Results: installResults}
		}
	}
	
	return func() tea.Msg {
		currentEnv := environments[currentIndex]
		
		// Apply the environment configuration using the installation engine
		installResult, err := m.installEngine.ApplyEnvironment(currentEnv)
		
		success := installResult.Success && err == nil
		message := installResult.Output
		if err != nil {
			message = fmt.Sprintf("Environment application failed: %v", err)
		}
		
		result := EnvironmentApplicationResult{
			EnvironmentName: currentEnv.Name,
			Success:         success,
			Message:         message,
			Error:           err,
		}
		
		// Add result to the list
		newResults := append(results, result)
		
		// Continue with next environment
		return EnvironmentApplicationNextMsg{
			Environments: environments,
			CurrentIndex: currentIndex + 1,
			Results:      newResults,
		}
	}
}