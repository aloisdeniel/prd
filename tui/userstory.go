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
	story        *prd.UserStory
	path         string
	completed    bool
	scrollOffset int
	width        int
	height       int
	err          error
	basePath     string
	featID       string
}

func newUserStoryDetailModel(basePath, featID, path string, story *prd.UserStory, completed bool) userStoryDetailModel {
	return userStoryDetailModel{
		basePath:  basePath,
		featID:    featID,
		path:      path,
		story:     story,
		completed: completed,
	}
}

type storyReloadedMsg struct {
	story     *prd.UserStory
	completed bool
}

func (m userStoryDetailModel) reloadStory() tea.Msg {
	_, feature, err := prd.LoadFeature(m.basePath, m.featID)
	if err != nil {
		return errMsg{err}
	}
	progress, _ := prd.LoadProgress(m.path)
	for i := range feature.UserStories {
		if feature.UserStories[i].ID == m.story.ID {
			return storyReloadedMsg{&feature.UserStories[i], progress[feature.UserStories[i].ID]}
		}
	}
	return errMsg{fmt.Errorf("user story US-%d not found after reload", m.story.ID)}
}

func (m userStoryDetailModel) Update(msg tea.Msg) (userStoryDetailModel, tea.Cmd) {
	switch msg := msg.(type) {
	case storyReloadedMsg:
		m.story = msg.story
		m.completed = msg.completed
		m.err = nil
	case errMsg:
		m.err = msg.err
	case tea.KeyMsg:
		if m.story == nil {
			return m, nil
		}
		switch {
		case key.Matches(msg, keys.Complete):
			if err := prd.CompleteUserStory(m.path, m.story.ID); err != nil {
				m.err = err
				return m, nil
			}
			return m, m.reloadStory
		case key.Matches(msg, keys.ScrollUp):
			m.scrollOffset -= m.scrollAmount()
			if m.scrollOffset < 0 {
				m.scrollOffset = 0
			}
		case key.Matches(msg, keys.ScrollDown):
			m.scrollOffset += m.scrollAmount()
		}
	}
	return m, nil
}

func (m userStoryDetailModel) scrollAmount() int {
	if m.height > 4 {
		return m.height / 2
	}
	return 5
}

func (m userStoryDetailModel) View() string {
	var b strings.Builder

	if m.story == nil {
		b.WriteString(dimStyle.PaddingLeft(2).Render("Loading..."))
		return b.String()
	}

	status := checkboxUnchecked
	if m.completed {
		status = checkboxChecked
	}
	b.WriteString(titleStyle.Render(fmt.Sprintf("%s %s US-%d | %s", status, renderPriority(m.story.Priority), m.story.ID, m.story.Name)))
	b.WriteString("\n")

	if m.story.Description != "" {
		b.WriteString(subtitleStyle.Render(m.story.Description))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(subtitleStyle.Render("Acceptance Criteria"))
	b.WriteString("\n\n")

	for _, ac := range m.story.AcceptanceCriteria {
		checkbox := checkboxUnchecked
		if ac.Completed {
			checkbox = checkboxChecked
		}

		line := fmt.Sprintf("  %s %s", checkbox, dimStyle.Render(ac.Text))
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

	return m.applyScroll(b.String())
}

func (m userStoryDetailModel) applyScroll(content string) string {
	lines := strings.Split(content, "\n")
	totalLines := len(lines)

	// Clamp scroll offset
	maxOffset := totalLines - m.height
	if maxOffset < 0 {
		maxOffset = 0
	}
	offset := m.scrollOffset
	if offset > maxOffset {
		offset = maxOffset
	}

	// Calculate visible range
	start := offset
	end := offset + m.height
	if end > totalLines {
		end = totalLines
	}

	if start >= totalLines {
		return ""
	}

	return strings.Join(lines[start:end], "\n")
}

func (m userStoryDetailModel) statusHelp() string {
	parts := []string{"ctrl+u/d scroll", "c toggle complete", "esc back"}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(strings.Join(parts, "  "))
}
