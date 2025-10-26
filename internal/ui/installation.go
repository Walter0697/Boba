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
		fmt.Printf("\n=== All Tools Processed ===\n")
		fmt.Printf("Total tools: %d\n", len(results))
		// All tools processed, check if we need to move to environments phase
		if m.installEverythingMode && len(m.pendingEnvironments) > 0 {
			fmt.Printf("Moving to environments phase...\n")
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
		
		fmt.Printf("\n=== Processing Tool %d/%d ===\n", currentIndex+1, len(tools))
		fmt.Printf("Tool: %s\n", currentTool.Name)
		fmt.Printf("This is a BLOCKING operation - next tool will wait\n")
		
		// Install the tool
		// This call BLOCKS until the tool installation is complete
		result, err := m.installEngine.InstallTool(currentTool)
		
		fmt.Printf("\n=== Tool %s Completed ===\n", currentTool.Name)
		fmt.Printf("Success: %v\n", result.Success)
		fmt.Printf("Duration: %v\n", result.Duration)
		
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
		
		fmt.Printf("Moving to next tool (if any)...\n")
		
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
		fmt.Printf("\n=== All Environments Processed ===\n")
		fmt.Printf("Total environments: %d\n", len(results))
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
		
		fmt.Printf("\n=== Processing Environment %d/%d ===\n", currentIndex+1, len(environments))
		fmt.Printf("Environment: %s\n", currentEnv.Name)
		fmt.Printf("This is a BLOCKING operation - next environment will wait\n")
		
		// Apply the environment configuration using the installation engine
		// This call BLOCKS until the environment setup is complete
		installResult, err := m.installEngine.ApplyEnvironment(currentEnv)
		
		fmt.Printf("\n=== Environment %s Completed ===\n", currentEnv.Name)
		fmt.Printf("Success: %v\n", installResult.Success)
		fmt.Printf("Duration: %v\n", installResult.Duration)
		
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
		
		fmt.Printf("Moving to next environment (if any)...\n")
		
		// Continue with next environment
		return EnvironmentApplicationNextMsg{
			Environments: environments,
			CurrentIndex: currentIndex + 1,
			Results:      newResults,
		}
	}
}