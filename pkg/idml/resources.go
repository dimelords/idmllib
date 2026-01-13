package idml

import (
	"bytes"
	"encoding/xml"

	"github.com/dimelords/idmllib/pkg/common"
)

// ResourceFile represents a generic resource XML file (Graphic, Fonts, Styles, Preferences).
// These files contain collections of resource definitions (colors, fonts, styles, etc.).
//
// The root element is <idPkg:ResourceType> (e.g., <idPkg:Graphic>) with the idPkg namespace.
type ResourceFile struct {
	// XMLName is not set directly - we handle it manually in MarshalXML/UnmarshalXML
	XMLName xml.Name `xml:"-"`

	// ResourceType is the type of resource (Graphic, Fonts, Styles, Preferences)
	ResourceType string `xml:"-"`

	// DOMVersion is the InDesign DOM version (e.g., "20.4")
	DOMVersion string `xml:"DOMVersion,attr"`

	// RawContent stores all child elements as raw XML
	// This preserves the complete structure without needing to model every element
	RawContent []byte `xml:",innerxml"`
}

// UnmarshalXML implements custom XML unmarshaling for ResourceFile.
// It handles the idPkg:ResourceType wrapper element and preserves the content.
func (r *ResourceFile) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// Store the resource type from element name
	r.ResourceType = start.Name.Local

	// Extract DOMVersion from attributes
	for _, attr := range start.Attr {
		if attr.Name.Local == "DOMVersion" {
			r.DOMVersion = attr.Value
			break
		}
	}

	// Read all content as raw XML
	var buf bytes.Buffer
	depth := 0

	for {
		tok, err := d.Token()
		if err != nil {
			return common.WrapError("idml", "unmarshal resource", err)
		}

		switch t := tok.(type) {
		case xml.StartElement:
			depth++
			buf.WriteString("<")
			buf.WriteString(t.Name.Local)
			for _, attr := range t.Attr {
				buf.WriteString(" ")
				buf.WriteString(attr.Name.Local)
				buf.WriteString(`="`)
				xml.EscapeText(&buf, []byte(attr.Value)) // nolint:errcheck
				buf.WriteString(`"`)
			}
			buf.WriteString(">")

		case xml.EndElement:
			depth--
			if depth < 0 {
				// End of root element
				r.RawContent = buf.Bytes()
				return nil
			}
			buf.WriteString("</")
			buf.WriteString(t.Name.Local)
			buf.WriteString(">")

		case xml.CharData:
			xml.EscapeText(&buf, t) // nolint:errcheck

		case xml.Comment:
			buf.WriteString("<!--")
			buf.Write(t)
			buf.WriteString("-->")

		case xml.ProcInst:
			buf.WriteString("<?")
			buf.WriteString(t.Target)
			buf.WriteString(" ")
			buf.Write(t.Inst)
			buf.WriteString("?>")
		}
	}
}

// MarshalXML implements custom XML marshaling for ResourceFile.
func (r *ResourceFile) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	// Create the idPkg:ResourceType wrapper element
	wrapper := xml.StartElement{
		Name: xml.Name{Local: "idPkg:" + r.ResourceType},
		Attr: []xml.Attr{
			{Name: xml.Name{Local: "xmlns:idPkg"}, Value: "http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging"},
			{Name: xml.Name{Local: "DOMVersion"}, Value: r.DOMVersion},
		},
	}

	// Start the wrapper element
	if err := e.EncodeToken(wrapper); err != nil {
		return err
	}

	// Write raw content directly
	if len(r.RawContent) > 0 {
		// Parse and re-encode to ensure proper formatting
		decoder := xml.NewDecoder(bytes.NewReader(r.RawContent))
		for {
			tok, err := decoder.Token()
			if err != nil {
				break
			}
			if err := e.EncodeToken(xml.CopyToken(tok)); err != nil {
				return err
			}
		}
	}

	// End the wrapper element
	if err := e.EncodeToken(wrapper.End()); err != nil {
		return err
	}

	return nil
}

// ParseResourceFile parses a resource XML file into a ResourceFile struct.
func ParseResourceFile(data []byte) (*ResourceFile, error) {
	// Add nil check for input data
	if data == nil {
		return nil, common.Errorf("idml", "parse resource", "", "input data is nil")
	}

	// Add empty data check
	if len(data) == 0 {
		return nil, common.Errorf("idml", "parse resource", "", "input data is empty")
	}

	var resource ResourceFile
	if err := xml.Unmarshal(data, &resource); err != nil {
		return nil, common.WrapError("idml", "parse resource", err)
	}
	return &resource, nil
}

// MarshalResourceFile marshals a ResourceFile struct back to XML with proper formatting.
func MarshalResourceFile(resource *ResourceFile) ([]byte, error) {
	var buf bytes.Buffer

	// Add XML declaration
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	buf.WriteByte('\n')

	// Marshal the resource
	resourceXML, err := xml.MarshalIndent(resource, "", "\t")
	if err != nil {
		return nil, common.WrapError("idml", "marshal resource", err)
	}

	buf.Write(resourceXML)
	buf.WriteByte('\n')

	return buf.Bytes(), nil
}
