package story

import "encoding/xml"

// TextVariableInstance is a text variable placed in the text. InDesign shows the variable's current value there and
// keeps the value it last showed in ResultText. It lives in a CharacterStyleRange of its own, which InDesign marks
// with PageNumberType="TextVariable":
//
//	<CharacterStyleRange AppliedCharacterStyle="..." PageNumberType="TextVariable">
//	  <TextVariableInstance Self="uf9" Name="Publication" ResultText="..." AssociatedTextVariable="dTextVariablenPublication" />
//	</CharacterStyleRange>
type TextVariableInstance struct {
	XMLName xml.Name `xml:"TextVariableInstance"`
	Self    string   `xml:"Self,attr"`
	// Name is the variable's name.
	Name string `xml:"Name,attr,omitempty"`
	// ResultText is the text the instance showed when the document was written.
	ResultText string `xml:"ResultText,attr"`
	// AssociatedTextVariable is the variable's Self in designmap.xml ("dTextVariablenPublication").
	AssociatedTextVariable string `xml:"AssociatedTextVariable,attr,omitempty"`

	// OtherAttrs preserves attributes not modeled by a typed field, so
	// nothing is lost when the element is written back.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// NewTextVariableRange returns a character style range holding one instance of a text variable, as InDesign writes
// it. self is the instance's own Self; name and variableSelf are the variable's Name and Self, and resultText the
// text it shows now. An empty appliedCharacterStyle means no character style.
func NewTextVariableRange(appliedCharacterStyle, self, name, variableSelf, resultText string) CharacterStyleRange {
	if appliedCharacterStyle == "" {
		appliedCharacterStyle = "CharacterStyle/$ID/[No character style]"
	}
	return CharacterStyleRange{
		XMLName:               xml.Name{Local: "CharacterStyleRange"},
		AppliedCharacterStyle: appliedCharacterStyle,
		OtherAttrs:            []xml.Attr{{Name: xml.Name{Local: "PageNumberType"}, Value: "TextVariable"}},
		Children: []CharacterChild{{TextVariableInstance: &TextVariableInstance{
			Self: self, Name: name, ResultText: resultText, AssociatedTextVariable: variableSelf,
		}}},
	}
}

// TextVariableInstances returns the story's text variable instances in text order, in its paragraphs and the
// tagged content inside them, for reading and for updating their ResultText in place.
func (s *Story) TextVariableInstances() []*TextVariableInstance {
	var out []*TextVariableInstance
	for _, p := range s.Paragraphs() {
		for _, c := range p.Ranges() {
			for i := range c.Children {
				if v := c.Children[i].TextVariableInstance; v != nil {
					out = append(out, v)
				}
			}
		}
	}
	return out
}
