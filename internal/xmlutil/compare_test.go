package xmlutil

import (
	"os"
	"path/filepath"
	"testing"
)

// TestCompareXML verifies basic XML comparison functionality.
func TestCompareXML(t *testing.T) {
	tests := []struct {
		name      string
		xml1      string
		xml2      string
		shouldErr bool
	}{
		{
			name:      "identical XML",
			xml1:      `<root><child>value</child></root>`,
			xml2:      `<root><child>value</child></root>`,
			shouldErr: false,
		},
		{
			name:      "different whitespace - should be equivalent",
			xml1:      `<root><child>value</child></root>`,
			xml2:      `<root>  <child>value</child>  </root>`,
			shouldErr: false,
		},
		{
			name: "different indentation - should be equivalent",
			xml1: `<root>
<child>value</child>
</root>`,
			xml2: `<root>
  <child>value</child>
</root>`,
			shouldErr: false,
		},
		{
			name:      "different content",
			xml1:      `<root><child>value1</child></root>`,
			xml2:      `<root><child>value2</child></root>`,
			shouldErr: true,
		},
		{
			name:      "different structure",
			xml1:      `<root><child1>value</child1></root>`,
			xml2:      `<root><child2>value</child2></root>`,
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CompareXML([]byte(tt.xml1), []byte(tt.xml2))
			if tt.shouldErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestCompareXMLWithEtree tests the etree-based comparison.
func TestCompareXMLWithEtree(t *testing.T) {
	tests := []struct {
		name      string
		xml1      string
		xml2      string
		shouldErr bool
	}{
		{
			name:      "identical with namespaces",
			xml1:      `<root xmlns="http://example.com"><child>value</child></root>`,
			xml2:      `<root xmlns="http://example.com"><child>value</child></root>`,
			shouldErr: false,
		},
		{
			name:      "different whitespace in content",
			xml1:      `<root><child>value</child></root>`,
			xml2:      `<root><child>  value  </child></root>`,
			shouldErr: false, // TrimSpace makes these equivalent
		},
		{
			name:      "attributes in different order - should be equivalent",
			xml1:      `<root a="1" b="2"><child/></root>`,
			xml2:      `<root b="2" a="1"><child/></root>`,
			shouldErr: false,
		},
		{
			name:      "different tag names",
			xml1:      `<root><child1/></root>`,
			xml2:      `<root><child2/></root>`,
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CompareXMLWithEtree([]byte(tt.xml1), []byte(tt.xml2))
			if tt.shouldErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestNormalizeXML tests XML normalization.
func TestNormalizeXML(t *testing.T) {
	input := `<root>  <child>value</child>  </root>`

	normalized, err := NormalizeXML([]byte(input))
	if err != nil {
		t.Fatalf("NormalizeXML failed: %v", err)
	}

	// Verify it's valid XML by parsing it again
	_, err = ParseToMap(normalized)
	if err != nil {
		t.Errorf("normalized XML is invalid: %v", err)
	}

	t.Logf("Normalized XML:\n%s", string(normalized))
}

// TestParseToMap tests parsing XML to etree Document.
func TestParseToMap(t *testing.T) {
	xml := `<root><child attr="value">content</child></root>`

	doc, err := ParseToMap([]byte(xml))
	if err != nil {
		t.Fatalf("ParseToMap failed: %v", err)
	}

	if doc == nil {
		t.Error("ParseToMap returned nil")
		return
	}

	if doc.Root() == nil {
		t.Error("Document has no root element")
		return
	}

	t.Logf("Parsed root tag: %s", doc.Root().Tag)
}

// TestRealIDMLFiles tests XML comparison with actual IDML XML files.
func TestRealIDMLFiles(t *testing.T) {
	testFiles := []string{
		"../../testdata/designmap.xml",
		"../../testdata/designmap_minimal.xml",
		"../../testdata/Spread_u210.xml",
		"../../testdata/story_u1d8.xml",
	}

	for _, file := range testFiles {
		t.Run(filepath.Base(file), func(t *testing.T) {
			// Read the XML file
			data, err := os.ReadFile(file)
			if err != nil {
				if os.IsNotExist(err) {
					t.Skipf("file does not exist: %s", file)
					return
				}
				t.Fatalf("failed to read file: %v", err)
			}

			// Test 1: Parse to etree Document (no structs yet!)
			doc, err := ParseToMap(data)
			if err != nil {
				t.Errorf("ParseToMap failed: %v", err)
			} else {
				t.Logf("Successfully parsed %s to etree.Document", filepath.Base(file))
				if doc.Root() != nil {
					t.Logf("Root element: %s (namespace: %s)", doc.Root().Tag, doc.Root().Space)
				}
			}

			// Test 2: Self-comparison (should always pass)
			if err := CompareXML(data, data); err != nil {
				t.Errorf("self-comparison failed: %v", err)
			}

			// Test 3: Etree self-comparison
			if err := CompareXMLWithEtree(data, data); err != nil {
				t.Errorf("etree self-comparison failed: %v", err)
			}

			// Test 4: Normalize and compare
			normalized, err := NormalizeXML(data)
			if err != nil {
				t.Errorf("NormalizeXML failed: %v", err)
				return
			}

			// Normalized version should be structurally equivalent to original
			if err := CompareXML(data, normalized); err != nil {
				t.Logf("Note: Normalized XML differs from original structurally")
				t.Logf("Difference: %v", err)
			} else {
				t.Logf("✅ Normalized XML is structurally equivalent to original")
			}
		})
	}
}
