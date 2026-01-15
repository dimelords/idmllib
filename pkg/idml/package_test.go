package idml

import (
	"testing"

	"github.com/dimelords/idmllib/v2/pkg/xmp"
)

// TestXMP_ReadIDMLExtractsMetadata tests that reading an IDML file extracts XMP metadata.
// This test verifies Requirements 2.1, 2.4.
func TestXMP_ReadIDMLExtractsMetadata(t *testing.T) {
	// Load example.idml which should contain XMP metadata
	pkg := loadExampleIDML(t)

	// Access XMP metadata
	xmpMeta := pkg.XMP()
	if xmpMeta == nil {
		t.Fatal("Expected non-nil XMP metadata")
	}

	// Verify XMP is not empty
	if xmpMeta.IsEmpty() {
		t.Error("Expected XMP metadata to be non-empty for example.idml")
	}

	// Verify XMP contains expected fields
	createDate, err := xmpMeta.GetField("xmp:CreateDate")
	if err != nil {
		t.Errorf("Expected xmp:CreateDate field to exist: %v", err)
	}
	if createDate == "" {
		t.Error("Expected xmp:CreateDate to have a value")
	}
}

// TestXMP_ReturnsNonNilMetadata tests that XMP() returns non-nil Metadata.
// This test verifies Requirements 3.1, 3.5.
func TestXMP_ReturnsNonNilMetadata(t *testing.T) {
	pkg := loadExampleIDML(t)

	xmpMeta := pkg.XMP()
	if xmpMeta == nil {
		t.Fatal("Expected XMP() to return non-nil Metadata")
	}
}

