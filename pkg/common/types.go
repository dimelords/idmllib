// Package common contains shared types used across all IDML domain packages.
//
// These types are extracted from pkg/idml to avoid circular dependencies
// and provide a common foundation for document, spread, story, and resource packages.
//
// Key types:
//   - RawXMLElement: Forward-compatible catch-all for unknown XML elements
//   - Properties: Common properties container with Label key-value pairs
//   - GridDataInformation: Grid layout configuration shared by Document and Spread
package common

import (
	"encoding/xml"
)

// RawXMLElement represents an arbitrary XML element that hasn't been explicitly modeled yet.
// This allows forward compatibility by preserving unknown elements during marshal/unmarshal.
//
// Example usage in catch-all fields:
//
//	OtherElements []RawXMLElement `xml:",any"`
//
// This captures elements like KinsokuTable, MojikumiTable, TextVariable, etc.
// until we add explicit struct definitions for them.
type RawXMLElement struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:",any,attr"`
	Content []byte     `xml:",innerxml"`
}

// Properties represents the Properties container element.
// Properties hold metadata and configuration as Label key-value pairs.
type Properties struct {
	XMLName xml.Name `xml:"Properties"`

	PathGeometry *PathGeometry `xml:"PathGeometry,omitempty"`

	// Label contains key-value pairs
	Label *Label `xml:"Label,omitempty"`

	// Catch-all for other Properties children (AppliedMathMLSwatch, etc.)
	// that are not yet explicitly modeled
	OtherElements []RawXMLElement `xml:",any"`
}

// Label represents a container for key-value pairs.
type Label struct {
	XMLName       xml.Name       `xml:"Label"`
	KeyValuePairs []KeyValuePair `xml:"KeyValuePair"`
}

// KeyValuePair represents a single key-value metadata pair.
type KeyValuePair struct {
	XMLName xml.Name `xml:"KeyValuePair"`
	Key     string   `xml:"Key,attr"`
	Value   string   `xml:"Value,attr"`
}

// PathGeometry represents path geometry information.
type PathGeometry struct {
	GeometryPathType *GeometryPathType `xml:"GeometryPathType,omitempty"`
}

// GeometryPathType defines a geometric path with points.
type GeometryPathType struct {
	PathOpen       string          `xml:"PathOpen,attr,omitempty"`
	PathPointArray *PathPointArray `xml:"PathPointArray,omitempty"`
}

// PathPointArray contains an array of path points.
type PathPointArray struct {
	PathPoints []PathPointType `xml:"PathPointType"`
}

// PathPointType represents a single point in a path with anchor and direction handles.
type PathPointType struct {
	Anchor         string `xml:"Anchor,attr"`
	LeftDirection  string `xml:"LeftDirection,attr,omitempty"`
	RightDirection string `xml:"RightDirection,attr,omitempty"`
}

// GetAppliedFont extracts the AppliedFont value from Properties.OtherElements.
// Returns the font family name (e.g., "Polaris Condensed") or empty string if not found.
//
// AppliedFont is stored in the Properties element as:
//
//	<AppliedFont type="string">Polaris Condensed</AppliedFont>
func (p *Properties) GetAppliedFont() string {
	if p == nil {
		return ""
	}

	// Search OtherElements for <AppliedFont> element
	for _, elem := range p.OtherElements {
		if elem.XMLName.Local == "AppliedFont" {
			// Extract text content, trim whitespace
			content := string(elem.Content)
			// Parse the content to extract just the text (skip any nested XML)
			// For simple text content, this will work directly
			if len(content) > 0 {
				// Find text between > and <
				start := 0
				end := len(content)
				for i, ch := range content {
					if ch == '>' {
						start = i + 1
					} else if ch == '<' {
						end = i
						break
					}
				}
				if start < end {
					return content[start:end]
				}
				// If no angle brackets, it's plain text
				return content
			}
		}
	}

	return ""
}

// GetBasedOn extracts the BasedOn value from Properties.OtherElements.
// Returns the parent style ID or empty string if not found.
//
// BasedOn is stored in the Properties element as:
//
//	<BasedOn type="string">$ID/[No character style]</BasedOn>
func (p *Properties) GetBasedOn() string {
	if p == nil {
		return ""
	}

	// Search OtherElements for <BasedOn> element
	for _, elem := range p.OtherElements {
		if elem.XMLName.Local == "BasedOn" {
			// Extract text content, trim whitespace
			content := string(elem.Content)
			// Parse the content to extract just the text (skip any nested XML)
			if len(content) > 0 {
				// Find text between > and <
				start := 0
				end := len(content)
				for i, ch := range content {
					if ch == '>' {
						start = i + 1
					} else if ch == '<' {
						end = i
						break
					}
				}
				if start < end {
					return content[start:end]
				}
				// If no angle brackets, it's plain text
				return content
			}
		}
	}

	return ""
}

// GridDataInformation contains the detailed configuration for a grid.
// This defines character/line spacing, alignment, and typographic settings.
type GridDataInformation struct {
	XMLName xml.Name `xml:"GridDataInformation"`

	// Typography
	FontStyle string `xml:"FontStyle,attr,omitempty"` // Font style (e.g., "Roman")
	PointSize string `xml:"PointSize,attr,omitempty"` // Font size in points

	// Character Spacing (Aki = space in Japanese)
	CharacterAki string `xml:"CharacterAki,attr,omitempty"` // Space between characters
	LineAki      string `xml:"LineAki,attr,omitempty"`      // Space between lines

	// Scaling
	HorizontalScale string `xml:"HorizontalScale,attr,omitempty"` // Horizontal scale percentage
	VerticalScale   string `xml:"VerticalScale,attr,omitempty"`   // Vertical scale percentage

	// Alignment
	LineAlignment      string `xml:"LineAlignment,attr,omitempty"`      // Line alignment (e.g., "LeftOrTopLineJustify")
	GridAlignment      string `xml:"GridAlignment,attr,omitempty"`      // Grid alignment (e.g., "AlignEmCenter")
	CharacterAlignment string `xml:"CharacterAlignment,attr,omitempty"` // Character alignment (e.g., "AlignEmCenter")

	// Properties may contain AppliedFont and other settings
	Properties *Properties `xml:"Properties,omitempty"`

	// Catch-all for other GridDataInformation children
	OtherElements []RawXMLElement `xml:",any"`
}
