package installer

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// SystemInstaller handles system-level installation of the BOBA binary
type SystemInstaller struct {
	binaryPath     string
	installPath    string
	backupPath     string
	zshrcPath      string
	zshrcBackupPath string
}

// SystemInstallationResult represents the result of system installation
type SystemInstallationResult struct {
	Success         bool
	BinaryInstalled bool
	ZshrcModified   bool
	BackupCreated   bool
	Message         string
	Error           error
	Duration        time.Duration
}

// NewSystemInstaller creates a new system installer instance
func NewSystemInstaller() (*SystemInstaller, error) {
	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %w", err)
	}
	
	// Get user home directory
	currentUser, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}
	
	installPath := "/usr/local/bin/boba"
	if runtime.GOOS == "windows" {
		// On Windows, install to a directory in PATH or create one
		installPath = filepath.Join(os.Getenv("PROGRAMFILES"), "BOBA", "boba.exe")
	}
	
	return &SystemInstaller{
		binaryPath:      execPath,
		installPath:     installPath,
		backupPath:      installPath + ".backup",
		zshrcPath:       filepath.Join(currentUser.HomeDir, ".zshrc"),
		zshrcBackupPath: filepath.Join(currentUser.HomeDir, ".zshrc.boba.backup"),
	}, nil
}

// IsSystemInstalled checks if BOBA is already installed system-wide
func (si *SystemInstaller) IsSystemInstalled() bool {
	// Check if binary exists in install location
	if _, err := os.Stat(si.installPath); err != nil {
		return false
	}
	
	// Check if it's accessible via PATH
	if _, err := exec.LookPath("boba"); err != nil {
		return false
	}
	
	return true
}

// RequiresSudo checks if sudo privileges are needed for installation
func (si *SystemInstaller) RequiresSudo() bool {
	if runtime.GOOS == "windows" {
		return false // Windows handles elevation differently
	}
	
	// Check if we can write to the install directory
	installDir := filepath.Dir(si.installPath)
	testFile := filepath.Join(installDir, ".boba_test_write")
	
	// Try to create a test file
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return true // Need sudo
	}
	
	// Clean up test file
	os.Remove(testFile)
	return false
}

// InstallToSystem installs BOBA binary to system location and sets up shell integration
func (si *SystemInstaller) InstallToSystem() (*SystemInstallationResult, error) {
	startTime := time.Now()
	result := &SystemInstallationResult{}
	
	// Step 1: Install binary
	if err := si.installBinary(); err != nil {
		result.Error = fmt.Errorf("failed to install binary: %w", err)
		result.Duration = time.Since(startTime)
		return result, result.Error
	}
	result.BinaryInstalled = true
	
	// Step 2: Setup shell integration
	if err := si.setupShellIntegration(); err != nil {
		result.Error = fmt.Errorf("failed to setup shell integration: %w", err)
		result.Duration = time.Since(startTime)
		return result, result.Error
	}
	result.ZshrcModified = true
	
	// Step 3: Verify installation
	if err := si.verifyInstallation(); err != nil {
		result.Error = fmt.Errorf("installation verification failed: %w", err)
		result.Duration = time.Since(startTime)
		return result, result.Error
	}
	
	result.Success = true
	result.Message = "BOBA successfully installed to system. Restart your shell or run 'source ~/.zshrc' to use the 'boba' command."
	result.Duration = time.Since(startTime)
	
	return result, nil
}

// installBinary copies the BOBA binary to the system location
func (si *SystemInstaller) installBinary() error {
	// Create backup if existing installation exists
	if _, err := os.Stat(si.installPath); err == nil {
		if err := si.createBinaryBackup(); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
	}
	
	// Ensure install directory exists
	installDir := filepath.Dir(si.installPath)
	if err := os.MkdirAll(installDir, 0755); err != nil {
		if si.RequiresSudo() {
			return si.installBinaryWithSudo()
		}
		return fmt.Errorf("failed to create install directory: %w", err)
	}
	
	// Copy binary
	if err := si.copyBinary(si.binaryPath, si.installPath); err != nil {
		if si.RequiresSudo() {
			return si.installBinaryWithSudo()
		}
		return fmt.Errorf("failed to copy binary: %w", err)
	}
	
	// Make executable
	if err := os.Chmod(si.installPath, 0755); err != nil {
		return fmt.Errorf("failed to make binary executable: %w", err)
	}
	
	return nil
}

