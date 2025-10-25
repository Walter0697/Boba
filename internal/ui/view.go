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
		Padding(1, 2).
		Margin(0, 1)
	
	// Header styles
	headerStyle = lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true).
		Align(lipgloss.Center).
		Margin(1, 0)
	
	titleStyle = lipgloss.NewStyle().
		Foreground(secondaryColor).
		Bold(true).
		Align(lipgloss.Center).
		Padding(1, 0).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(secondaryColor).
		Margin(0, 2)
	
	// Menu item styles
	menuItemStyle = lipgloss.NewStyle().
		Padding(0, 2).
		Margin(0, 1)
	
	selectedMenuItemStyle = lipgloss.NewStyle().
		Foreground(accentColor).
		Bold(true).
		Padding(0, 2).
		Margin(0, 1).
		Background(lipgloss.Color("235")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accentColor)
	
	// Help text style
	helpStyle = lipgloss.NewStyle().
		Foreground(mutedColor).
		Italic(true).
		Align(lipgloss.Center).
		Margin(1, 0).
		Padding(1, 2).
		Border(lipgloss.NormalBorder()).
		BorderForeground(mutedColor)
	
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
		Padding(1, 2).
		Border(lipgloss.ThickBorder()).
		BorderForeground(errorColor).
		Margin(1, 0)
	
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
		return "🫧 Select an option: 🫧"
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
	
	var s strings.Builder
	
	// ASCII Art Header with enhanced styling
	s.WriteString(m.renderHeader())
	s.WriteString("\n")
	
	// Menu title with enhanced border
	s.WriteString(titleStyle.Render(m.getMenuTitle()))
	s.WriteString("\n\n")
	
	// Enhanced menu items with better visual hierarchy
	s.WriteString(m.renderMenuItems())
	s.WriteString("\n")
	
	// Enhanced help text with border
	s.WriteString(helpStyle.Render(m.getHelpText()))
	
	// Show authentication error if present
	if m.authError != "" {
		s.WriteString("\n")
		s.WriteString(errorStyle.Render("⚠️  " + m.authError))
	}
	
	return baseStyle.Render(s.String())
}

// renderHeader creates the ASCII art header with enhanced styling
func (m MenuModel) renderHeader() string {
	myFigure := figure.NewFigure("BOBA", "", true)
	header := myFigure.String()
	
	// Add subtitle
	subtitle := "🫧 Development Environment Setup Tool 🫧"
	
	headerContent := header + "\n" + 
		lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true).
			Align(lipgloss.Center).
			Render(subtitle)
	
	return headerStyle.Render(headerContent)
}

// renderMenuItems creates enhanced menu items with better visual feedback
func (m MenuModel) renderMenuItems() string {
	var items []string
	currentChoices := m.getMenuChoices()
	
	for i, choice := range currentChoices {
		var itemText string
		
		if m.cursor == i {
			// Selected item with enhanced styling
			itemText = selectedMenuItemStyle.Render(fmt.Sprintf("▶ %s", choice))
		} else {
			// Regular item
			itemText = menuItemStyle.Render(fmt.Sprintf("  %s", choice))
		}
		
		items = append(items, itemText)
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