package story

import (
	"encoding/xml"
	"strings"
	"testing"

	"github.com/dimelords/idmllib/internal/testutil"
	"github.com/google/go-cmp/cmp"
)

// TestParseStory_ParsesStoryXML tests parsing a Story XML file.
func TestParseStory_ParsesStoryXML(t *testing.T) {
	// Read test file
	data := testutil.ReadTestData(t, "story_u1d8.xml")

	// Parse story
	story, err := ParseStory(data)
	if err != nil {
		t.Fatalf("ParseStory failed: %v", err)
	}

	// Verify basic structure
	if story.DOMVersion == "" {
		t.Error("DOMVersion is empty")
	}

	// Verify story element
	if story.StoryElement.Self == "" {
		t.Error("Story Self is empty")
	}

	// Verify story preferences exist
	if story.StoryElement.StoryPreference == nil {
		t.Error("StoryPreference is nil")
	}

	// Verify paragraph style ranges exist
	if len(story.StoryElement.ParagraphStyleRanges) == 0 {
		t.Error("No ParagraphStyleRanges found")
	}
}

// TestStoryRoundtrip tests that we can parse and marshal back to identical XML.
func TestStoryRoundtrip(t *testing.T) {
	// Read original file
	originalData := testutil.ReadTestData(t, "story_u1d8.xml")

	// Parse
	story, err := ParseStory(originalData)
	if err != nil {
		t.Fatalf("ParseStory failed: %v", err)
	}

	// Marshal back
	regenerated, err := MarshalStory(story)
	if err != nil {
		t.Fatalf("MarshalStory failed: %v", err)
	}

	// Parse again to compare structures (not byte-for-byte)
	story2, err := ParseStory(regenerated)
	if err != nil {
		t.Fatalf("ParseStory (second) failed: %v", err)
	}

	// Compare structures (ignoring XMLName differences since Go adds namespace to all elements)
	opts := []cmp.Option{
		cmp.Comparer(func(a, b xml.Name) bool {
			// Only compare local name, ignore namespace differences
			return a.Local == b.Local
		}),
	}
	if diff := cmp.Diff(story, story2, opts...); diff != "" {
		t.Errorf("Story structures differ after roundtrip (-original +regenerated):\n%s", diff)
	}
}

// TestStoryXMLDeclaration verifies that the XML declaration is included.
func TestStoryXMLDeclaration(t *testing.T) {
	// Read test file
	data := testutil.ReadTestData(t, "story_u1d8.xml")

	// Parse story
	story, err := ParseStory(data)
	if err != nil {
		t.Fatalf("ParseStory failed: %v", err)
	}

	// Marshal back
	regenerated, err := MarshalStory(story)
	if err != nil {
		t.Fatalf("MarshalStory failed: %v", err)
	}

	// Check for XML declaration
	expectedDecl := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n"
	if len(regenerated) < len(expectedDecl) {
		t.Fatal("Generated XML is too short to contain declaration")
	}

	actualDecl := string(regenerated[:len(expectedDecl)])
	if actualDecl != expectedDecl {
		t.Errorf("XML declaration mismatch:\nwant: %q\ngot:  %q", expectedDecl, actualDecl)
	}
}

// TestStoryNamespace verifies that the namespace is correctly handled.
func TestStoryNamespace(t *testing.T) {
	// Read test file
	data := testutil.ReadTestData(t, "story_u1d8.xml")

	// Parse story
	story, err := ParseStory(data)
	if err != nil {
		t.Fatalf("ParseStory failed: %v", err)
	}

	// Check namespace
	expectedNamespace := "http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging"
	if story.XMLName.Space != expectedNamespace {
		t.Errorf("Namespace mismatch:\nwant: %s\ngot:  %s", expectedNamespace, story.XMLName.Space)
	}

	if story.XMLName.Local != "Story" {
		t.Errorf("Local name mismatch:\nwant: Story\ngot:  %s", story.XMLName.Local)
	}
}

