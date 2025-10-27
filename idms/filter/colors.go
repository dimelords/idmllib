package filter

import (
	"regexp"

	"github.com/dimelords/idmllib/types"
)

// UsedColors holds all color and swatch Self references used in the story
type UsedColors struct {
	Colors   map[string]bool
	Swatches map[string]bool
}

// extractUsedColors analyzes styles to find all referenced colors and swatches
func extractUsedColors(idms *types.IDMS) *UsedColors {
	used := &UsedColors{
		Colors:   make(map[string]bool),
		Swatches: make(map[string]bool),
	}

	// Extract colors from used styles
	extractColorsFromStyles(idms, used)

	return used
}

// extractColorsFromStyles finds color references in the filtered styles
func extractColorsFromStyles(idms *types.IDMS, used *UsedColors) {
	colorPattern := regexp.MustCompile(`(?:FillColor|StrokeColor|GapColor|ColumnRuleStrokeColor)="((?:Color|Swatch)/[^"]+)"`)

	// Check character styles
	if idms.RootCharacterStyleGroup != nil {
		for _, group := range idms.RootCharacterStyleGroup.CharacterStyleGroups {
			extractColorsFromCharacterStyleGroup(&group, colorPattern, used)
		}
	}

	// Check paragraph styles
	if idms.RootParagraphStyleGroup != nil {
		for _, group := range idms.RootParagraphStyleGroup.ParagraphStyleGroups {
			extractColorsFromParagraphStyleGroup(&group, colorPattern, used)
		}
	}

	// Check object styles
	if idms.RootObjectStyleGroup != nil {
		for _, group := range idms.RootObjectStyleGroup.ObjectStyleGroups {
			extractColorsFromObjectStyleGroup(&group, colorPattern, used)
		}
	}
}

// filterColorsAndSwatches removes unused colors and swatches from the IDMS document
func filterColorsAndSwatches(idms *types.IDMS, used *UsedColors) {
	// System colors and swatches that should always be included
	// Based on InDesign native snippet behavior:
	// - Swatch/None is always included
	// - Color/Black is always included
	// - Color/Paper and Color/Registration are only included if explicitly used
	alwaysInclude := map[string]bool{
		"Swatch/None": true,
		"Color/Black": true,
	}

	// Filter colors - keep used colors and always-include colors
	var filteredColors []types.Color
	for _, color := range idms.Colors {
		if used.Colors[color.Self] || alwaysInclude[color.Self] {
			filteredColors = append(filteredColors, color)
			// Mark always-include items as used for ColorGroup filtering
			if alwaysInclude[color.Self] {
				used.Colors[color.Self] = true
			}
		}
	}
	idms.Colors = filteredColors

	// Filter swatches - keep used swatches and always-include swatches
	var filteredSwatches []types.Swatch
	for _, swatch := range idms.Swatches {
		if used.Swatches[swatch.Self] || alwaysInclude[swatch.Self] {
			filteredSwatches = append(filteredSwatches, swatch)
			// Mark always-include items as used for ColorGroup filtering
			if alwaysInclude[swatch.Self] {
				used.Swatches[swatch.Self] = true
			}
		}
	}
	idms.Swatches = filteredSwatches
}

// filterColorGroups removes ColorGroupSwatch entries that reference unused colors/swatches
func filterColorGroups(idms *types.IDMS, used *UsedColors) {
	for i := range idms.ColorGroups {
		var filteredSwatches []types.ColorGroupSwatch
		for _, cgSwatch := range idms.ColorGroups[i].ColorGroupSwatches {
			// Keep the swatch if the referenced color or swatch is used
			if used.Colors[cgSwatch.SwatchItemRef] || used.Swatches[cgSwatch.SwatchItemRef] {
				filteredSwatches = append(filteredSwatches, cgSwatch)
			}
		}
		idms.ColorGroups[i].ColorGroupSwatches = filteredSwatches
	}
}
