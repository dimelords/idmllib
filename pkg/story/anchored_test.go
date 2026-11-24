package story

import (
	"strings"
	"testing"
)

// anchoredXML holds every kind of page item anchored in text, at every level a story tags: in a story-level tag's
// paragraph, a paragraph-level tag's character range, character ranges, and an inline tag around a text frame.
const anchoredXML = `<?xml version="1.0" encoding="UTF-8"?>
<idPkg:Story xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging" DOMVersion="21.0">
<Story Self="us">
<XMLElement Self="x0" MarkupTag="XMLTag/Story" XMLContent="us">
<ParagraphStyleRange AppliedParagraphStyle="ParagraphStyle/$ID/NormalParagraphStyle">
<CharacterStyleRange AppliedCharacterStyle="CharacterStyle/$ID/[No character style]">
<Content>a </Content>
<TextFrame Self="tf1" ParentStory="us1"><AnchoredObjectSetting AnchoredPosition="InlinePosition" /></TextFrame>
<Content> b </Content>
<XMLElement Self="x1" MarkupTag="XMLTag/Item" XMLContent="us2" NoTextMarker="true"><TextFrame Self="tf2" ParentStory="us2" /><XMLAttribute Self="x1a" Name="id" Value="1" /></XMLElement>
<Content> c </Content>
<Group Self="g1"><Polygon Self="p1" /><Polygon Self="p2" /></Group>
<Polygon Self="p3" />
<Rectangle Self="r1" />
</CharacterStyleRange>
</ParagraphStyleRange>
</XMLElement>
<ParagraphStyleRange AppliedParagraphStyle="ParagraphStyle/$ID/NormalParagraphStyle">
<XMLElement Self="x2" MarkupTag="XMLTag/Item" XMLContent="us">
<CharacterStyleRange AppliedCharacterStyle="CharacterStyle/$ID/[No character style]">
<Oval Self="o1"><AnchoredObjectSetting AnchoredPosition="Anchored" /></Oval>
<Content>d</Content>
</CharacterStyleRange>
</XMLElement>
<CharacterStyleRange AppliedCharacterStyle="CharacterStyle/$ID/[No character style]">
<GraphicLine Self="l1" />
<Content>e</Content>
<XMLElement Self="x3" MarkupTag="XMLTag/Item" NoTextMarker="true"><Rectangle Self="r2" /><Group Self="g2" /><Polygon Self="p4" /><Oval Self="o2" /><GraphicLine Self="l2" /></XMLElement>
</CharacterStyleRange>
</ParagraphStyleRange>
</Story>
</idPkg:Story>`

func anchoredSelfs(st *Story) []string {
	var out []string
	for _, a := range st.Anchored() {
		switch {
		case a.TextFrame != nil:
			out = append(out, "TextFrame "+a.TextFrame.Self)
		case a.Rectangle != nil:
			out = append(out, "Rectangle "+a.Rectangle.Self)
		case a.Group != nil:
			out = append(out, "Group "+a.Group.Self)
		case a.Polygon != nil:
			out = append(out, "Polygon "+a.Polygon.Self)
		case a.Oval != nil:
			out = append(out, "Oval "+a.Oval.Self)
		case a.GraphicLine != nil:
			out = append(out, "GraphicLine "+a.GraphicLine.Self)
		}
	}
	return out
}

const wantAnchored = "TextFrame tf1, TextFrame tf2, Group g1, Polygon p3, Rectangle r1, Oval o1, GraphicLine l1, " +
	"Rectangle r2, Group g2, Polygon p4, Oval o2, GraphicLine l2"

func TestAnchoredItems(t *testing.T) {
	f, err := ParseStory([]byte(anchoredXML))
	if err != nil {
		t.Fatal(err)
	}
	st := &f.Story
	if got := strings.Join(anchoredSelfs(st), ", "); got != wantAnchored {
		t.Fatalf("anchored %s, want %s", got, wantAnchored)
	}
	items := st.Anchored()
	if g := items[2].Group; len(g.Polygons) != 2 {
		t.Errorf("group holds %d polygons", len(g.Polygons))
	}
	if len(items[0].TextFrame.OtherElements) != 1 || items[0].TextFrame.OtherElements[0].XMLName.Local != "AnchoredObjectSetting" {
		t.Errorf("the text frame's anchoring isn't kept: %+v", items[0].TextFrame.OtherElements)
	}
	// The inline tag holds its text frame typed, its attribute kept.
	x := st.XMLElements[0].ParagraphStyleRanges[0].CharacterStyleRanges[0].Children[3].XMLElement
	if x == nil || len(x.TextFrames) != 1 || len(x.XMLAttributes) != 1 || len(x.OtherElements) != 0 {
		t.Fatalf("inline tag %+v", x)
	}
	// Items are pointers into the story: renaming through them renames the story's.
	items[1].TextFrame.Self = "tf2b"
	if x.TextFrames[0].Self != "tf2b" {
		t.Error("Anchored doesn't point into the story")
	}
	items[1].TextFrame.Self = "tf2"

	// Written and read again, every item is where it was, among the text around it.
	out, err := MarshalStory(f)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{
		`<Content>a </Content><TextFrame Self="tf1"`, `</TextFrame><Content> b </Content><XMLElement Self="x1"`,
		`<TextFrame Self="tf2" ParentStory="us2"></TextFrame><XMLAttribute Self="x1a"`, `<Content> c </Content><Group Self="g1">`,
		`</Group><Polygon Self="p3"></Polygon><Rectangle Self="r1"`, `<Oval Self="o1">`, `</Oval><Content>d</Content>`,
		`<GraphicLine Self="l1"></GraphicLine><Content>e</Content>`,
		`<XMLElement Self="x3" MarkupTag="XMLTag/Item" NoTextMarker="true"><Rectangle Self="r2"`,
	} {
		if !strings.Contains(compact(string(out)), s) {
			t.Errorf("written story lacks %s", s)
		}
	}
	again, err := ParseStory(out)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(anchoredSelfs(&again.Story), ", "); got != wantAnchored {
		t.Errorf("anchored after writing %s", got)
	}
}

// compact drops the whitespace between elements, so the test can look for siblings.
func compact(s string) string {
	var b strings.Builder
	for _, line := range strings.Split(s, "\n") {
		b.WriteString(strings.TrimSpace(line))
	}
	return strings.ReplaceAll(b.String(), "> <", "><")
}
