package idms_test

import (
	"testing"

	"github.com/dimelords/idmllib/v2/pkg/idml"
	"github.com/dimelords/idmllib/v2/pkg/idms"
	"github.com/dimelords/idmllib/v2/pkg/spread"
)

func TestNewExporter(t *testing.T) {
	// Load test IDML package
	pkg, err := idml.Read("../../testdata/plain.idml")
	if err != nil {
		t.Fatalf("Failed to open test IDML: %v", err)
	}

	// Create exporter
	exporter := idms.NewExporter(pkg)
	if exporter == nil {
		t.Fatal("NewExporter returned nil")
	}

	// Verify dependencies are initialized
	deps := exporter.Dependencies()
	if deps == nil {
		t.Fatal("Dependencies are nil")
	}
}

func TestExportSelection_NilSelection(t *testing.T) {
	pkg, err := idml.Read("../../testdata/plain.idml")
	if err != nil {
		t.Fatalf("Failed to open test IDML: %v", err)
	}

	exporter := idms.NewExporter(pkg)

	// Test with nil selection
	_, err = exporter.ExportSelection(nil)
	if err == nil {
		t.Fatal("Expected error with nil selection, got nil")
	}
}

func TestExportSelection_EmptySelection(t *testing.T) {
	pkg, err := idml.Read("../../testdata/plain.idml")
	if err != nil {
		t.Fatalf("Failed to open test IDML: %v", err)
	}

	exporter := idms.NewExporter(pkg)

	// Test with empty selection
	sel := idml.NewSelection()
	_, err = exporter.ExportSelection(sel)
	if err == nil {
		t.Fatal("Expected error with empty selection, got nil")
	}
}

// TestExportSelection_BasicTextFrame tests exporting a single text frame.
// Currently returns "not yet implemented" as Phase 4.2 and 4.3 are pending.
func TestExportSelection_BasicTextFrame(t *testing.T) {
	pkg, err := idml.Read("../../testdata/plain.idml")
	if err != nil {
		t.Fatalf("Failed to open test IDML: %v", err)
	}

	// Get a text frame to export
	tf, err := pkg.SelectTextFrameByID("u116")
	if err != nil {
		t.Skipf("Could not find test text frame: %v", err)
	}

	// Create selection with the text frame
	sel := idml.NewSelection()
	sel.AddTextFrame(tf)

	exporter := idms.NewExporter(pkg)

	// Attempt export - should fail with "not yet implemented"
	_, err = exporter.ExportSelection(sel)
	if err == nil {
		t.Fatal("Expected error (not implemented), got nil")
	}

	// But dependencies should be collected
	deps := exporter.Dependencies()
	if len(deps.Stories) == 0 {
		t.Error("Expected at least one story dependency, got none")
	}
}

// TestResourceExtraction tests the resource extraction functionality.
func TestResourceExtraction(t *testing.T) {
	pkg, err := idml.Read("../../testdata/plain.idml")
	if err != nil {
		t.Fatalf("Failed to open test IDML: %v", err)
	}

	// Get all text frames
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Failed to get spreads: %v", err)
	}

	if len(spreads) == 0 {
		t.Skip("No spreads in test file")
	}

	// Find a text frame
	var testFrame *spread.SpreadTextFrame
	for _, spread := range spreads {
		if len(spread.InnerSpread.TextFrames) > 0 {
			testFrame = &spread.InnerSpread.TextFrames[0]
			break
		}
	}

	if testFrame == nil {
		t.Skip("No text frames found in test file")
	}

	// Create selection with the text frame
	sel := idml.NewSelection()
	sel.AddTextFrame(testFrame)

	exporter := idms.NewExporter(pkg)

	// Phase 4.3 NOW IMPLEMENTED - export should succeed
	result, err := exporter.ExportSelection(sel)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected IDMS package, got nil")
	}

	// Verify dependencies were collected
	deps := exporter.Dependencies()
	if len(deps.Stories) == 0 {
		t.Error("Expected at least one story dependency")
	}
	if len(deps.ParagraphStyles) == 0 {
		t.Error("Expected at least one paragraph style dependency")
	}
}

