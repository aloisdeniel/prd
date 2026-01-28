package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/aloisdeniel/prd/prd"
)

type screen int

const (
	screenFeatureList screen = iota
	screenFeatureDetail
	screenUserStoryDetail
	screenFeatureForm
	screenUserStoryForm
	screenSearch
)

type errMsg struct {
	err error
}

func (e errMsg) Error() string {
	if e.err != nil {
		return e.err.Error()
	}
	return ""
}

type screenState struct {
	screen screen
}

type Model struct {
	stack          []screenState
	featureList    featureListModel
	featureDetail  featureDetailModel
	storyDetail    userStoryDetailModel
	featureForm    featureFormModel
	storyForm      userStoryFormModel
	search         searchModel
	sidebar        sidebarModel
	help           help.Model
	wideMode       bool
	width          int
	height         int
	basePath       string
	err            error
	invalidErrors  []string // Errors for currently selected invalid entry
	invalidName    string   // Name of currently selected invalid entry
	invalidPath    string   // Path of currently selected invalid entry
}

func NewModel(basePath string) Model {
	h := help.New()
	h.ShowAll = false
	return Model{
		basePath:    basePath,
		featureList: newFeatureListModel(basePath),
		sidebar:     newSidebarModel(basePath),
		help:        h,
		stack:       []screenState{{screen: screenFeatureList}},
	}
}

func (m Model) currentScreen() screen {
	if len(m.stack) == 0 {
		return screenFeatureList
	}
	return m.stack[len(m.stack)-1].screen
}

func (m *Model) push(s screen) {
	m.stack = append(m.stack, screenState{screen: s})
}

func (m *Model) pop() {
	if len(m.stack) > 1 {
		m.stack = m.stack[:len(m.stack)-1]
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.featureList.Init(), m.sidebar.Init(), watchFiles(m.basePath))
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.wideMode = true
		m.sidebar.SetSize(sidebarWidth, msg.Height)
		// Update viewport sizes for detail models
		m.featureDetail.SetSize(m.detailPanelWidth(), m.detailPanelHeight())
		m.storyDetail.SetSize(m.detailPanelWidth(), m.detailPanelHeight())
		return m, nil
	case fileChangedMsg:
		// Reload everything and restart the watcher for the next change.
		return m, tea.Batch(
			m.featureList.loadEntries,
			m.sidebar.loadAll,
			watchFiles(m.basePath),
		)
	}

	// In wide mode with no overlay, route to sidebar-based update
	if m.wideMode && !m.isOverlayScreen() {
		return m.updateWide(msg)
	}

	cur := m.currentScreen()

	// Handle form screens separately (they consume all key input)
	if cur == screenFeatureForm {
		return m.updateFeatureForm(msg)
	}
	if cur == screenUserStoryForm {
		return m.updateStoryForm(msg)
	}
	if cur == screenSearch {
		return m.updateSearch(msg)
	}

	// Global key handling for non-form screens
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.Back):
			if len(m.stack) > 1 {
				m.pop()
				// Reload data when going back
				switch m.currentScreen() {
				case screenFeatureList:
					return m, m.featureList.loadEntries
				case screenFeatureDetail:
					return m, m.featureDetail.loadFeature
				}
				return m, nil
			}
			return m, tea.Quit
		}
	}

	switch cur {
	case screenFeatureList:
		return m.updateFeatureList(msg)
	case screenFeatureDetail:
		return m.updateFeatureDetail(msg)
	case screenUserStoryDetail:
		return m.updateStoryDetail(msg)
	}

	return m, nil
}

func (m Model) isOverlayScreen() bool {
	cur := m.currentScreen()
	return cur == screenFeatureForm || cur == screenUserStoryForm || cur == screenSearch
}

