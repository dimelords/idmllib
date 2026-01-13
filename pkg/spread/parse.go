package spread

import (
	"encoding/xml"

	"github.com/dimelords/idmllib/v2/internal/xmlutil"
	"github.com/dimelords/idmllib/v2/pkg/common"
)

// UnmarshalXML implements custom XML unmarshaling for Spread.
func (s *Spread) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// Add nil check for decoder
	if d == nil {
		return common.Errorf("spread", "unmarshal spread", "", "decoder is nil")
	}

	// Verify we're parsing an idPkg:Spread element
	if start.Name.Local != "Spread" {
		return common.WrapError("spread", "parse", common.ErrInvalidFormat)
	}

	// Extract DOMVersion from idPkg:Spread attributes
	for _, attr := range start.Attr {
		if attr.Name.Local == "DOMVersion" {
			s.DOMVersion = attr.Value
			break
		}
	}

	// Read tokens until we find the inner <Spread> element or hit the end
	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "Spread" {
				// Found the inner Spread element - unmarshal it
				if err := d.DecodeElement(&s.InnerSpread, &t); err != nil {
					return err
				}
			}

		case xml.EndElement:
			if t.Name.Local == "Spread" && t.Name.Space == start.Name.Space {
				// End of idPkg:Spread element
				return nil
			}
		}
	}
}

// MarshalXML implements custom XML marshaling for Spread.
func (s *Spread) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	// Create the idPkg:Spread wrapper element
	wrapper := xml.StartElement{
		Name: xml.Name{Local: "idPkg:Spread"},
		Attr: []xml.Attr{
			{Name: xml.Name{Local: "xmlns:idPkg"}, Value: "http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging"},
			{Name: xml.Name{Local: "DOMVersion"}, Value: s.DOMVersion},
		},
	}

	// Start the wrapper element
	if err := e.EncodeToken(wrapper); err != nil {
		return common.WrapError("spread", "marshal spread", err)
	}

	// Marshal the inner Spread element
	innerStart := xml.StartElement{Name: xml.Name{Local: "Spread"}}
	if err := e.EncodeElement(&s.InnerSpread, innerStart); err != nil {
		return common.WrapError("spread", "marshal spread", err)
	}

	// End the wrapper element
	if err := e.EncodeToken(wrapper.End()); err != nil {
		return common.WrapError("spread", "marshal spread", err)
	}

	return nil
}

// ParseSpread parses a Spread XML file into a Spread struct.
func ParseSpread(data []byte) (*Spread, error) {
	// Add nil check for input data
	if data == nil {
		return nil, common.Errorf("spread", "parse spread", "", "input data is nil")
	}

	// Add empty data check
	if len(data) == 0 {
		return nil, common.Errorf("spread", "parse spread", "", "input data is empty")
	}

	var spread Spread
	if err := xml.Unmarshal(data, &spread); err != nil {
		return nil, common.WrapError("spread", "parse spread", err)
	}
	return &spread, nil
}

// MarshalSpread marshals a Spread struct back to XML with proper formatting.
func MarshalSpread(spread *Spread) ([]byte, error) {
	// Add XML declaration and marshal with indentation
	data, err := xmlutil.MarshalIndentWithHeader(spread, "", "\t")
	if err != nil {
		return nil, common.WrapError("spread", "marshal spread", err)
	}

	return data, nil
}
