package story

import (
	"encoding/xml"

	"github.com/dimelords/idmllib/v2/pkg/common"
)

// kindOther is the ChildOrder kind used for children kept as RawXMLElement.
const kindOther = "other"

// UnmarshalXML decodes a <Story> element while recording the document order of
// its children, so that tagged content (XMLElement wrappers) and other unknown
// children keep their position relative to the paragraph ranges.
func (s *Story) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	if d == nil {
		return common.Errorf("story", "unmarshal story element", "", "decoder is nil")
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
			case "StoryPreference":
				var v StoryPreference
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				s.StoryPreference = &v
			case "InCopyExportOption":
				var v InCopyExportOption
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				s.InCopyExportOption = &v
			case "ParagraphStyleRange":
				var v ParagraphStyleRange
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				s.ParagraphStyleRanges = append(s.ParagraphStyleRanges, v)
			case "XMLElement":
				var v XMLElement
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				s.XMLElements = append(s.XMLElements, v)
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

// MarshalXML encodes the <Story> element, replaying the recorded child order.
// Children added after parsing are written after the recorded ones.
func (s Story) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name = xml.Name{Local: "Story"}
	start.Attr = common.MarshalAttrs(s)
	if err := e.EncodeToken(start); err != nil {
		return err
	}

	kinds := []common.ChildKind{
		common.OptionalKind("StoryPreference", s.StoryPreference != nil, func(e *xml.Encoder) error {
			return e.Encode(s.StoryPreference)
		}),
		common.OptionalKind("InCopyExportOption", s.InCopyExportOption != nil, func(e *xml.Encoder) error {
			return e.Encode(s.InCopyExportOption)
		}),
		{Name: "ParagraphStyleRange", Len: len(s.ParagraphStyleRanges), Encode: func(e *xml.Encoder, i int) error {
			return e.Encode(s.ParagraphStyleRanges[i])
		}},
		{Name: "XMLElement", Len: len(s.XMLElements), Encode: func(e *xml.Encoder, i int) error {
			return e.Encode(s.XMLElements[i])
		}},
		{Name: kindOther, Len: len(s.OtherElements), Encode: func(e *xml.Encoder, i int) error {
			return e.Encode(s.OtherElements[i])
		}},
	}
	if err := common.EncodeChildren(e, s.childOrder, kinds); err != nil {
		return err
	}
	return e.EncodeToken(start.End())
}

// UnmarshalXML decodes a <ParagraphStyleRange> while recording child order, so
// that a leading <Properties> block, XMLElement wrappers or other children keep
// their position relative to the character style ranges.
func (p *ParagraphStyleRange) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	if d == nil {
		return common.Errorf("story", "unmarshal paragraph style range", "", "decoder is nil")
	}
	p.XMLName = start.Name
	common.UnmarshalAttrs(p, start.Attr)
	p.childOrder = nil

	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "CharacterStyleRange":
				var v CharacterStyleRange
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				p.CharacterStyleRanges = append(p.CharacterStyleRanges, v)
			case "XMLElement":
				var v XMLElement
				if err := d.DecodeElement(&v, &t); err != nil {
					return err
				}
				p.XMLElements = append(p.XMLElements, v)
			default:
				raw, err := common.DecodeRaw(d, t)
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

// MarshalXML encodes the <ParagraphStyleRange>, replaying the recorded child order.
func (p ParagraphStyleRange) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name = xml.Name{Local: "ParagraphStyleRange"}
	start.Attr = common.MarshalAttrs(p)
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	kinds := []common.ChildKind{
		{Name: "CharacterStyleRange", Len: len(p.CharacterStyleRanges), Encode: func(e *xml.Encoder, i int) error {
			return e.Encode(p.CharacterStyleRanges[i])
		}},
		{Name: "XMLElement", Len: len(p.XMLElements), Encode: func(e *xml.Encoder, i int) error {
			return e.Encode(p.XMLElements[i])
		}},
		{Name: kindOther, Len: len(p.OtherElements), Encode: func(e *xml.Encoder, i int) error {
			return e.Encode(p.OtherElements[i])
		}},
	}
	if err := common.EncodeChildren(e, p.childOrder, kinds); err != nil {
		return err
	}
	return e.EncodeToken(start.End())
}
