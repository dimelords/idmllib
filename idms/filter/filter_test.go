package filter

import (
	"testing"

	"github.com/dimelords/idmllib/types"
)

func TestCharacterStyleGroupWrapper_FilterGroup(t *testing.T) {
	group := &types.CharacterStyleGroup{
		Self: "CharacterStyleGroup/test",
		Name: "Test Group",
		CharacterStyles: []types.CharacterStyle{
			{Self: "CharacterStyle/used", Name: "Used Style"},
			{Self: "CharacterStyle/unused", Name: "Unused Style"},
			{Self: "CharacterStyle/used2", Name: "Used Style 2"},
		},
	}

	usedStyles := map[string]bool{
		"CharacterStyle/used":  true,
		"CharacterStyle/used2": true,
	}

	wrapper := &CharacterStyleGroupWrapper{Group: group}
	hasStyles := wrapper.FilterGroup(usedStyles)

	if !hasStyles {
		t.Error("Expected FilterGroup to return true when styles are present")
	}

	if len(group.CharacterStyles) != 2 {
		t.Errorf("Expected 2 styles after filtering, got %d", len(group.CharacterStyles))
	}

	// Verify correct styles are kept
	foundUsed := false
	foundUsed2 := false
	for _, style := range group.CharacterStyles {
		if style.Self == "CharacterStyle/used" {
			foundUsed = true
		}
		if style.Self == "CharacterStyle/used2" {
			foundUsed2 = true
		}
	}

	if !foundUsed || !foundUsed2 {
		t.Error("Expected both used styles to be kept after filtering")
	}
}

func TestParagraphStyleGroupWrapper_FilterGroup(t *testing.T) {
	// Create a group with styles and sub-groups
	group := &types.ParagraphStyleGroup{
		Self: "ParagraphStyleGroup/test",
		Name: "Test Group",
		ParagraphStyles: []types.ParagraphStyle{
			{Self: "ParagraphStyle/used", Name: "Used Style"},
			{Self: "ParagraphStyle/unused", Name: "Unused Style"},
		},
		SubGroups: []types.ParagraphStyleGroup{
			{
				Self: "ParagraphStyleGroup/sub",
				Name: "Sub Group",
				ParagraphStyles: []types.ParagraphStyle{
					{Self: "ParagraphStyle/subused", Name: "Sub Used Style"},
					{Self: "ParagraphStyle/subunused", Name: "Sub Unused Style"},
				},
			},
		},
	}

	usedStyles := map[string]bool{
		"ParagraphStyle/used":    true,
		"ParagraphStyle/subused": true,
	}

	wrapper := &ParagraphStyleGroupWrapper{Group: group}
	hasStyles := wrapper.FilterGroup(usedStyles)

	if !hasStyles {
		t.Error("Expected FilterGroup to return true when styles are present")
	}

	if len(group.ParagraphStyles) != 1 {
		t.Errorf("Expected 1 style after filtering, got %d", len(group.ParagraphStyles))
	}

	if len(group.SubGroups) != 1 {
		t.Errorf("Expected 1 sub-group after filtering, got %d", len(group.SubGroups))
	}

	if len(group.SubGroups[0].ParagraphStyles) != 1 {
		t.Errorf("Expected 1 style in sub-group, got %d", len(group.SubGroups[0].ParagraphStyles))
	}
}

func TestObjectStyleGroupWrapper_FilterGroup(t *testing.T) {
	group := &types.ObjectStyleGroup{
		Self: "ObjectStyleGroup/test",
		Name: "Test Group",
		ObjectStyles: []types.ObjectStyle{
			{Self: "ObjectStyle/used", Name: "Used Style"},
			{Self: "ObjectStyle/unused", Name: "Unused Style"},
		},
		SubGroups: []types.ObjectStyleGroup{
			{
				Self: "ObjectStyleGroup/sub",
				Name: "Sub Group",
				ObjectStyles: []types.ObjectStyle{
					{Self: "ObjectStyle/subused", Name: "Sub Used Style"},
				},
			},
		},
	}

	usedStyles := map[string]bool{
		"ObjectStyle/used":    true,
		"ObjectStyle/subused": true,
	}

	wrapper := &ObjectStyleGroupWrapper{Group: group}
	hasStyles := wrapper.FilterGroup(usedStyles)

	if !hasStyles {
		t.Error("Expected FilterGroup to return true when styles are present")
	}

	if len(group.ObjectStyles) != 1 {
		t.Errorf("Expected 1 style after filtering, got %d", len(group.ObjectStyles))
	}

	if len(group.SubGroups) != 1 {
		t.Errorf("Expected 1 sub-group after filtering, got %d", len(group.SubGroups))
	}
}

func TestFilterGroup_EmptyResult(t *testing.T) {
	group := &types.CharacterStyleGroup{
		Self: "CharacterStyleGroup/test",
		Name: "Test Group",
		CharacterStyles: []types.CharacterStyle{
			{Self: "CharacterStyle/unused", Name: "Unused Style"},
		},
	}

	usedStyles := map[string]bool{
		"CharacterStyle/other": true,
	}

	wrapper := &CharacterStyleGroupWrapper{Group: group}
	hasStyles := wrapper.FilterGroup(usedStyles)

	if hasStyles {
		t.Error("Expected FilterGroup to return false when no styles match")
	}

	if len(group.CharacterStyles) != 0 {
		t.Errorf("Expected 0 styles after filtering, got %d", len(group.CharacterStyles))
	}
}
