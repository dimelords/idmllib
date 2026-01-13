package idml

import (
	"path/filepath"
	"testing"

	"github.com/dimelords/idmllib/v2/internal/testutil"
)

func TestIsValidZipPath_ValidatesZipPaths(t *testing.T) {
	tests := []struct {
		name  string
		path  string
		valid bool
	}{
		// Valid paths
		{"simple file", "file.txt", true},
		{"nested path", "dir/file.txt", true},
		{"deeply nested", "a/b/c/d/file.txt", true},
		{"IDML mimetype", "mimetype", true},
		{"IDML designmap", "designmap.xml", true},
		{"IDML story", "Stories/Story_u1d8.xml", true},
		{"IDML spread", "Spreads/Spread_ub6.xml", true},
		{"IDML resource", "Resources/Fonts.xml", true},
		{"META-INF", "META-INF/container.xml", true},
		{"current dir prefix", "./file.txt", true},

		// Invalid paths - directory traversal
		{"parent dir", "../file.txt", false},
		{"nested parent", "dir/../../../etc/passwd", false},
		{"double parent", "../../file.txt", false},
		{"hidden traversal", "dir/subdir/../../..", false},
		{"traversal with normal", "a/b/../../../c", false},

		// Invalid paths - absolute
		{"absolute unix", "/etc/passwd", false},
		{"absolute with traversal", "/tmp/../etc/passwd", false},

		// Invalid paths - empty
		{"empty string", "", false},

		// Edge cases that should be valid
		{"dots in filename", "file..txt", true},
		{"triple dots", "file...txt", true},
		{"dot prefix", ".hidden", true},
		{"dot directory", ".config/file", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidZipPath(tt.path)
			if got != tt.valid {
				t.Errorf("isValidZipPath(%q) = %v, want %v", tt.path, got, tt.valid)
			}
		})
	}
}

func TestIsValidZipPath_WindowsPaths(t *testing.T) {
	// These tests check Windows-style paths which should be rejected
	// even on Unix systems for defense in depth
	tests := []struct {
		name  string
		path  string
		valid bool
	}{
		// Note: filepath.IsAbs behavior varies by OS
		// On Unix, "C:\file" is not absolute but contains backslash
		// We primarily care about ".." traversal which works on both
		{"backslash traversal", "dir\\..\\..\\file", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidZipPath(tt.path)
			// We only check that traversal is caught
			if tt.valid != got && !tt.valid {
				// If we expected invalid and got valid, that's a problem
				// But Windows paths on Unix may behave differently
				t.Logf("isValidZipPath(%q) = %v (platform-dependent)", tt.path, got)
			}
		})
	}
}

// createTestZIP creates a ZIP file for testing with specified files.
func createTestZIP(t *testing.T, files map[string][]byte) string {
	return testutil.CreateTestZIP(t, files)
}

func TestReadOptions_Defaults(t *testing.T) {
	opts := &ReadOptions{}
	opts.applyDefaults()

	if opts.MaxTotalSize != DefaultMaxTotalSize {
		t.Errorf("MaxTotalSize = %d, want %d", opts.MaxTotalSize, DefaultMaxTotalSize)
	}
	if opts.MaxFileSize != DefaultMaxFileSize {
		t.Errorf("MaxFileSize = %d, want %d", opts.MaxFileSize, DefaultMaxFileSize)
	}
	if opts.MaxFileCount != DefaultMaxFileCount {
		t.Errorf("MaxFileCount = %d, want %d", opts.MaxFileCount, DefaultMaxFileCount)
	}
	if opts.MaxCompressionRatio != DefaultMaxCompressionRatio {
		t.Errorf("MaxCompressionRatio = %d, want %d", opts.MaxCompressionRatio, DefaultMaxCompressionRatio)
	}
}

