# Installation Engine

The Installation Engine is a comprehensive script execution system for installing and managing CLI tools from GitHub repositories. It provides secure, cross-platform script execution with proper error handling, output capture, and installation verification.

## Features

### Core Functionality
- **Script Execution**: Executes install.sh and uninstall.sh scripts from GitHub repositories
- **Cross-Platform Support**: Works on Linux, macOS, and Windows
- **Security Measures**: Secure script execution with proper permissions and timeouts
- **Output Capture**: Captures both stdout and stderr from script execution
- **Error Handling**: Comprehensive error handling with detailed error messages
- **Installation Verification**: Verifies tool installation by checking PATH availability

### Platform Detection
- **Linux**: Detects distribution (Ubuntu, CentOS, Arch, etc.) and package manager (apt, yum, pacman, etc.)
- **macOS**: Automatically configures for Homebrew
- **Windows**: Supports PowerShell script execution
- **WSL**: Properly detects WSL environment and uses appropriate Linux package managers

### Security Features
- **Timeout Protection**: 10-minute timeout for installations, 30-second timeout for commands
- **Environment Isolation**: Scripts run with controlled environment variables
- **Temporary File Management**: Secure temporary file handling with automatic cleanup
- **Permission Control**: Proper file permissions for script execution

## Usage

### Basic Installation

```go
import (
    "boba/internal/github"
    "boba/internal/installer"
    "boba/internal/parser"
)

// Create GitHub client
githubClient := github.NewGitHubClient("token", "username", "repo")

// Create installation engine
engine := installer.NewInstallationEngine(githubClient)
defer engine.Cleanup() // Always cleanup temporary files

// Create tool definition
tool := parser.Tool{
    Name:          "git",
    FolderName:    "git",
    InstallScript: "tools/git/install.sh",
}

// Install the tool
result, err := engine.InstallTool(tool)
if err != nil {
    log.Printf("Installation failed: %v", err)
    return
}

if result.Success {
    fmt.Printf("Installation successful! Duration: %v\n", result.Duration)
} else {
    fmt.Printf("Installation failed with exit code %d\n", result.ExitCode)
}
```

### Installation Verification

```go
// Check if tool is already installed
if engine.IsToolInstalled(tool) {
    fmt.Printf("Tool %s is already installed\n", tool.Name)
}

// Verify installation after installing
isVerified, message := engine.VerifyInstallation(tool)
if isVerified {
    fmt.Printf("✓ %s\n", message)
} else {
    fmt.Printf("✗ %s\n", message)
}
```

### Uninstallation

```go
tool := parser.Tool{
    Name:            "git",
    FolderName:      "git",
    UninstallScript: "tools/git/uninstall.sh",
}

result, err := engine.UninstallTool(tool)
if err != nil {
    log.Printf("Uninstallation failed: %v", err)
}
```

## Script Requirements

### Install Script (install.sh)
Scripts should be executable shell scripts that:
- Exit with code 0 on success, non-zero on failure
- Provide meaningful output for user feedback
- Handle platform-specific installation logic
- Use environment variables provided by the engine

### Environment Variables
The engine provides these environment variables to scripts:
- `BOBA_TOOL_NAME`: Name of the tool being installed
- `BOBA_PLATFORM`: Target platform (linux, darwin, windows)
- `BOBA_PACKAGE_MANAGER`: Detected package manager (apt, brew, etc.)
- `BOBA_TEMP_DIR`: Temporary directory for downloads and intermediate files
- `TMPDIR`, `TEMP`, `TMP`: Standard temp directory variables (all set to BOBA's temp dir)

**Note:** Scripts are executed with their working directory set to `$BOBA_TEMP_DIR`, so you can use relative paths for temporary files. All temporary files should be created in this directory to ensure proper cleanup.

### Example Install Script
```bash
#!/bin/bash
set -e

echo "Installing $BOBA_TOOL_NAME for platform $BOBA_PLATFORM"

case "$BOBA_PLATFORM" in
    "linux")
        case "$BOBA_PACKAGE_MANAGER" in
            "apt")
                sudo apt-get update
                sudo apt-get install -y git
                ;;
            "yum")
                sudo yum install -y git
                ;;
            *)
                echo "Unsupported package manager: $BOBA_PACKAGE_MANAGER"
                exit 1
                ;;
        esac
        ;;
    "darwin")
        brew install git
        ;;
    *)
        echo "Unsupported platform: $BOBA_PLATFORM"
        exit 1
        ;;
esac

echo "Installation of $BOBA_TOOL_NAME completed successfully"
```

## API Reference

### Types

#### InstallationEngine
Main engine for handling tool installations.

```go
type InstallationEngine struct {
    platform     Platform
    githubClient GitHubClientInterface
    tempDir      string
}
```

#### InstallationResult
Result of an installation or uninstallation operation.

```go
type InstallationResult struct {
    Success    bool
    Output     string
    Error      error
    ExitCode   int
    Duration   time.Duration
}
```

#### Platform
Information about the target platform.

```go
type Platform struct {
    OS             string
    Distribution   string
    PackageManager string
}
```

### Methods

#### NewInstallationEngine(githubClient GitHubClientInterface) *InstallationEngine
Creates a new installation engine instance.

#### IsToolInstalled(tool parser.Tool) bool
Checks if a tool is already installed on the system by looking for it in PATH.

#### InstallTool(tool parser.Tool) (*InstallationResult, error)
Installs a tool using its install script from the repository.

#### UninstallTool(tool parser.Tool) (*InstallationResult, error)
Uninstalls a tool using its uninstall script from the repository.

#### VerifyInstallation(tool parser.Tool) (bool, string)
Verifies that a tool was successfully installed and is accessible.

#### ExecuteCommand(command string) (string, error)
Executes a shell command and returns the output.

#### GetPlatform() Platform
Returns the detected platform information.

#### Cleanup() error
Removes temporary files and directories.

## Error Handling

The engine provides comprehensive error handling:

- **Script Download Errors**: When install/uninstall scripts cannot be downloaded
- **Script Execution Errors**: When scripts fail to execute or return non-zero exit codes
- **Timeout Errors**: When scripts take longer than the allowed timeout
- **Platform Errors**: When the platform cannot be detected or is unsupported
- **Permission Errors**: When script files cannot be created or executed

All errors include detailed context and suggestions for resolution.

## Testing

The engine includes comprehensive tests:

- **Unit Tests**: Test individual components and methods
- **Integration Tests**: Test realistic installation scenarios
- **Mock Testing**: Test with mock GitHub clients for isolated testing
- **Platform Testing**: Test platform detection and command execution

Run tests with:
```bash
go test ./internal/installer -v
```

Run integration tests:
```bash
go test ./internal/installer -v -run TestInstallationEngineIntegration
```

## Security Considerations

- Scripts are executed in isolated temporary directories
- Timeouts prevent long-running or hanging scripts
- Environment variables are controlled and limited
- Temporary files are automatically cleaned up
- Scripts cannot access sensitive system information beyond what's explicitly provided

## Future Enhancements

- Support for custom script interpreters (Python, Node.js, etc.)
- Dependency resolution and installation ordering
- Rollback capability for failed installations
- Progress reporting for long-running installations
- Parallel installation support
- Custom timeout configuration per tool