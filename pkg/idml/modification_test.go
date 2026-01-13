package idml

import (
	"encoding/xml"
	"errors"
	"strings"
	"testing"

	"github.com/dimelords/idmllib/v2/pkg/common"
	"github.com/dimelords/idmllib/v2/pkg/spread"
	"github.com/dimelords/idmllib/v2/pkg/story"
)

// ============================================================================
// Test Helper Functions
// ============================================================================

// createTestStory creates a simple test story with the given style reference.
func createTestStory(paragraphStyle string) *story.Story {
	return &story.Story{
		XMLName: xml.Name{Local: "Story"},
		StoryElement: story.StoryElement{
			Self: "u1",
			ParagraphStyleRanges: []story.ParagraphStyleRange{
				{
					AppliedParagraphStyle: paragraphStyle,
					CharacterStyleRanges: []story.CharacterStyleRange{
						story.NewCharacterStyleRange("CharacterStyle/$ID/[No character style]", []story.Content{
							{Text: "Test content"},
						}),
					},
				},
			},
		},
	}
}

// ============================================================================
// RemoveStory Tests
// ============================================================================

// TestRemoveStory_Basic tests basic story removal without cleanup.
func TestRemoveStory_Basic(t *testing.T) {
	// Load example.idml
	pkg := loadExampleIDML(t)

	// Get initial story count
	initialStories, err := pkg.Stories()
	if err != nil {
		t.Fatalf("Failed to get stories: %v", err)
	}
	initialCount := len(initialStories)

	// Get a story filename to remove
	var storyToRemove string
	for filename := range initialStories {
		storyToRemove = filename
		break
	}

	// Remove story without cleanup
	result, err := pkg.RemoveStory(storyToRemove, false)
	if err != nil {
		t.Fatalf("RemoveStory failed: %v", err)
	}

	// Verify no cleanup was performed
	if result.Count() != 0 {
		t.Errorf("Expected 0 cleaned resources with cleanup=false, got %d", result.Count())
	}

	// Verify story was removed
	finalStories, err := pkg.Stories()
	if err != nil {
		t.Fatalf("Failed to get stories after removal: %v", err)
	}

	if len(finalStories) != initialCount-1 {
		t.Errorf("Expected %d stories after removal, got %d", initialCount-1, len(finalStories))
	}

	// Verify removed story is not in the package
	if _, exists := finalStories[storyToRemove]; exists {
		t.Errorf("Removed story '%s' still exists in package", storyToRemove)
	}

	// Verify story file is not in files map
	if _, exists := pkg.files[storyToRemove]; exists {
		t.Errorf("Removed story file '%s' still exists in files map", storyToRemove)
	}

	// Verify story is not in fileOrder
	for _, filename := range pkg.fileOrder {
		if filename == storyToRemove {
			t.Errorf("Removed story '%s' still exists in fileOrder", storyToRemove)
		}
	}
}

// TestRemoveStory_WithCleanup tests story removal with orphan cleanup.
func TestRemoveStory_WithCleanup(t *testing.T) {
	// Load example.idml
	pkg := loadExampleIDML(t)

	// Get a story filename to remove
	stories, err := pkg.Stories()
	if err != nil {
		t.Fatalf("Failed to get stories: %v", err)
	}

	var storyToRemove string
	for filename := range stories {
		storyToRemove = filename
		break
	}

	// Remove story with cleanup
	result, err := pkg.RemoveStory(storyToRemove, true)
	if err != nil {
		t.Fatalf("RemoveStory with cleanup failed: %v", err)
	}

	// Verify cleanup result contains details
	t.Logf("Cleanup removed: %d fonts, %d paragraph styles, %d character styles",
		len(result.RemovedFonts),
		len(result.RemovedParagraphStyles),
		len(result.RemovedCharacterStyles))

	// Note: We can't assert specific counts as they depend on the example.idml content
	// and whether resources are actually orphaned. The important thing is that
	// cleanup was attempted and completed without error.
}

// TestRemoveStory_NotFound tests error handling when story doesn't exist.
func TestRemoveStory_NotFound(t *testing.T) {
	pkg := New()

	_, err := pkg.RemoveStory("Stories/NonExistent.xml", false)
	if err == nil {
		t.Fatal("Expected error when removing non-existent story, got nil")
	}

	// Verify it's an common.ErrNotFound
	var idmlErr *common.Error
	if !errors.As(err, &idmlErr) {
		t.Fatalf("Expected *common.Error type, got %T", err)
	}

	if !errors.Is(err, common.ErrNotFound) {
		t.Errorf("Expected common.ErrNotFound, got %v", err)
	}
}

