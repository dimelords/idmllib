package idml

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dimelords/idmllib/v2/internal/xmlutil"
)

// Fidelity tests guard the core promise of the library: parsing a file and
// writing it back must not change its meaning. They run over every IDML file in
// testdata and compare each XML file of the output against the input with
// order-sensitive structural comparison.

const largeFileThreshold = 50 << 20 // skip very large fixtures under -short

func fidelityFixtures(t *testing.T) []string {
	t.Helper()
	files, err := filepath.Glob(filepath.Join("..", "..", "testdata", "*.idml"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no IDML fixtures found in testdata")
	}
	return files
}

func readZipFiles(t *testing.T, path string) map[string][]byte {
	t.Helper()
	r, err := zip.OpenReader(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer r.Close()
	out := make(map[string][]byte, len(r.File))
	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("open %s in %s: %v", f.Name, path, err)
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			t.Fatalf("read %s in %s: %v", f.Name, path, err)
		}
		out[f.Name] = data
	}
	return out
}

func skipIfLarge(t *testing.T, path string) {
	t.Helper()
	if !testing.Short() {
		return
	}
	if fi, err := os.Stat(path); err == nil && fi.Size() > largeFileThreshold {
		t.Skipf("skipping %s (%d bytes) under -short", filepath.Base(path), fi.Size())
	}
}

// parseEverything forces every typed parser to run and returns the paths of
// the files that now have a cached, typed representation.
func parseEverything(t *testing.T, pkg *Package) []string {
	t.Helper()
	var paths []string
	if _, err := pkg.Document(); err != nil {
		t.Fatalf("Document(): %v", err)
	}
	paths = append(paths, PathDesignmap)
	stories, err := pkg.Stories()
	if err != nil {
		t.Fatalf("Stories(): %v", err)
	}
	for name := range stories {
		paths = append(paths, name)
	}
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Spreads(): %v", err)
	}
	for name := range spreads {
		paths = append(paths, name)
	}
	masters, err := pkg.MasterSpreads()
	if err != nil {
		t.Fatalf("MasterSpreads(): %v", err)
	}
	for name := range masters {
		paths = append(paths, name)
	}
	if _, err := pkg.Fonts(); err != nil {
		t.Fatalf("Fonts(): %v", err)
	}
	if _, err := pkg.Graphics(); err != nil {
		t.Fatalf("Graphics(): %v", err)
	}
	if _, err := pkg.Styles(); err != nil {
		t.Fatalf("Styles(): %v", err)
	}
	if _, err := pkg.Preferences(); err != nil {
		t.Fatalf("Preferences(): %v", err)
	}
	if _, err := pkg.Tags(); err != nil {
		t.Fatalf("Tags(): %v", err)
	}
	return append(paths, PathFonts, PathGraphic, PathStyles, PathPreferences, PathTags)
}

// TestFidelityReMarshalIsLossless parses every typed file, marks it modified
// so Write re-marshals it from the structs, and checks that the result is
// structurally identical to the input: same elements in the same order with
// the same attributes.
func TestFidelityReMarshalIsLossless(t *testing.T) {
	for _, path := range fidelityFixtures(t) {
		t.Run(filepath.Base(path), func(t *testing.T) {
			skipIfLarge(t, path)
			pkg, err := readFixture(path)
			if err != nil {
				t.Fatalf("Read: %v", err)
			}
			parsed := parseEverything(t, pkg)
			for _, p := range parsed {
				pkg.MarkModified(p)
			}
			out := filepath.Join(t.TempDir(), "out.idml")
			if err := Write(pkg, out); err != nil {
				t.Fatalf("Write: %v", err)
			}

			orig := readZipFiles(t, path)
			gen := readZipFiles(t, out)
			if len(orig) != len(gen) {
				t.Errorf("file count changed: %d -> %d", len(orig), len(gen))
			}
			for _, name := range parsed {
				in, ok := orig[name]
				if !ok {
					continue // e.g. a package without Fonts.xml
				}
				outData, ok := gen[name]
				if !ok {
					t.Errorf("%s missing from output", name)
					continue
				}
				if bytes.Equal(in, outData) {
					t.Errorf("%s was marked modified but written back byte-identical; re-marshal did not run", name)
					continue
				}
				diffs, err := xmlutil.CompareXMLWithDetails(in, outData, xmlutil.StrictCompareOptions())
				if err != nil {
					t.Errorf("%s: compare failed: %v", name, err)
					continue
				}
				for i, d := range diffs {
					if i == 5 {
						t.Errorf("%s: ... %d more differences", name, len(diffs)-5)
						break
					}
					t.Errorf("%s: %s at %s: expected %q, got %q", name, d.Description, d.Path, d.Expected, d.Got)
				}
			}
		})
	}
}