func (m Model) updateWide(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Pass sidebar-loaded messages to sidebar
	if _, ok := msg.(sidebarLoadedMsg); ok {
		var cmd tea.Cmd
		m.sidebar, cmd = m.sidebar.Update(msg)
		m.syncDetailFromSidebar()
		return m, cmd
	}

	// Pass feature/story loaded messages to the detail models
	switch msg.(type) {
	case featureLoadedMsg:
		var cmd tea.Cmd
		m.featureDetail, cmd = m.featureDetail.Update(msg)
		return m, cmd
	case storyReloadedMsg:
		var cmd tea.Cmd
		m.storyDetail, cmd = m.storyDetail.Update(msg)
		return m, cmd
	case featureListMsg:
		var cmd tea.Cmd
		m.featureList, cmd = m.featureList.Update(msg)
		return m, cmd
	}

	if msg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.NewFeature):
			m.featureForm = newFeatureFormModel(m.basePath)
			m.push(screenFeatureForm)
			return m, m.featureForm.inputs[0].Focus()
		case key.Matches(msg, keys.New), key.Matches(msg, keys.NewStory):
			return m.handleWideNewStory()
		case key.Matches(msg, keys.Delete):
			return m.handleWideDelete()
		case key.Matches(msg, keys.Search):
			entries := m.featureList.entries
			// Also try to build entries from sidebar data
			if len(entries) == 0 {
				for _, fe := range m.sidebar.allFeats {
					entries = append(entries, fe.entry)
				}
			}
			m.search = newSearchModel(m.basePath, entries)
			m.push(screenSearch)
			return m, m.search.input.Focus()
		case key.Matches(msg, keys.Complete):
			return m.handleWideComplete(msg)
		case key.Matches(msg, keys.ScrollUp), key.Matches(msg, keys.ScrollDown):
			// Forward scroll keys to the detail panel
			cur := m.currentScreen()
			if cur == screenFeatureDetail {
				var cmd tea.Cmd
				m.featureDetail, cmd = m.featureDetail.Update(msg)
				return m, cmd
			} else if cur == screenUserStoryDetail {
				var cmd tea.Cmd
				m.storyDetail, cmd = m.storyDetail.Update(msg)
				return m, cmd
			}
			return m, nil
		case key.Matches(msg, keys.Enter):
			// If selected item is invalid, do nothing (errors already shown)
			sel := m.sidebar.selected()
			if sel != nil && sel.invalid {
				return m, nil
			}
			// If selected item is a story, open it in detail pane
			if sel != nil && sel.isStory {
				fe := m.sidebar.selectedFeatureEntry()
				us := sel.feature.UserStories[sel.storyIdx]
				completed := sel.progress[us.ID]
				m.storyDetail = newUserStoryDetailModel(m.basePath, sel.featureID, fe.path, &us, completed)
				m.storyDetail.SetSize(m.detailPanelWidth(), m.detailPanelHeight())
				// Set stack to story detail so View knows what to render
				m.stack = []screenState{{screen: screenUserStoryDetail}}
				return m, nil
			}
			// For features, let sidebar handle expand/collapse
			var cmd tea.Cmd
			m.sidebar, cmd = m.sidebar.Update(msg)
			m.syncDetailFromSidebar()
			return m, cmd
		case key.Matches(msg, keys.Back):
			// In wide mode, back from story detail goes to feature detail
			cur := m.currentScreen()
			if cur == screenUserStoryDetail {
				m.stack = []screenState{{screen: screenFeatureDetail}}
				m.syncDetailFromSidebar()
				return m, nil
			}
			return m, tea.Quit
		default:
			// Navigation keys go to sidebar
			var cmd tea.Cmd
			oldCursor := m.sidebar.cursorPos()
			m.sidebar, cmd = m.sidebar.Update(msg)
			if m.sidebar.cursorPos() != oldCursor {
				m.syncDetailFromSidebar()
			}
			return m, cmd
		}
	}

	return m, nil
}

