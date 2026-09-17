package story

import (
	"encoding/xml"
	"strconv"
	"strings"

	"github.com/dimelords/idmllib/v2/pkg/common"
)

// Typed story content beyond plain paragraph and character ranges. All types
// preserve unknown attributes and children, and those with several child
// kinds keep document order (see docs/FIDELITY.md).

// XMLElement is a tagged region of a story, produced by InDesign's XML
// structure. It appears at three levels: inside <Story> wrapping whole
// paragraphs, inside <ParagraphStyleRange> wrapping character ranges, and
// inside <CharacterStyleRange> wrapping inline content.
type XMLElement struct {
	XMLName    xml.Name `xml:"XMLElement"`
	Self       string   `xml:"Self,attr,omitempty"`
	MarkupTag  string   `xml:"MarkupTag,attr,omitempty"`  // Reference to the XMLTag, e.g. "XMLTag/story"
	XMLContent string   `xml:"XMLContent,attr,omitempty"` // Story this element's content lives in

	XMLAttributes        []XMLAttribute         `xml:"XMLAttribute"`
	ParagraphStyleRanges []ParagraphStyleRange  `xml:"ParagraphStyleRange"`
	CharacterStyleRanges []CharacterStyleRange  `xml:"CharacterStyleRange"`
	XMLElements          []XMLElement           `xml:"XMLElement"`
	Contents             []Content              `xml:"Content"`
	Brs                  []Br                   `xml:"Br"`
	OtherElements        []common.RawXMLElement `xml:",any"`
	OtherAttrs           []xml.Attr             `xml:",any,attr"`

	childOrder common.ChildOrder
}

