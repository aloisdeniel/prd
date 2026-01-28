package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
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
	entryIdx  int      // index into allFeats
	invalid   bool     // True if the document has errors
	errors    []string // Parse or validation errors
}

// Implement list.Item interface
func (i sidebarItem) FilterValue() string {
	if i.isStory {
		return i.feature.UserStories[i.storyIdx].Name
	}
	return i.featureID
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
	list     list.Model
	allFeats []sidebarEntry
	width    int
	height   int
	err      error
}

// sidebarDelegate handles custom rendering for sidebar items
type sidebarDelegate struct {
	allFeats *[]sidebarEntry
}

func (d sidebarDelegate) Height() int                             { return 1 }
func (d sidebarDelegate) Spacing() int                            { return 0 }
func (d sidebarDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d sidebarDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(sidebarItem)
	if !ok {
		return
	}

	cursor := "  "
	style := normalStyle
	if index == m.Index() {
		cursor = selectedStyle.Render("> ")
		style = selectedStyle
	}

	var line string
	if !item.isStory {
		// Feature row
		fe := (*d.allFeats)[item.entryIdx]

		if item.invalid {
			// Invalid document - show with INVALID tag
			name := truncate(fe.entry.Name, sidebarWidth-20)
			line = fmt.Sprintf("%s  %s  %s %s", cursor, style.Render(fe.entry.ID), errorStyle.Render(name), invalidTag)
		} else {
			// Valid document
			arrow := "▸"
			if item.expanded {
				arrow = "▾"
			}
			progress := fmt.Sprintf("[%d/%d]", fe.entry.Completed, fe.entry.Total)
			pStyle := incompleteStyle
			if fe.entry.Completed == fe.entry.Total && fe.entry.Total > 0 {
				pStyle = completedStyle
			}
			name := truncate(fe.entry.Name, sidebarWidth-16)
			line = fmt.Sprintf("%s%s %s  %s %s", cursor, arrow, style.Render(fe.entry.ID), style.Render(name), pStyle.Render(progress))
		}
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
		line = fmt.Sprintf("%s  %s US-%d  %s%s", cursor, status, us.ID, style.Render(name), tag)
	}

	fmt.Fprint(w, line)
}

type sidebarLoadedMsg struct {
	entries []sidebarEntry
}

func newSidebarModel(basePath string) sidebarModel {
	// Create list with empty items initially
	delegate := sidebarDelegate{}
	l := list.New([]list.Item{}, delegate, sidebarWidth, 10)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowFilter(false)
	l.SetShowHelp(false)
	l.SetShowPagination(false)
	l.Styles.NoItems = dimStyle
	l.SetFilteringEnabled(false)

	m := sidebarModel{
		basePath: basePath,
		list:     l,
		width:    sidebarWidth,
	}
	// Update delegate with reference to allFeats
	delegate.allFeats = &m.allFeats
	m.list.SetDelegate(delegate)

	return m
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
	var items []list.Item
	for i := range m.allFeats {
		fe := &m.allFeats[i]
		items = append(items, sidebarItem{
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
				items = append(items, sidebarItem{
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

	// Preserve cursor position
	oldIndex := m.list.Index()
	m.list.SetItems(items)
	if oldIndex >= len(items) {
		oldIndex = max(0, len(items)-1)
	}
	m.list.Select(oldIndex)
}

func (m *sidebarModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	// Account for title and padding in view
	listHeight := height - 5
	if listHeight < 1 {
		listHeight = 1
	}
	m.list.SetSize(width, listHeight)
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
		// Update delegate reference after allFeats changes
		delegate := sidebarDelegate{allFeats: &m.allFeats}
		m.list.SetDelegate(delegate)
		m.rebuildItems()
	case errMsg:
		m.err = msg.err
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Up), key.Matches(msg, keys.Down):
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		case key.Matches(msg, keys.Space), key.Matches(msg, keys.Enter):
			if sel := m.selected(); sel != nil && !sel.isStory && !sel.invalid {
				idx := sel.entryIdx
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

	if len(m.list.Items()) == 0 {
		b.WriteString(dimStyle.Render(" No features."))
		return b.String()
	}

	b.WriteString(m.list.View())
	return b.String()
}

func (m sidebarModel) selected() *sidebarItem {
	if item := m.list.SelectedItem(); item != nil {
		if si, ok := item.(sidebarItem); ok {
			return &si
		}
	}
	return nil
}

// cursor returns the current cursor position for compatibility with tui.go
func (m sidebarModel) cursor() int {
	return m.list.Index()
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
