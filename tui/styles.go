package tui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("170")).
			PaddingLeft(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			PaddingLeft(1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("170")).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	completedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("34"))

	incompleteStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("203"))

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Background(lipgloss.Color("236")).
			PaddingLeft(1).
			PaddingRight(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("203")).
			PaddingLeft(1)

	inputLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("170")).
			Bold(true).
			PaddingLeft(1)

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("170"))

	checkboxChecked   = completedStyle.Render("[x]")
	checkboxUnchecked = incompleteStyle.Render("[ ]")
)
