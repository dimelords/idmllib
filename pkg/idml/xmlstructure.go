package idml

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"strings"

	"github.com/dimelords/idmllib/v3/pkg/common"
)

// ResetXMLStructure empties the document's XML structure (XML/BackingStory.xml): every tagged element is removed
// and only the root element is kept, as in a new document. Tagged elements point at stories and page items by
// their Self, so after those are removed the structure would point at nothing. The tag definitions in
// XML/Tags.xml are kept. It returns how many elements were removed.
func (p *Package) ResetXMLStructure() (removed int, err error) {
	if !p.hasFile(PathBackingStory) {
		return 0, nil
	}
	data, err := p.FileData(PathBackingStory)
	if err != nil {
		return 0, err
	}
	bs, err := parseBackingStory(data)
	if err != nil {
		return 0, common.WrapErrorWithPath("idml", "reset xml structure", PathBackingStory, err)
	}
	if bs.elements <= 1 {
		return 0, nil // only the root: nothing to reset
	}

	var b bytes.Buffer
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + "\n")
	b.WriteString(`<idPkg:BackingStory xmlns:idPkg="` + idPkgNamespace + `" DOMVersion="` + escapeAttr(bs.domVersion) + `">` + "\n")
	b.WriteString("\t<XmlStory" + attrString(bs.storyAttrs) + ">\n")
	b.WriteString("\t\t" + `<ParagraphStyleRange AppliedParagraphStyle="ParagraphStyle/$ID/NormalParagraphStyle">` + "\n")
	b.WriteString("\t\t\t" + `<CharacterStyleRange AppliedCharacterStyle="CharacterStyle/$ID/[No character style]">` + "\n")
	b.WriteString("\t\t\t\t<XMLElement" + attrString(bs.rootAttrs) + " />\n")
	b.WriteString("\t\t\t</CharacterStyleRange>\n\t\t</ParagraphStyleRange>\n\t</XmlStory>\n</idPkg:BackingStory>\n")

	p.setFileData(PathBackingStory, b.Bytes())
	p.markDirty(PathBackingStory)
	return bs.elements - 1, nil
}

type backingStory struct {
	domVersion string
	storyAttrs []xml.Attr // the XmlStory's own attributes
	rootAttrs  []xml.Attr // the root XMLElement's own attributes (Self, MarkupTag)
	elements   int        // XMLElements in the structure, the root included
}

func parseBackingStory(data []byte) (*backingStory, error) {
	bs := &backingStory{domVersion: "20.4"}
	d := xml.NewDecoder(bytes.NewReader(data))
	for {
		tok, err := d.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		switch start.Name.Local {
		case "BackingStory":
			for _, a := range start.Attr {
				if a.Name.Local == "DOMVersion" {
					bs.domVersion = a.Value
				}
			}
		case "XmlStory":
			if bs.storyAttrs == nil {
				bs.storyAttrs = plainAttrs(start.Attr)
			}
		case "XMLElement":
			if bs.elements == 0 {
				bs.rootAttrs = plainAttrs(start.Attr)
			}
			bs.elements++
		}
	}
	if bs.storyAttrs == nil {
		return nil, errors.New("no XmlStory")
	}
	if bs.rootAttrs == nil {
		return nil, errors.New("no root XMLElement")
	}
	return bs, nil
}

// plainAttrs copies attributes without namespaces, as IDML element attributes are written.
func plainAttrs(attrs []xml.Attr) []xml.Attr {
	out := make([]xml.Attr, 0, len(attrs))
	for _, a := range attrs {
		if a.Name.Space != "" {
			continue
		}
		out = append(out, xml.Attr{Name: xml.Name{Local: a.Name.Local}, Value: a.Value})
	}
	return out
}

func attrString(attrs []xml.Attr) string {
	var b strings.Builder
	for _, a := range attrs {
		b.WriteString(" " + a.Name.Local + `="` + escapeAttr(a.Value) + `"`)
	}
	return b.String()
}

func escapeAttr(s string) string {
	var b bytes.Buffer
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}