// XMLAttribute is an attribute on a tagged XMLElement.
type XMLAttribute struct {
	XMLName    xml.Name   `xml:"XMLAttribute"`
	Self       string     `xml:"Self,attr,omitempty"`
	Name       string     `xml:"Name,attr,omitempty"`
	Value      string     `xml:"Value,attr,omitempty"`
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// Change is a tracked change (Track Changes) wrapping inserted, deleted or
// moved text inside a character style range.
type Change struct {
	XMLName             xml.Name `xml:"Change"`
	ChangeType          string   `xml:"ChangeType,attr,omitempty"` // "InsertedText", "DeletedText", "MovedText"
	Date                string   `xml:"Date,attr,omitempty"`
	UserName            string   `xml:"UserName,attr,omitempty"`
	AppliedDocumentUser string   `xml:"AppliedDocumentUser,attr,omitempty"`

	Contents      []Content              `xml:"Content"`
	Brs           []Br                   `xml:"Br"`
	XMLElements   []XMLElement           `xml:"XMLElement"`
	OtherElements []common.RawXMLElement `xml:",any"`
	OtherAttrs    []xml.Attr             `xml:",any,attr"`

	childOrder common.ChildOrder
}

// Note is an editorial note anchored in the text. Its content is not part of
// the story's printed text.
type Note struct {
	XMLName             xml.Name `xml:"Note"`
	Collapsed           string   `xml:"Collapsed,attr,omitempty"`
	CreationDate        string   `xml:"CreationDate,attr,omitempty"`
	ModificationDate    string   `xml:"ModificationDate,attr,omitempty"`
	UserName            string   `xml:"UserName,attr,omitempty"`
	AppliedDocumentUser string   `xml:"AppliedDocumentUser,attr,omitempty"`

	ParagraphStyleRanges []ParagraphStyleRange  `xml:"ParagraphStyleRange"`
	OtherElements        []common.RawXMLElement `xml:",any"`
	OtherAttrs           []xml.Attr             `xml:",any,attr"`

	childOrder common.ChildOrder
}

// HyperlinkTextSource marks the text that is the source of a hyperlink
// defined in designmap.xml.
type HyperlinkTextSource struct {
	XMLName               xml.Name `xml:"HyperlinkTextSource"`
	Self                  string   `xml:"Self,attr,omitempty"`
	Name                  string   `xml:"Name,attr,omitempty"`
	Hidden                string   `xml:"Hidden,attr,omitempty"`
	AppliedCharacterStyle string   `xml:"AppliedCharacterStyle,attr,omitempty"`

	Contents      []Content              `xml:"Content"`
	Brs           []Br                   `xml:"Br"`
	XMLElements   []XMLElement           `xml:"XMLElement"`
	OtherElements []common.RawXMLElement `xml:",any"`
	OtherAttrs    []xml.Attr             `xml:",any,attr"`

	childOrder common.ChildOrder
}

// Footnote holds the paragraphs of a footnote anchored in the text.
type Footnote struct {
	XMLName              xml.Name               `xml:"Footnote"`
	ParagraphStyleRanges []ParagraphStyleRange  `xml:"ParagraphStyleRange"`
	OtherElements        []common.RawXMLElement `xml:",any"`
	OtherAttrs           []xml.Attr             `xml:",any,attr"`

	childOrder common.ChildOrder
}

// Table is a table anchored in the text. Rows and Columns describe the grid;
// Cells hold the content, named "column:row".
type Table struct {
	XMLName           xml.Name `xml:"Table"`
	Self              string   `xml:"Self,attr,omitempty"`
	HeaderRowCount    string   `xml:"HeaderRowCount,attr,omitempty"`
	FooterRowCount    string   `xml:"FooterRowCount,attr,omitempty"`
	BodyRowCount      string   `xml:"BodyRowCount,attr,omitempty"`
	ColumnCount       string   `xml:"ColumnCount,attr,omitempty"`
	AppliedTableStyle string   `xml:"AppliedTableStyle,attr,omitempty"`
	TableDirection    string   `xml:"TableDirection,attr,omitempty"`

	Properties    *common.Properties     `xml:"Properties,omitempty"`
	Rows          []Row                  `xml:"Row"`
	Columns       []Column               `xml:"Column"`
	Cells         []Cell                 `xml:"Cell"`
	OtherElements []common.RawXMLElement `xml:",any"`
	OtherAttrs    []xml.Attr             `xml:",any,attr"`

	childOrder common.ChildOrder
}

// Row describes one table row.
type Row struct {
	XMLName         xml.Name               `xml:"Row"`
	Self            string                 `xml:"Self,attr,omitempty"`
	Name            string                 `xml:"Name,attr,omitempty"` // row index as string
	SingleRowHeight string                 `xml:"SingleRowHeight,attr,omitempty"`
	Properties      *common.Properties     `xml:"Properties,omitempty"`
	OtherElements   []common.RawXMLElement `xml:",any"`
	OtherAttrs      []xml.Attr             `xml:",any,attr"`

	childOrder common.ChildOrder
}

// Column describes one table column.
type Column struct {
	XMLName           xml.Name               `xml:"Column"`
	Self              string                 `xml:"Self,attr,omitempty"`
	Name              string                 `xml:"Name,attr,omitempty"` // column index as string
	ColumnType        string                 `xml:"ColumnType,attr,omitempty"`
	SingleColumnWidth string                 `xml:"SingleColumnWidth,attr,omitempty"`
	Properties        *common.Properties     `xml:"Properties,omitempty"`
	OtherElements     []common.RawXMLElement `xml:",any"`
	OtherAttrs        []xml.Attr             `xml:",any,attr"`

	childOrder common.ChildOrder
}

// Cell is one table cell. Name is "column:row".
type Cell struct {
	XMLName          xml.Name `xml:"Cell"`
	Self             string   `xml:"Self,attr,omitempty"`
	Name             string   `xml:"Name,attr,omitempty"`
	RowSpan          string   `xml:"RowSpan,attr,omitempty"`
	ColumnSpan       string   `xml:"ColumnSpan,attr,omitempty"`
	CellType         string   `xml:"CellType,attr,omitempty"`
	AppliedCellStyle string   `xml:"AppliedCellStyle,attr,omitempty"`

	Properties           *common.Properties     `xml:"Properties,omitempty"`
	ParagraphStyleRanges []ParagraphStyleRange  `xml:"ParagraphStyleRange"`
	XMLElements          []XMLElement           `xml:"XMLElement"`
	OtherElements        []common.RawXMLElement `xml:",any"`
	OtherAttrs           []xml.Attr             `xml:",any,attr"`

	childOrder common.ChildOrder
}

// --- Ordered traversal -------------------------------------------------------

// paragraphsOf returns the paragraph ranges held directly in psrs and inside
// the given XMLElement wrappers, in document order.
func paragraphsOf(order common.ChildOrder, psrs []ParagraphStyleRange, xmls []XMLElement) []*ParagraphStyleRange {
	out := make([]*ParagraphStyleRange, 0, len(psrs))
	common.ForEachChild(order, []string{"ParagraphStyleRange", "XMLElement"}, []int{len(psrs), len(xmls)}, func(k, i int) {
		if k == 0 {
			out = append(out, &psrs[i])
		} else {
			out = append(out, xmls[i].Paragraphs()...)
		}
	})
	return out
}

// Paragraphs returns the story's paragraph ranges in document order,
// descending into XMLElement wrappers of tagged stories. Use it instead of
// ranging over ParagraphStyleRanges when the story may be tagged.
//
// The returned pointers refer into the underlying slices; they stay valid
// for in-place edits but not across appends to or removals from those slices.
func (s *Story) Paragraphs() []*ParagraphStyleRange {
	return paragraphsOf(s.childOrder, s.ParagraphStyleRanges, s.XMLElements)
}

// Paragraphs returns the paragraph ranges wrapped by this element and its
// nested elements, in document order.
func (x *XMLElement) Paragraphs() []*ParagraphStyleRange {
	return paragraphsOf(x.childOrder, x.ParagraphStyleRanges, x.XMLElements)
}

// Paragraphs returns the paragraph ranges of a footnote in document order.
func (f *Footnote) Paragraphs() []*ParagraphStyleRange {
	return paragraphsOf(f.childOrder, f.ParagraphStyleRanges, nil)
}

// Paragraphs returns the paragraph ranges of a note in document order.
func (n *Note) Paragraphs() []*ParagraphStyleRange {
	return paragraphsOf(n.childOrder, n.ParagraphStyleRanges, nil)
}

// Paragraphs returns the paragraph ranges of a cell in document order.
func (c *Cell) Paragraphs() []*ParagraphStyleRange {
	return paragraphsOf(c.childOrder, c.ParagraphStyleRanges, c.XMLElements)
}

// Ranges returns the paragraph's character style ranges in document order,
// descending into XMLElement wrappers.
func (p *ParagraphStyleRange) Ranges() []*CharacterStyleRange {
	out := make([]*CharacterStyleRange, 0, len(p.CharacterStyleRanges))
	common.ForEachChild(p.childOrder, []string{"CharacterStyleRange", "XMLElement"}, []int{len(p.CharacterStyleRanges), len(p.XMLElements)}, func(k, i int) {
		if k == 0 {
			out = append(out, &p.CharacterStyleRanges[i])
		} else {
			out = append(out, p.XMLElements[i].Ranges()...)
		}
	})
	return out
}

// Ranges returns the character style ranges wrapped by this element and its
// nested elements, in document order.
func (x *XMLElement) Ranges() []*CharacterStyleRange {
	out := make([]*CharacterStyleRange, 0, len(x.CharacterStyleRanges))
	common.ForEachChild(x.childOrder, []string{"CharacterStyleRange", "XMLElement"}, []int{len(x.CharacterStyleRanges), len(x.XMLElements)}, func(k, i int) {
		if k == 0 {
			out = append(out, &x.CharacterStyleRanges[i])
		} else {
			out = append(out, x.XMLElements[i].Ranges()...)
		}
	})
	return out
}

// Tables returns the tables anchored anywhere in the story's main text flow,
// in document order.
func (s *Story) Tables() []*Table {
	out := []*Table{}
	for _, psr := range s.Paragraphs() {
		for _, csr := range psr.Ranges() {
			for i := range csr.Children {
				if t := csr.Children[i].Table; t != nil {
					out = append(out, t)
				}
			}
		}
	}
	return out
}

// Cell returns the cell at the given column and row, or nil.
func (t *Table) Cell(column, row int) *Cell {
	name := strconv.Itoa(column) + ":" + strconv.Itoa(row)
	for i := range t.Cells {
		if t.Cells[i].Name == name {
			return &t.Cells[i]
		}
	}
	return nil
}

// --- Text extraction ---------------------------------------------------------

// inlineParts are the inline children of an element that can hold text
// directly: Content, Br, character ranges and nested XMLElements.
type inlineParts struct {
	order    common.ChildOrder
	contents []Content
	brs      []Br
	csrs     []CharacterStyleRange
	xmls     []XMLElement
}

// inlineText writes the text of inline children in document order.
func inlineText(b *strings.Builder, p inlineParts) {
	common.ForEachChild(p.order,
		[]string{"Content", "Br", "CharacterStyleRange", "XMLElement"},
		[]int{len(p.contents), len(p.brs), len(p.csrs), len(p.xmls)},
		func(k, i int) {
			switch k {
			case 0:
				b.WriteString(p.contents[i].Text)
			case 1:
				b.WriteString("\n")
			case 2:
				p.csrs[i].writeText(b)
			case 3:
				p.xmls[i].writeText(b)
			}
		})
}

func (x *XMLElement) writeText(b *strings.Builder) {
	// Paragraph-level wrappers hold paragraphs; inline wrappers hold content.
	for _, psr := range paragraphsOf(x.childOrder, x.ParagraphStyleRanges, nil) {
		psr.writeText(b)
	}
	inlineText(b, inlineParts{order: x.childOrder, contents: x.Contents, brs: x.Brs, csrs: x.CharacterStyleRanges, xmls: x.XMLElements})
}

func (p *ParagraphStyleRange) writeText(b *strings.Builder) {
	for _, csr := range p.Ranges() {
		csr.writeText(b)
	}
}

// writeText appends the printed text of the range: content, line breaks,
// hyperlink sources, tagged inline content and inserted or moved tracked
// changes. Deleted changes, notes, footnotes and tables are not part of the
// main text flow and are skipped.
func (c *CharacterStyleRange) writeText(b *strings.Builder) {
	for i := range c.Children {
		ch := &c.Children[i]
		switch {
		case ch.Content != nil:
			b.WriteString(ch.Content.Text)
		case ch.Br != nil:
			b.WriteString("\n")
		case ch.XMLElement != nil:
			ch.XMLElement.writeText(b)
		case ch.HyperlinkTextSource != nil:
			h := ch.HyperlinkTextSource
			inlineText(b, inlineParts{order: h.childOrder, contents: h.Contents, brs: h.Brs, xmls: h.XMLElements})
		case ch.Change != nil && ch.Change.ChangeType != "DeletedText":
			inlineText(b, inlineParts{order: ch.Change.childOrder, contents: ch.Change.Contents, brs: ch.Change.Brs, xmls: ch.Change.XMLElements})
		}
	}
}

// ExtractText returns the story's printed text in reading order. Tagged
// (XMLElement) content, hyperlink sources and inserted tracked changes are
// included; deleted changes, notes, footnotes and table content are not.
// Paragraph breaks (Br) become "\n".
func (s *Story) ExtractText() string {
	var b strings.Builder
	for _, psr := range s.Paragraphs() {
		psr.writeText(&b)
	}
	return b.String()
}

// ExtractText returns the printed text of a footnote.
func (f *Footnote) ExtractText() string {
	var b strings.Builder
	for _, psr := range f.Paragraphs() {
		psr.writeText(&b)
	}
	return b.String()
}

// ExtractText returns the text of a note.
func (n *Note) ExtractText() string {
	var b strings.Builder
	for _, psr := range n.Paragraphs() {
		psr.writeText(&b)
	}
	return b.String()
}

// ExtractText returns the text of a table cell.
func (c *Cell) ExtractText() string {
	var b strings.Builder
	for _, psr := range c.Paragraphs() {
		psr.writeText(&b)
	}
	return b.String()
}

// ExtractText returns the printed text of the story in this file; see
// Story.ExtractText.
func (f *File) ExtractText() string {
	return f.Story.ExtractText()
}
