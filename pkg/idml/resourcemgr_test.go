package idml

import (
	"testing"
)

// TestOrphanedResourcesHelpers_HelperMethods tests the helper methods on OrphanedResources.
func TestOrphanedResourcesHelpers_HelperMethods(t *testing.T) {
	tests := []struct {
		name        string
		orphans     *OrphanedResources
		wantOrphans bool
		wantCount   int
	}{
		{
			name:        "empty",
			orphans:     &OrphanedResources{},
			wantOrphans: false,
			wantCount:   0,
		},
		{
			name: "fonts only",
			orphans: &OrphanedResources{
				Fonts: []string{"Arial", "Helvetica"},
			},
			wantOrphans: true,
			wantCount:   2,
		},
		{
			name: "styles only",
			orphans: &OrphanedResources{
				ParagraphStyles: []string{"Style1", "Style2"},
				CharacterStyles: []string{"CharStyle1"},
			},
			wantOrphans: true,
			wantCount:   3,
		},
		{
			name: "mixed",
			orphans: &OrphanedResources{
				Fonts:           []string{"Arial"},
				ParagraphStyles: []string{"Style1"},
				CharacterStyles: []string{"CharStyle1"},
				ObjectStyles:    []string{"ObjStyle1"},
				Colors:          []string{"Red", "Blue"},
			},
			wantOrphans: true,
			wantCount:   6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.orphans.HasOrphans(); got != tt.wantOrphans {
				t.Errorf("HasOrphans() = %v, want %v", got, tt.wantOrphans)
			}
			if got := tt.orphans.Count(); got != tt.wantCount {
				t.Errorf("Count() = %v, want %v", got, tt.wantCount)
			}
		})
	}
}

// TestCleanupResultHelpers_HelperMethods tests the helper methods on CleanupResult.
func TestCleanupResultHelpers_HelperMethods(t *testing.T) {
	tests := []struct {
		name      string
		result    *CleanupResult
		wantCount int
	}{
		{
			name:      "empty",
			result:    &CleanupResult{},
			wantCount: 0,
		},
		{
			name: "fonts only",
			result: &CleanupResult{
				RemovedFonts: []string{"Arial", "Helvetica"},
			},
			wantCount: 2,
		},
		{
			name: "mixed",
			result: &CleanupResult{
				RemovedFonts:           []string{"Arial"},
				RemovedParagraphStyles: []string{"Style1", "Style2"},
				RemovedCharacterStyles: []string{"CharStyle1"},
			},
			wantCount: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.Count(); got != tt.wantCount {
				t.Errorf("Count() = %v, want %v", got, tt.wantCount)
			}
		})
	}
}

// TestMissingResourcesHelpers_HelperMethods tests the helper methods on MissingResources.
func TestMissingResourcesHelpers_HelperMethods(t *testing.T) {
	tests := []struct {
		name        string
		missing     *MissingResources
		wantMissing bool
	}{
		{
			name: "empty",
			missing: &MissingResources{
				Fonts:           make(map[string][]string),
				ParagraphStyles: make(map[string][]string),
				CharacterStyles: make(map[string][]string),
			},
			wantMissing: false,
		},
		{
			name: "fonts missing",
			missing: &MissingResources{
				Fonts: map[string][]string{
					"Arial": {"Story1"},
				},
				ParagraphStyles: make(map[string][]string),
				CharacterStyles: make(map[string][]string),
			},
			wantMissing: true,
		},
		{
			name: "styles missing",
			missing: &MissingResources{
				Fonts: make(map[string][]string),
				ParagraphStyles: map[string][]string{
					"Style1": {"Story1", "Story2"},
				},
				CharacterStyles: make(map[string][]string),
			},
			wantMissing: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.missing.HasMissing(); got != tt.wantMissing {
				t.Errorf("HasMissing() = %v, want %v", got, tt.wantMissing)
			}
		})
	}
}

// TestValidationError_ErrorInterface tests the ValidationError error interface.
func TestValidationError_ErrorInterface(t *testing.T) {
	ve := &ValidationError{
		ResourceType: "Font",
		ResourceID:   "Arial Black",
		UsedBy:       []string{"Story1", "Story2"},
		Message:      "font not found",
	}

	expectedError := "Font: Arial Black (used by 2 elements)"
	if got := ve.Error(); got != expectedError {
		t.Errorf("Error() = %q, want %q", got, expectedError)
	}
}

