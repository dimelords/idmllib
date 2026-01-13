package xmlutil

import (
	"os"
	"strings"
	"testing"
)

// TestCompareXMLWithDetails tests the enhanced detailed comparison.
func TestCompareXMLWithDetails(t *testing.T) {
	tests := []struct {
		name          string
		xml1          string
		xml2          string
		expectedDiffs int
		checkDiffs    func(t *testing.T, diffs []XMLDifference)
	}{
		{
			name:          "identical XML",
			xml1:          `<root><child>value</child></root>`,
			xml2:          `<root><child>value</child></root>`,
			expectedDiffs: 0,
		},
		{
			name:          "different whitespace - should be equivalent",
			xml1:          `<root><child>value</child></root>`,
			xml2:          `<root>  <child>value</child>  </root>`,
			expectedDiffs: 0,
		},
		{
			name:          "different tag",
			xml1:          `<root><child1>value</child1></root>`,
			xml2:          `<root><child2>value</child2></root>`,
			expectedDiffs: 1,
			checkDiffs: func(t *testing.T, diffs []XMLDifference) {
				if diffs[0].Type != "tag" {
					t.Errorf("expected type 'tag', got %q", diffs[0].Type)
				}
			},
		},
		{
			name:          "different attribute value",
			xml1:          `<root attr="value1"/>`,
			xml2:          `<root attr="value2"/>`,
			expectedDiffs: 1,
			checkDiffs: func(t *testing.T, diffs []XMLDifference) {
				if diffs[0].Type != "attribute" {
					t.Errorf("expected type 'attribute', got %q", diffs[0].Type)
				}
				if diffs[0].Expected != "value1" {
					t.Errorf("expected Expected='value1', got %q", diffs[0].Expected)
				}
				if diffs[0].Got != "value2" {
					t.Errorf("expected Got='value2', got %q", diffs[0].Got)
				}
			},
		},
		{
			name:          "missing attribute",
			xml1:          `<root attr1="val1" attr2="val2"/>`,
			xml2:          `<root attr1="val1"/>`,
			expectedDiffs: 1,
			checkDiffs: func(t *testing.T, diffs []XMLDifference) {
				if diffs[0].Type != "attribute" {
					t.Errorf("expected type 'attribute', got %q", diffs[0].Type)
				}
			},
		},
		{
			name:          "extra attribute",
			xml1:          `<root attr1="val1"/>`,
			xml2:          `<root attr1="val1" attr2="val2"/>`,
			expectedDiffs: 1,
			checkDiffs: func(t *testing.T, diffs []XMLDifference) {
				if diffs[0].Type != "attribute" {
					t.Errorf("expected type 'attribute', got %q", diffs[0].Type)
				}
			},
		},
		{
			name:          "different text content",
			xml1:          `<root>text1</root>`,
			xml2:          `<root>text2</root>`,
			expectedDiffs: 1,
			checkDiffs: func(t *testing.T, diffs []XMLDifference) {
				if diffs[0].Type != "text" {
					t.Errorf("expected type 'text', got %q", diffs[0].Type)
				}
			},
		},
		{
			name:          "different child count",
			xml1:          `<root><child1/><child2/></root>`,
			xml2:          `<root><child1/></root>`,
			expectedDiffs: 1,
			checkDiffs: func(t *testing.T, diffs []XMLDifference) {
				if diffs[0].Type != "structure" {
					t.Errorf("expected type 'structure', got %q", diffs[0].Type)
				}
			},
		},
		{
			name:          "multiple differences",
			xml1:          `<root attr="val1"><child>text1</child></root>`,
			xml2:          `<root attr="val2"><child>text2</child></root>`,
			expectedDiffs: 2, // attribute + text
			checkDiffs: func(t *testing.T, diffs []XMLDifference) {
				types := make(map[string]bool)
				for _, d := range diffs {
					types[d.Type] = true
				}
				if !types["attribute"] || !types["text"] {
					t.Error("expected both attribute and text differences")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diffs, err := CompareXMLWithDetails([]byte(tt.xml1), []byte(tt.xml2), nil)
			if err != nil {
				t.Fatalf("CompareXMLWithDetails failed: %v", err)
			}

			if len(diffs) != tt.expectedDiffs {
				t.Errorf("expected %d differences, got %d", tt.expectedDiffs, len(diffs))
				for i, d := range diffs {
					t.Logf("  Diff %d: %s - %s", i+1, d.Path, d.Description)
				}
			}

			if tt.checkDiffs != nil && len(diffs) > 0 {
				tt.checkDiffs(t, diffs)
			}
		})
	}
}

// TestCompareXMLWithSorting tests element sorting for order-independent comparison.
// NOTE: Current sorting implementation sorts by tag name only, which doesn't help
// when elements have the same tag but different attributes/content.
// For true IDML resource/style comparison, we'd need semantic matching by ID.
func TestCompareXMLWithSorting(t *testing.T) {
	tests := []struct {
		name          string
		xml1          string
		xml2          string
		sortElements  []string
		expectedDiffs int
		note          string
	}{
		{
			name:          "different order without sorting",
			xml1:          `<root><styles><style id="1"/><style id="2"/></styles></root>`,
			xml2:          `<root><styles><style id="2"/><style id="1"/></styles></root>`,
			sortElements:  []string{}, // No sorting
			expectedDiffs: 2,          // Will see attribute differences
			note:          "Without sorting, order differences are detected",
		},
		{
			name:          "different order with sorting - LIMITATION",
			xml1:          `<root><styles><style id="1"/><style id="2"/></styles></root>`,
			xml2:          `<root><styles><style id="2"/><style id="1"/></styles></root>`,
			sortElements:  []string{"styles"}, // Sort styles children
			expectedDiffs: 2,                  // Still sees diffs - sorts by tag, not attributes
			note:          "Current implementation sorts by tag name, doesn't handle same-tag-different-attrs",
		},
		{
			name:          "different tags with sorting",
			xml1:          `<root><container><beta/><alpha/></container></root>`,
			xml2:          `<root><container><alpha/><beta/></container></root>`,
			sortElements:  []string{"container"},
			expectedDiffs: 0,
			note:          "Sorting works when elements have different tag names",
		},
		{
			name:          "different order but different content",
			xml1:          `<root><styles><style id="1" value="a"/><style id="2" value="b"/></styles></root>`,
			xml2:          `<root><styles><style id="2" value="x"/><style id="1" value="y"/></styles></root>`,
			sortElements:  []string{"styles"},
			expectedDiffs: 4, // After sorting by tag, attributes still differ
			note:          "Content differences are still detected after sorting",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &CompareOptions{
				SortElements:     tt.sortElements,
				MaxDifferences:   100,
				IgnoreWhitespace: true,
			}

			diffs, err := CompareXMLWithDetails([]byte(tt.xml1), []byte(tt.xml2), opts)
			if err != nil {
				t.Fatalf("CompareXMLWithDetails failed: %v", err)
			}

			if len(diffs) != tt.expectedDiffs {
				t.Errorf("expected %d differences, got %d", tt.expectedDiffs, len(diffs))
				t.Logf("Note: %s", tt.note)
				for i, d := range diffs {
					t.Logf("  Diff %d: %s [%s] - %s", i+1, d.Path, d.Type, d.Description)
					t.Logf("    Expected: %s", d.Expected)
					t.Logf("    Got: %s", d.Got)
				}
			}
		})
	}
}