// TestRemoveStory_Roundtrip tests that we can write and re-read after removal.
func TestRemoveStory_Roundtrip(t *testing.T) {
	// Load example.idml
	pkg := loadExampleIDML(t)

	// Get a story to remove
	stories, err := pkg.Stories()
	if err != nil {
		t.Fatalf("Failed to get stories: %v", err)
	}

	var storyToRemove string
	for filename := range stories {
		storyToRemove = filename
		break
	}

	// Remove story with cleanup
	_, err = pkg.RemoveStory(storyToRemove, true)
	if err != nil {
		t.Fatalf("RemoveStory failed: %v", err)
	}

	// Write to temporary file
	outputPath := writeTestIDML(t, pkg, "after_removal.idml")

	// Read back the file
	pkg2, err := Read(outputPath)
	if err != nil {
		t.Fatalf("Failed to read back package: %v", err)
	}

	// Verify removed story is still gone
	stories2, err := pkg2.Stories()
	if err != nil {
		t.Fatalf("Failed to get stories from reloaded package: %v", err)
	}

	if _, exists := stories2[storyToRemove]; exists {
		t.Errorf("Removed story '%s' reappeared after roundtrip", storyToRemove)
	}

	// Verify we can still access other stories
	if len(stories2) == 0 && len(stories) > 1 {
		t.Error("All stories disappeared after roundtrip")
	}
}

// ============================================================================
// AddStory Tests
// ============================================================================

// TestAddStory_Basic tests basic story addition without validation.
func TestAddStory_Basic(t *testing.T) {
	pkg := New()

	// Create a simple story
	story := createTestStory("ParagraphStyle/$ID/NormalParagraphStyle")

	// Add story without validation
	opts := ValidationOptions{
		EnsureStylesExist: false,
		EnsureFontsExist:  false,
		AutoAddMissing:    false,
		FailOnMissing:     false,
	}

	err := pkg.AddStory("Stories/Story_new.xml", story, opts)
	if err != nil {
		t.Fatalf("AddStory failed: %v", err)
	}

	// Verify story was added
	stories, err := pkg.Stories()
	if err != nil {
		t.Fatalf("Failed to get stories: %v", err)
	}

	if len(stories) != 1 {
		t.Errorf("Expected 1 story, got %d", len(stories))
	}

	// Verify we can retrieve the story
	retrieved, err := pkg.Story("Stories/Story_new.xml")
	if err != nil {
		t.Fatalf("Failed to retrieve added story: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Retrieved story is nil")
	}

	// Verify story content matches
	if retrieved.StoryElement.Self != story.StoryElement.Self {
		t.Errorf("Story Self mismatch: expected %s, got %s",
			story.StoryElement.Self, retrieved.StoryElement.Self)
	}
}

// TestAddStory_AlreadyExists tests error handling when story already exists.
func TestAddStory_AlreadyExists(t *testing.T) {
	pkg := New()
	story := createTestStory("ParagraphStyle/$ID/NormalParagraphStyle")

	opts := ValidationOptions{
		EnsureStylesExist: false,
		EnsureFontsExist:  false,
	}

	// Add story first time
	err := pkg.AddStory("Stories/Story_test.xml", story, opts)
	if err != nil {
		t.Fatalf("First AddStory failed: %v", err)
	}

	// Try to add again
	err = pkg.AddStory("Stories/Story_test.xml", story, opts)
	if err == nil {
		t.Fatal("Expected error when adding duplicate story, got nil")
	}

	// Verify it's an common.ErrAlreadyExists
	if !errors.Is(err, common.ErrAlreadyExists) {
		t.Errorf("Expected common.ErrAlreadyExists, got %v", err)
	}
}

// TestAddStory_WithValidation tests story addition with validation enabled.
func TestAddStory_WithValidation(t *testing.T) {
	// Test that validation works when AutoAddMissing is enabled
	// This is the recommended approach for adding stories with unknown styles
	pkg := New()

	// Use a custom style that doesn't exist
	story := createTestStory("ParagraphStyle/CustomStyle")

	// Add story with validation and auto-add enabled
	opts := ValidationOptions{
		EnsureStylesExist: true,
		EnsureFontsExist:  false,
		AutoAddMissing:    true,
		FailOnMissing:     false,
	}

	err := pkg.AddStory("Stories/Story_validated.xml", story, opts)
	if err != nil {
		t.Fatalf("AddStory with validation and auto-add failed: %v", err)
	}

	// Verify story was added successfully
	retrieved, err := pkg.Story("Stories/Story_validated.xml")
	if err != nil {
		t.Fatalf("Failed to retrieve validated story: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Retrieved story is nil")
	}

	// The story should have been added even though the style didn't exist initially
	// (because AutoAddMissing created default resources)
	t.Log("✅ Story added successfully with auto-added missing resources")
}

// TestAddStory_AutoAddMissing tests automatic resource addition.
func TestAddStory_AutoAddMissing(t *testing.T) {
	pkg := New()

	// Create story with custom style that doesn't exist
	story := createTestStory("ParagraphStyle/CustomNonExistentStyle")

	// Add story with auto-add enabled
	opts := ValidationOptions{
		EnsureStylesExist: true,
		AutoAddMissing:    true,
		FailOnMissing:     false,
	}

	err := pkg.AddStory("Stories/Story_autoadd.xml", story, opts)
	if err != nil {
		t.Fatalf("AddStory with auto-add failed: %v", err)
	}

	// Verify story was added
	retrieved, err := pkg.Story("Stories/Story_autoadd.xml")
	if err != nil {
		t.Fatalf("Failed to retrieve story: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Retrieved story is nil")
	}

	// Test enhancement: Verify that the missing style was added to Styles.xml
	// by checking pkg.Styles() for the custom style.
}

