package ui

import (
	"testing"
	tea "github.com/charmbracelet/bubbletea"
	"boba/internal/parser"
)

func TestInstallationProgressNavigationReset(t *testing.T) {
	// Create a properly initialized model in loading state
	model := MenuModel{
		isLoading:              true,
		installationInProgress: true,
		loadingMessage:         "Installing something...",
		toolInstallStatus:      make(map[string]bool),
		availableTools:         []parser.Tool{},
		choices:                []string{},
	}
	
	// Simulate an InstallationProgressMsg
	progressMsg := InstallationProgressMsg{
		ToolName: "test-tool",
		Status:   "Installation completed successfully",
		Success:  true,
	}
	
	// Process the message
	updatedModel, _ := model.Update(progressMsg)
	
	// Cast back to MenuModel
	if menuModel, ok := updatedModel.(MenuModel); ok {
		// Verify that loading states are reset
		if menuModel.isLoading {
			t.Error("Expected isLoading to be false after InstallationProgressMsg")
		}
		
		if menuModel.installationInProgress {
			t.Error("Expected installationInProgress to be false after InstallationProgressMsg")
		}
		
		// Verify that we're now showing results instead of immediately returning to menu
		if !menuModel.showingResults {
			t.Error("Expected showingResults to be true after InstallationProgressMsg")
		}
		
		// Verify that results are populated
		if len(menuModel.installationResults) != 1 {
			t.Errorf("Expected 1 installation result, got %d", len(menuModel.installationResults))
		}
		
		if menuModel.installationResults[0].ToolName != "test-tool" {
			t.Errorf("Expected result for 'test-tool', got '%s'", menuModel.installationResults[0].ToolName)
		}
	} else {
		t.Fatal("Expected MenuModel type after update")
	}
}

func TestInstallationCompleteNavigationReset(t *testing.T) {
	// Create a properly initialized model in installation progress state
	model := MenuModel{
		isLoading:              true,
		installationInProgress: true,
		installEverythingMode:  false,
		pendingEnvironments:    nil,
		toolInstallStatus:      make(map[string]bool),
		availableTools:         []parser.Tool{},
		choices:                []string{},
	}
	
	// Simulate an InstallationCompleteMsg
	completeMsg := InstallationCompleteMsg{
		Results: []InstallationResult{
			{
				ToolName: "test-tool",
				Success:  true,
				Message:  "Installation successful",
			},
		},
	}
	
	// Process the message
	updatedModel, _ := model.Update(completeMsg)
	
	// Cast back to MenuModel
	if menuModel, ok := updatedModel.(MenuModel); ok {
		// Verify that all states are reset
		if menuModel.isLoading {
			t.Error("Expected isLoading to be false after InstallationCompleteMsg")
		}
		
		if menuModel.installationInProgress {
			t.Error("Expected installationInProgress to be false after InstallationCompleteMsg")
		}
		
		if menuModel.installEverythingMode {
			t.Error("Expected installEverythingMode to be false after InstallationCompleteMsg")
		}
		
		if menuModel.pendingEnvironments != nil {
			t.Error("Expected pendingEnvironments to be nil after InstallationCompleteMsg")
		}
		
		// Verify that we're showing results instead of immediately returning to menu
		if !menuModel.showingResults {
			t.Error("Expected showingResults to be true after InstallationCompleteMsg")
		}
		
		// Verify that results are stored
		if len(menuModel.installationResults) != 1 {
			t.Errorf("Expected 1 installation result, got %d", len(menuModel.installationResults))
		}
	} else {
		t.Fatal("Expected MenuModel type after update")
	}
}

func TestEnvironmentApplicationNavigationFlow(t *testing.T) {
	// Create a properly initialized model
	model := MenuModel{
		isLoading:              true,
		installationInProgress: false,
		toolInstallStatus:      make(map[string]bool),
		availableTools:         []parser.Tool{},
		choices:                []string{},
	}
	
	// Simulate environment application progress
	progressMsg := InstallationProgressMsg{
		ToolName: "dev-env",
		Status:   "Environment application completed successfully",
		Success:  true,
	}
	
	// Process the message
	updatedModel, _ := model.Update(progressMsg)
	
	// Cast back to MenuModel
	if menuModel, ok := updatedModel.(MenuModel); ok {
		// Verify that loading is complete
		if menuModel.isLoading {
			t.Error("Expected isLoading to be false after environment application")
		}
		
		// Verify that we're showing results screen
		if !menuModel.showingResults {
			t.Error("Expected showingResults to be true after environment application")
		}
		
		// Verify the result is properly set
		if len(menuModel.installationResults) != 1 {
			t.Errorf("Expected 1 result, got %d", len(menuModel.installationResults))
		}
		
		if menuModel.installationResults[0].ToolName != "dev-env" {
			t.Errorf("Expected result for 'dev-env', got '%s'", menuModel.installationResults[0].ToolName)
		}
	} else {
		t.Fatal("Expected MenuModel type after update")
	}
}

func TestResultsScreenKeyPressNavigation(t *testing.T) {
	// Create a model showing results
	model := MenuModel{
		showingResults: true,
		installationResults: []InstallationResult{
			{
				ToolName: "test-tool",
				Success:  true,
				Message:  "Installation successful",
			},
		},
		toolInstallStatus: make(map[string]bool),
		availableTools:    []parser.Tool{},
		choices:           []string{},
	}
	
	// Simulate any key press (space in this case)
	keyMsg := tea.KeyMsg{Type: tea.KeySpace}
	
	// Process the key press
	updatedModel, _ := model.Update(keyMsg)
	
	// Cast back to MenuModel
	if menuModel, ok := updatedModel.(MenuModel); ok {
		// Verify that we're no longer showing results
		if menuModel.showingResults {
			t.Error("Expected showingResults to be false after key press")
		}
		
		// Verify that results are cleared
		if len(menuModel.installationResults) != 0 {
			t.Errorf("Expected 0 installation results after returning to menu, got %d", len(menuModel.installationResults))
		}
	} else {
		t.Fatal("Expected MenuModel type after update")
	}
}