package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dimelords/idmllib/pkg/idml"
)

type RoundtripStep int

const (
	rtStepInputFile RoundtripStep = iota
	rtStepOutputFile
	rtStepProcessing
	rtStepDone
)

// RoundtripWizard handles roundtrip testing
type RoundtripWizard struct {
	step       RoundtripStep
	inputFile  string
	outputFile string
	inputText  string
	// cursor field removed as it was unused
	error      string
	success    bool
	fileCount  int
	storyCount int
	domVersion string
}

// NewRoundtripWizard creates a new roundtrip wizard
func NewRoundtripWizard() *RoundtripWizard {
	return &RoundtripWizard{
		step:       rtStepInputFile,
		outputFile: "roundtrip.idml",
	}
}

func (m *RoundtripWizard) Init() tea.Cmd {
	return nil
}

func (m *RoundtripWizard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.step == rtStepDone {
				return m, tea.Quit
			}
			m.error = "Cancelled by user"
			return m, tea.Quit

		case "esc":
			if m.step == rtStepOutputFile {
				m.step = rtStepInputFile
				m.inputText = ""
			}

		case "enter":
			return m.handleEnter()

		case "backspace":
			if len(m.inputText) > 0 {
				m.inputText = m.inputText[:len(m.inputText)-1]
			}

		default:
			// Add character to input
			if len(msg.String()) == 1 {
				m.inputText += msg.String()
			}
		}
	}

	return m, nil
}

func (m *RoundtripWizard) handleEnter() (tea.Model, tea.Cmd) {
	switch m.step {
	case rtStepInputFile:
		if m.inputText == "" {
			m.error = "Input file is required"
			return m, nil
		}
		m.inputFile = m.inputText
		m.inputText = ""
		m.step = rtStepOutputFile

	case rtStepOutputFile:
		if m.inputText != "" {
			m.outputFile = m.inputText
		}
		// Add .idml extension if needed
		if !strings.HasSuffix(strings.ToLower(m.outputFile), ".idml") {
			m.outputFile += ".idml"
		}
		return m.performRoundtrip()

	case rtStepDone:
		return m, tea.Quit
	}

	return m, nil
}

func (m *RoundtripWizard) performRoundtrip() (tea.Model, tea.Cmd) {
	m.step = rtStepProcessing

	// Read the file
	pkg, err := idml.Read(m.inputFile)
	if err != nil {
		m.error = fmt.Sprintf("Error reading file: %v", err)
		m.step = rtStepDone
		return m, nil
	}

	m.fileCount = pkg.FileCount()

	// Count stories in the package
	doc, err := pkg.Document()
	if err == nil {
		m.domVersion = doc.DOMVersion
		m.storyCount = len(doc.Stories)
	}

	// Write the file
	if err := idml.Write(pkg, m.outputFile); err != nil {
		m.error = fmt.Sprintf("Error writing file: %v", err)
		m.step = rtStepDone
		return m, nil
	}

	m.success = true
	m.step = rtStepDone
	return m, nil
}

func (m *RoundtripWizard) View() string {
	var s strings.Builder

	s.WriteString(TitleStyle.Render("🔄 Roundtrip Test"))
	s.WriteString("\n")
	s.WriteString(SubtitleStyle.Render("Read and write IDML files to verify compatibility"))
	s.WriteString("\n\n")

	switch m.step {
	case rtStepInputFile:
		s.WriteString(m.viewInputFile())
	case rtStepOutputFile:
		s.WriteString(m.viewOutputFile())
	case rtStepProcessing:
		s.WriteString(InfoStyle.Render("⚙️  Processing..."))
	case rtStepDone:
		if m.success {
			s.WriteString(m.viewSuccess())
		} else {
			s.WriteString(ErrorStyle.Render("❌ " + m.error))
			s.WriteString("\n\n")
			s.WriteString(HelpStyle.Render("q to quit"))
		}
	}

	return s.String()
}

func (m *RoundtripWizard) viewInputFile() string {
	var s strings.Builder

	s.WriteString(InputLabelStyle.Render("📂 Input IDML file:"))
	s.WriteString("\n\n")

	if m.inputText == "" {
		s.WriteString(SubtitleStyle.Render("  Enter path to IDML file..."))
	} else {
		s.WriteString(InputStyle.Render("  " + m.inputText + "█"))
	}

	s.WriteString("\n\n")

	if m.error != "" {
		s.WriteString(ErrorStyle.Render("  " + m.error))
		s.WriteString("\n\n")
		m.error = ""
	}

	s.WriteString(HelpStyle.Render("Type path and press enter • esc/q to cancel"))
	return s.String()
}

func (m *RoundtripWizard) viewOutputFile() string {
	var s strings.Builder

	s.WriteString(SuccessStyle.Render("  ✓ Input: " + m.inputFile))
	s.WriteString("\n\n")

	s.WriteString(InputLabelStyle.Render("💾 Output IDML file:"))
	s.WriteString("\n\n")

	if m.inputText == "" {
		s.WriteString(SubtitleStyle.Render("  " + m.outputFile + " (press enter to use)"))
	} else {
		s.WriteString(InputStyle.Render("  " + m.inputText + "█"))
	}

	s.WriteString("\n\n")
	s.WriteString(HelpStyle.Render("Type path or press enter for default • esc to go back"))
	return s.String()
}

func (m *RoundtripWizard) viewSuccess() string {
	var s strings.Builder

	s.WriteString(SuccessStyle.Render("✅ Roundtrip completed successfully!"))
	s.WriteString("\n\n")

	details := fmt.Sprintf(
		"📄 Files:\n"+
			"   Input:  %s\n"+
			"   Output: %s\n"+
			"\n"+
			"📊 Statistics:\n"+
			"   Files in package: %d\n"+
			"   Story references:  %d\n",
		m.inputFile,
		m.outputFile,
		m.fileCount,
		m.storyCount,
	)

	if m.domVersion != "" {
		details += fmt.Sprintf("   DOM version: %s\n", m.domVersion)
	}

	details += "\n🎨 Both files can be opened in Adobe InDesign!"

	s.WriteString(BoxStyle.Render(details))
	s.WriteString("\n\n")
	s.WriteString(HelpStyle.Render("q to quit"))

	return s.String()
}

// IsSuccess returns true if roundtrip succeeded
func (m *RoundtripWizard) IsSuccess() bool {
	return m.success
}
