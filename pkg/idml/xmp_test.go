package idml

import (
	"testing"
)

func TestXMPExtraction(t *testing.T) {
	// Read an IDML file that contains XMP metadata
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read IDML: %v", err)
	}

	// Verify XMP metadata was extracted
	if pkg.XMPMetadata == "" {
		t.Error("XMP metadata was not extracted")
	}

	t.Logf("XMP metadata extracted: %d bytes", len(pkg.XMPMetadata))
}

func TestXMPAccessor(t *testing.T) {
	// Read an IDML file
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read IDML: %v", err)
	}

	// Test XMP() accessor
	xmp := pkg.XMP()
	if xmp == nil {
		t.Fatal("XMP() returned nil")
	}

	if xmp.IsEmpty() {
		t.Error("XMP should not be empty")
	}

	// Try to get a field
	createDate, err := xmp.GetField("xmp:CreateDate")
	if err != nil {
		t.Errorf("Failed to get CreateDate: %v", err)
	} else {
		t.Logf("CreateDate: %s", createDate)
	}
}

func TestSetXMP(t *testing.T) {
	// Read an IDML file
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read IDML: %v", err)
	}

	// Get XMP and modify it
	xmp := pkg.XMP()
	err = xmp.SetField("xmp:CreatorTool", "Test Tool")
	if err != nil {
		t.Fatalf("Failed to set field: %v", err)
	}

	// Update package XMP
	pkg.SetXMP(xmp)

	// Verify the change was persisted
	xmp2 := pkg.XMP()
	creatorTool, err := xmp2.GetField("xmp:CreatorTool")
	if err != nil {
		t.Fatalf("Failed to get CreatorTool: %v", err)
	}

	if creatorTool != "Test Tool" {
		t.Errorf("Expected CreatorTool to be 'Test Tool', got '%s'", creatorTool)
	}
}

func TestXMPPersistenceInWrite(t *testing.T) {
	// Read an IDML file
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read IDML: %v", err)
	}

	// Get original XMP
	originalXMP := pkg.XMP()
	originalCreatorTool, err := originalXMP.GetField("xmp:CreatorTool")
	if err != nil {
		t.Fatalf("Failed to get original CreatorTool: %v", err)
	}

	// Modify XMP
	xmp := pkg.XMP()
	err = xmp.SetField("xmp:CreatorTool", "Modified Test Tool")
	if err != nil {
		t.Fatalf("Failed to set field: %v", err)
	}
	pkg.SetXMP(xmp)

	// Write to a temporary file
	tmpFile := t.TempDir() + "/test_xmp_write.idml"
	err = Write(pkg, tmpFile)
	if err != nil {
		t.Fatalf("Failed to write IDML: %v", err)
	}

	// Read the file back
	pkg2, err := Read(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read written IDML: %v", err)
	}

	// Verify XMP was persisted
	xmp2 := pkg2.XMP()
	creatorTool, err := xmp2.GetField("xmp:CreatorTool")
	if err != nil {
		t.Fatalf("Failed to get CreatorTool from written file: %v", err)
	}

	if creatorTool != "Modified Test Tool" {
		t.Errorf("Expected CreatorTool to be 'Modified Test Tool', got '%s'", creatorTool)
	}

	t.Logf("Original CreatorTool: %s", originalCreatorTool)
	t.Logf("Modified CreatorTool: %s", creatorTool)
}
