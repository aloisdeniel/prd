package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/aloisdeniel/prd/prd"
)

type userStoryDetailModel struct {
	story    *prd.UserStory
	path     string
	cursor   int
	width    int
	height   int
	err      error
	basePath string
	featID   string
}

func newUserStoryDetailModel(basePath, featID, path string, story *prd.UserStory) userStoryDetailModel {
	return userStoryDetailModel{
		basePath: basePath,
		featID:   featID,
		path:     path,
		story:    story,
	}
}

type storyReloadedMsg struct {
	story *prd.UserStory
}

func (m userStoryDetailModel) reloadStory() tea.Msg {
	_, feature, err := prd.LoadFeature(m.basePath, m.featID)
	if err != nil {
		return errMsg{err}
	}
	for i := range feature.UserStories {
		if feature.UserStories[i].ID == m.story.ID {
			return storyReloadedMsg{&feature.UserStories[i]}
		}
	}
	return errMsg{fmt.Errorf("user story US-%d not found after reload", m.story.ID)}
}

func (m userStoryDetailModel) Update(msg tea.Msg) (userStoryDetailModel, tea.Cmd) {
	switch msg := msg.(type) {
	case storyReloadedMsg:
		m.story = msg.story
		m.err = nil
		if m.cursor >= len(m.story.AcceptanceCriteria) {
			m.cursor = max(0, len(m.story.AcceptanceCriteria)-1)
		}
	case errMsg:
		m.err = msg.err
	case tea.KeyMsg:
		if m.story == nil {
			return m, nil
		}
		switch {
		case key.Matches(msg, keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, keys.Down):
			if m.cursor < len(m.story.AcceptanceCriteria)-1 {
				m.cursor++
			}
		case key.Matches(msg, keys.Space):
			if m.cursor >= 0 && m.cursor < len(m.story.AcceptanceCriteria) {
				acNum := m.cursor + 1
				if err := prd.ToggleAcceptanceCriterion(m.path, m.story.ID, acNum); err != nil {
					m.err = err
					return m, nil
				}
				return m, m.reloadStory
			}
		case key.Matches(msg, keys.Complete):
			if err := prd.CompleteUserStory(m.path, m.story.ID); err != nil {
				m.err = err
				return m, nil
			}
			return m, m.reloadStory
		}
	}
	return m, nil
}

func (m userStoryDetailModel) View() string {
	var b strings.Builder

	if m.story == nil {
		b.WriteString(dimStyle.PaddingLeft(2).Render("Loading..."))
		return b.String()
	}

	b.WriteString(titleStyle.Render(fmt.Sprintf("US-%d | %s", m.story.ID, m.story.Name)))
	b.WriteString("\n")

	if m.story.Description != "" {
		b.WriteString(subtitleStyle.Render(m.story.Description))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(subtitleStyle.Render("Acceptance Criteria"))
	b.WriteString("\n\n")

	for i, ac := range m.story.AcceptanceCriteria {
		cursor := "  "
		style := normalStyle
		if i == m.cursor {
			cursor = selectedStyle.Render("> ")
			style = selectedStyle
		}

		checkbox := checkboxUnchecked
		if ac.Completed {
			checkbox = checkboxChecked
		}

		line := fmt.Sprintf("%s%s %s", cursor, checkbox, style.Render(ac.Text))
		b.WriteString(line + "\n")
	}

	if m.story.TechnicalConsiderations != "" {
		b.WriteString("\n")
		b.WriteString(subtitleStyle.Render("Technical Considerations"))
		b.WriteString("\n")
		b.WriteString(dimStyle.PaddingLeft(2).Render(m.story.TechnicalConsiderations))
		b.WriteString("\n")
	}

	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render(m.err.Error()))
	}

	return b.String()
}

func (m userStoryDetailModel) statusHelp() string {
	parts := []string{"j/k navigate", "space toggle", "c complete all", "esc back"}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(strings.Join(parts, "  "))
}
