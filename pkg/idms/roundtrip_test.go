package idms

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRoundtripTextOnly tests reading and writing the text-only IDMS file
func TestRoundtripTextOnly(t *testing.T) {
	// Read original
	original, err := Read("../../testdata/Snippet_31F27A2D0.idms")
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}

	// Write to temp file
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "roundtrip.idms")

	if err := Write(original, outPath); err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	// Read back
	roundtrip, err := Read(outPath)
	if err != nil {
		t.Fatalf("Read roundtrip error: %v", err)
	}

	// Compare structures
	if roundtrip.Document == nil {
		t.Fatal("Roundtrip document is nil")
	}

	if roundtrip.Document.Self != original.Document.Self {
		t.Errorf("Document.Self mismatch: got '%s', want '%s'",
			roundtrip.Document.Self, original.Document.Self)
	}

	if roundtrip.Document.DOMVersion != original.Document.DOMVersion {
		t.Errorf("Document.DOMVersion mismatch: got '%s', want '%s'",
			roundtrip.Document.DOMVersion, original.Document.DOMVersion)
	}

	// Verify processing instructions preserved
	if len(roundtrip.AIDProcessingInstructions) != len(original.AIDProcessingInstructions) {
		t.Errorf("AID PI count mismatch: got %d, want %d",
			len(roundtrip.AIDProcessingInstructions), len(original.AIDProcessingInstructions))
	}

	// Verify XMP preserved
	if roundtrip.XMPMetadata == "" {
		t.Error("XMP metadata lost in roundtrip")
	}

	t.Logf("✅ Roundtrip successful")
	t.Logf("   Document structure preserved: %v", roundtrip.Document != nil)
	t.Logf("   AID PIs preserved: %d", len(roundtrip.AIDProcessingInstructions))
	t.Logf("   XMP preserved: %d bytes", len(roundtrip.XMPMetadata))
}

// TestRoundtripWithGraphics tests reading and writing the IDMS file with graphics
func TestRoundtripWithGraphics(t *testing.T) {
	// Read original
	original, err := Read("../../testdata/Snippet_31F27A387.idms")
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}

	// Write to temp file
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "roundtrip_graphics.idms")

	if err := Write(original, outPath); err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	// Read back
	roundtrip, err := Read(outPath)
	if err != nil {
		t.Fatalf("Read roundtrip error: %v", err)
	}

	// Compare structures
	if roundtrip.Document == nil {
		t.Fatal("Roundtrip document is nil")
	}

	if roundtrip.Document.Self != original.Document.Self {
		t.Errorf("Document.Self mismatch: got '%s', want '%s'",
			roundtrip.Document.Self, original.Document.Self)
	}

	// Verify graphics elements are preserved by checking the marshaled XML
	marshaledData, err := Marshal(roundtrip)
	if err != nil {
		t.Fatalf("Marshal roundtrip error: %v", err)
	}

	marshaledXML := string(marshaledData)

	// Count key graphics elements in both original and roundtrip
	origRectCount := countElements(string(original.rawXML), "<Rectangle")
	rtRectCount := countElements(marshaledXML, "<Rectangle")

	origImageCount := countElements(string(original.rawXML), "<Image")
	rtImageCount := countElements(marshaledXML, "<Image")

	origLinkCount := countElements(string(original.rawXML), "<Link")
	rtLinkCount := countElements(marshaledXML, "<Link")

	// Verify counts match
	if rtRectCount != origRectCount {
		t.Errorf("Rectangle count mismatch: got %d, want %d", rtRectCount, origRectCount)
	}
	if rtImageCount != origImageCount {
		t.Errorf("Image count mismatch: got %d, want %d", rtImageCount, origImageCount)
	}
	if rtLinkCount != origLinkCount {
		t.Errorf("Link count mismatch: got %d, want %d", rtLinkCount, origLinkCount)
	}

	t.Logf("✅ Graphics roundtrip successful")
	t.Logf("   Document structure preserved: %v", roundtrip.Document != nil)
	t.Logf("   Rectangles: %d", rtRectCount)
	t.Logf("   Images: %d", rtImageCount)
	t.Logf("   Links: %d", rtLinkCount)
}

// Helper function for counting element occurrences
func countElements(xml string, tag string) int {
	return strings.Count(xml, tag)
}

