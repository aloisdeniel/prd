package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/aloisdeniel/prd/prd"
)

type searchResult struct {
	featureID   string
	featureName string
	storyID     int
	storyName   string
	isStory     bool
}

type searchModel struct {
	input    textinput.Model
	entries  []prd.FeatureEntry
	features map[string]*prd.Feature
	results  []searchResult
	cursor   int
	basePath string
	err      error
}

func newSearchModel(basePath string, entries []prd.FeatureEntry) searchModel {
	ti := textinput.New()
	ti.Placeholder = "Search features and user stories..."
	ti.Focus()
	ti.CharLimit = 100

	m := searchModel{
		input:    ti,
		entries:  entries,
		features: make(map[string]*prd.Feature),
		basePath: basePath,
	}

	// Preload all features
	for _, e := range entries {
		_, feat, err := prd.LoadFeature(basePath, e.ID)
		if err == nil {
			m.features[e.ID] = feat
		}
	}

	m.filterResults()
	return m
}

func (m *searchModel) filterResults() {
	query := strings.ToLower(strings.TrimSpace(m.input.Value()))
	m.results = nil

	for _, e := range m.entries {
		matchFeature := query == "" || strings.Contains(strings.ToLower(e.Name), query) || strings.Contains(e.ID, query)

		if matchFeature {
			m.results = append(m.results, searchResult{
				featureID:   e.ID,
				featureName: e.Name,
			})
		}

		if feat, ok := m.features[e.ID]; ok {
			for _, us := range feat.UserStories {
				matchStory := query == "" || strings.Contains(strings.ToLower(us.Name), query) ||
					strings.Contains(fmt.Sprintf("us-%d", us.ID), query)
				if matchStory && !matchFeature {
					m.results = append(m.results, searchResult{
						featureID:   e.ID,
						featureName: e.Name,
						storyID:     us.ID,
						storyName:   us.Name,
						isStory:     true,
					})
				} else if matchStory && matchFeature {
					m.results = append(m.results, searchResult{
						featureID:   e.ID,
						featureName: e.Name,
						storyID:     us.ID,
						storyName:   us.Name,
						isStory:     true,
					})
				}
			}
		}
	}

	if m.cursor >= len(m.results) {
		m.cursor = max(0, len(m.results)-1)
	}
}

func (m searchModel) Update(msg tea.Msg) (searchModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case "down":
			if m.cursor < len(m.results)-1 {
				m.cursor++
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m.filterResults()
	return m, cmd
}

func (m searchModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Search"))
	b.WriteString("\n\n")
	b.WriteString("  " + m.input.View())
	b.WriteString("\n\n")

	if len(m.results) == 0 {
		b.WriteString(dimStyle.PaddingLeft(2).Render("No results"))
		return b.String()
	}

	for i, r := range m.results {
		cursor := "  "
		style := normalStyle
		if i == m.cursor {
			cursor = selectedStyle.Render("> ")
			style = selectedStyle
		}

		if r.isStory {
			line := fmt.Sprintf("%s  %s US-%d  %s",
				cursor,
				dimStyle.Render(r.featureID),
				r.storyID,
				style.Render(r.storyName),
			)
			b.WriteString(line + "\n")
		} else {
			line := fmt.Sprintf("%s%s  %s",
				cursor,
				style.Render(r.featureID),
				style.Render(r.featureName),
			)
			b.WriteString(line + "\n")
		}
	}

	return b.String()
}

func (m searchModel) selectedResult() *searchResult {
	if m.cursor >= 0 && m.cursor < len(m.results) {
		return &m.results[m.cursor]
	}
	return nil
}
