// Package tui provides shared styles and utilities for the TUI
package tui

import "github.com/charmbracelet/lipgloss"

// Adaptive colors that work on both light and dark backgrounds
var (
	// Primary colors - work on both backgrounds
	colorPrimary = lipgloss.AdaptiveColor{Light: "#FF06B7", Dark: "#FF06B7"} // Magenta
	colorAccent  = lipgloss.AdaptiveColor{Light: "#00A5D9", Dark: "#00D9FF"} // Cyan

	// Semantic colors
	colorSuccess = lipgloss.AdaptiveColor{Light: "#00AF87", Dark: "#00D787"} // Green
	colorError   = lipgloss.AdaptiveColor{Light: "#D70000", Dark: "#FF5F87"} // Red
	colorWarning = lipgloss.AdaptiveColor{Light: "#D78700", Dark: "#FFD75F"} // Yellow

	// Neutral colors - adapt to terminal theme
	colorText    = lipgloss.AdaptiveColor{Light: "#1A1A1A", Dark: "#E4E4E4"} // Main text
	colorTextDim = lipgloss.AdaptiveColor{Light: "#6C6C6C", Dark: "#6C6C6C"} // Dim gray
	colorBorder  = lipgloss.AdaptiveColor{Light: "#D0D0D0", Dark: "#3A3A3A"} // Border
)

// Title styles - minimal
var (
	TitleStyle = lipgloss.NewStyle().
			Foreground(colorText)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(colorTextDim)

	HeaderStyle = lipgloss.NewStyle().
			Foreground(colorText).
			MarginTop(1)
)

// Selection and menu styles - checkbox style
var (
	// Selected item with [x] indicator
	SelectedStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			PaddingLeft(1)

	// Unselected item with [ ] indicator
	UnselectedStyle = lipgloss.NewStyle().
			Foreground(colorText).
			PaddingLeft(1)

	// Description text
	DescStyle = lipgloss.NewStyle().
			Foreground(colorTextDim).
			PaddingLeft(5)

	// Legacy alias
	NormalStyle = UnselectedStyle
)

// Status styles
var (
	SuccessStyle = lipgloss.NewStyle().
			Foreground(colorSuccess)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(colorError)

	WarningStyle = lipgloss.NewStyle().
			Foreground(colorWarning)

	InfoStyle = lipgloss.NewStyle().
			Foreground(colorAccent)
)

// Input styles - clean with underline
var (
	InputStyle = lipgloss.NewStyle().
			Foreground(colorText)

	InputLabelStyle = lipgloss.NewStyle().
			Foreground(colorText).
			MarginBottom(1)

	InputPromptStyle = lipgloss.NewStyle().
				Foreground(colorPrimary)
)

// Help text style
var (
	HelpStyle = lipgloss.NewStyle().
			Foreground(colorTextDim).
			MarginTop(1)

	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(colorTextDim)

	HelpSepStyle = lipgloss.NewStyle().
			Foreground(colorTextDim)
)

// Box styles - minimal border
var (
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorBorder).
			Padding(1, 2)

	BoxTitleStyle = lipgloss.NewStyle().
			Foreground(colorText).
			Bold(true)
)

// Checkbox indicators
const (
	CheckboxEmpty    = "[ ]"
	CheckboxSelected = "[x]"
	CheckboxChecked  = "[✓]"
)

// Format help text in the huh style: "j/k, up/down: select • enter: choose • q, esc: quit"
func FormatHelp(items ...string) string {
	result := ""
	for i, item := range items {
		if i > 0 {
			result += HelpSepStyle.Render(" • ")
		}
		result += HelpKeyStyle.Render(item)
	}
	return HelpStyle.Render(result)
}
