// Package filter provides style and color filtering functionality for IDMS documents.
package filter

import (
	"log/slog"

	"github.com/dimelords/idmllib/types"
)

// Context contains all data needed for filtering operations.
// This allows the filter package to be independent of file I/O operations.
type Context struct {
	IDMS    *types.IDMS
	Story   *types.Story
	Spreads []types.Spread // Pre-loaded spreads that contain this story
}

// UnusedStyles filters style groups to only include styles that are actually
// referenced in the story content and TextFrames.
func UnusedStyles(ctx *Context) error {
	// Extract all style references from the story and spreads
	usedStyles := extractUsedStyles(ctx)

	slog.Debug("Used styles extracted (before dependency resolution)",
		"characterStyles", len(usedStyles.CharacterStyles),
		"paragraphStyles", len(usedStyles.ParagraphStyles),
		"objectStyles", len(usedStyles.ObjectStyles))

	// Resolve style dependencies (BasedOn, NextStyle, etc.)
	resolveStyleDependencies(ctx.IDMS, usedStyles)

	slog.Debug("Used styles after dependency resolution",
		"characterStyles", len(usedStyles.CharacterStyles),
		"paragraphStyles", len(usedStyles.ParagraphStyles),
		"objectStyles", len(usedStyles.ObjectStyles))

	// Filter each style group
	if ctx.IDMS.RootCharacterStyleGroup != nil {
		filterCharacterStyleGroups(ctx.IDMS.RootCharacterStyleGroup, usedStyles.CharacterStyles)
	}

	if ctx.IDMS.RootParagraphStyleGroup != nil {
		filterParagraphStyleGroups(ctx.IDMS.RootParagraphStyleGroup, usedStyles.ParagraphStyles)
	}

	if ctx.IDMS.RootObjectStyleGroup != nil {
		filterObjectStyleGroups(ctx.IDMS.RootObjectStyleGroup, usedStyles.ObjectStyles)
	}

	// Extract and filter colors and swatches
	usedColors := extractUsedColors(ctx.IDMS)
	slog.Debug("Used colors extracted",
		"colors", len(usedColors.Colors),
		"swatches", len(usedColors.Swatches))

	filterColorsAndSwatches(ctx.IDMS, usedColors)
	filterColorGroups(ctx.IDMS, usedColors)

	return nil
}
