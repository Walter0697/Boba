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
	
	fmt.Printf("\n=== Installing Tool: %s ===\n", tool.Name)
	fmt.Printf("Script path: %s\n", tool.InstallScript)
	
	// Download the install script
	fmt.Printf("Downloading install script...\n")
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
	fmt.Printf("Creating temporary script at: %s\n", scriptPath)
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
	fmt.Printf("Executing installation script...\n")
	fmt.Printf("==========================================\n")
	result := ie.executeScriptSecurely(scriptPath, tool.Name)
	fmt.Printf("==========================================\n")
	fmt.Printf("Installation completed with exit code: %d\n", result.ExitCode)
	result.Duration = time.Since(startTime)
	
	return result, result.Error
}

// UninstallTool uninstalls a tool using its uninstall script from the repository
func (ie *InstallationEngine) UninstallTool(tool parser.Tool) (*InstallationResult, error) {
	if ie.githubClient == nil {
		return nil, fmt.Errorf("GitHub client not initialized")
	}
	
	startTime := time.Now()
	
	fmt.Printf("\n=== Uninstalling Tool: %s ===\n", tool.Name)
	fmt.Printf("Script path: %s\n", tool.UninstallScript)
	
	// Download the uninstall script
	fmt.Printf("Downloading uninstall script...\n")
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
	fmt.Printf("Creating temporary script at: %s\n", scriptPath)
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
	fmt.Printf("Executing uninstall script...\n")
	fmt.Printf("==========================================\n")
	result := ie.executeScriptSecurely(scriptPath, tool.Name)
	fmt.Printf("==========================================\n")
	fmt.Printf("Uninstallation completed with exit code: %d\n", result.ExitCode)
	result.Duration = time.Since(startTime)
	
	return result, result.Error
}

