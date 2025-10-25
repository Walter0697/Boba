package installer

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"boba/internal/parser"
)

// Platform represents the target platform information
type Platform struct {
	OS             string
	Distribution   string
	PackageManager string
}

// InstallationResult represents the result of an installation operation
type InstallationResult struct {
	Success    bool
	Output     string
	Error      error
	ExitCode   int
	Duration   time.Duration
}

// InstallationEngine handles cross-platform tool installation
type InstallationEngine struct {
	platform     Platform
	githubClient GitHubClientInterface
	tempDir      string
}

// NewInstallationEngine creates a new installation engine instance
func NewInstallationEngine(githubClient GitHubClientInterface) *InstallationEngine {
	tempDir := filepath.Join(os.TempDir(), "boba-installer")
	os.MkdirAll(tempDir, 0755)
	
	return &InstallationEngine{
		platform:     detectPlatform(),
		githubClient: githubClient,
		tempDir:      tempDir,
	}
}

// detectPlatform determines the current platform
func detectPlatform() Platform {
	platform := Platform{
		OS: runtime.GOOS,
	}
	
	// Detect distribution and package manager for Linux
	if platform.OS == "linux" {
		platform.Distribution = detectLinuxDistribution()
		platform.PackageManager = detectPackageManager()
	} else if platform.OS == "darwin" {
		platform.PackageManager = "brew"
	}
	
	return platform
}

// detectLinuxDistribution attempts to detect the Linux distribution
func detectLinuxDistribution() string {
	// Check /etc/os-release first
	if content, err := os.ReadFile("/etc/os-release"); err == nil {
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "ID=") {
				return strings.Trim(strings.TrimPrefix(line, "ID="), "\"")
			}
		}
	}
	
	// Fallback checks
	if _, err := os.Stat("/etc/debian_version"); err == nil {
		return "debian"
	}
	if _, err := os.Stat("/etc/redhat-release"); err == nil {
		return "rhel"
	}
	if _, err := os.Stat("/etc/arch-release"); err == nil {
		return "arch"
	}
	
	return "unknown"
}

// detectPackageManager attempts to detect the available package manager
func detectPackageManager() string {
	managers := []string{"apt", "yum", "dnf", "pacman", "zypper", "apk"}
	
	for _, manager := range managers {
		if _, err := exec.LookPath(manager); err == nil {
			return manager
		}
	}
	
	return "unknown"
}

// IsToolInstalled checks if a tool is already installed on the system
func (ie *InstallationEngine) IsToolInstalled(tool parser.Tool) bool {
	// First, try to check if the tool name is available in PATH
	if _, err := exec.LookPath(tool.Name); err == nil {
		return true
	}
	
	// Try common variations of the tool name
	variations := []string{
		tool.Name,
		strings.ToLower(tool.Name),
		strings.ReplaceAll(tool.Name, "-", ""),
		strings.ReplaceAll(tool.Name, "_", ""),
	}
	
	for _, variation := range variations {
		if _, err := exec.LookPath(variation); err == nil {
			return true
		}
	}
	
	return false
}

// InstallTool installs a tool using its install script from the repository
func (ie *InstallationEngine) InstallTool(tool parser.Tool) (*InstallationResult, error) {
	if ie.githubClient == nil {
		return nil, fmt.Errorf("GitHub client not initialized")
	}
	
	startTime := time.Now()
	
	// Download the install script
	scriptContent, err := ie.githubClient.GetRepositoryContents(tool.InstallScript)
	if err != nil {
		return &InstallationResult{
			Success:  false,
			Error:    fmt.Errorf("failed to download install script: %w", err),
			Duration: time.Since(startTime),
		}, err
	}
	
	// Create a temporary script file
	scriptPath := filepath.Join(ie.tempDir, fmt.Sprintf("install_%s.sh", tool.FolderName))
	if err := os.WriteFile(scriptPath, scriptContent, 0755); err != nil {
		return &InstallationResult{
			Success:  false,
			Error:    fmt.Errorf("failed to create script file: %w", err),
			Duration: time.Since(startTime),
		}, err
	}
	
	// Ensure cleanup
	defer os.Remove(scriptPath)
	
	// Execute the script with security measures
	result := ie.executeScriptSecurely(scriptPath, tool.Name)
	result.Duration = time.Since(startTime)
	
	return result, result.Error
}