// ============================================================================
// Phase 5.1: Text-Only Export Tests
// ============================================================================

// TestExportTextFrame_Single tests exporting a single text frame as IDMS.
func TestExportTextFrame_Single(t *testing.T) {
	pkg, err := idml.Read("../../testdata/plain.idml")
	if err != nil {
		t.Fatalf("Failed to open test IDML: %v", err)
	}

	// Get all text frames from spreads
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Failed to get spreads: %v", err)
	}

	if len(spreads) == 0 {
		t.Skip("No spreads in test file")
	}

	// Find first text frame
	var testFrame *spread.SpreadTextFrame
	for _, spread := range spreads {
		if len(spread.InnerSpread.TextFrames) > 0 {
			testFrame = &spread.InnerSpread.TextFrames[0]
			break
		}
	}

	if testFrame == nil {
		t.Skip("No text frames found in test file")
	}

	// Create selection with single text frame
	sel := idml.NewSelection()
	sel.AddTextFrame(testFrame)

	// Export
	exporter := idms.NewExporter(pkg)
	result, err := exporter.ExportSelection(sel)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify result structure
	if result == nil {
		t.Fatal("Expected IDMS package, got nil")
	}

	if result.Document == nil {
		t.Fatal("Expected Document, got nil")
	}

	// Verify snippet type is set
	snippetType := result.SnippetType()
	if snippetType != "PageItem" {
		t.Errorf("Expected SnippetType 'PageItem', got '%s'", snippetType)
	}

	// Verify spread was created with text frame
	if len(result.Document.InlineSpreads) == 0 {
		t.Fatal("Expected at least one inline spread")
	}

	spread := result.Document.InlineSpreads[0]
	if len(spread.TextFrames) == 0 {
		t.Fatal("Expected at least one text frame in spread")
	}

	// Verify text frame was copied
	exportedFrame := spread.TextFrames[0]
	if exportedFrame.Self != testFrame.Self {
		t.Errorf("Expected frame ID '%s', got '%s'", testFrame.Self, exportedFrame.Self)
	}

	t.Logf("✅ Successfully exported single text frame")
	t.Logf("   Frame ID: %s", exportedFrame.Self)
	t.Logf("   Story reference: %s", exportedFrame.ParentStory)
}

// TestExportTextFrame_WithStory tests that referenced story is included in export.
func TestExportTextFrame_WithStory(t *testing.T) {
	pkg, err := idml.Read("../../testdata/plain.idml")
	if err != nil {
		t.Fatalf("Failed to open test IDML: %v", err)
	}

	// Get text frame with story
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Failed to get spreads: %v", err)
	}

	var testFrame *spread.SpreadTextFrame
	for _, spread := range spreads {
		for i := range spread.InnerSpread.TextFrames {
			if spread.InnerSpread.TextFrames[i].ParentStory != "" {
				testFrame = &spread.InnerSpread.TextFrames[i]
				break
			}
		}
		if testFrame != nil {
			break
		}
	}

	if testFrame == nil {
		t.Skip("No text frame with story found")
	}

	// Create selection
	sel := idml.NewSelection()
	sel.AddTextFrame(testFrame)

	// Export
	exporter := idms.NewExporter(pkg)
	result, err := exporter.ExportSelection(sel)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify story was included
	if len(result.Document.InlineStories) == 0 {
		t.Fatal("Expected at least one inline story")
	}

	// Find the referenced story
	storyFound := false
	for _, story := range result.Document.InlineStories {
		if story.Self == testFrame.ParentStory {
			storyFound = true
			t.Logf("✅ Story '%s' included in export", story.Self)
			break
		}
	}

	if !storyFound {
		t.Errorf("Referenced story '%s' not found in inline stories", testFrame.ParentStory)
	}
}

