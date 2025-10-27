package filter

import (
	"encoding/xml"
	"regexp"
	"strings"

	"github.com/dimelords/idmllib/types"
)

// UsedStyles holds all style Self references used in the story
type UsedStyles struct {
	CharacterStyles map[string]bool
	ParagraphStyles map[string]bool
	ObjectStyles    map[string]bool
}

// extractUsedStyles analyzes the story and spreads to find all referenced styles
func extractUsedStyles(ctx *Context) *UsedStyles {
	used := &UsedStyles{
		CharacterStyles: make(map[string]bool),
		ParagraphStyles: make(map[string]bool),
		ObjectStyles:    make(map[string]bool),
	}

	storyXML, err := xml.Marshal(ctx.Story)
	if err != nil {
		return used
	}

	// Extract CharacterStyle references: AppliedCharacterStyle="..."
	charStylePattern := regexp.MustCompile(`AppliedCharacterStyle="([^"]+)"`)
	for _, match := range charStylePattern.FindAllStringSubmatch(string(storyXML), -1) {
		if len(match) > 1 && match[1] != "n" && match[1] != "$ID/[No character style]" {
			used.CharacterStyles[match[1]] = true
		}
	}

	// Extract ParagraphStyle references: AppliedParagraphStyle="..."
	paraStylePattern := regexp.MustCompile(`AppliedParagraphStyle="([^"]+)"`)
	for _, match := range paraStylePattern.FindAllStringSubmatch(string(storyXML), -1) {
		if len(match) > 1 && match[1] != "n" && match[1] != "$ID/[No paragraph style]" {
			used.ParagraphStyles[match[1]] = true
		}
	}

	// Extract ObjectStyle references from the pre-loaded spreads
	for _, spread := range ctx.Spreads {
		for _, textFrame := range spread.TextFrames {
			// Only process TextFrames that reference our story
			if textFrame.ParentStory == ctx.Story.Self {
				style := textFrame.AppliedObjectStyle
				if style != "" && style != "n" && style != "$ID/[None]" {
					used.ObjectStyles[style] = true
				}
			}
		}
	}

	return used
}

// filterCharacterStyleGroups filters character style groups to only include used styles
func filterCharacterStyleGroups(root *types.RootCharacterStyleGroup, usedStyles map[string]bool) {
	if root == nil {
		return
	}
	root.CharacterStyleGroups = filterStyleGroups(
		root.CharacterStyleGroups,
		usedStyles,
		func(g *types.CharacterStyleGroup) StyleGroupFilter {
			return &CharacterStyleGroupWrapper{Group: g}
		})
}

// filterParagraphStyleGroups filters paragraph style groups to only include used styles
func filterParagraphStyleGroups(root *types.RootParagraphStyleGroup, usedStyles map[string]bool) {
	if root == nil {
		return
	}
	root.ParagraphStyleGroups = filterStyleGroups(
		root.ParagraphStyleGroups,
		usedStyles,
		func(g *types.ParagraphStyleGroup) StyleGroupFilter {
			return &ParagraphStyleGroupWrapper{Group: g}
		})
}

// filterObjectStyleGroups filters object style groups to only include used styles
func filterObjectStyleGroups(root *types.RootObjectStyleGroup, usedStyles map[string]bool) {
	if root == nil {
		return
	}
	root.ObjectStyleGroups = filterStyleGroups(
		root.ObjectStyleGroups,
		usedStyles,
		func(g *types.ObjectStyleGroup) StyleGroupFilter {
			return &ObjectStyleGroupWrapper{Group: g}
		})
}

// extractColorsFromCharacterStyleGroup extracts color references from character styles
func extractColorsFromCharacterStyleGroup(group *types.CharacterStyleGroup, pattern *regexp.Regexp, used *UsedColors) {
	for _, style := range group.CharacterStyles {
		// CharacterStyle doesn't have FontStyle, only Properties
		var styleXML string
		if style.Properties != nil {
			styleXML = style.Properties.InnerXML
		}
		for _, match := range pattern.FindAllStringSubmatch(styleXML, -1) {
			if len(match) > 1 && match[1] != "n" {
				if strings.HasPrefix(match[1], "Color/") {
					used.Colors[match[1]] = true
				} else if strings.HasPrefix(match[1], "Swatch/") {
					used.Swatches[match[1]] = true
				}
			}
		}
	}
}

// extractColorsFromParagraphStyleGroup extracts color references from paragraph styles
func extractColorsFromParagraphStyleGroup(group *types.ParagraphStyleGroup, pattern *regexp.Regexp, used *UsedColors) {
	for _, style := range group.ParagraphStyles {
		// Build string to search - check for nil Properties
		styleXML := style.FontStyle
		if style.Properties != nil {
			styleXML += " " + style.Properties.InnerXML
		}
		for _, match := range pattern.FindAllStringSubmatch(styleXML, -1) {
			if len(match) > 1 && match[1] != "n" {
				if strings.HasPrefix(match[1], "Color/") {
					used.Colors[match[1]] = true
				} else if strings.HasPrefix(match[1], "Swatch/") {
					used.Swatches[match[1]] = true
				}
			}
		}
	}

	// Recursively check sub-groups
	for _, subGroup := range group.SubGroups {
		extractColorsFromParagraphStyleGroup(&subGroup, pattern, used)
	}
}

// extractColorsFromObjectStyleGroup extracts color references from object styles
func extractColorsFromObjectStyleGroup(group *types.ObjectStyleGroup, pattern *regexp.Regexp, used *UsedColors) {
	for _, style := range group.ObjectStyles {
		// Check all color-related attributes
		styleAttrs := style.FillColor + " " + style.StrokeColor + " " + style.GapColor + " " + style.InnerXML
		for _, match := range pattern.FindAllStringSubmatch(styleAttrs, -1) {
			if len(match) > 1 && match[1] != "n" {
				if strings.HasPrefix(match[1], "Color/") {
					used.Colors[match[1]] = true
				} else if strings.HasPrefix(match[1], "Swatch/") {
					used.Swatches[match[1]] = true
				}
			}
		}
	}

	// Recursively check sub-groups
	for _, subGroup := range group.SubGroups {
		extractColorsFromObjectStyleGroup(&subGroup, pattern, used)
	}
}
