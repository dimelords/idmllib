package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dimelords/idmllib/pkg/idml"
	"github.com/dimelords/idmllib/pkg/idms"
)

type exportStep int

const (
	expStepInputFile exportStep = iota
	expStepSelectFrame
	expStepFrameActions
	expStepOutputFile
	expStepExporting
	expStepSuccess
	expStepNextAction
)

// ExportIDMSWizard handles IDMS export
type ExportIDMSWizard struct {
	step          exportStep
	inputFile     string
	outputFile    string
	inputText     string
	pkg           *idml.Package
	selectedFrame *TextFrameItem
	error         string
	success       bool
	frameSelector *TextFrameSelector
	actionMenu    *ActionMenu
	nextMenu      *ActionMenu
	exportCount   int
}

// NewExportIDMSWizard creates a new export wizard
func NewExportIDMSWizard() *ExportIDMSWizard {
	return &ExportIDMSWizard{
		step:       expStepInputFile,
		outputFile: "export.idms",
	}
}

func (m *ExportIDMSWizard) Init() tea.Cmd {
	return nil
}

func (m *ExportIDMSWizard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle next action menu after success
	if m.step == expStepNextAction && m.nextMenu != nil {
		updatedMenu, cmd := m.nextMenu.Update(msg)
		m.nextMenu = updatedMenu.(*ActionMenu)

		if m.nextMenu.selected != -1 {
			return m.handleNextAction(m.nextMenu.selected)
		}
		if m.nextMenu.quitting {
			return m, tea.Quit
		}

		return m, cmd
	}

	// Handle frame action menu
	if m.step == expStepFrameActions && m.actionMenu != nil {
		updatedMenu, cmd := m.actionMenu.Update(msg)
		m.actionMenu = updatedMenu.(*ActionMenu)

		if m.actionMenu.selected != -1 {
			return m.handleFrameAction(m.actionMenu.selected)
		}
		if m.actionMenu.quitting {
			m.step = expStepSelectFrame
			m.frameSelector, _ = NewTextFrameSelector(m.pkg)
			return m, nil
		}

		return m, cmd
	}

	// If we're in frame selection mode, delegate to frame selector
	if m.step == expStepSelectFrame && m.frameSelector != nil {
		// Check if frame selector is done
		if m.frameSelector.selectedFrame != nil {
			m.selectedFrame = m.frameSelector.selectedFrame

			// Show action menu for what to do with this frame
			m.actionMenu = NewActionMenu(
				"What would you like to do with this frame?",
				[]string{
					"Export as IDMS",
					"View frame details",
					"Select different frame",
					"Return to main menu",
				},
			)
			m.step = expStepFrameActions
			return m, nil
		}
		if m.frameSelector.quitting && m.frameSelector.selectedFrame == nil {
			// User pressed q/esc in frame selector - go back to main menu
			return m, tea.Quit
		}

		// Otherwise delegate to frame selector
		updatedSelector, cmd := m.frameSelector.Update(msg)
		m.frameSelector = updatedSelector.(*TextFrameSelector)
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			// Ctrl+C always quits
			return m, tea.Quit

		case "q":
			// 'q' always quits back to main menu
			return m, tea.Quit

		case "esc":
			if m.step == expStepOutputFile {
				// Go back to frame actions
				m.step = expStepFrameActions
				m.inputText = ""
				return m, nil
			} else if m.step == expStepInputFile {
				// Cancel entire export
				return m, tea.Quit
			}

		case "enter":
			return m.handleEnter()

		case "backspace":
			if len(m.inputText) > 0 {
				m.inputText = m.inputText[:len(m.inputText)-1]
			}

		default:
			// Add character to input (only in input steps)
			if m.step == expStepInputFile || m.step == expStepOutputFile {
				if len(msg.String()) == 1 {
					m.inputText += msg.String()
				}
			}
		}
	}

	return m, nil
}

func (m *ExportIDMSWizard) handleFrameAction(action int) (tea.Model, tea.Cmd) {
	switch action {
	case 0: // Export as IDMS
		m.step = expStepOutputFile
		m.inputText = ""
		return m, nil

	case 1: // View frame details
		// TODO: Implement frame details view
		m.error = "Frame details view not yet implemented"
		m.step = expStepFrameActions
		return m, nil

	case 2: // Select different frame
		m.step = expStepSelectFrame
		m.selectedFrame = nil
		selector, _ := NewTextFrameSelector(m.pkg)
		m.frameSelector = selector
		return m, selector.Init()

	case 3: // Return to main menu
		return m, tea.Quit
	}
	return m, nil
}

