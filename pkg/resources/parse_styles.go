package resources

import (
	"encoding/xml"

	"github.com/dimelords/idmllib/internal/xmlutil"
	"github.com/dimelords/idmllib/pkg/common"
)

// ParseStylesFile parses a Styles.xml file into a StylesFile struct.
func ParseStylesFile(data []byte) (*StylesFile, error) {
	// Add nil check for input data
	if data == nil {
		return nil, common.Errorf("resources", "parse styles", "", "input data is nil")
	}

	// Add empty data check
	if len(data) == 0 {
		return nil, common.Errorf("resources", "parse styles", "", "input data is empty")
	}

	var styles StylesFile
	if err := xml.Unmarshal(data, &styles); err != nil {
		return nil, common.WrapError("resources", "parse styles", err)
	}
	return &styles, nil
}

// MarshalStylesFile marshals a StylesFile struct back to XML with proper formatting.
func MarshalStylesFile(styles *StylesFile) ([]byte, error) {
	return xmlutil.MarshalIndentWithHeader(styles, "", "\t")
}

// UnmarshalXML implements custom XML unmarshaling for StylesFile.
func (s *StylesFile) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// Add nil check for decoder
	if d == nil {
		return common.Errorf("resources", "unmarshal styles", "", "decoder is nil")
	}

	// Verify we're parsing an idPkg:Styles element
	if start.Name.Local != "Styles" {
		return common.WrapError("resources", "unmarshal styles", common.ErrInvalidFormat)
	}

	// Extract DOMVersion from idPkg:Styles attributes
	for _, attr := range start.Attr {
		if attr.Name.Local == "DOMVersion" {
			s.DOMVersion = attr.Value
			break
		}
	}

	// Define a temporary struct for unmarshaling the inner content
	type stylesContent struct {
		RootCharacterStyleGroup *CharacterStyleGroup   `xml:"RootCharacterStyleGroup,omitempty"`
		RootParagraphStyleGroup *ParagraphStyleGroup   `xml:"RootParagraphStyleGroup,omitempty"`
		RootCellStyleGroup      *CellStyleGroup        `xml:"RootCellStyleGroup,omitempty"`
		RootTableStyleGroup     *TableStyleGroup       `xml:"RootTableStyleGroup,omitempty"`
		RootObjectStyleGroup    *ObjectStyleGroup      `xml:"RootObjectStyleGroup,omitempty"`
		TOCStyles               []TOCStyle             `xml:"TOCStyle,omitempty"`
		OtherElements           []common.RawXMLElement `xml:",any"`
	}

	var content stylesContent
	if err := d.DecodeElement(&content, &start); err != nil {
		return common.WrapError("resources", "unmarshal styles content", err)
	}

	// Copy parsed content to StylesFile
	s.RootCharacterStyleGroup = content.RootCharacterStyleGroup
	s.RootParagraphStyleGroup = content.RootParagraphStyleGroup
	s.RootCellStyleGroup = content.RootCellStyleGroup
	s.RootTableStyleGroup = content.RootTableStyleGroup
	s.RootObjectStyleGroup = content.RootObjectStyleGroup
	s.TOCStyles = content.TOCStyles
	s.OtherElements = content.OtherElements

	return nil
}

// MarshalXML implements custom XML marshaling for StylesFile.
func (s *StylesFile) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	// Create the idPkg:Styles wrapper element
	wrapper := xml.StartElement{
		Name: xml.Name{Local: "idPkg:Styles"},
		Attr: []xml.Attr{
			{Name: xml.Name{Local: "xmlns:idPkg"}, Value: "http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging"},
			{Name: xml.Name{Local: "DOMVersion"}, Value: s.DOMVersion},
		},
	}

	// Start the wrapper element
	if err := e.EncodeToken(wrapper); err != nil {
		return err
	}

	// Encode all style groups
	if s.RootCharacterStyleGroup != nil {
		if err := e.EncodeElement(s.RootCharacterStyleGroup, xml.StartElement{Name: xml.Name{Local: "RootCharacterStyleGroup"}}); err != nil {
			return err
		}
	}

	if s.RootParagraphStyleGroup != nil {
		if err := e.EncodeElement(s.RootParagraphStyleGroup, xml.StartElement{Name: xml.Name{Local: "RootParagraphStyleGroup"}}); err != nil {
			return err
		}
	}

	if s.RootCellStyleGroup != nil {
		if err := e.EncodeElement(s.RootCellStyleGroup, xml.StartElement{Name: xml.Name{Local: "RootCellStyleGroup"}}); err != nil {
			return err
		}
	}

	if s.RootTableStyleGroup != nil {
		if err := e.EncodeElement(s.RootTableStyleGroup, xml.StartElement{Name: xml.Name{Local: "RootTableStyleGroup"}}); err != nil {
			return err
		}
	}

	if s.RootObjectStyleGroup != nil {
		if err := e.EncodeElement(s.RootObjectStyleGroup, xml.StartElement{Name: xml.Name{Local: "RootObjectStyleGroup"}}); err != nil {
			return err
		}
	}

	// Encode TOC styles
	for _, toc := range s.TOCStyles {
		if err := e.EncodeElement(&toc, xml.StartElement{Name: xml.Name{Local: "TOCStyle"}}); err != nil {
			return err
		}
	}

	// Encode other elements
	for _, elem := range s.OtherElements {
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
