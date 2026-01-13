// IDML Builder - Main CLI Tool with Bubbletea TUI
package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dimelords/idmllib/v2/cmd/cli/tui"
)

func main() {
	for {
		// Show main menu
		menu := tui.NewMainMenu()
		p := tea.NewProgram(menu)
		m, err := p.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		finalMenu := m.(*tui.MainMenu)
		selected := finalMenu.GetSelected()

		// Handle selection
		switch selected {
		case 0:
			// Create new document
			runCreateDocument()

		case 1:
			// Roundtrip test
			runRoundtrip()

		case 2:
			// Export IDMS
			runExportIDMS()

		case 3:
			// Exit
			fmt.Println("\nGoodbye!")
			os.Exit(0)

		default:
			// User cancelled (Ctrl+C)
			fmt.Println()
			os.Exit(0)
		}

		// Add spacing between operations
		fmt.Println("\n" + strings.Repeat("─", 50) + "\n")
	}
}

func runCreateDocument() {
	wizard := tui.NewCreateDocumentWizard()
	p := tea.NewProgram(wizard)
	finalWizard, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}

	// Check if successful
	wizard = finalWizard.(*tui.CreateDocumentWizard)
	if wizard.IsSuccess() {
		fmt.Printf("\n✓ Document created: %s\n", wizard.GetFilename())
	}
}

func runRoundtrip() {
	wizard := tui.NewRoundtripWizard()
	p := tea.NewProgram(wizard)
	finalWizard, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}

	// Check if successful
	wizard = finalWizard.(*tui.RoundtripWizard)
	if wizard.IsSuccess() {
		fmt.Println("\n✓ Roundtrip test completed")
	}
}

func runExportIDMS() {
	wizard := tui.NewExportIDMSWizard()
	p := tea.NewProgram(wizard, tea.WithAltScreen())
	finalWizard, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}

	// Check if successful
	wizard = finalWizard.(*tui.ExportIDMSWizard)
	exportCount := wizard.GetExportCount()

	if exportCount > 0 {
		if exportCount == 1 {
			fmt.Println("\n✓ Successfully exported 1 IDMS snippet")
		} else {
			fmt.Printf("\n✓ Successfully exported %d IDMS snippets\n", exportCount)
		}

		// Show a reminder about where files are saved
		fmt.Println("  Check the success screen for file locations")
	}
}
