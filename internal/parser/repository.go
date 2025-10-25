package parser

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"boba/internal/github"
	"gopkg.in/yaml.v3"
)

// Tool represents a tool with its metadata and scripts
type Tool struct {
	Name         string   `yaml:"name" json:"name"`
	Description  string   `yaml:"description" json:"description"`
	Version      string   `yaml:"version,omitempty" json:"version,omitempty"`
	AutoInstall  bool     `yaml:"auto_install" json:"auto_install"`
	Dependencies []string `yaml:"dependencies,omitempty" json:"dependencies,omitempty"`
	Homepage     string   `yaml:"homepage,omitempty" json:"homepage,omitempty"`
	
	// Internal fields
	FolderName      string `yaml:"-" json:"-"`
	InstallScript   string `yaml:"-" json:"-"`
	UninstallScript string `yaml:"-" json:"-"`
}

// Environment represents an environment configuration with its metadata and scripts
type Environment struct {
	Name         string   `yaml:"name" json:"name"`
	Description  string   `yaml:"description" json:"description"`
	Shell        string   `yaml:"shell,omitempty" json:"shell,omitempty"` // zsh, bash, fish, etc.
	AutoApply    bool     `yaml:"auto_apply" json:"auto_apply"`
	Dependencies []string `yaml:"dependencies,omitempty" json:"dependencies,omitempty"`
	
	// Internal fields
	FolderName    string `yaml:"-" json:"-"`
	ConfigFiles   []string `yaml:"-" json:"-"` // List of config files (.zshrc, .bashrc, etc.)
	SetupScript   string `yaml:"-" json:"-"`
	RestoreScript string `yaml:"-" json:"-"`
}

// RepositoryContents represents the parsed repository structure
type RepositoryContents struct {
	Tools       []Tool    `json:"tools"`
	LastFetched time.Time `json:"last_fetched"`
}

// RepositoryParser handles parsing of repository configuration files
type RepositoryParser struct {
	github *github.GitHubClient
	cache  *RepositoryContents
}

// NewRepositoryParser creates a new repository parser instance
func NewRepositoryParser(githubClient *github.GitHubClient) *RepositoryParser {
	return &RepositoryParser{
		github: githubClient,
	}
}

// FetchTools fetches and parses all tools from the repository
func (rp *RepositoryParser) FetchTools() ([]Tool, error) {
	if rp.github == nil {
		return nil, fmt.Errorf("GitHub client not initialized")
	}

	// Get the tools directory listing
	toolNames, err := rp.github.GetDirectoryContents("tools")
	if err != nil {
		return nil, fmt.Errorf("cannot find 'tools' directory in your repository: %w", err)
	}

	var tools []Tool
	
	// Fetch each tool's configuration
	for _, toolName := range toolNames {
		tool, err := rp.fetchTool(toolName)
		if err != nil {
			// Tool doesn't exist or has issues, skip it but log the error
			fmt.Printf("Warning: Failed to fetch tool %s: %v\n", toolName, err)
			continue
		}
		tools = append(tools, tool)
	}

	// Cache the results
	rp.cache = &RepositoryContents{
		Tools:       tools,
		LastFetched: time.Now(),
	}

	return tools, nil
}

// fetchTool fetches a single tool's configuration
func (rp *RepositoryParser) fetchTool(toolName string) (Tool, error) {
	// Try to fetch tool.yaml first, then tool.json
	toolConfigPath := filepath.Join("tools", toolName, "tool.yaml")
	
	configContent, err := rp.github.GetRepositoryContents(toolConfigPath)
	if err != nil {
		// Try JSON format
		toolConfigPath = filepath.Join("tools", toolName, "tool.json")
		configContent, err = rp.github.GetRepositoryContents(toolConfigPath)
		if err != nil {
			return Tool{}, fmt.Errorf("failed to fetch tool config for %s: %w", toolName, err)
		}
	}

	// Parse the configuration
	var tool Tool
	if strings.HasSuffix(toolConfigPath, ".yaml") || strings.HasSuffix(toolConfigPath, ".yml") {
		err = yaml.Unmarshal(configContent, &tool)
	} else {
		err = json.Unmarshal(configContent, &tool)
	}
	
	if err != nil {
		return Tool{}, fmt.Errorf("failed to parse tool config for %s: %w", toolName, err)
	}

	// Set internal fields
	tool.FolderName = toolName
	tool.InstallScript = filepath.Join("tools", toolName, "install.sh")
	tool.UninstallScript = filepath.Join("tools", toolName, "uninstall.sh")

	return tool, nil
}

// GetTools returns cached tools or fetches them if not cached
func (rp *RepositoryParser) GetTools() ([]Tool, error) {
	if rp.cache != nil && time.Since(rp.cache.LastFetched) < 5*time.Minute {
		return rp.cache.Tools, nil
	}
	
	return rp.FetchTools()
}

