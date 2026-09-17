package resources

import (
	"encoding/xml"

	"github.com/dimelords/idmllib/v2/pkg/common"
)

// Style groups hold styles and nested groups interleaved in document order.
// InDesign shows styles in the Styles panel in this order, so the order must
// survive a roundtrip. The three group types that support nesting record it
// with a ChildOrder and replay it on marshal.

// decodeGroupChildren drives the token loop shared by the group types.
// onStart is called for each child start element and must consume it; it
// returns the kind to record.
func decodeGroupChildren(d *xml.Decoder, order *common.ChildOrder, onStart func(t xml.StartElement) (string, error)) error {
	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			kind, err := onStart(t)
			if err != nil {
				return err
			}
			order.Record(kind)
		case xml.EndElement:
			return nil
		}
	}
}

// groupStart returns the start element to use when marshaling a group: the
// caller-provided name (from a field tag or EncodeElement) when present,
// otherwise fallback.
func groupStart(start xml.StartElement, attrs []xml.Attr, fallback string) xml.StartElement {
	if start.Name.Local == "" {
		start.Name = xml.Name{Local: fallback}
	}
	start.Attr = attrs
	return start
}

func named(name string, v any) func(*xml.Encoder, int) error {
	return func(e *xml.Encoder, _ int) error {
		return e.EncodeElement(v, xml.StartElement{Name: xml.Name{Local: name}})
	}
}

// UnmarshalXML decodes a character style group recording child order.
func (g *CharacterStyleGroup) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	if d == nil {
		return common.Errorf("resources", "unmarshal character style group", "", "decoder is nil")
	}
	common.UnmarshalAttrs(g, start.Attr)
	g.childOrder = nil
	return decodeGroupChildren(d, &g.childOrder, func(t xml.StartElement) (string, error) {
		switch t.Name.Local {
		case "CharacterStyle":
			var v CharacterStyle
			if err := d.DecodeElement(&v, &t); err != nil {
				return "", err
			}
			g.CharacterStyles = append(g.CharacterStyles, v)
		case "CharacterStyleGroup":
			var v CharacterStyleGroup
			if err := d.DecodeElement(&v, &t); err != nil {
				return "", err
			}
			g.NestedGroups = append(g.NestedGroups, v)
		default:
			raw, err := common.DecodeRaw(d, t)
			if err != nil {
				return "", err
			}
			g.OtherElements = append(g.OtherElements, raw)
			return kindOther, nil
		}
		return t.Name.Local, nil
	})
}

// MarshalXML encodes a character style group replaying the recorded child order.
func (g CharacterStyleGroup) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start = groupStart(start, common.MarshalAttrs(g), "CharacterStyleGroup")
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	kinds := []common.ChildKind{
		{Name: "CharacterStyle", Len: len(g.CharacterStyles), Encode: func(e *xml.Encoder, i int) error {
			return named("CharacterStyle", g.CharacterStyles[i])(e, i)
		}},
		{Name: "CharacterStyleGroup", Len: len(g.NestedGroups), Encode: func(e *xml.Encoder, i int) error {
			return named("CharacterStyleGroup", g.NestedGroups[i])(e, i)
		}},
		{Name: kindOther, Len: len(g.OtherElements), Encode: func(e *xml.Encoder, i int) error {
			return e.Encode(g.OtherElements[i])
		}},
	}
	if err := common.EncodeChildren(e, g.childOrder, kinds); err != nil {
		return err
	}
	return e.EncodeToken(start.End())
}

// UnmarshalXML decodes a paragraph style group recording child order.
func (g *ParagraphStyleGroup) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	if d == nil {
		return common.Errorf("resources", "unmarshal paragraph style group", "", "decoder is nil")
	}
	common.UnmarshalAttrs(g, start.Attr)
	g.childOrder = nil
	return decodeGroupChildren(d, &g.childOrder, func(t xml.StartElement) (string, error) {
		switch t.Name.Local {
		case "ParagraphStyle":
			var v ParagraphStyle
			if err := d.DecodeElement(&v, &t); err != nil {
				return "", err
			}
			g.ParagraphStyles = append(g.ParagraphStyles, v)
		case "ParagraphStyleGroup":
			var v ParagraphStyleGroup
			if err := d.DecodeElement(&v, &t); err != nil {
				return "", err
			}
			g.NestedGroups = append(g.NestedGroups, v)
		default:
			raw, err := common.DecodeRaw(d, t)
			if err != nil {
				return "", err
			}
			g.OtherElements = append(g.OtherElements, raw)
			return kindOther, nil
		}
		return t.Name.Local, nil
	})
}

// MarshalXML encodes a paragraph style group replaying the recorded child order.
func (g ParagraphStyleGroup) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start = groupStart(start, common.MarshalAttrs(g), "ParagraphStyleGroup")
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	kinds := []common.ChildKind{
		{Name: "ParagraphStyle", Len: len(g.ParagraphStyles), Encode: func(e *xml.Encoder, i int) error {
			return named("ParagraphStyle", g.ParagraphStyles[i])(e, i)
		}},
		{Name: "ParagraphStyleGroup", Len: len(g.NestedGroups), Encode: func(e *xml.Encoder, i int) error {
			return named("ParagraphStyleGroup", g.NestedGroups[i])(e, i)
		}},
		{Name: kindOther, Len: len(g.OtherElements), Encode: func(e *xml.Encoder, i int) error {
			return e.Encode(g.OtherElements[i])
		}},
	}
	if err := common.EncodeChildren(e, g.childOrder, kinds); err != nil {
		return err
	}
	return e.EncodeToken(start.End())
}

// UnmarshalXML decodes an object style group recording child order.
func (g *ObjectStyleGroup) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	if d == nil {
		return common.Errorf("resources", "unmarshal object style group", "", "decoder is nil")
	}
	common.UnmarshalAttrs(g, start.Attr)
	g.childOrder = nil
	return decodeGroupChildren(d, &g.childOrder, func(t xml.StartElement) (string, error) {
		switch t.Name.Local {
		case "ObjectStyle":
			var v ObjectStyle
			if err := d.DecodeElement(&v, &t); err != nil {
				return "", err
			}
			g.ObjectStyles = append(g.ObjectStyles, v)
		case "ObjectStyleGroup":
			var v ObjectStyleGroup
			if err := d.DecodeElement(&v, &t); err != nil {
				return "", err
			}
			g.NestedGroups = append(g.NestedGroups, v)
		default:
			raw, err := common.DecodeRaw(d, t)
			if err != nil {
				return "", err
			}
			g.OtherElements = append(g.OtherElements, raw)
			return kindOther, nil
		}
		return t.Name.Local, nil
	})
}

// MarshalXML encodes an object style group replaying the recorded child order.
func (g ObjectStyleGroup) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start = groupStart(start, common.MarshalAttrs(g), "ObjectStyleGroup")
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	kinds := []common.ChildKind{
		{Name: "ObjectStyle", Len: len(g.ObjectStyles), Encode: func(e *xml.Encoder, i int) error {
			return named("ObjectStyle", g.ObjectStyles[i])(e, i)
		}},
		{Name: "ObjectStyleGroup", Len: len(g.NestedGroups), Encode: func(e *xml.Encoder, i int) error {
			return named("ObjectStyleGroup", g.NestedGroups[i])(e, i)
		}},
		{Name: kindOther, Len: len(g.OtherElements), Encode: func(e *xml.Encoder, i int) error {
			return e.Encode(g.OtherElements[i])
		}},
	}
	if err := common.EncodeChildren(e, g.childOrder, kinds); err != nil {
		return err
	}
	return e.EncodeToken(start.End())
}
