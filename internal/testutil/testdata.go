package testutil

import (
	"archive/zip"
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var (
	// preserveTestOutput is a flag to preserve test output files for debugging
	preserveTestOutput = flag.Bool("preserve-test-output", false, "preserve test output files for debugging")
)

// TestDataPath returns the absolute path to a test data file.
// It searches in the testdata directory relative to the package.
func TestDataPath(t *testing.T, filename string) string {
	t.Helper()

	// Try different relative paths to find testdata
	paths := []string{
		filepath.Join("../../testdata", filename),
		filepath.Join("../testdata", filename),
		filepath.Join("testdata", filename),
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			abs, _ := filepath.Abs(p)
			return abs
		}
	}

	t.Fatalf("Test data file not found: %s", filename)
	return ""
}

// TempIDML creates a temporary IDML file for testing.
// The file is automatically cleaned up when the test completes.
func TempIDML(t *testing.T, name string) string {
	t.Helper()

	dir := t.TempDir()
	return filepath.Join(dir, name)
}

// TempIDMLWithDebug creates a temporary IDML file for testing with optional debug preservation.
// If the -preserve-test-output flag is set, the file will be created in a debug directory
// and preserved after the test completes. Otherwise, it behaves like TempIDML.
func TempIDMLWithDebug(t *testing.T, name string) string {
	t.Helper()

	if *preserveTestOutput {
		// Create debug directory if it doesn't exist
		debugDir := "debug_test_output"
		if err := os.MkdirAll(debugDir, 0750); err != nil {
			t.Fatalf("Failed to create debug directory: %v", err)
		}

		// Create unique filename with test name
		debugName := t.Name() + "_" + name
		debugPath := filepath.Join(debugDir, debugName)

		// Clean up on test failure only if requested
		t.Cleanup(func() {
			if !t.Failed() && !*preserveTestOutput {
				os.Remove(debugPath)
			}
		})

		return debugPath
	}

	// Use regular temporary directory
	return TempIDML(t, name)
}

// ReadTestData reads the content of a test data file.
// It searches for the file in the testdata directory.
func ReadTestData(t *testing.T, filename string) []byte {
	t.Helper()

	path := TestDataPath(t, filename)
	// #nosec G304 - Test utility function with controlled paths
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read test data %s: %v", filename, err)
	}

	return data
}

// CreateTestZIP creates a ZIP file for testing with specified files.
// Returns the path to the created ZIP file in a temporary directory.
func CreateTestZIP(t *testing.T, files map[string][]byte) string {
	t.Helper()

	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "test.idml")

	// #nosec G304 - Test utility function with controlled temp directory path
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("failed to create test ZIP: %v", err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	for name, data := range files {
		fw, err := w.Create(name)
		if err != nil {
			t.Fatalf("failed to create file in ZIP: %v", err)
		}
		if _, err := fw.Write(data); err != nil {
			t.Fatalf("failed to write file in ZIP: %v", err)
		}
	}

	return zipPath
}

// CreateTestZIPWithDebug creates a ZIP file for testing with optional debug preservation.
// If the -preserve-test-output flag is set, the ZIP file will be preserved in a debug directory.
func CreateTestZIPWithDebug(t *testing.T, files map[string][]byte, name string) string {
	t.Helper()

	var zipPath string

	if *preserveTestOutput {
		// Create debug directory if it doesn't exist
		debugDir := "debug_test_output"
		if err := os.MkdirAll(debugDir, 0750); err != nil {
			t.Fatalf("Failed to create debug directory: %v", err)
		}

		// Create unique filename with test name
		debugName := t.Name() + "_" + name
		zipPath = filepath.Join(debugDir, debugName)

		// Clean up on test success only if not preserving
		t.Cleanup(func() {
			if !t.Failed() && !*preserveTestOutput {
				os.Remove(zipPath)
			}
		})
	} else {
		tmpDir := t.TempDir()
		zipPath = filepath.Join(tmpDir, name)
	}

	// #nosec G304 - Test utility function with controlled temp directory path
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("failed to create test ZIP: %v", err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	for filename, data := range files {
		fw, err := w.Create(filename)
		if err != nil {
			t.Fatalf("failed to create file in ZIP: %v", err)
		}
		if _, err := fw.Write(data); err != nil {
			t.Fatalf("failed to write file in ZIP: %v", err)
		}
	}

	return zipPath
}
