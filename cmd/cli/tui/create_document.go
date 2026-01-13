package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dimelords/idmllib/v2/pkg/idml"
)

type createStep int

const (
	stepPreset createStep = iota
	stepOrientation
	stepColumns
	stepGutter
	stepFilename
	stepCreating
	stepDone
)

// CreateDocumentWizard guides the user through document creation
type CreateDocumentWizard struct {
	step        createStep
	preset      idml.DocumentPreset
	orientation string
	columns     int
	gutter      float64
	filename    string
	cursor      int
	error       string
	success     bool
	pkg         *idml.Package
}

// NewCreateDocumentWizard creates a new wizard
func NewCreateDocumentWizard() *CreateDocumentWizard {
	return &CreateDocumentWizard{
		step:        stepPreset,
		preset:      idml.PresetLetterUS,
		orientation: "Portrait",
		columns:     1,
		gutter:      12.0,
		filename:    "output.idml",
	}
}

func (m *CreateDocumentWizard) Init() tea.Cmd {
	return nil
}

func (m *CreateDocumentWizard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.step == stepDone {
				return m, tea.Quit
			}
			m.error = "Cancelled by user"
			return m, tea.Quit

		case "esc":
			if m.step > 0 && m.step < stepCreating {
				m.step--
			}

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			switch m.step {
			case stepPreset:
				if m.cursor < 2 {
					m.cursor++
				}
			case stepOrientation:
				if m.cursor < 1 {
					m.cursor++
				}
			}

		case "enter", " ":
			return m.handleSelection()
		}
	}

	return m, nil
}

func (m *CreateDocumentWizard) handleSelection() (tea.Model, tea.Cmd) {
	switch m.step {
	case stepPreset:
		switch m.cursor {
		case 0:
			m.preset = idml.PresetLetterUS
		case 1:
			m.preset = idml.PresetA4
		case 2:
			m.preset = idml.PresetTabloid
		}
		m.step = stepOrientation
		m.cursor = 0

	case stepOrientation:
		if m.cursor == 0 {
			m.orientation = "Portrait"
		} else {
			m.orientation = "Landscape"
		}
		m.step = stepColumns
		m.cursor = 0

	case stepColumns:
		// Simple input - just proceed for now
		m.step = stepGutter
		m.cursor = 0

	case stepGutter:
		m.step = stepFilename
		m.cursor = 0

	case stepFilename:
		return m.createDocument()

	case stepDone:
		return m, tea.Quit
	}

	return m, nil
}

func (m *CreateDocumentWizard) createDocument() (tea.Model, tea.Cmd) {
	m.step = stepCreating

	// Add .idml extension if needed
	if !strings.HasSuffix(strings.ToLower(m.filename), ".idml") {
		m.filename += ".idml"
	}

	// Create document options
	opts := &idml.TemplateOptions{
		Preset:       m.preset,
		Orientation:  m.orientation,
		ColumnCount:  m.columns,
		ColumnGutter: m.gutter,
	}

	// Create the package
	pkg, err := idml.NewFromTemplate(opts)
	if err != nil {
		m.error = err.Error()
		m.step = stepDone
		return m, nil
	}

	// Save the document
	if err := idml.Write(pkg, m.filename); err != nil {
		m.error = err.Error()
		m.step = stepDone
		return m, nil
	}

	m.pkg = pkg
	m.success = true
	m.step = stepDone
	return m, nil
}

func (m *CreateDocumentWizard) View() string {
	var s strings.Builder

	s.WriteString(TitleStyle.Render("📄 Create New Document"))
	s.WriteString("\n\n")

	switch m.step {
	case stepPreset:
		s.WriteString(m.viewPresetSelection())
	case stepOrientation:
		s.WriteString(m.viewOrientationSelection())
	case stepColumns:
		s.WriteString(m.viewColumnsInput())
	case stepGutter:
		s.WriteString(m.viewGutterInput())
	case stepFilename:
		s.WriteString(m.viewFilenameInput())
	case stepCreating:
		s.WriteString(InfoStyle.Render("⚙️  Creating document..."))
	case stepDone:
		if m.success {
			s.WriteString(m.viewSuccess())
		} else {
			s.WriteString(ErrorStyle.Render("❌ Error: " + m.error))
		}
	}

	return s.String()
}

func (m *CreateDocumentWizard) viewPresetSelection() string {
	var s strings.Builder
	s.WriteString(InputLabelStyle.Render("📐 Select Page Size:"))
	s.WriteString("\n\n")

	presets := []struct {
		name   string
		desc   string
		preset idml.DocumentPreset
	}{
		{"Letter", "8.5\" × 11\" (US)", idml.PresetLetterUS},
		{"A4", "210mm × 297mm (ISO)", idml.PresetA4},
		{"Tabloid", "11\" × 17\"", idml.PresetTabloid},
	}

	for i, p := range presets {
		cursor := "  "
		if m.cursor == i {
			cursor = "▸ "
		}

		line := fmt.Sprintf("%s%d. %s", cursor, i+1, p.name)
		desc := fmt.Sprintf("     %s", p.desc)

		if m.cursor == i {
			s.WriteString(SelectedStyle.Render(line))
			s.WriteString("\n")
			s.WriteString(SubtitleStyle.Render(desc))
		} else {
			s.WriteString(NormalStyle.Render(line))
		}
		s.WriteString("\n")
	}

	s.WriteString(HelpStyle.Render("\n↑/k up • ↓/j down • enter select"))
	return s.String()
}

