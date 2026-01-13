package xmlutil

import (
	"fmt"
	"sort"
	"strings"

	"github.com/beevik/etree"
	"github.com/google/go-cmp/cmp"
)

// XMLDifference represents a single difference found during XML comparison.
type XMLDifference struct {
	Path        string // XPath-like path to the element (e.g., "/root/Story[0]/ParagraphStyleRange[2]")
	Type        string // "tag", "namespace", "attribute", "text", "structure"
	Description string // Human-readable description of the difference
	Expected    string // Expected value (from original)
	Got         string // Actual value (from generated)
}

// CompareOptions controls how XML comparison is performed.
type CompareOptions struct {
	// SortElements lists element tag names that should be sorted before comparison.
	// Useful for IDML Resources, Styles where order doesn't matter.
	SortElements []string

	// MaxDifferences limits how many differences to collect (0 = unlimited).
	// Prevents collecting thousands of diffs in severely broken XML.
	MaxDifferences int

	// IgnoreWhitespace controls whether text whitespace differences are ignored.
	// Default is true (whitespace is trimmed before comparison).
	IgnoreWhitespace bool
}

// DefaultCompareOptions returns sensible defaults for IDML comparison.
func DefaultCompareOptions() *CompareOptions {
	return &CompareOptions{
		SortElements: []string{
			// IDML elements where order doesn't matter
			"FontFamily",
			"Color",
			"ParagraphStyle",
			"CharacterStyle",
			"ObjectStyle",
		},
		MaxDifferences:   100,
		IgnoreWhitespace: true,
	}
}

// CompareXMLWithDetails compares two XML byte slices and returns all differences found.
// This is more informative than CompareXML which stops at the first error.
//
// Returns a slice of XMLDifference structs describing what differs, or an error if parsing fails.
// An empty slice means the XML is structurally equivalent.
func CompareXMLWithDetails(original, generated []byte, opts *CompareOptions) ([]XMLDifference, error) {
	if opts == nil {
		opts = DefaultCompareOptions()
	}

	// Parse original document
	origDoc := etree.NewDocument()
	if err := origDoc.ReadFromBytes(original); err != nil {
		return nil, fmt.Errorf("failed to parse original XML: %w", err)
	}

	// Parse generated document
	genDoc := etree.NewDocument()
	if err := genDoc.ReadFromBytes(generated); err != nil {
		return nil, fmt.Errorf("failed to parse generated XML: %w", err)
	}

	// Compare root elements
	origRoot := origDoc.Root()
	genRoot := genDoc.Root()

	if origRoot == nil && genRoot == nil {
		return nil, nil // Both empty, no differences
	}

	if origRoot == nil {
		return []XMLDifference{{
			Path:        "/",
			Type:        "structure",
			Description: "original has no root element, generated does",
			Expected:    "(none)",
			Got:         genRoot.Tag,
		}}, nil
	}

	if genRoot == nil {
		return []XMLDifference{{
			Path:        "/",
			Type:        "structure",
			Description: "original has root element, generated does not",
			Expected:    origRoot.Tag,
			Got:         "(none)",
		}}, nil
	}

	// Collect differences
	diffs := []XMLDifference{}
	compareElementsDetailed(origRoot, genRoot, "root", &diffs, opts)

	return diffs, nil
}

// compareElementsDetailed recursively compares two etree elements and collects all differences.
func compareElementsDetailed(orig, gen *etree.Element, path string, diffs *[]XMLDifference, opts *CompareOptions) {
	// Check if we've hit the max differences limit
	if opts.MaxDifferences > 0 && len(*diffs) >= opts.MaxDifferences {
		return
	}

	// Compare tag names
	if orig.Tag != gen.Tag {
		*diffs = append(*diffs, XMLDifference{
			Path:        path,
			Type:        "tag",
			Description: "tag name mismatch",
			Expected:    orig.Tag,
			Got:         gen.Tag,
		})
		return // Can't meaningfully continue comparing different elements
	}

	// Compare namespaces
	if orig.Space != gen.Space {
		*diffs = append(*diffs, XMLDifference{
			Path:        path,
			Type:        "namespace",
			Description: "namespace mismatch",
			Expected:    orig.Space,
			Got:         gen.Space,
		})
	}

	// Compare attributes (order-independent)
	compareAttributes(orig, gen, path, diffs, opts)

	// Compare text content
	compareText(orig, gen, path, diffs, opts)

	// Compare child elements
	compareChildren(orig, gen, path, diffs, opts)
}

