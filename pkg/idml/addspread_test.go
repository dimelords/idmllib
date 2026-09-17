package idml

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/dimelords/idmllib/v3/pkg/spread"
)

// newTestSpread returns a minimal one-page spread.
func newTestSpread(self string) *spread.Spread {
	sp := &spread.Spread{}
	sp.Self = self
	sp.Pages = []spread.Page{{
		Self:            "up_" + self,
		AppliedMaster:   "ub4",
		GeometricBounds: "0 0 792 612",
	}}
	return sp
}

// TestAddSpreadRegistersInDesignmap checks that a spread added through the
// API is referenced from designmap.xml. A spread file present in the archive
// but not referenced there is ignored by InDesign, the same trap AddStory had.
func TestAddSpreadRegistersInDesignmap(t *testing.T) {
	pkg, err := NewFromTemplate(nil)
	if err != nil {
		t.Fatal(err)
	}
	const id = "uspread1"
	filename := SpreadPath(id)

	if err := pkg.AddSpread(filename, newTestSpread(id)); err != nil {
		t.Fatalf("AddSpread: %v", err)
	}
	if err := pkg.AddSpread(filename, newTestSpread(id)); err == nil {
		t.Error("adding the same spread twice should fail")
	}

	out := filepath.Join(t.TempDir(), "out.idml")
	if err := Write(pkg, out); err != nil {
		t.Fatal(err)
	}
	pkg2, err := Read(out)
	if err != nil {
		t.Fatal(err)
	}

	doc, err := pkg2.Document()
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, ref := range doc.Spreads {
		if ref.Src == filename {
			found = true
		}
	}
	if !found {
		t.Errorf("designmap has no idPkg:Spread reference to %s", filename)
	}
	designmap := string(mustFileData(t, pkg2, PathDesignmap))
	if !strings.Contains(designmap, `<idPkg:Spread src="`+filename+`"`) {
		t.Errorf("designmap.xml has no prefixed idPkg:Spread element for %s", filename)
	}

	sp, err := pkg2.Spread(filename)
	if err != nil {
		t.Fatalf("Spread(%s): %v", filename, err)
	}
	if sp.Self != id {
		t.Errorf("Self = %q, want %q", sp.Self, id)
	}
	if len(sp.Pages) != 1 || sp.Pages[0].AppliedMaster != "ub4" {
		t.Errorf("page not preserved: %+v", sp.Pages)
	}
}

// TestRectangleFillAndStrokeAreTyped covers the six attributes Rectangle was
// missing while Oval and Polygon already had them.
func TestRectangleFillAndStrokeAreTyped(t *testing.T) {
	pkg, err := NewFromTemplate(nil)
	if err != nil {
		t.Fatal(err)
	}
	const spreadID = "uspread2"
	filename := SpreadPath(spreadID)
	if err := pkg.AddSpread(filename, newTestSpread(spreadID)); err != nil {
		t.Fatal(err)
	}

	rect := &spread.Rectangle{}
	rect.Self = "urect1"
	rect.ItemTransform = "1 0 0 1 0 0"
	rect.FillColor = "Color/Black"
	rect.FillTint = "50"
	rect.StrokeColor = "Color/Black"
	rect.StrokeWeight = "2"
	rect.StrokeType = "StrokeStyle/$ID/Solid"
	rect.StrokeTint = "100"
	if err := pkg.AddRectangle(filename, rect, ValidationOptions{}); err != nil {
		t.Fatalf("AddRectangle: %v", err)
	}

	out := filepath.Join(t.TempDir(), "out.idml")
	if err := Write(pkg, out); err != nil {
		t.Fatal(err)
	}
	pkg2, err := Read(out)
	if err != nil {
		t.Fatal(err)
	}
	sp, err := pkg2.Spread(filename)
	if err != nil {
		t.Fatal(err)
	}
	if len(sp.Rectangles) != 1 {
		t.Fatalf("got %d rectangles, want 1", len(sp.Rectangles))
	}
	got := sp.Rectangles[0]
	for _, c := range []struct{ name, got, want string }{
		{"FillColor", got.FillColor, "Color/Black"},
		{"FillTint", got.FillTint, "50"},
		{"StrokeColor", got.StrokeColor, "Color/Black"},
		{"StrokeWeight", got.StrokeWeight, "2"},
		{"StrokeType", got.StrokeType, "StrokeStyle/$ID/Solid"},
		{"StrokeTint", got.StrokeTint, "100"},
	} {
		if c.got != c.want {
			t.Errorf("%s = %q, want %q", c.name, c.got, c.want)
		}
	}
}