func (m *CreateDocumentWizard) viewOrientationSelection() string {
	var s strings.Builder
	s.WriteString(InputLabelStyle.Render("📱 Select Orientation:"))
	s.WriteString("\n\n")

	orientations := []struct {
		name string
		icon string
	}{
		{"Portrait", "📄"},
		{"Landscape", "🖼️ "},
	}

	for i, o := range orientations {
		cursor := "  "
		if m.cursor == i {
			cursor = "▸ "
		}

		line := fmt.Sprintf("%s%d. %s %s", cursor, i+1, o.icon, o.name)

		if m.cursor == i {
			s.WriteString(SelectedStyle.Render(line))
		} else {
			s.WriteString(NormalStyle.Render(line))
		}
		s.WriteString("\n")
	}

	s.WriteString(HelpStyle.Render("\n↑/k up • ↓/j down • enter select • esc back"))
	return s.String()
}

func (m *CreateDocumentWizard) viewColumnsInput() string {
	var s strings.Builder
	s.WriteString(InputLabelStyle.Render("📰 Number of Columns:"))
	s.WriteString("\n\n")
	s.WriteString(InfoStyle.Render(fmt.Sprintf("Current: %d", m.columns)))
	s.WriteString("\n")
	s.WriteString(SubtitleStyle.Render("Type a number (1-12) and press enter"))
	s.WriteString("\n")
	s.WriteString(HelpStyle.Render("enter to confirm • esc back"))
	return s.String()
}

func (m *CreateDocumentWizard) viewGutterInput() string {
	var s strings.Builder
	s.WriteString(InputLabelStyle.Render("📏 Column Gutter:"))
	s.WriteString("\n\n")
	s.WriteString(InfoStyle.Render(fmt.Sprintf("Current: %.1f points", m.gutter)))
	s.WriteString("\n")
	s.WriteString(SubtitleStyle.Render("Common: 12pt (1/6\"), 18pt (1/4\"), 24pt (1/3\")"))
	s.WriteString("\n")
	s.WriteString(HelpStyle.Render("enter to use current • esc back"))
	return s.String()
}

func (m *CreateDocumentWizard) viewFilenameInput() string {
	var s strings.Builder
	s.WriteString(InputLabelStyle.Render("💾 Output Filename:"))
	s.WriteString("\n\n")
	s.WriteString(InfoStyle.Render(m.filename))
	s.WriteString("\n")
	s.WriteString(SubtitleStyle.Render("Type filename and press enter"))
	s.WriteString("\n")
	s.WriteString(HelpStyle.Render("enter to use current • esc back"))
	return s.String()
}

func (m *CreateDocumentWizard) viewSuccess() string {
	var s strings.Builder
	s.WriteString(SuccessStyle.Render("✅ Document created successfully!"))
	s.WriteString("\n\n")

	opts := &idml.TemplateOptions{
		Preset:       m.preset,
		Orientation:  m.orientation,
		ColumnCount:  m.columns,
		ColumnGutter: m.gutter,
	}
	dims := opts.GetDimensions()

	s.WriteString(BoxStyle.Render(fmt.Sprintf(
		"📄 Document Details:\n"+
			"   Preset: %s\n"+
			"   Orientation: %s\n"+
			"   Dimensions: %.1f × %.1f points\n"+
			"   Columns: %d\n"+
			"   Gutter: %.1f points\n"+
			"\n"+
			"💾 Saved to: %s",
		m.presetName(),
		m.orientation,
		dims.Width, dims.Height,
		m.columns,
		m.gutter,
		m.filename,
	)))

	s.WriteString("\n\n")
	s.WriteString(HelpStyle.Render("q to quit"))
	return s.String()
}

func (m *CreateDocumentWizard) presetName() string {
	switch m.preset {
	case idml.PresetLetterUS:
		return "Letter (8.5\" × 11\")"
	case idml.PresetA4:
		return "A4 (210mm × 297mm)"
	case idml.PresetTabloid:
		return "Tabloid (11\" × 17\")"
	default:
		return "Unknown"
	}
}

// GetFilename returns the created filename
func (m *CreateDocumentWizard) GetFilename() string {
	return m.filename
}

// IsSuccess returns true if creation succeeded
func (m *CreateDocumentWizard) IsSuccess() bool {
	return m.success
}