func (m *Model) selectInSidebar(featureID string, isStory bool, storyID int) {
	for i, fe := range m.sidebar.allFeats {
		if fe.entry.ID == featureID {
			m.sidebar.allFeats[i].expanded = true
			m.sidebar.rebuildItems()
			for j, item := range m.sidebar.items {
				if isStory {
					if item.isStory && item.featureID == featureID && item.feature.UserStories[item.storyIdx].ID == storyID {
						m.sidebar.cursor = j
						return
					}
				} else {
					if !item.isStory && item.featureID == featureID {
						m.sidebar.cursor = j
						return
					}
				}
			}
			return
		}
	}
}

func (m *Model) detailPanelHeight() int {
	// Account for title line, spacing, and status bar in viewWide
	return m.height - 6
}

func (m *Model) detailPanelWidth() int {
	return max(m.width-sidebarWidth-3, 10)
}

func (m *Model) syncDetailFromSidebar() {
	sel := m.sidebar.selected()
	if sel == nil {
		return
	}

	// Handle invalid entries - show errors in detail pane
	if sel.invalid {
		fe := m.sidebar.selectedFeatureEntry()
		m.invalidErrors = sel.errors
		if fe != nil {
			m.invalidName = fe.entry.Name
			m.invalidPath = fe.path
		}
		m.stack = []screenState{{screen: screenFeatureList}} // Use a neutral screen state
		return
	}

	// Clear invalid state when selecting valid entry
	m.invalidErrors = nil
	m.invalidName = ""
	m.invalidPath = ""

	if sel.isStory {
		// Don't auto-switch to story detail on cursor move; only on Enter
		// Show the parent feature detail instead
		fe := m.sidebar.selectedFeatureEntry()
		if fe != nil {
			m.featureDetail = newFeatureDetailModel(m.basePath, fe.entry.ID)
			m.featureDetail.feature = fe.feature
			m.featureDetail.path = fe.path
			m.featureDetail.progress = fe.progress
			m.featureDetail.cursor = sel.storyIdx
			m.featureDetail.SetSize(m.detailPanelWidth(), m.detailPanelHeight())
			m.stack = []screenState{{screen: screenFeatureDetail}}
		}
	} else {
		fe := m.sidebar.selectedFeatureEntry()
		if fe != nil {
			m.featureDetail = newFeatureDetailModel(m.basePath, fe.entry.ID)
			m.featureDetail.feature = fe.feature
			m.featureDetail.path = fe.path
			m.featureDetail.progress = fe.progress
			m.featureDetail.SetSize(m.detailPanelWidth(), m.detailPanelHeight())
			m.stack = []screenState{{screen: screenFeatureDetail}}
		}
	}
}

func (m Model) handleWideNewStory() (tea.Model, tea.Cmd) {
	fe := m.sidebar.selectedFeatureEntry()
	if fe != nil && fe.path != "" && !fe.invalid {
		m.storyForm = newUserStoryFormModel(fe.path)
		m.push(screenUserStoryForm)
		return m, m.storyForm.inputs[0].Focus()
	}
	return m, nil
}

func (m Model) handleWideDelete() (tea.Model, tea.Cmd) {
	sel := m.sidebar.selected()
	if sel == nil {
		return m, nil
	}
	if sel.isStory {
		fe := m.sidebar.selectedFeatureEntry()
		if fe != nil {
			us := sel.feature.UserStories[sel.storyIdx]
			if err := prd.DeleteUserStory(fe.path, us.ID); err != nil {
				m.err = err
				return m, nil
			}
			return m, m.sidebar.loadAll
		}
	}
	return m, nil
}

func (m Model) handleWideComplete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If viewing a story detail, toggle via the detail model
	if m.currentScreen() == screenUserStoryDetail {
		var cmd tea.Cmd
		m.storyDetail, cmd = m.storyDetail.Update(msg)
		return m, cmd
	}
	// Toggle completion for the story under the sidebar cursor
	sel := m.sidebar.selected()
	if sel != nil && sel.isStory {
		fe := m.sidebar.selectedFeatureEntry()
		if fe != nil {
			us := sel.feature.UserStories[sel.storyIdx]
			if err := prd.CompleteUserStory(fe.path, us.ID); err != nil {
				m.err = err
				return m, nil
			}
			return m, m.sidebar.loadAll
		}
	}
	return m, nil
}

