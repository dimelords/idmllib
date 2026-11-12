package idml

import (
	"math"
	"path/filepath"
	"testing"

	"github.com/dimelords/idmllib/v3/pkg/spread"
)

// embeddedimage.idml was written by InDesign 2026: a JPEG (150 ppi) and a PNG (96 ppi) it made itself, each a
// coloured box of 200 × 120 pt, placed in a frame and their links embedded.
const embeddedImageFixture = "../../testdata/embeddedimage.idml"

func TestReadEmbeddedImages(t *testing.T) {
	pkg, err := Read(embeddedImageFixture)
	if err != nil {
		t.Fatal(err)
	}
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatal(err)
	}
	found := map[string]bool{}
	for _, sp := range spreads {
		for _, r := range sp.Rectangles {
			img := r.Image
			if img == nil {
				continue
			}
			data, ok, err := img.EmbeddedContents()
			if err != nil || !ok {
				t.Fatalf("%s: contents %v, %v", r.Self, ok, err)
			}
			info, err := spread.ImageInfo(data)
			if err != nil {
				t.Fatalf("%s: %v", r.Self, err)
			}
			if info.Format != img.ImageTypeName || img.Link == nil || img.Link.StoredState != "Embedded" || img.Link.LinkResourceFormat != info.Format {
				t.Errorf("%s: %+v, image %s, link %+v", r.Self, info, img.ImageTypeName, img.Link)
			}
			// InDesign's GraphicBounds are the size ImageInfo computes from the pixels and the resolution.
			l, tp, rt, bt, ok := img.GraphicBounds()
			w, h := info.Size()
			if !ok || l != 0 || tp != 0 || math.Abs(rt-w) > 0.001 || math.Abs(bt-h) > 0.001 {
				t.Errorf("%s: graphic bounds %v %v %v %v, ImageInfo %v × %v", r.Self, l, tp, rt, bt, w, h)
			}
			found[filepath.Ext(img.Link.LinkResourceURI)] = true
		}
	}
	if !found[".jpg"] || !found[".png"] {
		t.Errorf("found %v, want a .jpg and a .png", found)
	}
}
