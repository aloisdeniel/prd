package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/aloisdeniel/prd/prd"
)

const sidebarWidth = 35

type sidebarItem struct {
	isStory   bool
	featureID string
	feature   *prd.Feature
	storyIdx  int
	progress  map[int]bool
	expanded  bool
	entryIdx  int      // index into allFeats
	invalid   bool     // True if the document has errors
	errors    []string // Parse or validation errors
}

type sidebarEntry struct {
	entry    prd.FeatureEntry
	feature  *prd.Feature
	progress map[int]bool
	expanded bool
	path     string
	invalid  bool     // True if the document has errors
	errors   []string // Parse or validation errors
}

type sidebarModel struct {
	basePath string
	items    []sidebarItem
	allFeats []sidebarEntry
	cursor   int
	width    int
	height   int
	err      error
}

type sidebarLoadedMsg struct {
	entries []sidebarEntry
}

func newSidebarModel(basePath string) sidebarModel {
	return sidebarModel{basePath: basePath, width: sidebarWidth}
}

func (m sidebarModel) loadAll() tea.Msg {
	entries, err := prd.ListFeatures(m.basePath)
	if err != nil {
		return errMsg{err}
	}
	var result []sidebarEntry
	for _, e := range entries {
		// Check if entry has errors (invalid document)
		if len(e.Errors) > 0 {
			result = append(result, sidebarEntry{
				entry:   e,
				path:    e.Path,
				invalid: true,
				errors:  e.Errors,
			})
			continue
		}

		path, feature, err := prd.LoadFeature(m.basePath, e.ID)
		if err != nil {
			continue
		}
		progress, _ := prd.LoadProgress(path)
		result = append(result, sidebarEntry{
			entry:    e,
			feature:  feature,
			progress: progress,
			path:     path,
		})
	}
	return sidebarLoadedMsg{result}
}

func (m *sidebarModel) rebuildItems() {
	m.items = nil
	for i := range m.allFeats {
		fe := &m.allFeats[i]
		m.items = append(m.items, sidebarItem{
			isStory:   false,
			featureID: fe.entry.ID,
			feature:   fe.feature,
			progress:  fe.progress,
			expanded:  fe.expanded,
			entryIdx:  i,
			invalid:   fe.invalid,
			errors:    fe.errors,
		})
		if fe.expanded && fe.feature != nil {
			for si := range fe.feature.UserStories {
				m.items = append(m.items, sidebarItem{
					isStory:   true,
					featureID: fe.entry.ID,
					feature:   fe.feature,
					storyIdx:  si,
					progress:  fe.progress,
					entryIdx:  i,
				})
			}
		}
	}
	if m.cursor >= len(m.items) {
		m.cursor = max(0, len(m.items)-1)
	}
}

func (m *sidebarModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m sidebarModel) Init() tea.Cmd {
	return m.loadAll
}

func (m sidebarModel) Update(msg tea.Msg) (sidebarModel, tea.Cmd) {
	switch msg := msg.(type) {
	case sidebarLoadedMsg:
		// Preserve expanded state from previous allFeats
		oldExpanded := make(map[string]bool)
		for _, fe := range m.allFeats {
			oldExpanded[fe.entry.ID] = fe.expanded
		}
		m.allFeats = msg.entries
		for i := range m.allFeats {
			if exp, ok := oldExpanded[m.allFeats[i].entry.ID]; ok {
				m.allFeats[i].expanded = exp
			}
		}
		m.err = nil
		m.rebuildItems()
	case errMsg:
		m.err = msg.err
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, keys.Down):
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case key.Matches(msg, keys.Space), key.Matches(msg, keys.Enter):
			if len(m.items) > 0 && !m.items[m.cursor].isStory && !m.items[m.cursor].invalid {
				idx := m.items[m.cursor].entryIdx
				m.allFeats[idx].expanded = !m.allFeats[idx].expanded
				m.rebuildItems()
			}
			// If it's a story, the parent tui.go handles opening the detail
		}
	}
	return m, nil
}

