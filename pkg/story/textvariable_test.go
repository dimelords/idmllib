package story

import (
	"strings"
	"testing"
)

// The range InDesign writes for a text variable instance roundtrips, and its result text is part of the story's
// printed text.
func TestTextVariableRange(t *testing.T) {
	const x = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<idPkg:Story xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging" DOMVersion="21.0">
	<Story Self="ue3">
		<ParagraphStyleRange AppliedParagraphStyle="ParagraphStyle/$ID/NormalParagraphStyle">
			<CharacterStyleRange AppliedCharacterStyle="CharacterStyle/$ID/[No character style]">
				<Content>Head: </Content>
			</CharacterStyleRange>
			<CharacterStyleRange AppliedCharacterStyle="CharacterStyle/$ID/[No character style]" PageNumberType="TextVariable">
				<TextVariableInstance Self="uf9" Name="Publication" ResultText="The Daily Example" AssociatedTextVariable="dTextVariablenPublication" />
			</CharacterStyleRange>
		</ParagraphStyleRange>
	</Story>
</idPkg:Story>`
	f, err := ParseStory([]byte(x))
	if err != nil {
		t.Fatal(err)
	}
	insts := f.Story.TextVariableInstances()
	if len(insts) != 1 || insts[0].Self != "uf9" || insts[0].Name != "Publication" || insts[0].AssociatedTextVariable != "dTextVariablenPublication" {
		t.Fatalf("instances %+v", insts)
	}
	if got := f.Story.ExtractText(); got != "Head: The Daily Example" {
		t.Errorf("ExtractText() = %q", got)
	}
	insts[0].ResultText = "The Evening Example"
	data, err := MarshalStory(f)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `<TextVariableInstance Self="uf9" Name="Publication" ResultText="The Evening Example" AssociatedTextVariable="dTextVariablenPublication">`) &&
		!strings.Contains(string(data), `<TextVariableInstance Self="uf9" Name="Publication" ResultText="The Evening Example" AssociatedTextVariable="dTextVariablenPublication"></TextVariableInstance>`) {
		t.Errorf("written as\n%s", data)
	}

	r := NewTextVariableRange("", "u1", "Weekday", "dTextVariablenWeekday", "Tuesday")
	if r.AppliedCharacterStyle != "CharacterStyle/$ID/[No character style]" || len(r.OtherAttrs) != 1 || r.OtherAttrs[0].Value != "TextVariable" {
		t.Errorf("range %+v", r)
	}
}
