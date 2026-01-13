package idml

import (
	"strings"
	"testing"
)

// TestExtractColorsFromParagraphStyles_ExtractsColors tests color extraction from paragraph style definitions.
func TestExtractColorsFromParagraphStyles_ExtractsColors(t *testing.T) {
	// Read the example.idml file which contains various colors
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read example.idml: %v", err)
	}

	// Create ResourceManager
	rm := NewResourceManager(pkg)

	// Analyze dependencies to get the dependency set
	deps, err := rm.analyzeDependencies()
	if err != nil {
		t.Fatalf("Failed to analyze dependencies: %v", err)
	}

	// Check if any colors were extracted from paragraph styles
	if len(deps.colors) == 0 {
		t.Logf("Note: No colors found in paragraph styles (may be expected for this test data)")
	} else {
		t.Logf("Found %d colors in dependency set:", len(deps.colors))
		for color := range deps.colors {
			t.Logf("  - %s", color)
		}
	}

	// Verify that special values are NOT tracked
	specialValues := []string{"Text Color", "Swatch/None", "$ID/[No color]"}
	for _, special := range specialValues {
		if deps.colors[special] {
			t.Errorf("Special value %q should not be tracked as a color", special)
		}
	}
}

// TestExtractColorsFromCharacterStyles_ExtractsColors tests color extraction from character style definitions.
func TestExtractColorsFromCharacterStyles_ExtractsColors(t *testing.T) {
	// Read the example.idml file
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read example.idml: %v", err)
	}

	rm := NewResourceManager(pkg)

	// Get styles to inspect what colors are available
	styles, err := pkg.Styles()
	if err != nil {
		t.Fatalf("Failed to get styles: %v", err)
	}

	// Count character styles with colors
	colorCount := 0
	if styles.RootCharacterStyleGroup != nil {
		for _, cs := range styles.RootCharacterStyleGroup.CharacterStyles {
			if cs.FillColor != "" && cs.FillColor != "Text Color" {
				colorCount++
				t.Logf("Character style %s has FillColor: %s", cs.Name, cs.FillColor)
			}
			if cs.StrokeColor != "" && cs.StrokeColor != "Swatch/None" {
				colorCount++
				t.Logf("Character style %s has StrokeColor: %s", cs.Name, cs.StrokeColor)
			}
		}
	}

	if colorCount == 0 {
		t.Logf("Note: No character styles with custom colors found in test data")
	}

	// Analyze dependencies
	deps, err := rm.analyzeDependencies()
	if err != nil {
		t.Fatalf("Failed to analyze dependencies: %v", err)
	}

	// Verify colors were extracted if they exist
	if colorCount > 0 && len(deps.colors) == 0 {
		t.Errorf("Expected colors to be extracted from character styles, but none were found")
	}
}

// TestFindColorUsage_ReportsUsageCorrectly tests that color usage reporting works correctly.
func TestFindColorUsage_ReportsUsageCorrectly(t *testing.T) {
	// Read the example.idml file
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read example.idml: %v", err)
	}

	rm := NewResourceManager(pkg)

	// Get all colors from the document to test with a real color
	graphic, err := pkg.Graphics()
	if err != nil {
		t.Fatalf("Failed to get graphic resources: %v", err)
	}

	var testColor string
	if len(graphic.Colors) > 0 {
		// Use the first non-built-in color
		for _, color := range graphic.Colors {
			if color.Self != "" && !strings.HasPrefix(color.Self, "$ID/") {
				testColor = color.Self
				break
			}
		}
	}

	if testColor == "" {
		t.Skip("No custom colors found in test data")
	}

	t.Logf("Testing with color: %s", testColor)

	// Find where the color is used
	usedBy := rm.findColorUsage(testColor)

	if len(usedBy) == 0 {
		t.Logf("Warning: Color %q not found in any spreads or styles (might be unused)", testColor)
	} else {
		t.Logf("Color %q found in %d locations:", testColor, len(usedBy))
		for _, location := range usedBy {
			t.Logf("  - %s", location)
		}
	}

	// Test with nonexistent color
	usedBy = rm.findColorUsage("Color/NonExistentColor")
	if len(usedBy) > 0 {
		t.Errorf("Expected 'Color/NonExistentColor' to not be used anywhere, but found in %d locations", len(usedBy))
	}
}

