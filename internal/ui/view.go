package ui

import (
	"fmt"
	"strings"
	
	"github.com/charmbracelet/lipgloss"
	"github.com/common-nighthawk/go-figure"
)

// Enhanced styling constants and styles
var (
	// Color palette
	primaryColor   = lipgloss.Color("205") // Pink/magenta
	secondaryColor = lipgloss.Color("86")  // Green
	accentColor    = lipgloss.Color("212") // Light pink
	mutedColor     = lipgloss.Color("241") // Gray
	errorColor     = lipgloss.Color("196") // Red
	successColor   = lipgloss.Color("46")  // Bright green
	warningColor   = lipgloss.Color("226") // Yellow
	
	// Base styles
	baseStyle = lipgloss.NewStyle().
		Padding(0, 1).
		Margin(0, 0)
	
	// Header styles
	headerStyle = lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true).
		Align(lipgloss.Center).
		Margin(0, 0)
	
	titleStyle = lipgloss.NewStyle().
		Foreground(secondaryColor).
		Bold(true).
		Align(lipgloss.Center).
		Padding(0, 0).
		Margin(0, 0)
	
	// Menu item styles
	menuItemStyle = lipgloss.NewStyle().
		Padding(0, 1).
		Margin(0, 0)
	
	selectedMenuItemStyle = lipgloss.NewStyle().
		Foreground(accentColor).
		Bold(true).
		Padding(0, 1).
		Margin(0, 0)
	
	// Help text style
	helpStyle = lipgloss.NewStyle().
		Foreground(mutedColor).
		Italic(true).
		Align(lipgloss.Center).
		Margin(1, 0).
		Padding(0, 0)
	
	// Status styles
	loadingStyle = lipgloss.NewStyle().
		Foreground(warningColor).
		Bold(true).
		Align(lipgloss.Center).
		Padding(1, 2).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(warningColor)
	
	errorStyle = lipgloss.NewStyle().
		Foreground(errorColor).
		Bold(true).
		Padding(0, 0).
		Margin(0, 0)
	
	successStyle = lipgloss.NewStyle().
		Foreground(successColor).
		Bold(true).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(successColor).
		Margin(1, 0)
)

// getMenuTitle returns the title for the current menu
func (m MenuModel) getMenuTitle() string {
	switch m.currentMenu {
	case MainMenu:
		return "Select an option:"
	case InstallEverythingMenu:
		return "🚀 Install Everything"
	case ToolsListMenu:
		return "📋 List of Available Tools"
	case EnvironmentMenu:
		return "🌍 Setup Environment"
	case ConfigurationMenu:
		return "⚙️ Installation Configuration"
	case RepositoryConfigMenu:
		return "📁 Repository Configuration"
	case ToolOverrideMenu:
		return "🔧 Tool Override Management"
	case EnvironmentOverrideMenu:
		return "🌍 Environment Override Management"
	case GitHubAuthMenu:
		return "🔐 GitHub Authentication"
	default:
		return "Menu"
	}
}

// getHelpText returns context-appropriate help text
func (m MenuModel) getHelpText() string {
	helpText := ""
	if m.currentMenu == MainMenu {
		helpText = "Navigate: ↑/↓ or j/k • Select: Enter/Space • Quit: q or Ctrl+C"
	} else {
		helpText = "Navigate: ↑/↓ or j/k • Select: Enter/Space • Back: esc/b • Quit: q or Ctrl+C"
	}
	
	// Add authentication status if relevant
	if m.requiresAuthentication() {
		helpText += "\n🔐 GitHub authentication required for this feature"
	}
	
	return helpText
}

// View renders the UI with enhanced styling
func (m MenuModel) View() string {
	// Handle authentication screen
	if m.currentMenu == GitHubAuthMenu && m.authModel != nil {
		return m.renderAuthScreen()
	}
	
	// Handle loading states
	if m.isLoading {
		return m.renderLoadingScreen()
	}
	
	// Handle installation progress
	if m.installationInProgress {
		return m.renderInstallationScreen()
	}
	
	// Handle showing results
	if m.showingResults {
		return m.renderResultsScreen()
	}
	
	var s strings.Builder
	
	// ASCII Art Header
	s.WriteString(m.renderHeader())
	s.WriteString("\n")
	
	// Menu title
	s.WriteString(titleStyle.Render(m.getMenuTitle()))
	s.WriteString("\n")
	
	// Menu items
	s.WriteString(m.renderMenuItems())
	s.WriteString("\n")
	
	// Help text
	s.WriteString(helpStyle.Render(m.getHelpText()))
	
	// Show authentication error as simple line
	if m.authError != "" {
		s.WriteString("\n")
		s.WriteString(errorStyle.Render(m.authError))
	}
	
	return baseStyle.Render(s.String())
}

