package idml

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dimelords/idmllib/idms"
)

// TestOpen verifies that we can open and parse an IDML file
func TestOpen(t *testing.T) {
	pkg, err := Open("../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to open IDML: %v", err)
	}
	defer pkg.Close()

	if pkg.reader == nil {
		t.Error("Package reader is nil")
	}

	if len(pkg.Stories) == 0 {
		t.Error("Expected stories, got none")
	}

	t.Logf("Successfully opened IDML with %d stories", len(pkg.Stories))
}

// TestOpenNonExistentFile verifies error handling for missing files
func TestOpenNonExistentFile(t *testing.T) {
	_, err := Open("../testdata/nonexistent.idml")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

// TestExportStoryAsIDMSXML verifies XML export functionality
func TestExportStoryAsIDMSXML(t *testing.T) {
	pkg, err := Open("../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to open IDML: %v", err)
	}
	defer pkg.Close()

	if len(pkg.Stories) == 0 {
		t.Fatal("No stories to export")
	}

	outputDir := t.TempDir()
	outputPath := filepath.Join(outputDir, "test_export_xml.idms")

	// Create IDMS exporter and export first story as XML
	exporter := idms.NewExporter(pkg)
	err = exporter.ExportStoryXML(pkg.Stories[0].Self, outputPath)
	if err != nil {
		t.Fatalf("XML export failed: %v", err)
	}

	// Verify file exists and has content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if len(content) == 0 {
		t.Error("Output file is empty")
	}

	t.Logf("Successfully exported story as XML (%d bytes)", len(content))
}

// TestGetStory verifies that we can retrieve a story by ID
func TestGetStory(t *testing.T) {
	pkg, err := Open("../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to open IDML: %v", err)
	}
	defer func(pkg *Package) {
		_ = pkg.Close()
	}(pkg)

	if len(pkg.Stories) == 0 {
		t.Fatal("No stories found")
	}

	// Get first story
	storyID := pkg.Stories[0].Self
	story, err := pkg.GetStory(storyID)
	if err != nil {
		t.Fatalf("Failed to get story: %v", err)
	}

	if story == nil {
		t.Fatal("Story is nil")
	}

	if story.Self != storyID {
		t.Errorf("Wrong story returned: got %s, want %s", story.Self, storyID)
	}

	// Test non-existent story
	_, err = pkg.GetStory("nonexistent_story")
	if err == nil {
		t.Error("Expected error for non-existent story, got nil")
	}

	t.Logf("Successfully retrieved story: %s", story.Self)
}
