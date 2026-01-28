package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("170")).
			PaddingLeft(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			PaddingLeft(1)

	sectionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("81")).
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

	sidebarTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("170")).
				PaddingLeft(1)

	nextTagStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("0")).
			Background(lipgloss.Color("213"))

	invalidTagStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("0")).
			Background(lipgloss.Color("203"))

	nextTag    = nextTagStyle.Render("NEXT")
	invalidTag = invalidTagStyle.Render("INVALID")

	checkboxChecked = completedStyle.Render("[x]")
	checkboxUnchecked = incompleteStyle.Render("[ ]")

	priorityStyles = map[int]lipgloss.Style{
		1: lipgloss.NewStyle().Foreground(lipgloss.Color("196")),  // red
		2: lipgloss.NewStyle().Foreground(lipgloss.Color("208")),  // orange
		3: lipgloss.NewStyle().Foreground(lipgloss.Color("226")),  // yellow
		4: lipgloss.NewStyle().Foreground(lipgloss.Color("118")),  // light green
		5: lipgloss.NewStyle().Foreground(lipgloss.Color("34")),   // green
	}
)

func renderPriority(p int) string {
	s, ok := priorityStyles[p]
	if !ok {
		s = dimStyle
	}
	return s.Render(fmt.Sprintf("P%d", p))
}

// truncate truncates a string to maxLen characters, adding "..." if truncated
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
