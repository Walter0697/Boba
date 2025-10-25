package ui

// navigateToMenu changes the current menu and updates the navigation stack
func (m *MenuModel) navigateToMenu(menuType MenuType) {
	m.menuStack = append(m.menuStack, m.currentMenu)
	m.currentMenu = menuType
	m.choices = m.getMenuChoices()
	m.cursor = 0
}

// navigateBack returns to the previous menu
func (m *MenuModel) navigateBack() {
	if len(m.menuStack) > 0 {
		// Pop the last menu from the stack
		m.currentMenu = m.menuStack[len(m.menuStack)-1]
		m.menuStack = m.menuStack[:len(m.menuStack)-1]
		m.choices = m.getMenuChoices()
		m.cursor = 0
	}
}

// requiresAuthentication checks if the current menu action requires GitHub auth
func (m *MenuModel) requiresAuthentication() bool {
	switch m.currentMenu {
	case InstallEverythingMenu, ToolsListMenu:
		return !m.isGitHubAuthenticated()
	default:
		return false
	}
}