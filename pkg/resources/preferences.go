package resources

import (
	"encoding/xml"

	"github.com/dimelords/idmllib/v3/internal/xmlutil"
	"github.com/dimelords/idmllib/v3/pkg/common"
)

// PreferencesFile represents Resources/Preferences.xml: document, text, view,
// grid and export preferences. The individual preference elements are not
// modeled yet; they are preserved verbatim and in order in Elements.
type PreferencesFile struct {
	XMLName    xml.Name `xml:"-"`
	DOMVersion string   `xml:"DOMVersion,attr"`

	// Elements holds every preference element in document order.
	Elements []common.RawXMLElement `xml:",any"`
	// OtherAttrs preserves attributes of the wrapper element.
	OtherAttrs []xml.Attr `xml:",any,attr"`

	childOrder common.ChildOrder
}

// ParsePreferencesFile parses the contents of Resources/Preferences.xml.
func ParsePreferencesFile(data []byte) (*PreferencesFile, error) {
	if len(data) == 0 {
		return nil, common.Errorf("resources", "parse preferences", "", "input data is empty")
	}
	var prefs PreferencesFile
	if err := xml.Unmarshal(data, &prefs); err != nil {
		return nil, common.WrapError("resources", "parse preferences", err)
	}
	return &prefs, nil
}

// MarshalPreferencesFile serializes a PreferencesFile to XML with the standard
// IDML declaration.
func MarshalPreferencesFile(prefs *PreferencesFile) ([]byte, error) {
	if prefs == nil {
		return nil, common.Errorf("resources", "marshal preferences", "", "preferences is nil")
	}
	data, err := xmlutil.MarshalIndentWithHeader(prefs, "", "\t")
	if err != nil {
		return nil, common.WrapError("resources", "marshal preferences", err)
	}
	return data, nil
}

// UnmarshalXML decodes an <idPkg:Preferences> file recording child order.
func (x *PreferencesFile) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	if start.Name.Local != "Preferences" {
		return common.WrapError("resources", "unmarshal preferences", common.ErrInvalidFormat)
	}
	return common.UnmarshalOrdered(x, &x.childOrder, d, start)
}

// MarshalXML encodes the <idPkg:Preferences> file wrapper.
func (x *PreferencesFile) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	wrapper := xml.StartElement{
		Name: xml.Name{Local: "idPkg:Preferences"},
		Attr: []xml.Attr{{Name: xml.Name{Local: "xmlns:idPkg"}, Value: "http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging"}},
	}
	return common.MarshalOrdered(x, x.childOrder, e, wrapper)
}