// TestXMP_EmptyMetadataForPackageWithoutXMP tests that XMP() returns empty Metadata
// when the package doesn't contain XMP.
// This test verifies Requirements 2.4, 3.5.
func TestXMP_EmptyMetadataForPackageWithoutXMP(t *testing.T) {
	// Create a minimal package without XMP
	pkg := New()
	
	// Add minimal required files for a valid IDML package
	designmap := []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Document xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging" DOMVersion="20.4">
</Document>`)
	
	pkg.files["designmap.xml"] = &fileEntry{data: designmap}
	pkg.fileOrder = append(pkg.fileOrder, "designmap.xml")

	xmpMeta := pkg.XMP()
	if xmpMeta == nil {
		t.Fatal("Expected XMP() to return non-nil Metadata even for empty XMP")
	}

	if !xmpMeta.IsEmpty() {
		t.Error("Expected XMP metadata to be empty for package without XMP")
	}
}

// TestSetXMP_UpdatesPackageMetadata tests that SetXMP() updates the package XMP metadata.
// This test verifies Requirements 3.3.
func TestSetXMP_UpdatesPackageMetadata(t *testing.T) {
	pkg := loadExampleIDML(t)

	// Get current XMP
	xmpMeta := pkg.XMP()
	if xmpMeta.IsEmpty() {
		t.Skip("Test requires IDML file with XMP metadata")
	}

	// Modify XMP - update a field
	originalCreatorTool, err := xmpMeta.GetField("xmp:CreatorTool")
	if err != nil {
		t.Fatalf("Failed to get xmp:CreatorTool: %v", err)
	}

	newCreatorTool := "Test Tool v1.0"
	err = xmpMeta.SetField("xmp:CreatorTool", newCreatorTool)
	if err != nil {
		t.Fatalf("Failed to set xmp:CreatorTool: %v", err)
	}

	// Update package with modified XMP
	pkg.SetXMP(xmpMeta)

	// Verify the change was applied
	updatedXMP := pkg.XMP()
	updatedCreatorTool, err := updatedXMP.GetField("xmp:CreatorTool")
	if err != nil {
		t.Fatalf("Failed to get updated xmp:CreatorTool: %v", err)
	}

	if updatedCreatorTool != newCreatorTool {
		t.Errorf("Expected xmp:CreatorTool to be %q, got %q", newCreatorTool, updatedCreatorTool)
	}

	if updatedCreatorTool == originalCreatorTool {
		t.Error("Expected xmp:CreatorTool to be different from original")
	}
}

// TestXMP_ModifyAndPersist tests modifying XMP metadata and verifying it persists.
// This test verifies Requirements 3.1, 3.3, 7.1.
func TestXMP_ModifyAndPersist(t *testing.T) {
	pkg := loadExampleIDML(t)

	// Get XMP and modify it
	xmpMeta := pkg.XMP()
	if xmpMeta.IsEmpty() {
		t.Skip("Test requires IDML file with XMP metadata")
	}

	// Update timestamps
	err := xmpMeta.UpdateTimestamps()
	if err != nil {
		t.Fatalf("Failed to update timestamps: %v", err)
	}

	// Remove thumbnails
	err = xmpMeta.RemoveThumbnails()
	if err != nil && err.Error() != "no XMP metadata" {
		// It's okay if there are no thumbnails to remove
		t.Logf("RemoveThumbnails returned: %v", err)
	}

	// Set a custom field
	err = xmpMeta.SetField("xmp:CreatorTool", "idmllib Test Suite")
	if err != nil {
		t.Fatalf("Failed to set xmp:CreatorTool: %v", err)
	}

	// Update package
	pkg.SetXMP(xmpMeta)

	// Write to temporary file
	outputPath := writeTestIDML(t, pkg, "xmp_modified.idml")

	// Read back
	pkg2, err := Read(outputPath)
	if err != nil {
		t.Fatalf("Failed to read modified IDML: %v", err)
	}

	// Verify XMP was persisted
	xmpMeta2 := pkg2.XMP()
	if xmpMeta2.IsEmpty() {
		t.Fatal("Expected XMP metadata to be persisted")
	}

	// Verify the custom field
	creatorTool, err := xmpMeta2.GetField("xmp:CreatorTool")
	if err != nil {
		t.Fatalf("Failed to get xmp:CreatorTool from persisted XMP: %v", err)
	}

	if creatorTool != "idmllib Test Suite" {
		t.Errorf("Expected xmp:CreatorTool to be %q, got %q", "idmllib Test Suite", creatorTool)
	}

	// Verify thumbnails were removed (should not contain xmp:Thumbnails)
	xmpString := xmpMeta2.String()
	if len(xmpString) > 0 && containsSubstring(xmpString, "<xmp:Thumbnails>") {
		t.Error("Expected thumbnails to be removed from persisted XMP")
	}
}

// TestXMP_ChainedOperations tests performing multiple XMP operations in sequence.
// This test verifies Requirements 10.1, 10.2, 10.3.
func TestXMP_ChainedOperations(t *testing.T) {
	pkg := loadExampleIDML(t)

	xmpMeta := pkg.XMP()
	if xmpMeta.IsEmpty() {
		t.Skip("Test requires IDML file with XMP metadata")
	}

	// Perform chained operations
	err := xmpMeta.UpdateTimestamps()
	if err != nil {
		t.Fatalf("Failed to update timestamps: %v", err)
	}

	err = xmpMeta.RemoveThumbnails()
	if err != nil && err.Error() != "no XMP metadata" {
		t.Logf("RemoveThumbnails returned: %v", err)
	}

	err = xmpMeta.SetField("xmp:CreatorTool", "Chained Operations Test")
	if err != nil {
		t.Fatalf("Failed to set field: %v", err)
	}

	// Update package
	pkg.SetXMP(xmpMeta)

	// Verify all changes were applied
	finalXMP := pkg.XMP()
	
	creatorTool, err := finalXMP.GetField("xmp:CreatorTool")
	if err != nil {
		t.Fatalf("Failed to get xmp:CreatorTool: %v", err)
	}
	if creatorTool != "Chained Operations Test" {
		t.Errorf("Expected xmp:CreatorTool to be %q, got %q", "Chained Operations Test", creatorTool)
	}

	// Verify timestamps were updated (they should be recent)
	modifyDate, err := finalXMP.GetField("xmp:ModifyDate")
	if err != nil {
		t.Fatalf("Failed to get xmp:ModifyDate: %v", err)
	}
	if modifyDate == "" {
		t.Error("Expected xmp:ModifyDate to have a value")
	}
}

// containsSubstring is a helper function to check if a string contains a substring.
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

// findSubstring performs a simple substring search.
func findSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestXMP_Integration_RoundTrip tests the complete XMP round-trip workflow.
// This is a comprehensive integration test that verifies:
// - Reading IDML extracts XMP
// - Modifying XMP works correctly
// - Writing IDML includes modified XMP
// - Reading the written file preserves XMP changes
// This test verifies Requirements 2.1, 3.1, 3.3, 7.1, 7.3.
func TestXMP_Integration_RoundTrip(t *testing.T) {
	// Load original IDML
	pkg := loadExampleIDML(t)

	// Get and verify XMP exists
	xmpMeta := pkg.XMP()
	if xmpMeta.IsEmpty() {
		t.Skip("Test requires IDML file with XMP metadata")
	}

	// Store original CreateDate (should be preserved)
	originalCreateDate, err := xmpMeta.GetField("xmp:CreateDate")
	if err != nil {
		t.Fatalf("Failed to get original xmp:CreateDate: %v", err)
	}

	// Modify XMP with multiple operations
	testCreatorTool := "idmllib Round-Trip Test v1.0"
	
	err = xmpMeta.UpdateTimestamps()
	if err != nil {
		t.Fatalf("Failed to update timestamps: %v", err)
	}

	err = xmpMeta.SetField("xmp:CreatorTool", testCreatorTool)
	if err != nil {
		t.Fatalf("Failed to set xmp:CreatorTool: %v", err)
	}

	// Update package
	pkg.SetXMP(xmpMeta)

	// Write to file
	outputPath := writeTestIDML(t, pkg, "xmp_roundtrip.idml")

	// Read back
	pkg2, err := Read(outputPath)
	if err != nil {
		t.Fatalf("Failed to read round-trip IDML: %v", err)
	}

	// Verify XMP was persisted correctly
	xmpMeta2 := pkg2.XMP()
	if xmpMeta2.IsEmpty() {
		t.Fatal("Expected XMP metadata to be persisted in round-trip")
	}

	// Verify CreateDate was preserved
	createDate, err := xmpMeta2.GetField("xmp:CreateDate")
	if err != nil {
		t.Fatalf("Failed to get xmp:CreateDate from round-trip: %v", err)
	}
	if createDate != originalCreateDate {
		t.Errorf("Expected xmp:CreateDate to be preserved as %q, got %q", originalCreateDate, createDate)
	}

	// Verify CreatorTool was updated
	creatorTool, err := xmpMeta2.GetField("xmp:CreatorTool")
	if err != nil {
		t.Fatalf("Failed to get xmp:CreatorTool from round-trip: %v", err)
	}
	if creatorTool != testCreatorTool {
		t.Errorf("Expected xmp:CreatorTool to be %q, got %q", testCreatorTool, creatorTool)
	}

	// Verify ModifyDate and MetadataDate exist and are not empty
	modifyDate, err := xmpMeta2.GetField("xmp:ModifyDate")
	if err != nil {
		t.Fatalf("Failed to get xmp:ModifyDate from round-trip: %v", err)
	}
	if modifyDate == "" {
		t.Error("Expected xmp:ModifyDate to have a value")
	}

	metadataDate, err := xmpMeta2.GetField("xmp:MetadataDate")
	if err != nil {
		t.Fatalf("Failed to get xmp:MetadataDate from round-trip: %v", err)
	}
	if metadataDate == "" {
		t.Error("Expected xmp:MetadataDate to have a value")
	}
}

// TestXMP_EmptyPackage tests XMP operations on a newly created empty package.
// This test verifies Requirements 2.4, 3.5.
func TestXMP_EmptyPackage(t *testing.T) {
	pkg := New()

	// XMP should return non-nil but empty metadata
	xmpMeta := pkg.XMP()
	if xmpMeta == nil {
		t.Fatal("Expected XMP() to return non-nil Metadata for empty package")
	}

	if !xmpMeta.IsEmpty() {
		t.Error("Expected XMP metadata to be empty for new package")
	}

	// Operations on empty XMP should return errors
	err := xmpMeta.UpdateTimestamps()
	if err == nil {
		t.Error("Expected error when updating timestamps on empty XMP")
	}

	err = xmpMeta.RemoveThumbnails()
	if err == nil {
		t.Error("Expected error when removing thumbnails from empty XMP")
	}

	_, err = xmpMeta.GetField("xmp:CreateDate")
	if err == nil {
		t.Error("Expected error when getting field from empty XMP")
	}

	err = xmpMeta.SetField("xmp:CreatorTool", "Test")
	if err == nil {
		t.Error("Expected error when setting field on empty XMP")
	}
}

// TestXMP_ParseFromString tests creating XMP metadata from a string.
// This verifies the xmp.Parse function works correctly with IDML integration.
func TestXMP_ParseFromString(t *testing.T) {
	xmpString := `<?xpacket begin="" id="W5M0MpCehiHzreSzNTczkc9d"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
  <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
    <rdf:Description xmlns:xmp="http://ns.adobe.com/xap/1.0/">
      <xmp:CreateDate>2026-01-13T11:50:31+01:00</xmp:CreateDate>
      <xmp:MetadataDate>2026-01-13T11:50:31+01:00</xmp:MetadataDate>
      <xmp:ModifyDate>2026-01-13T11:50:31+01:00</xmp:ModifyDate>
      <xmp:CreatorTool>Test Tool</xmp:CreatorTool>
    </rdf:Description>
  </rdf:RDF>
</x:xmpmeta>
<?xpacket end="r"?>`

	xmpMeta := xmp.Parse(xmpString)
	if xmpMeta == nil {
		t.Fatal("Expected non-nil XMP metadata from Parse")
	}

	if xmpMeta.IsEmpty() {
		t.Error("Expected XMP metadata to be non-empty")
	}

	// Verify we can read fields
	creatorTool, err := xmpMeta.GetField("xmp:CreatorTool")
	if err != nil {
		t.Fatalf("Failed to get xmp:CreatorTool: %v", err)
	}
	if creatorTool != "Test Tool" {
		t.Errorf("Expected xmp:CreatorTool to be %q, got %q", "Test Tool", creatorTool)
	}

	// Verify String() returns the original
	if xmpMeta.String() != xmpString {
		t.Error("Expected String() to return original XMP string")
	}
}
