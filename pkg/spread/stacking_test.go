package spread

import (
	"regexp"
	"slices"
	"strings"
	"testing"
)

const stackingXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<idPkg:Spread xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging" DOMVersion="20.4">
	<Spread Self="ud3" PageCount="1">
		<FlattenerPreference LineArtAndTextResolution="300"/>
		<Page Self="ud8" Name="1"/>
		<Rectangle Self="box" ItemTransform="1 0 0 1 0 0"/>
		<TextFrame Self="date" ParentStory="st1" ItemTransform="1 0 0 1 0 0"/>
		<EPS Self="raw"/>
		<GraphicLine Self="rule" ItemTransform="1 0 0 1 0 0"/>
		<TextFrame Self="title" ParentStory="st2" ItemTransform="1 0 0 1 0 0"/>
	</Spread>
</idPkg:Spread>`

// written returns the Selfs of a marshaled spread's children in the order they are written.
func written(t *testing.T, f *File) []string {
	t.Helper()
	data, err := MarshalSpread(f)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, m := range regexp.MustCompile(`<(\w+) Self="([^"]+)"`).FindAllStringSubmatch(string(data), -1) {
		if m[1] != "Spread" {
			out = append(out, m[2])
		}
	}
	return out
}

func TestStackingOrderOfAParsedSpread(t *testing.T) {
	f, err := ParseSpread([]byte(stackingXML))
	if err != nil {
		t.Fatal(err)
	}
	// Pages and raw elements are not typed page items; the typed ones are listed bottom to top.
	if got, want := f.Spread.StackingOrder(), []string{"box", "date", "rule", "title"}; !slices.Equal(got, want) {
		t.Errorf("StackingOrder() = %v, want %v", got, want)
	}

	// Restacking the same order changes nothing.
	if err := f.Spread.SetStackingOrder(f.Spread.StackingOrder()); err != nil {
		t.Fatal(err)
	}
	if got, want := written(t, f), []string{"ud8", "box", "date", "raw", "rule", "title"}; !slices.Equal(got, want) {
		t.Errorf("written %v, want %v", got, want)
	}

	// Moving the box to the top: the typed items fill their slots in the new order, while the page and the raw
	// element keep their places.
	if err := f.Spread.SetStackingOrder([]string{"date", "rule", "title", "box"}); err != nil {
		t.Fatal(err)
	}
	if got, want := written(t, f), []string{"ud8", "date", "rule", "raw", "title", "box"}; !slices.Equal(got, want) {
		t.Errorf("written %v, want %v", got, want)
	}
	if f.Spread.TextFrames[0].Self != "date" || f.Spread.Rectangles[0].Self != "box" {
		t.Errorf("the typed slices lost items: %v %v", f.Spread.TextFrames, f.Spread.Rectangles)
	}

	// The order survives a roundtrip.
	data, err := MarshalSpread(f)
	if err != nil {
		t.Fatal(err)
	}
	g, err := ParseSpread(data)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := g.Spread.StackingOrder(), []string{"date", "rule", "title", "box"}; !slices.Equal(got, want) {
		t.Errorf("after a roundtrip StackingOrder() = %v, want %v", got, want)
	}
}

// TestStackingOrderOfAppendedItems is the case the method is for: items appended through the typed slices are
// written grouped by kind, so a text frame appended after a rectangle ends up below it until it is restacked.
func TestStackingOrderOfAppendedItems(t *testing.T) {
	var sp Spread
	sp.Pages = append(sp.Pages, Page{Self: "p1"})
	sp.Rectangles = append(sp.Rectangles, Rectangle{PageItemBase: PageItemBase{Self: "box"}})
	sp.TextFrames = append(sp.TextFrames, TextFrame{PageItemBase: PageItemBase{Self: "date"}})
	if got, want := sp.StackingOrder(), []string{"date", "box"}; !slices.Equal(got, want) {
		t.Errorf("appended StackingOrder() = %v, want %v (grouped by kind, text frames first)", got, want)
	}
	if err := sp.SetStackingOrder([]string{"box", "date"}); err != nil {
		t.Fatal(err)
	}
	if got, want := written(t, &File{Spread: sp}), []string{"p1", "box", "date"}; !slices.Equal(got, want) {
		t.Errorf("written %v, want %v", got, want)
	}

	// Items appended after a restack go on top, as with a parsed spread.
	sp.GraphicLines = append(sp.GraphicLines, GraphicLine{PageItemBase: PageItemBase{Self: "rule"}})
	if got, want := sp.StackingOrder(), []string{"box", "date", "rule"}; !slices.Equal(got, want) {
		t.Errorf("after appending StackingOrder() = %v, want %v", got, want)
	}
}

func TestSetStackingOrderNeedsEveryItemOnce(t *testing.T) {
	f, err := ParseSpread([]byte(stackingXML))
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		order []string
		want  string
	}{
		{[]string{"box", "date", "rule"}, "title"},
		{[]string{"box", "date", "rule", "title", "nope"}, "nope"},
		{[]string{"box", "date", "rule", "rule"}, "rule"},
		{[]string{"box", "date", "rule", "title", "ud8"}, "ud8"},
	} {
		err := f.Spread.SetStackingOrder(c.order)
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("SetStackingOrder(%v) = %v, want an error naming %s", c.order, err, c.want)
		}
	}
	if got, want := f.Spread.StackingOrder(), []string{"box", "date", "rule", "title"}; !slices.Equal(got, want) {
		t.Errorf("a failed SetStackingOrder changed the order: %v", got)
	}
}
