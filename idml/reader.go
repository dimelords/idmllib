package idml

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/dimelords/idmllib/types"
)

// GetSpreads returns all loaded spreads
func (p *Package) GetSpreads() []types.Spread {
	return p.Spreads
}

// CollectTextFrames iterates through all spreads and collects TextFrames matching the predicate
func (p *Package) CollectTextFrames(shouldInclude types.TextFramePredicate) ([]types.SelectedFrame, error) {
	var selected []types.SelectedFrame

	// Loop through all spread files in the IDML
	for _, f := range p.reader.File {
		if !strings.HasPrefix(f.Name, "Spreads/Spread_") || !strings.HasSuffix(f.Name, ".xml") {
			continue
		}

		spreadData, err := p.readFileFromZipFile(f)
		if err != nil {
			slog.Warn("Failed to read spread file", "file", f.Name, "error", err)
			continue
		}

		// Parse the spread
		spread, err := p.readSpread(spreadData)
		if err != nil {
			slog.Warn("Failed to parse spread", "file", f.Name, "error", err)
			continue
		}

		// Check each TextFrame with the predicate
		for _, tf := range spread.TextFrames {
			if shouldInclude(&tf) {
				selected = append(selected, types.SelectedFrame{
					TextFrame: tf,
					Spread:    spread,
				})
			}
		}
	}

	return selected, nil
}

// LoadResources loads all resources (styles, colors, layers, etc.) into the IDMS document
func (p *Package) LoadResources(idms *types.IDMS) error {
	if err := p.loadStyles(idms); err != nil {
		slog.Warn("Failed to load styles", "error", err)
	}

	if err := p.loadColorsAndSwatches(idms); err != nil {
		slog.Warn("Failed to load colors and swatches", "error", err)
	}

	if err := p.loadTransparencyDefaults(idms); err != nil {
		slog.Warn("Failed to load transparency defaults", "error", err)
	}

	if err := p.loadColorGroups(idms); err != nil {
		slog.Warn("Failed to load color groups", "error", err)
	}

	if err := p.loadLayers(idms); err != nil {
		slog.Warn("Failed to load layers", "error", err)
	}

	return nil
}

// getSpreadsForStory finds and returns Spread structs for spreads that contain
// TextFrames referencing the given story. Only TextFrames for this story are included.
func (p *Package) getSpreadsForStory(story *types.Story) ([]types.Spread, error) {
	var spreads []types.Spread

	// Loop through all spread files in the IDML
	for _, f := range p.reader.File {
		if !strings.HasPrefix(f.Name, "Spreads/Spread_") || !strings.HasSuffix(f.Name, ".xml") {
			continue
		}

		spreadData, err := p.readFileFromZipFile(f)
		if err != nil {
			continue
		}

		// Check if this spread contains our story
		spreadXML := string(spreadData)
		if !strings.Contains(spreadXML, fmt.Sprintf(`ParentStory="%s"`, story.Self)) {
			continue
		}

		// Parse the spread
		spread, err := p.readSpread(spreadData)
		if err != nil {
			slog.Warn("Failed to parse spread", "file", f.Name, "error", err)
			continue
		}

		// Filter TextFrames to only include those for this story
		var filteredTextFrames []types.TextFrame
		for _, tf := range spread.TextFrames {
			if tf.ParentStory == story.Self {
				filteredTextFrames = append(filteredTextFrames, tf)
			}
		}

		// Only add the spread if it has TextFrames for this story
		if len(filteredTextFrames) > 0 {
			spread.TextFrames = filteredTextFrames
			// Clear other elements (Pages, Rectangles) as we only want TextFrames
			spread.Pages = nil
			spread.Rectangles = nil
			spreads = append(spreads, spread)
		}
	}

	if len(spreads) == 0 {
		return nil, &SpreadNotFoundError{StoryID: story.Self}
	}

	return spreads, nil
}
