package fixturegen

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/dimelords/idmllib/v3/pkg/idml"
	"github.com/dimelords/idmllib/v3/pkg/idms"
)

var update = flag.Bool("update", false, "rewrite the fixtures under testdata")

// TestBuildRoundtrips checks every fixture can be written and read back
// before any of them is allowed near testdata. A fixture that does not
// survive its own roundtrip would fail the whole suite in a way that
// looks like a library bug.
func TestBuildRoundtrips(t *testing.T) {
	for _, spec := range Specs() {
		t.Run(spec.Name, func(t *testing.T) {
			pkg, err := Build(spec)
			if err != nil {
				t.Fatalf("Build: %v", err)
			}
			path := filepath.Join(t.TempDir(), spec.Name)
			if err := idml.Write(pkg, path); err != nil {
				t.Fatalf("Write: %v", err)
			}

			back, err := idml.Read(path)
			if err != nil {
				t.Fatalf("Read: %v", err)
			}
			stories, err := back.Stories()
			if err != nil {
				t.Fatalf("Stories: %v", err)
			}
			if len(stories) != spec.Stories {
				t.Errorf("stories = %d, want %d", len(stories), spec.Stories)
			}
			spreads, err := back.Spreads()
			if err != nil {
				t.Fatalf("Spreads: %v", err)
			}
			if len(spreads) != spec.Spreads {
				t.Errorf("spreads = %d, want %d", len(spreads), spec.Spreads)
			}
		})
	}
}

// TestWriteFixtures regenerates testdata. It only writes with -update, so
// an ordinary test run never rewrites the inputs other tests read.
func TestWriteFixtures(t *testing.T) {
	if !*update {
		t.Skip("pass -update to rewrite the fixtures")
	}
	dir := filepath.Join("..", "..", "testdata")
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("testdata: %v", err)
	}
	for _, spec := range Specs() {
		pkg, err := Build(spec)
		if err != nil {
			t.Fatalf("Build %s: %v", spec.Name, err)
		}
		out := filepath.Join(dir, spec.Name)
		if err := idml.Write(pkg, out); err != nil {
			t.Fatalf("Write %s: %v", out, err)
		}
		t.Logf("wrote %s", out)
	}

	// Two fixtures are single XML files that used to be lifted out of the
	// captured document. They are written from the generated package
	// instead, so they cannot drift from it and cannot reacquire anything
	// that was in the original.
	pkg, err := Build(Specs()[0])
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if err := idml.Write(pkg, filepath.Join(t.TempDir(), "tmp.idml")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	// The graphics snippet, exported from the generated document. It used
	// to be a selection lifted out of the customer's page; nothing else in
	// testdata carries an image, so it could not simply be swapped for one
	// of the other snippets.
	sel := idml.NewSelection()
	sp, err := pkg.Spread("Spreads/Spread_u210.xml")
	if err != nil {
		t.Fatalf("Spread: %v", err)
	}
	for i := range sp.Rectangles {
		sel.AddRectangle(&sp.Rectangles[i])
	}
	for i := range sp.TextFrames {
		sel.AddTextFrame(&sp.TextFrames[i])
		break
	}
	snippet, err := idms.NewExporter(pkg).ExportSelection(sel)
	if err != nil {
		t.Fatalf("ExportSelection: %v", err)
	}
	snippetPath := filepath.Join(dir, "Snippet_31F27A387.idms")
	if err := idms.Write(snippet, snippetPath); err != nil {
		t.Fatalf("Write snippet: %v", err)
	}
	t.Logf("wrote %s", snippetPath)

	for _, name := range []string{"designmap.xml", "Spreads/Spread_u210.xml"} {
		data, err := pkg.FileData(name)
		if err != nil {
			t.Fatalf("FileData %s: %v", name, err)
		}
		out := filepath.Join(dir, filepath.Base(name))
		if err := os.WriteFile(out, data, 0o644); err != nil {
			t.Fatalf("write %s: %v", out, err)
		}
		t.Logf("wrote %s", out)
	}
}
