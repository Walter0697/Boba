package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
	
	"boba/internal/ui"
)

// TestApplicationIntegration tests the complete application from main entry point
func TestApplicationIntegration(t *testing.T) {
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
	
	t.Run("ApplicationStartup", func(t *testing.T) {
		// Test that the application can start without crashing
		uiManager := ui.NewUIManager()
		
		if uiManager == nil {
			t.Fatal("UIManager creation failed")
		}
		
		// We can't easily test the full UI interaction without a terminal,
		// but we can test that the initial model is created correctly
		model := ui.InitialModel()
		
		// Verify initial state
		if model.GetCurrentMenu() != ui.MainMenu {
			t.Errorf("Expected MainMenu, got %v", model.GetCurrentMenu())
		}
		
		// Verify configuration manager is initialized
		if model.GetConfigManager() == nil {
			t.Error("Expected ConfigManager to be initialized")
		}
		
		// Test initial view rendering
		view := model.View()
		if view == "" {
			t.Error("Expected non-empty view")
		}
		
		// Should contain application header
		if !containsString(view, "BOBA") {
			t.Error("Expected BOBA header in initial view")
		}
	})
	
	t.Run("ConfigurationPersistence", func(t *testing.T) {
		// Test that configuration persists across application restarts
		
		// First application instance
		model1 := ui.InitialModel()
		configManager1 := model1.GetConfigManager()
		
		// Configure the application
		err := configManager1.SetRepositoryURL("testuser/test-config")
		if err != nil {
			t.Fatalf("Failed to set repository URL: %v", err)
		}
		
		err = configManager1.SetGitHubToken("test_token_123")
		if err != nil {
			t.Fatalf("Failed to set GitHub token: %v", err)
		}
		
		err = configManager1.SetToolOverride("docker", false)
		if err != nil {
			t.Fatalf("Failed to set tool override: %v", err)
		}
		
		// Verify configuration files were created
		configDir := configManager1.GetConfigDir()
		expectedFiles := []string{
			filepath.Join(configDir, "config.json"),
			filepath.Join(configDir, "credentials.json"),
		}
		
		for _, file := range expectedFiles {
			if _, err := os.Stat(file); os.IsNotExist(err) {
				t.Errorf("Expected configuration file not created: %s", file)
			}
		}
		
		// Second application instance (simulating restart)
		model2 := ui.InitialModel()
		configManager2 := model2.GetConfigManager()
		
		// Verify configuration persisted
		config := configManager2.GetConfig()
		if config.RepositoryURL != "testuser/test-config" {
			t.Errorf("Repository URL not persisted: expected testuser/test-config, got %s", config.RepositoryURL)
		}
		
		credentials := configManager2.GetCredentials()
		if credentials.GitHubToken != "test_token_123" {
			t.Errorf("GitHub token not persisted: expected test_token_123, got %s", credentials.GitHubToken)
		}
		
		// Verify tool override persisted
		if enabled, exists := configManager2.GetToolOverride("docker"); !exists || enabled {
			t.Error("Tool override not persisted correctly")
		}
	})
	
	t.Run("ErrorRecovery", func(t *testing.T) {
		// Test application behavior with various error conditions
		
		// Test with corrupted config file
		model := ui.InitialModel()
		configManager := model.GetConfigManager()
		configDir := configManager.GetConfigDir()
		
		// Write invalid JSON to config file
		configFile := filepath.Join(configDir, "config.json")
		err := os.WriteFile(configFile, []byte("invalid json"), 0644)
		if err != nil {
			t.Fatalf("Failed to write invalid config: %v", err)
		}
		
		// Application should handle corrupted config gracefully
		model2 := ui.InitialModel()
		view := model2.View()
		
		// Should still render without crashing
		if view == "" {
			t.Error("Expected application to handle corrupted config gracefully")
		}
		
		// Should show some indication of configuration issues
		if model2.GetAuthError() == "" {
			t.Error("Expected some indication of configuration issues")
		}
	})
	
	t.Run("MenuNavigation", func(t *testing.T) {
		// Test complete menu navigation workflow
		model := ui.InitialModel()
		
		// Test cursor movement
		for i := 0; i < 3; i++ {
			model = model.SetCursor(i)
			
			// Verify cursor position
			if model.GetCursor() != i {
				t.Errorf("Failed to set cursor to position %d", i)
			}
			
			// Test that view renders without error
			view := model.View()
			if view == "" {
				t.Error("View should not be empty")
			}
		}
		
		// Test that we can navigate to configuration menu
		model = model.SetCursor(4) // Installation Configuration (assuming it's at index 4)
		
		// Test that view still renders
		view := model.View()
		if view == "" {
			t.Error("View should not be empty after cursor change")
		}
	})
	
	t.Run("ViewRendering", func(t *testing.T) {
		// Test that all view states render without errors
		model := ui.InitialModel()
		
		// Test main menu view
		mainView := model.View()
		if mainView == "" {
			t.Error("Main menu view should not be empty")
		}
		
		// Test loading state view
		model = model.SetLoading(true, "Loading test data...")
		loadingView := model.View()
		if !containsString(loadingView, "Loading test data...") {
			t.Error("Loading view should contain loading message")
		}
		
		// Test installation progress view
		model = model.SetLoading(false, "")
		model = model.SetInstallationInProgress(true)
		installView := model.View()
		if !containsString(installView, "Installation") {
			t.Error("Installation view should contain installation text")
		}
		
		// Test error state view
		model = model.SetInstallationInProgress(false)
		model = model.SetAuthError("Test error message")
		errorView := model.View()
		if !containsString(errorView, "Test error message") {
			t.Error("Error view should contain error message")
		}
	})
}

// TestApplicationPerformance tests application performance characteristics
func TestApplicationPerformance(t *testing.T) {
	// Skip performance tests in CI or when SKIP_INTEGRATION is set
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
	
	t.Run("StartupPerformance", func(t *testing.T) {
		// Test that application starts up quickly
		start := time.Now()
		
		model := ui.InitialModel()
		view := model.View()
		
		duration := time.Since(start)
		
		// Application should start up in less than 1 second
		if duration > time.Second {
			t.Errorf("Application startup took too long: %v", duration)
		}
		
		// View should be rendered
		if view == "" {
			t.Error("View should be rendered after startup")
		}
	})
	
	t.Run("ViewRenderingPerformance", func(t *testing.T) {
		model := ui.InitialModel()
		
		// Test that view rendering is fast
		start := time.Now()
		
		for i := 0; i < 100; i++ {
			view := model.View()
			if view == "" {
				t.Error("View should not be empty")
			}
		}
		
		duration := time.Since(start)
		
		// 100 view renders should complete in less than 100ms
		if duration > 100*time.Millisecond {
			t.Errorf("View rendering too slow: %v for 100 renders", duration)
		}
	})
	
	t.Run("ConfigurationPerformance", func(t *testing.T) {
		model := ui.InitialModel()
		configManager := model.GetConfigManager()
		
		// Test configuration operations performance
		start := time.Now()
		
		for i := 0; i < 50; i++ {
			err := configManager.SetToolOverride("test-tool", i%2 == 0)
			if err != nil {
				t.Fatalf("Failed to set tool override: %v", err)
			}
		}
		
		duration := time.Since(start)
		
		// 50 configuration operations should complete quickly
		if duration > 500*time.Millisecond {
			t.Errorf("Configuration operations too slow: %v for 50 operations", duration)
		}
	})
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