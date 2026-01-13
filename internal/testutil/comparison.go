// Package testutil provides testing utilities for idmlbuild.
// This package contains helpers for golden file testing, comparison utilities,
// and test data management.
package testutil

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/beevik/etree"
	"github.com/google/go-cmp/cmp"
)

// CompareZIPContents compares the contents of two ZIP files.
// It returns a detailed report of any differences found.
func CompareZIPContents(t *testing.T, path1, path2 string) (*ZIPComparison, error) {
	t.Helper()

	zip1, err := zip.OpenReader(path1)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", path1, err)
	}
	defer zip1.Close()

	zip2, err := zip.OpenReader(path2)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", path2, err)
	}
	defer zip2.Close()

	comp := &ZIPComparison{
		Path1: path1,
		Path2: path2,
	}

	// Compare file counts
	comp.FileCount1 = len(zip1.File)
	comp.FileCount2 = len(zip2.File)

	if comp.FileCount1 != comp.FileCount2 {
		comp.Differences = append(comp.Differences, fmt.Sprintf(
			"File count mismatch: %d vs %d", comp.FileCount1, comp.FileCount2))
	}

	// Compare each file
	fileMap1 := make(map[string]*zip.File)
	for _, f := range zip1.File {
		fileMap1[f.Name] = f
	}

	fileMap2 := make(map[string]*zip.File)
	for _, f := range zip2.File {
		fileMap2[f.Name] = f
	}

	// Check for files in zip1 but not zip2
	for name := range fileMap1 {
		if _, exists := fileMap2[name]; !exists {
			comp.Differences = append(comp.Differences,
				fmt.Sprintf("File %q exists in first but not second", name))
		}
	}

	// Check for files in zip2 but not zip1
	for name := range fileMap2 {
		if _, exists := fileMap1[name]; !exists {
			comp.Differences = append(comp.Differences,
				fmt.Sprintf("File %q exists in second but not first", name))
		}
	}

	// Compare common files
	for name, file1 := range fileMap1 {
		file2, exists := fileMap2[name]
		if !exists {
			continue
		}

		rc1, err := file1.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open %s in first ZIP: %w", name, err)
		}
		data1, _ := io.ReadAll(rc1)
		rc1.Close()

		rc2, err := file2.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open %s in second ZIP: %w", name, err)
		}
		data2, _ := io.ReadAll(rc2)
		rc2.Close()

		if !bytes.Equal(data1, data2) {
			comp.Differences = append(comp.Differences,
				fmt.Sprintf("File %q content differs (%d vs %d bytes)",
					name, len(data1), len(data2)))
		} else {
			comp.IdenticalFiles++
		}
	}

	comp.IsIdentical = len(comp.Differences) == 0

	return comp, nil
}

// ZIPComparison holds the results of comparing two ZIP files.
type ZIPComparison struct {
	Path1          string
	Path2          string
	FileCount1     int
	FileCount2     int
	IdenticalFiles int
	Differences    []string
	IsIdentical    bool
}

// Report returns a human-readable report of the comparison.
func (c *ZIPComparison) Report() string {
	var buf strings.Builder
	buf.WriteString("ZIP Comparison Report\n")
	buf.WriteString(strings.Repeat("=", 50))
	buf.WriteString("\n")
	buf.WriteString(fmt.Sprintf("File 1: %s (%d files)\n", c.Path1, c.FileCount1))
	buf.WriteString(fmt.Sprintf("File 2: %s (%d files)\n", c.Path2, c.FileCount2))
	buf.WriteString(fmt.Sprintf("Identical files: %d\n", c.IdenticalFiles))

	if c.IsIdentical {
		buf.WriteString("\n✅ Files are identical!\n")
	} else {
		buf.WriteString(fmt.Sprintf("\n❌ Found %d difference(s):\n", len(c.Differences)))
		for i, diff := range c.Differences {
			buf.WriteString(fmt.Sprintf("  %d. %s\n", i+1, diff))
		}
	}

	return buf.String()
}

// CompareXMLStructurally compares two XML documents structurally using etree.
// This ignores whitespace differences and focuses on element structure.
func CompareXMLStructurally(t *testing.T, xml1, xml2 []byte) error {
	t.Helper()

	doc1 := etree.NewDocument()
	if err := doc1.ReadFromBytes(xml1); err != nil {
		return fmt.Errorf("failed to parse first XML: %w", err)
	}

	doc2 := etree.NewDocument()
	if err := doc2.ReadFromBytes(xml2); err != nil {
		return fmt.Errorf("failed to parse second XML: %w", err)
	}

	// Normalize both documents
	normalizeDocument(doc1)
	normalizeDocument(doc2)

	// Compare
	str1, _ := doc1.WriteToString()
	str2, _ := doc2.WriteToString()

	if str1 != str2 {
		return fmt.Errorf("XML structures differ:\n%s",
			cmp.Diff(str1, str2))
	}

	return nil
}

// normalizeDocument removes insignificant whitespace from an etree document.
func normalizeDocument(doc *etree.Document) {
	normalizeElement(doc.Root())
}

// normalizeElement recursively normalizes an element.
func normalizeElement(elem *etree.Element) {
	if elem == nil {
		return
	}

	// Remove whitespace-only text nodes
	for i := len(elem.Child) - 1; i >= 0; i-- {
		if charData, ok := elem.Child[i].(*etree.CharData); ok {
			if strings.TrimSpace(charData.Data) == "" {
				elem.RemoveChildAt(i)
			}
		}
	}

	// Recursively normalize children
	for _, child := range elem.ChildElements() {
		normalizeElement(child)
	}
}
