package idml

import (
	"path/filepath"
	"testing"

	"github.com/dimelords/idmllib/v3/pkg/spread"
)

// textframewrap.idml was written by InDesign 2026: a text frame of placeholder text, and over it a text frame
// "Headline" wrapped by its bounding box, 5.3 pt on top, 5.7 at the sides and none at the bottom.
const textFrameWrapFixture = "../../testdata/textframewrap.idml"

func TestTextFrameWrap(t *testing.T) {
	pkg, err := Read(textFrameWrapFixture)
	if err != nil {
		t.Fatal(err)
	}
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatal(err)
	}
	var src string
	var head *spread.TextFrame
	modes := map[string]int{}
	for path, sp := range spreads {
		for i := range sp.TextFrames {
			tf := &sp.TextFrames[i]
			if tf.TextWrapPreference == nil {
				t.Fatalf("text frame %s: no typed TextWrapPreference", tf.Self)
			}
			for _, el := range tf.OtherElements {
				if el.XMLName.Local == "TextWrapPreference" {
					t.Errorf("text frame %s: TextWrapPreference also kept raw", tf.Self)
				}
			}
			modes[tf.TextWrapPreference.TextWrapMode]++
			if tf.TextWrapPreference.TextWrapMode == spread.WrapBoundingBox {
				head, src = tf, path
			}
		}
	}
	if modes[spread.WrapBoundingBox] != 1 || modes[spread.WrapNone] != 1 || head == nil {
		t.Fatalf("wrap modes %v", modes)
	}
	if top, left, bottom, right, ok := head.TextWrapPreference.Offsets(); !ok || top != 5.3 || left != 5.7 || bottom != 0 || right != 5.7 {
		t.Errorf("offsets %v %v %v %v %v", top, left, bottom, right, ok)
	}

	// A new wrap is written and read back.
	head.TextWrapPreference = spread.NewBoundingBoxWrap(10, 12, 4, 12)
	pkg.MarkModified(src)
	out := filepath.Join(t.TempDir(), "wrap.idml")
	if err := Write(pkg, out); err != nil {
		t.Fatal(err)
	}
	back, err := Read(out)
	if err != nil {
		t.Fatal(err)
	}
	sp, err := back.Spread(src)
	if err != nil {
		t.Fatal(err)
	}
	for _, tf := range sp.TextFrames {
		if tf.Self != head.Self {
			continue
		}
		if top, left, bottom, right, ok := tf.TextWrapPreference.Offsets(); !ok || top != 10 || left != 12 || bottom != 4 || right != 12 {
			t.Errorf("written offsets %v %v %v %v %v", top, left, bottom, right, ok)
		}
	}
}
