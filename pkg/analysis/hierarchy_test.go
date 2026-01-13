package analysis

import (
	"testing"

	"github.com/dimelords/idmllib/v2/internal/testutil"
	"github.com/dimelords/idmllib/v2/pkg/idml"
)

// loadExampleIDML loads the standard example.idml test file.
func loadExampleIDML(t *testing.T) *idml.Package {
	t.Helper()
	path := testutil.TestDataPath(t, "example.idml")
	pkg, err := idml.Read(path)
	if err != nil {
		t.Fatalf("Failed to read example.idml: %v", err)
	}
	return pkg
}

// loadPlainIDML loads the standard plain.idml test file.
func loadPlainIDML(t *testing.T) *idml.Package {
	t.Helper()
	path := testutil.TestDataPath(t, "plain.idml")
	pkg, err := idml.Read(path)
	if err != nil {
		t.Fatalf("Failed to read plain.idml: %v", err)
	}
	return pkg
}

// TestResolveStyleHierarchies tests the ResolveStyleHierarchies method.
func TestResolveStyleHierarchies(t *testing.T) {
	// Read a real IDML file with style hierarchies
	pkg := loadExampleIDML(t)

	// Create a tracker
	tracker := NewDependencyTracker(pkg)

	// Add some styles that we know have parents
	// Based on the Styles.xml in example.idml:
	// - CharacterStyle/Naviga%3aFreddans is based on CharacterStyle/Naviga%3aInitial Kepler REP
	// - CharacterStyle/Naviga%3aInitial Kepler REP is based on $ID/[No character style]
	tracker.deps.CharacterStyles["CharacterStyle/Naviga%3aFreddans"] = true

	// Resolve the hierarchies
	if err := tracker.ResolveStyleHierarchies(); err != nil {
		t.Fatalf("ResolveStyleHierarchies failed: %v", err)
	}

	// Check that the parent style was added
	if !tracker.deps.CharacterStyles["CharacterStyle/Naviga%3aInitial Kepler REP"] {
		t.Errorf("Expected parent style 'CharacterStyle/Naviga%%3aInitial Kepler REP' to be in dependencies")
	}

	// Check that the built-in style was NOT added (it's a $ID/ style)
	if tracker.deps.CharacterStyles["$ID/[No character style]"] {
		t.Errorf("Built-in style '$ID/[No character style]' should not be in dependencies")
	}

	t.Logf("Total character styles after hierarchy resolution: %d", len(tracker.deps.CharacterStyles))
	for style := range tracker.deps.CharacterStyles {
		t.Logf("  - %s", style)
	}
}

// TestResolveStyleHierarchies_MultiLevel tests multi-level style inheritance.
func TestResolveStyleHierarchies_MultiLevel(t *testing.T) {
	pkg := loadExampleIDML(t)

	tracker := NewDependencyTracker(pkg)

	// Add a style that we know has a multi-level hierarchy
	// CharacterStyle/Naviga%3aFreddans -> CharacterStyle/Naviga%3aInitial Kepler REP -> $ID/[No character style]
	tracker.deps.CharacterStyles["CharacterStyle/Naviga%3aFreddans"] = true

	if err := tracker.ResolveStyleHierarchies(); err != nil {
		t.Fatalf("ResolveStyleHierarchies failed: %v", err)
	}

	// Both the immediate parent and grandparent (before built-in) should be included
	if !tracker.deps.CharacterStyles["CharacterStyle/Naviga%3aInitial Kepler REP"] {
		t.Errorf("Expected intermediate parent style to be in dependencies")
	}

	// The built-in style should NOT be included
	if tracker.deps.CharacterStyles["$ID/[No character style]"] {
		t.Errorf("Built-in style should not be in dependencies")
	}
}

// TestResolveStyleHierarchies_NoParent tests styles with no parent.
func TestResolveStyleHierarchies_NoParent(t *testing.T) {
	pkg := loadExampleIDML(t)

	tracker := NewDependencyTracker(pkg)

	// Add a style that is based directly on the built-in style
	tracker.deps.CharacterStyles["CharacterStyle/BIL fotokreditering"] = true

	initialCount := len(tracker.deps.CharacterStyles)

	if err := tracker.ResolveStyleHierarchies(); err != nil {
		t.Fatalf("ResolveStyleHierarchies failed: %v", err)
	}

	// The count should be the same (no additional styles added)
	// because the parent is a built-in style
	if len(tracker.deps.CharacterStyles) != initialCount {
		t.Errorf("Expected %d styles, got %d", initialCount, len(tracker.deps.CharacterStyles))
	}
}

// TestResolveStyleHierarchies_ParagraphStyles tests paragraph style hierarchies.
func TestResolveStyleHierarchies_ParagraphStyles(t *testing.T) {
	pkg := loadExampleIDML(t)

	tracker := NewDependencyTracker(pkg)

	// Add some paragraph styles
	// We need to check the actual Styles.xml to see which styles have parents
	// For now, let's just verify the method doesn't crash with paragraph styles
	tracker.deps.ParagraphStyles["ParagraphStyle/SomeStyle"] = true

	if err := tracker.ResolveStyleHierarchies(); err != nil {
		t.Fatalf("ResolveStyleHierarchies failed: %v", err)
	}
}

// TestResolveStyleHierarchies_ObjectStyles tests object style hierarchies.
func TestResolveStyleHierarchies_ObjectStyles(t *testing.T) {
	pkg := loadExampleIDML(t)

	tracker := NewDependencyTracker(pkg)

	// Add some object styles
	tracker.deps.ObjectStyles["ObjectStyle/SomeStyle"] = true

	if err := tracker.ResolveStyleHierarchies(); err != nil {
		t.Fatalf("ResolveStyleHierarchies failed: %v", err)
	}
}

// TestResolveStyleHierarchies_EmptyDeps tests with no dependencies.
func TestResolveStyleHierarchies_EmptyDeps(t *testing.T) {
	pkg := loadExampleIDML(t)

	tracker := NewDependencyTracker(pkg)

	// Don't add any styles - just test that it doesn't crash
	if err := tracker.ResolveStyleHierarchies(); err != nil {
		t.Fatalf("ResolveStyleHierarchies failed: %v", err)
	}

	// All counts should be zero
	if len(tracker.deps.ParagraphStyles) != 0 {
		t.Errorf("Expected 0 paragraph styles, got %d", len(tracker.deps.ParagraphStyles))
	}
	if len(tracker.deps.CharacterStyles) != 0 {
		t.Errorf("Expected 0 character styles, got %d", len(tracker.deps.CharacterStyles))
	}
	if len(tracker.deps.ObjectStyles) != 0 {
		t.Errorf("Expected 0 object styles, got %d", len(tracker.deps.ObjectStyles))
	}
}

// TestResolveStyleHierarchies_NoStylesFile tests with a package that has no Styles.xml.
func TestResolveStyleHierarchies_NoStylesFile(t *testing.T) {
	pkg := loadPlainIDML(t)

	tracker := NewDependencyTracker(pkg)
	tracker.deps.CharacterStyles["CharacterStyle/SomeStyle"] = true

	// Should not crash even if there's no Styles file
	if err := tracker.ResolveStyleHierarchies(); err != nil {
		t.Fatalf("ResolveStyleHierarchies failed: %v", err)
	}
}