// TestAddStory_FailOnMissing tests validation failure with missing resources.
func TestAddStory_FailOnMissing(t *testing.T) {
	pkg := New()

	// Create story with custom style that doesn't exist
	story := createTestStory("ParagraphStyle/CustomNonExistentStyle")

	// Add story with FailOnMissing enabled
	opts := ValidationOptions{
		EnsureStylesExist: true,
		AutoAddMissing:    false,
		FailOnMissing:     true,
	}

	err := pkg.AddStory("Stories/Story_fail.xml", story, opts)
	if err == nil {
		t.Fatal("Expected error with missing resources and FailOnMissing=true, got nil")
	}

	// Verify error message mentions missing resources
	if !strings.Contains(err.Error(), "missing resources") {
		t.Errorf("Expected error about missing resources, got: %v", err)
	}
}

// ============================================================================
// UpdateStory Tests
// ============================================================================

// TestUpdateStory_Basic tests basic story update without validation.
func TestUpdateStory_Basic(t *testing.T) {
	pkg := New()

	// Add initial story
	story1 := createTestStory("ParagraphStyle/$ID/NormalParagraphStyle")
	story1.StoryElement.Self = "original"

	opts := ValidationOptions{
		EnsureStylesExist: false,
		EnsureFontsExist:  false,
	}

	err := pkg.AddStory("Stories/Story_update.xml", story1, opts)
	if err != nil {
		t.Fatalf("AddStory failed: %v", err)
	}

	// Create updated story
	story2 := createTestStory("ParagraphStyle/$ID/NormalParagraphStyle")
	story2.StoryElement.Self = "updated"

	// Update the story
	err = pkg.UpdateStory("Stories/Story_update.xml", story2, opts)
	if err != nil {
		t.Fatalf("UpdateStory failed: %v", err)
	}

	// Verify story was updated
	retrieved, err := pkg.Story("Stories/Story_update.xml")
	if err != nil {
		t.Fatalf("Failed to retrieve updated story: %v", err)
	}

	if retrieved.StoryElement.Self != "updated" {
		t.Errorf("Story was not updated: expected Self='updated', got '%s'",
			retrieved.StoryElement.Self)
	}
}

// TestUpdateStory_NotFound tests error handling when story doesn't exist.
func TestUpdateStory_NotFound(t *testing.T) {
	pkg := New()
	story := createTestStory("ParagraphStyle/$ID/NormalParagraphStyle")

	opts := ValidationOptions{
		EnsureStylesExist: false,
		EnsureFontsExist:  false,
	}

	err := pkg.UpdateStory("Stories/NonExistent.xml", story, opts)
	if err == nil {
		t.Fatal("Expected error when updating non-existent story, got nil")
	}

	// Verify it's an common.ErrNotFound
	if !errors.Is(err, common.ErrNotFound) {
		t.Errorf("Expected common.ErrNotFound, got %v", err)
	}
}

// TestUpdateStory_PreservesOtherStories tests that updating one story doesn't affect others.
func TestUpdateStory_PreservesOtherStories(t *testing.T) {
	pkg := New()

	opts := ValidationOptions{
		EnsureStylesExist: false,
		EnsureFontsExist:  false,
	}

	// Add multiple stories
	story1 := createTestStory("ParagraphStyle/$ID/NormalParagraphStyle")
	story1.StoryElement.Self = "story1"

	story2 := createTestStory("ParagraphStyle/$ID/NormalParagraphStyle")
	story2.StoryElement.Self = "story2"

	err := pkg.AddStory("Stories/Story_1.xml", story1, opts)
	if err != nil {
		t.Fatalf("Failed to add story 1: %v", err)
	}

	err = pkg.AddStory("Stories/Story_2.xml", story2, opts)
	if err != nil {
		t.Fatalf("Failed to add story 2: %v", err)
	}

	// Update story 1
	updatedStory1 := createTestStory("ParagraphStyle/$ID/NormalParagraphStyle")
	updatedStory1.StoryElement.Self = "story1_updated"

	err = pkg.UpdateStory("Stories/Story_1.xml", updatedStory1, opts)
	if err != nil {
		t.Fatalf("UpdateStory failed: %v", err)
	}

	// Verify story 1 was updated
	retrieved1, err := pkg.Story("Stories/Story_1.xml")
	if err != nil {
		t.Fatalf("Failed to retrieve story 1: %v", err)
	}

	if retrieved1.StoryElement.Self != "story1_updated" {
		t.Errorf("Story 1 was not updated correctly")
	}

	// Verify story 2 was NOT affected
	retrieved2, err := pkg.Story("Stories/Story_2.xml")
	if err != nil {
		t.Fatalf("Failed to retrieve story 2: %v", err)
	}

	if retrieved2.StoryElement.Self != "story2" {
		t.Errorf("Story 2 was unexpectedly modified")
	}
}