// TestDefaultCleanupOptions_SensibleDefaults tests that default cleanup options are sensible.
func TestDefaultCleanupOptions_SensibleDefaults(t *testing.T) {
	opts := DefaultCleanupOptions()

	// These should be enabled by default (safe to remove)
	if !opts.RemoveOrphanedFonts {
		t.Error("RemoveOrphanedFonts should be true by default")
	}
	if !opts.RemoveOrphanedParagraphStyles {
		t.Error("RemoveOrphanedParagraphStyles should be true by default")
	}
	if !opts.RemoveOrphanedCharacterStyles {
		t.Error("RemoveOrphanedCharacterStyles should be true by default")
	}

	// These should be disabled by default (preserve libraries)
	if opts.RemoveOrphanedObjectStyles {
		t.Error("RemoveOrphanedObjectStyles should be false by default")
	}
	if opts.RemoveOrphanedColors {
		t.Error("RemoveOrphanedColors should be false by default")
	}
	if opts.RemoveOrphanedSwatches {
		t.Error("RemoveOrphanedSwatches should be false by default")
	}
	if opts.RemoveOrphanedLayers {
		t.Error("RemoveOrphanedLayers should be false by default")
	}

	// DryRun should be false by default
	if opts.DryRun {
		t.Error("DryRun should be false by default")
	}
}

// TestDefaultValidationOptions_SensibleDefaults tests that default validation options are sensible.
func TestDefaultValidationOptions_SensibleDefaults(t *testing.T) {
	opts := DefaultValidationOptions()

	// All validation checks should be enabled by default
	if !opts.EnsureStylesExist {
		t.Error("EnsureStylesExist should be true by default")
	}
	if !opts.EnsureFontsExist {
		t.Error("EnsureFontsExist should be true by default")
	}
	if !opts.EnsureColorsExist {
		t.Error("EnsureColorsExist should be true by default")
	}
	if !opts.EnsureLayersExist {
		t.Error("EnsureLayersExist should be true by default")
	}

	// AutoAddMissing should be false by default (conservative)
	if opts.AutoAddMissing {
		t.Error("AutoAddMissing should be false by default")
	}

	// FailOnMissing should be true by default (fail fast)
	if !opts.FailOnMissing {
		t.Error("FailOnMissing should be true by default")
	}
}

// TestFindOrphans_EmptyPackage tests orphan detection with an empty package.
func TestFindOrphans_EmptyPackage(t *testing.T) {
	pkg := New()
	rm := NewResourceManager(pkg)

	orphans, err := rm.FindOrphans()
	if err != nil {
		t.Fatalf("FindOrphans() error = %v", err)
	}

	if orphans.HasOrphans() {
		t.Errorf("Empty package should have no orphans, got %d", orphans.Count())
	}
}

// TestFindOrphans_RealDocument tests orphan detection with a real IDML document.
func TestFindOrphans_RealDocument(t *testing.T) {
	// Load the example.idml test file
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read example.idml: %v", err)
	}

	rm := NewResourceManager(pkg)

	// Find orphans
	orphans, err := rm.FindOrphans()
	if err != nil {
		t.Fatalf("FindOrphans() error = %v", err)
	}

	// The example.idml likely has some orphaned resources
	// We don't check for specific orphans since the test file might change,
	// but we verify the method runs without error
	t.Logf("Found %d orphaned fonts", len(orphans.Fonts))
	t.Logf("Found %d orphaned paragraph styles", len(orphans.ParagraphStyles))
	t.Logf("Found %d orphaned character styles", len(orphans.CharacterStyles))
	t.Logf("Total orphans: %d", orphans.Count())
}

// TestCleanupOrphans_DryRun tests that DryRun mode doesn't actually remove anything.
func TestCleanupOrphans_DryRun(t *testing.T) {
	// Load the example.idml test file
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read example.idml: %v", err)
	}

	rm := NewResourceManager(pkg)

	// First, find orphans to know what we're working with
	orphansBefore, err := rm.FindOrphans()
	if err != nil {
		t.Fatalf("FindOrphans() error = %v", err)
	}

	// Run cleanup in DryRun mode
	opts := DefaultCleanupOptions()
	opts.DryRun = true
	result, err := rm.CleanupOrphans(opts)
	if err != nil {
		t.Fatalf("CleanupOrphans() error = %v", err)
	}

	// Verify the result shows what would be removed
	if len(result.RemovedFonts) != len(orphansBefore.Fonts) {
		t.Errorf("DryRun result mismatch: got %d fonts, want %d", len(result.RemovedFonts), len(orphansBefore.Fonts))
	}

	// Run FindOrphans again - should have the same results since DryRun didn't change anything
	orphansAfter, err := rm.FindOrphans()
	if err != nil {
		t.Fatalf("FindOrphans() after DryRun error = %v", err)
	}

	if orphansAfter.Count() != orphansBefore.Count() {
		t.Errorf("DryRun changed orphan count: before=%d, after=%d", orphansBefore.Count(), orphansAfter.Count())
	}
}