// UninstallTool uninstalls a tool using its uninstall script from the repository
func (ie *InstallationEngine) UninstallTool(tool parser.Tool) (*InstallationResult, error) {
	if ie.githubClient == nil {
		return nil, fmt.Errorf("GitHub client not initialized")
	}
	
	startTime := time.Now()
	
	// Download the uninstall script
	scriptContent, err := ie.githubClient.GetRepositoryContents(tool.UninstallScript)
	if err != nil {
		return &InstallationResult{
			Success:  false,
			Error:    fmt.Errorf("failed to download uninstall script: %w", err),
			Duration: time.Since(startTime),
		}, err
	}
	
	// Create a temporary script file
	scriptPath := filepath.Join(ie.tempDir, fmt.Sprintf("uninstall_%s.sh", tool.FolderName))
	if err := os.WriteFile(scriptPath, scriptContent, 0755); err != nil {
		return &InstallationResult{
			Success:  false,
			Error:    fmt.Errorf("failed to create script file: %w", err),
			Duration: time.Since(startTime),
		}, err
	}
	
	// Ensure cleanup
	defer os.Remove(scriptPath)
	
	// Execute the script with security measures
	result := ie.executeScriptSecurely(scriptPath, tool.Name)
	result.Duration = time.Since(startTime)
	
	return result, result.Error
}

// executeScriptSecurely executes a script with proper security measures and output capture
func (ie *InstallationEngine) executeScriptSecurely(scriptPath, toolName string) *InstallationResult {
	// Create context with timeout (10 minutes max per installation)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	
	// Prepare the command
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// On Windows, use PowerShell or cmd to execute scripts
		cmd = exec.CommandContext(ctx, "powershell", "-ExecutionPolicy", "Bypass", "-File", scriptPath)
	} else {
		// On Unix-like systems, use bash
		cmd = exec.CommandContext(ctx, "/bin/bash", scriptPath)
	}
	
	// Set up environment variables
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("BOBA_TOOL_NAME=%s", toolName),
		fmt.Sprintf("BOBA_PLATFORM=%s", ie.platform.OS),
		fmt.Sprintf("BOBA_PACKAGE_MANAGER=%s", ie.platform.PackageManager),
		fmt.Sprintf("BOBA_TEMP_DIR=%s", ie.tempDir),
		fmt.Sprintf("TMPDIR=%s", ie.tempDir),
		fmt.Sprintf("TEMP=%s", ie.tempDir),
		fmt.Sprintf("TMP=%s", ie.tempDir),
	)
	
	// Set working directory to temp directory
	cmd.Dir = ie.tempDir
	
	// Create pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return &InstallationResult{
			Success: false,
			Error:   fmt.Errorf("failed to create stdout pipe: %w", err),
		}
	}
	
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return &InstallationResult{
			Success: false,
			Error:   fmt.Errorf("failed to create stderr pipe: %w", err),
		}
	}
	
	// Start the command
	if err := cmd.Start(); err != nil {
		return &InstallationResult{
			Success: false,
			Error:   fmt.Errorf("failed to start script: %w", err),
		}
	}
	
	// Capture output
	var outputBuilder strings.Builder
	
	// Read stdout and stderr concurrently
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			outputBuilder.WriteString("STDOUT: " + line + "\n")
		}
	}()
	
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			outputBuilder.WriteString("STDERR: " + line + "\n")
		}
	}()
	
	// Wait for the command to complete
	err = cmd.Wait()
	
	// Get exit code
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
			}
		}
	}
	
	output := outputBuilder.String()
	success := exitCode == 0
	
	result := &InstallationResult{
		Success:  success,
		Output:   output,
		ExitCode: exitCode,
	}
	
	if !success {
		result.Error = fmt.Errorf("script execution failed with exit code %d: %s", exitCode, output)
	}
	
	return result
}

// ExecuteCommand executes a shell command and returns the output
func (ie *InstallationEngine) ExecuteCommand(command string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/C", command)
	} else {
		cmd = exec.CommandContext(ctx, "/bin/bash", "-c", command)
	}
	
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// VerifyInstallation verifies that a tool was successfully installed
func (ie *InstallationEngine) VerifyInstallation(tool parser.Tool) (bool, string) {
	// First check if the tool is now available in PATH
	if ie.IsToolInstalled(tool) {
		// Try to get version information
		versionOutput, err := ie.ExecuteCommand(fmt.Sprintf("%s --version", tool.Name))
		if err == nil && versionOutput != "" {
			return true, fmt.Sprintf("Tool '%s' is installed and accessible. Version info: %s", tool.Name, strings.TrimSpace(versionOutput))
		}
		
		// Try alternative version commands
		for _, versionCmd := range []string{"-v", "version", "--help"} {
			versionOutput, err := ie.ExecuteCommand(fmt.Sprintf("%s %s", tool.Name, versionCmd))
			if err == nil && versionOutput != "" {
				return true, fmt.Sprintf("Tool '%s' is installed and accessible. Info: %s", tool.Name, strings.TrimSpace(versionOutput))
			}
		}
		
		return true, fmt.Sprintf("Tool '%s' is installed and accessible in PATH", tool.Name)
	}
	
	return false, fmt.Sprintf("Tool '%s' is not accessible in PATH after installation", tool.Name)
}

// GetPlatform returns the detected platform information
func (ie *InstallationEngine) GetPlatform() Platform {
	return ie.platform
}

// Cleanup removes temporary files and directories
func (ie *InstallationEngine) Cleanup() error {
	return os.RemoveAll(ie.tempDir)
}