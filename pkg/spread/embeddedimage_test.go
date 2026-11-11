package spread

import (
	"bytes"
	"encoding/xml"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"math"
	"strings"
	"testing"

	"github.com/dimelords/idmllib/v3/pkg/common"
)

func testJPEG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for i := range img.Pix {
		img.Pix[i] = 0x80
	}
	var b bytes.Buffer
	if err := jpeg.Encode(&b, img, nil); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func testPNG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	img.Set(0, 0, color.RGBA{R: 0xff, A: 0xff})
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

// withJFIFDensity puts a JFIF segment (APP0) with a resolution after a JPEG's start of image, where InDesign and
// most encoders write it (Go's writes none): unit 1 is dots per inch, 2 dots per centimetre, 0 an aspect ratio.
func withJFIFDensity(t *testing.T, data []byte, unit byte, x, y uint16) []byte {
	t.Helper()
	if !bytes.HasPrefix(data, []byte{0xFF, 0xD8}) {
		t.Fatal("not a JPEG")
	}
	app0 := []byte{0xFF, 0xE0, 0, 16, 'J', 'F', 'I', 'F', 0, 1, 1, unit, byte(x >> 8), byte(x), byte(y >> 8), byte(y), 0, 0}
	return append(append([]byte{0xFF, 0xD8}, app0...), data[2:]...)
}

// withPHYs inserts a pHYs chunk (pixels per metre) after a PNG's IHDR. Its CRC isn't checked by the reader.
func withPHYs(data []byte, ppm uint32) []byte {
	ihdrEnd := 8 + 4 + 4 + 13 + 4
	chunk := []byte{0, 0, 0, 9, 'p', 'H', 'Y', 's',
		byte(ppm >> 24), byte(ppm >> 16), byte(ppm >> 8), byte(ppm), byte(ppm >> 24), byte(ppm >> 16), byte(ppm >> 8), byte(ppm), 1,
		0, 0, 0, 0}
	return append(append(bytes.Clone(data[:ihdrEnd]), chunk...), data[ihdrEnd:]...)
}

func TestImageInfo(t *testing.T) {
	near := func(a, b float64) bool { return math.Abs(a-b) < 1e-9 }
	jpg := testJPEG(t, 300, 200)
	for _, c := range []struct {
		name   string
		data   []byte
		format string
		ppi    float64
	}{
		{"jpeg without a resolution", jpg, FormatJPEG, 72},
		{"jpeg at 300 dpi", withJFIFDensity(t, jpg, 1, 300, 300), FormatJPEG, 300},
		{"jpeg at 118 dots per cm", withJFIFDensity(t, jpg, 2, 118, 118), FormatJPEG, 118 * 2.54},
		{"jpeg with an aspect ratio only", withJFIFDensity(t, jpg, 0, 1, 1), FormatJPEG, 72},
		{"png without pHYs", testPNG(t, 300, 200), FormatPNG, 72},
		{"png at 3779 pixels per metre", withPHYs(testPNG(t, 300, 200), 3779), FormatPNG, 3779 * 0.0254},
	} {
		info, err := ImageInfo(c.data)
		if err != nil {
			t.Errorf("%s: %v", c.name, err)
			continue
		}
		if info.Format != c.format || info.Width != 300 || info.Height != 200 || !near(info.PPIX, c.ppi) || !near(info.PPIY, c.ppi) {
			t.Errorf("%s: %+v", c.name, info)
		}
		w, h := info.Size()
		if !near(w, 300/c.ppi*72) || !near(h, 200/c.ppi*72) {
			t.Errorf("%s: size %v × %v pt", c.name, w, h)
		}
	}
	if _, err := ImageInfo([]byte("GIF89a")); err == nil {
		t.Error("not an image: no error")
	}
}

