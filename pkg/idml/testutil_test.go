package idml

import (
	"testing"

	"github.com/dimelords/idmllib/v2/internal/testutil"
)

// loadTestIDML loads a test IDML file from testdata directory.
// This is a convenience wrapper around idml.Read for test files.
func loadTestIDML(t *testing.T, filename string) *Package {
	t.Helper()

	path := testutil.TestDataPath(t, filename)
	pkg, err := Read(path)
	if err != nil {
		t.Fatalf("Failed to read test IDML %s: %v", filename, err)
	}

	return pkg
}

// loadExampleIDML loads the standard example.idml test file.
// This is commonly used across many tests.
func loadExampleIDML(t *testing.T) *Package {
	t.Helper()
	return loadTestIDML(t, "example.idml")
}

// loadPlainIDML loads the standard plain.idml test file.
// This is commonly used across many tests.
func loadPlainIDML(t *testing.T) *Package {
	t.Helper()
	return loadTestIDML(t, "plain.idml")
}

// writeTestIDML writes an IDML package to a temporary file and returns the path.
// The file is automatically cleaned up when the test completes.
func writeTestIDML(t *testing.T, pkg *Package, name string) string {
	t.Helper()

	outputPath := testutil.TempIDML(t, name)
	if err := Write(pkg, outputPath); err != nil {
		t.Fatalf("Failed to write test IDML %s: %v", name, err)
	}

	return outputPath
}

// writeTestIDMLWithDebug writes an IDML package to a file with optional debug preservation.
// If the -preserve-test-output flag is set, the file will be preserved in a debug directory.
func writeTestIDMLWithDebug(t *testing.T, pkg *Package, name string) string {
	t.Helper()

	outputPath := testutil.TempIDMLWithDebug(t, name)
	if err := Write(pkg, outputPath); err != nil {
		t.Fatalf("Failed to write test IDML %s: %v", name, err)
	}

	return outputPath
}