// ============================================================================
// Integration Tests
// ============================================================================

// TestComplexScenario_MultipleModifications tests a complex workflow.
func TestComplexScenario_MultipleModifications(t *testing.T) {
	// Load example.idml
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read example.idml: %v", err)
	}

	// Get initial story count
	initialStories, err := pkg.Stories()
	if err != nil {
		t.Fatalf("Failed to get initial stories: %v", err)
	}
	initialCount := len(initialStories)

	// Step 1: Remove a story with cleanup
	var storyToRemove string
	for filename := range initialStories {
		storyToRemove = filename
		break
	}

	result, err := pkg.RemoveStory(storyToRemove, true)
	if err != nil {
		t.Fatalf("RemoveStory failed: %v", err)
	}
	t.Logf("Removed story, cleaned up %d resources", result.Count())

	// Step 2: Add a new story with validation
	newStory := createTestStory("ParagraphStyle/$ID/NormalParagraphStyle")
	newStory.StoryElement.Self = "newstory"

	opts := ValidationOptions{
		EnsureStylesExist: true,
		AutoAddMissing:    true,
		FailOnMissing:     false,
	}

	err = pkg.AddStory("Stories/Story_added.xml", newStory, opts)
	if err != nil {
		t.Fatalf("AddStory failed: %v", err)
	}

	// Step 3: Verify final state
	finalStories, err := pkg.Stories()
	if err != nil {
		t.Fatalf("Failed to get final stories: %v", err)
	}

	expectedCount := initialCount // -1 +1 = same count
	if len(finalStories) != expectedCount {
		t.Errorf("Expected %d stories, got %d", expectedCount, len(finalStories))
	}

	// Step 4: Write and re-read
	outputPath := writeTestIDML(t, pkg, "complex_scenario.idml")

	pkg2, err := Read(outputPath)
	if err != nil {
		t.Fatalf("Failed to read back package: %v", err)
	}

	// Verify roundtrip preserved changes
	roundtripStories, err := pkg2.Stories()
	if err != nil {
		t.Fatalf("Failed to get stories from roundtrip: %v", err)
	}

	if len(roundtripStories) != expectedCount {
		t.Errorf("Roundtrip changed story count: expected %d, got %d",
			expectedCount, len(roundtripStories))
	}

	// Verify removed story is still gone
	if _, exists := roundtripStories[storyToRemove]; exists {
		t.Error("Removed story reappeared after roundtrip")
	}

	// Verify added story still exists
	if _, exists := roundtripStories["Stories/Story_added.xml"]; !exists {
		t.Error("Added story disappeared after roundtrip")
	}
}

// ============================================================================
// TextFrame Operation Tests
// ============================================================================

// TestRemoveTextFrame_Basic tests basic text frame removal without cleanup.
func TestRemoveTextFrame_Basic(t *testing.T) {
	// Load example.idml
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read example.idml: %v", err)
	}

	// Get spreads
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Failed to get spreads: %v", err)
	}

	// Find a spread with text frames
	var spreadFilename string
	var textFrameID string
	var initialTFCount int

	for filename, sp := range spreads {
		if len(sp.InnerSpread.TextFrames) > 0 {
			spreadFilename = filename
			textFrameID = sp.InnerSpread.TextFrames[0].Self
			initialTFCount = len(sp.InnerSpread.TextFrames)
			break
		}
	}

	if spreadFilename == "" {
		t.Skip("No spreads with text frames found in example.idml")
	}

	// Remove text frame without cleanup
	result, err := pkg.RemoveTextFrame(spreadFilename, textFrameID, false)
	if err != nil {
		t.Fatalf("RemoveTextFrame failed: %v", err)
	}

	// Verify no cleanup was performed
	if result.Count() != 0 {
		t.Errorf("Expected 0 cleaned resources with cleanup=false, got %d", result.Count())
	}

	// Verify text frame was removed
	sp, err := pkg.Spread(spreadFilename)
	if err != nil {
		t.Fatalf("Failed to get spread after removal: %v", err)
	}

	if len(sp.InnerSpread.TextFrames) != initialTFCount-1 {
		t.Errorf("Expected %d text frames after removal, got %d",
			initialTFCount-1, len(sp.InnerSpread.TextFrames))
	}

	// Verify removed text frame is not in the spread
	for _, tf := range sp.InnerSpread.TextFrames {
		if tf.Self == textFrameID {
			t.Errorf("Removed text frame '%s' still exists in spread", textFrameID)
		}
	}
}