// TestExportTextFrame_WithStyles tests that referenced styles are included.
func TestExportTextFrame_WithStyles(t *testing.T) {
	pkg, err := idml.Read("../../testdata/plain.idml")
	if err != nil {
		t.Fatalf("Failed to open test IDML: %v", err)
	}

	// Get text frame
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Failed to get spreads: %v", err)
	}

	var testFrame *spread.SpreadTextFrame
	for _, spread := range spreads {
		if len(spread.InnerSpread.TextFrames) > 0 {
			testFrame = &spread.InnerSpread.TextFrames[0]
			break
		}
	}

	if testFrame == nil {
		t.Skip("No text frames found")
	}

	// Create selection
	sel := idml.NewSelection()
	sel.AddTextFrame(testFrame)

	// Export
	exporter := idms.NewExporter(pkg)
	result, err := exporter.ExportSelection(sel)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify style groups exist
	if result.Document.RootParagraphStyleGroup == nil {
		t.Error("Expected RootParagraphStyleGroup, got nil")
	}

	if result.Document.RootCharacterStyleGroup == nil {
		t.Error("Expected RootCharacterStyleGroup, got nil")
	}

	if result.Document.RootObjectStyleGroup == nil {
		t.Error("Expected RootObjectStyleGroup, got nil")
	}

	// Check dependencies to see what was extracted
	deps := exporter.Dependencies()
	t.Logf("✅ Style dependencies collected:")
	t.Logf("   Paragraph styles: %d", len(deps.ParagraphStyles))
	t.Logf("   Character styles: %d", len(deps.CharacterStyles))
	t.Logf("   Object styles: %d", len(deps.ObjectStyles))
}

// TestExportTextFrame_WithColors tests that referenced colors are included.
func TestExportTextFrame_WithColors(t *testing.T) {
	pkg, err := idml.Read("../../testdata/plain.idml")
	if err != nil {
		t.Fatalf("Failed to open test IDML: %v", err)
	}

	// Get text frame
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Failed to get spreads: %v", err)
	}

	var testFrame *spread.SpreadTextFrame
	for _, spread := range spreads {
		if len(spread.InnerSpread.TextFrames) > 0 {
			testFrame = &spread.InnerSpread.TextFrames[0]
			break
		}
	}

	if testFrame == nil {
		t.Skip("No text frames found")
	}

	// Create selection
	sel := idml.NewSelection()
	sel.AddTextFrame(testFrame)

	// Export
	exporter := idms.NewExporter(pkg)
	result, err := exporter.ExportSelection(sel)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify default colors are always included
	if len(result.Document.Colors) == 0 {
		t.Error("Expected at least default colors")
	}

	// Verify Black color exists
	blackFound := false
	for _, color := range result.Document.Colors {
		if color.Self == "Color/Black" {
			blackFound = true
			break
		}
	}

	if !blackFound {
		t.Error("Expected default 'Color/Black' in colors")
	}

	// Verify None swatch exists
	if len(result.Document.Swatches) == 0 {
		t.Error("Expected at least default swatches")
	}

	noneFound := false
	for _, swatch := range result.Document.Swatches {
		if swatch.Self == "Swatch/None" {
			noneFound = true
			break
		}
	}

	if !noneFound {
		t.Error("Expected default 'Swatch/None' in swatches")
	}

	t.Logf("✅ Color/swatch dependencies collected:")
	t.Logf("   Colors: %d", len(result.Document.Colors))
	t.Logf("   Swatches: %d", len(result.Document.Swatches))
}

// TestExportTextFrame_MultipleFrames tests exporting multiple text frames.
func TestExportTextFrame_MultipleFrames(t *testing.T) {
	pkg, err := idml.Read("../../testdata/plain.idml")
	if err != nil {
		t.Fatalf("Failed to open test IDML: %v", err)
	}

	// Get multiple text frames
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Failed to get spreads: %v", err)
	}

	var testFrames []*spread.SpreadTextFrame
	for _, spread := range spreads {
		for i := range spread.InnerSpread.TextFrames {
			testFrames = append(testFrames, &spread.InnerSpread.TextFrames[i])
			if len(testFrames) >= 3 {
				break
			}
		}
		if len(testFrames) >= 3 {
			break
		}
	}

	if len(testFrames) < 2 {
		t.Skip("Need at least 2 text frames for this test")
	}

	// Create selection with multiple frames
	sel := idml.NewSelection()
	for _, frame := range testFrames {
		sel.AddTextFrame(frame)
	}

	// Export
	exporter := idms.NewExporter(pkg)
	result, err := exporter.ExportSelection(sel)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify all frames are in export
	if len(result.Document.InlineSpreads) == 0 {
		t.Fatal("Expected at least one inline spread")
	}

	spread := result.Document.InlineSpreads[0]

	exportedCount := len(spread.TextFrames)

	if exportedCount != len(testFrames) {
		t.Errorf("Expected %d text frames, got %d", len(testFrames), exportedCount)
	}

	t.Logf("✅ Successfully exported %d text frames", exportedCount)
}

