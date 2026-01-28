package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/aloisdeniel/prd/prd"
)

type featureDetailModel struct {
	feature  *prd.Feature
	path     string
	id       string
	progress map[int]bool
	cursor   int
	width    int
	height   int
	err      error
	basePath string
}

func newFeatureDetailModel(basePath, id string) featureDetailModel {
	return featureDetailModel{basePath: basePath, id: id}
}

type featureLoadedMsg struct {
	path     string
	feature  *prd.Feature
	progress map[int]bool
}

func (m featureDetailModel) loadFeature() tea.Msg {
	path, feature, err := prd.LoadFeature(m.basePath, m.id)
	if err != nil {
		return errMsg{err}
	}
	progress, _ := prd.LoadProgress(path)
	return featureLoadedMsg{path, feature, progress}
}

func (m featureDetailModel) Init() tea.Cmd {
	return m.loadFeature
}

func (m featureDetailModel) Update(msg tea.Msg) (featureDetailModel, tea.Cmd) {
	switch msg := msg.(type) {
	case featureLoadedMsg:
		m.feature = msg.feature
		m.path = msg.path
		m.progress = msg.progress
		m.err = nil
		if m.cursor >= len(m.feature.UserStories) {
			m.cursor = max(0, len(m.feature.UserStories)-1)
		}
	case errMsg:
		m.err = msg.err
	case tea.KeyMsg:
		if m.feature == nil {
			return m, nil
		}
		switch {
		case key.Matches(msg, keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, keys.Down):
			if m.cursor < len(m.feature.UserStories)-1 {
				m.cursor++
			}
		}
	}
	return m, nil
}

func (m featureDetailModel) View() string {
	var b strings.Builder

	if m.feature == nil {
		b.WriteString(dimStyle.PaddingLeft(2).Render("Loading..."))
		return b.String()
	}

	b.WriteString(titleStyle.Render(m.feature.Name))
	b.WriteString("\n")

	if m.feature.Description != "" {
		desc := m.feature.Description
		if len(desc) > 200 {
			desc = desc[:200] + "..."
		}
		b.WriteString(subtitleStyle.Render(desc))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(subtitleStyle.Render("User Stories"))
	b.WriteString("\n\n")

	if len(m.feature.UserStories) == 0 {
		b.WriteString(dimStyle.PaddingLeft(2).Render("No user stories. Press 'n' to create one."))
		return b.String()
	}

	nextID := nextStoryID(m.feature.UserStories, m.progress)

	for i, us := range m.feature.UserStories {
		cursor := "  "
		style := normalStyle
		if i == m.cursor {
			cursor = selectedStyle.Render("> ")
			style = selectedStyle
		}

		status := checkboxUnchecked
		if m.progress[us.ID] {
			status = checkboxChecked
		}

		tag := ""
		if us.ID == nextID {
			tag = " " + nextTag
		}

		line := fmt.Sprintf("%s%s %s US-%d  %s%s",
			cursor,
			status,
			renderPriority(us.Priority),
			us.ID,
			style.Render(us.Name),
			tag,
		)
		b.WriteString(line + "\n")
	}

	return b.String()
}

func (m featureDetailModel) selectedStory() *prd.UserStory {
	if m.feature != nil && m.cursor >= 0 && m.cursor < len(m.feature.UserStories) {
		return &m.feature.UserStories[m.cursor]
	}
	return nil
}

func nextStoryID(stories []prd.UserStory, progress map[int]bool) int {
	bestID := -1
	bestPriority := 6 // higher than max (5)
	for _, us := range stories {
		if !progress[us.ID] && us.Priority < bestPriority {
			bestPriority = us.Priority
			bestID = us.ID
		}
	}
	return bestID
}

func (m featureDetailModel) statusHelp() string {
	parts := []string{"j/k navigate", "enter open", "u/n new story", "d delete", "esc back"}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(strings.Join(parts, "  "))
}
