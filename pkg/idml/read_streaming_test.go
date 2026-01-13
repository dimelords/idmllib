package idml

import (
	"bytes"
	"os"
	"testing"
)

func TestReadBytes_BasicFunctionality(t *testing.T) {
	// Read a valid IDML file into memory
	data, err := os.ReadFile("../../testdata/plain.idml")
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	// Parse from bytes
	pkg, err := ReadBytes(data)
	if err != nil {
		t.Fatalf("ReadBytes failed: %v", err)
	}

	if pkg.FileCount() == 0 {
		t.Error("expected files in package")
	}

	// Verify designmap.xml was parsed
	doc, err := pkg.Document()
	if err != nil {
		t.Errorf("failed to get document: %v", err)
	}
	if doc == nil {
		t.Error("expected document to be parsed")
	}
}

func TestReadBytes_vs_Read(t *testing.T) {
	// Read the same file using both methods and compare results
	testFile := "../../testdata/plain.idml"

	// Method 1: Traditional Read
	pkg1, err := Read(testFile)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	// Method 2: ReadBytes
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	pkg2, err := ReadBytes(data)
	if err != nil {
		t.Fatalf("ReadBytes failed: %v", err)
	}

	// Compare file counts
	if pkg1.FileCount() != pkg2.FileCount() {
		t.Errorf("file count mismatch: Read=%d, ReadBytes=%d", pkg1.FileCount(), pkg2.FileCount())
	}

	// Compare file order
	files1 := pkg1.fileOrder
	files2 := pkg2.fileOrder
	if len(files1) != len(files2) {
		t.Errorf("file order length mismatch: %d vs %d", len(files1), len(files2))
	}
	for i := range files1 {
		if files1[i] != files2[i] {
			t.Errorf("file order mismatch at index %d: %q vs %q", i, files1[i], files2[i])
		}
	}

	// Compare document parsing
	doc1, _ := pkg1.Document()
	doc2, _ := pkg2.Document()
	if doc1.DOMVersion != doc2.DOMVersion {
		t.Errorf("DOMVersion mismatch: %q vs %q", doc1.DOMVersion, doc2.DOMVersion)
	}
}

func TestReadBytesWithOptions_FileCountLimit(t *testing.T) {
	designmap := []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Document xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging" DOMVersion="20.4">
</Document>`)

	// Create ZIP in memory with many files
	files := map[string][]byte{
		"designmap.xml": designmap,
	}
	for i := 0; i < 15; i++ {
		files["Stories/file"+string(rune('a'+i))+".xml"] = []byte("<test/>")
	}

	zipPath := createTestZIP(t, files)
	data, err := os.ReadFile(zipPath)
	if err != nil {
		t.Fatalf("failed to read test ZIP: %v", err)
	}

	// Should fail with strict file count limit
	_, err = ReadBytesWithOptions(data, &ReadOptions{
		MaxFileCount: 5,
	})
	if err == nil {
		t.Error("expected error for file count limit, got nil")
	}

	// Should succeed with higher limit
	_, err = ReadBytesWithOptions(data, &ReadOptions{
		MaxFileCount: 100,
	})
	if err != nil {
		t.Errorf("unexpected error with sufficient file count limit: %v", err)
	}
}

func TestReadBytesWithOptions_TotalSizeLimit(t *testing.T) {
	// Create large content
	largeContent := make([]byte, 10000)
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
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
	data, err := os.ReadFile(zipPath)
	if err != nil {
		t.Fatalf("failed to read test ZIP: %v", err)
	}

	// Should fail with small total size limit
	_, err = ReadBytesWithOptions(data, &ReadOptions{
		MaxTotalSize:        5000,
		MaxCompressionRatio: -1,
	})
	if err == nil {
		t.Error("expected error for total size limit, got nil")
	}

	// Should succeed with higher limit
	_, err = ReadBytesWithOptions(data, &ReadOptions{
		MaxTotalSize:        100000,
		MaxCompressionRatio: -1,
	})
	if err != nil {
		t.Errorf("unexpected error with sufficient total size limit: %v", err)
	}
}

func TestReadFrom_BasicFunctionality(t *testing.T) {
	// Open file and get size
	file, err := os.Open("../../testdata/plain.idml")
	if err != nil {
		t.Fatalf("failed to open test file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}

	// Parse from io.ReaderAt
	pkg, err := ReadFrom(file, stat.Size())
	if err != nil {
		t.Fatalf("ReadFrom failed: %v", err)
	}

	if pkg.FileCount() == 0 {
		t.Error("expected files in package")
	}

	// Verify designmap.xml was parsed
	doc, err := pkg.Document()
	if err != nil {
		t.Errorf("failed to get document: %v", err)
	}
	if doc == nil {
		t.Error("expected document to be parsed")
	}
}

func TestReadFrom_vs_Read(t *testing.T) {
	testFile := "../../testdata/plain.idml"

	// Method 1: Traditional Read
	pkg1, err := Read(testFile)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	// Method 2: ReadFrom
	file, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}

	pkg2, err := ReadFrom(file, stat.Size())
	if err != nil {
		t.Fatalf("ReadFrom failed: %v", err)
	}

	// Compare file counts
	if pkg1.FileCount() != pkg2.FileCount() {
		t.Errorf("file count mismatch: Read=%d, ReadFrom=%d", pkg1.FileCount(), pkg2.FileCount())
	}
}

func TestReadFrom_WithBytesReader(t *testing.T) {
	// Read file into memory
	data, err := os.ReadFile("../../testdata/plain.idml")
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	// Create bytes.Reader
	reader := bytes.NewReader(data)

	// Parse using ReadFrom
	pkg, err := ReadFrom(reader, int64(len(data)))
	if err != nil {
		t.Fatalf("ReadFrom with bytes.Reader failed: %v", err)
	}

	if pkg.FileCount() == 0 {
		t.Error("expected files in package")
	}
}

func TestReadFromWithOptions_Limits(t *testing.T) {
	designmap := []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Document xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging" DOMVersion="20.4">
</Document>`)

	files := map[string][]byte{
		"designmap.xml": designmap,
	}
	for i := 0; i < 15; i++ {
		files["Stories/file"+string(rune('a'+i))+".xml"] = []byte("<test/>")
	}

	zipPath := createTestZIP(t, files)
	data, err := os.ReadFile(zipPath)
	if err != nil {
		t.Fatalf("failed to read test ZIP: %v", err)
	}

	reader := bytes.NewReader(data)

	// Should fail with strict file count limit
	_, err = ReadFromWithOptions(reader, int64(len(data)), &ReadOptions{
		MaxFileCount: 5,
	})
	if err == nil {
		t.Error("expected error for file count limit, got nil")
	}

	// Create new reader (old one was consumed)
	reader = bytes.NewReader(data)

	// Should succeed with higher limit
	_, err = ReadFromWithOptions(reader, int64(len(data)), &ReadOptions{
		MaxFileCount: 100,
	})
	if err != nil {
		t.Errorf("unexpected error with sufficient file count limit: %v", err)
	}
}

