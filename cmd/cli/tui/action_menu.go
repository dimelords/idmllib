package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ActionMenu is a reusable component for showing contextual action options
type ActionMenu struct {
	title    string
	options  []string
	cursor   int
	selected int
	quitting bool
}

// NewActionMenu creates a new action menu with a title and options
func NewActionMenu(title string, options []string) *ActionMenu {
	return &ActionMenu{
		title:    title,
		options:  options,
		selected: -1,
	}
}

func (m *ActionMenu) Init() tea.Cmd {
	return nil
}

func (m *ActionMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "q", "esc":
			m.quitting = true
			return m, nil

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}

		case "enter", " ":
			m.selected = m.cursor
			return m, nil

		// Quick number selection (1-9)
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			idx := int(msg.String()[0] - '1')
			if idx >= 0 && idx < len(m.options) {
				m.cursor = idx
				m.selected = idx
				return m, nil
			}
		}
	}

	return m, nil
}

func (m *ActionMenu) View() string {
	var s strings.Builder

	// Title
	s.WriteString(TitleStyle.Render(m.title))
	s.WriteString("\n\n")

	// Options with checkbox style
	for i, option := range m.options {
		// Checkbox indicator
		checkbox := CheckboxEmpty
		style := UnselectedStyle

		if m.cursor == i {
			checkbox = CheckboxSelected
			style = SelectedStyle
		}

		// Build line: [x] 1. Option
		line := fmt.Sprintf("%s %d. %s", checkbox, i+1, option)
		s.WriteString(style.Render(line))
		s.WriteString("\n")
	}

	// Help text
	s.WriteString(FormatHelp(
		"j/k, up/down: select",
		"enter: choose",
		"1-9: quick select",
		"q/esc: go back",
	))

	return s.String()
}

// GetSelected returns the selected option index, or -1 if none selected
func (m *ActionMenu) GetSelected() int {
	return m.selected
}

// IsQuitting returns true if the user wants to quit this menu
func (m *ActionMenu) IsQuitting() bool {
	return m.quitting
}