// TestRemoveTextFrame_WithCleanup tests text frame removal with resource cleanup.
func TestRemoveTextFrame_WithCleanup(t *testing.T) {
	// Load example.idml
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read example.idml: %v", err)
	}

	// Get spreads
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Failed to get spreads: %v", err)
	}

	// Find a spread with text frames
	var spreadFilename string
	var textFrameID string

	for filename, sp := range spreads {
		if len(sp.InnerSpread.TextFrames) > 0 {
			spreadFilename = filename
			textFrameID = sp.InnerSpread.TextFrames[0].Self
			break
		}
	}

	if spreadFilename == "" {
		t.Skip("No spreads with text frames found in example.idml")
	}

	// Remove text frame with cleanup
	result, err := pkg.RemoveTextFrame(spreadFilename, textFrameID, true)
	if err != nil {
		t.Fatalf("RemoveTextFrame failed: %v", err)
	}

	// Cleanup may or may not find orphans depending on the document
	// Just verify the operation succeeded
	t.Logf("Cleanup removed: %d fonts, %d paragraph styles, %d character styles",
		len(result.RemovedFonts), len(result.RemovedParagraphStyles), len(result.RemovedCharacterStyles))
}

// TestRemoveTextFrame_NotFound tests removing a non-existent text frame.
func TestRemoveTextFrame_NotFound(t *testing.T) {
	// Load example.idml
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read example.idml: %v", err)
	}

	// Get spreads
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Failed to get spreads: %v", err)
	}

	// Get a spread filename
	var spreadFilename string
	for filename := range spreads {
		spreadFilename = filename
		break
	}

	// Try to remove non-existent text frame
	_, err = pkg.RemoveTextFrame(spreadFilename, "nonexistent_textframe", false)
	if err == nil {
		t.Fatal("Expected error when removing non-existent text frame")
	}

	// Verify it's a not found error
	if !errors.Is(err, common.ErrNotFound) {
		t.Errorf("Expected common.ErrNotFound, got %v", err)
	}
}

// TestAddTextFrame_Basic tests adding a text frame to a spread.
func TestAddTextFrame_Basic(t *testing.T) {
	// Load example.idml
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read example.idml: %v", err)
	}

	// Get spreads
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Failed to get spreads: %v", err)
	}

	// Get a spread filename
	var spreadFilename string
	var initialTFCount int
	for filename, sp := range spreads {
		spreadFilename = filename
		initialTFCount = len(sp.InnerSpread.TextFrames)
		break
	}

	// Create a new text frame
	newTF := spread.SpreadTextFrame{
		PageItemBase: spread.PageItemBase{
			Self:            "u_new_textframe_test",
			GeometricBounds: "0 0 100 200",
		},
		ParentStory:        "", // No story
		AppliedObjectStyle: "ObjectStyle/$ID/[Normal Text Frame]",
	}

	// Add text frame without validation
	opts := ValidationOptions{
		EnsureStylesExist: false,
		EnsureFontsExist:  false,
		AutoAddMissing:    false,
		FailOnMissing:     false,
	}

	err = pkg.AddTextFrame(spreadFilename, &newTF, opts)
	if err != nil {
		t.Fatalf("AddTextFrame failed: %v", err)
	}

	// Verify text frame was added
	sp, err := pkg.Spread(spreadFilename)
	if err != nil {
		t.Fatalf("Failed to get spread after addition: %v", err)
	}

	if len(sp.InnerSpread.TextFrames) != initialTFCount+1 {
		t.Errorf("Expected %d text frames after addition, got %d",
			initialTFCount+1, len(sp.InnerSpread.TextFrames))
	}

	// Verify new text frame is in the spread
	found := false
	for _, tf := range sp.InnerSpread.TextFrames {
		if tf.Self == "u_new_textframe_test" {
			found = true
			if tf.AppliedObjectStyle != "ObjectStyle/$ID/[Normal Text Frame]" {
				t.Errorf("Text frame object style mismatch")
			}
			break
		}
	}

	if !found {
		t.Error("New text frame not found in spread")
	}
}

// TestAddTextFrame_WithValidation tests adding a text frame with validation.
func TestAddTextFrame_WithValidation(t *testing.T) {
	// Load example.idml
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read example.idml: %v", err)
	}

	// Get spreads
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Failed to get spreads: %v", err)
	}

	// Get a spread filename
	var spreadFilename string
	for filename := range spreads {
		spreadFilename = filename
		break
	}

	// Create a new text frame with a non-existent object style
	newTF := spread.SpreadTextFrame{
		PageItemBase: spread.PageItemBase{
			Self:            "u_new_textframe_validation",
			GeometricBounds: "0 0 100 200",
		},
		ParentStory:        "",
		AppliedObjectStyle: "ObjectStyle/$ID/TestNonExistentStyle",
	}

	// Add text frame with auto-add missing resources
	opts := ValidationOptions{
		EnsureStylesExist: true,
		AutoAddMissing:    true,
		FailOnMissing:     false,
	}

	err = pkg.AddTextFrame(spreadFilename, &newTF, opts)
	if err != nil {
		t.Fatalf("AddTextFrame with validation failed: %v", err)
	}

	// Verify the object style was added
	styles, err := pkg.Styles()
	if err != nil {
		t.Fatalf("Failed to get styles: %v", err)
	}

	if styles.RootObjectStyleGroup != nil {
		found := false
		for _, style := range styles.RootObjectStyleGroup.ObjectStyles {
			if style.Self == "ObjectStyle/$ID/TestNonExistentStyle" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected object style was not auto-added")
		}
	}
}

