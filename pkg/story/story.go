package story

import (
	"encoding/xml"
	"strings"

	"github.com/dimelords/idmllib/v2/pkg/common"
)

// Story represents an InDesign Story XML file.
// Stories contain text content with formatting information.
type Story struct {
	XMLName    xml.Name `xml:"http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging Story"`
	DOMVersion string   `xml:"DOMVersion,attr"`

	// The actual story content
	StoryElement StoryElement `xml:"Story"`
}

// ExtractText returns all text content from the story concatenated as a single string.
// Line breaks (<Br> elements) are converted to newline characters.
// This is a convenience method that navigates the story structure automatically.
func (s *Story) ExtractText() string {
	var buf strings.Builder
	for _, psr := range s.StoryElement.ParagraphStyleRanges {
		for _, csr := range psr.CharacterStyleRanges {
			for _, child := range csr.Children {
				if child.Content != nil {
					buf.WriteString(child.Content.Text)
				} else if child.Br != nil {
					buf.WriteString("\n")
				}
			}
		}
	}
	return buf.String()
}

// StoryElement represents the main Story element containing all content.
type StoryElement struct {
	XMLName xml.Name `xml:"Story"`

	// Identity
	Self string `xml:"Self,attr"`

	// Story metadata
	UserText         string `xml:"UserText,attr,omitempty"`         // "true"/"false"
	IsEndnoteStory   string `xml:"IsEndnoteStory,attr,omitempty"`   // "true"/"false"
	AppliedTOCStyle  string `xml:"AppliedTOCStyle,attr,omitempty"`  // Reference to TOC style
	TrackChanges     string `xml:"TrackChanges,attr,omitempty"`     // "true"/"false"
	StoryTitle       string `xml:"StoryTitle,attr,omitempty"`       // Story title
	AppliedNamedGrid string `xml:"AppliedNamedGrid,attr,omitempty"` // Reference to named grid

	// Story preferences
	StoryPreference *StoryPreference `xml:"StoryPreference,omitempty"`

	// InCopy export options
	InCopyExportOption *InCopyExportOption `xml:"InCopyExportOption,omitempty"`

	// Content - paragraph style ranges
	ParagraphStyleRanges []ParagraphStyleRange `xml:"ParagraphStyleRange"`

	// Catch-all for unknown elements
	OtherElements []common.RawXMLElement `xml:",any"`
}

// StoryPreference represents story-level preferences.
type StoryPreference struct {
	XMLName xml.Name `xml:"StoryPreference"`

	// Optical margin alignment
	OpticalMarginAlignment string `xml:"OpticalMarginAlignment,attr,omitempty"` // "true"/"false"
	OpticalMarginSize      string `xml:"OpticalMarginSize,attr,omitempty"`      // Point size

	// Frame type and orientation
	FrameType        string `xml:"FrameType,attr,omitempty"`        // "TextFrameType", etc.
	StoryOrientation string `xml:"StoryOrientation,attr,omitempty"` // "Horizontal"/"Vertical"
	StoryDirection   string `xml:"StoryDirection,attr,omitempty"`   // "LeftToRightDirection", etc.

	// Catch-all for unknown attributes/elements
	OtherElements []common.RawXMLElement `xml:",any"`
}

// InCopyExportOption represents InCopy export settings for the story.
type InCopyExportOption struct {
	XMLName xml.Name `xml:"InCopyExportOption"`

	IncludeGraphicProxies string `xml:"IncludeGraphicProxies,attr,omitempty"` // "true"/"false"
	IncludeAllResources   string `xml:"IncludeAllResources,attr,omitempty"`   // "true"/"false"

	// Catch-all for unknown attributes/elements
	OtherElements []common.RawXMLElement `xml:",any"`
}

// ParagraphStyleRange represents a range of paragraphs with the same paragraph style.
type ParagraphStyleRange struct {
	XMLName xml.Name `xml:"ParagraphStyleRange"`

	// Applied paragraph style reference
	AppliedParagraphStyle string `xml:"AppliedParagraphStyle,attr"`

	// Character style ranges within this paragraph
	CharacterStyleRanges []CharacterStyleRange `xml:"CharacterStyleRange"`

	// Catch-all for unknown elements
	OtherElements []common.RawXMLElement `xml:",any"`
}

