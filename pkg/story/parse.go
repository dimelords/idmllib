package story

import (
	"encoding/xml"

	"github.com/dimelords/idmllib/v2/pkg/common"
)

// ParseStory parses a Story XML file into a Story struct.
// It handles the IDML namespace and preserves unknown elements.
func ParseStory(data []byte) (*Story, error) {
	// Add nil check for input data
	if data == nil {
		return nil, common.Errorf("story", "parse story", "", "input data is nil")
	}

	// Add empty data check
	if len(data) == 0 {
		return nil, common.Errorf("story", "parse story", "", "input data is empty")
	}

	var story Story
	if err := xml.Unmarshal(data, &story); err != nil {
		return nil, common.WrapError("story", "parse story", err)
	}
	return &story, nil
}

// MarshalStory marshals a Story struct back to XML with proper formatting.
// It includes the XML declaration and proper namespace with idPkg prefix.
func MarshalStory(story *Story) ([]byte, error) {
	// Marshal the story with proper indentation
	data, err := xml.MarshalIndent(story, "", "\t")
	if err != nil {
		return nil, common.WrapError("story", "marshal story", err)
	}

	// Add XML declaration (use the original format for backward compatibility)
	result := []byte(xml.Header)
	result = append(result, data...)
	result = append(result, '\n')

	return result, nil
}

// UnmarshalXML implements custom unmarshaling for Story to handle idPkg namespace prefix.
func (s *Story) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// Add nil check for decoder
	if d == nil {
		return common.Errorf("story", "unmarshal story", "", "decoder is nil")
	}

	// Set the XMLName based on the start element
	s.XMLName = start.Name

	// Parse attributes
	for _, attr := range start.Attr {
		if attr.Name.Local == "DOMVersion" {
			s.DOMVersion = attr.Value
		}
	}

	// Parse child elements
	for {
		token, err := d.Token()
		if err != nil {
			return err
		}

		switch t := token.(type) {
		case xml.StartElement:
			if t.Name.Local == "Story" {
				if err := d.DecodeElement(&s.StoryElement, &t); err != nil {
					return err
				}
			}
		case xml.EndElement:
			return nil
		}
	}
}

// MarshalXML implements custom marshaling for Story to use idPkg namespace prefix.
func (s Story) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	// Set the element name with idPkg prefix
	start.Name = xml.Name{Local: "idPkg:Story"}

	// Add namespace declaration
	start.Attr = append(start.Attr, xml.Attr{
		Name:  xml.Name{Local: "xmlns:idPkg"},
		Value: "http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging",
	})

	// Add DOMVersion attribute
	start.Attr = append(start.Attr, xml.Attr{
		Name:  xml.Name{Local: "DOMVersion"},
		Value: s.DOMVersion,
	})

	// Write opening tag
	if err := e.EncodeToken(start); err != nil {
		return err
	}

	// Encode the story element
	if err := e.Encode(s.StoryElement); err != nil {
		return err
	}

	// Write closing tag
	if err := e.EncodeToken(xml.EndElement{Name: start.Name}); err != nil {
		return err
	}

	return nil
}

// UnmarshalXML implements custom unmarshaling for CharacterStyleRange to preserve element order.
func (c *CharacterStyleRange) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// Add nil check for decoder
	if d == nil {
		return common.Errorf("story", "unmarshal character style range", "", "decoder is nil")
	}

	// Set XMLName
	c.XMLName = start.Name

	// Parse attributes
	for _, attr := range start.Attr {
		switch attr.Name.Local {
		case "AppliedCharacterStyle":
			c.AppliedCharacterStyle = attr.Value
		case "HorizontalScale":
			c.HorizontalScale = attr.Value
		case "Tracking":
			c.Tracking = attr.Value
		default:
			// Store unknown attributes
			c.OtherAttrs = append(c.OtherAttrs, attr)
		}
	}

	// Parse child elements in order
	for {
		token, err := d.Token()
		if err != nil {
			return err
		}

		switch t := token.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "Content":
				var content Content
				if err := d.DecodeElement(&content, &t); err != nil {
					return err
				}
				c.Children = append(c.Children, CharacterChild{Content: &content})

			case "Br":
				// Br is self-closing, just consume the element
				if err := d.Skip(); err != nil {
					return err
				}
				c.Children = append(c.Children, CharacterChild{Br: &Br{XMLName: t.Name}})

			default:
				// Unknown element - store as RawXMLElement
				var raw common.RawXMLElement
				raw.XMLName = t.Name
				raw.Attrs = t.Attr

				// Read inner content
				if err := d.DecodeElement(&raw, &t); err != nil {
					return err
				}
				c.Children = append(c.Children, CharacterChild{Other: &raw})
			}

		case xml.EndElement:
			return nil
		}
	}
}

// MarshalXML implements custom marshaling for CharacterStyleRange to preserve element order.
func (c CharacterStyleRange) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	// Set element name
	start.Name = c.XMLName

	// Add attributes
	if c.AppliedCharacterStyle != "" {
		start.Attr = append(start.Attr, xml.Attr{
			Name:  xml.Name{Local: "AppliedCharacterStyle"},
			Value: c.AppliedCharacterStyle,
		})
	}
	if c.HorizontalScale != "" {
		start.Attr = append(start.Attr, xml.Attr{
			Name:  xml.Name{Local: "HorizontalScale"},
			Value: c.HorizontalScale,
		})
	}
	if c.Tracking != "" {
		start.Attr = append(start.Attr, xml.Attr{
			Name:  xml.Name{Local: "Tracking"},
			Value: c.Tracking,
		})
	}

	// Add other attributes
	start.Attr = append(start.Attr, c.OtherAttrs...)

	// Write opening tag
	if err := e.EncodeToken(start); err != nil {
		return err
	}

	// Write children in order
	for _, child := range c.Children {
		if child.Content != nil {
			if err := e.Encode(child.Content); err != nil {
				return err
			}
		} else if child.Br != nil {
			// Encode Br as self-closing tag
			brStart := xml.StartElement{Name: xml.Name{Local: "Br"}}
			if err := e.EncodeToken(brStart); err != nil {
				return err
			}
			if err := e.EncodeToken(xml.EndElement{Name: brStart.Name}); err != nil {
				return err
			}
		} else if child.Other != nil {
			if err := e.Encode(child.Other); err != nil {
				return err
			}
		}
	}

	// Write closing tag
	if err := e.EncodeToken(xml.EndElement{Name: start.Name}); err != nil {
		return err
	}

	return nil
}
