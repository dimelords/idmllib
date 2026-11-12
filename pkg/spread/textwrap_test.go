package spread

import (
	"encoding/xml"
	"strings"
	"testing"
)

func TestNewBoundingBoxWrap(t *testing.T) {
	w := NewBoundingBoxWrap(5.3, 5.7, 0, 5.7)
	if w.TextWrapMode != WrapBoundingBox || w.TextWrapSide != "BothSides" || w.Inverse != "false" || w.ApplyToMasterPageOnly != "false" {
		t.Errorf("wrap %+v", w)
	}
	if top, left, bottom, right, ok := w.Offsets(); !ok || top != 5.3 || left != 5.7 || bottom != 0 || right != 5.7 {
		t.Errorf("offsets %v %v %v %v %v", top, left, bottom, right, ok)
	}
	w.SetOffsets(1, 2, 3, 4)
	n := 0
	for _, el := range w.Properties.OtherElements {
		if el.XMLName.Local == "TextWrapOffset" {
			n++
		}
	}
	if top, left, bottom, right, _ := w.Offsets(); n != 1 || top != 1 || left != 2 || bottom != 3 || right != 4 {
		t.Errorf("after SetOffsets: %d offsets, %v %v %v %v", n, top, left, bottom, right)
	}
	var none *TextWrapPreference
	if _, _, _, _, ok := none.Offsets(); ok {
		t.Error("a nil wrap has offsets")
	}

	// A text frame writes it as InDesign does: a TextWrapPreference child holding the offsets.
	tf := TextFrame{PageItemBase: PageItemBase{Self: "tf1"}, ParentStory: "s1", ContentType: "TextType", TextWrapPreference: NewBoundingBoxWrap(5.3, 5.7, 0, 5.7)}
	out, err := xml.Marshal(tf)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`<TextWrapPreference Inverse="false" ApplyToMasterPageOnly="false" TextWrapSide="BothSides" TextWrapMode="BoundingBoxTextWrap">`,
		`<TextWrapOffset Top="5.3" Left="5.7" Bottom="0" Right="5.7"`} {
		if !strings.Contains(string(out), want) {
			t.Errorf("no %s in %s", want, out)
		}
	}
	var back TextFrame
	if err := xml.Unmarshal(out, &back); err != nil {
		t.Fatal(err)
	}
	if back.TextWrapPreference == nil || back.TextWrapPreference.TextWrapMode != WrapBoundingBox || len(back.OtherElements) != 0 {
		t.Errorf("read back %+v, other elements %v", back.TextWrapPreference, back.OtherElements)
	}
}
