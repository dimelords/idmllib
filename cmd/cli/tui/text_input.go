package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// TextInput is a simple text input model
type TextInput struct {
	prompt      string
	placeholder string
	value       string
	// cursor field removed as it was unused
	submitted bool
	cancelled bool
}

// NewTextInput creates a new text input
func NewTextInput(prompt, placeholder string) *TextInput {
	return &TextInput{
		prompt:      prompt,
		placeholder: placeholder,
	}
}

func (m *TextInput) Init() tea.Cmd {
	return nil
}

func (m *TextInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			return m, tea.Quit

		case "enter":
			m.submitted = true
			return m, tea.Quit

		case "backspace":
			if len(m.value) > 0 {
				m.value = m.value[:len(m.value)-1]
			}

		default:
			// Add character
			m.value += msg.String()
		}
	}

	return m, nil
}

func (m *TextInput) View() string {
	var s strings.Builder

	s.WriteString(InputLabelStyle.Render(m.prompt))
	s.WriteString("\n")

	displayValue := m.value
	if displayValue == "" {
		displayValue = SubtitleStyle.Render(m.placeholder)
	} else {
		displayValue = InputStyle.Render(displayValue + "█")
	}

	s.WriteString(displayValue)
	s.WriteString("\n\n")
	s.WriteString(HelpStyle.Render("enter to confirm • esc to cancel"))

	return s.String()
}

// GetValue returns the input value
func (m *TextInput) GetValue() string {
	return m.value
}

// WasCancelled returns true if the input was cancelled
func (m *TextInput) WasCancelled() bool {
	return m.cancelled
}
