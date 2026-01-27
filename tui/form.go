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
	inputs   []textinput.Model
	focused  int
	basePath string
	err      error
	done     bool
}

func newFeatureFormModel(basePath string) featureFormModel {
	nameInput := textinput.New()
	nameInput.Placeholder = "Feature name"
	nameInput.Focus()
	nameInput.CharLimit = 100

	descInput := textinput.New()
	descInput.Placeholder = "Description (optional)"
	descInput.CharLimit = 500

	return featureFormModel{
		inputs:   []textinput.Model{nameInput, descInput},
		basePath: basePath,
	}
}

type featureCreatedMsg struct{}

func (m featureFormModel) Update(msg tea.Msg) (featureFormModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Tab):
			m.inputs[m.focused].Blur()
			m.focused = (m.focused + 1) % len(m.inputs)
			m.inputs[m.focused].Focus()
			return m, nil
		case key.Matches(msg, keys.Enter):
			if m.focused < len(m.inputs)-1 {
				m.inputs[m.focused].Blur()
				m.focused++
				m.inputs[m.focused].Focus()
				return m, nil
			}
			// Submit
			name := strings.TrimSpace(m.inputs[0].Value())
			if name == "" {
				m.err = errEmpty
				return m, nil
			}
			desc := strings.TrimSpace(m.inputs[1].Value())
			_, err := prd.CreateFeature(m.basePath, name, desc)
			if err != nil {
				m.err = err
				return m, nil
			}
			m.done = true
			return m, func() tea.Msg { return featureCreatedMsg{} }
		}
	}

	var cmd tea.Cmd
	m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
	return m, cmd
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
	b.WriteString("\n")

	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render(m.err.Error()))
	}

	return b.String()
}

func (m featureFormModel) statusHelp() string {
	return dimStyle.Render("tab next field  enter submit  esc cancel")
}

// userStoryFormModel handles creating a new user story.
type userStoryFormModel struct {
	inputs   []textinput.Model
	focused  int
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
		inputs:  []textinput.Model{nameInput, descInput, acInput},
		prdPath: prdPath,
	}
}

type storyCreatedMsg struct{}

func (m userStoryFormModel) Update(msg tea.Msg) (userStoryFormModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Tab):
			m.inputs[m.focused].Blur()
			m.focused = (m.focused + 1) % len(m.inputs)
			m.inputs[m.focused].Focus()
			return m, nil
		case key.Matches(msg, keys.Enter):
			if m.focused < len(m.inputs)-1 {
				m.inputs[m.focused].Blur()
				m.focused++
				m.inputs[m.focused].Focus()
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
			err := prd.CreateUserStory(m.prdPath, name, desc, criteria, "")
			if err != nil {
				m.err = err
				return m, nil
			}
			m.done = true
			return m, func() tea.Msg { return storyCreatedMsg{} }
		}
	}

	var cmd tea.Cmd
	m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
	return m, cmd
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
	b.WriteString("\n")

	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render(m.err.Error()))
	}

	return b.String()
}

func (m userStoryFormModel) statusHelp() string {
	return dimStyle.Render("tab next field  enter submit  esc cancel")
}

var errEmpty = fmt.Errorf("name cannot be empty")