// CharacterStyleRange represents a range of characters with the same character style.
// IMPORTANT: This struct uses custom marshaling to preserve the order of Content and Br elements.
type CharacterStyleRange struct {
	XMLName xml.Name `xml:"CharacterStyleRange"`

	// Applied character style reference
	AppliedCharacterStyle string `xml:"AppliedCharacterStyle,attr"`

	// Common character formatting attributes
	// These are optional and preserve InDesign's text formatting
	HorizontalScale string `xml:"HorizontalScale,attr,omitempty"` // Horizontal scaling percentage
	Tracking        string `xml:"Tracking,attr,omitempty"`        // Letter spacing/tracking value

	// Additional formatting attributes (catch-all)
	OtherAttrs []xml.Attr `xml:"-"` // Not used by encoding/xml, manually handled

	// Mixed content: Content and Br elements in order
	// This field stores both Content and Br elements in the order they appear
	Children []CharacterChild `xml:"-"` // Manually marshaled to preserve order
}

// CharacterChild represents either a Content element or a Br element
type CharacterChild struct {
	Content *Content              // If non-nil, this is a Content element
	Br      *Br                   // If non-nil, this is a Br element
	Other   *common.RawXMLElement // If non-nil, this is an unknown element
}

// Content represents actual text content.
type Content struct {
	XMLName xml.Name `xml:"Content"`
	Text    string   `xml:",chardata"`
}

// Br represents a line break element.
type Br struct {
	XMLName xml.Name `xml:"Br"`
}

// NewCharacterStyleRange creates a new CharacterStyleRange with the given style and content.
// This is a convenience constructor for backward compatibility with code that used struct literals.
// If appliedStyle is empty, it defaults to the no-style marker.
func NewCharacterStyleRange(appliedStyle string, contents []Content) CharacterStyleRange {
	if appliedStyle == "" {
		appliedStyle = "CharacterStyle/$ID/[No character style]"
	}

	csr := CharacterStyleRange{
		XMLName:               xml.Name{Local: "CharacterStyleRange"},
		AppliedCharacterStyle: appliedStyle,
	}

	// Convert Content array to Children with interleaved Br elements
	for _, content := range contents {
		csr.Children = append(csr.Children, CharacterChild{Content: &Content{
			XMLName: xml.Name{Local: "Content"},
			Text:    content.Text,
		}})
		csr.Children = append(csr.Children, CharacterChild{Br: &Br{XMLName: xml.Name{Local: "Br"}}})
	}

	return csr
}

// GetContent returns all Content elements in order (for backward compatibility).
// This allows existing code to continue accessing content without knowing about the new Children structure.
func (c *CharacterStyleRange) GetContent() []Content {
	var contents []Content
	for _, child := range c.Children {
		if child.Content != nil {
			contents = append(contents, *child.Content)
		}
	}
	return contents
}

// SetContent sets content elements, replacing any existing children with new Content + Br structure.
// This provides backward compatibility for code that builds CharacterStyleRanges programmatically.
// Line breaks are added after each Content element.
func (c *CharacterStyleRange) SetContent(contents []Content) {
	c.Children = nil
	for _, content := range contents {
		// Add content
		c.Children = append(c.Children, CharacterChild{Content: &Content{
			XMLName: xml.Name{Local: "Content"},
			Text:    content.Text,
		}})
		// Add line break after content
		c.Children = append(c.Children, CharacterChild{Br: &Br{XMLName: xml.Name{Local: "Br"}}})
	}
}

// AddContent appends a Content element followed by a Br (for backward compatibility).
func (c *CharacterStyleRange) AddContent(text string) {
	c.Children = append(c.Children, CharacterChild{Content: &Content{
		XMLName: xml.Name{Local: "Content"},
		Text:    text,
	}})
	c.Children = append(c.Children, CharacterChild{Br: &Br{XMLName: xml.Name{Local: "Br"}}})
}