func (m Model) updateFeatureList(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.featureList, cmd = m.featureList.Update(msg)

	if msg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(msg, keys.Enter):
			if e := m.featureList.selectedEntry(); e != nil {
				m.featureDetail = newFeatureDetailModel(m.basePath, e.ID)
				m.push(screenFeatureDetail)
				return m, m.featureDetail.Init()
			}
		case key.Matches(msg, keys.New), key.Matches(msg, keys.NewFeature):
			m.featureForm = newFeatureFormModel(m.basePath)
			m.push(screenFeatureForm)
			return m, m.featureForm.inputs[0].Focus()
		case key.Matches(msg, keys.Search):
			m.search = newSearchModel(m.basePath, m.featureList.entries)
			m.push(screenSearch)
			return m, m.search.input.Focus()
		}
	}

	return m, cmd
}

func (m Model) updateFeatureDetail(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.featureDetail, cmd = m.featureDetail.Update(msg)

	if msg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(msg, keys.Enter):
			if s := m.featureDetail.selectedStory(); s != nil {
				completed := m.featureDetail.progress[s.ID]
				m.storyDetail = newUserStoryDetailModel(m.basePath, m.featureDetail.id, m.featureDetail.path, s, completed)
				m.storyDetail.SetSize(m.detailPanelWidth(), m.detailPanelHeight())
				m.push(screenUserStoryDetail)
				return m, nil
			}
		case key.Matches(msg, keys.New), key.Matches(msg, keys.NewStory):
			if m.featureDetail.path != "" {
				m.storyForm = newUserStoryFormModel(m.featureDetail.path)
				m.push(screenUserStoryForm)
				return m, m.storyForm.inputs[0].Focus()
			}
		case key.Matches(msg, keys.Delete):
			if s := m.featureDetail.selectedStory(); s != nil {
				if err := prd.DeleteUserStory(m.featureDetail.path, s.ID); err != nil {
					m.err = err
					return m, nil
				}
				return m, m.featureDetail.loadFeature
			}
		}
	}

	return m, cmd
}

func (m Model) updateStoryDetail(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.storyDetail, cmd = m.storyDetail.Update(msg)
	return m, cmd
}

func (m Model) updateFeatureForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		if key.Matches(msg, keys.Back) {
			m.pop()
			return m, nil
		}
	}

	switch msg.(type) {
	case featureCreatedMsg:
		m.pop()
		if m.wideMode {
			return m, m.sidebar.loadAll
		}
		return m, m.featureList.loadEntries
	}

	var cmd tea.Cmd
	m.featureForm, cmd = m.featureForm.Update(msg)
	return m, cmd
}

func (m Model) updateStoryForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		if key.Matches(msg, keys.Back) {
			m.pop()
			return m, nil
		}
	}

	switch msg.(type) {
	case storyCreatedMsg:
		m.pop()
		if m.wideMode {
			return m, m.sidebar.loadAll
		}
		return m, m.featureDetail.loadFeature
	}

	var cmd tea.Cmd
	m.storyForm, cmd = m.storyForm.Update(msg)
	return m, cmd
}

