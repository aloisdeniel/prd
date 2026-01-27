package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

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
	stack       []screenState
	featureList featureListModel
	featureDetail featureDetailModel
	storyDetail userStoryDetailModel
	featureForm featureFormModel
	storyForm   userStoryFormModel
	search      searchModel
	width       int
	height      int
	basePath    string
	err         error
}

func NewModel(basePath string) Model {
	return Model{
		basePath:    basePath,
		featureList: newFeatureListModel(basePath),
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
	return m.featureList.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
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
		case key.Matches(msg, keys.New):
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
				m.storyDetail = newUserStoryDetailModel(m.basePath, m.featureDetail.id, m.featureDetail.path, s)
				m.push(screenUserStoryDetail)
				return m, nil
			}
		case key.Matches(msg, keys.New):
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
				if r.isStory {
					m.featureDetail = newFeatureDetailModel(m.basePath, r.featureID)
					// Replace stack: list -> detail -> story
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

	switch m.currentScreen() {
	case screenFeatureList:
		content = m.featureList.View()
	case screenFeatureDetail:
		content = m.featureDetail.View()
	case screenUserStoryDetail:
		content = m.storyDetail.View()
	case screenFeatureForm:
		content = m.featureForm.View()
	case screenUserStoryForm:
		content = m.storyForm.View()
	case screenSearch:
		content = m.search.View()
	}

	// Status bar
	statusHelp := m.currentStatusHelp()
	statusBar := statusBarStyle.Width(m.width).Render(statusHelp)

	if m.err != nil {
		content += "\n" + errorStyle.Render(m.err.Error())
	}

	// Layout: content + status bar at bottom
	availHeight := m.height - 1
	contentLines := strings.Count(content, "\n")
	if contentLines < availHeight {
		content += strings.Repeat("\n", availHeight-contentLines)
	}

	return fmt.Sprintf("%s\n%s", content, statusBar)
}

func (m Model) currentStatusHelp() string {
	switch m.currentScreen() {
	case screenFeatureList:
		return m.featureList.statusHelp()
	case screenFeatureDetail:
		return m.featureDetail.statusHelp()
	case screenUserStoryDetail:
		return m.storyDetail.statusHelp()
	case screenFeatureForm:
		return m.featureForm.statusHelp()
	case screenUserStoryForm:
		return m.storyForm.statusHelp()
	case screenSearch:
		return m.search.statusHelp()
	}
	return ""
}

func Run(basePath string) error {
	p := tea.NewProgram(NewModel(basePath), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
