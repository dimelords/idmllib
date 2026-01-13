package idms

import (
	"bytes"
	"testing"

	"github.com/dimelords/idmllib/internal/testutil"
)

// TestGoldenMarshal_TextOnly tests that marshaling produces consistent output
// using golden file comparison.
func TestGoldenMarshal_TextOnly(t *testing.T) {
	// Read the original IDMS
	pkg, err := Read("../../testdata/Snippet_31F27A2D0.idms")
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}

	// Marshal it
	data, err := Marshal(pkg)
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}

	// Validate no namespace pollution
	if bytes.Contains(data, []byte("xmlns:idPkg")) {
		t.Error("Marshaled output contains xmlns:idPkg namespace declaration")
	}
	if bytes.Contains(data, []byte("idPkg:")) {
		t.Error("Marshaled output contains idPkg: prefix")
	}

	// Golden file comparison
	g := testutil.NewGoldenFileInTestdata(t)
	g.Assert(t, "marshal_text_only", data)
}

// TestGoldenMarshal_WithGraphics tests marshaling of IDMS with graphics
func TestGoldenMarshal_WithGraphics(t *testing.T) {
	// Read the graphics IDMS
	pkg, err := Read("../../testdata/Snippet_31F27A387.idms")
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}

	// Marshal it
	data, err := Marshal(pkg)
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}

	// Validate no namespace pollution
	if bytes.Contains(data, []byte("xmlns:idPkg")) {
		t.Error("Marshaled output contains xmlns:idPkg namespace declaration")
	}
	if bytes.Contains(data, []byte("idPkg:")) {
		t.Error("Marshaled output contains idPkg: prefix")
	}

	// Golden file comparison
	g := testutil.NewGoldenFileInTestdata(t)
	g.Assert(t, "marshal_with_graphics", data)
}

// TestNamespaceExclusion specifically validates that IDMS marshaling
// excludes the idPkg namespace that IDML uses.
func TestNamespaceExclusion(t *testing.T) {
	pkg, err := Read("../../testdata/Snippet_31F27A2D0.idms")
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}

	data, err := Marshal(pkg)
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}

	// Validate no namespace pollution
	testCases := []struct {
		name    string
		pattern []byte
	}{
		{"xmlns:idPkg declaration", []byte("xmlns:idPkg")},
		{"idPkg: prefix on Graphic", []byte("idPkg:Graphic")},
		{"idPkg: prefix on Fonts", []byte("idPkg:Fonts")},
		{"idPkg: prefix on Styles", []byte("idPkg:Styles")},
		{"idPkg: prefix on Spread", []byte("idPkg:Spread")},
		{"idPkg: prefix on Story", []byte("idPkg:Story")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if bytes.Contains(data, tc.pattern) {
				t.Errorf("Marshaled output contains forbidden pattern: %s", string(tc.pattern))
			}
		})
	}
}
