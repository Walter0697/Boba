package ui

import (
	"testing"
	
	"boba/internal/installer"
)

func TestSystemInstallMenuNavigation(t *testing.T) {
	// Create a test model
	model := InitialModel()
	
	// Verify system installer is initialized
	if model.systemInstaller == nil {
		t.Error("System installer should be initialized")
	}
	
	// Test navigation to system install menu
	model.cursor = 4 // Position of "Install BOBA to System" in main menu
	updatedModel, _ := model.HandleMenuSelection()
	
	if updatedModel.GetCurrentMenu() != SystemInstallMenu {
		t.Errorf("Expected SystemInstallMenu, got %v", updatedModel.GetCurrentMenu())
	}
}

func TestSystemInstallMenuChoices(t *testing.T) {
	// Create a test model
	model := InitialModel()
	model.currentMenu = SystemInstallMenu
	
	// Get menu choices
	choices := model.getSystemInstallChoices()
	
	if len(choices) == 0 {
		t.Error("System install menu should have choices")
	}
	
	// Should have a back option
	lastChoice := choices[len(choices)-1]
	if lastChoice != "← Back to Main Menu" {
		t.Errorf("Expected back option, got: %s", lastChoice)
	}
}

func TestSystemInstallMenuWithoutInstaller(t *testing.T) {
	// Create a test model without system installer
	model := InitialModel()
	model.systemInstaller = nil
	model.currentMenu = SystemInstallMenu
	
	// Get menu choices
	choices := model.getSystemInstallChoices()
	
	// Should show error message
	if len(choices) < 2 {
		t.Error("Should have error message and back option")
	}
	
	if choices[0] != "❌ System installer not available" {
		t.Errorf("Expected error message, got: %s", choices[0])
	}
}

func TestSystemInstallationResult(t *testing.T) {
	// Create a test model
	model := InitialModel()
	model.currentMenu = SystemInstallMenu
	
	// Simulate successful installation result
	model.systemInstallResult = &installer.SystemInstallationResult{
		Success:         true,
		BinaryInstalled: true,
		ZshrcModified:   true,
		Message:         "Installation successful",
	}
	
	// Get menu choices
	choices := model.getSystemInstallChoices()
	
	// Should show success message
	found := false
	for _, choice := range choices {
		if choice == "✅ System Installation Successful!" {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("Should show success message")
	}
}

func TestSystemInstallationFailureResult(t *testing.T) {
	// Create a test model
	model := InitialModel()
	model.currentMenu = SystemInstallMenu
	
	// Simulate failed installation result
	model.systemInstallResult = &installer.SystemInstallationResult{
		Success: false,
		Error:   &testError{msg: "Installation failed"},
		Message: "Installation failed",
	}
	
	// Get menu choices
	choices := model.getSystemInstallChoices()
	
	// Should show failure message
	found := false
	for _, choice := range choices {
		if choice == "❌ System Installation Failed" {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("Should show failure message")
	}
}

// Test error type for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}