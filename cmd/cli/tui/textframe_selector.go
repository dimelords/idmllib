// Package tui provides interactive terminal UI components using bubbletea
package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dimelords/idmllib/v2/pkg/idml"
	"github.com/dimelords/idmllib/v2/pkg/spread"
	"github.com/dimelords/idmllib/v2/pkg/story"
)

// TextFrameItem represents a textframe with preview text
type TextFrameItem struct {
	ID         string
	StoryID    string
	Preview    string
	SpreadName string
	FrameIndex int
}

// TextFrameSelector is the bubbletea model for selecting textframes
type TextFrameSelector struct {
	items         []TextFrameItem
	cursor        int
	selected      int
	pkg           *idml.Package
	stories       map[string]*story.Story
	spreads       map[string]*spread.Spread
	width         int
	height        int
	quitting      bool
	selectedFrame *TextFrameItem
}

// NewTextFrameSelector creates a new TextFrameSelector
func NewTextFrameSelector(pkg *idml.Package) (*TextFrameSelector, error) {
	spreads, err := pkg.Spreads()
	if err != nil {
		return nil, fmt.Errorf("failed to get spreads: %w", err)
	}

	stories, err := pkg.Stories()
	if err != nil {
		return nil, fmt.Errorf("failed to get stories: %w", err)
	}

	// Build story map for quick lookup by Self ID
	storyMap := make(map[string]*story.Story)
	for _, st := range stories {
		// Key by the story's Self attribute (e.g., "u1d8")
		storyMap[st.StoryElement.Self] = st
	}

	// Collect all textframes
	var items []TextFrameItem
	for spreadName, spread := range spreads {
		for frameIdx, tf := range spread.InnerSpread.TextFrames {
			preview := getStoryPreview(tf.ParentStory, storyMap)
			items = append(items, TextFrameItem{
				ID:         tf.Self,
				StoryID:    tf.ParentStory,
				Preview:    preview,
				SpreadName: spreadName,
				FrameIndex: frameIdx,
			})
		}
	}

	return &TextFrameSelector{
		items:    items,
		pkg:      pkg,
		stories:  storyMap,
		spreads:  spreads,
		selected: -1,
	}, nil
}

// getStoryPreview extracts the first ~100 chars of story content
func getStoryPreview(storyID string, stories map[string]*story.Story) string {
	st, ok := stories[storyID]
	if !ok {
		// Story not found - return a helpful debug message
		return fmt.Sprintf("(story %s not found)", storyID)
	}

	// Get first paragraph content
	var preview string
	for _, para := range st.StoryElement.ParagraphStyleRanges {
		for _, charRange := range para.CharacterStyleRanges {
			for _, child := range charRange.Children {
				if child.Content != nil {
					preview += child.Content.Text
					if len(preview) >= 100 {
						break
					}
				}
			}
			if len(preview) >= 100 {
				break
			}
		}
		if len(preview) >= 100 {
			break
		}
	}

	if preview == "" {
		return "(empty story)"
	}

	// Truncate and clean
	if len(preview) > 100 {
		preview = preview[:100] + "..."
	}
	preview = strings.ReplaceAll(preview, "\n", " ")
	preview = strings.ReplaceAll(preview, "\t", " ")

	return preview
}

func (m *TextFrameSelector) Init() tea.Cmd {
	return nil
}

func (m *TextFrameSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			// Ctrl+C should quit the whole app
			m.quitting = true
			return m, tea.Quit

		case "q", "esc":
			// q/esc just cancels selection, doesn't quit app
			m.quitting = true
			return m, nil

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
			m.selectedFrame = &m.items[m.cursor]
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func (m *TextFrameSelector) View() string {
	if m.quitting && m.selected == -1 {
		return ""
	}

	if m.selectedFrame != nil {
		return ""
	}

	var s strings.Builder

	// Title with count
	title := fmt.Sprintf("Select a text frame to export (found %d)", len(m.items))
	s.WriteString(TitleStyle.Render(title))
	s.WriteString("\n\n")

	// Show items with checkbox style
	for i, item := range m.items {
		// Checkbox indicator
		checkbox := CheckboxEmpty
		style := UnselectedStyle

		if m.cursor == i {
			checkbox = CheckboxSelected
			style = SelectedStyle
		}

		// Truncate preview for cleaner display
		preview := item.Preview
		if len(preview) > 70 {
			preview = preview[:67] + "..."
		}

		// Format: [x] FrameID
		line := fmt.Sprintf("%s %s", checkbox, item.ID)
		s.WriteString(style.Render(line))
		s.WriteString("\n")

		// Show preview for selected item only
		if m.cursor == i {
			previewLine := fmt.Sprintf("    %s", preview)
			s.WriteString(DescStyle.Render(previewLine))
			s.WriteString("\n")
		}
	}

	// Help text
	s.WriteString(FormatHelp(
		"j/k, up/down: select",
		"enter: choose",
		"q, esc: cancel",
	))

	return s.String()
}

// GetSelectedFrame returns the selected textframe, or nil if none selected
func (m *TextFrameSelector) GetSelectedFrame() *TextFrameItem {
	return m.selectedFrame
}
