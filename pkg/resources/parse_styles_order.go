package resources

import (
	"encoding/xml"

	"github.com/dimelords/idmllib/v3/pkg/common"
)

// kindOther is the ChildOrder kind used for children kept as RawXMLElement.
const kindOther = "other"

// UnmarshalXML decodes an <idPkg:Styles> file while recording the document
// order of its top-level children (style groups, TOC styles and anything
// preserved raw such as TrapPreset), so they are written back in the same
// order.
func (s *StylesFile) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	if d == nil {
		return common.Errorf("resources", "unmarshal styles", "", "decoder is nil")
	}
	if start.Name.Local != "Styles" {
		return common.WrapError("resources", "unmarshal styles", common.ErrInvalidFormat)
	}
	s.XMLName = start.Name
	common.UnmarshalAttrs(s, start.Attr)
	s.childOrder = nil

	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "RootCharacterStyleGroup":
				var v CharacterStyleGroup
				if err := d.DecodeElement(&v, &t); err != nil {
					return common.WrapError("resources", "unmarshal styles content", err)
				}
				s.RootCharacterStyleGroup = &v
			case "RootParagraphStyleGroup":
				var v ParagraphStyleGroup
				if err := d.DecodeElement(&v, &t); err != nil {
					return common.WrapError("resources", "unmarshal styles content", err)
				}
				s.RootParagraphStyleGroup = &v
			case "RootCellStyleGroup":
				var v CellStyleGroup
				if err := d.DecodeElement(&v, &t); err != nil {
					return common.WrapError("resources", "unmarshal styles content", err)
				}
				s.RootCellStyleGroup = &v
			case "RootTableStyleGroup":
				var v TableStyleGroup
				if err := d.DecodeElement(&v, &t); err != nil {
					return common.WrapError("resources", "unmarshal styles content", err)
				}
				s.RootTableStyleGroup = &v
			case "RootObjectStyleGroup":
				var v ObjectStyleGroup
				if err := d.DecodeElement(&v, &t); err != nil {
					return common.WrapError("resources", "unmarshal styles content", err)
				}
				s.RootObjectStyleGroup = &v
			case "TOCStyle":
				var v TOCStyle
				if err := d.DecodeElement(&v, &t); err != nil {
					return common.WrapError("resources", "unmarshal styles content", err)
				}
				s.TOCStyles = append(s.TOCStyles, v)
			default:
				raw, err := common.DecodeRaw(d, t)
				if err != nil {
					return common.WrapError("resources", "unmarshal styles content", err)
				}
				s.OtherElements = append(s.OtherElements, raw)
				s.childOrder.Record(kindOther)
				continue
			}
			s.childOrder.Record(t.Name.Local)
		case xml.EndElement:
			return nil
		}
	}
}

// MarshalXML encodes the <idPkg:Styles> file, replaying the recorded child order.
func (s *StylesFile) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	wrapper := xml.StartElement{
		Name: xml.Name{Local: "idPkg:Styles"},
		Attr: []xml.Attr{
			{Name: xml.Name{Local: "xmlns:idPkg"}, Value: "http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging"},
			{Name: xml.Name{Local: "DOMVersion"}, Value: s.DOMVersion},
		},
	}
	wrapper.Attr = append(wrapper.Attr, s.OtherAttrs...)
	if err := e.EncodeToken(wrapper); err != nil {
		return err
	}

	named := func(name string, v any) func(*xml.Encoder) error {
		return func(e *xml.Encoder) error {
			return e.EncodeElement(v, xml.StartElement{Name: xml.Name{Local: name}})
		}
	}
	kinds := []common.ChildKind{
		common.OptionalKind("RootCharacterStyleGroup", s.RootCharacterStyleGroup != nil, named("RootCharacterStyleGroup", s.RootCharacterStyleGroup)),
		common.OptionalKind("RootParagraphStyleGroup", s.RootParagraphStyleGroup != nil, named("RootParagraphStyleGroup", s.RootParagraphStyleGroup)),
		common.OptionalKind("RootCellStyleGroup", s.RootCellStyleGroup != nil, named("RootCellStyleGroup", s.RootCellStyleGroup)),
		common.OptionalKind("RootTableStyleGroup", s.RootTableStyleGroup != nil, named("RootTableStyleGroup", s.RootTableStyleGroup)),
		common.OptionalKind("RootObjectStyleGroup", s.RootObjectStyleGroup != nil, named("RootObjectStyleGroup", s.RootObjectStyleGroup)),
		{Name: "TOCStyle", Len: len(s.TOCStyles), Encode: func(e *xml.Encoder, i int) error {
			return e.EncodeElement(&s.TOCStyles[i], xml.StartElement{Name: xml.Name{Local: "TOCStyle"}})
		}},
		{Name: kindOther, Len: len(s.OtherElements), Encode: func(e *xml.Encoder, i int) error {
			return e.EncodeElement(&s.OtherElements[i], xml.StartElement{Name: s.OtherElements[i].XMLName})
		}},
	}
	if err := common.EncodeChildren(e, s.childOrder, kinds); err != nil {
		return err
	}
	return e.EncodeToken(wrapper.End())
}
