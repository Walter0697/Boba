package installer

import "boba/internal/parser"

// GitHubClientInterface defines the interface for GitHub operations needed by the installer
type GitHubClientInterface interface {
	GetRepositoryContents(path string) ([]byte, error)
}

// InstallationEngineInterface defines the interface for installation operations
type InstallationEngineInterface interface {
	// Tool operations
	IsToolInstalled(tool parser.Tool) bool
	InstallTool(tool parser.Tool) (*InstallationResult, error)
	UninstallTool(tool parser.Tool) (*InstallationResult, error)
	VerifyInstallation(tool parser.Tool) (bool, string)
	
	// Environment operations
	IsEnvironmentApplied(env parser.Environment) bool
	ApplyEnvironment(env parser.Environment) (*InstallationResult, error)
	RestoreEnvironment(env parser.Environment) (*InstallationResult, error)
	VerifyEnvironmentApplication(env parser.Environment) (bool, string)
	
	// Utility operations
	ExecuteCommand(command string) (string, error)
	GetPlatform() Platform
	Cleanup() error
}