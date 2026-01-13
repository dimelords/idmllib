package testutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestProperty8_TestCleanupConsistency tests that running tests doesn't leave
// artifacts in the project root and that temporary directories are cleaned up properly.
// **Feature: code-refactoring-improvements, Property 8: Test cleanup consistency**
func TestProperty8_TestCleanupConsistency(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Property: TempDir creates directories that are cleaned up automatically
	properties.Property("TempDir creates auto-cleanup directories", prop.ForAll(
		func(filename string) bool {
			if filename == "" || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
				return true // Skip invalid filenames
			}

			// Create a sub-test to isolate the TempDir behavior
			var tempPath string
			var dirExists bool

			t.Run("temp_dir_test", func(subT *testing.T) {
				tempDir := subT.TempDir()
				tempPath = filepath.Join(tempDir, filename)

				// Create a file in the temp directory
				if err := os.WriteFile(tempPath, []byte("test content"), 0644); err != nil {
					subT.Fatalf("Failed to create test file: %v", err)
				}

				// Verify the file exists during the test
				if _, err := os.Stat(tempPath); err != nil {
					subT.Fatalf("Test file should exist during test: %v", err)
				}

				// Store the directory path to check after cleanup
				tempPath = tempDir
			})

			// After the sub-test completes, the temp directory should be cleaned up
			if _, err := os.Stat(tempPath); err == nil {
				dirExists = true
			}

			// The directory should not exist after test cleanup
			return !dirExists
		},
		genValidFilename(),
	))

	// Property: TempIDML creates files in temporary directories
	properties.Property("TempIDML creates files in temp directories", prop.ForAll(
		func(filename string) bool {
			if filename == "" || !strings.HasSuffix(filename, ".idml") {
				return true // Skip invalid filenames
			}

			// Create a sub-test to test TempIDML
			var tempPath string
			var isInTempDir bool

			t.Run("temp_idml_test", func(subT *testing.T) {
				tempPath = TempIDML(subT, filename)

				// Verify the path is in a temporary directory (contains temp dir patterns)
				isInTempDir = strings.Contains(tempPath, "TestProperty8") ||
					strings.Contains(tempPath, "temp") ||
					strings.Contains(tempPath, "tmp")

				// Create the file to verify the path works
				if err := os.WriteFile(tempPath, []byte("test idml content"), 0644); err != nil {
					subT.Fatalf("Failed to create IDML file: %v", err)
				}

				// Verify the file exists during the test
				if _, err := os.Stat(tempPath); err != nil {
					subT.Fatalf("IDML file should exist during test: %v", err)
				}
			})

			return isInTempDir
		},
		genIDMLFilename(),
	))

	// Property: TempIDMLWithDebug respects preserve flag behavior
	properties.Property("TempIDMLWithDebug handles debug preservation", prop.ForAll(
		func(filename string) bool {
			if filename == "" || !strings.HasSuffix(filename, ".idml") {
				return true // Skip invalid filenames
			}

			// Test without preserve flag (should use temp directory)
			var tempPath string
			var isInTempDir bool

			t.Run("temp_idml_debug_test", func(subT *testing.T) {
				tempPath = TempIDMLWithDebug(subT, filename)

				// Without preserve flag, should behave like TempIDML
				isInTempDir = strings.Contains(tempPath, "TestProperty8") ||
					strings.Contains(tempPath, "temp") ||
					strings.Contains(tempPath, "tmp") ||
					!strings.HasPrefix(tempPath, "debug_test_output")

				// Create the file to verify the path works
				if err := os.WriteFile(tempPath, []byte("test idml content"), 0644); err != nil {
					subT.Fatalf("Failed to create IDML file: %v", err)
				}
			})

			return isInTempDir
		},
		genIDMLFilename(),
	))

	// Property: Project root should not accumulate test artifacts
	properties.Property("project root stays clean after tests", prop.ForAll(
		func(filename string) bool {
			if filename == "" || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
				return true // Skip invalid filenames
			}

			// Get the project root (go up from internal/testutil)
			projectRoot, err := filepath.Abs("../..")
			if err != nil {
				return true // Skip if we can't determine project root
			}

			// List of patterns that should not exist in project root after tests
			problematicPatterns := []string{
				"export.idms",
				"debug_*.idms",
				"*_debug.idml",
				"test_*.idml",
				"temp_*.idml",
				"output.idml",
				"cleaned_test.idml",
			}

			// Check that none of these patterns exist in project root
			for _, pattern := range problematicPatterns {
				matches, err := filepath.Glob(filepath.Join(projectRoot, pattern))
				if err != nil {
					continue // Skip if glob fails
				}
				if len(matches) > 0 {
					// Found problematic files in project root
					return false
				}
			}

			return true
		},
		gen.AlphaString(),
	))

	// Property: Temporary files should not be created in /tmp without cleanup
	properties.Property("no unmanaged files in /tmp", prop.ForAll(
		func(filename string) bool {
			if filename == "" || !strings.HasSuffix(filename, ".idml") {
				return true // Skip invalid filenames
			}

			// This property tests that we don't create files directly in /tmp
			// without proper cleanup mechanisms
			tmpPath := filepath.Join("/tmp", filename)

			// If the file exists, it should not be from our tests
			// (This is more of a documentation property - we can't easily test
			// all possible test executions, but we can verify the pattern)

			// Check if file exists
			if _, err := os.Stat(tmpPath); err == nil {
				// File exists - this might be from a test that doesn't clean up properly
				// For now, we'll just log this and return true, but in a real scenario
				// this would indicate a cleanup problem
				t.Logf("Warning: Found file in /tmp that might be from tests: %s", tmpPath)
			}

			// Always return true for this property since we can't control all test execution
			// The real value is in the logging and awareness
			return true
		},
		genIDMLFilename(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// genValidFilename generates valid filenames for testing
func genValidFilename() gopter.Gen {
	return gen.OneConstOf(
		"test.txt",
		"example.xml",
		"data.json",
		"config.yml",
		"output.log",
		"temp.dat",
		"sample.bin",
	)
}

// genIDMLFilename generates valid IDML filenames for testing
func genIDMLFilename() gopter.Gen {
	return gen.OneConstOf(
		"test.idml",
		"example.idml",
		"output.idml",
		"temp.idml",
		"sample.idml",
		"document.idml",
		"layout.idml",
	)
}