// GetToolsByCategory returns tools filtered by category (deprecated - categories removed)
// This method is kept for backward compatibility but will return all tools
func (rp *RepositoryParser) GetToolsByCategory(category string) ([]Tool, error) {
	// Since categories are removed, just return all tools
	return rp.GetTools()
}

// GetToolByName returns a specific tool by name
func (rp *RepositoryParser) GetToolByName(name string) (*Tool, error) {
	tools, err := rp.GetTools()
	if err != nil {
		return nil, err
	}

	for _, tool := range tools {
		if tool.FolderName == name || tool.Name == name {
			return &tool, nil
		}
	}

	return nil, fmt.Errorf("tool %s not found", name)
}

// GetAutoInstallTools returns tools that should be installed automatically
func (rp *RepositoryParser) GetAutoInstallTools() ([]Tool, error) {
	tools, err := rp.GetTools()
	if err != nil {
		return nil, err
	}

	var autoInstallTools []Tool
	for _, tool := range tools {
		if tool.AutoInstall {
			autoInstallTools = append(autoInstallTools, tool)
		}
	}

	return autoInstallTools, nil
}

// GetManualInstallTools returns tools that should only be installed manually
func (rp *RepositoryParser) GetManualInstallTools() ([]Tool, error) {
	tools, err := rp.GetTools()
	if err != nil {
		return nil, err
	}

	var manualInstallTools []Tool
	for _, tool := range tools {
		if !tool.AutoInstall {
			manualInstallTools = append(manualInstallTools, tool)
		}
	}

	return manualInstallTools, nil
}

// FetchEnvironments fetches and parses all environment configurations from the repository
func (rp *RepositoryParser) FetchEnvironments() ([]Environment, error) {
	if rp.github == nil {
		return nil, fmt.Errorf("GitHub client not initialized")
	}

	// Get the environments directory listing
	envNames, err := rp.github.GetDirectoryContents("environments")
	if err != nil {
		return nil, fmt.Errorf("cannot find 'environments' directory in your repository: %w", err)
	}

	var environments []Environment
	
	// Fetch each environment's configuration
	for _, envName := range envNames {
		env, err := rp.fetchEnvironment(envName)
		if err != nil {
			// Environment doesn't exist or has issues, skip it but log the error
			fmt.Printf("Warning: Failed to fetch environment %s: %v\n", envName, err)
			continue
		}
		environments = append(environments, env)
	}

	return environments, nil
}

// fetchEnvironment fetches a single environment's configuration
func (rp *RepositoryParser) fetchEnvironment(envName string) (Environment, error) {
	// Try to fetch environment.yaml first, then environment.json
	envConfigPath := filepath.Join("environments", envName, "environment.yaml")
	
	configContent, err := rp.github.GetRepositoryContents(envConfigPath)
	if err != nil {
		// Try JSON format
		envConfigPath = filepath.Join("environments", envName, "environment.json")
		configContent, err = rp.github.GetRepositoryContents(envConfigPath)
		if err != nil {
			return Environment{}, fmt.Errorf("failed to fetch environment config for %s: %w", envName, err)
		}
	}

	// Parse the configuration
	var env Environment
	if strings.HasSuffix(envConfigPath, ".yaml") || strings.HasSuffix(envConfigPath, ".yml") {
		err = yaml.Unmarshal(configContent, &env)
	} else {
		err = json.Unmarshal(configContent, &env)
	}
	
	if err != nil {
		return Environment{}, fmt.Errorf("failed to parse environment config for %s: %w", envName, err)
	}

	// Set internal fields
	env.FolderName = envName
	env.SetupScript = filepath.Join("environments", envName, "setup.sh")
	env.RestoreScript = filepath.Join("environments", envName, "restore.sh")
	
	// Detect config files (common shell config files)
	configFiles := []string{".zshrc", ".bashrc", ".profile", ".bash_profile", ".fishrc"}
	for _, configFile := range configFiles {
		configPath := filepath.Join("environments", envName, configFile)
		if _, err := rp.github.GetRepositoryContents(configPath); err == nil {
			env.ConfigFiles = append(env.ConfigFiles, configPath)
		}
	}

	return env, nil
}

// GetEnvironmentByName returns a specific environment by name
func (rp *RepositoryParser) GetEnvironmentByName(name string) (*Environment, error) {
	environments, err := rp.FetchEnvironments()
	if err != nil {
		return nil, err
	}

	for _, env := range environments {
		if env.FolderName == name || env.Name == name {
			return &env, nil
		}
	}

	return nil, fmt.Errorf("environment %s not found", name)
}

// GetAutoApplyEnvironments returns environments that should be applied automatically
func (rp *RepositoryParser) GetAutoApplyEnvironments() ([]Environment, error) {
	environments, err := rp.FetchEnvironments()
	if err != nil {
		return nil, err
	}

	var autoApplyEnvironments []Environment
	for _, env := range environments {
		if env.AutoApply {
			autoApplyEnvironments = append(autoApplyEnvironments, env)
		}
	}

	return autoApplyEnvironments, nil
}