// TestMaxDifferences tests the max differences limit.
func TestMaxDifferences(t *testing.T) {
	// XML with many differences
	xml1 := `<root>
		<item attr="1">text1</item>
		<item attr="2">text2</item>
		<item attr="3">text3</item>
		<item attr="4">text4</item>
		<item attr="5">text5</item>
	</root>`

	xml2 := `<root>
		<item attr="99">different1</item>
		<item attr="99">different2</item>
		<item attr="99">different3</item>
		<item attr="99">different4</item>
		<item attr="99">different5</item>
	</root>`

	opts := &CompareOptions{
		MaxDifferences:   3,
		IgnoreWhitespace: true,
	}

	diffs, err := CompareXMLWithDetails([]byte(xml1), []byte(xml2), opts)
	if err != nil {
		t.Fatalf("CompareXMLWithDetails failed: %v", err)
	}

	// Should stop at limit
	if len(diffs) > 3 {
		t.Errorf("expected at most 3 differences (limit), got %d", len(diffs))
	}

	if len(diffs) < 3 {
		t.Errorf("expected to collect up to limit (3), but only got %d", len(diffs))
	}

	t.Logf("Collected %d differences (limit was 3):", len(diffs))
	for i, d := range diffs {
		t.Logf("  Diff %d: %s - %s", i+1, d.Path, d.Description)
	}
}