// TestCleanupOrphans_ActualCleanup tests that cleanup actually removes resources.
func TestCleanupOrphans_ActualCleanup(t *testing.T) {
	// Load the example.idml test file
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read example.idml: %v", err)
	}

	// Create a new resource manager for the initial scan
	rm1 := NewResourceManager(pkg)

	// Find orphans before cleanup
	orphansBefore, err := rm1.FindOrphans()
	if err != nil {
		t.Fatalf("FindOrphans() before cleanup error = %v", err)
	}

	// If there are no orphans, we can't test cleanup
	if !orphansBefore.HasOrphans() {
		t.Skip("No orphans found in example.idml, skipping cleanup test")
	}

	t.Logf("Found %d orphans before cleanup", orphansBefore.Count())

	// Run actual cleanup
	opts := DefaultCleanupOptions()
	opts.DryRun = false
	result, err := rm1.CleanupOrphans(opts)
	if err != nil {
		t.Fatalf("CleanupOrphans() error = %v", err)
	}

	t.Logf("Removed %d resources", result.Count())

	// Create a new resource manager after cleanup to re-analyze
	rm2 := NewResourceManager(pkg)

	// Find orphans after cleanup
	orphansAfter, err := rm2.FindOrphans()
	if err != nil {
		t.Fatalf("FindOrphans() after cleanup error = %v", err)
	}

	// After cleanup, there should be fewer or no orphans
	if orphansAfter.Count() > orphansBefore.Count() {
		t.Errorf("Cleanup increased orphan count: before=%d, after=%d", orphansBefore.Count(), orphansAfter.Count())
	}

	// Verify specific resource types were cleaned
	if len(result.RemovedFonts) > 0 {
		t.Logf("Removed %d fonts", len(result.RemovedFonts))
	}
	if len(result.RemovedParagraphStyles) > 0 {
		t.Logf("Removed %d paragraph styles", len(result.RemovedParagraphStyles))
	}
	if len(result.RemovedCharacterStyles) > 0 {
		t.Logf("Removed %d character styles", len(result.RemovedCharacterStyles))
	}
}

// TestCleanupOrphans_SelectiveCleanup tests cleanup with selective options.
func TestCleanupOrphans_SelectiveCleanup(t *testing.T) {
	// Load the example.idml test file
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read example.idml: %v", err)
	}

	rm := NewResourceManager(pkg)

	// Find all orphans
	orphans, err := rm.FindOrphans()
	if err != nil {
		t.Fatalf("FindOrphans() error = %v", err)
	}

	// Run cleanup with only fonts enabled
	opts := CleanupOptions{
		RemoveOrphanedFonts:           true,
		RemoveOrphanedParagraphStyles: false,
		RemoveOrphanedCharacterStyles: false,
		RemoveOrphanedObjectStyles:    false,
		RemoveOrphanedColors:          false,
		RemoveOrphanedSwatches:        false,
		RemoveOrphanedLayers:          false,
		DryRun:                        false,
	}

	result, err := rm.CleanupOrphans(opts)
	if err != nil {
		t.Fatalf("CleanupOrphans() error = %v", err)
	}

	// Should only have removed fonts
	if len(result.RemovedFonts) != len(orphans.Fonts) {
		t.Errorf("Expected to remove %d fonts, got %d", len(orphans.Fonts), len(result.RemovedFonts))
	}

	// Should not have removed styles
	if len(result.RemovedParagraphStyles) > 0 {
		t.Errorf("Expected to remove 0 paragraph styles, got %d", len(result.RemovedParagraphStyles))
	}
	if len(result.RemovedCharacterStyles) > 0 {
		t.Errorf("Expected to remove 0 character styles, got %d", len(result.RemovedCharacterStyles))
	}
}

// TestCleanupOrphans_RoundTrip tests that cleaned documents can be written and re-read.
func TestCleanupOrphans_RoundTrip(t *testing.T) {
	// Load the example.idml test file
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read example.idml: %v", err)
	}

	// Clean up orphans
	rm := NewResourceManager(pkg)
	_, err = rm.CleanupOrphans(DefaultCleanupOptions())
	if err != nil {
		t.Fatalf("CleanupOrphans() error = %v", err)
	}

	// Write to a temporary file
	tmpPath := "/tmp/cleaned_test.idml"
	if err := Write(pkg, tmpPath); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	// Re-read the cleaned file
	pkg2, err := Read(tmpPath)
	if err != nil {
		t.Fatalf("Read() after cleanup error = %v", err)
	}

	// Verify the cleaned file has no (or fewer) orphans
	rm2 := NewResourceManager(pkg2)
	orphans, err := rm2.FindOrphans()
	if err != nil {
		t.Fatalf("FindOrphans() on cleaned file error = %v", err)
	}

	t.Logf("Cleaned file has %d orphans", orphans.Count())
}