func TestReadBytes_InvalidZIP(t *testing.T) {
	// Try to parse invalid ZIP data
	invalidData := []byte("not a valid ZIP file")

	_, err := ReadBytes(invalidData)
	if err == nil {
		t.Error("expected error for invalid ZIP data, got nil")
	}
}

func TestReadBytes_EmptyData(t *testing.T) {
	// Try to parse empty data
	_, err := ReadBytes([]byte{})
	if err == nil {
		t.Error("expected error for empty data, got nil")
	}
}

func TestReadBytes_NilOptions(t *testing.T) {
	// ReadBytes should handle nil options gracefully
	data, err := os.ReadFile("../../testdata/plain.idml")
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	// nil options should use defaults
	pkg, err := ReadBytesWithOptions(data, nil)
	if err != nil {
		t.Fatalf("ReadBytesWithOptions with nil options failed: %v", err)
	}

	if pkg.FileCount() == 0 {
		t.Error("expected files in package")
	}
}

func TestReadFrom_NilOptions(t *testing.T) {
	// ReadFrom should handle nil options gracefully
	data, err := os.ReadFile("../../testdata/plain.idml")
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	reader := bytes.NewReader(data)

	// nil options should use defaults
	pkg, err := ReadFromWithOptions(reader, int64(len(data)), nil)
	if err != nil {
		t.Fatalf("ReadFromWithOptions with nil options failed: %v", err)
	}

	if pkg.FileCount() == 0 {
		t.Error("expected files in package")
	}
}

func TestReadBytes_DisabledLimits(t *testing.T) {
	designmap := []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Document xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging" DOMVersion="20.4">
</Document>`)

	// Create many files
	files := map[string][]byte{
		"designmap.xml": designmap,
	}
	for i := 0; i < 50; i++ {
		files["Stories/file"+string(rune('a'+i))+".xml"] = []byte("<test/>")
	}

	zipPath := createTestZIP(t, files)
	data, err := os.ReadFile(zipPath)
	if err != nil {
		t.Fatalf("failed to read test ZIP: %v", err)
	}

	// Should succeed with disabled file count limit (-1)
	_, err = ReadBytesWithOptions(data, &ReadOptions{
		MaxFileCount: -1,
	})
	if err != nil {
		t.Errorf("unexpected error with disabled file count limit: %v", err)
	}
}

// Benchmark comparison
func BenchmarkRead(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := Read("../../testdata/plain.idml")
		if err != nil {
			b.Fatalf("Read failed: %v", err)
		}
	}
}

func BenchmarkReadBytes(b *testing.B) {
	// Pre-load file data
	data, err := os.ReadFile("../../testdata/plain.idml")
	if err != nil {
		b.Fatalf("failed to read test file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ReadBytes(data)
		if err != nil {
			b.Fatalf("ReadBytes failed: %v", err)
		}
	}
}

func BenchmarkReadFrom(b *testing.B) {
	// Pre-load file data
	data, err := os.ReadFile("../../testdata/plain.idml")
	if err != nil {
		b.Fatalf("failed to read test file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(data)
		_, err := ReadFrom(reader, int64(len(data)))
		if err != nil {
			b.Fatalf("ReadFrom failed: %v", err)
		}
	}
}