// TestMarshalTextOnly tests marshaling the text-only IDMS
func TestMarshalTextOnly(t *testing.T) {
	pkg, err := Read("../../testdata/Snippet_31F27A2D0.idms")
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}

	data, err := Marshal(pkg)
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("Marshaled data is empty")
	}

	// Verify it starts with XML declaration
	if len(data) < 5 || string(data[:5]) != "<?xml" {
		t.Error("Marshaled data does not start with XML declaration")
	}

	t.Logf("✅ Successfully marshaled IDMS")
	t.Logf("   Output size: %d bytes", len(data))

	// Save for manual inspection if needed
	tmpDir := t.TempDir()
	debugPath := filepath.Join(tmpDir, "marshaled.idms")
	if err := os.WriteFile(debugPath, data, 0644); err != nil {
		t.Logf("Could not write debug file: %v", err)
	} else {
		t.Logf("   Debug output: %s", debugPath)
	}
}

// TestSpreadAndStoryParsing tests that inline spreads and stories are properly parsed
func TestSpreadAndStoryParsing(t *testing.T) {
	pkg, err := Read("../../testdata/Snippet_31F27A2D0.idms")
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}

	// Test that Document.InlineSpreads and InlineStories are populated
	if len(pkg.Document.InlineSpreads) == 0 {
		t.Error("Expected inline spreads to be parsed, but got 0")
	}

	if len(pkg.Document.InlineStories) == 0 {
		t.Error("Expected inline stories to be parsed, but got 0")
	}

	// Test the accessor methods
	spreads := pkg.Spreads()
	stories := pkg.Stories()

	if len(spreads) == 0 {
		t.Error("Spreads() returned 0 spreads, expected at least 1")
	}

	if len(stories) == 0 {
		t.Error("Stories() returned 0 stories, expected at least 1")
	}

	// Verify spread structure
	if len(spreads) > 0 {
		spread := spreads[0]
		if spread.Self == "" {
			t.Error("First spread has empty Self attribute")
		}
		if len(spread.TextFrames) == 0 {
			t.Error("Expected text frames in spread, but got 0")
		}
		t.Logf("✅ Spread parsed correctly")
		t.Logf("   Self: %s", spread.Self)
		t.Logf("   TextFrames: %d", len(spread.TextFrames))
	}

	// Verify story structure
	if len(stories) > 0 {
		story := stories[0]
		if story.Self == "" {
			t.Error("First story has empty Self attribute")
		}
		t.Logf("✅ Story parsed correctly")
		t.Logf("   Self: %s", story.Self)
	}

	// Test roundtrip preservation
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "spread_story_roundtrip.idms")

	if err := Write(pkg, outPath); err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	// Read back and verify
	roundtrip, err := Read(outPath)
	if err != nil {
		t.Fatalf("Read roundtrip error: %v", err)
	}

	rtSpreads := roundtrip.Spreads()
	rtStories := roundtrip.Stories()

	if len(rtSpreads) != len(spreads) {
		t.Errorf("Roundtrip spread count mismatch: got %d, want %d", len(rtSpreads), len(spreads))
	}

	if len(rtStories) != len(stories) {
		t.Errorf("Roundtrip story count mismatch: got %d, want %d", len(rtStories), len(stories))
	}

	// Verify spread details preserved
	if len(rtSpreads) > 0 && len(spreads) > 0 {
		if rtSpreads[0].Self != spreads[0].Self {
			t.Errorf("Spread Self mismatch: got %s, want %s", rtSpreads[0].Self, spreads[0].Self)
		}
		if len(rtSpreads[0].TextFrames) != len(spreads[0].TextFrames) {
			t.Errorf("TextFrame count mismatch: got %d, want %d",
				len(rtSpreads[0].TextFrames), len(spreads[0].TextFrames))
		}
	}

	// Verify story details preserved
	if len(rtStories) > 0 && len(stories) > 0 {
		if rtStories[0].Self != stories[0].Self {
			t.Errorf("Story Self mismatch: got %s, want %s", rtStories[0].Self, stories[0].Self)
		}
	}

	t.Logf("✅ Spread and story roundtrip successful")
	t.Logf("   Original spreads: %d, roundtrip: %d", len(spreads), len(rtSpreads))
	t.Logf("   Original stories: %d, roundtrip: %d", len(stories), len(rtStories))
}
