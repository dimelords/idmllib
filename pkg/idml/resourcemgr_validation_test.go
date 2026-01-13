package idml

import (
	"encoding/xml"
	"testing"

	"github.com/dimelords/idmllib/pkg/resources"
	"github.com/dimelords/idmllib/pkg/story"
)

// TestValidateReferences_EmptyPackage tests validation with an empty package.
func TestValidateReferences_EmptyPackage(t *testing.T) {
	pkg := New()
	rm := NewResourceManager(pkg)

	errors, err := rm.ValidateReferences()
	if err != nil {
		t.Fatalf("ValidateReferences() error = %v", err)
	}

	if len(errors) > 0 {
		t.Errorf("Empty package should have no validation errors, got %d", len(errors))
	}
}

// TestValidateReferences_ValidDocument tests validation with a document that may have some missing references.
// Note: example.idml actually has some missing styles, which is fine for testing validation.
func TestValidateReferences_ValidDocument(t *testing.T) {
	// Load the example.idml test file
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read example.idml: %v", err)
	}

	rm := NewResourceManager(pkg)

	errors, err := rm.ValidateReferences()
	if err != nil {
		t.Fatalf("ValidateReferences() error = %v", err)
	}

	// Log any validation errors found
	if len(errors) > 0 {
		t.Logf("Found %d validation errors (this is expected for example.idml):", len(errors))
		for i, verr := range errors {
			if i < 5 { // Only log first 5 to avoid cluttering output
				t.Logf("  - %s: %s", verr.ResourceType, verr.Message)
			}
		}
		if len(errors) > 5 {
			t.Logf("  ... and %d more", len(errors)-5)
		}
	} else {
		t.Log("No validation errors found")
	}
}

// TestFindMissingResources_EmptyPackage tests missing resource detection with empty package.
func TestFindMissingResources_EmptyPackage(t *testing.T) {
	pkg := New()
	rm := NewResourceManager(pkg)

	missing, err := rm.FindMissingResources()
	if err != nil {
		t.Fatalf("FindMissingResources() error = %v", err)
	}

	if missing.HasMissing() {
		t.Errorf("Empty package should have no missing resources, got Fonts=%d, ParagraphStyles=%d, CharacterStyles=%d",
			len(missing.Fonts), len(missing.ParagraphStyles), len(missing.CharacterStyles))
	}
}

// TestFindMissingResources_ValidDocument tests missing resource detection on example.idml.
// Note: example.idml may have some missing styles, which helps validate the detection logic.
func TestFindMissingResources_ValidDocument(t *testing.T) {
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read example.idml: %v", err)
	}

	rm := NewResourceManager(pkg)

	missing, err := rm.FindMissingResources()
	if err != nil {
		t.Fatalf("FindMissingResources() error = %v", err)
	}

	// Log what we found
	if missing.HasMissing() {
		t.Logf("Found missing resources in example.idml (this is expected):")

		if len(missing.Fonts) > 0 {
			t.Logf("  Missing fonts: %d", len(missing.Fonts))
		}

		if len(missing.ParagraphStyles) > 0 {
			t.Logf("  Missing paragraph styles: %d", len(missing.ParagraphStyles))
			// Show first few examples
			count := 0
			for style, usedBy := range missing.ParagraphStyles {
				if count < 3 {
					t.Logf("    - %s (used by %d files)", style, len(usedBy))
					count++
				}
			}
		}

		if len(missing.CharacterStyles) > 0 {
			t.Logf("  Missing character styles: %d", len(missing.CharacterStyles))
		}
	} else {
		t.Log("No missing resources found")
	}
}