// TestFindParagraphStyleByID_FindsParagraphStyle tests the paragraph style lookup functionality.
func TestFindParagraphStyleByID_FindsParagraphStyle(t *testing.T) {
	// Read the example.idml file
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read example.idml: %v", err)
	}

	rm := NewResourceManager(pkg)
	styles, err := pkg.Styles()
	if err != nil {
		t.Fatalf("Failed to get styles: %v", err)
	}

	if styles.RootParagraphStyleGroup == nil {
		t.Skip("No paragraph style group found in test data")
	}

	// Get the first paragraph style ID from the test data
	if len(styles.RootParagraphStyleGroup.ParagraphStyles) == 0 {
		t.Skip("No paragraph styles found in test data")
	}

	testStyle := &styles.RootParagraphStyleGroup.ParagraphStyles[0]
	testStyleID := testStyle.Self

	// Test finding an existing style
	found := rm.findParagraphStyleByID(styles.RootParagraphStyleGroup, testStyleID)
	if found == nil {
		t.Errorf("findParagraphStyleByID(%s) = nil, want non-nil", testStyleID)
	} else if found.Self != testStyleID {
		t.Errorf("findParagraphStyleByID(%s).Self = %s, want %s", testStyleID, found.Self, testStyleID)
	}

	t.Logf("Successfully found style: %s (Name: %s)", found.Self, found.Name)

	// Test with non-existent style
	notFound := rm.findParagraphStyleByID(styles.RootParagraphStyleGroup, "ParagraphStyle/NonExistent")
	if notFound != nil {
		t.Errorf("findParagraphStyleByID(NonExistent) = %v, want nil", notFound)
	}

	// Test with nil group
	nilResult := rm.findParagraphStyleByID(nil, testStyleID)
	if nilResult != nil {
		t.Errorf("findParagraphStyleByID(nil, %s) = %v, want nil", testStyleID, nilResult)
	}
}

// TestSpreadUsesColor_DetectsColorUsage tests detecting color usage in spreads.
func TestSpreadUsesColor_DetectsColorUsage(t *testing.T) {
	// Read the example.idml file
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read example.idml: %v", err)
	}

	rm := NewResourceManager(pkg)

	// Get spreads
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Failed to get spreads: %v", err)
	}

	if len(spreads) == 0 {
		t.Skip("No spreads found in test data")
	}

	// Get a color from the document
	graphic, err := pkg.Graphics()
	if err != nil {
		t.Fatalf("Failed to get graphic resources: %v", err)
	}

	var testColor string
	if len(graphic.Colors) > 0 {
		for _, color := range graphic.Colors {
			if color.Self != "" && !strings.HasPrefix(color.Self, "$ID/") {
				testColor = color.Self
				break
			}
		}
	}

	if testColor == "" {
		t.Skip("No custom colors found in test data")
	}

	// Test each spread
	foundInSpread := false
	for filename, spread := range spreads {
		if rm.spreadUsesColor(spread, testColor) {
			t.Logf("Found color %q in spread: %s", testColor, filename)
			foundInSpread = true
			break
		}
	}

	if !foundInSpread {
		t.Logf("Note: Color %q not found in any spreads (may be used only in styles)", testColor)
	}

	// Test with nonexistent color - should always return false
	for filename, spread := range spreads {
		if rm.spreadUsesColor(spread, "Color/NonExistentColor") {
			t.Errorf("Spread %s should not use 'Color/NonExistentColor'", filename)
		}
	}
}

// TestParagraphStylesUseColor_DetectsColorUsage tests detecting color usage in paragraph styles.
func TestParagraphStylesUseColor_DetectsColorUsage(t *testing.T) {
	// Read the example.idml file
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read example.idml: %v", err)
	}

	rm := NewResourceManager(pkg)

	// Get styles
	styles, err := pkg.Styles()
	if err != nil {
		t.Fatalf("Failed to get styles: %v", err)
	}

	if styles.RootParagraphStyleGroup == nil {
		t.Skip("No paragraph style group found in test data")
	}

	// Find a paragraph style that has a color
	var testColor string
	for _, ps := range styles.RootParagraphStyleGroup.ParagraphStyles {
		if ps.FillColor != "" && ps.FillColor != "Text Color" && !strings.HasPrefix(ps.FillColor, "$ID/") {
			testColor = ps.FillColor
			t.Logf("Found paragraph style %s with color %s", ps.Name, testColor)
			break
		}
	}

	if testColor == "" {
		t.Skip("No paragraph styles with custom colors found in test data")
	}

	// Test detection
	if !rm.paragraphStylesUseColor(styles, testColor) {
		t.Errorf("Expected paragraphStylesUseColor to return true for color %q", testColor)
	}

	// Test with nonexistent color
	if rm.paragraphStylesUseColor(styles, "Color/NonExistentColor") {
		t.Errorf("Expected paragraphStylesUseColor to return false for nonexistent color")
	}
}

