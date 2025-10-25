package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"boba/internal/config"
	"boba/internal/github"
	"boba/internal/installer"
	"boba/internal/parser"
)

// MenuType represents different menu screens
type MenuType int

const (
	MainMenu MenuType = iota
	InstallEverythingMenu
	UpdateEverythingMenu
	ToolsListMenu
	EnvironmentMenu
	ConfigurationMenu
	RepositoryConfigMenu
	ToolOverrideMenu
	EnvironmentOverrideMenu
	GitHubAuthMenu
)

// MenuModel represents the state of our menu system
type MenuModel struct {
	currentMenu      MenuType
	menuStack        []MenuType
	choices          []string
	cursor           int
	selected         map[int]struct{}
	configManager    *config.ConfigManager
	githubClient     *github.GitHubClient
	authModel        *github.AuthModel
	repoParser       *parser.RepositoryParser
	installEngine    *installer.InstallationEngine
	availableTools   []parser.Tool
	toolInstallStatus map[string]bool // Cache for tool installation status
	availableEnvironments []parser.Environment // Available environment configurations
	isLoading        bool
	loadingMessage   string
	installationInProgress bool
	installationResults    []InstallationResult
	installEverythingMode  bool // Flag to track if we're in "Install Everything" mode
	pendingEnvironments    []parser.Environment // Environments to apply after tools
	authError              string // Store authentication error for display
}

// MenuItem represents a menu option
type MenuItem struct {
	Title       string
	Description string
}

// InstallationResult represents the result of a tool installation
type InstallationResult struct {
	ToolName string
	Success  bool
	Message  string
	Error    error
}

// EnvironmentApplicationResult represents the result of environment application
type EnvironmentApplicationResult struct {
	EnvironmentName string
	Success         bool
	Message         string
	Error           error
}

// Message types for tea.Cmd communication
type ToolsListMsg struct {
	Tools []parser.Tool
}

type EnvironmentsListMsg struct {
	Environments []parser.Environment
}

type InstallationProgressMsg struct {
	ToolName string
	Status   string
	Success  bool
}

type InstallEverythingPhaseMsg struct {
	Phase        string // "tools" or "environments"
	Tools        []parser.Tool
	Environments []parser.Environment
}

type InstallationStartMsg struct {
	Tools        []parser.Tool
	CurrentIndex int
	Results      []InstallationResult
}

type InstallationNextMsg struct {
	Tools        []parser.Tool
	CurrentIndex int
	Results      []InstallationResult
}

type EnvironmentApplicationNextMsg struct {
	Environments []parser.Environment
	CurrentIndex int
	Results      []EnvironmentApplicationResult
}

type InstallationCompleteMsg struct {
	Results []InstallationResult
}

// Init is called when the program starts
func (m MenuModel) Init() tea.Cmd {
	return nil
}

// Getter methods for testing and external access
func (m MenuModel) GetCurrentMenu() MenuType {
	return m.currentMenu
}

func (m MenuModel) GetConfigManager() *config.ConfigManager {
	return m.configManager
}

func (m MenuModel) GetCursor() int {
	return m.cursor
}

func (m MenuModel) GetAuthError() string {
	return m.authError
}

func (m MenuModel) GetIsLoading() bool {
	return m.isLoading
}

func (m MenuModel) GetInstallationInProgress() bool {
	return m.installationInProgress
}

// Setter methods for testing
func (m MenuModel) SetCursor(cursor int) MenuModel {
	m.cursor = cursor
	return m
}

func (m MenuModel) SetLoading(loading bool, message string) MenuModel {
	m.isLoading = loading
	m.loadingMessage = message
	return m
}

func (m MenuModel) SetInstallationInProgress(inProgress bool) MenuModel {
	m.installationInProgress = inProgress
	return m
}

func (m MenuModel) SetAuthError(error string) MenuModel {
	m.authError = error
	return m
}

func (m MenuModel) NavigateBack() MenuModel {
	if len(m.menuStack) > 0 {
		m.currentMenu = m.menuStack[len(m.menuStack)-1]
		m.menuStack = m.menuStack[:len(m.menuStack)-1]
		m = m.updateMenuChoices()
	}
	return m
}

func (m MenuModel) HandleMenuSelection() (MenuModel, tea.Cmd) {
	model, cmd := m.handleMenuSelection()
	if menuModel, ok := model.(MenuModel); ok {
		return menuModel, cmd
	}
	return m, cmd
}

// updateMenuChoices updates the menu choices based on current menu
func (m MenuModel) updateMenuChoices() MenuModel {
	m.choices = m.getMenuChoices()
	m.cursor = 0
	return m
}