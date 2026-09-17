package idms

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dimelords/idmllib/v3/internal/xmlutil"
)

// TestFidelityReMarshalIsLossless parses every IDMS snippet in testdata and
// marshals it back, requiring the result to be structurally identical to the
// input: same elements in the same order with the same attributes.
func TestFidelityReMarshalIsLossless(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("..", "..", "testdata", "*.idms"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no IDMS fixtures found in testdata")
	}
	for _, path := range files {
		t.Run(filepath.Base(path), func(t *testing.T) {
			in, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			pkg, err := Parse(in)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			out, err := Marshal(pkg)
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			diffs, err := xmlutil.CompareXMLWithDetails(in, out, xmlutil.StrictCompareOptions())
			if err != nil {
				t.Fatalf("compare: %v", err)
			}
			for i, d := range diffs {
				if i == 10 {
					t.Errorf("... %d more differences", len(diffs)-10)
					break
				}
				t.Errorf("%s at %s: expected %q, got %q", d.Description, d.Path, d.Expected, d.Got)
			}
		})
	}
}