// compareAttributes compares element attributes (order-independent).
func compareAttributes(orig, gen *etree.Element, path string, diffs *[]XMLDifference, opts *CompareOptions) {
	// Check limit before processing
	if opts.MaxDifferences > 0 && len(*diffs) >= opts.MaxDifferences {
		return
	}

	origAttrs := make(map[string]string)
	for _, attr := range orig.Attr {
		key := attr.Space + ":" + attr.Key
		origAttrs[key] = attr.Value
	}

	genAttrs := make(map[string]string)
	for _, attr := range gen.Attr {
		key := attr.Space + ":" + attr.Key
		genAttrs[key] = attr.Value
	}

	// Find missing attributes in generated
	for key, origVal := range origAttrs {
		if opts.MaxDifferences > 0 && len(*diffs) >= opts.MaxDifferences {
			return
		}
		if genVal, exists := genAttrs[key]; !exists {
			*diffs = append(*diffs, XMLDifference{
				Path:        path,
				Type:        "attribute",
				Description: fmt.Sprintf("attribute %q missing in generated", key),
				Expected:    origVal,
				Got:         "(missing)",
			})
		} else if genVal != origVal {
			*diffs = append(*diffs, XMLDifference{
				Path:        path,
				Type:        "attribute",
				Description: fmt.Sprintf("attribute %q value differs", key),
				Expected:    origVal,
				Got:         genVal,
			})
		}
	}

	// Find extra attributes in generated
	for key, genVal := range genAttrs {
		if opts.MaxDifferences > 0 && len(*diffs) >= opts.MaxDifferences {
			return
		}
		if _, exists := origAttrs[key]; !exists {
			*diffs = append(*diffs, XMLDifference{
				Path:        path,
				Type:        "attribute",
				Description: fmt.Sprintf("attribute %q exists in generated but not in original", key),
				Expected:    "(none)",
				Got:         genVal,
			})
		}
	}
}

// compareText compares text content of elements.
func compareText(orig, gen *etree.Element, path string, diffs *[]XMLDifference, opts *CompareOptions) {
	// Check limit before processing
	if opts.MaxDifferences > 0 && len(*diffs) >= opts.MaxDifferences {
		return
	}

	origText := orig.Text()
	genText := gen.Text()

	// Apply whitespace trimming if enabled
	if opts.IgnoreWhitespace {
		origText = strings.TrimSpace(origText)
		genText = strings.TrimSpace(genText)
	}

	if origText != genText {
		*diffs = append(*diffs, XMLDifference{
			Path:        path,
			Type:        "text",
			Description: "text content differs",
			Expected:    truncate(origText, 100),
			Got:         truncate(genText, 100),
		})
	}
}