// TestStoryContentExtraction tests extracting text content from a story.
func TestStoryContentExtraction(t *testing.T) {
	// Read test file
	data := testutil.ReadTestData(t, "story_u1d8.xml")

	// Parse story
	story, err := ParseStory(data)
	if err != nil {
		t.Fatalf("ParseStory failed: %v", err)
	}

	// Extract all text content
	var allText string
	for _, psr := range story.StoryElement.ParagraphStyleRanges {
		for _, csr := range psr.CharacterStyleRanges {
			for _, content := range csr.GetContent() {
				allText += content.Text
			}
		}
	}

	if allText == "" {
		t.Error("No text content extracted")
	}

	// Verify we got some expected content
	expectedSubstring := "Lorem ipsum"
	if len(allText) < len(expectedSubstring) {
		t.Errorf("Extracted text too short, got: %q", allText)
	}
}

// TestStoryAttributes tests that all attributes are correctly parsed.
func TestStoryAttributes(t *testing.T) {
	// Read test file
	data := testutil.ReadTestData(t, "story_u1d8.xml")

	// Parse story
	story, err := ParseStory(data)
	if err != nil {
		t.Fatalf("ParseStory failed: %v", err)
	}

	// Verify DOMVersion
	if story.DOMVersion != "20.4" {
		t.Errorf("DOMVersion: want 20.4, got %s", story.DOMVersion)
	}

	// Verify Story element attributes
	se := story.StoryElement
	if se.Self != "u1d8" {
		t.Errorf("Self: want u1d8, got %s", se.Self)
	}

	if se.UserText != "true" {
		t.Errorf("UserText: want true, got %s", se.UserText)
	}

	if se.IsEndnoteStory != "false" {
		t.Errorf("IsEndnoteStory: want false, got %s", se.IsEndnoteStory)
	}

	if se.TrackChanges != "false" {
		t.Errorf("TrackChanges: want false, got %s", se.TrackChanges)
	}
}

// TestStoryPreference tests that StoryPreference is correctly parsed.
func TestStoryPreference(t *testing.T) {
	// Read test file
	data := testutil.ReadTestData(t, "story_u1d8.xml")

	// Parse story
	story, err := ParseStory(data)
	if err != nil {
		t.Fatalf("ParseStory failed: %v", err)
	}

	// Verify StoryPreference exists
	sp := story.StoryElement.StoryPreference
	if sp == nil {
		t.Fatal("StoryPreference is nil")
	}

	// Verify attributes
	if sp.OpticalMarginAlignment != "false" {
		t.Errorf("OpticalMarginAlignment: want false, got %s", sp.OpticalMarginAlignment)
	}

	if sp.OpticalMarginSize != "12" {
		t.Errorf("OpticalMarginSize: want 12, got %s", sp.OpticalMarginSize)
	}

	if sp.FrameType != "TextFrameType" {
		t.Errorf("FrameType: want TextFrameType, got %s", sp.FrameType)
	}

	if sp.StoryOrientation != "Horizontal" {
		t.Errorf("StoryOrientation: want Horizontal, got %s", sp.StoryOrientation)
	}

	if sp.StoryDirection != "LeftToRightDirection" {
		t.Errorf("StoryDirection: want LeftToRightDirection, got %s", sp.StoryDirection)
	}
}

// TestParagraphStyleRanges tests parsing of paragraph style ranges.
func TestParagraphStyleRanges(t *testing.T) {
	// Read test file
	data := testutil.ReadTestData(t, "story_u1d8.xml")

	// Parse story
	story, err := ParseStory(data)
	if err != nil {
		t.Fatalf("ParseStory failed: %v", err)
	}

	// Verify we have paragraph style ranges
	ranges := story.StoryElement.ParagraphStyleRanges
	if len(ranges) == 0 {
		t.Fatal("No ParagraphStyleRanges found")
	}

	// Check first range
	psr := ranges[0]
	if psr.AppliedParagraphStyle == "" {
		t.Error("AppliedParagraphStyle is empty")
	}

	// Check that it has character style ranges
	if len(psr.CharacterStyleRanges) == 0 {
		t.Error("No CharacterStyleRanges in first ParagraphStyleRange")
	}
}

