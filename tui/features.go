package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/aloisdeniel/prd/prd"
)

type featureListModel struct {
	entries  []prd.FeatureEntry
	cursor   int
	width    int
	height   int
	err      error
	basePath string
}

func newFeatureListModel(basePath string) featureListModel {
	return featureListModel{basePath: basePath}
}

func (m featureListModel) loadEntries() tea.Msg {
	entries, err := prd.ListFeatures(m.basePath)
	if err != nil {
		return errMsg{err}
	}
	return featureListMsg(entries)
}

type featureListMsg []prd.FeatureEntry

func (m featureListModel) Init() tea.Cmd {
	return m.loadEntries
}

func (m featureListModel) Update(msg tea.Msg) (featureListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case featureListMsg:
		m.entries = msg
		m.err = nil
		if m.cursor >= len(m.entries) {
			m.cursor = max(0, len(m.entries)-1)
		}
	case errMsg:
		m.err = msg.err
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, keys.Down):
			if m.cursor < len(m.entries)-1 {
				m.cursor++
			}
		}
	}
	return m, nil
}

func (m featureListModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Features"))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(errorStyle.Render(m.err.Error()))
		return b.String()
	}

	if len(m.entries) == 0 {
		b.WriteString(dimStyle.PaddingLeft(2).Render("No features found. Press 'n' to create one."))
		return b.String()
	}

	for i, e := range m.entries {
		cursor := "  "
		style := normalStyle
		if i == m.cursor {
			cursor = selectedStyle.Render("> ")
			style = selectedStyle
		}

		progress := fmt.Sprintf("[%d/%d]", e.Completed, e.Total)
		pStyle := incompleteStyle
		if e.Completed == e.Total && e.Total > 0 {
			pStyle = completedStyle
		}

		line := fmt.Sprintf("%s%s  %s  %s",
			cursor,
			style.Render(e.ID),
			style.Render(e.Name),
			pStyle.Render(progress),
		)
		b.WriteString(line + "\n")
	}

	return b.String()
}

func (m featureListModel) selectedEntry() *prd.FeatureEntry {
	if m.cursor >= 0 && m.cursor < len(m.entries) {
		return &m.entries[m.cursor]
	}
	return nil
}

func (m featureListModel) statusHelp() string {
	parts := []string{"j/k navigate", "enter open", "f/n new feature", "/ search", "q quit"}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(strings.Join(parts, "  "))
}
