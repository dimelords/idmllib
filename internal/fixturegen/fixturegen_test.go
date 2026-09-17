package fixturegen

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/dimelords/idmllib/v3/pkg/idml"
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
}
