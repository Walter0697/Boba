package github

import (
	"testing"
	
	tea "github.com/charmbracelet/bubbletea"
)

// TestAuthModelCreation tests that the authentication model can be created properly
func TestAuthModelCreation(t *testing.T) {
	onComplete := func(client *GitHubClient, repoURL string) tea.Cmd {
		return nil
	}
	
	onCancel := func() tea.Cmd {
		return nil
	}
	
	authModel := NewAuthModel(onComplete, onCancel)
	
	if authModel == nil {
		t.Fatal("NewAuthModel() returned nil")
	}
	
	if authModel.state != AuthStateTokenInput {
		t.Errorf("NewAuthModel() initial state = %v, want %v", authModel.state, AuthStateTokenInput)
	}
	
	if authModel.onComplete == nil {
		t.Error("NewAuthModel() onComplete callback is nil")
	}
	
	if authModel.onCancel == nil {
		t.Error("NewAuthModel() onCancel callback is nil")
	}
}

// TestAuthModelGetters tests the getter methods
func TestAuthModelGetters(t *testing.T) {
	authModel := NewAuthModel(nil, nil)
	
	// Test initial empty values
	if authModel.GetToken() != "" {
		t.Errorf("GetToken() = %v, want empty string", authModel.GetToken())
	}
	
	if authModel.GetRepoURL() != "boba-config" {
		t.Errorf("GetRepoURL() = %v, want boba-config", authModel.GetRepoURL())
	}
	
	if authModel.GetClient() != nil {
		t.Errorf("GetClient() = %v, want nil", authModel.GetClient())
	}
}

// TestAuthModelView tests that the view method returns content
func TestAuthModelView(t *testing.T) {
	authModel := NewAuthModel(nil, nil)
	
	view := authModel.View()
	if view == "" {
		t.Error("View() returned empty string")
	}
	
	// Should contain authentication header
	if !contains(view, "GitHub Authentication") {
		t.Error("View() should contain 'GitHub Authentication' header")
	}
	
	// Should contain token input prompt
	if !contains(view, "GitHub Personal Access Token") {
		t.Error("View() should contain token input prompt")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || 
		s[len(s)-len(substr):] == substr || 
		containsInMiddle(s, substr))))
}

func containsInMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}