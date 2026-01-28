package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/aloisdeniel/prd/prd"
)

type featureDetailModel struct {
	feature  *prd.Feature
	path     string
	id       string
	progress map[int]bool
	cursor   int
	viewport viewport.Model
	ready    bool
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
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case featureLoadedMsg:
		m.feature = msg.feature
		m.path = msg.path
		m.progress = msg.progress
		m.err = nil
		if m.cursor >= len(m.feature.UserStories) {
			m.cursor = max(0, len(m.feature.UserStories)-1)
		}
		m.updateViewportContent()
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
		case key.Matches(msg, keys.ScrollUp), key.Matches(msg, keys.ScrollDown):
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}
	}
	return m, nil
}

func (m *featureDetailModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.viewport.Width = width
	m.viewport.Height = height
	m.ready = true
	m.updateViewportContent()
}

func (m *featureDetailModel) updateViewportContent() {
	if !m.ready {
		return
	}
	m.viewport.SetContent(m.renderContent())
}

func (m featureDetailModel) View() string {
	if m.feature == nil {
		return dimStyle.PaddingLeft(2).Render("Loading...")
	}
	if !m.ready {
		return dimStyle.PaddingLeft(2).Render("Initializing...")
	}
	return m.viewport.View()
}

func (m featureDetailModel) renderContent() string {
	if m.feature == nil {
		return ""
	}

	var b strings.Builder

	b.WriteString(titleStyle.Render(m.feature.Name))
	b.WriteString("\n\n")

	if m.feature.Description != "" {
		desc := m.feature.Description
		if len(desc) > 200 {
			desc = desc[:200] + "..."
		}
		b.WriteString(subtitleStyle.Render(desc))
		b.WriteString("\n")
	}

	// User Stories section (always first)
	b.WriteString("\n")
	b.WriteString(sectionStyle.Render("User Stories"))
	b.WriteString("\n\n")

	if len(m.feature.UserStories) == 0 {
		b.WriteString(dimStyle.PaddingLeft(2).Render("No user stories. Press 'n' to create one."))
	} else {
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
	}

	// Other sections
	m.renderSection(&b, "Goals", m.feature.Goals)
	m.renderSection(&b, "Functional Requirements", m.feature.FunctionalRequirements)
	m.renderSection(&b, "Non-Goals", m.feature.NonGoals)
	m.renderSection(&b, "Technical Considerations", m.feature.TechnicalConsiderations)
	m.renderSection(&b, "Analytics", m.feature.Analytics)
	m.renderSection(&b, "Risks", m.feature.Risks)
	m.renderSection(&b, "Success Metrics", m.feature.SuccessMetrics)
	m.renderSection(&b, "Open Questions", m.feature.OpenQuestions)
	m.renderNotes(&b)

	return b.String()
}

func (m featureDetailModel) renderSection(b *strings.Builder, title, content string) {
	if content == "" {
		return
	}
	b.WriteString("\n")
	b.WriteString(sectionStyle.Render(title))
	b.WriteString("\n\n")
	// Truncate long content for display
	display := content
	if len(display) > 300 {
		display = display[:300] + "..."
	}
	b.WriteString(dimStyle.PaddingLeft(2).Render(display))
	b.WriteString("\n")
}

func (m featureDetailModel) renderNotes(b *strings.Builder) {
	if len(m.feature.Notes) == 0 {
		return
	}
	b.WriteString("\n")
	b.WriteString(sectionStyle.Render("Notes"))
	b.WriteString("\n\n")
	for _, note := range m.feature.Notes {
		dateStr := ""
		if note.Date != "" {
			dateStr = note.Date + ": "
		}
		content := note.Content
		if len(content) > 100 {
			content = content[:100] + "..."
		}
		b.WriteString(dimStyle.PaddingLeft(2).Render(dateStr + content))
		b.WriteString("\n")
	}
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