// compareChildren compares child elements, with optional sorting.
func compareChildren(orig, gen *etree.Element, path string, diffs *[]XMLDifference, opts *CompareOptions) {
	// Check limit before processing
	if opts.MaxDifferences > 0 && len(*diffs) >= opts.MaxDifferences {
		return
	}

	origChildren := orig.ChildElements()
	genChildren := gen.ChildElements()

	// Check if this element type should be sorted before comparison
	shouldSort := contains(opts.SortElements, orig.Tag)

	if shouldSort {
		// Sort both children by tag name for order-independent comparison
		sortElementsByTag(origChildren)
		sortElementsByTag(genChildren)
	}

	if len(origChildren) != len(genChildren) {
		*diffs = append(*diffs, XMLDifference{
			Path:        path,
			Type:        "structure",
			Description: "child element count mismatch",
			Expected:    fmt.Sprintf("%d children", len(origChildren)),
			Got:         fmt.Sprintf("%d children", len(genChildren)),
		})

		// Check limit after adding diff
		if opts.MaxDifferences > 0 && len(*diffs) >= opts.MaxDifferences {
			return
		}
	}

	// Compare common children
	minLen := len(origChildren)
	if len(genChildren) < minLen {
		minLen = len(genChildren)
	}

	for i := 0; i < minLen; i++ {
		childPath := fmt.Sprintf("%s/%s[%d]", path, origChildren[i].Tag, i)
		compareElementsDetailed(origChildren[i], genChildren[i], childPath, diffs, opts)

		// Check limit after each child
		if opts.MaxDifferences > 0 && len(*diffs) >= opts.MaxDifferences {
			return
		}
	}
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func sortElementsByTag(elements []*etree.Element) {
	sort.Slice(elements, func(i, j int) bool {
		return elements[i].Tag < elements[j].Tag
	})
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// Backward-compatible functions that maintain existing API

// CompareXML compares two XML byte slices for structural equivalence.
// It uses etree library to parse and compare XML documents.
// This correctly handles namespaces, attributes, and ignores insignificant whitespace.
//
// Returns nil if XML is structurally equivalent, error with details if different.
// This is the simple API - use CompareXMLWithDetails for more information.
func CompareXML(original, generated []byte) error {
	return CompareXMLWithEtree(original, generated)
}

// CompareXMLWithEtree compares two XML byte slices using etree library.
// This provides more detailed control over XML comparison and can handle
// more complex namespace scenarios.
//
// Returns nil if XML trees are equivalent, error with details if different.
// This stops at the first error - use CompareXMLWithDetails to collect all differences.
func CompareXMLWithEtree(original, generated []byte) error {
	// Parse original document
	origDoc := etree.NewDocument()
	if err := origDoc.ReadFromBytes(original); err != nil {
		return fmt.Errorf("failed to parse original XML with etree: %w", err)
	}

	// Parse generated document
	genDoc := etree.NewDocument()
	if err := genDoc.ReadFromBytes(generated); err != nil {
		return fmt.Errorf("failed to parse generated XML with etree: %w", err)
	}

	// Compare root elements
	origRoot := origDoc.Root()
	genRoot := genDoc.Root()

	if origRoot == nil && genRoot == nil {
		return nil // Both empty
	}
	if origRoot == nil || genRoot == nil {
		return fmt.Errorf("one document has no root element")
	}

	// Compare recursively
	return compareElements(origRoot, genRoot, "root")
}

// compareElements recursively compares two etree elements.
// This is the old implementation that returns on first error.
func compareElements(orig, gen *etree.Element, path string) error {
	// Compare tag names
	if orig.Tag != gen.Tag {
		return fmt.Errorf("%s: tag mismatch: %q vs %q", path, orig.Tag, gen.Tag)
	}

	// Compare namespaces
	if orig.Space != gen.Space {
		return fmt.Errorf("%s: namespace mismatch: %q vs %q", path, orig.Space, gen.Space)
	}

	// Compare attributes (order-independent)
	origAttrs := make(map[string]string)
	for _, attr := range orig.Attr {
		key := attr.Space + ":" + attr.Key
		origAttrs[key] = attr.Value
	}

	genAttrs := make(map[string]string)
	for _, attr := range gen.Attr {
		key := attr.Space + ":" + attr.Key
		genAttrs[key] = attr.Value
	}

	if !cmp.Equal(origAttrs, genAttrs) {
		return fmt.Errorf("%s: attributes differ:\n%s", path, cmp.Diff(origAttrs, genAttrs))
	}

	// Compare text content (ignoring insignificant whitespace)
	origText := strings.TrimSpace(orig.Text())
	genText := strings.TrimSpace(gen.Text())
	if origText != genText {
		return fmt.Errorf("%s: text content differs:\norig: %q\ngen:  %q", path, origText, genText)
	}

	// Compare child elements
	origChildren := orig.ChildElements()
	genChildren := gen.ChildElements()

	if len(origChildren) != len(genChildren) {
		return fmt.Errorf("%s: child count mismatch: %d vs %d", path, len(origChildren), len(genChildren))
	}

	for i := range origChildren {
		childPath := fmt.Sprintf("%s/%s[%d]", path, origChildren[i].Tag, i)
		if err := compareElements(origChildren[i], genChildren[i], childPath); err != nil {
			return err
		}
	}

	return nil
}

// NormalizeXML parses and re-formats XML with consistent indentation.
// Useful for comparing XML files where formatting differs.
//
// Returns the normalized XML bytes and any parsing error.
func NormalizeXML(data []byte) ([]byte, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(data); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	// Format with indentation
	doc.Indent(2)

	// Write to bytes
	return doc.WriteToBytes()
}

// ParseToMap parses XML into an etree Document structure.
// This is useful for Phase 1 where we don't have typed structs yet.
// Returns the parsed document or an error.
//
// Note: In Phase 1, we work with generic XML structures (etree.Document)
// rather than strongly-typed structs. This allows us to handle any XML
// without defining structs first.
func ParseToMap(data []byte) (*etree.Document, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(data); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}
	return doc, nil
}

// FormatDifferences returns a human-readable summary of XML differences.
func FormatDifferences(diffs []XMLDifference) string {
	if len(diffs) == 0 {
		return "No differences found"
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Found %d difference(s):\n\n", len(diffs)))

	for i, diff := range diffs {
		b.WriteString(fmt.Sprintf("%d. %s [%s]\n", i+1, diff.Path, diff.Type))
		b.WriteString(fmt.Sprintf("   %s\n", diff.Description))
		b.WriteString(fmt.Sprintf("   Expected: %s\n", diff.Expected))
		b.WriteString(fmt.Sprintf("   Got:      %s\n", diff.Got))
		if i < len(diffs)-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}