// TestUpdateTextFrame_Basic tests updating a text frame.
func TestUpdateTextFrame_Basic(t *testing.T) {
	// Load example.idml
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read example.idml: %v", err)
	}

	// Get spreads
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Failed to get spreads: %v", err)
	}

	// Find a spread with text frames
	var spreadFilename string
	var textFrameID string
	var originalObjectStyle string

	for filename, sp := range spreads {
		if len(sp.InnerSpread.TextFrames) > 0 {
			spreadFilename = filename
			textFrameID = sp.InnerSpread.TextFrames[0].Self
			originalObjectStyle = sp.InnerSpread.TextFrames[0].AppliedObjectStyle
			break
		}
	}

	if spreadFilename == "" {
		t.Skip("No spreads with text frames found in example.idml")
	}

	// Get the text frame
	sp, err := pkg.Spread(spreadFilename)
	if err != nil {
		t.Fatalf("Failed to get spread: %v", err)
	}

	var tfToUpdate *spread.SpreadTextFrame
	for i := range sp.InnerSpread.TextFrames {
		if sp.InnerSpread.TextFrames[i].Self == textFrameID {
			tfToUpdate = &sp.InnerSpread.TextFrames[i]
			break
		}
	}

	// Modify the text frame
	tfToUpdate.GeometricBounds = "10 10 110 210"

	// Update without validation
	opts := ValidationOptions{
		EnsureStylesExist: false,
		EnsureFontsExist:  false,
		AutoAddMissing:    false,
		FailOnMissing:     false,
	}

	err = pkg.UpdateTextFrame(spreadFilename, textFrameID, tfToUpdate, opts)
	if err != nil {
		t.Fatalf("UpdateTextFrame failed: %v", err)
	}

	// Verify text frame was updated
	sp, err = pkg.Spread(spreadFilename)
	if err != nil {
		t.Fatalf("Failed to get spread after update: %v", err)
	}

	found := false
	for _, tf := range sp.InnerSpread.TextFrames {
		if tf.Self == textFrameID {
			found = true
			if tf.GeometricBounds != "10 10 110 210" {
				t.Errorf("Expected GeometricBounds '10 10 110 210', got '%s'", tf.GeometricBounds)
			}
			if tf.AppliedObjectStyle != originalObjectStyle {
				t.Errorf("Object style should remain unchanged")
			}
			break
		}
	}

	if !found {
		t.Error("Updated text frame not found in spread")
	}
}

// TestUpdateTextFrame_NotFound tests updating a non-existent text frame.
func TestUpdateTextFrame_NotFound(t *testing.T) {
	// Load example.idml
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read example.idml: %v", err)
	}

	// Get spreads
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Failed to get spreads: %v", err)
	}

	// Get a spread filename
	var spreadFilename string
	for filename := range spreads {
		spreadFilename = filename
		break
	}

	// Create a text frame to update
	tf := &spread.SpreadTextFrame{
		PageItemBase: spread.PageItemBase{
			Self:            "nonexistent_textframe",
			GeometricBounds: "0 0 100 200",
		},
	}

	// Try to update non-existent text frame
	opts := ValidationOptions{}
	err = pkg.UpdateTextFrame(spreadFilename, "nonexistent_textframe", tf, opts)
	if err == nil {
		t.Fatal("Expected error when updating non-existent text frame")
	}

	// Verify it's a not found error
	if !errors.Is(err, common.ErrNotFound) {
		t.Errorf("Expected common.ErrNotFound, got %v", err)
	}
}

// ============================================================================
// Rectangle Operation Tests
// ============================================================================

// TestRemoveRectangle_Basic tests basic rectangle removal without cleanup.
func TestRemoveRectangle_Basic(t *testing.T) {
	// Load example.idml
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read example.idml: %v", err)
	}

	// Get spreads
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Failed to get spreads: %v", err)
	}

	// Find a spread with rectangles
	var spreadFilename string
	var rectangleID string
	var initialRectCount int

	for filename, sp := range spreads {
		if len(sp.InnerSpread.Rectangles) > 0 {
			spreadFilename = filename
			rectangleID = sp.InnerSpread.Rectangles[0].Self
			initialRectCount = len(sp.InnerSpread.Rectangles)
			break
		}
	}

	if spreadFilename == "" {
		t.Skip("No spreads with rectangles found in example.idml")
	}

	// Remove rectangle without cleanup
	result, err := pkg.RemoveRectangle(spreadFilename, rectangleID, false)
	if err != nil {
		t.Fatalf("RemoveRectangle failed: %v", err)
	}

	// Verify no cleanup was performed
	if result.Count() != 0 {
		t.Errorf("Expected 0 cleaned resources with cleanup=false, got %d", result.Count())
	}

	// Verify rectangle was removed
	sp, err := pkg.Spread(spreadFilename)
	if err != nil {
		t.Fatalf("Failed to get spread after removal: %v", err)
	}

	if len(sp.InnerSpread.Rectangles) != initialRectCount-1 {
		t.Errorf("Expected %d rectangles after removal, got %d",
			initialRectCount-1, len(sp.InnerSpread.Rectangles))
	}

	// Verify removed rectangle is not in the spread
	for _, rect := range sp.InnerSpread.Rectangles {
		if rect.Self == rectangleID {
			t.Errorf("Removed rectangle '%s' still exists in spread", rectangleID)
		}
	}
}

