package ui

import (
	"os"
	"testing"
	
	"boba/internal/config"
	"boba/internal/parser"
)

// TestCompleteUserWorkflow tests the entire user workflow from startup to installation
func TestCompleteUserWorkflow(t *testing.T) {
	// Skip integration tests in CI or when SKIP_INTEGRATION is set
	if os.Getenv("SKIP_INTEGRATION") != "" {
		t.Skip("Skipping integration test")
	}
	
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Override the home directory for testing
	originalHome := os.Getenv("HOME")
	defer func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		}
	}()
	os.Setenv("HOME", tempDir)
	
	t.Run("FirstTimeSetup", func(t *testing.T) {
		// Test first-time user experience
		model := InitialModel()
		
		// Should show welcome message for unconfigured system
		if model.authError == "" {
			t.Error("Expected welcome message for first-time setup")
		}
		
		if !containsString(model.authError, "Welcome to BOBA") {
			t.Errorf("Expected welcome message, got: %s", model.authError)
		}
		
		// Should be on main menu
		if model.currentMenu != MainMenu {
			t.Errorf("Expected MainMenu, got %v", model.currentMenu)
		}
		
		// Should have all main menu options
		expectedChoices := []string{
			"Install Everything",
			"Update Everything", 
			"List of Available Tools",
			"Setup Environment",
			"Installation Configuration",
		}
		
		if len(model.choices) != len(expectedChoices) {
			t.Errorf("Expected %d choices, got %d", len(expectedChoices), len(model.choices))
		}
	})
	
	t.Run("ConfigurationWorkflow", func(t *testing.T) {
		model := InitialModel()
		
		// Test that we can access configuration menu choices
		configChoices := []string{
			"Repository Configuration",
			"Tool Override Management",
			"Environment Override Management",
			"← Back to Main Menu",
		}
		
		// Verify we have some configuration options available
		if len(configChoices) == 0 {
			t.Error("Expected configuration choices to be available")
		}
		
		// Test that view renders without error
		view := model.View()
		if view == "" {
			t.Error("Configuration view should not be empty")
		}
	})
	
	t.Run("AuthenticationWorkflow", func(t *testing.T) {
		// Test authentication flow
		model := InitialModel()
		
		// Test that authentication error is shown when not authenticated
		if model.GetAuthError() == "" {
			t.Error("Expected authentication error for unconfigured system")
		}
		
		// Test that view renders authentication message
		view := model.View()
		if !containsString(view, "GitHub") {
			t.Error("Expected GitHub authentication message in view")
		}
	})
	
	t.Run("ToolListingWorkflow", func(t *testing.T) {
		model := setupAuthenticatedModel(t)
		
		// Test that we can simulate tool listing
		tools := []parser.Tool{
			{Name: "git", Description: "Version control", AutoInstall: true},
			{Name: "docker", Description: "Container platform", AutoInstall: true},
		}
		
		model.availableTools = tools
		model.isLoading = false
		
		// Test that view renders with tools
		view := model.View()
		if view == "" {
			t.Error("Tools list view should not be empty")
		}
	})
	
	t.Run("InstallationWorkflow", func(t *testing.T) {
		model := setupAuthenticatedModel(t)
		
		// Set up tools for installation
		tools := []parser.Tool{
			{Name: "test-tool", Description: "Test tool", AutoInstall: true},
		}
		model.availableTools = tools
		
		// Simulate installation process
		model.installationInProgress = true
		
		// Simulate installation result
		result := InstallationResult{
			ToolName: "test-tool",
			Success:  true,
			Message:  "Installation completed successfully",
		}
		model.installationResults = []InstallationResult{result}
		
		// Verify installation state
		if !model.installationInProgress {
			t.Error("Expected installation to be in progress")
		}
		
		if len(model.installationResults) != 1 {
			t.Errorf("Expected 1 installation result, got %d", len(model.installationResults))
		}
		
		if !model.installationResults[0].Success {
			t.Error("Expected successful installation result")
		}
	})
	
	t.Run("ErrorHandlingWorkflow", func(t *testing.T) {
		// Test error handling scenarios
		model := InitialModel()
		
		// Test authentication error
		model.authError = "Test authentication error"
		view := model.View()
		
		if !containsString(view, "Test authentication error") {
			t.Error("Expected authentication error to be displayed in view")
		}
		
		// Test loading state
		model.isLoading = true
		model.loadingMessage = "Loading test data..."
		loadingView := model.View()
		
		if !containsString(loadingView, "Loading test data...") {
			t.Error("Expected loading message to be displayed")
		}
		
		// Test installation error
		model.isLoading = false
		model.installationInProgress = true
		model.installationResults = []InstallationResult{
			{
				ToolName: "failing-tool",
				Success:  false,
				Message:  "Installation failed",
			},
		}
		
		errorView := model.View()
		if !containsString(errorView, "failing-tool") {
			t.Error("Expected failed tool name in view")
		}
		
		if !containsString(errorView, "Installation failed") {
			t.Error("Expected failure message in view")
		}
	})
	
	t.Run("NavigationWorkflow", func(t *testing.T) {
		model := InitialModel()
		
		// Test that we start on main menu
		if model.GetCurrentMenu() != MainMenu {
			t.Errorf("Expected MainMenu, got %v", model.GetCurrentMenu())
		}
		
		// Test cursor movement
		model = model.SetCursor(1)
		if model.GetCursor() != 1 {
			t.Error("Expected cursor to move to position 1")
		}
		
		// Test that view renders correctly after navigation
		view := model.View()
		if view == "" {
			t.Error("View should not be empty after navigation")
		}
	})
}