// TestCharacterStylesUseColor_DetectsColorUsage tests detecting color usage in character styles.
func TestCharacterStylesUseColor_DetectsColorUsage(t *testing.T) {
	// Read the example.idml file
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read example.idml: %v", err)
	}

	rm := NewResourceManager(pkg)

	// Get styles
	styles, err := pkg.Styles()
	if err != nil {
		t.Fatalf("Failed to get styles: %v", err)
	}

	if styles.RootCharacterStyleGroup == nil {
		t.Skip("No character style group found in test data")
	}

	// Find a character style that has a color
	var testColor string
	for _, cs := range styles.RootCharacterStyleGroup.CharacterStyles {
		if cs.FillColor != "" && cs.FillColor != "Text Color" && !strings.HasPrefix(cs.FillColor, "$ID/") {
			testColor = cs.FillColor
			t.Logf("Found character style %s with FillColor %s", cs.Name, testColor)
			break
		}
		if cs.StrokeColor != "" && cs.StrokeColor != "Swatch/None" && !strings.HasPrefix(cs.StrokeColor, "$ID/") {
			testColor = cs.StrokeColor
			t.Logf("Found character style %s with StrokeColor %s", cs.Name, testColor)
			break
		}
	}

	if testColor == "" {
		t.Skip("No character styles with custom colors found in test data")
	}

	// Test detection
	if !rm.characterStylesUseColor(styles, testColor) {
		t.Errorf("Expected characterStylesUseColor to return true for color %q", testColor)
	}

	// Test with nonexistent color
	if rm.characterStylesUseColor(styles, "Color/NonExistentColor") {
		t.Errorf("Expected characterStylesUseColor to return false for nonexistent color")
	}
}

// TestColorTrackingIntegration_IntegratesWithValidation tests color tracking with FindOrphans and ValidateReferences.
func TestColorTrackingIntegration_IntegratesWithValidation(t *testing.T) {
	// Read the example.idml file
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read example.idml: %v", err)
	}

	rm := NewResourceManager(pkg)

	// Test FindOrphans - colors used in styles should be tracked
	orphans, err := rm.FindOrphans()
	if err != nil {
		t.Fatalf("FindOrphans() error: %v", err)
	}

	t.Logf("FindOrphans found %d total orphaned resources", orphans.Count())
	t.Logf("  - Fonts: %d", len(orphans.Fonts))
	t.Logf("  - Paragraph Styles: %d", len(orphans.ParagraphStyles))
	t.Logf("  - Character Styles: %d", len(orphans.CharacterStyles))
	t.Logf("  - Colors: %d", len(orphans.Colors))

	// Test ValidateReferences - should detect missing color references
	validationErrors, err := rm.ValidateReferences()
	if err != nil {
		t.Fatalf("ValidateReferences() error: %v", err)
	}

	if len(validationErrors) > 0 {
		t.Logf("ValidateReferences found %d errors:", len(validationErrors))
		for _, valErr := range validationErrors {
			t.Logf("  - %s", valErr.Message)
		}
	} else {
		t.Logf("ValidateReferences found no errors (all references valid)")
	}

	// Note: Colors can be reported as orphaned if they're defined in Graphic.xml
	// but not actually used in any content (stories/spreads). A color defined in
	// an unused paragraph style is still considered unused at the document level.
	// This is expected behavior - the test just verifies the methods run without error.

	// Log some statistics about color usage
	styles, err := pkg.Styles()
	if err == nil {
		colorsInParagraphStyles := 0
		colorsInCharacterStyles := 0

		if styles.RootParagraphStyleGroup != nil {
			for _, ps := range styles.RootParagraphStyleGroup.ParagraphStyles {
				if ps.FillColor != "" && ps.FillColor != "Text Color" && !strings.HasPrefix(ps.FillColor, "$ID/") {
					colorsInParagraphStyles++
				}
			}
		}

		if styles.RootCharacterStyleGroup != nil {
			for _, cs := range styles.RootCharacterStyleGroup.CharacterStyles {
				if cs.FillColor != "" && cs.FillColor != "Text Color" && !strings.HasPrefix(cs.FillColor, "$ID/") {
					colorsInCharacterStyles++
				}
				if cs.StrokeColor != "" && cs.StrokeColor != "Swatch/None" && !strings.HasPrefix(cs.StrokeColor, "$ID/") {
					colorsInCharacterStyles++
				}
			}
		}

		t.Logf("Colors defined in paragraph styles: %d", colorsInParagraphStyles)
		t.Logf("Colors defined in character styles: %d", colorsInCharacterStyles)
	}
}

// TestBuiltInColorFiltering_FiltersBuiltInColors tests that built-in colors and special values are filtered out.
func TestBuiltInColorFiltering_FiltersBuiltInColors(t *testing.T) {
	// Read the example.idml file
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read example.idml: %v", err)
	}

	rm := NewResourceManager(pkg)

	// Analyze dependencies
	deps, err := rm.analyzeDependencies()
	if err != nil {
		t.Fatalf("Failed to analyze dependencies: %v", err)
	}

	// Verify built-in values are not tracked
	builtInValues := []string{
		"$ID/[No Color]",
		"$ID/[Paper]",
		"$ID/[Registration]",
		"Text Color",
		"Swatch/None",
	}

	for _, builtIn := range builtInValues {
		if deps.colors[builtIn] {
			t.Errorf("Built-in/special value %q should not be tracked as a color", builtIn)
		}
	}

	// Log all tracked colors for inspection
	if len(deps.colors) > 0 {
		t.Logf("Tracked colors:")
		for color := range deps.colors {
			t.Logf("  - %s", color)

			// Verify no tracked color is a built-in
			for _, builtIn := range builtInValues {
				if color == builtIn {
					t.Errorf("Built-in color %q found in tracked colors", color)
				}
			}
		}
	}
}
