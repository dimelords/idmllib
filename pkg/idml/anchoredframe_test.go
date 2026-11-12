package idml

import (
	"bytes"
	"encoding/xml"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dimelords/idmllib/v3/pkg/spread"
	"github.com/dimelords/idmllib/v3/pkg/story"
)

// anchoredimage.idml was written by InDesign 2026: a text frame whose second and third paragraphs each hold a 32 pt
// frame anchored in the text with a PNG (144 ppi) placed in it and its link embedded; the first frame is custom
// anchored (its bottom right corner to the text), the second inline.
const anchoredImageFixture = "../../testdata/anchoredimage.idml"

// anchoredStory returns the fixture's story with the anchored frames.
func anchoredStory(t *testing.T, pkg *Package) (string, *story.Story) {
	t.Helper()
	stories, err := pkg.Stories()
	if err != nil {
		t.Fatal(err)
	}
	for path, st := range stories {
		if len(st.Rectangles()) > 0 {
			return path, st
		}
	}
	t.Fatal("no story with anchored frames")
	return "", nil
}

// anchoring is a frame's AnchoredObjectSetting AnchoredPosition, kept with its other children.
func anchoring(r *spread.Rectangle) string {
	for _, e := range r.OtherElements {
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

func TestReadAnchoredFrames(t *testing.T) {
	pkg, err := Read(anchoredImageFixture)
	if err != nil {
		t.Fatal(err)
	}
	_, st := anchoredStory(t, pkg)
	frames := st.Rectangles()
	if len(frames) != 2 {
		t.Fatalf("%d anchored frames, want 2", len(frames))
	}
	for i, want := range []string{"Anchored", "InlinePosition"} {
		r := frames[i]
		if got := anchoring(r); got != want {
			t.Errorf("frame %d: anchored position %q, want %q", i, got, want)
		}
		if r.Image == nil {
			t.Fatalf("frame %d: no image", i)
		}
		data, ok, err := r.Image.EmbeddedContents()
		if err != nil || !ok {
			t.Fatalf("frame %d: contents %v, %v", i, ok, err)
		}
		if info, err := spread.ImageInfo(data); err != nil || info.Format != spread.FormatPNG {
			t.Errorf("frame %d: %+v, %v", i, info, err)
		}
	}
}

// A frame built with NewEmbeddedImageRectangle goes into a story's text, and is there with its image when the
// document is written and read again.
func TestWriteAnchoredFrame(t *testing.T) {
	pkg, err := Read(anchoredImageFixture)
	if err != nil {
		t.Fatal(err)
	}
	path, st := anchoredStory(t, pkg)
	data, _, err := st.Rectangles()[0].Image.EmbeddedContents()
	if err != nil {
		t.Fatal(err)
	}
	info, err := spread.ImageInfo(data)
	if err != nil {
		t.Fatal(err)
	}
	w, h := info.Size()
	r, err := spread.NewEmbeddedImageRectangle("uanc", "uanci", "uancl", 0, 0, 32, 32, 32/w, 32/h, 0, 0,
		spread.EmbeddedImage{Data: data, Name: "byline.png", Format: info.Format, Width: w, Height: h, PPIX: info.PPIX, PPIY: info.PPIY})
	if err != nil {
		t.Fatal(err)
	}
	st.ParagraphStyleRanges = append(st.ParagraphStyleRanges, story.ParagraphStyleRange{
		XMLName:               xml.Name{Local: "ParagraphStyleRange"},
		AppliedParagraphStyle: "ParagraphStyle/$ID/NormalParagraphStyle",
		CharacterStyleRanges: []story.CharacterStyleRange{{
			XMLName:               xml.Name{Local: "CharacterStyleRange"},
			AppliedCharacterStyle: "CharacterStyle/$ID/[No character style]",
			Children:              []story.CharacterChild{{Rectangle: r}},
		}},
	})
	pkg.MarkModified(path)
	out := filepath.Join(t.TempDir(), "out.idml")
	if err := Write(pkg, out); err != nil {
		t.Fatal(err)
	}
	again, err := Read(out)
	if err != nil {
		t.Fatal(err)
	}
	_, st2 := anchoredStory(t, again)
	frames := st2.Rectangles()
	if len(frames) != 3 || frames[2].Self != "uanc" || frames[2].Image == nil {
		t.Fatalf("frames after writing: %d", len(frames))
	}
	got, ok, err := frames[2].Image.EmbeddedContents()
	if err != nil || !ok || !bytes.Equal(got, data) {
		t.Errorf("the new frame's image: %d bytes, %v, %v", len(got), ok, err)
	}
	// The frame is written inside the character range, where InDesign expects an anchored object.
	raw, err := again.FileData(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `<Rectangle Self="uanc"`) {
		t.Error("the anchored frame isn't in the story's XML")
	}
}