// TestCharacterStyleRanges tests parsing of character style ranges.
func TestCharacterStyleRanges(t *testing.T) {
	// Read test file
	data := testutil.ReadTestData(t, "story_u1d8.xml")

	// Parse story
	story, err := ParseStory(data)
	if err != nil {
		t.Fatalf("ParseStory failed: %v", err)
	}

	// Get first character style range
	if len(story.StoryElement.ParagraphStyleRanges) == 0 {
		t.Fatal("No ParagraphStyleRanges")
	}

	csrs := story.StoryElement.ParagraphStyleRanges[0].CharacterStyleRanges
	if len(csrs) == 0 {
		t.Fatal("No CharacterStyleRanges")
	}

	// Check first range
	csr := csrs[0]
	if csr.AppliedCharacterStyle == "" {
		t.Error("AppliedCharacterStyle is empty")
	}

	// Check content
	contents := csr.GetContent()
	if len(contents) == 0 {
		t.Error("No Content in CharacterStyleRange")
	}

	if contents[0].Text == "" {
		t.Error("Content text is empty")
	}
}

// TestStoryNamespacePrefix verifies that the idPkg prefix is used correctly.
func TestStoryNamespacePrefix(t *testing.T) {
	// Read test file
	data := testutil.ReadTestData(t, "story_u1d8.xml")

	// Parse story
	story, err := ParseStory(data)
	if err != nil {
		t.Fatalf("ParseStory failed: %v", err)
	}

	// Marshal back
	regenerated, err := MarshalStory(story)
	if err != nil {
		t.Fatalf("MarshalStory failed: %v", err)
	}

	// Check for idPkg prefix
	regeneratedStr := string(regenerated)
	if !strings.Contains(regeneratedStr, "<idPkg:Story") {
		t.Error("Output missing <idPkg:Story> tag")
	}
	if !strings.Contains(regeneratedStr, "</idPkg:Story>") {
		t.Error("Output missing </idPkg:Story> tag")
	}
	if !strings.Contains(regeneratedStr, `xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging"`) {
		t.Error("Output missing xmlns:idPkg namespace declaration")
	}
}

// TestStoryOutputDebug helps debug the marshaled output.
func TestStoryOutputDebug(t *testing.T) {
	// Read test file
	data := testutil.ReadTestData(t, "story_u1d8.xml")

	t.Logf("Original XML:\n%s", string(data))

	// Parse story
	story, err := ParseStory(data)
	if err != nil {
		t.Fatalf("ParseStory failed: %v", err)
	}

	// Marshal back
	regenerated, err := MarshalStory(story)
	if err != nil {
		t.Fatalf("MarshalStory failed: %v", err)
	}

	t.Logf("Regenerated XML:\n%s", string(regenerated))
}

// TestNewCharacterStyleRange tests the CharacterStyleRange constructor.
func TestNewCharacterStyleRange(t *testing.T) {
	tests := []struct {
		name         string
		appliedStyle string
		contents     []Content
		wantStyle    string
		wantChildren int
	}{
		{
			name:         "empty style uses default",
			appliedStyle: "",
			contents: []Content{
				{Text: "Hello"},
			},
			wantStyle:    "CharacterStyle/$ID/[No character style]",
			wantChildren: 2, // Content + Br
		},
		{
			name:         "custom style",
			appliedStyle: "CharacterStyle/MyStyle",
			contents: []Content{
				{Text: "Hello"},
				{Text: "World"},
			},
			wantStyle:    "CharacterStyle/MyStyle",
			wantChildren: 4, // 2 * (Content + Br)
		},
		{
			name:         "empty contents",
			appliedStyle: "CharacterStyle/Test",
			contents:     []Content{},
			wantStyle:    "CharacterStyle/Test",
			wantChildren: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			csr := NewCharacterStyleRange(tt.appliedStyle, tt.contents)

			if csr.AppliedCharacterStyle != tt.wantStyle {
				t.Errorf("AppliedCharacterStyle = %q, want %q", csr.AppliedCharacterStyle, tt.wantStyle)
			}

			if len(csr.Children) != tt.wantChildren {
				t.Errorf("Children count = %d, want %d", len(csr.Children), tt.wantChildren)
			}

			// Verify GetContent returns the original texts
			gotContents := csr.GetContent()
			if len(gotContents) != len(tt.contents) {
				t.Errorf("GetContent count = %d, want %d", len(gotContents), len(tt.contents))
			}
			for i, got := range gotContents {
				if got.Text != tt.contents[i].Text {
					t.Errorf("Content[%d].Text = %q, want %q", i, got.Text, tt.contents[i].Text)
				}
			}
		})
	}
}

