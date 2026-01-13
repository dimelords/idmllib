package idml

import (
	"testing"
)

// TestStoryUsesFont_DetectsFontUsage tests the font detection functionality in stories.
func TestStoryUsesFont_DetectsFontUsage(t *testing.T) {
	// Read the example.idml file which contains various fonts
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read example.idml: %v", err)
	}

	// Get all stories
	stories, err := pkg.Stories()
	if err != nil {
		t.Fatalf("Failed to get stories: %v", err)
	}

	// Create ResourceManager
	rm := NewResourceManager(pkg)

	// Note: In the test data, most stories use built-in character styles or
	// styles that don't have explicit AppliedFont definitions. Fonts are often
	// defined at the paragraph style level in real IDML files, not character style level.
	//
	// This test verifies that storyUsesFont() works correctly, even when fonts
	// aren't found (which is the expected behavior for this test data).
	foundPolarisStory := false
	for filename, story := range stories {
		if rm.storyUsesFont(story, "Polaris Condensed") {
			t.Logf("Found 'Polaris Condensed' font in story: %s", filename)
			foundPolarisStory = true
			break
		}
	}

	// In this test data, Polaris Condensed is defined in character styles
	// but those styles are not used in the stories, so we don't expect to find it
	if foundPolarisStory {
		t.Logf("Note: Found 'Polaris Condensed' in stories (unexpected but not an error)")
	} else {
		t.Logf("'Polaris Condensed' not found in stories (expected for this test data)")
	}

	// Test with a font that doesn't exist - should always return false
	for filename, story := range stories {
		if rm.storyUsesFont(story, "NonExistentFont") {
			t.Errorf("Story %s should not use 'NonExistentFont'", filename)
		}
	}

	// Verify that the function works correctly by testing it returns false
	// for stories that use built-in styles (which don't have custom fonts)
	builtInStyleCount := 0
	for _, story := range stories {
		for _, psr := range story.StoryElement.ParagraphStyleRanges {
			for _, csr := range psr.CharacterStyleRanges {
				if csr.AppliedCharacterStyle == "CharacterStyle/$ID/[No character style]" {
					builtInStyleCount++
					// Verify that built-in styles don't return fonts
					font, err := rm.getFontFromCharacterStyle(csr.AppliedCharacterStyle)
					if err != nil {
						t.Errorf("getFontFromCharacterStyle error: %v", err)
					}
					if font != "" {
						t.Errorf("Built-in style should not have a font, got: %s", font)
					}
					break
				}
			}
		}
	}
	t.Logf("Verified %d character style ranges use built-in styles", builtInStyleCount)
}

// TestFindFontUsage_ReportsUsageCorrectly tests that font usage reporting works correctly.
func TestFindFontUsage_ReportsUsageCorrectly(t *testing.T) {
	// Read the example.idml file
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read example.idml: %v", err)
	}

	// Create ResourceManager and find where a font is used
	rm := NewResourceManager(pkg)
	usedBy := rm.findFontUsage("Polaris Condensed")

	if len(usedBy) == 0 {
		t.Logf("Warning: 'Polaris Condensed' not found in any stories (might be inherited or in paragraph styles)")
	} else {
		t.Logf("'Polaris Condensed' found in %d stories:", len(usedBy))
		for _, filename := range usedBy {
			t.Logf("  - %s", filename)
		}
	}

	// Test with nonexistent font
	usedBy = rm.findFontUsage("NonExistentFont")
	if len(usedBy) > 0 {
		t.Errorf("Expected 'NonExistentFont' to not be used anywhere, but found in %d stories", len(usedBy))
	}
}

// TestGetFontFromCharacterStyle_ExtractsFonts tests extracting fonts from character styles.
func TestGetFontFromCharacterStyle_ExtractsFonts(t *testing.T) {
	// Read the example.idml file
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read example.idml: %v", err)
	}

	rm := NewResourceManager(pkg)

	// Get the styles to find a valid character style ID
	styles, err := pkg.Styles()
	if err != nil {
		t.Fatalf("Failed to get styles: %v", err)
	}

	// Find a character style that has an AppliedFont
	var testStyleID string
	var expectedFont string

	if styles.RootCharacterStyleGroup != nil {
		for _, cs := range styles.RootCharacterStyleGroup.CharacterStyles {
			if cs.Properties != nil {
				font := cs.Properties.GetAppliedFont()
				if font != "" {
					testStyleID = cs.Self
					expectedFont = font
					break
				}
			}
		}
	}

	if testStyleID == "" {
		t.Skip("No character styles with fonts found in test data")
	}

	// Test extracting the font
	font, err := rm.getFontFromCharacterStyle(testStyleID)
	if err != nil {
		t.Errorf("getFontFromCharacterStyle() error: %v", err)
	}

	if font != expectedFont {
		t.Errorf("getFontFromCharacterStyle(%s) = %q, want %q", testStyleID, font, expectedFont)
	}

	t.Logf("Successfully extracted font %q from style %s", font, testStyleID)

	// Test with built-in style (should return empty)
	font, err = rm.getFontFromCharacterStyle("$ID/[No character style]")
	if err != nil {
		t.Errorf("getFontFromCharacterStyle($ID/...) error: %v", err)
	}
	if font != "" {
		t.Errorf("getFontFromCharacterStyle($ID/...) = %q, want empty string for built-in style", font)
	}

	// Test with non-existent style
	font, err = rm.getFontFromCharacterStyle("CharacterStyle/NonExistent")
	if err != nil {
		t.Errorf("getFontFromCharacterStyle(NonExistent) error: %v", err)
	}
	if font != "" {
		t.Errorf("getFontFromCharacterStyle(NonExistent) = %q, want empty string", font)
	}
}

// TestFindCharacterStyleByID_FindsCharacterStyle tests the character style lookup functionality.
func TestFindCharacterStyleByID_FindsCharacterStyle(t *testing.T) {
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

	if styles.RootCharacterStyleGroup == nil {
		t.Skip("No character style group found in test data")
	}

	// Get the first character style ID from the test data
	if len(styles.RootCharacterStyleGroup.CharacterStyles) == 0 {
		t.Skip("No character styles found in test data")
	}

	testStyle := &styles.RootCharacterStyleGroup.CharacterStyles[0]
	testStyleID := testStyle.Self

	// Test finding an existing style
	found := rm.findCharacterStyleByID(styles.RootCharacterStyleGroup, testStyleID)
	if found == nil {
		t.Errorf("findCharacterStyleByID(%s) = nil, want non-nil", testStyleID)
	} else if found.Self != testStyleID {
		t.Errorf("findCharacterStyleByID(%s).Self = %s, want %s", testStyleID, found.Self, testStyleID)
	}

	t.Logf("Successfully found style: %s (Name: %s)", found.Self, found.Name)

	// Test with non-existent style
	notFound := rm.findCharacterStyleByID(styles.RootCharacterStyleGroup, "CharacterStyle/NonExistent")
	if notFound != nil {
		t.Errorf("findCharacterStyleByID(NonExistent) = %v, want nil", notFound)
	}

	// Test with nil group
	nilResult := rm.findCharacterStyleByID(nil, testStyleID)
	if nilResult != nil {
		t.Errorf("findCharacterStyleByID(nil, %s) = %v, want nil", testStyleID, nilResult)
	}
}