func (m *ExportIDMSWizard) handleNextAction(action int) (tea.Model, tea.Cmd) {
	switch action {
	case 0: // Export another frame from this file
		m.selectedFrame = nil
		m.step = expStepSelectFrame
		m.outputFile = fmt.Sprintf("export_%d.idms", m.exportCount+1)
		selector, _ := NewTextFrameSelector(m.pkg)
		m.frameSelector = selector
		return m, selector.Init()

	case 1: // Export from different file
		m.pkg = nil
		m.selectedFrame = nil
		m.step = expStepInputFile
		m.inputText = ""
		m.success = false
		return m, nil

	case 2: // Return to main menu
		return m, tea.Quit
	}
	return m, nil
}

func (m *ExportIDMSWizard) handleEnter() (tea.Model, tea.Cmd) {
	switch m.step {
	case expStepInputFile:
		if m.inputText == "" {
			m.inputText = "testdata/plain.idml"
		}
		m.inputFile = m.inputText
		m.inputText = ""

		// Load the IDML file
		pkg, err := idml.Read(m.inputFile)
		if err != nil {
			m.error = fmt.Sprintf("Error reading file: %v", err)
			m.step = expStepInputFile
			m.inputFile = ""
			return m, nil
		}
		m.pkg = pkg

		// Create frame selector
		selector, err := NewTextFrameSelector(pkg)
		if err != nil {
			m.error = fmt.Sprintf("Error creating selector: %v", err)
			m.step = expStepInputFile
			return m, nil
		}

		m.frameSelector = selector
		m.step = expStepSelectFrame
		return m, selector.Init()

	case expStepOutputFile:
		if m.inputText != "" {
			m.outputFile = m.inputText
		}
		// Add .idms extension if needed
		if !strings.HasSuffix(strings.ToLower(m.outputFile), ".idms") {
			m.outputFile += ".idms"
		}
		return m.performExport()

	case expStepSuccess:
		// Show next action menu
		m.nextMenu = NewActionMenu(
			"Export successful! What would you like to do next?",
			[]string{
				"Export another frame from this file",
				"Export from a different file",
				"Return to main menu",
			},
		)
		m.step = expStepNextAction
		return m, nil
	}

	return m, nil
}

func (m *ExportIDMSWizard) performExport() (tea.Model, tea.Cmd) {
	m.step = expStepExporting

	// Get spreads
	spreads, err := m.pkg.Spreads()
	if err != nil {
		m.error = fmt.Sprintf("Error getting spreads: %v", err)
		m.step = expStepFrameActions
		return m, nil
	}

	// Get the selected spread
	spread, ok := spreads[m.selectedFrame.SpreadName]
	if !ok {
		m.error = fmt.Sprintf("Spread %s not found", m.selectedFrame.SpreadName)
		m.step = expStepFrameActions
		return m, nil
	}

	// Get the textframe
	if m.selectedFrame.FrameIndex >= len(spread.InnerSpread.TextFrames) {
		m.error = fmt.Sprintf("Invalid frame index %d", m.selectedFrame.FrameIndex)
		m.step = expStepFrameActions
		return m, nil
	}

	textFrame := &spread.InnerSpread.TextFrames[m.selectedFrame.FrameIndex]

	// Create selection
	selection := idml.NewSelection()
	selection.AddTextFrame(textFrame)

	// Export
	exporter := idms.NewExporter(m.pkg)
	snippet, err := exporter.ExportSelection(selection)
	if err != nil {
		m.error = fmt.Sprintf("Error exporting: %v", err)
		m.step = expStepFrameActions
		return m, nil
	}

	// Write file - get absolute path for display
	absPath := m.outputFile
	if !filepath.IsAbs(m.outputFile) {
		cwd, _ := os.Getwd()
		absPath = filepath.Join(cwd, m.outputFile)
	}

	if err := idms.Write(snippet, m.outputFile); err != nil {
		m.error = fmt.Sprintf("Error writing file to %s: %v", absPath, err)
		m.step = expStepFrameActions
		return m, nil
	}

	// Store the absolute path for success message
	m.outputFile = absPath
	m.success = true
	m.exportCount++
	m.step = expStepSuccess
	return m, nil
}

