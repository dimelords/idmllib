package idml

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/dimelords/idmllib/pkg/common"
)

// TestMetadataFileParsing_ParsesIndividualFiles tests parsing individual metadata files
func TestMetadataFileParsing_ParsesIndividualFiles(t *testing.T) {
	tests := []struct {
		name     string
		filename string
	}{
		{
			name:     "container.xml",
			filename: "META-INF/container.xml",
		},
		{
			name:     "metadata.xml",
			filename: "META-INF/metadata.xml",
		},
		{
			name:     "Tags.xml",
			filename: "XML/Tags.xml",
		},
		{
			name:     "BackingStory.xml",
			filename: "XML/BackingStory.xml",
		},
	}

	pkg, err := Read("../../testdata/plain.idml")
	if err != nil {
		t.Fatalf("Read() failed: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mf, err := pkg.MetadataFile(tt.filename)
			if err != nil {
				t.Fatalf("MetadataFile() failed: %v", err)
			}

			if mf.Filename != tt.filename {
				t.Errorf("Filename = %q, want %q", mf.Filename, tt.filename)
			}

			if len(mf.RawContent) == 0 {
				t.Error("RawContent is empty")
			}

			// Verify it starts with XML declaration
			if !bytes.HasPrefix(mf.RawContent, []byte("<?xml")) {
				t.Error("RawContent doesn't start with XML declaration")
			}
		})
	}
}

// TestMetadataFilesAll_RetrievesAllFiles tests retrieving all metadata files
func TestMetadataFilesAll_RetrievesAllFiles(t *testing.T) {
	pkg, err := Read("../../testdata/plain.idml")
	if err != nil {
		t.Fatalf("Read() failed: %v", err)
	}

	metadata, err := pkg.MetadataFiles()
	if err != nil {
		t.Fatalf("MetadataFiles() failed: %v", err)
	}

	// Verify we got the expected files
	expected := map[string]bool{
		"META-INF/container.xml": false,
		"META-INF/metadata.xml":  false,
		"XML/Tags.xml":           false,
		"XML/BackingStory.xml":   false,
	}

	for path := range metadata {
		if _, exists := expected[path]; exists {
			expected[path] = true
		}
	}

	// Check all expected files were found
	for path, found := range expected {
		if !found {
			t.Errorf("Expected metadata file %q not found", path)
		}
	}

	// Verify each file has content
	for path, mf := range metadata {
		if len(mf.RawContent) == 0 {
			t.Errorf("File %q has empty content", path)
		}
	}
}

// TestMetadataRoundtrip_PreservesMetadata tests that metadata files survive roundtrip
func TestMetadataRoundtrip_PreservesMetadata(t *testing.T) {
	tests := []struct {
		name     string
		idmlFile string
	}{
		{
			name:     "plain.idml",
			idmlFile: "../../testdata/plain.idml",
		},
		{
			name:     "example.idml",
			idmlFile: "../../testdata/example.idml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Read original
			pkg1, err := Read(tt.idmlFile)
			if err != nil {
				t.Fatalf("Read() failed: %v", err)
			}

			// Parse all metadata files to trigger caching
			metadata1, err := pkg1.MetadataFiles()
			if err != nil {
				t.Fatalf("MetadataFiles() failed: %v", err)
			}

			if len(metadata1) == 0 {
				t.Skip("No metadata files in test file")
			}

			t.Logf("Found %d metadata files", len(metadata1))

			// Write to temp file
			tmpDir := t.TempDir()
			outputPath := filepath.Join(tmpDir, "output.idml")
			if err := Write(pkg1, outputPath); err != nil {
				t.Fatalf("Write() failed: %v", err)
			}

			// Read back
			pkg2, err := Read(outputPath)
			if err != nil {
				t.Fatalf("Read() output failed: %v", err)
			}

			// Parse metadata files again
			metadata2, err := pkg2.MetadataFiles()
			if err != nil {
				t.Fatalf("MetadataFiles() from output failed: %v", err)
			}

			// Compare counts
			if len(metadata1) != len(metadata2) {
				t.Errorf("Metadata file count mismatch: %d vs %d",
					len(metadata1), len(metadata2))
			}

			// Compare each file
			for path, mf1 := range metadata1 {
				mf2, exists := metadata2[path]
				if !exists {
					t.Errorf("Metadata file %q missing in output", path)
					continue
				}

				// Compare content
				if !bytes.Equal(mf1.RawContent, mf2.RawContent) {
					t.Errorf("Metadata file %q differs after roundtrip", path)
					t.Logf("  Original: %d bytes", len(mf1.RawContent))
					t.Logf("  Output:   %d bytes", len(mf2.RawContent))
				}
			}

			t.Logf("✅ Roundtrip successful: %d metadata files preserved", len(metadata1))
		})
	}
}

// TestMetadataNotFound_HandlesErrors tests error handling for missing files
func TestMetadataNotFound_HandlesErrors(t *testing.T) {
	pkg, err := Read("../../testdata/plain.idml")
	if err != nil {
		t.Fatalf("Read() failed: %v", err)
	}

	_, err = pkg.MetadataFile("META-INF/nonexistent.xml")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}

	// Verify it's the right error type
	if !common.IsNotFound(err) {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}