// installBinaryWithSudo installs the binary using sudo privileges
func (si *SystemInstaller) installBinaryWithSudo() error {
	if runtime.GOOS == "windows" {
		return si.installBinaryWindows()
	}
	
	// Create install directory with sudo
	installDir := filepath.Dir(si.installPath)
	cmd := exec.Command("sudo", "mkdir", "-p", installDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create install directory with sudo: %w", err)
	}
	
	// Copy binary with sudo
	cmd = exec.Command("sudo", "cp", si.binaryPath, si.installPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to copy binary with sudo: %w", err)
	}
	
	// Make executable with sudo
	cmd = exec.Command("sudo", "chmod", "755", si.installPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to make binary executable with sudo: %w", err)
	}
	
	return nil
}

// installBinaryWindows handles Windows-specific installation
func (si *SystemInstaller) installBinaryWindows() error {
	// On Windows, we might need to handle UAC elevation
	// For now, try to install to a user-accessible location
	userBinDir := filepath.Join(os.Getenv("USERPROFILE"), "bin")
	si.installPath = filepath.Join(userBinDir, "boba.exe")
	
	// Create directory
	if err := os.MkdirAll(userBinDir, 0755); err != nil {
		return fmt.Errorf("failed to create user bin directory: %w", err)
	}
	
	// Copy binary
	if err := si.copyBinary(si.binaryPath, si.installPath); err != nil {
		return fmt.Errorf("failed to copy binary: %w", err)
	}
	
	return nil
}

// copyBinary copies a file from source to destination
func (si *SystemInstaller) copyBinary(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()
	
	// Copy file contents
	_, err = destFile.ReadFrom(sourceFile)
	return err
}

// createBinaryBackup creates a backup of existing binary
func (si *SystemInstaller) createBinaryBackup() error {
	return si.copyBinary(si.installPath, si.backupPath)
}

// setupShellIntegration modifies ~/.zshrc to add BOBA to PATH and create alias
func (si *SystemInstaller) setupShellIntegration() error {
	// Check if zsh is available
	if _, err := exec.LookPath("zsh"); err != nil {
		return fmt.Errorf("zsh is not installed or not in PATH")
	}
	
	// Create backup of .zshrc if it exists
	if _, err := os.Stat(si.zshrcPath); err == nil {
		if err := si.copyBinary(si.zshrcPath, si.zshrcBackupPath); err != nil {
			return fmt.Errorf("failed to backup .zshrc: %w", err)
		}
	}
	
	// Read existing .zshrc content
	var existingContent string
	if content, err := os.ReadFile(si.zshrcPath); err == nil {
		existingContent = string(content)
	}
	
	// Check if BOBA configuration already exists
	bobaMarker := "# BOBA CLI Tool Configuration"
	if strings.Contains(existingContent, bobaMarker) {
		// Configuration already exists, update it
		return si.updateExistingConfiguration(existingContent)
	}
	
	// Add BOBA configuration
	bobaConfig := si.generateBobaConfiguration()
	
	// Append to .zshrc
	file, err := os.OpenFile(si.zshrcPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open .zshrc for writing: %w", err)
	}
	defer file.Close()
	
	if _, err := file.WriteString(bobaConfig); err != nil {
		return fmt.Errorf("failed to write BOBA configuration to .zshrc: %w", err)
	}
	
	return nil
}

// generateBobaConfiguration generates the shell configuration for BOBA
func (si *SystemInstaller) generateBobaConfiguration() string {
	installDir := filepath.Dir(si.installPath)
	
	config := fmt.Sprintf(`

# BOBA CLI Tool Configuration
# Added by BOBA installer on %s
export PATH="%s:$PATH"
alias boba-update="boba"
alias dev-setup="boba"

# BOBA completion (if available)
if command -v boba >/dev/null 2>&1; then
    # Add any completion setup here in the future
    :
fi
`, time.Now().Format("2006-01-02 15:04:05"), installDir)
	
	return config
}

// updateExistingConfiguration updates existing BOBA configuration in .zshrc
func (si *SystemInstaller) updateExistingConfiguration(content string) error {
	lines := strings.Split(content, "\n")
	var newLines []string
	inBobaSection := false
	
	for _, line := range lines {
		if strings.Contains(line, "# BOBA CLI Tool Configuration") {
			inBobaSection = true
			// Add updated configuration
			newLines = append(newLines, si.generateBobaConfiguration())
			continue
		}
		
		if inBobaSection && strings.HasPrefix(line, "#") && !strings.Contains(line, "BOBA") {
			inBobaSection = false
		}
		
		if !inBobaSection {
			newLines = append(newLines, line)
		}
	}
	
	// Write updated content
	return os.WriteFile(si.zshrcPath, []byte(strings.Join(newLines, "\n")), 0644)
}

