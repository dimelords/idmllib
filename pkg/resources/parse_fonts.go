package resources

import (
	"encoding/xml"

	"github.com/dimelords/idmllib/internal/xmlutil"
	"github.com/dimelords/idmllib/pkg/common"
)

// ParseFontsFile parses a Fonts.xml file into a FontsFile struct.
func ParseFontsFile(data []byte) (*FontsFile, error) {
	// Add nil check for input data
	if data == nil {
		return nil, common.Errorf("resources", "parse fonts", "", "input data is nil")
	}

	// Add empty data check
	if len(data) == 0 {
		return nil, common.Errorf("resources", "parse fonts", "", "input data is empty")
	}

	var fonts FontsFile
	if err := xml.Unmarshal(data, &fonts); err != nil {
		return nil, common.WrapError("resources", "parse fonts", err)
	}
	return &fonts, nil
}

// MarshalFontsFile marshals a FontsFile struct back to XML with proper formatting.
func MarshalFontsFile(fonts *FontsFile) ([]byte, error) {
	return xmlutil.MarshalIndentWithHeader(fonts, "", "\t")
}

// UnmarshalXML implements custom XML unmarshaling for FontsFile.
func (f *FontsFile) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// Add nil check for decoder
	if d == nil {
		return common.Errorf("resources", "unmarshal fonts", "", "decoder is nil")
	}

	// Verify we're parsing an idPkg:Fonts element
	if start.Name.Local != "Fonts" {
		return common.WrapError("resources", "unmarshal fonts", common.ErrInvalidFormat)
	}

	// Extract DOMVersion from idPkg:Fonts attributes
	for _, attr := range start.Attr {
		if attr.Name.Local == "DOMVersion" {
			f.DOMVersion = attr.Value
			break
		}
	}

	// Define a temporary struct for unmarshaling the inner content
	type fontsContent struct {
		FontFamilies   []FontFamily           `xml:"FontFamily,omitempty"`
		CompositeFonts []CompositeFont        `xml:"CompositeFont,omitempty"`
		OtherElements  []common.RawXMLElement `xml:",any"`
	}

	var content fontsContent
	if err := d.DecodeElement(&content, &start); err != nil {
		return common.WrapError("resources", "unmarshal fonts content", err)
	}

	// Copy parsed content to FontsFile
	f.FontFamilies = content.FontFamilies
	f.CompositeFonts = content.CompositeFonts
	f.OtherElements = content.OtherElements

	return nil
}

// MarshalXML implements custom XML marshaling for FontsFile.
func (f *FontsFile) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	// Create the idPkg:Fonts wrapper element
	wrapper := xml.StartElement{
		Name: xml.Name{Local: "idPkg:Fonts"},
		Attr: []xml.Attr{
			{Name: xml.Name{Local: "xmlns:idPkg"}, Value: "http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging"},
			{Name: xml.Name{Local: "DOMVersion"}, Value: f.DOMVersion},
		},
	}

	// Start the wrapper element
	if err := e.EncodeToken(wrapper); err != nil {
		return err
	}

	// Encode all font families
	for _, family := range f.FontFamilies {
		if err := e.EncodeElement(&family, xml.StartElement{Name: xml.Name{Local: "FontFamily"}}); err != nil {
			return err
		}
	}

	// Encode all composite fonts
	for _, composite := range f.CompositeFonts {
		if err := e.EncodeElement(&composite, xml.StartElement{Name: xml.Name{Local: "CompositeFont"}}); err != nil {
			return err
		}
	}

	// Encode other elements
	for _, elem := range f.OtherElements {
		if err := e.EncodeElement(&elem, xml.StartElement{Name: elem.XMLName}); err != nil {
			return err
		}
	}

	// End the wrapper element
	if err := e.EncodeToken(wrapper.End()); err != nil {
		return err
	}

	return nil
}
