package story

import "encoding/xml"

// ContentSegment is one run of text or one processing instruction inside a
// <Content> element. InDesign encodes special characters as processing
// instructions, for example <?ACE 4?> for a footnote reference marker,
// <?ACE 18?> for an automatic page number and <?ACE 7?> for indent-to-here.
type ContentSegment struct {
	// Text is a run of literal text. Empty when this segment is an instruction.
	Text string
	// Instruction is set for a processing instruction segment.
	Instruction *xml.ProcInst
}

// SetText replaces the content with plain text, dropping any special
// character instructions.
func (c *Content) SetText(text string) {
	c.Text = text
	c.Segments = nil
}

// HasInstructions reports whether the content carries processing
// instructions (special characters) in addition to its text.
func (c *Content) HasInstructions() bool {
	for _, s := range c.Segments {
		if s.Instruction != nil {
			return true
		}
	}
	return false
}

// UnmarshalXML decodes <Content>, keeping text and processing instructions in
// order. Text receives the plain text; Segments is populated only when at
// least one instruction is present.
func (c *Content) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	c.XMLName = start.Name
	c.Text = ""
	c.Segments = nil
	var segs []ContentSegment
	hasPI := false
	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.CharData:
			s := string(t)
			c.Text += s
			segs = append(segs, ContentSegment{Text: s})
		case xml.ProcInst:
			hasPI = true
			inst := xml.ProcInst{Target: t.Target, Inst: append([]byte(nil), t.Inst...)}
			segs = append(segs, ContentSegment{Instruction: &inst})
		case xml.StartElement:
			// Content has no child elements; skip anything unexpected.
			if err := d.Skip(); err != nil {
				return err
			}
		case xml.EndElement:
			if hasPI {
				c.Segments = segs
			}
			return nil
		}
	}
}

// MarshalXML encodes <Content>. When Segments is set the segments are written
// in order, otherwise Text is written as character data.
func (c Content) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name = xml.Name{Local: "Content"}
	start.Attr = nil
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	if len(c.Segments) == 0 {
		if err := e.EncodeToken(xml.CharData(c.Text)); err != nil {
			return err
		}
	} else {
		for _, s := range c.Segments {
			if s.Instruction != nil {
				if err := e.EncodeToken(*s.Instruction); err != nil {
					return err
				}
				continue
			}
			if err := e.EncodeToken(xml.CharData(s.Text)); err != nil {
				return err
			}
		}
	}
	return e.EncodeToken(start.End())
}