// TestFindMissingResources_MissingStyles tests detection of missing styles.
func TestFindMissingResources_MissingStyles(t *testing.T) {
	// Create a package with a story that references a non-existent style
	pkg := New()

	// Create a story with a reference to a missing paragraph style
	st := &story.Story{
		XMLName: xml.Name{Local: "Story"},
		StoryElement: story.StoryElement{
			ParagraphStyleRanges: []story.ParagraphStyleRange{
				{
					AppliedParagraphStyle: "ParagraphStyle/MissingStyle",
					CharacterStyleRanges: []story.CharacterStyleRange{
						story.NewCharacterStyleRange("CharacterStyle/$ID/[No character style]", []story.Content{{Text: "Test text"}}),
					},
				},
			},
		},
	}

	// Marshal the story to bytes
	storyData, err := story.MarshalStory(st)
	if err != nil {
		t.Fatalf("Failed to marshal story: %v", err)
	}

	// Add the story to the package's file map
	pkg.files = make(map[string]*fileEntry)
	pkg.files["Stories/Story_u1.xml"] = &fileEntry{data: storyData}

	// Also cache it in stories map
	pkg.stories = map[string]*story.Story{
		"Stories/Story_u1.xml": st,
	}

	// Create an empty styles file
	pkg.styles = &resources.StylesFile{
		XMLName: xml.Name{Local: "idPkg:Styles"},
		RootParagraphStyleGroup: &resources.ParagraphStyleGroup{
			ParagraphStyles: []resources.ParagraphStyle{}, // Empty - no styles defined
		},
		RootCharacterStyleGroup: &resources.CharacterStyleGroup{
			CharacterStyles: []resources.CharacterStyle{}, // Empty - no styles defined
		},
	}

	rm := NewResourceManager(pkg)

	missing, err := rm.FindMissingResources()
	if err != nil {
		t.Fatalf("FindMissingResources() error = %v", err)
	}

	// Should detect the missing paragraph style
	if len(missing.ParagraphStyles) == 0 {
		t.Error("Expected to find missing paragraph style")
	}

	if _, exists := missing.ParagraphStyles["ParagraphStyle/MissingStyle"]; !exists {
		t.Error("Expected 'ParagraphStyle/MissingStyle' to be in missing resources")
	}

	// Check that the usage information is correct
	if usedBy, exists := missing.ParagraphStyles["ParagraphStyle/MissingStyle"]; exists {
		if len(usedBy) == 0 {
			t.Error("Expected usage information for missing style")
		}
		t.Logf("Missing style 'ParagraphStyle/MissingStyle' is used by: %v", usedBy)
	}
}

// TestFindMissingResources_NoStylesFile tests handling when Styles.xml doesn't exist.
func TestFindMissingResources_NoStylesFile(t *testing.T) {
	// Create a package with a story but no Styles.xml
	pkg := New()

	st := &story.Story{
		XMLName: xml.Name{Local: "Story"},
		StoryElement: story.StoryElement{
			ParagraphStyleRanges: []story.ParagraphStyleRange{
				{
					AppliedParagraphStyle: "ParagraphStyle/SomeStyle",
					CharacterStyleRanges: []story.CharacterStyleRange{
						story.NewCharacterStyleRange("CharacterStyle/SomeCharStyle", []story.Content{{Text: "Test"}}),
					},
				},
			},
		},
	}

	// Marshal the story
	storyData, err := story.MarshalStory(st)
	if err != nil {
		t.Fatalf("Failed to marshal story: %v", err)
	}

	// Add to files map
	pkg.files = make(map[string]*fileEntry)
	pkg.files["Stories/Story_u1.xml"] = &fileEntry{data: storyData}

	// Cache in stories map
	pkg.stories = map[string]*story.Story{
		"Stories/Story_u1.xml": st,
	}

	// Don't set pkg.styles - it will be nil

	rm := NewResourceManager(pkg)

	missing, err := rm.FindMissingResources()
	if err != nil {
		t.Fatalf("FindMissingResources() error = %v", err)
	}

	// When Styles.xml doesn't exist, all used styles should be reported as missing
	if len(missing.ParagraphStyles) == 0 {
		t.Error("Expected to find missing paragraph styles when Styles.xml doesn't exist")
	}

	if len(missing.CharacterStyles) == 0 {
		t.Error("Expected to find missing character styles when Styles.xml doesn't exist")
	}
}

