package filter

import (
	"log/slog"

	"github.com/dimelords/idmllib/types"
)

// StyleWithDependencies represents a style that can have BasedOn/NextStyle references
type StyleWithDependencies interface {
	GetSelf() string
	GetBasedOn() string
	GetNextStyle() string
	IsSystemStyle(ref string) bool
}

// resolveStyleDependencies recursively adds base styles (BasedOn) and related styles (NextStyle)
// for all used styles to ensure the complete style hierarchy is included
func resolveStyleDependencies(idms *types.IDMS, used *UsedStyles) {
	// Keep resolving until no new dependencies are found
	maxIterations := 50 // Prevent infinite loops
	for iteration := 0; iteration < maxIterations; iteration++ {
		foundNew := false

		// Resolve CharacterStyle dependencies
		if idms.RootCharacterStyleGroup != nil {
			if resolveCharacterStyleDependencies(idms.RootCharacterStyleGroup, used) {
				foundNew = true
			}
		}

		// Resolve ParagraphStyle dependencies
		if idms.RootParagraphStyleGroup != nil {
			if resolveParagraphStyleDependencies(idms.RootParagraphStyleGroup, used) {
				foundNew = true
			}
		}

		// Resolve ObjectStyle dependencies
		if idms.RootObjectStyleGroup != nil {
			if resolveObjectStyleDependencies(idms.RootObjectStyleGroup, used) {
				foundNew = true
			}
		}

		// If no new dependencies were found, we're done
		if !foundNew {
			slog.Debug("Style dependency resolution completed", "iterations", iteration+1)
			break
		}
	}
}

// resolveGenericStyleDependencies resolves dependencies for any style type
func resolveGenericStyleDependencies[S StyleWithDependencies](
	styles []S,
	used map[string]bool,
	styleType string,
) bool {
	foundNew := false
	for _, style := range styles {
		// Only check styles that are already marked as used
		if !used[style.GetSelf()] {
			continue
		}

		// Check BasedOn
		if basedOn := style.GetBasedOn(); basedOn != "" && !style.IsSystemStyle(basedOn) {
			if !used[basedOn] {
				used[basedOn] = true
				foundNew = true
				slog.Debug("Added base "+styleType, "derived", style.GetSelf(), "basedOn", basedOn)
			}
		}

		// Check NextStyle (for paragraph styles)
		if nextStyle := style.GetNextStyle(); nextStyle != "" && !style.IsSystemStyle(nextStyle) {
			if !used[nextStyle] {
				used[nextStyle] = true
				foundNew = true
				slog.Debug("Added NextStyle "+styleType, "current", style.GetSelf(), "nextStyle", nextStyle)
			}
		}
	}
	return foundNew
}

// resolveCharacterStyleDependencies finds and marks BasedOn styles for character styles
func resolveCharacterStyleDependencies(root *types.RootCharacterStyleGroup, used *UsedStyles) bool {
	foundNew := false
	for _, group := range root.CharacterStyleGroups {
		if resolveGenericStyleDependencies(group.CharacterStyles, used.CharacterStyles, "CharacterStyle") {
			foundNew = true
		}
	}
	return foundNew
}

// resolveParagraphStyleDependencies finds and marks BasedOn and NextStyle for paragraph styles
func resolveParagraphStyleDependencies(root *types.RootParagraphStyleGroup, used *UsedStyles) bool {
	foundNew := false
	for _, group := range root.ParagraphStyleGroups {
		if resolveParagraphStyleGroupDependencies(&group, used) {
			foundNew = true
		}
	}
	return foundNew
}

func resolveParagraphStyleGroupDependencies(group *types.ParagraphStyleGroup, used *UsedStyles) bool {
	// Resolve dependencies for styles in this group
	foundNew := resolveGenericStyleDependencies(group.ParagraphStyles, used.ParagraphStyles, "ParagraphStyle")

	// Recursively check sub-groups
	for _, subGroup := range group.SubGroups {
		if resolveParagraphStyleGroupDependencies(&subGroup, used) {
			foundNew = true
		}
	}

	return foundNew
}

// resolveObjectStyleDependencies finds and marks BasedOn styles for object styles
func resolveObjectStyleDependencies(root *types.RootObjectStyleGroup, used *UsedStyles) bool {
	foundNew := false
	for _, group := range root.ObjectStyleGroups {
		if resolveObjectStyleGroupDependencies(&group, used) {
			foundNew = true
		}
	}
	return foundNew
}

func resolveObjectStyleGroupDependencies(group *types.ObjectStyleGroup, used *UsedStyles) bool {
	// Resolve dependencies for styles in this group
	foundNew := resolveGenericStyleDependencies(group.ObjectStyles, used.ObjectStyles, "ObjectStyle")

	// Recursively check sub-groups
	for _, subGroup := range group.SubGroups {
		if resolveObjectStyleGroupDependencies(&subGroup, used) {
			foundNew = true
		}
	}

	return foundNew
}
