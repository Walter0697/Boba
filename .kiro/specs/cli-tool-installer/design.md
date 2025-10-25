# Design Document

## Overview

The CLI Tool Installer is a Go application that provides an interactive terminal interface for comprehensive development environment setup. The application uses a GitHub-driven configuration approach where all tool definitions and setup instructions are stored in the user's private repository, with local override capabilities for selective installation.

The application assumes zsh is already installed and configured as the default shell, simplifying the environment setup process by focusing on zsh configuration management.

**Project Name:** `boba`
*Configuration directory: `~/.boba/`*

The application follows a menu-driven architecture with four main components: GitHub authentication, interactive UI, configuration management, and cross-platform installation execution.

## Architecture

### High-Level Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Interactive   │    │     GitHub       │    │  Configuration  │
│   UI Layer      │◄──►│   Integration    │◄──►│   Manager       │
│  (Bubble Tea)   │    │    (go-github)   │    │   (Local JSON)  │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                        │                        │
         ▼                        ▼                        ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Installation  │    │   Repository     │    │   Platform      │
│   Orchestrator  │◄──►│   Parser         │◄──►│   Detector      │
│                 │    │                  │    │                 │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

### Core Components

1. **Main Application Controller**: Orchestrates the entire application flow
2. **Interactive UI Manager**: Handles menu navigation and user interactions
3. **GitHub Client**: Manages authentication and repository operations
4. **Configuration Manager**: Handles local overrides and settings persistence
5. **Installation Engine**: Executes platform-specific installation commands
6. **Repository Parser**: Parses configuration files from the GitHub repository

## Components and Interfaces

### 1. Main Application Structure

```go
type App struct {
    ui           *UIManager
    github       *GitHubClient
    config       *ConfigManager
    installer    *InstallationEngine
    repoParser   *RepositoryParser
}

type Config struct {
    RepositoryURL    string            `json:"repository_url"`
    ToolOverrides    map[string]bool   `json:"tool_overrides"`
    LastSync         time.Time         `json:"last_sync"`
}

// Credentials stored separately for security
type Credentials struct {
    GitHubToken      string            `json:"github_token"`
}
```

### 2. Interactive UI Manager

```go
type UIManager struct {
    model    tea.Model
    program  *tea.Program
}

type MenuModel struct {
    currentMenu  MenuType
    menuStack    []MenuType
    choices      []MenuItem
    cursor       int
    selected     map[int]struct{}
}

type MenuType int
const (
    MainMenu MenuType = iota
    ToolsMenu
    EnvironmentMenu
    ConfigurationMenu
    ToolCategoryMenu
    ToolOverrideMenu
)

type MenuItem struct {
    Title       string
    Description string
    Action      func() tea.Cmd
    Enabled     bool
}
```

### 3. GitHub Integration

```go
type GitHubClient struct {
    client *github.Client
    token  string
    owner  string
    repo   string
}

type RepositoryConfig struct {
    Tools       []ToolDefinition `yaml:"tools"`
    Environment EnvConfig        `yaml:"environment"`
    Version     string           `yaml:"version"`
}

type ToolDefinition struct {
    Name         string            `yaml:"name"`
    Description  string            `yaml:"description"`
    Category     string            `yaml:"category"`
    Installation map[string]string `yaml:"installation"` // platform -> command
    CheckCommand string            `yaml:"check_command"`
}

type EnvConfig struct {
    ZshrcPath    string `yaml:"zshrc_path"`
    BackupSuffix string `yaml:"backup_suffix"`
    // Note: Assumes zsh is already installed and set as default shell
}
```

### 4. Configuration Manager

```go
type ConfigManager struct {
    configDir     string  // ~/.{project_name}/
    configPath    string  // ~/.{project_name}/config.json
    credPath      string  // ~/.{project_name}/credentials.json
    config        *Config
    credentials   *Credentials
}

func (cm *ConfigManager) LoadConfig() error
func (cm *ConfigManager) SaveConfig() error
func (cm *ConfigManager) LoadCredentials() error
func (cm *ConfigManager) SaveCredentials() error
func (cm *ConfigManager) GetToolOverride(toolName string) (bool, bool)
func (cm *ConfigManager) SetToolOverride(toolName string, enabled bool) error
func (cm *ConfigManager) GetEnabledTools(allTools []ToolDefinition) []ToolDefinition
func (cm *ConfigManager) InitConfigDir() error
```

### 5. Installation Engine

```go
type InstallationEngine struct {
    platform Platform
}

type Platform struct {
    OS           string
    Distribution string
    PackageManager string
}

func (ie *InstallationEngine) DetectPlatform() Platform
func (ie *InstallationEngine) IsToolInstalled(tool ToolDefinition) bool
func (ie *InstallationEngine) InstallTool(tool ToolDefinition) error
func (ie *InstallationEngine) ExecuteCommand(command string) (string, error)
```