// TestValidationError_Error tests the Error() method of ValidationError.
func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name     string
		verr     ValidationError
		expected string
	}{
		{
			name: "font error",
			verr: ValidationError{
				ResourceType: "Font",
				ResourceID:   "Arial Black",
				UsedBy:       []string{"Story1", "Story2"},
			},
			expected: "Font: Arial Black (used by 2 elements)",
		},
		{
			name: "paragraph style error",
			verr: ValidationError{
				ResourceType: "ParagraphStyle",
				ResourceID:   "ParagraphStyle/Custom",
				UsedBy:       []string{"Story1"},
			},
			expected: "ParagraphStyle: ParagraphStyle/Custom (used by 1 elements)",
		},
		{
			name: "no usage",
			verr: ValidationError{
				ResourceType: "Color",
				ResourceID:   "Color/Red",
				UsedBy:       []string{},
			},
			expected: "Color: Color/Red (used by 0 elements)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.verr.Error()
			if got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestFindParagraphStyleUsage_FindsUsage tests finding stories that use a specific paragraph style.
func TestFindParagraphStyleUsage_FindsUsage(t *testing.T) {
	pkg := New()

	// Create stories with different paragraph styles
	st1 := &story.Story{
		XMLName: xml.Name{Local: "Story"},
		StoryElement: story.StoryElement{
			ParagraphStyleRanges: []story.ParagraphStyleRange{
				{
					AppliedParagraphStyle: "ParagraphStyle/Style1",
					CharacterStyleRanges: []story.CharacterStyleRange{
						story.NewCharacterStyleRange("", []story.Content{{Text: "Text"}}),
					},
				},
			},
		},
	}

	st2 := &story.Story{
		XMLName: xml.Name{Local: "Story"},
		StoryElement: story.StoryElement{
			ParagraphStyleRanges: []story.ParagraphStyleRange{
				{
					AppliedParagraphStyle: "ParagraphStyle/Style2",
					CharacterStyleRanges: []story.CharacterStyleRange{
						story.NewCharacterStyleRange("", []story.Content{{Text: "Text"}}),
					},
				},
			},
		},
	}

	st3 := &story.Story{
		XMLName: xml.Name{Local: "Story"},
		StoryElement: story.StoryElement{
			ParagraphStyleRanges: []story.ParagraphStyleRange{
				{
					AppliedParagraphStyle: "ParagraphStyle/Style1", // Uses Style1
					CharacterStyleRanges: []story.CharacterStyleRange{
						story.NewCharacterStyleRange("", []story.Content{{Text: "Text"}}),
					},
				},
			},
		},
	}

	pkg.stories = map[string]*story.Story{
		"Stories/Story_u1.xml": st1,
		"Stories/Story_u2.xml": st2,
		"Stories/Story_u3.xml": st3,
	}

	rm := NewResourceManager(pkg)

	// Find usage of Style1
	usedBy := rm.findParagraphStyleUsage("ParagraphStyle/Style1")

	// Should be used by Story_u1.xml and Story_u3.xml
	if len(usedBy) != 2 {
		t.Errorf("Expected Style1 to be used by 2 stories, got %d", len(usedBy))
	}

	// Find usage of Style2
	usedBy = rm.findParagraphStyleUsage("ParagraphStyle/Style2")

	// Should be used by Story_u2.xml only
	if len(usedBy) != 1 {
		t.Errorf("Expected Style2 to be used by 1 story, got %d", len(usedBy))
	}

	// Find usage of non-existent style
	usedBy = rm.findParagraphStyleUsage("ParagraphStyle/NonExistent")

	// Should not be used by any story
	if len(usedBy) != 0 {
		t.Errorf("Expected non-existent style to be used by 0 stories, got %d", len(usedBy))
	}
}

// TestFindCharacterStyleUsage_FindsUsage tests finding stories that use a specific character style.
func TestFindCharacterStyleUsage_FindsUsage(t *testing.T) {
	pkg := New()

	st1 := &story.Story{
		XMLName: xml.Name{Local: "Story"},
		StoryElement: story.StoryElement{
			ParagraphStyleRanges: []story.ParagraphStyleRange{
				{
					AppliedParagraphStyle: "ParagraphStyle/Normal",
					CharacterStyleRanges: []story.CharacterStyleRange{
						story.NewCharacterStyleRange("CharacterStyle/Bold", []story.Content{{Text: "Text"}}),
					},
				},
			},
		},
	}

	st2 := &story.Story{
		XMLName: xml.Name{Local: "Story"},
		StoryElement: story.StoryElement{
			ParagraphStyleRanges: []story.ParagraphStyleRange{
				{
					AppliedParagraphStyle: "ParagraphStyle/Normal",
					CharacterStyleRanges: []story.CharacterStyleRange{
						story.NewCharacterStyleRange("CharacterStyle/Italic", []story.Content{{Text: "Text"}}),
					},
				},
			},
		},
	}

	pkg.stories = map[string]*story.Story{
		"Stories/Story_u1.xml": st1,
		"Stories/Story_u2.xml": st2,
	}

	rm := NewResourceManager(pkg)

	// Find usage of Bold character style
	usedBy := rm.findCharacterStyleUsage("CharacterStyle/Bold")
	if len(usedBy) != 1 {
		t.Errorf("Expected Bold character style to be used by 1 story, got %d", len(usedBy))
	}

	// Find usage of Italic character style
	usedBy = rm.findCharacterStyleUsage("CharacterStyle/Italic")
	if len(usedBy) != 1 {
		t.Errorf("Expected Italic character style to be used by 1 story, got %d", len(usedBy))
	}

	// Find usage of non-existent character style
	usedBy = rm.findCharacterStyleUsage("CharacterStyle/NonExistent")
	if len(usedBy) != 0 {
		t.Errorf("Expected non-existent character style to be used by 0 stories, got %d", len(usedBy))
	}
}
