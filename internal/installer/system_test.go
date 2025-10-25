package installer

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestNewSystemInstaller(t *testing.T) {
	installer, err := NewSystemInstaller()
	if err != nil {
		t.Fatalf("Failed to create system installer: %v", err)
	}
	
	if installer == nil {
		t.Fatal("System installer is nil")
	}
	
	// Check that paths are set
	if installer.binaryPath == "" {
		t.Error("Binary path is empty")
	}
	
	if installer.installPath == "" {
		t.Error("Install path is empty")
	}
	
	if installer.zshrcPath == "" {
		t.Error("Zshrc path is empty")
	}
}

func TestRequiresSudo(t *testing.T) {
	installer, err := NewSystemInstaller()
	if err != nil {
		t.Fatalf("Failed to create system installer: %v", err)
	}
	
	// On Windows, should not require sudo
	if runtime.GOOS == "windows" {
		if installer.RequiresSudo() {
			t.Error("Windows should not require sudo")
		}
	}
	
	// Test is platform-dependent, so we just ensure it doesn't panic
	_ = installer.RequiresSudo()
}

func TestGetInstallationInfo(t *testing.T) {
	installer, err := NewSystemInstaller()
	if err != nil {
		t.Fatalf("Failed to create system installer: %v", err)
	}
	
	info := installer.GetInstallationInfo()
	
	// Check required fields
	requiredFields := []string{"binary_path", "install_path", "zshrc_path", "is_installed", "requires_sudo", "platform", "has_backup"}
	
	for _, field := range requiredFields {
		if _, exists := info[field]; !exists {
			t.Errorf("Missing required field: %s", field)
		}
	}
	
	// Check platform matches runtime
	if info["platform"] != runtime.GOOS {
		t.Errorf("Platform mismatch: got %v, expected %s", info["platform"], runtime.GOOS)
	}
}

func TestGenerateBobaConfiguration(t *testing.T) {
	installer, err := NewSystemInstaller()
	if err != nil {
		t.Fatalf("Failed to create system installer: %v", err)
	}
	
	config := installer.generateBobaConfiguration()
	
	// Check that configuration contains expected elements
	expectedElements := []string{
		"# BOBA CLI Tool Configuration",
		"export PATH=",
		"alias boba-update=",
		"alias dev-setup=",
	}
	
	for _, element := range expectedElements {
		if !contains(config, element) {
			t.Errorf("Configuration missing expected element: %s", element)
		}
	}
}

func TestCopyBinary(t *testing.T) {
	installer, err := NewSystemInstaller()
	if err != nil {
		t.Fatalf("Failed to create system installer: %v", err)
	}
	
	// Create a temporary source file
	tempDir := t.TempDir()
	srcFile := filepath.Join(tempDir, "source.txt")
	dstFile := filepath.Join(tempDir, "destination.txt")
	
	testContent := "test content"
	if err := os.WriteFile(srcFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Test copy
	if err := installer.copyBinary(srcFile, dstFile); err != nil {
		t.Fatalf("Failed to copy binary: %v", err)
	}
	
	// Verify copy
	copiedContent, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("Failed to read copied file: %v", err)
	}
	
	if string(copiedContent) != testContent {
		t.Errorf("Copied content mismatch: got %s, expected %s", string(copiedContent), testContent)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		containsAt(s, substr))))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}