### 6. Repository Parser

```go
type RepositoryParser struct {
    github *GitHubClient
    cache  *RepositoryConfig
}

func (rp *RepositoryParser) FetchConfig() (*RepositoryConfig, error)
func (rp *RepositoryParser) GetToolsByCategory(category string) []ToolDefinition
func (rp *RepositoryParser) GetEnvironmentConfig() EnvConfig
```

## Data Models

### Repository Configuration Format (YAML)

```yaml
version: "1.0"
tools:
  - name: "git"
    description: "Version control system"
    category: "development"
    installation:
      linux: "sudo apt-get install -y git"
      darwin: "brew install git"
      windows: "winget install Git.Git"
    check_command: "git --version"
  
  - name: "docker"
    description: "Container platform"
    category: "development"
    installation:
      linux: "curl -fsSL https://get.docker.com | sh"
      darwin: "brew install --cask docker"
    check_command: "docker --version"

environment:
  zshrc_path: "dotfiles/.zshrc"
  backup_suffix: ".backup"
  # Assumes zsh is pre-installed and configured as default shell
```

### Local Configuration Format

**Config File (~/.{project_name}/config.json)**
```json
{
  "repository_url": "https://github.com/username/dev-setup",
  "tool_overrides": {
    "docker": false,
    "kubernetes": true
  },
  "last_sync": "2024-10-24T10:30:00Z"
}
```

**Credentials File (~/.{project_name}/credentials.json)**
```json
{
  "github_token": "ghp_xxxxxxxxxxxxxxxxxxxx"
}
```

**Directory Structure:**
```
~/.boba/
├── config.json
├── credentials.json
└── cache/
    └── repository_config.yaml
```

## Error Handling

### Error Types

```go
type AppError struct {
    Type    ErrorType
    Message string
    Cause   error
}

type ErrorType int
const (
    GitHubAuthError ErrorType = iota
    RepositoryAccessError
    InstallationError
    ConfigurationError
    PlatformNotSupportedError
)
```

### Error Handling Strategy

1. **GitHub Authentication Errors**: Prompt for re-authentication with clear error messages
2. **Repository Access Errors**: Provide specific feedback about repository permissions or network issues
3. **Installation Errors**: Display command output and suggest manual installation steps
4. **Configuration Errors**: Validate configuration files and provide helpful error messages
5. **Platform Errors**: Gracefully handle unsupported platforms with informative messages

### Recovery Mechanisms

- Retry logic for network operations with exponential backoff
- Graceful degradation when repository is unavailable (use cached config)
- Rollback capability for failed installations
- Configuration validation with helpful error messages

## Testing Strategy

### Unit Testing

1. **Configuration Manager Tests**
   - Test config loading/saving with various file states
   - Test tool override logic
   - Test configuration validation

2. **GitHub Client Tests**
   - Mock GitHub API responses
   - Test authentication flow
   - Test repository parsing

3. **Installation Engine Tests**
   - Mock command execution
   - Test platform detection
   - Test tool installation status checking

4. **Repository Parser Tests**
   - Test YAML parsing with various configurations
   - Test error handling for malformed configs
   - Test caching behavior

### Integration Testing

1. **End-to-End Menu Navigation**
   - Test complete user workflows
   - Test menu state transitions
   - Test error recovery flows

2. **GitHub Integration**
   - Test with real repository (using test repo)
   - Test authentication with various token states
   - Test repository configuration parsing

3. **Cross-Platform Testing**
   - Test platform detection on different OS
   - Test installation commands (in isolated environments)
   - Test zsh configuration management across platforms
   - Verify zsh prerequisite detection

### Manual Testing

1. **User Experience Testing**
   - Test menu navigation responsiveness
   - Test error message clarity
   - Test installation progress feedback

2. **Platform Compatibility**
   - Test on WSL, Linux, and macOS
   - Test with different package managers
   - Test with various shell configurations

## Implementation Phases

### Phase 1: Core Infrastructure
- Basic application structure and configuration management
- GitHub authentication and repository access
- Platform detection and basic command execution

### Phase 2: Interactive UI
- Bubble Tea integration for menu system
- Main menu and basic navigation
- Error display and user feedback

### Phase 3: Tool Management
- Repository configuration parsing
- Tool installation engine
- Installation status checking

### Phase 4: Advanced Features
- Local configuration overrides
- Environment setup (zshrc management)
- Installation progress and error handling

### Phase 5: Polish and Testing
- Comprehensive error handling
- Cross-platform testing
- User experience improvements