func (m sidebarModel) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")).
		MarginBottom(1).
		PaddingLeft(1)

	title := titleStyle.Render("Features")

	if m.err != nil {
		return lipgloss.JoinVertical(lipgloss.Left, title, errorStyle.Render(m.err.Error()))
	}

	if len(m.items) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left, title, dimStyle.PaddingLeft(1).Render("No features."))
	}

	// Render items directly without tree (simpler, no extra padding)
	var lines []string
	for i, item := range m.items {
		if item.isStory {
			lines = append(lines, m.renderStoryLine(i, item))
		} else {
			lines = append(lines, m.renderFeatureLine(i, item))
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return lipgloss.JoinVertical(lipgloss.Left, title, content)
}

func (m sidebarModel) renderFeatureLine(idx int, item sidebarItem) string {
	fe := &m.allFeats[item.entryIdx]
	isSelected := idx == m.cursor
	style := normalStyle
	if isSelected {
		style = selectedStyle
	}

	cursor := " "
	if isSelected {
		cursor = selectedStyle.Render(">")
	}

	if fe.invalid {
		// Invalid document - calculate available width for name
		// Layout: "> F-XX  name INVALID"
		fixedWidth := 3 + len(fe.entry.ID) + 2 + 1 + 7 // cursor + id + spaces + space + INVALID
		maxNameWidth := m.width - fixedWidth
		if maxNameWidth < 5 {
			maxNameWidth = 5
		}
		name := truncate(fe.entry.Name, maxNameWidth)
		return fmt.Sprintf("%s %s  %s %s", cursor, style.Render(fe.entry.ID), errorStyle.Render(name), invalidTag)
	}

	// Valid document
	// Layout: ">▾ F-XX  name [0/0]"
	arrow := "▸"
	if fe.expanded {
		arrow = "▾"
	}
	progress := fmt.Sprintf("[%d/%d]", fe.entry.Completed, fe.entry.Total)
	fixedWidth := 1 + 1 + 1 + len(fe.entry.ID) + 2 + 1 + len(progress) // cursor + arrow + space + id + spaces + space + progress
	maxNameWidth := m.width - fixedWidth
	if maxNameWidth < 5 {
		maxNameWidth = 5
	}

	pStyle := incompleteStyle
	if fe.entry.Completed == fe.entry.Total && fe.entry.Total > 0 {
		pStyle = completedStyle
	}
	name := truncate(fe.entry.Name, maxNameWidth)
	return fmt.Sprintf("%s%s %s  %s %s", cursor, arrow, style.Render(fe.entry.ID), style.Render(name), pStyle.Render(progress))
}

func (m sidebarModel) renderStoryLine(idx int, item sidebarItem) string {
	fe := &m.allFeats[item.entryIdx]
	us := fe.feature.UserStories[item.storyIdx]

	isSelected := idx == m.cursor
	style := normalStyle
	if isSelected {
		style = selectedStyle
	}

	cursor := " "
	if isSelected {
		cursor = selectedStyle.Render(">")
	}

	status := checkboxUnchecked
	if fe.progress[us.ID] {
		status = checkboxChecked
	}

	isNext := us.ID == nextStoryID(fe.feature.UserStories, fe.progress)
	tag := ""
	// Layout: ">   [x] US-XX  name NEXT"
	fixedWidth := 1 + 3 + 3 + 1 + 4 + 2 // cursor + indent + checkbox + space + US-XX + spaces
	if isNext {
		tag = " " + nextTag
		fixedWidth += 5 // " NEXT"
	}
	maxNameWidth := m.width - fixedWidth
	if maxNameWidth < 5 {
		maxNameWidth = 5
	}

	name := truncate(us.Name, maxNameWidth)
	return fmt.Sprintf("%s   %s US-%d  %s%s", cursor, status, us.ID, style.Render(name), tag)
}


func (m sidebarModel) selected() *sidebarItem {
	if m.cursor >= 0 && m.cursor < len(m.items) {
		return &m.items[m.cursor]
	}
	return nil
}

// cursor returns the current cursor position for compatibility with tui.go
func (m sidebarModel) cursorPos() int {
	return m.cursor
}

func (m sidebarModel) selectedFeatureEntry() *sidebarEntry {
	sel := m.selected()
	if sel == nil {
		return nil
	}
	return &m.allFeats[sel.entryIdx]
}
