package idml

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/dimelords/idmllib/v3/pkg/common"
	"github.com/dimelords/idmllib/v3/pkg/spread"
	"github.com/dimelords/idmllib/v3/pkg/story"
)

// anchoreditems.idml was written by InDesign 2026: a text frame whose story, tagged as a whole, holds in order an
// inline text frame with a story of its own, an inline text frame tagged with an XML element of its own, an inline
// group of two triangles, and a custom anchored oval and line.
const anchoredItemsFixture = "../../testdata/anchoreditems.idml"

// itemsStory returns the fixture's story with the anchored items.
func itemsStory(t *testing.T, pkg *Package) (string, *story.Story) {
	t.Helper()
	stories, err := pkg.Stories()
	if err != nil {
		t.Fatal(err)
	}
	for path, st := range stories {
		if len(st.Anchored()) > 0 {
			return path, st
		}
	}
	t.Fatal("no story with anchored items")
	return "", nil
}

// kindOf names the item an Anchored holds, with its Self.
func kindOf(a story.Anchored) (string, string) {
	switch {
	case a.TextFrame != nil:
		return "TextFrame", a.TextFrame.Self
	case a.Rectangle != nil:
		return "Rectangle", a.Rectangle.Self
	case a.Group != nil:
		return "Group", a.Group.Self
	case a.Polygon != nil:
		return "Polygon", a.Polygon.Self
	case a.Oval != nil:
		return "Oval", a.Oval.Self
	case a.GraphicLine != nil:
		return "GraphicLine", a.GraphicLine.Self
	}
	return "", ""
}

// anchoredPosition is an item's AnchoredObjectSetting AnchoredPosition, kept with its other children.
func anchoredPosition(els []common.RawXMLElement) string {
	for _, e := range els {
		if e.XMLName.Local != "AnchoredObjectSetting" {
			continue
		}
		for _, a := range e.Attrs {
			if a.Name.Local == "AnchoredPosition" {
				return a.Value
			}
		}
	}
	return ""
}

func TestReadAnchoredItems(t *testing.T) {
	pkg, err := Read(anchoredItemsFixture)
	if err != nil {
		t.Fatal(err)
	}
	_, st := itemsStory(t, pkg)
	items := st.Anchored()
	var kinds []string
	for _, a := range items {
		k, _ := kindOf(a)
		kinds = append(kinds, k)
	}
	if strings.Join(kinds, " ") != "TextFrame TextFrame Group Oval GraphicLine" {
		t.Fatalf("anchored items %v", kinds)
	}
	if g := items[2].Group; len(g.Polygons) != 2 {
		t.Errorf("the group holds %d polygons, want 2", len(g.Polygons))
	}
	positions := []string{
		anchoredPosition(items[0].TextFrame.OtherElements), anchoredPosition(items[1].TextFrame.OtherElements),
		anchoredPosition(items[2].Group.OtherElements), anchoredPosition(items[3].Oval.OtherElements),
		anchoredPosition(items[4].GraphicLine.OtherElements),
	}
	// InDesign writes no AnchoredObjectSetting on the inline group itself (its polygons have one each).
	if strings.Join(positions, ",") != "InlinePosition,InlinePosition,,Anchored,Anchored" {
		t.Errorf("anchored positions %v", positions)
	}
	// Each anchored text frame shows a story of its own in the package.
	stories, err := pkg.Stories()
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range items[:2] {
		if _, ok := stories["Stories/Story_"+a.TextFrame.ParentStory+".xml"]; !ok {
			t.Errorf("frame %s: no story %s", a.TextFrame.Self, a.TextFrame.ParentStory)
		}
	}
	// The tagged frame is held by its XML element, typed, and nothing is left raw.
	tagged := false
	var walk func(x *story.XMLElement)
	walk = func(x *story.XMLElement) {
		if slices.ContainsFunc(x.TextFrames, func(f spread.TextFrame) bool { return f.Self == items[1].TextFrame.Self }) {
			tagged = true
		}
		for _, o := range x.OtherElements {
			t.Errorf("XML element %s keeps a raw %s", x.Self, o.XMLName.Local)
		}
		for i := range x.XMLElements {
			walk(&x.XMLElements[i])
		}
		var ranges []*story.CharacterStyleRange
		for _, p := range x.Paragraphs() {
			ranges = append(ranges, p.Ranges()...)
		}
		for _, c := range append(ranges, x.Ranges()...) {
			for _, ch := range c.Children {
				if ch.XMLElement != nil {
					walk(ch.XMLElement)
				}
			}
		}
	}
	for i := range st.XMLElements {
		walk(&st.XMLElements[i])
	}
	if !tagged {
		t.Error("the tagged frame isn't in an XML element's TextFrames")
	}
	for _, p := range st.Paragraphs() {
		for _, c := range p.Ranges() {
			for _, ch := range c.Children {
				if ch.Other != nil {
					t.Errorf("a character range keeps a raw %s", ch.Other.XMLName.Local)
				}
			}
		}
	}
}

// A story copied with its anchored items under Selfs of their own, and the anchored frames' stories copied too, is
// there typed after writing and reading again, with nothing pointing at the originals.
func TestWriteAnchoredItems(t *testing.T) {
	pkg, err := Read(anchoredItemsFixture)
	if err != nil {
		t.Fatal(err)
	}
	path, st := itemsStory(t, pkg)
	stories, err := pkg.Stories()
	if err != nil {
		t.Fatal(err)
	}
	var old []string
	n := 0
	next := func() string { n++; return "ucopy" + strings.Repeat("x", n) }
	for _, a := range st.Anchored() {
		_, self := kindOf(a)
		old = append(old, self)
		switch {
		case a.TextFrame != nil:
			a.TextFrame.Self = next()
			src := stories["Stories/Story_"+a.TextFrame.ParentStory+".xml"]
			cp := *src
			cp.Self = next()
			if err := pkg.AddStory("Stories/Story_"+cp.Self+".xml", &cp, ValidationOptions{}); err != nil {
				t.Fatal(err)
			}
			a.TextFrame.ParentStory = cp.Self
		case a.Group != nil:
			a.Group.Self = next()
			for i := range a.Group.Polygons {
				old = append(old, a.Group.Polygons[i].Self)
				a.Group.Polygons[i].Self = next()
			}
		case a.Oval != nil:
			a.Oval.Self = next()
		case a.GraphicLine != nil:
			a.GraphicLine.Self = next()
		}
	}
	pkg.MarkModified(path)
	out := filepath.Join(t.TempDir(), "out.idml")
	if err := Write(pkg, out); err != nil {
		t.Fatal(err)
	}
	again, err := Read(out)
	if err != nil {
		t.Fatal(err)
	}
	_, st2 := itemsStory(t, again)
	items := st2.Anchored()
	if len(items) != 5 {
		t.Fatalf("%d anchored items after writing, want 5", len(items))
	}
	stories2, err := again.Stories()
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range items {
		k, self := kindOf(a)
		if !strings.HasPrefix(self, "ucopy") {
			t.Errorf("%s %s kept its Self", k, self)
		}
		if a.TextFrame != nil {
			if _, ok := stories2["Stories/Story_"+a.TextFrame.ParentStory+".xml"]; !ok || !strings.HasPrefix(a.TextFrame.ParentStory, "ucopy") {
				t.Errorf("frame %s shows story %s", self, a.TextFrame.ParentStory)
			}
		}
	}
	raw, err := again.FileData(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, self := range old {
		if strings.Contains(string(raw), `Self="`+self+`"`) {
			t.Errorf("the story still holds %s", self)
		}
	}
}
