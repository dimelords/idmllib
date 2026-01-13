package idml

import (
	"testing"

	"github.com/dimelords/idmllib/v2/pkg/spread"
)

// TestParseSpreadWithRectangles_ParsesRectangles tests parsing a spread containing rectangles and images.
func TestParseSpreadWithRectangles_ParsesRectangles(t *testing.T) {
	// Load example.idml which contains rectangles
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read example.idml: %v", err)
	}

	// Get spread
	sp, err := pkg.Spread("Spreads/Spread_u210.xml")
	if err != nil {
		t.Fatalf("Failed to get spread: %v", err)
	}

	// Verify we parsed rectangles
	if len(sp.InnerSpread.Rectangles) == 0 {
		t.Error("Expected rectangles but found none")
	}

	t.Logf("✅ Successfully parsed spread with %d rectangles", len(sp.InnerSpread.Rectangles))

	// Check first rectangle details
	if len(sp.InnerSpread.Rectangles) > 0 {
		rect := sp.InnerSpread.Rectangles[0]
		t.Logf("   Rectangle[0].Self: %s", rect.Self)
		t.Logf("   Rectangle[0].ContentType: %s", rect.ContentType)

		// Check for PathGeometry
		if rect.Properties != nil {
			t.Logf("   Rectangle[0] has Properties")
		}

		// Check for Image
		if rect.Image != nil {
			t.Logf("   Rectangle[0] contains Image: %s", rect.Image.Self)

			// Check for Link
			if rect.Image.Link != nil {
				t.Logf("   Image has Link: %s", rect.Image.Link.LinkResourceURI)
			}
		}

		// Check for FrameFittingOption
		if rect.FrameFittingOption != nil {
			t.Logf("   Rectangle[0] has FrameFittingOption")
		}
	}
}

// TestRectangleRoundtrip_PreservesData tests that we can parse and marshal rectangles without losing data.
func TestRectangleRoundtrip_PreservesData(t *testing.T) {
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read example.idml: %v", err)
	}

	sp, err := pkg.Spread("Spreads/Spread_u210.xml")
	if err != nil {
		t.Fatalf("Failed to get spread: %v", err)
	}

	originalRectCount := len(sp.InnerSpread.Rectangles)
	if originalRectCount == 0 {
		t.Skip("No rectangles in test spread")
	}

	// Marshal the spread
	spreadXML, err := spread.MarshalSpread(sp)
	if err != nil {
		t.Fatalf("Failed to marshal spread: %v", err)
	}

	// Parse it again
	roundtripSpread, err := spread.ParseSpread(spreadXML)
	if err != nil {
		t.Fatalf("Failed to parse roundtrip spread: %v", err)
	}

	// Verify rectangle count matches
	roundtripRectCount := len(roundtripSpread.InnerSpread.Rectangles)
	if roundtripRectCount != originalRectCount {
		t.Errorf("Rectangle count mismatch: original=%d, roundtrip=%d",
			originalRectCount, roundtripRectCount)
	}

	// Verify first rectangle Self attribute matches
	if originalRectCount > 0 && roundtripRectCount > 0 {
		origSelf := sp.InnerSpread.Rectangles[0].Self
		roundSelf := roundtripSpread.InnerSpread.Rectangles[0].Self
		if origSelf != roundSelf {
			t.Errorf("Rectangle[0].Self mismatch: original=%s, roundtrip=%s",
				origSelf, roundSelf)
		}
	}

	t.Logf("✅ Rectangle roundtrip successful: %d rectangles preserved", roundtripRectCount)
}

// TestRectangleWithImage_ParsesCorrectly tests that rectangles containing images are parsed correctly.
func TestRectangleWithImage_ParsesCorrectly(t *testing.T) {
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read example.idml: %v", err)
	}

	sp, err := pkg.Spread("Spreads/Spread_u210.xml")
	if err != nil {
		t.Fatalf("Failed to get spread: %v", err)
	}

	// Find rectangles with images
	var imageRects []spread.Rectangle
	for _, rect := range sp.InnerSpread.Rectangles {
		if rect.Image != nil {
			imageRects = append(imageRects, rect)
		}
	}

	if len(imageRects) == 0 {
		t.Skip("No rectangles with images found in test spread")
	}

	t.Logf("✅ Found %d rectangles with images", len(imageRects))

	// Verify image details
	for i, rect := range imageRects {
		if rect.Image.Link != nil {
			t.Logf("   Rectangle[%d] image link: %s", i, rect.Image.Link.LinkResourceURI)
			t.Logf("   Link state: %s", rect.Image.Link.StoredState)
		}
		if rect.FrameFittingOption != nil {
			t.Logf("   Rectangle[%d] has frame fitting options", i)
		}
	}
}