// TestNewImageRectangle covers the placed-photo constructor: a graphic
// Rectangle wrapping an Image with a file Link, with four-corner geometry.
func TestNewImageRectangle(t *testing.T) {
	rect := spread.NewImageRectangle(
		"urect", "uimage", "ulink",
		10, 20, 100, 50,
		1.5, 1.5, 3, 4,
		"/tmp/photo.jpg", "$ID/JPEG",
	)

	if rect.Self != "urect" || rect.ContentType != "GraphicType" {
		t.Errorf("Self=%q ContentType=%q", rect.Self, rect.ContentType)
	}
	// Scribus's IDML importer skips page items with no ItemTransform.
	if rect.ItemTransform != "1 0 0 1 0 0" {
		t.Errorf("ItemTransform = %q, want the identity transform", rect.ItemTransform)
	}
	if rect.Image == nil || rect.Image.Self != "uimage" {
		t.Fatalf("image not set: %+v", rect.Image)
	}
	if rect.Image.ImageTypeName != "$ID/JPEG" {
		t.Errorf("ImageTypeName = %q", rect.Image.ImageTypeName)
	}
	if want := "1.5 0 0 1.5 3 4"; rect.Image.ItemTransform != want {
		t.Errorf("ItemTransform = %q, want %q", rect.Image.ItemTransform, want)
	}
	link := rect.Image.Link
	if link == nil || link.LinkResourceURI != "file:/tmp/photo.jpg" {
		t.Fatalf("link not built correctly: %+v", link)
	}
	if link.LinkResourceFormat != "$ID/JPEG" || link.StoredState != "Normal" {
		t.Errorf("link format=%q state=%q", link.LinkResourceFormat, link.StoredState)
	}

	// Four corners, anchors in X Y order, clockwise from the frame origin.
	pts := rect.Properties.PathGeometry.GeometryPathType.PathPointArray.PathPoints
	want := []string{"10 20", "10 70", "110 70", "110 20"}
	if len(pts) != len(want) {
		t.Fatalf("got %d path points, want %d", len(pts), len(want))
	}
	for i, w := range want {
		if pts[i].Anchor != w {
			t.Errorf("point %d anchor = %q, want %q", i, pts[i].Anchor, w)
		}
		if pts[i].LeftDirection != w || pts[i].RightDirection != w {
			t.Errorf("point %d directions should match the anchor", i)
		}
	}
}

// TestAddedItemsGetAnItemTransform locks in the default. Every page item in
// real IDML carries ItemTransform, and Scribus's importer silently skips any
// item that lacks it, so an item added without one must still get it.
func TestAddedItemsGetAnItemTransform(t *testing.T) {
	pkg, err := NewFromTemplate(nil)
	if err != nil {
		t.Fatal(err)
	}
	filename := SpreadPath("uspread3")
	if err := pkg.AddSpread(filename, newTestSpread("uspread3")); err != nil {
		t.Fatal(err)
	}

	rect := &spread.Rectangle{}
	rect.Self = "urect_noxform"
	if err := pkg.AddRectangle(filename, rect, ValidationOptions{}); err != nil {
		t.Fatal(err)
	}
	tf := &spread.TextFrame{}
	tf.Self = "utf_noxform"
	if err := pkg.AddTextFrame(filename, tf, ValidationOptions{}); err != nil {
		t.Fatal(err)
	}
	// An explicit transform must be left alone.
	kept := &spread.Rectangle{}
	kept.Self = "urect_keep"
	kept.ItemTransform = "2 0 0 2 10 20"
	if err := pkg.AddRectangle(filename, kept, ValidationOptions{}); err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(t.TempDir(), "out.idml")
	if err := Write(pkg, out); err != nil {
		t.Fatal(err)
	}
	pkg2, err := Read(out)
	if err != nil {
		t.Fatal(err)
	}
	sp, err := pkg2.Spread(filename)
	if err != nil {
		t.Fatal(err)
	}

	got := map[string]string{}
	for _, r := range sp.Rectangles {
		got[r.Self] = r.ItemTransform
	}
	for _, f := range sp.TextFrames {
		got[f.Self] = f.ItemTransform
	}
	for self, want := range map[string]string{
		"urect_noxform": "1 0 0 1 0 0",
		"utf_noxform":   "1 0 0 1 0 0",
		"urect_keep":    "2 0 0 2 10 20",
	} {
		if got[self] != want {
			t.Errorf("%s ItemTransform = %q, want %q", self, got[self], want)
		}
	}
}