func (m *ExportIDMSWizard) View() string {
	// Delegate to next action menu
	if m.step == expStepNextAction && m.nextMenu != nil {
		return m.nextMenu.View()
	}

	// Delegate to frame action menu
	if m.step == expStepFrameActions && m.actionMenu != nil {
		var s strings.Builder

		// Show selected frame info compactly
		s.WriteString(SuccessStyle.Render(fmt.Sprintf("✓ Frame: %s", m.selectedFrame.ID)))

		// Truncate preview
		preview := m.selectedFrame.Preview
		if len(preview) > 50 {
			preview = preview[:47] + "..."
		}
		s.WriteString(SubtitleStyle.Render(fmt.Sprintf(" - %s", preview)))
		s.WriteString("\n\n")

		s.WriteString(m.actionMenu.View())
		return s.String()
	}

	// Delegate to frame selector if in that step
	if m.step == expStepSelectFrame && m.frameSelector != nil {
		return m.frameSelector.View()
	}

	var s strings.Builder

	switch m.step {
	case expStepInputFile:
		s.WriteString(m.viewInputFile())
	case expStepOutputFile:
		s.WriteString(m.viewOutputFile())
	case expStepExporting:
		s.WriteString(InfoStyle.Render("Exporting..."))
	case expStepSuccess:
		s.WriteString(m.viewSuccess())
	}

	return s.String()
}

func (m *ExportIDMSWizard) viewInputFile() string {
	var s strings.Builder

	if m.exportCount > 0 {
		s.WriteString(SuccessStyle.Render(fmt.Sprintf("✓ %d export(s) completed", m.exportCount)))
		s.WriteString("\n\n")
	}

	// Question style prompt
	s.WriteString(TitleStyle.Render("Input IDML file path?"))
	s.WriteString("\n")

	if m.inputText == "" {
		s.WriteString(SubtitleStyle.Render("  (press enter for: testdata/plain.idml)"))
	} else {
		s.WriteString(InputStyle.Render("  " + m.inputText))
		s.WriteString(InputPromptStyle.Render("█"))
	}

	if m.error != "" {
		s.WriteString("\n")
		s.WriteString(ErrorStyle.Render("  " + m.error))
		m.error = ""
	}

	s.WriteString("\n")
	s.WriteString(FormatHelp("type path", "enter: confirm", "esc/ctrl+c: cancel"))
	return s.String()
}

func (m *ExportIDMSWizard) viewOutputFile() string {
	var s strings.Builder

	// Show selected frame info compactly
	s.WriteString(SuccessStyle.Render(fmt.Sprintf("✓ Frame: %s", m.selectedFrame.ID)))

	// Truncate preview
	preview := m.selectedFrame.Preview
	if len(preview) > 50 {
		preview = preview[:47] + "..."
	}
	s.WriteString(SubtitleStyle.Render(fmt.Sprintf(" - %s", preview)))
	s.WriteString("\n\n")

	// Compact prompt
	s.WriteString(TitleStyle.Render("Output IDMS file path?"))
	s.WriteString("\n")

	if m.inputText == "" {
		s.WriteString(SubtitleStyle.Render(fmt.Sprintf("  (press enter for: %s)", m.outputFile)))
	} else {
		s.WriteString(InputStyle.Render("  " + m.inputText))
		s.WriteString(InputPromptStyle.Render("█"))
	}

	s.WriteString("\n")
	s.WriteString(FormatHelp("type path", "enter: save", "esc: go back"))
	return s.String()
}

func (m *ExportIDMSWizard) viewSuccess() string {
	var s strings.Builder

	s.WriteString(SuccessStyle.Render("✓ IDMS exported successfully!"))
	s.WriteString("\n")

	// Show full path prominently
	s.WriteString(HeaderStyle.Render("Saved to:"))
	s.WriteString("\n")
	s.WriteString(InputStyle.Render("  " + m.outputFile))
	s.WriteString("\n")

	s.WriteString(FormatHelp("enter: continue"))

	return s.String()
}

// IsSuccess returns true if export succeeded
func (m *ExportIDMSWizard) IsSuccess() bool {
	return m.success
}

// GetExportCount returns the number of successful exports
func (m *ExportIDMSWizard) GetExportCount() int {
	return m.exportCount
}

// HasError returns true if there was an error (not just cancellation)
func (m *ExportIDMSWizard) HasError() bool {
	return m.error != ""
}