// renderHeader creates the ASCII art header with enhanced styling
func (m MenuModel) renderHeader() string {
	myFigure := figure.NewFigure("BOBA", "", true)
	header := myFigure.String()
	
	// Add subtitle
	subtitle := "Development Environment Setup Tool"
	
	headerContent := header + "\n" + 
		lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true).
			Align(lipgloss.Center).
			Render(subtitle)
	
	return headerStyle.Render(headerContent)
}

// renderMenuItems creates simple menu items without borders
func (m MenuModel) renderMenuItems() string {
	var items []string
	currentChoices := m.getMenuChoices()
	
	for i, choice := range currentChoices {
		if m.cursor == i {
			// Selected item
			items = append(items, selectedMenuItemStyle.Render(fmt.Sprintf("▶ %s", choice)))
		} else {
			// Regular item
			items = append(items, menuItemStyle.Render(fmt.Sprintf("  %s", choice)))
		}
	}
	
	return strings.Join(items, "\n")
}

// renderLoadingScreen shows loading state with spinner animation
func (m MenuModel) renderLoadingScreen() string {
	var s strings.Builder
	
	s.WriteString(m.renderHeader())
	s.WriteString("\n")
	
	// Loading message with spinner
	loadingMsg := "🔄 " + m.loadingMessage
	if m.loadingMessage == "" {
		loadingMsg = "🔄 Loading..."
	}
	
	s.WriteString(loadingStyle.Render(loadingMsg))
	s.WriteString("\n")
	
	// Loading help text
	loadingHelp := "Please wait while we fetch the latest configuration..."
	s.WriteString(helpStyle.Render(loadingHelp))
	
	return baseStyle.Render(s.String())
}

// renderInstallationScreen shows installation progress
func (m MenuModel) renderInstallationScreen() string {
	var s strings.Builder
	
	s.WriteString(m.renderHeader())
	s.WriteString("\n")
	
	// Installation title
	installTitle := "🚀 Installation in Progress"
	s.WriteString(titleStyle.Render(installTitle))
	s.WriteString("\n\n")
	
	// Show installation results
	if len(m.installationResults) > 0 {
		for _, result := range m.installationResults {
			var resultStyle lipgloss.Style
			var icon string
			
			if result.Success {
				resultStyle = successStyle
				icon = "✅"
			} else {
				resultStyle = errorStyle
				icon = "❌"
			}
			
			resultText := fmt.Sprintf("%s %s: %s", icon, result.ToolName, result.Message)
			s.WriteString(resultStyle.Render(resultText))
			s.WriteString("\n")
		}
	}
	
	// Installation help text
	installHelp := "Installation is running... Press 'q' to cancel (may leave partial installations)"
	s.WriteString(helpStyle.Render(installHelp))
	
	return baseStyle.Render(s.String())
}

// renderResultsScreen shows installation/environment results with option to return to menu
func (m MenuModel) renderResultsScreen() string {
	var s strings.Builder
	
	// ASCII Art Header
	s.WriteString(m.renderHeader())
	s.WriteString("\n")
	
	// Results title
	resultsTitle := "📋 Operation Results"
	s.WriteString(titleStyle.Render(resultsTitle))
	s.WriteString("\n\n")
	
	// Show results
	if len(m.installationResults) > 0 {
		for _, result := range m.installationResults {
			var resultStyle lipgloss.Style
			var icon string
			
			if result.Success {
				resultStyle = successStyle
				icon = "✅"
			} else {
				resultStyle = errorStyle
				icon = "❌"
			}
			
			resultText := fmt.Sprintf("%s %s", icon, result.ToolName)
			s.WriteString(resultStyle.Render(resultText))
			s.WriteString("\n")
			
			// Show detailed message
			if result.Message != "" {
				messageLines := strings.Split(result.Message, "\n")
				for _, line := range messageLines {
					if strings.TrimSpace(line) != "" {
						s.WriteString(fmt.Sprintf("   %s\n", line))
					}
				}
			}
			s.WriteString("\n")
		}
	}
	
	// Instructions
	instructionText := "Press any key to return to the main menu"
	s.WriteString(helpStyle.Render(instructionText))
	
	return baseStyle.Render(s.String())
}

// renderAuthScreen shows authentication screen with enhanced styling
func (m MenuModel) renderAuthScreen() string {
	var s strings.Builder
	
	s.WriteString(m.renderHeader())
	s.WriteString("\n")
	
	// Auth title
	authTitle := "🔐 GitHub Authentication Required"
	s.WriteString(titleStyle.Render(authTitle))
	s.WriteString("\n")
	
	// Render the auth model with enhanced container
	authContent := m.authModel.View()
	authStyle := lipgloss.NewStyle().
		Padding(2, 4).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(secondaryColor)
	
	s.WriteString(authStyle.Render(authContent))
	
	return baseStyle.Render(s.String())
}