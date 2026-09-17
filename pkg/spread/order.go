package spread

import (
	"encoding/xml"

	"github.com/dimelords/idmllib/v2/pkg/common"
)

// kindOther is the ChildOrder kind used for children kept as RawXMLElement,
// which includes page item types not yet modeled (Button, EPS, ...).
const kindOther = "other"

// elementName returns the local element name this spread was read from:
// "Spread" for regular spreads and "MasterSpread" for master spreads.
// It defaults to "Spread" for elements created programmatically.
func (s *Spread) elementName() string {
	if s.XMLName.Local != "" {
		return s.XMLName.Local
	}
	return "Spread"
}

// IsMaster reports whether this element was read from a <MasterSpread>.
func (s *Spread) IsMaster() bool {
	return s.XMLName.Local == "MasterSpread"
}

// UnmarshalXML decodes a <Spread> or <MasterSpread> element while recording
// the document order of its children. In IDML the order of page items on a
// spread is their stacking order, so it must survive a roundtrip.
func (s *Spread) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	if d == nil {
		return common.Errorf("spread", "unmarshal spread element", "", "decoder is nil")
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
			case "FlattenerPreference":
				var v FlattenerPreference
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				s.FlattenerPreference = &v
			case "Page":
				var v Page
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				s.Pages = append(s.Pages, v)
			case "TextFrame":
				var v TextFrame
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				s.TextFrames = append(s.TextFrames, v)
			case "Rectangle":
				var v Rectangle
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				s.Rectangles = append(s.Rectangles, v)
			case "Image":
				var v Image
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				s.Images = append(s.Images, v)
			case "Oval":
				var v Oval
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				s.Ovals = append(s.Ovals, v)
			case "Polygon":
				var v Polygon
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				s.Polygons = append(s.Polygons, v)
			case "GraphicLine":
				var v GraphicLine
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				s.GraphicLines = append(s.GraphicLines, v)
			case "Group":
				var v Group
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				s.Groups = append(s.Groups, v)
			case "Button":
				var v Button
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				s.Buttons = append(s.Buttons, v)
			default:
				raw, err := common.DecodeRaw(d, t)
				if err != nil {
					return err
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

// MarshalXML encodes the spread element, replaying the recorded child order so
// stacking order is preserved. Items appended after parsing are written last,
// which places them on top of the stack.
func (s Spread) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name = xml.Name{Local: s.elementName()}
	start.Attr = common.MarshalAttrs(s)
	if err := e.EncodeToken(start); err != nil {
		return err
	}

	kinds := []common.ChildKind{
		common.OptionalKind("FlattenerPreference", s.FlattenerPreference != nil, func(e *xml.Encoder) error {
			return e.EncodeElement(s.FlattenerPreference, xml.StartElement{Name: xml.Name{Local: "FlattenerPreference"}})
		}),
		{Name: "Page", Len: len(s.Pages), Encode: func(e *xml.Encoder, i int) error {
			return e.EncodeElement(s.Pages[i], xml.StartElement{Name: xml.Name{Local: "Page"}})
		}},
		{Name: "TextFrame", Len: len(s.TextFrames), Encode: func(e *xml.Encoder, i int) error {
			return e.EncodeElement(s.TextFrames[i], xml.StartElement{Name: xml.Name{Local: "TextFrame"}})
		}},
		{Name: "Rectangle", Len: len(s.Rectangles), Encode: func(e *xml.Encoder, i int) error {
			return e.EncodeElement(s.Rectangles[i], xml.StartElement{Name: xml.Name{Local: "Rectangle"}})
		}},
		{Name: "Image", Len: len(s.Images), Encode: func(e *xml.Encoder, i int) error {
			return e.EncodeElement(s.Images[i], xml.StartElement{Name: xml.Name{Local: "Image"}})
		}},
		{Name: "Oval", Len: len(s.Ovals), Encode: func(e *xml.Encoder, i int) error {
			return e.EncodeElement(s.Ovals[i], xml.StartElement{Name: xml.Name{Local: "Oval"}})
		}},
		{Name: "Polygon", Len: len(s.Polygons), Encode: func(e *xml.Encoder, i int) error {
			return e.EncodeElement(s.Polygons[i], xml.StartElement{Name: xml.Name{Local: "Polygon"}})
		}},
		{Name: "GraphicLine", Len: len(s.GraphicLines), Encode: func(e *xml.Encoder, i int) error {
			return e.EncodeElement(s.GraphicLines[i], xml.StartElement{Name: xml.Name{Local: "GraphicLine"}})
		}},
		{Name: "Group", Len: len(s.Groups), Encode: func(e *xml.Encoder, i int) error {
			return e.EncodeElement(s.Groups[i], xml.StartElement{Name: xml.Name{Local: "Group"}})
		}},
		{Name: "Button", Len: len(s.Buttons), Encode: func(e *xml.Encoder, i int) error {
			return e.EncodeElement(s.Buttons[i], xml.StartElement{Name: xml.Name{Local: "Button"}})
		}},
		{Name: kindOther, Len: len(s.OtherElements), Encode: func(e *xml.Encoder, i int) error { return e.Encode(s.OtherElements[i]) }},
	}
	if err := common.EncodeChildren(e, s.childOrder, kinds); err != nil {
		return err
	}
	return e.EncodeToken(start.End())
}