// TestCharacterStyleRange_SetContent tests the SetContent method.
func TestCharacterStyleRange_SetContent(t *testing.T) {
	csr := CharacterStyleRange{
		AppliedCharacterStyle: "CharacterStyle/Test",
	}

	// Set initial content
	csr.SetContent([]Content{
		{Text: "First"},
		{Text: "Second"},
	})

	// Verify structure: 2 Content + 2 Br = 4 children
	if len(csr.Children) != 4 {
		t.Errorf("After SetContent: Children count = %d, want 4", len(csr.Children))
	}

	// Verify GetContent
	contents := csr.GetContent()
	if len(contents) != 2 {
		t.Fatalf("GetContent count = %d, want 2", len(contents))
	}
	if contents[0].Text != "First" {
		t.Errorf("contents[0].Text = %q, want %q", contents[0].Text, "First")
	}
	if contents[1].Text != "Second" {
		t.Errorf("contents[1].Text = %q, want %q", contents[1].Text, "Second")
	}

	// SetContent again should replace
	csr.SetContent([]Content{
		{Text: "Replaced"},
	})

	if len(csr.Children) != 2 {
		t.Errorf("After second SetContent: Children count = %d, want 2", len(csr.Children))
	}

	contents = csr.GetContent()
	if len(contents) != 1 || contents[0].Text != "Replaced" {
		t.Errorf("After replace: got %v, want [{Replaced}]", contents)
	}
}

// TestCharacterStyleRange_AddContent tests the AddContent method.
func TestCharacterStyleRange_AddContent(t *testing.T) {
	csr := CharacterStyleRange{
		AppliedCharacterStyle: "CharacterStyle/Test",
	}

	// Add first content
	csr.AddContent("First")
	if len(csr.Children) != 2 { // Content + Br
		t.Errorf("After first AddContent: Children count = %d, want 2", len(csr.Children))
	}

	// Add second content
	csr.AddContent("Second")
	if len(csr.Children) != 4 { // 2 * (Content + Br)
		t.Errorf("After second AddContent: Children count = %d, want 4", len(csr.Children))
	}

	// Verify GetContent
	contents := csr.GetContent()
	if len(contents) != 2 {
		t.Fatalf("GetContent count = %d, want 2", len(contents))
	}
	if contents[0].Text != "First" {
		t.Errorf("contents[0].Text = %q, want %q", contents[0].Text, "First")
	}
	if contents[1].Text != "Second" {
		t.Errorf("contents[1].Text = %q, want %q", contents[1].Text, "Second")
	}
}

// TestCharacterStyleRange_GetContent_Empty tests GetContent on empty range.
func TestCharacterStyleRange_GetContent_Empty(t *testing.T) {
	csr := CharacterStyleRange{}
	contents := csr.GetContent()
	if len(contents) != 0 {
		t.Errorf("Empty CSR GetContent = %v, want empty", contents)
	}
}

// TestCharacterStyleRange_GetContent_MixedChildren tests GetContent with mixed children.
func TestCharacterStyleRange_GetContent_MixedChildren(t *testing.T) {
	csr := CharacterStyleRange{
		Children: []CharacterChild{
			{Content: &Content{Text: "A"}},
			{Br: &Br{}},
			{Content: &Content{Text: "B"}},
			{Br: &Br{}},
			{Br: &Br{}}, // Extra Br should be ignored
		},
	}

	contents := csr.GetContent()
	if len(contents) != 2 {
		t.Fatalf("GetContent count = %d, want 2", len(contents))
	}
	if contents[0].Text != "A" || contents[1].Text != "B" {
		t.Errorf("contents = %v, want [A, B]", contents)
	}
}
