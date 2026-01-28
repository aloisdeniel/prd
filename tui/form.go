package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/aloisdeniel/prd/prd"
)

// featureFormModel handles creating a new feature.
type featureFormModel struct {
	inputs      []textinput.Model
	focused     int
	allSections bool
	basePath    string
	err         error
	done        bool
}

const featFormFieldCount = 3 // name, desc, allSections toggle

func newFeatureFormModel(basePath string) featureFormModel {
	nameInput := textinput.New()
	nameInput.Placeholder = "Feature name"
	nameInput.Focus()
	nameInput.CharLimit = 100

	descInput := textinput.New()
	descInput.Placeholder = "Description (optional)"
	descInput.CharLimit = 500

	return featureFormModel{
		inputs:      []textinput.Model{nameInput, descInput},
		allSections: true,
		basePath:    basePath,
	}
}

type featureCreatedMsg struct{}

func (m featureFormModel) Update(msg tea.Msg) (featureFormModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Toggle field (focused == 2)
		if m.focused == 2 {
			switch {
			case key.Matches(msg, keys.Space), key.Matches(msg, keys.Left), key.Matches(msg, keys.Right):
				m.allSections = !m.allSections
				return m, nil
			}
		}

		switch {
		case key.Matches(msg, keys.Tab):
			if m.focused < len(m.inputs) {
				m.inputs[m.focused].Blur()
			}
			m.focused = (m.focused + 1) % featFormFieldCount
			if m.focused < len(m.inputs) {
				m.inputs[m.focused].Focus()
			}
			return m, nil
		case key.Matches(msg, keys.Enter):
			if m.focused < featFormFieldCount-1 {
				if m.focused < len(m.inputs) {
					m.inputs[m.focused].Blur()
				}
				m.focused++
				if m.focused < len(m.inputs) {
					m.inputs[m.focused].Focus()
				}
				return m, nil
			}
			// Submit
			name := strings.TrimSpace(m.inputs[0].Value())
			if name == "" {
				m.err = errEmpty
				return m, nil
			}
			desc := strings.TrimSpace(m.inputs[1].Value())
			_, err := prd.CreateFeature(m.basePath, name, desc, m.allSections)
			if err != nil {
				m.err = err
				return m, nil
			}
			m.done = true
			return m, func() tea.Msg { return featureCreatedMsg{} }
		}
	}

	if m.focused < len(m.inputs) {
		var cmd tea.Cmd
		m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m featureFormModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("New Feature"))
	b.WriteString("\n\n")

	b.WriteString(inputLabelStyle.Render("Name"))
	b.WriteString("\n")
	b.WriteString("  " + m.inputs[0].View())
	b.WriteString("\n\n")

	b.WriteString(inputLabelStyle.Render("Description"))
	b.WriteString("\n")
	b.WriteString("  " + m.inputs[1].View())
	b.WriteString("\n\n")

	b.WriteString(inputLabelStyle.Render("All sections"))
	b.WriteString("\n  ")
	if m.allSections {
		b.WriteString(selectedStyle.Render("[x] Include all sections"))
	} else {
		b.WriteString(dimStyle.Render("[ ] Required sections only"))
	}
	if m.focused == 2 {
		b.WriteString(dimStyle.Render("  space to toggle"))
	}
	b.WriteString("\n")

	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render(m.err.Error()))
	}

	return b.String()
}

func (m featureFormModel) statusHelp() string {
	return dimStyle.Render("tab next field  space toggle  enter submit  esc cancel")
}

// userStoryFormModel handles creating a new user story.
type userStoryFormModel struct {
	inputs   []textinput.Model
	focused  int
	priority int // 1-5
	prdPath  string
	err      error
	done     bool
}

