package resources

import (
	"encoding/xml"

	"github.com/dimelords/idmllib/v3/internal/xmlutil"
	"github.com/dimelords/idmllib/v3/pkg/common"
)

// TagsFile represents XML/Tags.xml: the XML tags available for structured
// content. A story's XMLElement refers to one of these by its Self attribute,
// for example MarkupTag="XMLTag/story".
type TagsFile struct {
	XMLName    xml.Name `xml:"-"`
	DOMVersion string   `xml:"DOMVersion,attr"`

	XMLTags []XMLTag `xml:"XMLTag"`

	// Elements preserves any other child of the file.
	Elements []common.RawXMLElement `xml:",any"`
	// OtherAttrs preserves attributes of the wrapper element.
	OtherAttrs []xml.Attr `xml:",any,attr"`

	childOrder common.ChildOrder
}

// XMLTag is one tag definition, e.g. <XMLTag Self="XMLTag/story" Name="story">.
type XMLTag struct {
	XMLName xml.Name `xml:"XMLTag"`
	Self    string   `xml:"Self,attr,omitempty"`
	Name    string   `xml:"Name,attr,omitempty"`

	Properties    *common.Properties     `xml:"Properties,omitempty"`
	OtherElements []common.RawXMLElement `xml:",any"`
	OtherAttrs    []xml.Attr             `xml:",any,attr"`

	childOrder common.ChildOrder
}

// Tag returns the tag with the given Self reference, e.g. "XMLTag/story".
func (f *TagsFile) Tag(self string) (*XMLTag, bool) {
	for i := range f.XMLTags {
		if f.XMLTags[i].Self == self {
			return &f.XMLTags[i], true
		}
	}
	return nil, false
}

// ParseTagsFile parses the contents of XML/Tags.xml.
func ParseTagsFile(data []byte) (*TagsFile, error) {
	if len(data) == 0 {
		return nil, common.Errorf("resources", "parse tags", "", "input data is empty")
	}
	var tags TagsFile
	if err := xml.Unmarshal(data, &tags); err != nil {
		return nil, common.WrapError("resources", "parse tags", err)
	}
	return &tags, nil
}

// MarshalTagsFile serializes a TagsFile to XML with the standard IDML declaration.
func MarshalTagsFile(tags *TagsFile) ([]byte, error) {
	if tags == nil {
		return nil, common.Errorf("resources", "marshal tags", "", "tags is nil")
	}
	data, err := xmlutil.MarshalIndentWithHeader(tags, "", "\t")
	if err != nil {
		return nil, common.WrapError("resources", "marshal tags", err)
	}
	return data, nil
}

// UnmarshalXML decodes an <idPkg:Tags> file recording child order.
func (f *TagsFile) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	if start.Name.Local != "Tags" {
		return common.WrapError("resources", "unmarshal tags", common.ErrInvalidFormat)
	}
	return common.UnmarshalOrdered(f, &f.childOrder, d, start)
}

// MarshalXML encodes the <idPkg:Tags> file wrapper.
func (f *TagsFile) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	wrapper := xml.StartElement{
		Name: xml.Name{Local: "idPkg:Tags"},
		Attr: []xml.Attr{{Name: xml.Name{Local: "xmlns:idPkg"}, Value: "http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging"}},
	}
	return common.MarshalOrdered(f, f.childOrder, e, wrapper)
}

// UnmarshalXML decodes an <XMLTag> recording child order.
func (t *XMLTag) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return common.UnmarshalOrdered(t, &t.childOrder, d, start)
}

// MarshalXML encodes an <XMLTag> replaying the recorded child order.
func (t XMLTag) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return common.MarshalOrdered(t, t.childOrder, e, start)
}
