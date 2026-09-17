package common

import "encoding/xml"

// kindOther is the ChildOrder kind used for children kept as RawXMLElement.
const kindOther = "other"

// UnmarshalXML decodes a <Properties> element while recording the document
// order of its children, so typed children (PathGeometry, Label) keep their
// position among the many property elements that are preserved raw.
func (p *Properties) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	if d == nil {
		return Errorf("common", "unmarshal properties", "", "decoder is nil")
	}
	p.XMLName = start.Name
	p.childOrder = nil

	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "PathGeometry":
				var v PathGeometry
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				p.PathGeometry = &v
			case "Label":
				var v Label
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				p.Label = &v
			default:
				raw, err := DecodeRaw(d, t)
				if err != nil {
					return err
				}
				p.OtherElements = append(p.OtherElements, raw)
				p.childOrder.Record(kindOther)
				continue
			}
			p.childOrder.Record(t.Name.Local)
		case xml.EndElement:
			return nil
		}
	}
}

// MarshalXML encodes the <Properties> element, replaying the recorded child order.
func (p Properties) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name = xml.Name{Local: "Properties"}
	start.Attr = nil
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	kinds := []ChildKind{
		OptionalKind("PathGeometry", p.PathGeometry != nil, func(e *xml.Encoder) error {
			return e.Encode(p.PathGeometry)
		}),
		OptionalKind("Label", p.Label != nil, func(e *xml.Encoder) error {
			return e.Encode(p.Label)
		}),
		{Name: kindOther, Len: len(p.OtherElements), Encode: func(e *xml.Encoder, i int) error { return e.Encode(p.OtherElements[i]) }},
	}
	if err := EncodeChildren(e, p.childOrder, kinds); err != nil {
		return err
	}
	return e.EncodeToken(start.End())
}