// TestFidelityReadingIsNotDestructive checks that merely reading a package
// through every getter and writing it back leaves every file byte-identical.
func TestFidelityReadingIsNotDestructive(t *testing.T) {
	for _, path := range fidelityFixtures(t) {
		t.Run(filepath.Base(path), func(t *testing.T) {
			skipIfLarge(t, path)
			pkg, err := readFixture(path)
			if err != nil {
				t.Fatalf("Read: %v", err)
			}
			parseEverything(t, pkg)
			if _, err := pkg.MetadataFiles(); err != nil {
				t.Fatalf("MetadataFiles(): %v", err)
			}
			out := filepath.Join(t.TempDir(), "out.idml")
			if err := Write(pkg, out); err != nil {
				t.Fatalf("Write: %v", err)
			}
			orig := readZipFiles(t, path)
			gen := readZipFiles(t, out)
			for name, in := range orig {
				if !bytes.Equal(in, gen[name]) {
					t.Errorf("%s changed although nothing was modified (%d -> %d bytes)", name, len(in), len(gen[name]))
				}
			}
		})
	}
}

// TestMasterSpreadsAreParsed checks that master spreads are exposed through
// the typed API and keep their element name on write.
func TestMasterSpreadsAreParsed(t *testing.T) {
	pkg, err := Read(filepath.Join("..", "..", "testdata", "example.idml"))
	if err != nil {
		t.Fatal(err)
	}
	masters, err := pkg.MasterSpreads()
	if err != nil {
		t.Fatal(err)
	}
	if len(masters) == 0 {
		t.Fatal("expected at least one master spread")
	}
	for name, ms := range masters {
		if !IsMasterSpreadPath(name) {
			t.Errorf("%s is not a master spread path", name)
		}
		if !ms.IsMaster() {
			t.Errorf("%s: IsMaster() = false", name)
		}
		if len(ms.Pages) == 0 {
			t.Errorf("%s: no pages parsed", name)
		}
		pkg.MarkModified(name)
	}
	out := filepath.Join(t.TempDir(), "out.idml")
	if err := Write(pkg, out); err != nil {
		t.Fatal(err)
	}
	for name, data := range readZipFiles(t, out) {
		if IsMasterSpreadPath(name) && !strings.Contains(string(data), "<idPkg:MasterSpread ") {
			t.Errorf("%s lost its MasterSpread root element", name)
		}
	}
}

// readFixture reads a test package with limits high enough for the large
// fixture, whose single spread exceeds the default per-file size limit.
func readFixture(path string) (*Package, error) {
	return ReadWithOptions(path, &ReadOptions{
		MaxTotalSize: 2 << 30,
		MaxFileSize:  1 << 30,
	})
}

// bigFixture is a large image-heavy document that is deliberately not committed
// (92 MB). Tests that need it skip when it is absent.
const bigFixture = "../../testdata/example_big.idml"

// requireBigFixture skips the test unless the large fixture is present locally.
func requireBigFixture(t *testing.T) string {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping large fixture under -short")
	}
	if _, err := os.Stat(bigFixture); err != nil {
		t.Skipf("large fixture not present: %v", err)
	}
	return bigFixture
}
