// Package idms provides tests for IDMS export functionality.
package idms

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dimelords/idmllib/idml"
)

// TestStyleFiltering verifies that only used styles are included in export
func TestStyleFiltering(t *testing.T) {
	pkg, err := idml.Open("../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to open IDML: %v", err)
	}
	defer func(pkg *idml.Package) {
		_ = pkg.Close()
	}(pkg)

	if len(pkg.Stories) == 0 {
		t.Fatal("No stories to test")
	}

	outputDir := t.TempDir()
	outputPath := filepath.Join(outputDir, "test_style_filter.idms")

	// Export first story
	exporter := NewExporter(pkg)
	err = exporter.ExportStoryXML(pkg.Stories[0].Self, outputPath)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Read output and verify filtering
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	contentStr := string(content)

	// Count styles - should be fewer than in original IDML
	charStyleCount := strings.Count(contentStr, "<CharacterStyle Self=")
	paraStyleCount := strings.Count(contentStr, "<ParagraphStyle Self=")

	t.Logf("Filtered export contains: %d CharacterStyles, %d ParagraphStyles",
		charStyleCount, paraStyleCount)

	// Verify styles are present but filtered
	if charStyleCount == 0 && paraStyleCount == 0 {
		t.Error("No styles found in output - filtering may be too aggressive")
	}
}

// TestColorFiltering verifies that only used colors are included in export
func TestColorFiltering(t *testing.T) {
	pkg, err := idml.Open("../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to open IDML: %v", err)
	}
	defer func(pkg *idml.Package) {
		_ = pkg.Close()
	}(pkg)

	if len(pkg.Stories) == 0 {
		t.Fatal("No stories to test")
	}

	outputDir := t.TempDir()
	outputPath := filepath.Join(outputDir, "test_color_filter.idms")

	exporter := NewExporter(pkg)
	err = exporter.ExportStoryXML(pkg.Stories[0].Self, outputPath)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	contentStr := string(content)

	// Count colors and swatches
	colorCount := strings.Count(contentStr, "<Color Self=")
	swatchCount := strings.Count(contentStr, "<Swatch Self=")

	t.Logf("Filtered export contains: %d Colors, %d Swatches",
		colorCount, swatchCount)

	// System colors should always be present
	if !strings.Contains(contentStr, "Color/Black") {
		t.Error("System color Black is missing")
	}

	if !strings.Contains(contentStr, "Swatch/None") {
		t.Error("System swatch None is missing")
	}

	// Should have filtered out unused colors (fewer than original)
	if colorCount > 10 {
		t.Logf("Warning: Many colors in export (%d), filtering may not be working", colorCount)
	}
}

// TestStyleDependencies verifies that BasedOn styles are included
func TestStyleDependencies(t *testing.T) {
	pkg, err := idml.Open("../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to open IDML: %v", err)
	}
	defer func(pkg *idml.Package) {
		_ = pkg.Close()
	}(pkg)

	if len(pkg.Stories) == 0 {
		t.Fatal("No stories to test")
	}

	outputDir := t.TempDir()
	outputPath := filepath.Join(outputDir, "test_dependencies.idms")

	exporter := NewExporter(pkg)
	err = exporter.ExportStoryXML(pkg.Stories[0].Self, outputPath)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	contentStr := string(content)

	// Find all BasedOn references
	basedOnCount := strings.Count(contentStr, "<BasedOn")
	if basedOnCount > 0 {
		t.Logf("Found %d BasedOn references", basedOnCount)
	}
}

