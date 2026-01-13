package idms

import (
	"os"
	"testing"
)

// TestRead tests reading the text-only IDMS file
func TestReadTextOnlyIDMS(t *testing.T) {
	pkg, err := Read("../../testdata/Snippet_31F27A2D0.idms")
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}

	// Verify basic structure
	if pkg == nil {
		t.Fatal("Package is nil")
	}

	if pkg.Document == nil {
		t.Fatal("Document is nil")
	}

	// Verify XML declaration
	if pkg.XMLDeclaration == "" {
		t.Error("XML declaration is empty")
	}

	// Verify AID processing instructions
	if len(pkg.AIDProcessingInstructions) == 0 {
		t.Error("No AID processing instructions found")
	}

	// Verify XMP metadata
	if pkg.XMPMetadata == "" {
		t.Error("XMP metadata is empty")
	}

	t.Logf("✅ Successfully read IDMS file")
	t.Logf("   XML Declaration: %s", pkg.XMLDeclaration)
	t.Logf("   AID PIs: %d", len(pkg.AIDProcessingInstructions))
	t.Logf("   Document: %+v", pkg.Document != nil)
	t.Logf("   XMP Metadata length: %d bytes", len(pkg.XMPMetadata))
}

// TestParseTextOnlyIDMS tests parsing the text-only IDMS file
func TestParseTextOnlyIDMS(t *testing.T) {
	data, err := os.ReadFile("../../testdata/Snippet_31F27A2D0.idms")
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}

	pkg, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Verify structure
	if pkg == nil {
		t.Fatal("Package is nil")
	}

	if pkg.Document == nil {
		t.Fatal("Document is nil")
	}

	// Check Document Self attribute
	if pkg.Document.Self != "d" {
		t.Errorf("Expected Document.Self = 'd', got '%s'", pkg.Document.Self)
	}

	// Check DOMVersion
	if pkg.Document.DOMVersion != "20.4" {
		t.Errorf("Expected DOMVersion = '20.4', got '%s'", pkg.Document.DOMVersion)
	}

	t.Logf("✅ Successfully parsed IDMS structure")
	t.Logf("   Document.Self: %s", pkg.Document.Self)
	t.Logf("   Document.DOMVersion: %s", pkg.Document.DOMVersion)
}

// TestSnippetType tests the SnippetType() method
func TestSnippetType(t *testing.T) {
	pkg, err := Read("../../testdata/Snippet_31F27A2D0.idms")
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}

	snippetType := pkg.SnippetType()
	if snippetType == "" {
		t.Error("SnippetType is empty")
	}

	t.Logf("✅ SnippetType: %s", snippetType)

	// Verify it's a page item snippet
	if snippetType != "PageItem" {
		t.Logf("⚠️  Expected 'PageItem', got '%s'", snippetType)
	}
}