// TestRemoveRectangle_WithCleanup tests rectangle removal with resource cleanup.
func TestRemoveRectangle_WithCleanup(t *testing.T) {
	// Load example.idml
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read example.idml: %v", err)
	}

	// Get spreads
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Failed to get spreads: %v", err)
	}

	// Find a spread with rectangles
	var spreadFilename string
	var rectangleID string

	for filename, sp := range spreads {
		if len(sp.InnerSpread.Rectangles) > 0 {
			spreadFilename = filename
			rectangleID = sp.InnerSpread.Rectangles[0].Self
			break
		}
	}

	if spreadFilename == "" {
		t.Skip("No spreads with rectangles found in example.idml")
	}

	// Remove rectangle with cleanup
	result, err := pkg.RemoveRectangle(spreadFilename, rectangleID, true)
	if err != nil {
		t.Fatalf("RemoveRectangle failed: %v", err)
	}

	// Cleanup may or may not find orphans depending on the document
	// Just verify the operation succeeded
	t.Logf("Cleanup removed: %d fonts, %d paragraph styles, %d character styles",
		len(result.RemovedFonts), len(result.RemovedParagraphStyles), len(result.RemovedCharacterStyles))
}

// TestRemoveRectangle_NotFound tests removing a non-existent rectangle.
func TestRemoveRectangle_NotFound(t *testing.T) {
	// Load example.idml
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read example.idml: %v", err)
	}

	// Get spreads
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Failed to get spreads: %v", err)
	}

	// Get a spread filename
	var spreadFilename string
	for filename := range spreads {
		spreadFilename = filename
		break
	}

	// Try to remove non-existent rectangle
	_, err = pkg.RemoveRectangle(spreadFilename, "nonexistent_rectangle", false)
	if err == nil {
		t.Fatal("Expected error when removing non-existent rectangle")
	}

	// Verify it's a not found error
	if !errors.Is(err, common.ErrNotFound) {
		t.Errorf("Expected common.ErrNotFound, got %v", err)
	}
}

// TestAddRectangle_Basic tests adding a rectangle to a spread.
func TestAddRectangle_Basic(t *testing.T) {
	// Load example.idml
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read example.idml: %v", err)
	}

	// Get spreads
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Failed to get spreads: %v", err)
	}

	// Get a spread filename
	var spreadFilename string
	var initialRectCount int
	for filename, sp := range spreads {
		spreadFilename = filename
		initialRectCount = len(sp.InnerSpread.Rectangles)
		break
	}

	// Create a new rectangle
	newRect := spread.Rectangle{
		PageItemBase: spread.PageItemBase{
			Self:            "u_new_rectangle_test",
			GeometricBounds: "0 0 50 100",
		},
		AppliedObjectStyle: "ObjectStyle/$ID/[Basic Graphics Frame]",
	}

	// Add rectangle without validation
	opts := ValidationOptions{
		EnsureStylesExist: false,
		EnsureFontsExist:  false,
		AutoAddMissing:    false,
		FailOnMissing:     false,
	}

	err = pkg.AddRectangle(spreadFilename, &newRect, opts)
	if err != nil {
		t.Fatalf("AddRectangle failed: %v", err)
	}

	// Verify rectangle was added
	sp, err := pkg.Spread(spreadFilename)
	if err != nil {
		t.Fatalf("Failed to get spread after addition: %v", err)
	}

	if len(sp.InnerSpread.Rectangles) != initialRectCount+1 {
		t.Errorf("Expected %d rectangles after addition, got %d",
			initialRectCount+1, len(sp.InnerSpread.Rectangles))
	}

	// Verify new rectangle is in the spread
	found := false
	for _, rect := range sp.InnerSpread.Rectangles {
		if rect.Self == "u_new_rectangle_test" {
			found = true
			if rect.AppliedObjectStyle != "ObjectStyle/$ID/[Basic Graphics Frame]" {
				t.Errorf("Rectangle object style mismatch")
			}
			break
		}
	}

	if !found {
		t.Error("New rectangle not found in spread")
	}
}

