// Package analysis provides tools for analyzing IDML documents and tracking dependencies.
//
// This package is used primarily for IDMS export to collect all resources needed
// for selected page items. It analyzes the dependency graph of IDML elements to
// ensure that exported snippets contain all necessary styles, fonts, colors, and
// other resources.
//
// # Key Types
//
//   - DependencySet: Tracks all dependencies for a set of page items
//   - DependencyTracker: Analyzes IDML elements and builds dependency sets
//   - DependencySummary: Provides counts of each type of dependency
//
// # Usage
//
// Analyze dependencies for a selection:
//
//	// Create a tracker for the IDML package
//	tracker := analysis.NewDependencyTracker(pkg)
//
//	// Analyze a selection of page items
//	selection := idml.NewSelection()
//	selection.AddTextFrame(textFrame)
//	selection.AddRectangle(rectangle)
//
//	err := tracker.AnalyzeSelection(selection)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Resolve style hierarchies to include parent styles
//	err = tracker.ResolveStyleHierarchies()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Get the dependency set
//	deps := tracker.Dependencies()
//	fmt.Printf("Found %d stories, %d styles\n",
//	    len(deps.Stories), len(deps.ParagraphStyles))
//
// # Dependency Types
//
// The tracker identifies these types of dependencies:
//   - Stories: Referenced story files by filename
//   - ParagraphStyles: Referenced paragraph style IDs
//   - CharacterStyles: Referenced character style IDs
//   - ObjectStyles: Referenced object style IDs
//   - Colors: Referenced color IDs
//   - Swatches: Referenced swatch IDs
//   - Fonts: Referenced font families
//   - Layers: Referenced layer IDs
//   - Links: Referenced external file links (for images)
//   - ColorSpaces: Referenced color spaces (RGB, CMYK, Lab, etc.)
//
// # Style Hierarchy Resolution
//
// The ResolveStyleHierarchies method walks through style inheritance chains
// to ensure that when exporting an IDMS, all styles in the BasedOn hierarchy
// are included. This handles:
//   - Multi-level inheritance (grandparent styles, etc.)
//   - Circular reference detection (to prevent infinite loops)
//   - Built-in InDesign styles (which don't need to be included)
//
// # Architecture
//
// This package is part of the domain-specific architecture that supports
// IDMS export functionality. It works closely with:
//   - pkg/idml: For accessing IDML package content
//   - pkg/spread: For analyzing page items
//   - pkg/story: For analyzing text content
//   - pkg/resources: For style hierarchy information
package analysis