// verifyInstallation verifies that the system installation was successful
func (si *SystemInstaller) verifyInstallation() error {
	// Check if binary exists and is executable
	if _, err := os.Stat(si.installPath); err != nil {
		return fmt.Errorf("binary not found at %s: %w", si.installPath, err)
	}
	
	// Check if binary is accessible via PATH (this might not work immediately due to shell not being reloaded)
	if _, err := exec.LookPath("boba"); err != nil {
		// This is expected if shell hasn't been reloaded, so just warn
		fmt.Printf("Warning: 'boba' command not immediately available in PATH. Restart your shell or run 'source ~/.zshrc'\n")
	}
	
	// Verify .zshrc was modified
	if content, err := os.ReadFile(si.zshrcPath); err == nil {
		if !strings.Contains(string(content), "# BOBA CLI Tool Configuration") {
			return fmt.Errorf(".zshrc was not properly modified")
		}
	} else {
		return fmt.Errorf("failed to verify .zshrc modification: %w", err)
	}
	
	return nil
}

// UninstallFromSystem removes BOBA from system and reverts shell integration
func (si *SystemInstaller) UninstallFromSystem() (*SystemInstallationResult, error) {
	startTime := time.Now()
	result := &SystemInstallationResult{}
	
	// Remove binary
	if err := si.removeBinary(); err != nil {
		result.Error = fmt.Errorf("failed to remove binary: %w", err)
		result.Duration = time.Since(startTime)
		return result, result.Error
	}
	
	// Restore .zshrc
	if err := si.restoreShellConfiguration(); err != nil {
		result.Error = fmt.Errorf("failed to restore shell configuration: %w", err)
		result.Duration = time.Since(startTime)
		return result, result.Error
	}
	
	result.Success = true
	result.Message = "BOBA successfully uninstalled from system. Restart your shell to complete the removal."
	result.Duration = time.Since(startTime)
	
	return result, nil
}

// removeBinary removes the installed binary
func (si *SystemInstaller) removeBinary() error {
	if _, err := os.Stat(si.installPath); err != nil {
		return nil // Already removed
	}
	
	// Try to remove normally first
	if err := os.Remove(si.installPath); err != nil {
		if si.RequiresSudo() {
			// Use sudo to remove
			cmd := exec.Command("sudo", "rm", si.installPath)
			return cmd.Run()
		}
		return err
	}
	
	return nil
}

// restoreShellConfiguration restores the original .zshrc
func (si *SystemInstaller) restoreShellConfiguration() error {
	// Check if backup exists
	if _, err := os.Stat(si.zshrcBackupPath); err != nil {
		// No backup, try to remove BOBA configuration manually
		return si.removeBobaConfiguration()
	}
	
	// Restore from backup
	return si.copyBinary(si.zshrcBackupPath, si.zshrcPath)
}

// removeBobaConfiguration removes BOBA configuration from .zshrc
func (si *SystemInstaller) removeBobaConfiguration() error {
	content, err := os.ReadFile(si.zshrcPath)
	if err != nil {
		return err
	}
	
	lines := strings.Split(string(content), "\n")
	var newLines []string
	inBobaSection := false
	
	for _, line := range lines {
		if strings.Contains(line, "# BOBA CLI Tool Configuration") {
			inBobaSection = true
			continue
		}
		
		if inBobaSection && (strings.TrimSpace(line) == "" || (!strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "export") && !strings.HasPrefix(line, "alias"))) {
			inBobaSection = false
		}
		
		if !inBobaSection {
			newLines = append(newLines, line)
		}
	}
	
	return os.WriteFile(si.zshrcPath, []byte(strings.Join(newLines, "\n")), 0644)
}

// GetInstallationInfo returns information about the current installation
func (si *SystemInstaller) GetInstallationInfo() map[string]interface{} {
	info := map[string]interface{}{
		"binary_path":       si.binaryPath,
		"install_path":      si.installPath,
		"zshrc_path":        si.zshrcPath,
		"is_installed":      si.IsSystemInstalled(),
		"requires_sudo":     si.RequiresSudo(),
		"platform":          runtime.GOOS,
	}
	
	// Check if backup exists
	if _, err := os.Stat(si.backupPath); err == nil {
		info["has_backup"] = true
	} else {
		info["has_backup"] = false
	}
	
	return info
}