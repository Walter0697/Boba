package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// UIManager handles the interactive terminal interface
type UIManager struct {
	program *tea.Program
}

// NewUIManager creates a new UI manager
func NewUIManager() *UIManager {
	return &UIManager{}
}

// Start initializes and runs the UI
func (ui *UIManager) Start() error {
	p := tea.NewProgram(InitialModel(), tea.WithAltScreen())
	ui.program = p
	_, err := p.Run()
	return err
}