// ============================================================================
// Phase 5.2: Graphic Export Tests
// ============================================================================

// TestExportRectangle_Single tests exporting a single rectangle as IDMS.
func TestExportRectangle_Single(t *testing.T) {
	pkg, err := idml.Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to open test IDML: %v", err)
	}

	// Get all rectangles from spreads
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Failed to get spreads: %v", err)
	}

	if len(spreads) == 0 {
		t.Skip("No spreads in test file")
	}

	// Find first rectangle
	var testRect *spread.Rectangle
	for _, spread := range spreads {
		if len(spread.InnerSpread.Rectangles) > 0 {
			testRect = &spread.InnerSpread.Rectangles[0]
			break
		}
	}

	if testRect == nil {
		t.Skip("No rectangles found in test file")
	}

	// Create selection with single rectangle
	sel := idml.NewSelection()
	sel.AddRectangle(testRect)

	// Export
	exporter := idms.NewExporter(pkg)
	result, err := exporter.ExportSelection(sel)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify result structure
	if result == nil {
		t.Fatal("Expected IDMS package, got nil")
	}

	if result.Document == nil {
		t.Fatal("Expected Document, got nil")
	}

	// Verify snippet type is set
	snippetType := result.SnippetType()
	if snippetType != "PageItem" {
		t.Errorf("Expected SnippetType 'PageItem', got '%s'", snippetType)
	}

	// Verify spread was created with rectangle
	if len(result.Document.InlineSpreads) == 0 {
		t.Fatal("Expected at least one inline spread")
	}

	spread := result.Document.InlineSpreads[0]

	if len(spread.Rectangles) == 0 {
		t.Fatal("Expected at least one rectangle in spread")
	}

	// Verify rectangle was copied
	exportedRect := spread.Rectangles[0]
	if exportedRect.Self != testRect.Self {
		t.Errorf("Expected rectangle ID '%s', got '%s'", testRect.Self, exportedRect.Self)
	}

	t.Logf("✅ Successfully exported single rectangle")
	t.Logf("   Rectangle ID: %s", exportedRect.Self)
}

// TestExportRectangle_WithImage tests exporting rectangle containing an image.
func TestExportRectangle_WithImage(t *testing.T) {
	pkg, err := idml.Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to open test IDML: %v", err)
	}

	// Get rectangles
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Failed to get spreads: %v", err)
	}

	var testRect *spread.Rectangle
	for _, spread := range spreads {
		for i := range spread.InnerSpread.Rectangles {
			// Look for rectangle with image
			rect := &spread.InnerSpread.Rectangles[i]
			if rect.Image != nil {
				testRect = rect
				break
			}
		}
		if testRect != nil {
			break
		}
	}

	if testRect == nil {
		t.Skip("No rectangle with image found")
	}

	// Create selection
	sel := idml.NewSelection()
	sel.AddRectangle(testRect)

	// Export
	exporter := idms.NewExporter(pkg)
	result, err := exporter.ExportSelection(sel)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify rectangle and image are in export
	if len(result.Document.InlineSpreads) == 0 {
		t.Fatal("Expected at least one inline spread")
	}

	spread := result.Document.InlineSpreads[0]

	if len(spread.Rectangles) == 0 {
		t.Fatal("Expected at least one rectangle")
	}

	exportedRect := spread.Rectangles[0]
	if exportedRect.Image == nil {
		t.Error("Expected rectangle to contain image")
	}

	t.Logf("✅ Successfully exported rectangle with image")
	t.Logf("   Rectangle ID: %s", exportedRect.Self)
	if exportedRect.Image != nil {
		t.Logf("   Has image: true")
	}
}

