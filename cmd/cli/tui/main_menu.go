package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// MenuItem represents a menu option
type MenuItem struct {
	Title       string
	Description string
	Key         string
}

// MainMenu is the main menu model
type MainMenu struct {
	items    []MenuItem
	cursor   int
	selected int
	quitting bool
}

// NewMainMenu creates a new main menu
func NewMainMenu() *MainMenu {
	return &MainMenu{
		items: []MenuItem{
			{
				Title:       "Create new document",
				Description: "Interactive document creation wizard",
				Key:         "1",
			},
			{
				Title:       "Roundtrip test",
				Description: "Read and write IDML files",
				Key:         "2",
			},
			{
				Title:       "Export IDMS snippet",
				Description: "Export textframes as IDMS snippets",
				Key:         "3",
			},
			{
				Title:       "Exit",
				Description: "Quit the application",
				Key:         "q",
			},
		},
		selected: -1,
	}
}

func (m *MainMenu) Init() tea.Cmd {
	return nil
}

func (m *MainMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.selected = len(m.items) - 1
			m.quitting = true
			return m, tea.Quit

		case "q":
			// 'q' always quits regardless of cursor position
			m.selected = len(m.items) - 1
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}

		case "enter", " ":
			m.selected = m.cursor
			return m, tea.Quit

		case "1", "2", "3":
			// Direct number selection
			idx := int(msg.String()[0] - '1')
			if idx >= 0 && idx < len(m.items)-1 {
				m.cursor = idx
				m.selected = idx
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (m *MainMenu) View() string {
	if m.quitting {
		return ""
	}

	var s strings.Builder

	// Title
	s.WriteString(TitleStyle.Render("What would you like to do?"))
	s.WriteString("\n\n")

	// Menu items with checkbox style
	for i, item := range m.items {
		// Checkbox indicator
		checkbox := CheckboxEmpty
		style := UnselectedStyle

		if m.cursor == i {
			checkbox = CheckboxSelected
			style = SelectedStyle
		}

		// Build line: [x] 1. Title
		line := fmt.Sprintf("%s %s. %s", checkbox, item.Key, item.Title)
		s.WriteString(style.Render(line))
		s.WriteString("\n")

		// Description (only for selected item)
		if m.cursor == i && item.Description != "" {
			s.WriteString(DescStyle.Render(item.Description))
			s.WriteString("\n")
		}
	}

	// Help text
	s.WriteString(FormatHelp(
		"j/k, up/down: select",
		"enter: choose",
		"1-3: quick select",
		"q: quit",
	))

	return s.String()
}

// GetSelected returns the selected menu index
func (m *MainMenu) GetSelected() int {
	return m.selected
}