// TestFormatDifferences tests the formatting function.
func TestFormatDifferences(t *testing.T) {
	tests := []struct {
		name  string
		diffs []XMLDifference
		check func(t *testing.T, output string)
	}{
		{
			name:  "no differences",
			diffs: []XMLDifference{},
			check: func(t *testing.T, output string) {
				if output != "No differences found" {
					t.Errorf("unexpected output: %s", output)
				}
			},
		},
		{
			name: "single difference",
			diffs: []XMLDifference{{
				Path:        "/root/child",
				Type:        "text",
				Description: "text content differs",
				Expected:    "original",
				Got:         "modified",
			}},
			check: func(t *testing.T, output string) {
				if !containsAll(output, []string{"1 difference", "/root/child", "text", "original", "modified"}) {
					t.Errorf("output missing expected content:\n%s", output)
				}
			},
		},
		{
			name: "multiple differences",
			diffs: []XMLDifference{
				{
					Path:        "/root/child1",
					Type:        "attribute",
					Description: "attribute differs",
					Expected:    "val1",
					Got:         "val2",
				},
				{
					Path:        "/root/child2",
					Type:        "text",
					Description: "text differs",
					Expected:    "text1",
					Got:         "text2",
				},
			},
			check: func(t *testing.T, output string) {
				if !containsAll(output, []string{"2 difference", "/root/child1", "/root/child2"}) {
					t.Errorf("output missing expected content:\n%s", output)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := FormatDifferences(tt.diffs)
			t.Logf("Output:\n%s", output)
			tt.check(t, output)
		})
	}
}

// Helper function to check if string contains all substrings.
func containsAll(s string, substrs []string) bool {
	for _, substr := range substrs {
		if !strings.Contains(s, substr) {
			return false
		}
	}
	return true
}

// TestBackwardCompatibility ensures existing tests still work.
func TestBackwardCompatibility(t *testing.T) {
	tests := []struct {
		name      string
		xml1      string
		xml2      string
		shouldErr bool
	}{
		{
			name:      "identical",
			xml1:      `<root><child>value</child></root>`,
			xml2:      `<root><child>value</child></root>`,
			shouldErr: false,
		},
		{
			name:      "different",
			xml1:      `<root><child>value1</child></root>`,
			xml2:      `<root><child>value2</child></root>`,
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test old CompareXML function still works
			err := CompareXML([]byte(tt.xml1), []byte(tt.xml2))
			if tt.shouldErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Test old CompareXMLWithEtree function still works
			err = CompareXMLWithEtree([]byte(tt.xml1), []byte(tt.xml2))
			if tt.shouldErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestDefaultCompareOptions tests the default options.
func TestDefaultCompareOptions(t *testing.T) {
	opts := DefaultCompareOptions()

	if opts == nil {
		t.Fatal("DefaultCompareOptions returned nil")
	}

	if len(opts.SortElements) == 0 {
		t.Error("expected some default sort elements")
	}

	if opts.MaxDifferences <= 0 {
		t.Error("expected positive max differences")
	}

	if !opts.IgnoreWhitespace {
		t.Error("expected IgnoreWhitespace to be true by default")
	}

	t.Logf("Default options:")
	t.Logf("  SortElements: %v", opts.SortElements)
	t.Logf("  MaxDifferences: %d", opts.MaxDifferences)
	t.Logf("  IgnoreWhitespace: %v", opts.IgnoreWhitespace)
}

// TestRealIDMLFilesWithDetails tests the enhanced comparison with real IDML files.
func TestRealIDMLFilesWithDetails(t *testing.T) {
	testFiles := []string{
		"../../testdata/designmap.xml",
		"../../testdata/designmap_minimal.xml",
		"../../testdata/Spread_u210.xml",
		"../../testdata/story_u1d8.xml",
	}

	for _, file := range testFiles {
		t.Run(file, func(t *testing.T) {
			// Read file
			data, err := readTestFile(file)
			if err != nil {
				t.Skipf("file not found or unreadable: %s", file)
				return
			}

			// Test 1: Self-comparison with new API (should have zero diffs)
			opts := DefaultCompareOptions()
			diffs, err := CompareXMLWithDetails(data, data, opts)
			if err != nil {
				t.Errorf("CompareXMLWithDetails failed: %v", err)
			}

			if len(diffs) > 0 {
				t.Errorf("self-comparison found %d differences:", len(diffs))
				for _, d := range diffs {
					t.Logf("  %s [%s]: %s", d.Path, d.Type, d.Description)
				}
			}

			// Test 2: Normalize and compare
			normalized, err := NormalizeXML(data)
			if err != nil {
				t.Errorf("NormalizeXML failed: %v", err)
				return
			}

			diffs, err = CompareXMLWithDetails(data, normalized, opts)
			if err != nil {
				t.Errorf("comparison with normalized failed: %v", err)
			}

			// Log any formatting differences (expected, but should be minor)
			if len(diffs) > 0 {
				t.Logf("Normalized XML has %d formatting differences (expected):", len(diffs))
				for i, d := range diffs {
					if i >= 5 {
						t.Logf("  ... and %d more", len(diffs)-5)
						break
					}
					t.Logf("  %s [%s]: %s", d.Path, d.Type, d.Description)
				}
			} else {
				t.Logf("✅ Normalized XML is structurally equivalent")
			}
		})
	}
}

// Helper to read test files safely.
func readTestFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, err
		}
		return nil, err
	}
	return data, nil
}