func TestNewEmbeddedImageRectangle(t *testing.T) {
	data := testJPEG(t, 400, 200)
	r, err := NewEmbeddedImageRectangle("r1", "i1", "l1", 10, 20, 100, 50, 0.25, 0.25, 10, 20,
		EmbeddedImage{Data: data, Name: "photo.jpg", Format: FormatJPEG, Width: 400, Height: 200, PPIX: 72, PPIY: 72})
	if err != nil {
		t.Fatal(err)
	}
	if r.ContentType != "GraphicType" || r.Image == nil || r.Image.Link == nil || r.Image.Link.StoredState != "Embedded" ||
		r.Image.Link.LinkResourceURI != "file:photo.jpg" || r.Image.Link.LinkResourceFormat != FormatJPEG ||
		r.Image.ImageTypeName != FormatJPEG || r.Image.ItemTransform != "0.25 0 0 0.25 10 20" || r.Image.ActualPpi != "72 72" {
		t.Fatalf("rectangle %+v, image %+v", r, r.Image)
	}

	// It writes and reads back with the file and its bounds.
	out, err := xml.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "<![CDATA[") || !strings.Contains(string(out), `StoredState="Embedded"`) {
		t.Errorf("written: %.300s", out)
	}
	var back Rectangle
	if err := xml.Unmarshal(out, &back); err != nil {
		t.Fatal(err)
	}
	got, ok, err := back.Image.EmbeddedContents()
	if err != nil || !ok || !bytes.Equal(got, data) {
		t.Errorf("contents: %d bytes, %v, %v", len(got), ok, err)
	}
	if l, tp, rt, bt, ok := back.Image.GraphicBounds(); !ok || l != 0 || tp != 0 || rt != 400 || bt != 200 {
		t.Errorf("graphic bounds %v %v %v %v %v", l, tp, rt, bt, ok)
	}

	for name, img := range map[string]EmbeddedImage{
		"no data":       {Format: FormatJPEG, Width: 1, Height: 1},
		"a PNG as JPEG": {Data: testPNG(t, 2, 2), Format: FormatJPEG, Width: 1, Height: 1},
		"a GIF":         {Data: []byte("GIF89a"), Format: "$ID/GIF", Width: 1, Height: 1},
		"no size":       {Data: data, Format: FormatJPEG},
	} {
		if _, err := NewEmbeddedImageRectangle("r", "i", "l", 0, 0, 1, 1, 1, 1, 0, 0, img); err == nil {
			t.Errorf("%s: no error", name)
		}
	}
	if _, err := NewEmbeddedImageRectangle("r", "i", "l", 0, 0, 1, 1, 0, 1, 0, 0,
		EmbeddedImage{Data: data, Format: FormatJPEG, Width: 1, Height: 1}); err == nil {
		t.Error("a zero scale: no error")
	}
}

// A linked image's bounds can be set, once, for InDesign to apply its scale.
func TestSetGraphicBounds(t *testing.T) {
	r := NewImageRectangle("r", "i", "l", 0, 0, 100, 50, 0.5, 0.5, 0, 0, "/tmp/photo.jpg", FormatJPEG)
	if _, _, _, _, ok := r.Image.GraphicBounds(); ok {
		t.Fatal("a new linked image has bounds")
	}
	r.Image.SetGraphicBounds(0, 0, 300, 100)
	r.Image.SetGraphicBounds(0, 0, 200, 100)
	if l, tp, rt, bt, ok := r.Image.GraphicBounds(); !ok || l != 0 || tp != 0 || rt != 200 || bt != 100 {
		t.Errorf("bounds %v %v %v %v %v", l, tp, rt, bt, ok)
	}
	n := 0
	for _, el := range r.Image.Properties.OtherElements {
		if el.XMLName.Local == "GraphicBounds" {
			n++
		}
	}
	if n != 1 {
		t.Errorf("%d GraphicBounds", n)
	}
	var nilImage *Image
	if _, ok, err := nilImage.EmbeddedContents(); ok || err != nil {
		t.Error("a nil image has contents")
	}
	if _, ok, _ := (&Image{Properties: &common.Properties{}}).EmbeddedContents(); ok {
		t.Error("an image without Contents has contents")
	}
}
