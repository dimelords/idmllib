package idml

import (
	"bytes"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func readFixtureLazy(path string) (*Package, error) {
	return ReadWithOptions(path, &ReadOptions{
		MaxTotalSize: 2 << 30,
		MaxFileSize:  1 << 30,
		Lazy:         true,
	})
}

// TestLazyWriteMatchesEager checks that a lazily read package writes the same
// archive contents as an eagerly read one, both untouched and after every
// story has been marked modified.
func TestLazyWriteMatchesEager(t *testing.T) {
	for _, path := range fidelityFixtures(t) {
		t.Run(filepath.Base(path), func(t *testing.T) {
			skipIfLarge(t, path)
			for _, modify := range []bool{false, true} {
				name := "untouched"
				if modify {
					name = "modified"
				}
				t.Run(name, func(t *testing.T) {
					eager, err := readFixture(path)
					if err != nil {
						t.Fatal(err)
					}
					lazy, err := readFixtureLazy(path)
					if err != nil {
						t.Fatal(err)
					}
					defer func() { _ = lazy.Close() }()
					if !lazy.Lazy() {
						t.Error("Lazy() = false")
					}

					for _, pkg := range []*Package{eager, lazy} {
						if !modify {
							continue
						}
						stories, err := pkg.Stories()
						if err != nil {
							t.Fatal(err)
						}
						for name := range stories {
							pkg.MarkModified(name)
						}
					}

					dir := t.TempDir()
					eagerOut := filepath.Join(dir, "eager.idml")
					lazyOut := filepath.Join(dir, "lazy.idml")
					if err := Write(eager, eagerOut); err != nil {
						t.Fatal(err)
					}
					if err := Write(lazy, lazyOut); err != nil {
						t.Fatal(err)
					}
					a, b := readZipFiles(t, eagerOut), readZipFiles(t, lazyOut)
					if len(a) != len(b) {
						t.Fatalf("file count differs: %d vs %d", len(a), len(b))
					}
					for name, want := range a {
						if !bytes.Equal(want, b[name]) {
							t.Errorf("%s differs between eager and lazy output (%d vs %d bytes)", name, len(want), len(b[name]))
						}
					}
				})
			}
		})
	}
}

// TestLazyLoadsOnlyWhatIsUsed documents which entries a lazily read package
// actually pulls in, which is the point of the mode.
func TestLazyLoadsOnlyWhatIsUsed(t *testing.T) {
	pkg, err := readFixtureLazy(filepath.Join("..", "..", "testdata", "example.idml"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = pkg.Close() }()

	total := pkg.FileCount()
	afterRead := pkg.LoadedFiles()
	if afterRead >= total {
		t.Errorf("read loaded %d of %d entries; expected only the ones it needs", afterRead, total)
	}
	if _, err := pkg.Styles(); err != nil {
		t.Fatal(err)
	}
	afterStyles := pkg.LoadedFiles()
	if afterStyles != afterRead+1 {
		t.Errorf("reading Styles.xml loaded %d entries, want exactly 1", afterStyles-afterRead)
	}
	if err := pkg.LoadAll(); err != nil {
		t.Fatal(err)
	}
	if got := pkg.LoadedFiles(); got != total {
		t.Errorf("LoadAll left %d of %d entries unloaded", total-got, total)
	}
}

// TestLazyCloseReleasesSource checks that Close frees the archive, that an
// entry loaded beforehand stays usable, and that one that was not gives a
// clear error rather than silently empty data.
func TestLazyCloseReleasesSource(t *testing.T) {
	pkg, err := readFixtureLazy(filepath.Join("..", "..", "testdata", "example.idml"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pkg.Styles(); err != nil {
		t.Fatal(err)
	}
	if err := pkg.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if err := pkg.Close(); err != nil {
		t.Errorf("second Close: %v", err)
	}
	if _, err := pkg.FileData(PathStyles); err != nil {
		t.Errorf("an entry loaded before Close should stay readable: %v", err)
	}
	_, err = pkg.FileData(PathFonts)
	if err == nil {
		t.Fatal("expected an error reading an unloaded entry after Close")
	}
	if !strings.Contains(err.Error(), "closed") {
		t.Errorf("error should explain the package was closed, got: %v", err)
	}
}

// TestLazyCloseOnEagerPackageIsNoOp keeps Close safe to call unconditionally.
func TestLazyCloseOnEagerPackageIsNoOp(t *testing.T) {
	pkg, err := readFixture(filepath.Join("..", "..", "testdata", "plain.idml"))
	if err != nil {
		t.Fatal(err)
	}
	if err := pkg.Close(); err != nil {
		t.Errorf("Close on an eager package: %v", err)
	}
	if _, err := pkg.Document(); err != nil {
		t.Errorf("eager package unusable after Close: %v", err)
	}
}

// TestLazyReducesMemoryForMetadataWorkload measures the case lazy reading is
// for: opening a large document to look at its structure only.
func TestLazyReducesMemoryForMetadataWorkload(t *testing.T) {
	fixture := requireBigFixture(t)
	measure := func(lazy bool) uint64 {
		runtime.GC()
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		var pkg *Package
		var err error
		if lazy {
			pkg, err = readFixtureLazy(fixture)
		} else {
			pkg, err = readFixture(fixture)
		}
		if err != nil {
			t.Fatal(err)
		}
		doc, err := pkg.Document()
		if err != nil {
			t.Fatal(err)
		}
		runtime.GC()
		runtime.ReadMemStats(&after)
		runtime.KeepAlive(pkg)
		runtime.KeepAlive(doc)
		_ = pkg.Close()
		if after.HeapAlloc < before.HeapAlloc {
			return 0
		}
		return after.HeapAlloc - before.HeapAlloc
	}
	eager := measure(false)
	lazy := measure(true)
	t.Logf("held heap for document structure only: eager %.1f MB, lazy %.1f MB",
		float64(eager)/(1<<20), float64(lazy)/(1<<20))
	if lazy >= eager/2 {
		t.Errorf("lazy reading held %d bytes vs eager %d; expected far less", lazy, eager)
	}
}
