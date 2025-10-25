package github

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AuthState represents the current state of the authentication process
type AuthState int

const (
	AuthStateTokenInput AuthState = iota
	AuthStateRepoInput
	AuthStateValidating
	AuthStateSuccess
	AuthStateError
)

// AuthModel represents the authentication UI model
type AuthModel struct {
	state           AuthState
	tokenInput      string
	repoInput       string
	errorMessage    string
	successMessage  string
	cursor          int
	client          *GitHubClient
	onComplete      func(client *GitHubClient, repoURL string) tea.Cmd
	onCancel        func() tea.Cmd
}

// AuthMsg represents messages for the authentication flow
type AuthMsg struct {
	Type     string
	Success  bool
	Error    error
	User     string
	RepoName string
	CloneDir string
}

// NewAuthModel creates a new authentication model
func NewAuthModel(onComplete func(*GitHubClient, string) tea.Cmd, onCancel func() tea.Cmd) *AuthModel {
	return &AuthModel{
		state:      AuthStateTokenInput,
		repoInput:  "boba-config", // Default repository name
		onComplete: onComplete,
		onCancel:   onCancel,
	}
}

// NewAuthModelWithRepo creates a new authentication model with a custom default repository
func NewAuthModelWithRepo(defaultRepo string, onComplete func(*GitHubClient, string) tea.Cmd, onCancel func() tea.Cmd) *AuthModel {
	return &AuthModel{
		state:      AuthStateTokenInput,
		repoInput:  defaultRepo,
		onComplete: onComplete,
		onCancel:   onCancel,
	}
}

// NewRepoConfigModel creates a new model for repository configuration only (skips token input)
func NewRepoConfigModel(defaultRepo string, onComplete func(*GitHubClient, string) tea.Cmd, onCancel func() tea.Cmd) *AuthModel {
	return &AuthModel{
		state:      AuthStateRepoInput, // Start directly at repository input
		repoInput:  defaultRepo,
		onComplete: onComplete,
		onCancel:   onCancel,
	}
}

// Init initializes the authentication model
func (m *AuthModel) Init() tea.Cmd {
	return nil
}

// Update handles authentication model updates
func (m *AuthModel) Update(msg tea.Msg) (*AuthModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case AuthStateTokenInput:
			return m.handleTokenInput(msg)
		case AuthStateRepoInput:
			return m.handleRepoInput(msg)
		case AuthStateError:
			return m.handleErrorState(msg)
		case AuthStateSuccess:
			return m.handleSuccessState(msg)
		}
	case AuthMsg:
		return m.handleAuthMsg(msg)
	}
	return m, nil
}

// handleTokenInput handles token input state
func (m *AuthModel) handleTokenInput(msg tea.KeyMsg) (*AuthModel, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		// Ctrl+C should quit the entire application
		return m, tea.Quit
	case "esc":
		// Esc should go back to main menu
		if m.onCancel != nil {
			return m, m.onCancel()
		}
		return m, tea.Quit
	case "enter":
		if strings.TrimSpace(m.tokenInput) == "" {
			m.errorMessage = "GitHub token cannot be empty"
			return m, nil
		}
		// Skip repository input and go directly to validation
		return m, m.validateCredentials()
	case "backspace":
		if len(m.tokenInput) > 0 {
			m.tokenInput = m.tokenInput[:len(m.tokenInput)-1]
		}
		return m, nil
	case "ctrl+a":
		// Select all - clear the input for easy replacement
		m.tokenInput = ""
		return m, nil
	case "ctrl+u":
		// Clear line (common terminal shortcut)
		m.tokenInput = ""
		return m, nil
	default:
		// Handle different key types
		switch msg.Type {
		case tea.KeyRunes:
			// This handles pasted content (multiple characters at once)
			pastedText := string(msg.Runes)
			// Filter to only allow valid token characters (alphanumeric and common token chars)
			filteredText := ""
			for _, r := range pastedText {
				if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
					filteredText += string(r)
				}
			}
			m.tokenInput += filteredText
		default:
			// Add single character to token input
			char := msg.String()
			if len(char) == 1 {
				r := rune(char[0])
				// Only allow valid token characters
				if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
					m.tokenInput += char
				}
			}
		}
		return m, nil
	}
}