func (m Model) updateSearch(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(msg, keys.Back):
			m.pop()
			return m, nil
		case key.Matches(msg, keys.Enter):
			if r := m.search.selectedResult(); r != nil {
				if m.wideMode {
					m.pop() // remove search overlay
					m.selectInSidebar(r.featureID, r.isStory, r.storyID)
					m.syncDetailFromSidebar()
					return m, nil
				}
				if r.isStory {
					m.featureDetail = newFeatureDetailModel(m.basePath, r.featureID)
					m.stack = []screenState{{screen: screenFeatureList}}
					m.push(screenFeatureDetail)
					return m, m.featureDetail.Init()
				}
				m.featureDetail = newFeatureDetailModel(m.basePath, r.featureID)
				m.stack = []screenState{{screen: screenFeatureList}}
				m.push(screenFeatureDetail)
				return m, m.featureDetail.Init()
			}
		}
	}

	var cmd tea.Cmd
	m.search, cmd = m.search.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	var content string

	// Handle overlay screens (forms, search)
	if m.isOverlayScreen() {
		switch m.currentScreen() {
		case screenFeatureForm:
			content = m.featureForm.View()
		case screenUserStoryForm:
			content = m.storyForm.View()
		case screenSearch:
			content = m.search.View()
		}
	} else {
		content = m.viewWide()
	}

	// Status bar
	statusHelp := m.currentStatusHelp()
	statusBar := statusBarStyle.Width(m.width).Render(statusHelp)

	if m.err != nil {
		content += "\n" + errorStyle.Render(m.err.Error())
	}

	return lipgloss.JoinVertical(lipgloss.Left, content, statusBar)
}

func (m Model) viewWide() string {
	if m.height < 2 || m.width < sidebarWidth+10 {
		return "Window too small"
	}

	sidebarContent := m.sidebar.View()
	detailWidth := max(m.width-sidebarWidth-1, 10)

	var detailContent string

	// Check if we're showing an invalid entry
	if len(m.invalidErrors) > 0 {
		detailContent = m.viewInvalidDetail()
	} else {
		switch m.currentScreen() {
		case screenFeatureDetail:
			detailContent = sidebarTitleStyle.Render("Feature") + "\n\n" + m.featureDetail.View()
		case screenUserStoryDetail:
			detailContent = sidebarTitleStyle.Render("User Story") + "\n\n" + m.storyDetail.View()
		default:
			detailContent = dimStyle.PaddingLeft(2).Render("Select a feature to view details.")
		}
	}

	contentHeight := m.height - 2 // Leave room for status bar

	sidebarPane := lipgloss.NewStyle().
		Width(sidebarWidth).
		MaxHeight(contentHeight).
		Render(sidebarContent)

	separator := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render(strings.Repeat("│\n", contentHeight-1) + "│")

	detailPane := lipgloss.NewStyle().
		Width(detailWidth).
		MaxHeight(contentHeight).
		Render(detailContent)

	return lipgloss.JoinHorizontal(lipgloss.Top, sidebarPane, separator, detailPane)
}

func (m Model) viewInvalidDetail() string {
	var b strings.Builder

	b.WriteString(sidebarTitleStyle.Render("Invalid Document"))
	b.WriteString("\n\n")

	if m.invalidName != "" {
		b.WriteString(titleStyle.Render(m.invalidName))
		b.WriteString("\n")
	}

	if m.invalidPath != "" {
		b.WriteString(dimStyle.Render(m.invalidPath))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(sectionStyle.Render("Errors"))
	b.WriteString("\n\n")

	for _, err := range m.invalidErrors {
		b.WriteString(errorStyle.Render("• " + err))
		b.WriteString("\n")
	}

	return b.String()
}

func (m Model) currentStatusHelp() string {
	var keyMap contextualKeyMap
	if m.wideMode && !m.isOverlayScreen() {
		keyMap = wideKeys()
	} else {
		switch m.currentScreen() {
		case screenFeatureList:
			keyMap = featureListKeys()
		case screenFeatureDetail:
			keyMap = featureDetailKeys()
		case screenUserStoryDetail:
			keyMap = storyDetailKeys()
		case screenFeatureForm:
			keyMap = featureFormKeys()
		case screenUserStoryForm:
			keyMap = storyFormKeys()
		case screenSearch:
			keyMap = searchKeys()
		default:
			return ""
		}
	}
	return m.help.View(keyMap)
}

func Run(basePath string) error {
	p := tea.NewProgram(NewModel(basePath), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
