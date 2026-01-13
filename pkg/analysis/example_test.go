package analysis_test

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/dimelords/idmllib/v2/pkg/analysis"
	"github.com/dimelords/idmllib/v2/pkg/idml"
)

// Example demonstrates basic usage of the analysis package.
func Example() {
	// Load a test IDML package
	path := filepath.Join("../../testdata", "example.idml")
	pkg, err := idml.Read(path)
	if err != nil {
		log.Fatal(err)
	}

	// Create a dependency tracker
	tracker := analysis.NewDependencyTracker(pkg)

	// Manually add some dependencies to demonstrate
	deps := tracker.Dependencies()
	deps.ParagraphStyles["ParagraphStyle/Heading1"] = true
	deps.CharacterStyles["CharacterStyle/Bold"] = true
	deps.Stories["Stories/Story_u1d8.xml"] = true

	// Get summary
	summary := tracker.Summary()
	fmt.Printf("Paragraph styles: %d\n", summary.ParagraphStylesCount)
	fmt.Printf("Character styles: %d\n", summary.CharacterStylesCount)
	fmt.Printf("Stories: %d\n", summary.StoriesCount)

	// Output:
	// Paragraph styles: 1
	// Character styles: 1
	// Stories: 1
}

// ExampleDependencyTracker_ResolveStyleHierarchies demonstrates style hierarchy resolution.
func ExampleDependencyTracker_ResolveStyleHierarchies() {
	// Load a test IDML package
	path := filepath.Join("../../testdata", "example.idml")
	pkg, err := idml.Read(path)
	if err != nil {
		log.Fatal(err)
	}

	// Create a dependency tracker
	tracker := analysis.NewDependencyTracker(pkg)

	// Add a style that has a parent hierarchy
	deps := tracker.Dependencies()
	deps.CharacterStyles["CharacterStyle/Naviga%3aFreddans"] = true

	// Resolve style hierarchies
	if err := tracker.ResolveStyleHierarchies(); err != nil {
		log.Fatal(err)
	}

	// Check if parent styles were added
	summary := tracker.Summary()
	fmt.Printf("Character styles after hierarchy resolution: %d\n", summary.CharacterStylesCount)

	// Output:
	// Character styles after hierarchy resolution: 2
}
