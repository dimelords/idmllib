package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sebdah/goldie/v2"
)

// GoldenFile provides golden file testing utilities.
type GoldenFile struct {
	g *goldie.Goldie
}

// NewGoldenFile creates a new GoldenFile tester.
// The dir parameter specifies where golden files are stored.
func NewGoldenFile(t *testing.T, dir string) *GoldenFile {
	t.Helper()

	return &GoldenFile{
		g: goldie.New(t,
			goldie.WithFixtureDir(dir),
			goldie.WithNameSuffix(".golden"),
		),
	}
}

// NewGoldenFileInTestdata creates a GoldenFile tester using testdata/golden directory.
// This is a convenience method for the common case.
func NewGoldenFileInTestdata(t *testing.T) *GoldenFile {
	t.Helper()
	return NewGoldenFile(t, filepath.Join("testdata", "golden"))
}

// Assert compares the actual data against the golden file.
// If they differ, the test fails with a detailed diff.
func (gf *GoldenFile) Assert(t *testing.T, name string, actual []byte) {
	t.Helper()
	gf.g.Assert(t, name, actual)
}

// AssertFile compares a file's contents against the golden file.
func (gf *GoldenFile) AssertFile(t *testing.T, name, filePath string) {
	t.Helper()

	// #nosec G304 - Test utility function with controlled test data paths
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", filePath, err)
	}

	gf.g.Assert(t, name, data)
}

// Update is a convenience wrapper around Assert that makes intent clear.
// Golden files are actually updated via the -update flag when running tests.
//
// Usage:
//
//	go test ./pkg/idml/... -run TestGolden -update
//
// This method exists primarily for documentation and clarity in test code.
func (gf *GoldenFile) Update(t *testing.T, name string, actual []byte) {
	t.Helper()

	// Note: goldie handles updates automatically when -update flag is used
	// This just calls Assert, which compares or updates depending on the flag
	gf.g.Assert(t, name, actual)

	t.Logf("✅ Golden file checked/updated: %s", name)
}