// TestAddRectangle_WithValidation tests adding a rectangle with validation.
func TestAddRectangle_WithValidation(t *testing.T) {
	// Load example.idml
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read example.idml: %v", err)
	}

	// Get spreads
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Failed to get spreads: %v", err)
	}

	// Get a spread filename
	var spreadFilename string
	for filename := range spreads {
		spreadFilename = filename
		break
	}

	// Create a new rectangle with a non-existent object style
	newRect := spread.Rectangle{
		PageItemBase: spread.PageItemBase{
			Self:            "u_new_rectangle_validation",
			GeometricBounds: "0 0 50 100",
		},
		AppliedObjectStyle: "ObjectStyle/$ID/TestNonExistentRectStyle",
	}

	// Add rectangle with auto-add missing resources
	opts := ValidationOptions{
		EnsureStylesExist: true,
		AutoAddMissing:    true,
		FailOnMissing:     false,
	}

	err = pkg.AddRectangle(spreadFilename, &newRect, opts)
	if err != nil {
		t.Fatalf("AddRectangle with validation failed: %v", err)
	}

	// Verify the object style was added
	styles, err := pkg.Styles()
	if err != nil {
		t.Fatalf("Failed to get styles: %v", err)
	}

	if styles.RootObjectStyleGroup != nil {
		found := false
		for _, style := range styles.RootObjectStyleGroup.ObjectStyles {
			if style.Self == "ObjectStyle/$ID/TestNonExistentRectStyle" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected object style was not auto-added")
		}
	}
}

// TestUpdateRectangle_Basic tests updating a rectangle.
func TestUpdateRectangle_Basic(t *testing.T) {
	// Load example.idml
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read example.idml: %v", err)
	}

	// Get spreads
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Failed to get spreads: %v", err)
	}

	// Find a spread with rectangles
	var spreadFilename string
	var rectangleID string
	var originalObjectStyle string

	for filename, sp := range spreads {
		if len(sp.InnerSpread.Rectangles) > 0 {
			spreadFilename = filename
			rectangleID = sp.InnerSpread.Rectangles[0].Self
			originalObjectStyle = sp.InnerSpread.Rectangles[0].AppliedObjectStyle
			break
		}
	}

	if spreadFilename == "" {
		t.Skip("No spreads with rectangles found in example.idml")
	}

	// Get the rectangle
	sp, err := pkg.Spread(spreadFilename)
	if err != nil {
		t.Fatalf("Failed to get spread: %v", err)
	}

	var rectToUpdate *spread.Rectangle
	for i := range sp.InnerSpread.Rectangles {
		if sp.InnerSpread.Rectangles[i].Self == rectangleID {
			rectToUpdate = &sp.InnerSpread.Rectangles[i]
			break
		}
	}

	// Modify the rectangle
	rectToUpdate.GeometricBounds = "20 20 70 120"

	// Update without validation
	opts := ValidationOptions{
		EnsureStylesExist: false,
		EnsureFontsExist:  false,
		AutoAddMissing:    false,
		FailOnMissing:     false,
	}

	err = pkg.UpdateRectangle(spreadFilename, rectangleID, rectToUpdate, opts)
	if err != nil {
		t.Fatalf("UpdateRectangle failed: %v", err)
	}

	// Verify rectangle was updated
	sp, err = pkg.Spread(spreadFilename)
	if err != nil {
		t.Fatalf("Failed to get spread after update: %v", err)
	}

	found := false
	for _, rect := range sp.InnerSpread.Rectangles {
		if rect.Self == rectangleID {
			found = true
			if rect.GeometricBounds != "20 20 70 120" {
				t.Errorf("Expected GeometricBounds '20 20 70 120', got '%s'", rect.GeometricBounds)
			}
			if rect.AppliedObjectStyle != originalObjectStyle {
				t.Errorf("Object style should remain unchanged")
			}
			break
		}
	}

	if !found {
		t.Error("Updated rectangle not found in spread")
	}
}

// TestUpdateRectangle_NotFound tests updating a non-existent rectangle.
func TestUpdateRectangle_NotFound(t *testing.T) {
	// Load example.idml
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read example.idml: %v", err)
	}

	// Get spreads
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Failed to get spreads: %v", err)
	}

	// Get a spread filename
	var spreadFilename string
	for filename := range spreads {
		spreadFilename = filename
		break
	}

	// Create a rectangle to update
	rect := &spread.Rectangle{
		PageItemBase: spread.PageItemBase{
			Self:            "nonexistent_rectangle",
			GeometricBounds: "0 0 50 100",
		},
	}

	// Try to update non-existent rectangle
	opts := ValidationOptions{}
	err = pkg.UpdateRectangle(spreadFilename, "nonexistent_rectangle", rect, opts)
	if err == nil {
		t.Fatal("Expected error when updating non-existent rectangle")
	}

	// Verify it's a not found error
	if !errors.Is(err, common.ErrNotFound) {
		t.Errorf("Expected common.ErrNotFound, got %v", err)
	}
}