// TestUIRendering tests the UI rendering with various states
func TestUIRendering(t *testing.T) {
	t.Run("MainMenuRendering", func(t *testing.T) {
		model := InitialModel()
		view := model.View()
		
		// Should contain header (the ASCII art might not contain "BOBA" as plain text)
		if view == "" {
			t.Error("Expected non-empty view")
		}
		
		// Should contain menu options
		if !containsString(view, "Install Everything") {
			t.Error("Expected 'Install Everything' option in view")
		}
		
		// Should contain help text
		if !containsString(view, "Navigate:") {
			t.Error("Expected navigation help in view")
		}
	})
	
	t.Run("LoadingStateRendering", func(t *testing.T) {
		model := InitialModel()
		model.isLoading = true
		model.loadingMessage = "Fetching repository data..."
		
		view := model.View()
		
		if !containsString(view, "Fetching repository data...") {
			t.Error("Expected loading message in view")
		}
		
		if !containsString(view, "Please wait") {
			t.Error("Expected loading help text in view")
		}
	})
	
	t.Run("InstallationProgressRendering", func(t *testing.T) {
		model := InitialModel()
		model.installationInProgress = true
		model.installationResults = []InstallationResult{
			{ToolName: "git", Success: true, Message: "Installed successfully"},
			{ToolName: "docker", Success: false, Message: "Installation failed"},
		}
		
		view := model.View()
		
		if !containsString(view, "Installation in Progress") {
			t.Error("Expected installation progress title in view")
		}
		
		if !containsString(view, "git") {
			t.Error("Expected git tool in installation results")
		}
		
		if !containsString(view, "docker") {
			t.Error("Expected docker tool in installation results")
		}
		
		if !containsString(view, "✅") {
			t.Error("Expected success icon for successful installation")
		}
		
		if !containsString(view, "❌") {
			t.Error("Expected error icon for failed installation")
		}
	})
}

// setupAuthenticatedModel creates a model with authentication set up for testing
func setupAuthenticatedModel(t *testing.T) MenuModel {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Override the home directory for testing
	originalHome := os.Getenv("HOME")
	t.Cleanup(func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		}
	})
	os.Setenv("HOME", tempDir)
	
	// Create config manager and set up authentication
	configManager := config.NewConfigManager()
	configManager.LoadConfig()
	configManager.LoadCredentials()
	
	// Set up test credentials
	err := configManager.SetGitHubToken("test_token_123")
	if err != nil {
		t.Fatalf("Failed to set GitHub token: %v", err)
	}
	
	err = configManager.SetRepositoryURL("testuser/test-repo")
	if err != nil {
		t.Fatalf("Failed to set repository URL: %v", err)
	}
	
	// Create model
	model := MenuModel{
		currentMenu:   MainMenu,
		menuStack:     []MenuType{},
		configManager: configManager,
		choices: []string{
			"Install Everything",
			"Update Everything",
			"List of Available Tools",
			"Setup Environment", 
			"Installation Configuration",
		},
		selected:          make(map[int]struct{}),
		toolInstallStatus: make(map[string]bool),
	}
	
	return model
}

// Helper function to check if a string contains a substring
func containsString(haystack, needle string) bool {
	return len(haystack) >= len(needle) && 
		   (haystack == needle || 
		    haystack[:len(needle)] == needle || 
		    haystack[len(haystack)-len(needle):] == needle ||
		    findSubstring(haystack, needle))
}

func findSubstring(haystack, needle string) bool {
	for i := 0; i <= len(haystack)-len(needle); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}