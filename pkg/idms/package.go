// Package idms provides support for reading and writing InDesign Snippet (IDMS) files.
//
// IDMS files are single XML files that contain a self-contained subset of an InDesign
// document, typically representing selected page items and their dependencies.
//
// Key differences from IDML:
//   - IDMS is a single XML file, not a ZIP archive
//   - IDMS contains XML processing instructions for snippet metadata
//   - IDMS embeds XMP metadata as a processing instruction at the end
//   - IDMS represents a selection of items, not a complete document
//
// Architecture:
//   - Reuses types from pkg/idml (Document, Spread, Story, etc.)
//   - Adds IDMS-specific processing instruction handling
//   - Provides custom XML marshaling/unmarshaling for PIs
package idms

import (
	"fmt"

	"github.com/dimelords/idmllib/v2/pkg/document"
	"github.com/dimelords/idmllib/v2/pkg/spread"
	"github.com/dimelords/idmllib/v2/pkg/story"
)

// Package represents an InDesign Snippet (IDMS) file.
//
// An IDMS file contains:
//   - XML processing instructions for snippet metadata
//   - A Document element (reusing idml.Document)
//   - XMP metadata as a processing instruction
type Package struct {
	// XMLDeclaration is the XML declaration (e.g., <?xml version="1.0" encoding="UTF-8"?>)
	XMLDeclaration string

	// AIDProcessingInstructions contains the <?aid ...?> directives
	// Example: <?aid style="50" type="snippet" readerVersion="6.0" ...?>
	AIDProcessingInstructions []document.ProcessingInstruction

	// Document is the root IDMS document (reuses document.Document)
	Document *document.Document

	// XMPMetadata contains the XMP packet at the end of the file
	// Example: <?xpacket begin="" id="..."?>...<x:xmpmeta>...</x:xmpmeta><?xpacket end="r"?>
	XMPMetadata string

	// RawXML stores the original XML for perfect roundtripping
	// This is only populated when reading an existing IDMS file
	rawXML []byte
}

// New creates a new empty IDMS Package.
func New() *Package {
	return &Package{
		XMLDeclaration: `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`,
		Document:       &document.Document{},
	}
}

// SnippetType returns the snippet type from the AID processing instructions.
// Common values: "PageItem", "TextFrame", "Rectangle", "Image"
func (p *Package) SnippetType() string {
	for _, pi := range p.AIDProcessingInstructions {
		if pi.Target == "aid" && len(pi.Inst) > 0 {
			// Parse SnippetType from data
			// Example: 'SnippetType="PageItem"'
			// Simple string parsing for SnippetType
			if len(pi.Inst) > 13 && pi.Inst[:12] == "SnippetType=" {
				// Extract value between quotes
				data := pi.Inst[13:]
				if idx := len(data) - 1; idx >= 0 && data[idx] == '"' {
					return data[:idx]
				}
			}
		}
	}
	return ""
}

// SetSnippetType sets the snippet type in the AID processing instructions.
func (p *Package) SetSnippetType(snippetType string) {
	// Add or update the SnippetType PI
	found := false
	for i, pi := range p.AIDProcessingInstructions {
		if pi.Target == "aid" && len(pi.Inst) > 12 && pi.Inst[:12] == "SnippetType=" {
			p.AIDProcessingInstructions[i].Inst = fmt.Sprintf(`SnippetType="%s"`, snippetType)
			found = true
			break
		}
	}
	if !found {
		p.AIDProcessingInstructions = append(p.AIDProcessingInstructions, document.ProcessingInstruction{
			Target: "aid",
			Inst:   fmt.Sprintf(`SnippetType="%s"`, snippetType),
		})
	}
}

// SetDefaultAIDProcessingInstructions sets up the standard AID processing instructions
// for an InDesign snippet. This includes both the main AID PI with version/feature info
// and the SnippetType PI.
//
// Default values match InDesign 2025 (version 20.5):
//   - style="50"
//   - type="snippet"
//   - readerVersion="6.0"
//   - featureSet="257"
//   - product="20.5(66)" (InDesign 2025)
func (p *Package) SetDefaultAIDProcessingInstructions(snippetType string) {
	// Clear existing PIs and add the main AID PI first
	p.AIDProcessingInstructions = []document.ProcessingInstruction{
		{
			Target: "aid",
			Inst:   `style="50" type="snippet" readerVersion="6.0" featureSet="257" product="20.5(66)"`,
		},
		{
			Target: "aid",
			Inst:   fmt.Sprintf(`SnippetType="%s"`, snippetType),
		},
	}
}

// Validate checks if the Package is valid.
func (p *Package) Validate() error {
	if p.Document == nil {
		return fmt.Errorf("document is nil")
	}
	if len(p.AIDProcessingInstructions) == 0 {
		return fmt.Errorf("missing AID processing instructions")
	}
	return nil
}

// Spreads returns all inline spreads from the IDMS document.
// Unlike IDML where spreads are in separate files, IDMS embeds them directly
// in the Document element.
//
// Returns a slice of spread elements. Each element contains the full spread
// structure including page items (TextFrames, Rectangles, etc.).
func (p *Package) Spreads() []spread.SpreadElement {
	if p.Document == nil {
		return nil
	}
	return p.Document.InlineSpreads
}

// Stories returns all inline stories from the IDMS document.
// Unlike IDML where stories are in separate files, IDMS embeds them directly
// in the Document element.
//
// Returns a slice of story elements. Each element contains the full story
// structure including text content and style ranges.
func (p *Package) Stories() []story.StoryElement {
	if p.Document == nil {
		return nil
	}
	return p.Document.InlineStories
}
