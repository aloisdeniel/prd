package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/list"

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

	var sections []string

	// Title with bottom margin
	sections = append(sections, lipgloss.NewStyle().MarginBottom(1).Render(titleStyle.Render(m.feature.Name)))

	// Description
	if m.feature.Description != "" {
		desc := lipgloss.NewStyle().MaxWidth(m.width - 4).Render(m.feature.Description)
		sections = append(sections, subtitleStyle.Render(desc))
	}

	// User Stories section with top margin
	sectionHeaderStyle := sectionStyle.MarginTop(1)
	sections = append(sections, sectionHeaderStyle.Render("User Stories"))

	if len(m.feature.UserStories) == 0 {
		sections = append(sections, dimStyle.PaddingLeft(2).Render("No user stories. Press 'n' to create one."))
	} else {
		sections = append(sections, m.renderUserStoriesList())
	}

	// Other sections
	sections = append(sections, m.renderTextSection("Goals", m.feature.Goals)...)
	sections = append(sections, m.renderTextSection("Functional Requirements", m.feature.FunctionalRequirements)...)
	sections = append(sections, m.renderTextSection("Non-Goals", m.feature.NonGoals)...)
	sections = append(sections, m.renderTextSection("Technical Considerations", m.feature.TechnicalConsiderations)...)
	sections = append(sections, m.renderTextSection("Analytics", m.feature.Analytics)...)
	sections = append(sections, m.renderTextSection("Risks", m.feature.Risks)...)
	sections = append(sections, m.renderTextSection("Success Metrics", m.feature.SuccessMetrics)...)
	sections = append(sections, m.renderTextSection("Open Questions", m.feature.OpenQuestions)...)
	sections = append(sections, m.renderNotesSection()...)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m featureDetailModel) renderUserStoriesList() string {
	nextID := nextStoryID(m.feature.UserStories, m.progress)

	// Build list items
	l := list.New().
		Enumerator(func(_ list.Items, _ int) string { return "" })

	for i, us := range m.feature.UserStories {
		l.Item(m.renderStoryItem(i, us, nextID))
	}

	return l.String()
}

func (m featureDetailModel) renderStoryItem(idx int, us prd.UserStory, nextID int) string {
	isSelected := idx == m.cursor
	style := normalStyle
	if isSelected {
		style = selectedStyle
	}

	cursor := "  "
	if isSelected {
		cursor = selectedStyle.Render("> ")
	}

	status := checkboxUnchecked
	if m.progress[us.ID] {
		status = checkboxChecked
	}

	tag := ""
	// Layout: "> [x] P1 US-XX  name NEXT"
	fixedWidth := 2 + 3 + 1 + 2 + 1 + 5 + 2 // cursor + checkbox + space + priority + space + US-XX + spaces
	if us.ID == nextID {
		tag = " " + nextTag
		fixedWidth += 5
	}
	maxNameWidth := m.width - fixedWidth
	if maxNameWidth < 10 {
		maxNameWidth = 10
	}

	name := truncate(us.Name, maxNameWidth)
	return fmt.Sprintf("%s%s %s US-%d  %s%s", cursor, status, renderPriority(us.Priority), us.ID, style.Render(name), tag)
}

func (m featureDetailModel) renderTextSection(title, content string) []string {
	if content == "" {
		return nil
	}

	// Section header with top margin
	headerStyle := sectionStyle.MarginTop(1)
	// Content with proper width
	contentStyle := dimStyle.PaddingLeft(2).MaxWidth(m.width - 4)

	return []string{
		headerStyle.Render(title),
		contentStyle.Render(content),
	}
}

func (m featureDetailModel) renderNotesSection() []string {
	if len(m.feature.Notes) == 0 {
		return nil
	}

	// Section header with top margin
	headerStyle := sectionStyle.MarginTop(1)
	result := []string{headerStyle.Render("Notes")}

	// Build notes as a list
	l := list.New().
		EnumeratorStyle(dimStyle).
		ItemStyle(dimStyle)

	for _, note := range m.feature.Notes {
		dateStr := ""
		if note.Date != "" {
			dateStr = note.Date + ": "
		}
		noteText := lipgloss.NewStyle().MaxWidth(m.width - 8).Render(dateStr + note.Content)
		l.Item(noteText)
	}

	result = append(result, l.String())
	return result
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
