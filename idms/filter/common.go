package filter

import "github.com/dimelords/idmllib/types"

// StyleGroupFilter provides a generic interface for filtering style groups
type StyleGroupFilter interface {
	// FilterGroup filters a style group to only include used styles
	FilterGroup(usedStyles map[string]bool) bool
}

// CharacterStyleGroupWrapper wraps CharacterStyleGroup for generic filtering
type CharacterStyleGroupWrapper struct {
	Group *types.CharacterStyleGroup
}

// FilterGroup filters character styles in the group based on used styles
func (w *CharacterStyleGroupWrapper) FilterGroup(usedStyles map[string]bool) bool {
	var filteredStyles []types.CharacterStyle
	for _, style := range w.Group.CharacterStyles {
		if usedStyles[style.Self] {
			filteredStyles = append(filteredStyles, style)
		}
	}

	w.Group.CharacterStyles = filteredStyles
	return len(filteredStyles) > 0
}

// ParagraphStyleGroupWrapper wraps ParagraphStyleGroup for generic filtering
type ParagraphStyleGroupWrapper struct {
	Group *types.ParagraphStyleGroup
}

// FilterGroup filters paragraph styles in the group based on used styles
func (w *ParagraphStyleGroupWrapper) FilterGroup(usedStyles map[string]bool) bool {
	// Filter styles in this group
	var filteredStyles []types.ParagraphStyle
	for _, style := range w.Group.ParagraphStyles {
		if usedStyles[style.Self] {
			filteredStyles = append(filteredStyles, style)
		}
	}

	// Recursively filter sub-groups
	var filteredSubGroups []types.ParagraphStyleGroup
	for i := range w.Group.SubGroups {
		subWrapper := &ParagraphStyleGroupWrapper{Group: &w.Group.SubGroups[i]}
		if subWrapper.FilterGroup(usedStyles) {
			filteredSubGroups = append(filteredSubGroups, w.Group.SubGroups[i])
		}
	}

	w.Group.ParagraphStyles = filteredStyles
	w.Group.SubGroups = filteredSubGroups

	return len(filteredStyles) > 0 || len(filteredSubGroups) > 0
}

// ObjectStyleGroupWrapper wraps ObjectStyleGroup for generic filtering
type ObjectStyleGroupWrapper struct {
	Group *types.ObjectStyleGroup
}

// FilterGroup filters object styles in the group based on used styles
func (w *ObjectStyleGroupWrapper) FilterGroup(usedStyles map[string]bool) bool {
	// Filter styles in this group
	var filteredStyles []types.ObjectStyle
	for _, style := range w.Group.ObjectStyles {
		if usedStyles[style.Self] {
			filteredStyles = append(filteredStyles, style)
		}
	}

	// Recursively filter sub-groups
	var filteredSubGroups []types.ObjectStyleGroup
	for i := range w.Group.SubGroups {
		subWrapper := &ObjectStyleGroupWrapper{Group: &w.Group.SubGroups[i]}
		if subWrapper.FilterGroup(usedStyles) {
			filteredSubGroups = append(filteredSubGroups, w.Group.SubGroups[i])
		}
	}

	w.Group.ObjectStyles = filteredStyles
	w.Group.SubGroups = filteredSubGroups

	return len(filteredStyles) > 0 || len(filteredSubGroups) > 0
}

// filterStyleGroups is a generic function that filters any style group type
// It eliminates code duplication between Character, Paragraph, and Object style filtering
func filterStyleGroups[G any, W StyleGroupFilter](groups []G, usedStyles map[string]bool, wrapFn func(*G) W) []G {
	var filtered []G
	for i := range groups {
		wrapper := wrapFn(&groups[i])
		if wrapper.FilterGroup(usedStyles) {
			filtered = append(filtered, groups[i])
		}
	}
	return filtered
}