// handleRepoInput handles repository URL input state
func (m *AuthModel) handleRepoInput(msg tea.KeyMsg) (*AuthModel, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		// Ctrl+C should quit the entire application
		return m, tea.Quit
	case "esc":
		// Esc should go back to main menu
		if m.onCancel != nil {
			return m, m.onCancel()
		}
		return m, tea.Quit
	case "enter":
		if strings.TrimSpace(m.repoInput) == "" {
			m.errorMessage = "Repository URL cannot be empty"
			return m, nil
		}
		return m, m.validateCredentials()
	case "backspace":
		if len(m.repoInput) > 0 {
			m.repoInput = m.repoInput[:len(m.repoInput)-1]
		}
		return m, nil
	case "ctrl+a":
		// Select all - clear the input for easy replacement
		m.repoInput = ""
		return m, nil
	case "ctrl+u":
		// Clear line (common terminal shortcut)
		m.repoInput = ""
		return m, nil
	default:
		// Handle different key types
		switch msg.Type {
		case tea.KeyRunes:
			// This handles pasted content (multiple characters at once)
			pastedText := string(msg.Runes)
			// Allow all printable characters for URLs
			filteredText := ""
			for _, r := range pastedText {
				if r >= 32 && r <= 126 { // Printable ASCII characters
					filteredText += string(r)
				}
			}
			m.repoInput += filteredText
		default:
			// Add single character to repo input
			char := msg.String()
			if len(char) == 1 && char[0] >= 32 && char[0] <= 126 {
				m.repoInput += char
			}
		}
		return m, nil
	}
}

// handleErrorState handles error display state
func (m *AuthModel) handleErrorState(msg tea.KeyMsg) (*AuthModel, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		// Ctrl+C should quit the entire application
		return m, tea.Quit
	case "esc":
		// Esc should go back to main menu
		if m.onCancel != nil {
			return m, m.onCancel()
		}
		return m, tea.Quit
	case "r":
		// Retry - go back to token input
		m.state = AuthStateTokenInput
		m.errorMessage = ""
		m.tokenInput = ""
		m.repoInput = ""
		return m, nil
	case "enter":
		// Continue with error - go back to main menu
		if m.onCancel != nil {
			return m, m.onCancel()
		}
		return m, nil
	}
	return m, nil
}

// handleSuccessState handles success display state
func (m *AuthModel) handleSuccessState(msg tea.KeyMsg) (*AuthModel, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "esc", "enter":
		if m.onComplete != nil && m.client != nil {
			return m, m.onComplete(m.client, m.repoInput)
		}
		return m, nil
	}
	return m, nil
}

// handleAuthMsg handles authentication result messages
func (m *AuthModel) handleAuthMsg(msg AuthMsg) (*AuthModel, tea.Cmd) {
	switch msg.Type {
	case "validation_complete":
		if msg.Success {
			m.state = AuthStateSuccess
			if msg.RepoName != "" && msg.CloneDir != "" {
				m.successMessage = fmt.Sprintf("✅ Successfully authenticated as %s\n🔄 Repository '%s' cloned to:\n   %s", msg.User, msg.RepoName, msg.CloneDir)
			} else {
				m.successMessage = fmt.Sprintf("✅ Successfully authenticated as %s", msg.User)
			}
		} else {
			m.state = AuthStateError
			m.errorMessage = fmt.Sprintf("❌ Authentication failed: %s", msg.Error.Error())
		}
	}
	return m, nil
}

// validateCredentials validates the GitHub token and repository access
func (m *AuthModel) validateCredentials() tea.Cmd {
	m.state = AuthStateValidating
	m.errorMessage = ""
	
	return func() tea.Msg {
		// First, create a client to get the authenticated user
		tempClient := NewGitHubClient(m.tokenInput, "", "")
		
		// Validate token and get user info
		authResult, err := tempClient.ValidateToken()
		if err != nil {
			return AuthMsg{
				Type:    "validation_complete",
				Success: false,
				Error:   fmt.Errorf("token validation failed: %w", err),
			}
		}

		if !authResult.Success {
			return AuthMsg{
				Type:    "validation_complete",
				Success: false,
				Error:   authResult.Error,
			}
		}

		// Get the username from the authenticated user
		var owner string
		if authResult.User != nil && authResult.User.Login != nil {
			owner = *authResult.User.Login
		} else {
			return AuthMsg{
				Type:    "validation_complete",
				Success: false,
				Error:   fmt.Errorf("could not get GitHub username from token"),
			}
		}

		// Handle repository name - if it's just a name, prepend the username
		var repoURL string
		if strings.Contains(m.repoInput, "/") {
			// Already has owner/repo format
			repoURL = m.repoInput
		} else {
			// Just a repo name, prepend the username
			repoURL = fmt.Sprintf("%s/%s", owner, m.repoInput)
		}

		// Parse the full repository URL
		repoOwner, repo, err := ParseRepositoryURL(repoURL)
		if err != nil {
			return AuthMsg{
				Type:    "validation_complete",
				Success: false,
				Error:   fmt.Errorf("invalid repository URL '%s' (from input '%s'): %w", repoURL, m.repoInput, err),
			}
		}

		// Create GitHub client with proper owner/repo
		client := NewGitHubClient(m.tokenInput, repoOwner, repo)

		// Validate repository access
		if err := client.ValidateRepositoryAccess(); err != nil {
			return AuthMsg{
				Type:    "validation_complete",
				Success: false,
				Error:   fmt.Errorf("repository access failed for '%s/%s': %w", repoOwner, repo, err),
			}
		}

		// Clone the repository
		targetDir, err := client.GetCloneTargetDir()
		if err != nil {
			return AuthMsg{
				Type:    "validation_complete",
				Success: false,
				Error:   fmt.Errorf("failed to determine clone directory: %w", err),
			}
		}

		if err := client.CloneRepository(targetDir); err != nil {
			return AuthMsg{
				Type:    "validation_complete",
				Success: false,
				Error:   fmt.Errorf("failed to clone repository '%s': %w", client.GetFullRepoName(), err),
			}
		}

		// Store the client for later use
		m.client = client

		userName := "Unknown"
		if authResult.User != nil && authResult.User.Login != nil {
			userName = *authResult.User.Login
		}

		return AuthMsg{
			Type:    "validation_complete",
			Success: true,
			User:    userName,
			RepoName: client.GetFullRepoName(),
			CloneDir: targetDir,
		}
	}
}

