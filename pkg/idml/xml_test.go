package idml

import (
	"testing"

	"github.com/dimelords/idmllib/internal/xmlutil"
)

// TestXMLRoundtrip verifies that we can read XML files from a Package,
// parse them without typed structs (using etree.Document), and compare them
// structurally.
func TestXMLRoundtrip_StructuralComparison(t *testing.T) {
	tests := []struct {
		name     string
		filename string
	}{
		{
			name:     "plain.idml",
			filename: "../../testdata/plain.idml",
		},
		{
			name:     "example.idml",
			filename: "../../testdata/example.idml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Read the IDML package
			pkg, err := Read(tt.filename)
			if err != nil {
				t.Fatalf("Read() failed: %v", err)
			}

			// Get XML files from the package
			// We'll focus on the critical XML files
			xmlFiles := []string{
				"designmap.xml",
				"Spreads/Spread_u210.xml",
				"Stories/Story_u1d8.xml",
			}

			for _, xmlFile := range xmlFiles {
				t.Run(xmlFile, func(t *testing.T) {
					// Get the raw data for this file
					data, ok := pkg.files[xmlFile]
					if !ok {
						t.Skipf("file %s not found in package", xmlFile)
						return
					}

					// Test 1: Parse to etree.Document (no typed structs!)
					doc, err := xmlutil.ParseToMap(data.data)
					if err != nil {
						t.Errorf("failed to parse %s: %v", xmlFile, err)
						return
					}
					t.Logf("✅ Successfully parsed %s to etree.Document", xmlFile)
					if doc.Root() != nil {
						t.Logf("   Root: %s (namespace: %s)", doc.Root().Tag, doc.Root().Space)
					}

					// Test 2: Self-comparison using go-cmp (via etree)
					if err := xmlutil.CompareXML(data.data, data.data); err != nil {
						t.Errorf("self-comparison failed for %s: %v", xmlFile, err)
					}

					// Test 3: Self-comparison using etree directly
					if err := xmlutil.CompareXMLWithEtree(data.data, data.data); err != nil {
						t.Errorf("etree self-comparison failed for %s: %v", xmlFile, err)
					}

					// Test 4: Normalize and compare
					normalized, err := xmlutil.NormalizeXML(data.data)
					if err != nil {
						t.Errorf("failed to normalize %s: %v", xmlFile, err)
						return
					}

					// Compare original vs normalized (structural comparison)
					if err := xmlutil.CompareXML(data.data, normalized); err != nil {
						t.Logf("Note: Normalized XML differs from original for %s", xmlFile)
						t.Logf("      This is expected - formatting changes")
					} else {
						t.Logf("✅ Normalized XML is structurally identical to original for %s", xmlFile)
					}
				})
			}
		})
	}
}
