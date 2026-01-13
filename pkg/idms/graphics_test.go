package idms

import (
	"strings"
	"testing"
)

// TestParseGraphicsIDMS tests parsing the IDMS file with graphics
func TestParseGraphicsIDMS(t *testing.T) {
	pkg, err := Read("../../testdata/Snippet_31F27A387.idms")
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}

	// Verify basic structure
	if pkg == nil {
		t.Fatal("Package is nil")
	}

	if pkg.Document == nil {
		t.Fatal("Document is nil")
	}

	// Verify it contains more content than text-only IDMS
	if pkg.Document.Self != "d" {
		t.Errorf("Expected Document.Self = 'd', got '%s'", pkg.Document.Self)
	}

	t.Logf("✅ Successfully parsed graphics IDMS")
	t.Logf("   Document.Self: %s", pkg.Document.Self)
	t.Logf("   Document.DOMVersion: %s", pkg.Document.DOMVersion)
	t.Logf("   XMP Metadata: %d bytes", len(pkg.XMPMetadata))
}

// TestGraphicsIDMSHasImage tests that the graphics IDMS contains image references
func TestGraphicsIDMSHasImage(t *testing.T) {
	pkg, err := Read("../../testdata/Snippet_31F27A387.idms")
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}

	// Check if raw XML contains image-related elements
	rawXML := string(pkg.rawXML)

	// Verify Rectangle element exists
	if !strings.Contains(rawXML, "<Rectangle") {
		t.Error("Expected <Rectangle> element in graphics IDMS")
	}

	// Verify Image element exists
	if !strings.Contains(rawXML, "<Image") {
		t.Error("Expected <Image> element in graphics IDMS")
	}

	// Verify Link element exists (image file reference)
	if !strings.Contains(rawXML, "<Link") {
		t.Error("Expected <Link> element in graphics IDMS")
	}

	// Verify it references an image file
	if !strings.Contains(rawXML, ".jpg") {
		t.Error("Expected image file reference (.jpg) in graphics IDMS")
	}

	t.Logf("✅ Graphics IDMS contains:")
	t.Logf("   Rectangle element: %v", strings.Contains(rawXML, "<Rectangle"))
	t.Logf("   Image element: %v", strings.Contains(rawXML, "<Image"))
	t.Logf("   Link element: %v", strings.Contains(rawXML, "<Link"))
	t.Logf("   Image file reference: %v", strings.Contains(rawXML, ".jpg"))
}

// TestGraphicsIDMSStructure tests the structure of graphics elements
func TestGraphicsIDMSStructure(t *testing.T) {
	pkg, err := Read("../../testdata/Snippet_31F27A387.idms")
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}

	rawXML := string(pkg.rawXML)

	// Count graphic elements
	rectangleCount := strings.Count(rawXML, "<Rectangle")
	imageCount := strings.Count(rawXML, "<Image")
	linkCount := strings.Count(rawXML, "<Link")
	textFrameCount := strings.Count(rawXML, "<TextFrame")

	t.Logf("✅ Graphics IDMS structure:")
	t.Logf("   Rectangles: %d", rectangleCount)
	t.Logf("   Images: %d", imageCount)
	t.Logf("   Links: %d", linkCount)
	t.Logf("   TextFrames: %d", textFrameCount)

	if rectangleCount == 0 {
		t.Error("Expected at least one Rectangle element")
	}

	if imageCount == 0 {
		t.Error("Expected at least one Image element")
	}

	if linkCount == 0 {
		t.Error("Expected at least one Link element")
	}
}

// TestGraphicsIDMSHasColorSwatches tests that graphics IDMS has color definitions
func TestGraphicsIDMSHasColorSwatches(t *testing.T) {
	pkg, err := Read("../../testdata/Snippet_31F27A387.idms")
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}

	rawXML := string(pkg.rawXML)

	// Verify color-related elements
	if !strings.Contains(rawXML, "<Color") {
		t.Error("Expected <Color> element")
	}

	if !strings.Contains(rawXML, "<Swatch") {
		t.Error("Expected <Swatch> element")
	}

	t.Logf("✅ Graphics IDMS has color definitions")
}

// TestGraphicsIDMSHasStyles tests that graphics IDMS has style definitions
func TestGraphicsIDMSHasStyles(t *testing.T) {
	pkg, err := Read("../../testdata/Snippet_31F27A387.idms")
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}

	rawXML := string(pkg.rawXML)

	// Count style elements
	charStyleCount := strings.Count(rawXML, "<CharacterStyle")
	paraStyleCount := strings.Count(rawXML, "<ParagraphStyle")
	objStyleCount := strings.Count(rawXML, "<ObjectStyle")

	t.Logf("✅ Graphics IDMS styles:")
	t.Logf("   Character styles: %d", charStyleCount)
	t.Logf("   Paragraph styles: %d", paraStyleCount)
	t.Logf("   Object styles: %d", objStyleCount)

	if charStyleCount == 0 {
		t.Error("Expected at least one CharacterStyle")
	}

	if paraStyleCount == 0 {
		t.Error("Expected at least one ParagraphStyle")
	}

	if objStyleCount == 0 {
		t.Error("Expected at least one ObjectStyle")
	}
}