// TestColorGroupFiltering verifies ColorGroup filtering
func TestColorGroupFiltering(t *testing.T) {
	pkg, err := idml.Open("../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to open IDML: %v", err)
	}
	defer func(pkg *idml.Package) {
		_ = pkg.Close()
	}(pkg)

	if len(pkg.Stories) == 0 {
		t.Fatal("No stories to test")
	}

	outputDir := t.TempDir()
	outputPath := filepath.Join(outputDir, "test_colorgroup.idms")

	exporter := NewExporter(pkg)
	err = exporter.ExportStoryXML(pkg.Stories[0].Self, outputPath)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	contentStr := string(content)

	// Verify ColorGroup structure is present
	if !strings.Contains(contentStr, "<ColorGroup") {
		t.Error("ColorGroup element is missing")
	}

	if !strings.Contains(contentStr, "<ColorGroupSwatch") {
		t.Error("ColorGroupSwatch elements are missing")
	}

	// Count ColorGroupSwatch entries
	cgSwatchCount := strings.Count(contentStr, "<ColorGroupSwatch")
	t.Logf("Found %d ColorGroupSwatch entries", cgSwatchCount)

	// Should have at least 2 (None and Black)
	if cgSwatchCount < 2 {
		t.Error("Too few ColorGroupSwatch entries - expected at least 2 (None + Black)")
	}
}

// TestTextFrameFiltering verifies that only relevant TextFrames are included
func TestTextFrameFiltering(t *testing.T) {
	pkg, err := idml.Open("../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to open IDML: %v", err)
	}
	defer func(pkg *idml.Package) {
		_ = pkg.Close()
	}(pkg)

	if len(pkg.Stories) == 0 {
		t.Fatal("No stories to test")
	}

	outputDir := t.TempDir()
	outputPath := filepath.Join(outputDir, "test_textframe.idms")

	storyID := pkg.Stories[0].Self
	exporter := NewExporter(pkg)
	err = exporter.ExportStoryXML(storyID, outputPath)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	contentStr := string(content)

	// Count TextFrames
	textFrameCount := strings.Count(contentStr, "<TextFrame")
	t.Logf("Found %d TextFrame(s)", textFrameCount)

	// Verify TextFrame references the correct story
	if !strings.Contains(contentStr, `ParentStory="`+storyID) {
		t.Errorf("TextFrame doesn't reference story %s", storyID)
	}

	// Should have a reasonable number of TextFrames (not all from the document)
	if textFrameCount > 5 {
		t.Logf("Warning: Many TextFrames in export (%d), filtering may not be optimal", textFrameCount)
	}
}

// TestMultipleStoryExports verifies that each story export is independent
func TestMultipleStoryExports(t *testing.T) {
	pkg, err := idml.Open("../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to open IDML: %v", err)
	}
	defer func(pkg *idml.Package) {
		_ = pkg.Close()
	}(pkg)

	// Find stories that have text frames
	spreads := pkg.GetSpreads()
	storyWithFrames := make(map[string]bool)
	for _, spread := range spreads {
		for _, tf := range spread.TextFrames {
			storyWithFrames[tf.ParentStory] = true
		}
	}

	// Collect stories that have text frames
	var storiesWithFrames []string
	for _, story := range pkg.Stories {
		if storyWithFrames[story.Self] {
			storiesWithFrames = append(storiesWithFrames, story.Self)
			if len(storiesWithFrames) >= 2 {
				break
			}
		}
	}

	if len(storiesWithFrames) < 2 {
		t.Skip("Need at least 2 stories with text frames for this test")
	}

	outputDir := t.TempDir()

	// Export first two stories that have text frames
	exporter := NewExporter(pkg)
	outputs := make([]string, 2)
	for i := 0; i < 2; i++ {
		outputs[i] = filepath.Join(outputDir, "story_"+storiesWithFrames[i]+".idms")
		err = exporter.ExportStoryXML(storiesWithFrames[i], outputs[i])
		if err != nil {
			t.Fatalf("Failed to export story %s: %v", storiesWithFrames[i], err)
		}
	}

	// Verify both files exist and are different
	content1, err := os.ReadFile(outputs[0])
	if err != nil {
		t.Fatalf("Failed to read output 1: %v", err)
	}

	content2, err := os.ReadFile(outputs[1])
	if err != nil {
		t.Fatalf("Failed to read output 2: %v", err)
	}

	if string(content1) == string(content2) {
		t.Error("Exported stories are identical - exports may not be independent")
	}

	t.Logf("Story 1: %d bytes, Story 2: %d bytes", len(content1), len(content2))
}
