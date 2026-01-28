package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

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
	entryIdx  int // index into allFeats
}

type sidebarEntry struct {
	entry    prd.FeatureEntry
	feature  *prd.Feature
	progress map[int]bool
	expanded bool
	path     string
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
		})
		if fe.expanded {
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
		case key.Matches(msg, keys.Space):
			if len(m.items) > 0 && !m.items[m.cursor].isStory {
				idx := m.items[m.cursor].entryIdx
				m.allFeats[idx].expanded = !m.allFeats[idx].expanded
				m.rebuildItems()
			}
		case key.Matches(msg, keys.Enter):
			if len(m.items) > 0 && !m.items[m.cursor].isStory {
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
	var b strings.Builder

	b.WriteString("\n\n")
	b.WriteString(sidebarTitleStyle.Render("Features"))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(errorStyle.Render(m.err.Error()))
		return b.String()
	}

	if len(m.items) == 0 {
		b.WriteString(dimStyle.Render(" No features."))
		return b.String()
	}

	for i, item := range m.items {
		cursor := "  "
		style := normalStyle
		if i == m.cursor {
			cursor = selectedStyle.Render("> ")
			style = selectedStyle
		}

		if !item.isStory {
			// Feature row
			arrow := "▸"
			if item.expanded {
				arrow = "▾"
			}
			fe := m.allFeats[item.entryIdx]
			progress := fmt.Sprintf("[%d/%d]", fe.entry.Completed, fe.entry.Total)
			pStyle := incompleteStyle
			if fe.entry.Completed == fe.entry.Total && fe.entry.Total > 0 {
				pStyle = completedStyle
			}
			name := truncate(fe.entry.Name, sidebarWidth-16)
			line := fmt.Sprintf("%s%s %s  %s %s", cursor, arrow, style.Render(fe.entry.ID), style.Render(name), pStyle.Render(progress))
			b.WriteString(line + "\n")
		} else {
			// User story row
			us := item.feature.UserStories[item.storyIdx]
			status := checkboxUnchecked
			if item.progress[us.ID] {
				status = checkboxChecked
			}
			isNext := us.ID == nextStoryID(item.feature.UserStories, item.progress)
			tag := ""
			if isNext {
				tag = " " + nextTag
			}
			maxName := sidebarWidth - 15
			if isNext {
				maxName -= 6 // room for " NEXT"
			}
			name := truncate(us.Name, maxName)
			line := fmt.Sprintf("%s  %s US-%d  %s%s", cursor, status, us.ID, style.Render(name), tag)
			b.WriteString(line + "\n")
		}
	}

	return b.String()
}

func (m sidebarModel) selected() *sidebarItem {
	if m.cursor >= 0 && m.cursor < len(m.items) {
		return &m.items[m.cursor]
	}
	return nil
}

func (m sidebarModel) selectedFeatureEntry() *sidebarEntry {
	sel := m.selected()
	if sel == nil {
		return nil
	}
	return &m.allFeats[sel.entryIdx]
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
