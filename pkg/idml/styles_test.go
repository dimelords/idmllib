package idml

import (
	"testing"
)

// TestParseStylesForHierarchy_ParsesHierarchy tests parsing style hierarchy information
func TestParseStylesForHierarchy_ParsesHierarchy(t *testing.T) {
	// Load a test IDML to get the Styles.xml file
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read IDML: %v", err)
	}

	// Get the Styles resource
	stylesResource, err := pkg.Resource("Resources/Styles.xml")
	if err != nil {
		t.Fatalf("Failed to get Styles.xml: %v", err)
	}

	// Get the raw bytes
	stylesData := stylesResource.RawContent

	// Parse style hierarchy
	styleInfos, err := ParseStylesForHierarchy(stylesData)
	if err != nil {
		t.Fatalf("ParseStylesForHierarchy() error: %v", err)
	}

	if len(styleInfos) == 0 {
		t.Fatal("Expected to find some styles")
	}

	t.Logf("✅ Found %d styles", len(styleInfos))

	// Build a map of style IDs to their info
	styleMap := make(map[string]StyleInfo)
	for _, info := range styleInfos {
		styleMap[info.Self] = info
	}

	// Verify we found built-in styles
	builtinStyles := []string{
		"CharacterStyle/$ID/[No character style]",
		"ParagraphStyle/$ID/[No paragraph style]",
		"ObjectStyle/$ID/[None]",
	}

	for _, styleID := range builtinStyles {
		if _, found := styleMap[styleID]; !found {
			t.Errorf("Expected to find built-in style: %s", styleID)
		}
	}

	// Count styles by type based on Self prefix
	var charCount, paraCount, objCount int
	for _, info := range styleInfos {
		if len(info.Self) >= 14 && info.Self[:14] == "CharacterStyle" {
			charCount++
		} else if len(info.Self) >= 14 && info.Self[:14] == "ParagraphStyle" {
			paraCount++
		} else if len(info.Self) >= 11 && info.Self[:11] == "ObjectStyle" {
			objCount++
		}
	}

	t.Logf("   Character styles: %d", charCount)
	t.Logf("   Paragraph styles: %d", paraCount)
	t.Logf("   Object styles: %d", objCount)

	// Verify some styles have BasedOn relationships
	stylesWithParents := 0
	for _, info := range styleInfos {
		if info.BasedOn != "" {
			stylesWithParents++
		}
	}

	if stylesWithParents == 0 {
		t.Error("Expected to find some styles with BasedOn relationships")
	}

	t.Logf("   Styles with parents: %d", stylesWithParents)

	// Find and verify a specific style chain
	// Look for any character style that's based on the default
	for _, info := range styleInfos {
		if len(info.Self) >= 14 && info.Self[:14] == "CharacterStyle" &&
			info.BasedOn == "$ID/[No character style]" {
			t.Logf("   Example character style: '%s' based on '%s'", info.Self, info.BasedOn)
			break
		}
	}

	// Find and verify a paragraph style chain
	for _, info := range styleInfos {
		if len(info.Self) >= 14 && info.Self[:14] == "ParagraphStyle" &&
			info.BasedOn != "" && info.BasedOn != "$ID/[No paragraph style]" {
			// This is a style based on a non-default style
			t.Logf("   Example paragraph style chain: '%s' → '%s'", info.Self, info.BasedOn)

			// Check if the parent exists
			if parent, found := styleMap[info.BasedOn]; found {
				t.Logf("      Parent exists: '%s' (based on: '%s')", parent.Self, parent.BasedOn)
			}
			break
		}
	}
}

// TestResolveStyleHierarchies_ResolvesHierarchies tests the DependencyTracker's style hierarchy resolution
func TestResolveStyleHierarchies_ResolvesHierarchies(t *testing.T) {
	// This test is in the analysis package
	// We're importing it here to verify the integration
	t.Skip("Implemented in pkg/analysis/tracker_test.go")
}