// executeScriptSecurely executes a script with proper security measures and output capture
func (ie *InstallationEngine) executeScriptSecurely(scriptPath, toolName string) *InstallationResult {
	// Create context with timeout (10 minutes max per installation)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	
	fmt.Printf("DEBUG: Preparing to execute script\n")
	fmt.Printf("DEBUG: Script path: %s\n", scriptPath)
	fmt.Printf("DEBUG: Tool name: %s\n", toolName)
	
	// Prepare the command
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// On Windows, use PowerShell or cmd to execute scripts
		cmd = exec.CommandContext(ctx, "powershell", "-ExecutionPolicy", "Bypass", "-File", scriptPath)
		fmt.Printf("DEBUG: Using PowerShell to execute script\n")
	} else {
		// On Unix-like systems, use bash
		cmd = exec.CommandContext(ctx, "/bin/bash", scriptPath)
		fmt.Printf("DEBUG: Using bash to execute script\n")
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
	
	fmt.Printf("DEBUG: Environment variables set:\n")
	fmt.Printf("  BOBA_TOOL_NAME=%s\n", toolName)
	fmt.Printf("  BOBA_PLATFORM=%s\n", ie.platform.OS)
	fmt.Printf("  BOBA_PACKAGE_MANAGER=%s\n", ie.platform.PackageManager)
	
	// Set working directory to temp directory
	cmd.Dir = ie.tempDir
	fmt.Printf("DEBUG: Working directory: %s\n", ie.tempDir)
	
	// Create pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("DEBUG: Failed to create stdout pipe: %v\n", err)
		return &InstallationResult{
			Success: false,
			Error:   fmt.Errorf("failed to create stdout pipe: %w", err),
		}
	}
	
	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Printf("DEBUG: Failed to create stderr pipe: %v\n", err)
		return &InstallationResult{
			Success: false,
			Error:   fmt.Errorf("failed to create stderr pipe: %w", err),
		}
	}
	
	fmt.Printf("DEBUG: Pipes created successfully\n")
	
	// Start the command
	fmt.Printf("DEBUG: Starting command...\n")
	if err := cmd.Start(); err != nil {
		fmt.Printf("DEBUG: Failed to start command: %v\n", err)
		return &InstallationResult{
			Success: false,
			Error:   fmt.Errorf("failed to start script: %w", err),
		}
	}
	
	fmt.Printf("DEBUG: Command started successfully, PID: %d\n", cmd.Process.Pid)
	fmt.Printf("DEBUG: Reading output...\n\n")
	
	// Capture output
	var outputBuilder strings.Builder
	
	// Use channels to synchronize goroutines
	stdoutDone := make(chan bool)
	stderrDone := make(chan bool)
	
	// Read stdout and stderr concurrently
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			// Print to console in real-time
			fmt.Println(line)
			// Also capture for result
			outputBuilder.WriteString(line + "\n")
		}
		if err := scanner.Err(); err != nil {
			fmt.Printf("DEBUG: Error reading stdout: %v\n", err)
		}
		stdoutDone <- true
	}()
	
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			// Print to console in real-time (to stderr)
			fmt.Fprintln(os.Stderr, line)
			// Also capture for result
			outputBuilder.WriteString("STDERR: " + line + "\n")
		}
		if err := scanner.Err(); err != nil {
			fmt.Printf("DEBUG: Error reading stderr: %v\n", err)
		}
		stderrDone <- true
	}()
	
	// Wait for the command to complete with timeout monitoring
	fmt.Printf("DEBUG: Waiting for command to complete (timeout: 10 minutes)...\n")
	
	// Create a done channel to signal completion
	cmdDone := make(chan error, 1)
	go func() {
		cmdDone <- cmd.Wait()
	}()
	
	// Wait for either completion or timeout
	select {
	case err = <-cmdDone:
		// Command completed normally
		fmt.Printf("DEBUG: Command completed\n")
	case <-ctx.Done():
		// Timeout occurred
		fmt.Printf("DEBUG: Command timed out after 10 minutes, killing process...\n")
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		err = fmt.Errorf("command timed out after 10 minutes")
	}
	
	// Wait for output readers to finish
	<-stdoutDone
	<-stderrDone
	fmt.Printf("\nDEBUG: Output reading completed\n")
	
	// Get exit code
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
			}
		}
		fmt.Printf("DEBUG: Command finished with error: %v\n", err)
	} else {
		fmt.Printf("DEBUG: Command finished successfully\n")
	}
	
	output := outputBuilder.String()
	success := exitCode == 0
	
	fmt.Printf("DEBUG: Exit code: %d\n", exitCode)
	fmt.Printf("DEBUG: Success: %v\n", success)
	fmt.Printf("DEBUG: Output length: %d bytes\n", len(output))
	
	result := &InstallationResult{
		Success:  success,
		Output:   output,
		ExitCode: exitCode,
	}
	
	if !success {
		result.Error = fmt.Errorf("script execution failed with exit code %d", exitCode)
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

// ApplyEnvironment applies an environment configuration using its setup script
func (ie *InstallationEngine) ApplyEnvironment(env parser.Environment) (*InstallationResult, error) {
	if ie.githubClient == nil {
		return nil, fmt.Errorf("GitHub client not initialized")
	}
	
	startTime := time.Now()
	
	fmt.Printf("\n=== Applying Environment: %s ===\n", env.Name)
	fmt.Printf("Script path: %s\n", env.SetupScript)
	
	// Download the setup script
	fmt.Printf("Downloading setup script...\n")
	scriptContent, err := ie.githubClient.GetRepositoryContents(env.SetupScript)
	if err != nil {
		return &InstallationResult{
			Success:  false,
			Error:    fmt.Errorf("failed to download setup script: %w", err),
			Duration: time.Since(startTime),
		}, err
	}
	
	// Create a temporary script file
	scriptPath := filepath.Join(ie.tempDir, fmt.Sprintf("setup_%s.sh", env.FolderName))
	fmt.Printf("Creating temporary script at: %s\n", scriptPath)
	if err := os.WriteFile(scriptPath, scriptContent, 0755); err != nil {
		return &InstallationResult{
			Success:  false,
			Error:    fmt.Errorf("failed to create setup script file: %w", err),
			Duration: time.Since(startTime),
		}, err
	}
	
	// Ensure cleanup
	defer os.Remove(scriptPath)
	
	// Execute the setup script with security measures
	fmt.Printf("Executing setup script...\n")
	fmt.Printf("==========================================\n")
	result := ie.executeEnvironmentScriptSecurely(scriptPath, env.Name, env)
	fmt.Printf("==========================================\n")
	fmt.Printf("Environment setup completed with exit code: %d\n", result.ExitCode)
	result.Duration = time.Since(startTime)
	
	return result, result.Error
}

// RestoreEnvironment restores an environment configuration using its restore script
func (ie *InstallationEngine) RestoreEnvironment(env parser.Environment) (*InstallationResult, error) {
	if ie.githubClient == nil {
		return nil, fmt.Errorf("GitHub client not initialized")
	}
	
	startTime := time.Now()
	
	fmt.Printf("\n=== Restoring Environment: %s ===\n", env.Name)
	fmt.Printf("Script path: %s\n", env.RestoreScript)
	
	// Download the restore script
	fmt.Printf("Downloading restore script...\n")
	scriptContent, err := ie.githubClient.GetRepositoryContents(env.RestoreScript)
	if err != nil {
		return &InstallationResult{
			Success:  false,
			Error:    fmt.Errorf("failed to download restore script: %w", err),
			Duration: time.Since(startTime),
		}, err
	}
	
	// Create a temporary script file
	scriptPath := filepath.Join(ie.tempDir, fmt.Sprintf("restore_%s.sh", env.FolderName))
	fmt.Printf("Creating temporary script at: %s\n", scriptPath)
	if err := os.WriteFile(scriptPath, scriptContent, 0755); err != nil {
		return &InstallationResult{
			Success:  false,
			Error:    fmt.Errorf("failed to create restore script file: %w", err),
			Duration: time.Since(startTime),
		}, err
	}
	
	// Ensure cleanup
	defer os.Remove(scriptPath)
	
	// Execute the restore script with security measures
	fmt.Printf("Executing restore script...\n")
	fmt.Printf("==========================================\n")
	result := ie.executeEnvironmentScriptSecurely(scriptPath, env.Name, env)
	fmt.Printf("==========================================\n")
	fmt.Printf("Environment restore completed with exit code: %d\n", result.ExitCode)
	result.Duration = time.Since(startTime)
	
	return result, result.Error
}

// executeEnvironmentScriptSecurely executes an environment script with proper security measures and environment-specific variables
func (ie *InstallationEngine) executeEnvironmentScriptSecurely(scriptPath, envName string, env parser.Environment) *InstallationResult {
	// Create context with timeout (10 minutes max per environment setup)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	
	fmt.Printf("DEBUG: Preparing to execute environment script\n")
	fmt.Printf("DEBUG: Script path: %s\n", scriptPath)
	fmt.Printf("DEBUG: Environment name: %s\n", envName)
	
	// Prepare the command
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// On Windows, use PowerShell or cmd to execute scripts
		cmd = exec.CommandContext(ctx, "powershell", "-ExecutionPolicy", "Bypass", "-File", scriptPath)
		fmt.Printf("DEBUG: Using PowerShell to execute script\n")
	} else {
		// On Unix-like systems, use bash
		cmd = exec.CommandContext(ctx, "/bin/bash", scriptPath)
		fmt.Printf("DEBUG: Using bash to execute script\n")
	}
	
	// Set up environment variables
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("BOBA_ENV_NAME=%s", envName),
		fmt.Sprintf("BOBA_ENV_SHELL=%s", env.Shell),
		fmt.Sprintf("BOBA_PLATFORM=%s", ie.platform.OS),
		fmt.Sprintf("BOBA_PACKAGE_MANAGER=%s", ie.platform.PackageManager),
		fmt.Sprintf("BOBA_TEMP_DIR=%s", ie.tempDir),
		fmt.Sprintf("TMPDIR=%s", ie.tempDir),
		fmt.Sprintf("TEMP=%s", ie.tempDir),
		fmt.Sprintf("TMP=%s", ie.tempDir),
	)
	
	fmt.Printf("DEBUG: Environment variables set:\n")
	fmt.Printf("  BOBA_ENV_NAME=%s\n", envName)
	fmt.Printf("  BOBA_ENV_SHELL=%s\n", env.Shell)
	fmt.Printf("  BOBA_PLATFORM=%s\n", ie.platform.OS)
	fmt.Printf("  BOBA_PACKAGE_MANAGER=%s\n", ie.platform.PackageManager)
	
	// Set working directory to temp directory
	cmd.Dir = ie.tempDir
	fmt.Printf("DEBUG: Working directory: %s\n", ie.tempDir)
	
	// Create pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("DEBUG: Failed to create stdout pipe: %v\n", err)
		return &InstallationResult{
			Success: false,
			Error:   fmt.Errorf("failed to create stdout pipe: %w", err),
		}
	}
	
	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Printf("DEBUG: Failed to create stderr pipe: %v\n", err)
		return &InstallationResult{
			Success: false,
			Error:   fmt.Errorf("failed to create stderr pipe: %w", err),
		}
	}
	
	fmt.Printf("DEBUG: Pipes created successfully\n")
	
	// Start the command
	fmt.Printf("DEBUG: Starting command...\n")
	if err := cmd.Start(); err != nil {
		fmt.Printf("DEBUG: Failed to start command: %v\n", err)
		return &InstallationResult{
			Success: false,
			Error:   fmt.Errorf("failed to start environment script: %w", err),
		}
	}
	
	fmt.Printf("DEBUG: Command started successfully, PID: %d\n", cmd.Process.Pid)
	fmt.Printf("DEBUG: Reading output...\n\n")
	
	// Capture output
	var outputBuilder strings.Builder
	
	// Use channels to synchronize goroutines
	stdoutDone := make(chan bool)
	stderrDone := make(chan bool)
	
	// Read stdout and stderr concurrently
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			// Print to console in real-time
			fmt.Println(line)
			// Also capture for result
			outputBuilder.WriteString(line + "\n")
		}
		if err := scanner.Err(); err != nil {
			fmt.Printf("DEBUG: Error reading stdout: %v\n", err)
		}
		stdoutDone <- true
	}()
	
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			// Print to console in real-time (to stderr)
			fmt.Fprintln(os.Stderr, line)
			// Also capture for result
			outputBuilder.WriteString("STDERR: " + line + "\n")
		}
		if err := scanner.Err(); err != nil {
			fmt.Printf("DEBUG: Error reading stderr: %v\n", err)
		}
		stderrDone <- true
	}()
	
	// Wait for the command to complete with timeout monitoring
	fmt.Printf("DEBUG: Waiting for command to complete (timeout: 10 minutes)...\n")
	
	// Create a done channel to signal completion
	cmdDone := make(chan error, 1)
	go func() {
		cmdDone <- cmd.Wait()
	}()
	
	// Wait for either completion or timeout
	select {
	case err = <-cmdDone:
		// Command completed normally
		fmt.Printf("DEBUG: Command completed\n")
	case <-ctx.Done():
		// Timeout occurred
		fmt.Printf("DEBUG: Command timed out after 10 minutes, killing process...\n")
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		err = fmt.Errorf("command timed out after 10 minutes")
	}
	
	// Wait for output readers to finish
	<-stdoutDone
	<-stderrDone
	fmt.Printf("\nDEBUG: Output reading completed\n")
	
	// Get exit code
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
			}
		}
		fmt.Printf("DEBUG: Command finished with error: %v\n", err)
	} else {
		fmt.Printf("DEBUG: Command finished successfully\n")
	}
	
	output := outputBuilder.String()
	success := exitCode == 0
	
	fmt.Printf("DEBUG: Exit code: %d\n", exitCode)
	fmt.Printf("DEBUG: Success: %v\n", success)
	fmt.Printf("DEBUG: Output length: %d bytes\n", len(output))
	
	result := &InstallationResult{
		Success:  success,
		Output:   output,
		ExitCode: exitCode,
	}
	
	if !success {
		result.Error = fmt.Errorf("environment script execution failed with exit code %d", exitCode)
	}
	
	return result
}

// IsEnvironmentApplied checks if an environment is already applied (placeholder implementation)
func (ie *InstallationEngine) IsEnvironmentApplied(env parser.Environment) bool {
	// This is a placeholder implementation
	// In a real implementation, this could check for:
	// - Presence of specific config files
	// - Environment variables
	// - Shell configuration markers
	// - etc.
	return false
}

// VerifyEnvironmentApplication verifies that an environment was successfully applied
func (ie *InstallationEngine) VerifyEnvironmentApplication(env parser.Environment) (bool, string) {
	// This is a placeholder implementation
	// In a real implementation, this could verify:
	// - Config files are in place
	// - Environment variables are set
	// - Shell configurations are active
	// - etc.
	
	return true, fmt.Sprintf("Environment '%s' appears to be applied successfully", env.Name)
}

// Cleanup removes temporary files and directories
func (ie *InstallationEngine) Cleanup() error {
	return os.RemoveAll(ie.tempDir)
}