// TestExportRectangle_WithLinks tests that image links are preserved.
func TestExportRectangle_WithLinks(t *testing.T) {
	pkg, err := idml.Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to open test IDML: %v", err)
	}

	// Get rectangles
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Failed to get spreads: %v", err)
	}

	var testRect *spread.Rectangle
	for _, spread := range spreads {
		for i := range spread.InnerSpread.Rectangles {
			rect := &spread.InnerSpread.Rectangles[i]
			if rect.Image != nil && rect.Image.Link != nil {
				testRect = rect
				break
			}
		}
		if testRect != nil {
			break
		}
	}

	if testRect == nil {
		t.Skip("No rectangle with image link found")
	}

	// Create selection
	sel := idml.NewSelection()
	sel.AddRectangle(testRect)

	// Export
	exporter := idms.NewExporter(pkg)
	result, err := exporter.ExportSelection(sel)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify links are preserved
	spread := result.Document.InlineSpreads[0]

	exportedRect := spread.Rectangles[0]

	if exportedRect.Image == nil {
		t.Fatal("Expected image in rectangle")
	}

	if exportedRect.Image.Link == nil {
		t.Error("Expected link in image")
	}

	t.Logf("✅ Successfully exported rectangle with image link")
	t.Logf("   Rectangle ID: %s", exportedRect.Self)
	if exportedRect.Image != nil {
		t.Logf("   Has image: true")
		if exportedRect.Image.Link != nil {
			t.Logf("   Has link: true")
		}
	}
}

// TestExportRectangle_MultipleRectangles tests exporting multiple rectangles.
func TestExportRectangle_MultipleRectangles(t *testing.T) {
	pkg, err := idml.Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to open test IDML: %v", err)
	}

	// Get multiple rectangles
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Failed to get spreads: %v", err)
	}

	var testRects []*spread.Rectangle
	for _, spread := range spreads {
		for i := range spread.InnerSpread.Rectangles {
			testRects = append(testRects, &spread.InnerSpread.Rectangles[i])
			if len(testRects) >= 3 {
				break
			}
		}
		if len(testRects) >= 3 {
			break
		}
	}

	if len(testRects) < 2 {
		t.Skip("Need at least 2 rectangles for this test")
	}

	// Create selection with multiple rectangles
	sel := idml.NewSelection()
	for _, rect := range testRects {
		sel.AddRectangle(rect)
	}

	// Export
	exporter := idms.NewExporter(pkg)
	result, err := exporter.ExportSelection(sel)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify all rectangles are in export
	if len(result.Document.InlineSpreads) == 0 {
		t.Fatal("Expected at least one inline spread")
	}

	spread := result.Document.InlineSpreads[0]

	exportedCount := len(spread.Rectangles)

	if exportedCount != len(testRects) {
		t.Errorf("Expected %d rectangles, got %d", len(testRects), exportedCount)
	}

	t.Logf("✅ Successfully exported %d rectangles", exportedCount)
}

// TestExportRectangle_WithObjectStyle tests that object styles are included.
func TestExportRectangle_WithObjectStyle(t *testing.T) {
	pkg, err := idml.Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to open test IDML: %v", err)
	}

	// Get rectangle
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Failed to get spreads: %v", err)
	}

	var testRect *spread.Rectangle
	for _, spread := range spreads {
		if len(spread.InnerSpread.Rectangles) > 0 {
			testRect = &spread.InnerSpread.Rectangles[0]
			break
		}
	}

	if testRect == nil {
		t.Skip("No rectangles found")
	}

	// Create selection
	sel := idml.NewSelection()
	sel.AddRectangle(testRect)

	// Export
	exporter := idms.NewExporter(pkg)
	result, err := exporter.ExportSelection(sel)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify object style group exists
	if result.Document.RootObjectStyleGroup == nil {
		t.Error("Expected RootObjectStyleGroup, got nil")
	}

	// Check dependencies
	deps := exporter.Dependencies()
	t.Logf("✅ Object style dependencies collected:")
	t.Logf("   Object styles: %d", len(deps.ObjectStyles))
}

