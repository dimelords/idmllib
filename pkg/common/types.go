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
	"fmt"
	"unicode"
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

// UnmarshalXML decodes the element as it is, after checking its name. encoding/xml resolves a prefixed name such as
// a:b to its namespace and local part and checks only the whole name, so it accepts <a:0 xmlns:a="…">, whose local
// part isn't a name of its own. Written back, the element would be <0 xmlns="…">, which no XML parser reads.
// Namespaces in XML requires the local part to be a name (an NCName), so such an element is refused.
func (r *RawXMLElement) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	if !isNCName(start.Name.Local) {
		return fmt.Errorf("element %q: its local name %q is not an XML name", start.Name.Space+":"+start.Name.Local, start.Name.Local)
	}
	type plain RawXMLElement // the default decoding, without this method
	var p plain
	if err := d.DecodeElement(&p, &start); err != nil {
		return err
	}
	*r = RawXMLElement(p)
	return nil
}

// isNCName reports whether s can be a name's local part: a letter or an underscore, then letters, digits, marks,
// underscores, hyphens and full stops, with no colon. encoding/xml has already checked the characters of the whole
// name; this is about where the local part starts.
func isNCName(s string) bool {
	for i, r := range s {
		switch {
		case r == ':':
			return false
		case unicode.IsLetter(r) || r == '_':
		case i == 0:
			return false
		case unicode.IsDigit(r) || r == '-' || r == '.' || r == '\u00B7' || unicode.In(r, unicode.Mn, unicode.Mc):
		default:
			return false
		}
	}
	return s != ""
}

// Properties represents the Properties container element.
// Properties hold metadata and configuration as Label key-value pairs.
type Properties struct {
	XMLName xml.Name `xml:"Properties"`

	PathGeometry *PathGeometry `xml:"PathGeometry,omitempty"`

	// Label contains key-value pairs
	Label *Label `xml:"Label,omitempty"`

	// childOrder records the document order of children so MarshalXML can
	// replay them; see ChildOrder.
	childOrder ChildOrder

	// Catch-all for other Properties children (AppliedMathMLSwatch, etc.)
	// that are not yet explicitly modeled
	OtherElements []RawXMLElement `xml:",any"`

	// OtherAttrs preserves attributes not modeled by a typed field, so
	// nothing is lost when the element is written back.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// Label represents a container for key-value pairs.
type Label struct {
	XMLName       xml.Name       `xml:"Label"`
	KeyValuePairs []KeyValuePair `xml:"KeyValuePair"`

	// OtherAttrs preserves attributes not modeled by a typed field, so
	// nothing is lost when the element is written back.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// KeyValuePair represents a single key-value metadata pair.
type KeyValuePair struct {
	XMLName xml.Name `xml:"KeyValuePair"`
	Key     string   `xml:"Key,attr"`
	Value   string   `xml:"Value,attr"`

	// OtherAttrs preserves attributes not modeled by a typed field, so
	// nothing is lost when the element is written back.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// PathGeometry represents path geometry information: an item's paths. Most items have one. A compound shape, such
// as a letter with a counter or an icon with holes, has several, each its own <GeometryPathType>; Paths lists them
// all, in document order.
type PathGeometry struct {
	// GeometryPathType is the item's first path.
	GeometryPathType *GeometryPathType `xml:"GeometryPathType,omitempty"`
	// MorePaths are the item's paths after the first, in document order.
	MorePaths []GeometryPathType `xml:"-"`

	// OtherElements preserves children not modeled above.
	OtherElements []RawXMLElement `xml:"-"`
	// OtherAttrs preserves attributes not modeled by a typed field, so
	// nothing is lost when the element is written back.
	OtherAttrs []xml.Attr `xml:",any,attr"`

	childOrder ChildOrder
}

// Paths returns the item's paths, the first and then MorePaths, as pointers into p so they can be changed in
// place; none when p is nil.
func (p *PathGeometry) Paths() []*GeometryPathType {
	if p == nil {
		return nil
	}
	var out []*GeometryPathType
	if p.GeometryPathType != nil {
		out = append(out, p.GeometryPathType)
	}
	for i := range p.MorePaths {
		out = append(out, &p.MorePaths[i])
	}
	return out
}

// UnmarshalXML decodes every <GeometryPathType>, the first into GeometryPathType and the rest into MorePaths.
// encoding/xml would decode them all into the one GeometryPathType, appending each path's points to the first's, so
// a compound shape came back as a single path.
func (p *PathGeometry) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	*p = PathGeometry{}
	UnmarshalAttrs(p, start.Attr)
	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "GeometryPathType" {
				var g GeometryPathType
				if err := d.DecodeElement(&g, &t); err != nil {
					return err
				}
				if p.GeometryPathType == nil {
					p.GeometryPathType = &g
				} else {
					p.MorePaths = append(p.MorePaths, g)
				}
				p.childOrder.Record("GeometryPathType")
				continue
			}
			raw, err := DecodeRaw(d, t)
			if err != nil {
				return err
			}
			p.OtherElements = append(p.OtherElements, raw)
			p.childOrder.Record(kindOther)
		case xml.EndElement:
			return nil
		}
	}
}

// MarshalXML writes the paths and any other children in the order they were read.
func (p PathGeometry) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Attr = MarshalAttrs(p)
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	paths := p.Paths()
	pathStart := xml.StartElement{Name: xml.Name{Local: "GeometryPathType"}}
	kinds := []ChildKind{
		{Name: "GeometryPathType", Len: len(paths), Encode: func(e *xml.Encoder, i int) error {
			return e.EncodeElement(paths[i], pathStart)
		}},
		{Name: kindOther, Len: len(p.OtherElements), Encode: func(e *xml.Encoder, i int) error {
			return e.Encode(p.OtherElements[i])
		}},
	}
	if err := EncodeChildren(e, p.childOrder, kinds); err != nil {
		return err
	}
	return e.EncodeToken(start.End())
}

// GeometryPathType defines a geometric path with points.
type GeometryPathType struct {
	PathOpen       string          `xml:"PathOpen,attr,omitempty"`
	PathPointArray *PathPointArray `xml:"PathPointArray,omitempty"`

	// OtherAttrs preserves attributes not modeled by a typed field, so
	// nothing is lost when the element is written back.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// PathPointArray contains an array of path points.
type PathPointArray struct {
	PathPoints []PathPointType `xml:"PathPointType"`

	// OtherAttrs preserves attributes not modeled by a typed field, so
	// nothing is lost when the element is written back.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// PathPointType represents a single point in a path with anchor and direction handles.
type PathPointType struct {
	Anchor         string `xml:"Anchor,attr"`
	LeftDirection  string `xml:"LeftDirection,attr,omitempty"`
	RightDirection string `xml:"RightDirection,attr,omitempty"`

	// OtherAttrs preserves attributes not modeled by a typed field, so
	// nothing is lost when the element is written back.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// GetAppliedFont extracts the AppliedFont value from Properties.OtherElements.
// Returns the font family name (e.g., "Minion Pro") or empty string if not found.
//
// AppliedFont is stored in the Properties element as:
//
//	<AppliedFont type="string">Minion Pro</AppliedFont>
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

	// OtherAttrs preserves attributes not modeled by a typed field, so
	// nothing is lost when the element is written back.
	OtherAttrs []xml.Attr `xml:",any,attr"`

	// childOrder records the document order of children so MarshalXML can
	// replay them; see common.ChildOrder.
	childOrder ChildOrder
}