// View renders the authentication UI
func (m *AuthModel) View() string {
	var s strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FF6B9D")).
		MarginBottom(1)

	s.WriteString(headerStyle.Render("🔐 GitHub Authentication") + "\n\n")

	// Content based on state
	switch m.state {
	case AuthStateTokenInput:
		s.WriteString("Please enter your GitHub Personal Access Token:\n\n")
		s.WriteString("Token: ")
		// Show asterisks for security
		if len(m.tokenInput) > 0 {
			s.WriteString(strings.Repeat("*", len(m.tokenInput)))
		}
		s.WriteString("█\n\n") // Cursor
		
		if m.errorMessage != "" {
			errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
			s.WriteString(errorStyle.Render(m.errorMessage) + "\n\n")
		}
		
		s.WriteString("💡 You can create a token at: https://github.com/settings/tokens\n")
		s.WriteString("   Required scopes: repo (for private repositories)\n\n")
		s.WriteString("📋 Tip: You can paste your token with Ctrl+V\n")
		s.WriteString("🔧 Shortcuts: Ctrl+A (clear), Ctrl+U (clear line)\n\n")
		
		// Show which repository will be used
		repoHintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#87CEEB")).Italic(true)
		if strings.Contains(m.repoInput, "/") {
			s.WriteString(repoHintStyle.Render(fmt.Sprintf("Will use repository: %s", m.repoInput)) + "\n\n")
		} else {
			s.WriteString(repoHintStyle.Render(fmt.Sprintf("Will use repository: <your-username>/%s", m.repoInput)) + "\n\n")
		}
		
		s.WriteString("Press Enter to continue, Esc to go back, Ctrl+C to quit")

	case AuthStateRepoInput:
		s.WriteString("Please enter your repository URL:\n\n")
		s.WriteString("Repository: " + m.repoInput + "█\n")
		
		// Show default repository hint
		if m.repoInput == "boba-config" {
			hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#87CEEB")).Italic(true)
			s.WriteString(hintStyle.Render("(using default repository name)") + "\n")
		}
		s.WriteString("\n")
		
		if m.errorMessage != "" {
			errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
			s.WriteString(errorStyle.Render(m.errorMessage) + "\n\n")
		}
		
		s.WriteString("💡 Examples:\n")
		s.WriteString("   • https://github.com/username/repo\n")
		s.WriteString("   • username/repo\n")
		s.WriteString("   • git@github.com:username/repo.git\n\n")
		s.WriteString("📋 Tip: You can paste your URL with Ctrl+V\n")
		s.WriteString("🔧 Shortcuts: Ctrl+A (clear), Ctrl+U (clear line)\n\n")
		s.WriteString("Press Enter to continue, Esc to go back, Ctrl+C to quit")

	case AuthStateValidating:
		s.WriteString("🔄 Validating credentials and cloning repository...\n\n")
		s.WriteString("Please wait while we:\n")
		s.WriteString("  • Verify your GitHub token\n")
		s.WriteString("  • Check repository access\n")
		s.WriteString("  • Clone the repository locally")

	case AuthStateSuccess:
		successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
		s.WriteString(successStyle.Render(m.successMessage) + "\n\n")
		s.WriteString("Press Enter to continue")

	case AuthStateError:
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
		s.WriteString(errorStyle.Render(m.errorMessage) + "\n\n")
		s.WriteString("Press 'r' to retry, Enter/Esc to go back, or Ctrl+C to quit")
	}

	return s.String()
}

// GetToken returns the entered token
func (m *AuthModel) GetToken() string {
	return m.tokenInput
}

// GetRepoURL returns the entered repository URL
func (m *AuthModel) GetRepoURL() string {
	return m.repoInput
}

// GetClient returns the authenticated GitHub client
func (m *AuthModel) GetClient() *GitHubClient {
	return m.client
}