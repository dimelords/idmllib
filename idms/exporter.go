package idms

import (
	"encoding/xml"
	"fmt"
	"log/slog"
	"os"

	"github.com/dimelords/idmllib/idms/filter"
	"github.com/dimelords/idmllib/types"
)

// Reader provides read-only access to IDML package contents needed for export
type Reader interface {
	GetStory(id string) (*types.Story, error)
	GetSpreads() []types.Spread
	CollectTextFrames(shouldInclude types.TextFramePredicate) ([]types.SelectedFrame, error)
	LoadResources(idms *types.IDMS) error
}

// Exporter handles IDMS export operations
type Exporter struct {
	reader Reader
}

// NewExporter creates a new IDMS exporter
func NewExporter(reader Reader) *Exporter {
	return &Exporter{reader: reader}
}

// ExportXML exports TextFrames matching the predicate as an IDMS XML file
// The predicate function is called for each TextFrame and should return true if it should be included
func (e *Exporter) ExportXML(outputPath string, shouldInclude types.TextFramePredicate) error {
	// Collect all TextFrames that match the predicate
	selectedFrames, err := e.reader.CollectTextFrames(shouldInclude)
	if err != nil {
		return fmt.Errorf("failed to collect text frames: %w", err)
	}

	if len(selectedFrames) == 0 {
		return fmt.Errorf("no text frames matched the predicate")
	}

	// Group by stories - collect unique stories referenced by selected TextFrames
	storyMap := make(map[string]*types.Story)
	for _, sf := range selectedFrames {
		if _, exists := storyMap[sf.TextFrame.ParentStory]; !exists {
			story, err := e.reader.GetStory(sf.TextFrame.ParentStory)
			if err != nil {
				slog.Warn("Could not load story", "storyID", sf.TextFrame.ParentStory, "error", err)
				continue
			}
			storyMap[sf.TextFrame.ParentStory] = story
		}
	}

	// Group by spreads - organize TextFrames by their parent spreads
	spreadMap := make(map[string]*types.Spread)
	for _, sf := range selectedFrames {
		spreadID := sf.Spread.Self
		if spread, exists := spreadMap[spreadID]; exists {
			// Spread already exists, add TextFrame to it
			spread.TextFrames = append(spread.TextFrames, sf.TextFrame)
		} else {
			// New spread, create entry
			newSpread := sf.Spread
			newSpread.TextFrames = []types.TextFrame{sf.TextFrame}
			newSpread.Pages = nil      // Clear pages
			newSpread.Rectangles = nil // Clear rectangles
			spreadMap[spreadID] = &newSpread
		}
	}

	// Convert maps to slices
	var stories []types.Story
	for _, story := range storyMap {
		stories = append(stories, *story)
	}

	var spreads []types.Spread
	for _, spread := range spreadMap {
		spreads = append(spreads, *spread)
	}

	// Build IDMS document with all resources (no filtering yet)
	idms, err := e.buildIDMSDocument()
	if err != nil {
		return fmt.Errorf("failed to build IDMS document: %w", err)
	}

	// Add the spreads and stories
	idms.Spreads = spreads
	idms.Stories = stories

	// Filter layers to only include those used by TextFrames
	filterUnusedLayers(idms, spreads)

	// Filter styles using the filter package
	// We create a context for the FIRST story to maintain compatibility
	// In the future, we could enhance filter to handle multiple stories
	if len(stories) > 0 {
		ctx := &filter.Context{
			IDMS:    idms,
			Story:   &stories[0],
			Spreads: spreads,
		}
		if err := filter.UnusedStyles(ctx); err != nil {
			slog.Warn("Failed to filter styles", "error", err)
		}
	}

	// Write to file
	return writeIDMSToFile(outputPath, idms)
}

// ExportStoryXML exports a complete story as an IDMS XML file
// This is a convenience wrapper around ExportXML for the common use case of exporting an entire story
func (e *Exporter) ExportStoryXML(storyID, outputPath string) error {
	return e.ExportXML(outputPath, func(tf *types.TextFrame) bool {
		return tf.ParentStory == storyID
	})
}

// buildIDMSDocument constructs an IDMS struct with all resources loaded
func (e *Exporter) buildIDMSDocument() (*types.IDMS, error) {
	idms := &types.IDMS{
		DOMVersion: "20.4",
		Self:       "d",
	}

	// Load all resources from the reader
	if err := e.reader.LoadResources(idms); err != nil {
		return nil, fmt.Errorf("failed to load resources: %w", err)
	}

	// Add default stroke style
	idms.StrokeStyles = []types.StrokeStyle{
		{Self: "StrokeStyle/$ID/Solid", Name: "$ID/Solid"},
	}

	// Add TinDocumentDataObject
	idms.TinDocumentDataObject = &types.TinDocumentDataObject{
		Properties: &types.Properties{
			InnerXML: "<GaijiRefMaps><![CDATA[/////wAAAAAAAAAA]]></GaijiRefMaps>",
		},
	}

	return idms, nil
}

// writeIDMSToFile writes an IDMS document to a file with proper XML headers
func writeIDMSToFile(outputPath string, idms *types.IDMS) error {
	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer func(out *os.File) {
		_ = out.Close()
	}(out)

	// Write XML header and processing instructions
	header := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<?aid style="50" type="snippet" readerVersion="6.0" featureSet="257" product="20.5(48)" ?>
<?aid SnippetType="PageItem"?>
`
	if _, err := out.WriteString(header); err != nil {
		return err
	}

	// Marshal the complete IDMS document to XML
	encoder := xml.NewEncoder(out)
	encoder.Indent("", "\t")
	if err := encoder.Encode(idms); err != nil {
		return fmt.Errorf("failed to encode XML: %w", err)
	}

	return nil
}

// filterUnusedLayers removes layers that are not used by any TextFrames
func filterUnusedLayers(idms *types.IDMS, spreads []types.Spread) {
	usedLayers := make(map[string]bool)
	for _, spread := range spreads {
		for _, tf := range spread.TextFrames {
			usedLayers[tf.ItemLayer] = true
		}
	}

	var filteredLayers []types.Layer
	for _, layer := range idms.Layers {
		if usedLayers[layer.Self] {
			filteredLayers = append(filteredLayers, layer)
		}
	}
	idms.Layers = filteredLayers
}