// ============================================================================
// Phase 5.3: Mixed Selection Export Tests
// ============================================================================

// TestExportMixed_TextFrameAndRectangle tests exporting text frame + rectangle together.
func TestExportMixed_TextFrameAndRectangle(t *testing.T) {
	pkg, err := idml.Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to open test IDML: %v", err)
	}

	// Get spreads
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Failed to get spreads: %v", err)
	}

	if len(spreads) == 0 {
		t.Skip("No spreads in test file")
	}

	// Find a text frame and a rectangle
	var testFrame *spread.SpreadTextFrame
	var testRect *spread.Rectangle

	for _, spread := range spreads {
		if testFrame == nil && len(spread.InnerSpread.TextFrames) > 0 {
			testFrame = &spread.InnerSpread.TextFrames[0]
		}
		if testRect == nil && len(spread.InnerSpread.Rectangles) > 0 {
			testRect = &spread.InnerSpread.Rectangles[0]
		}
		if testFrame != nil && testRect != nil {
			break
		}
	}

	if testFrame == nil || testRect == nil {
		t.Skip("Need both text frame and rectangle for this test")
	}

	// Create mixed selection
	sel := idml.NewSelection()
	sel.AddTextFrame(testFrame)
	sel.AddRectangle(testRect)

	// Export
	exporter := idms.NewExporter(pkg)
	result, err := exporter.ExportSelection(sel)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify both elements are in export
	if len(result.Document.InlineSpreads) == 0 {
		t.Fatal("Expected at least one inline spread")
	}

	spread := result.Document.InlineSpreads[0]

	// Verify text frame
	if len(spread.TextFrames) == 0 {
		t.Error("Expected at least one text frame")
	}

	// Verify rectangle
	if len(spread.Rectangles) == 0 {
		t.Error("Expected at least one rectangle")
	}

	t.Logf("✅ Successfully exported mixed selection")
	t.Logf("   Text frames: %d", len(spread.TextFrames))
	t.Logf("   Rectangles: %d", len(spread.Rectangles))
}

// TestExportMixed_MultipleOfEachType tests exporting multiple text frames and rectangles.
func TestExportMixed_MultipleOfEachType(t *testing.T) {
	pkg, err := idml.Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to open test IDML: %v", err)
	}

	// Get spreads
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Failed to get spreads: %v", err)
	}

	// Collect multiple of each type
	var textFrames []*spread.SpreadTextFrame
	var rectangles []*spread.Rectangle

	for _, spread := range spreads {
		for i := range spread.InnerSpread.TextFrames {
			textFrames = append(textFrames, &spread.InnerSpread.TextFrames[i])
			if len(textFrames) >= 2 {
				break
			}
		}
		for i := range spread.InnerSpread.Rectangles {
			rectangles = append(rectangles, &spread.InnerSpread.Rectangles[i])
			if len(rectangles) >= 2 {
				break
			}
		}
		if len(textFrames) >= 2 && len(rectangles) >= 2 {
			break
		}
	}

	if len(textFrames) < 1 || len(rectangles) < 1 {
		t.Skip("Need at least one text frame and one rectangle")
	}

	// Create mixed selection
	sel := idml.NewSelection()
	for _, tf := range textFrames {
		sel.AddTextFrame(tf)
	}
	for _, rect := range rectangles {
		sel.AddRectangle(rect)
	}

	// Export
	exporter := idms.NewExporter(pkg)
	result, err := exporter.ExportSelection(sel)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify counts
	spread := result.Document.InlineSpreads[0]

	exportedFrames := len(spread.TextFrames)
	exportedRects := len(spread.Rectangles)

	if exportedFrames != len(textFrames) {
		t.Errorf("Expected %d text frames, got %d", len(textFrames), exportedFrames)
	}

	if exportedRects != len(rectangles) {
		t.Errorf("Expected %d rectangles, got %d", len(rectangles), exportedRects)
	}

	t.Logf("✅ Successfully exported multiple mixed elements")
	t.Logf("   Text frames: %d", exportedFrames)
	t.Logf("   Rectangles: %d", exportedRects)
}