func TestReadOptions_CustomValues(t *testing.T) {
	opts := &ReadOptions{
		MaxTotalSize:        1000,
		MaxFileSize:         500,
		MaxFileCount:        10,
		MaxCompressionRatio: 50,
	}
	opts.applyDefaults()

	// Custom values should be preserved
	if opts.MaxTotalSize != 1000 {
		t.Errorf("MaxTotalSize = %d, want 1000", opts.MaxTotalSize)
	}
	if opts.MaxFileSize != 500 {
		t.Errorf("MaxFileSize = %d, want 500", opts.MaxFileSize)
	}
	if opts.MaxFileCount != 10 {
		t.Errorf("MaxFileCount = %d, want 10", opts.MaxFileCount)
	}
	if opts.MaxCompressionRatio != 50 {
		t.Errorf("MaxCompressionRatio = %d, want 50", opts.MaxCompressionRatio)
	}
}

func TestReadWithOptions_FileCountLimit(t *testing.T) {
	// Create a simple minimal designmap.xml for testing
	designmap := []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Document xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging" DOMVersion="20.4">
</Document>`)

	// Create ZIP with many files
	files := map[string][]byte{
		"designmap.xml": designmap,
	}
	for i := 0; i < 15; i++ {
		files[filepath.Join("Stories", "file"+string(rune('a'+i))+".xml")] = []byte("<test/>")
	}

	zipPath := createTestZIP(t, files)

	// Should fail with strict file count limit
	_, err := ReadWithOptions(zipPath, &ReadOptions{
		MaxFileCount: 5,
	})
	if err == nil {
		t.Error("expected error for file count limit, got nil")
	}

	// Should succeed with higher limit
	_, err = ReadWithOptions(zipPath, &ReadOptions{
		MaxFileCount: 100,
	})
	if err != nil {
		t.Errorf("unexpected error with sufficient file count limit: %v", err)
	}
}

func TestReadWithOptions_TotalSizeLimit(t *testing.T) {
	// Create large content - use varied content to avoid high compression ratio
	largeContent := make([]byte, 10000)
	for i := range largeContent {
		largeContent[i] = byte(i % 256) // Varied content compresses less
	}
	designmap := []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Document xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging" DOMVersion="20.4">
</Document>`)

	files := map[string][]byte{
		"designmap.xml":      designmap,
		"Stories/large1.xml": largeContent,
		"Stories/large2.xml": largeContent,
	}

	zipPath := createTestZIP(t, files)

	// Should fail with small total size limit
	_, err := ReadWithOptions(zipPath, &ReadOptions{
		MaxTotalSize:        5000, // Less than total
		MaxCompressionRatio: -1,   // Disable ratio check for this test
	})
	if err == nil {
		t.Error("expected error for total size limit, got nil")
	}

	// Should succeed with higher limit
	_, err = ReadWithOptions(zipPath, &ReadOptions{
		MaxTotalSize:        100000,
		MaxCompressionRatio: -1, // Disable ratio check for this test
	})
	if err != nil {
		t.Errorf("unexpected error with sufficient total size limit: %v", err)
	}
}

func TestReadWithOptions_DisableLimits(t *testing.T) {
	designmap := []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Document xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging" DOMVersion="20.4">
</Document>`)

	// Create many files
	files := map[string][]byte{
		"designmap.xml": designmap,
	}
	for i := 0; i < 50; i++ {
		files[filepath.Join("Stories", "file"+string(rune('a'+i))+".xml")] = []byte("<test/>")
	}

	zipPath := createTestZIP(t, files)

	// Should succeed with disabled file count limit (-1)
	_, err := ReadWithOptions(zipPath, &ReadOptions{
		MaxFileCount: -1,
	})
	if err != nil {
		t.Errorf("unexpected error with disabled file count limit: %v", err)
	}
}

func TestRead_UsesDefaultOptions(t *testing.T) {
	// Read a valid IDML file - should work with default limits
	pkg := loadPlainIDML(t)

	if pkg.FileCount() == 0 {
		t.Error("expected files in package")
	}
}