func newUserStoryFormModel(prdPath string) userStoryFormModel {
	nameInput := textinput.New()
	nameInput.Placeholder = "User story name"
	nameInput.Focus()
	nameInput.CharLimit = 100

	descInput := textinput.New()
	descInput.Placeholder = "Description (optional)"
	descInput.CharLimit = 500

	acInput := textinput.New()
	acInput.Placeholder = "Acceptance criteria (comma-separated)"
	acInput.CharLimit = 500

	return userStoryFormModel{
		inputs:   []textinput.Model{nameInput, descInput, acInput},
		priority: 3,
		prdPath:  prdPath,
	}
}

type storyCreatedMsg struct{}

const storyFormFieldCount = 4 // name, desc, ac, priority

func (m userStoryFormModel) Update(msg tea.Msg) (userStoryFormModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Priority picker (focused == 3)
		if m.focused == 3 {
			switch {
			case key.Matches(msg, keys.Left):
				if m.priority > 1 {
					m.priority--
				}
				return m, nil
			case key.Matches(msg, keys.Right):
				if m.priority < 5 {
					m.priority++
				}
				return m, nil
			}
		}

		switch {
		case key.Matches(msg, keys.Tab):
			if m.focused < len(m.inputs) {
				m.inputs[m.focused].Blur()
			}
			m.focused = (m.focused + 1) % storyFormFieldCount
			if m.focused < len(m.inputs) {
				m.inputs[m.focused].Focus()
			}
			return m, nil
		case key.Matches(msg, keys.Enter):
			if m.focused < storyFormFieldCount-1 {
				if m.focused < len(m.inputs) {
					m.inputs[m.focused].Blur()
				}
				m.focused++
				if m.focused < len(m.inputs) {
					m.inputs[m.focused].Focus()
				}
				return m, nil
			}
			// Submit
			name := strings.TrimSpace(m.inputs[0].Value())
			if name == "" {
				m.err = errEmpty
				return m, nil
			}
			desc := strings.TrimSpace(m.inputs[1].Value())
			acText := strings.TrimSpace(m.inputs[2].Value())
			var criteria []string
			if acText != "" {
				for _, c := range strings.Split(acText, ",") {
					c = strings.TrimSpace(c)
					if c != "" {
						criteria = append(criteria, c)
					}
				}
			}
			err := prd.CreateUserStory(m.prdPath, name, desc, criteria, "", m.priority)
			if err != nil {
				m.err = err
				return m, nil
			}
			m.done = true
			return m, func() tea.Msg { return storyCreatedMsg{} }
		}
	}

	if m.focused < len(m.inputs) {
		var cmd tea.Cmd
		m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m userStoryFormModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("New User Story"))
	b.WriteString("\n\n")

	b.WriteString(inputLabelStyle.Render("Name"))
	b.WriteString("\n")
	b.WriteString("  " + m.inputs[0].View())
	b.WriteString("\n\n")

	b.WriteString(inputLabelStyle.Render("Description"))
	b.WriteString("\n")
	b.WriteString("  " + m.inputs[1].View())
	b.WriteString("\n\n")

	b.WriteString(inputLabelStyle.Render("Acceptance Criteria"))
	b.WriteString("\n")
	b.WriteString("  " + m.inputs[2].View())
	b.WriteString("\n\n")

	b.WriteString(inputLabelStyle.Render("Priority"))
	b.WriteString("\n  ")
	for p := 1; p <= 5; p++ {
		label := fmt.Sprintf(" P%d ", p)
		if p == m.priority {
			b.WriteString(selectedStyle.Render(label))
		} else {
			b.WriteString(dimStyle.Render(label))
		}
	}
	if m.focused == 3 {
		b.WriteString(dimStyle.Render("  ←/→ to change"))
	}
	b.WriteString("\n")

	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render(m.err.Error()))
	}

	return b.String()
}

func (m userStoryFormModel) statusHelp() string {
	return dimStyle.Render("tab next field  ←/→ priority  enter submit  esc cancel")
}

var errEmpty = fmt.Errorf("name cannot be empty")