// TestExportMixed_WithSharedDependencies tests that shared dependencies are handled correctly.
func TestExportMixed_WithSharedDependencies(t *testing.T) {
	pkg, err := idml.Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to open test IDML: %v", err)
	}

	// Get spreads
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Failed to get spreads: %v", err)
	}

	// Find text frame and rectangle
	var testFrame *spread.SpreadTextFrame
	var testRect *spread.Rectangle

	for _, spread := range spreads {
		if len(spread.InnerSpread.TextFrames) > 0 {
			testFrame = &spread.InnerSpread.TextFrames[0]
		}
		if len(spread.InnerSpread.Rectangles) > 0 {
			testRect = &spread.InnerSpread.Rectangles[0]
		}
		if testFrame != nil && testRect != nil {
			break
		}
	}

	if testFrame == nil || testRect == nil {
		t.Skip("Need both text frame and rectangle")
	}

	// Create mixed selection
	sel := idml.NewSelection()
	sel.AddTextFrame(testFrame)
	sel.AddRectangle(testRect)

	// Export
	exporter := idms.NewExporter(pkg)
	result, err := exporter.ExportSelection(sel)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify style groups exist (shared dependency)
	if result.Document.RootParagraphStyleGroup == nil {
		t.Error("Expected RootParagraphStyleGroup")
	}

	if result.Document.RootObjectStyleGroup == nil {
		t.Error("Expected RootObjectStyleGroup")
	}

	// Both text frames and rectangles use colors/swatches
	if len(result.Document.Colors) == 0 {
		t.Error("Expected colors (shared dependency)")
	}

	// Check dependency collection
	deps := exporter.Dependencies()
	t.Logf("✅ Shared dependencies collected:")
	t.Logf("   Paragraph styles: %d", len(deps.ParagraphStyles))
	t.Logf("   Object styles: %d", len(deps.ObjectStyles))
	t.Logf("   Colors: %d", len(deps.Colors))
}

// TestExportMixed_CompleteDocument tests exporting all elements from a spread.
func TestExportMixed_CompleteDocument(t *testing.T) {
	pkg, err := idml.Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to open test IDML: %v", err)
	}

	// Get first spread
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Failed to get spreads: %v", err)
	}

	if len(spreads) == 0 {
		t.Skip("No spreads in test file")
	}

	// Get first spread from map
	var sp *spread.Spread
	for _, s := range spreads {
		sp = s
		break
	}

	// Create selection with ALL elements from the spread
	sel := idml.NewSelection()

	for i := range sp.InnerSpread.TextFrames {
		sel.AddTextFrame(&sp.InnerSpread.TextFrames[i])
	}

	for i := range sp.InnerSpread.Rectangles {
		sel.AddRectangle(&sp.InnerSpread.Rectangles[i])
	}

	totalElements := len(sp.InnerSpread.TextFrames) + len(sp.InnerSpread.Rectangles)

	if totalElements == 0 {
		t.Skip("No elements in spread")
	}

	// Export
	exporter := idms.NewExporter(pkg)
	result, err := exporter.ExportSelection(sel)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify all elements exported
	exportedSpread := result.Document.InlineSpreads[0]
	exportedTotal := len(exportedSpread.TextFrames) + len(exportedSpread.Rectangles)

	if exportedTotal != totalElements {
		t.Errorf("Expected %d total elements, got %d", totalElements, exportedTotal)
	}

	t.Logf("✅ Successfully exported complete spread")
	t.Logf("   Text frames: %d", len(exportedSpread.TextFrames))
	t.Logf("   Rectangles: %d", len(exportedSpread.Rectangles))
	t.Logf("   Total elements: %d", exportedTotal)
}
