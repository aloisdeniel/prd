package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/aloisdeniel/prd/prd"
)

type userStoryDetailModel struct {
	story     *prd.UserStory
	path      string
	completed bool
	viewport  viewport.Model
	ready     bool
	width     int
	height    int
	err       error
	basePath  string
	featID    string
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
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case storyReloadedMsg:
		m.story = msg.story
		m.completed = msg.completed
		m.err = nil
		m.updateViewportContent()
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
		case key.Matches(msg, keys.ScrollUp), key.Matches(msg, keys.ScrollDown):
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}
	}
	return m, nil
}

func (m *userStoryDetailModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.viewport.Width = width
	m.viewport.Height = height
	m.ready = true
	m.updateViewportContent()
}

func (m *userStoryDetailModel) updateViewportContent() {
	if !m.ready {
		return
	}
	m.viewport.SetContent(m.renderContent())
}

func (m userStoryDetailModel) View() string {
	if m.story == nil {
		return dimStyle.PaddingLeft(2).Render("Loading...")
	}
	if !m.ready {
		return dimStyle.PaddingLeft(2).Render("Initializing...")
	}
	return m.viewport.View()
}

func (m userStoryDetailModel) renderContent() string {
	if m.story == nil {
		return ""
	}

	var b strings.Builder

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

	return b.String()
}

