package idml

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dimelords/idmllib/v2/pkg/common"
	"github.com/dimelords/idmllib/v2/pkg/resources"
	"github.com/dimelords/idmllib/v2/pkg/spread"
	"github.com/dimelords/idmllib/v2/pkg/story"
)

func (rm *ResourceManager) FindOrphans() (*OrphanedResources, error) {
	result := &OrphanedResources{}

	// Step 1: Analyze the entire document to find what's actually used
	deps, err := rm.analyzeDependencies()
	if err != nil {
		return nil, common.WrapError("idml", "find orphans", fmt.Errorf("failed to analyze content: %w", err))
	}

	// Step 2: Find orphaned fonts
	if err := rm.findOrphanedFonts(deps, result); err != nil {
		return nil, common.WrapError("idml", "find orphaned fonts", err)
	}

	// Step 3: Find orphaned styles
	if err := rm.findOrphanedStyles(deps, result); err != nil {
		return nil, common.WrapError("idml", "find orphaned styles", err)
	}

	// Step 4: Find orphaned object styles
	if err := rm.findOrphanedObjectStyles(deps, result); err != nil {
		return nil, common.WrapError("idml", "find orphaned object styles", err)
	}

	// Step 5: Find orphaned colors
	if err := rm.findOrphanedColors(deps, result); err != nil {
		return nil, common.WrapError("idml", "find orphaned colors", err)
	}

	// Step 6: Find orphaned swatches
	if err := rm.findOrphanedSwatches(deps, result); err != nil {
		return nil, common.WrapError("idml", "find orphaned swatches", err)
	}

	// Step 7: Find orphaned layers
	if err := rm.findOrphanedLayers(deps, result); err != nil {
		return nil, common.WrapError("idml", "find orphaned layers", err)
	}

	return result, nil
}

// analyzeDependencies analyzes all stories and spreads in the document
// to build a complete dependency set.
func (rm *ResourceManager) analyzeDependencies() (*dependencySet, error) {
	deps := newDependencySet()

	// Analyze all stories to find style dependencies
	stories, err := rm.pkg.Stories()
	if err != nil {
		return nil, common.WrapError("idml", "analyze dependencies", fmt.Errorf("failed to get stories: %w", err))
	}

	for _, story := range stories {
		if err := rm.analyzeStory(story, deps); err != nil {
			return nil, common.WrapError("idml", "analyze dependencies", fmt.Errorf("failed to analyze story: %w", err))
		}
	}

	// Analyze all spreads to find object dependencies
	spreads, err := rm.pkg.Spreads()
	if err != nil {
		return nil, common.WrapError("idml", "analyze dependencies", fmt.Errorf("failed to get spreads: %w", err))
	}

	for _, spread := range spreads {
		if err := rm.analyzeSpread(spread, deps); err != nil {
			return nil, common.WrapError("idml", "analyze dependencies", fmt.Errorf("failed to analyze spread: %w", err))
		}
	}

	// Extract colors from used paragraph styles
	if err := rm.extractColorsFromParagraphStyles(deps); err != nil {
		return nil, common.WrapError("idml", "analyze dependencies", fmt.Errorf("failed to extract paragraph style colors: %w", err))
	}

	// Extract colors from used character styles
	if err := rm.extractColorsFromCharacterStyles(deps); err != nil {
		return nil, common.WrapError("idml", "analyze dependencies", fmt.Errorf("failed to extract character style colors: %w", err))
	}

	return deps, nil
}

// analyzeStory analyzes a story and tracks all style dependencies.
func (rm *ResourceManager) analyzeStory(st *story.Story, deps *dependencySet) error {
	// Analyze each paragraph style range
	for _, psr := range st.StoryElement.ParagraphStyleRanges {
		// Track the paragraph style
		if psr.AppliedParagraphStyle != "" {
			deps.paragraphStyles[psr.AppliedParagraphStyle] = true
		}

		// Colors from paragraph styles are extracted in analyzeDependencies()
		// after all styles are collected. See extractColorsFromParagraphStyles().

		// Analyze each character style range within the paragraph
		for _, csr := range psr.CharacterStyleRanges {
			// Track the character style
			if csr.AppliedCharacterStyle != "" {
				deps.characterStyles[csr.AppliedCharacterStyle] = true
			}

			// Colors from character styles are extracted in analyzeDependencies()
			// after all styles are collected. See extractColorsFromCharacterStyles().
		}
	}

	return nil
}

// analyzeSpread analyzes a spread and tracks all object dependencies.
func (rm *ResourceManager) analyzeSpread(sp *spread.Spread, deps *dependencySet) error {
	// Analyze text frames
	for _, tf := range sp.InnerSpread.TextFrames {
		// Track the applied object style
		if tf.AppliedObjectStyle != "" {
			deps.objectStyles[tf.AppliedObjectStyle] = true
		}

		// Track the layer
		if tf.ItemLayer != "" {
			deps.layers[tf.ItemLayer] = true
		}

		// TODO: Color tracking in text frame fills and strokes.
		// Text frames get their colors from AppliedObjectStyle or Properties.
		// This requires extracting colors from object style definitions or
		// parsing Properties.OtherElements for direct color references.
		// See: https://app.clickup.com/t/86c6nxkk4

		// Analyze the parent story if present
		if tf.ParentStory != "" {
			storyFilename := StoryPath(tf.ParentStory)
			st, err := rm.pkg.Story(storyFilename)
			if err == nil {
				if err := rm.analyzeStory(st, deps); err != nil {
					// Don't fail - just skip this story
					continue
				}
			}
		}
	}

	// Analyze rectangles
	for _, rect := range sp.InnerSpread.Rectangles {
		// Track the applied object style
		if rect.AppliedObjectStyle != "" {
			deps.objectStyles[rect.AppliedObjectStyle] = true
		}

		// Track the layer
		if rect.ItemLayer != "" {
			deps.layers[rect.ItemLayer] = true
		}

		// Note: Rectangle doesn't have direct StrokeColor/FillColor attributes.
		// Colors are inherited from AppliedObjectStyle or defined in Properties element.
	}

	// Analyze ovals
	for i := range sp.InnerSpread.Ovals {
		oval := &sp.InnerSpread.Ovals[i]
		if oval.AppliedObjectStyle != "" {
			deps.objectStyles[oval.AppliedObjectStyle] = true
		}

		// Track stroke and fill colors
		if oval.StrokeColor != "" {
			deps.colors[oval.StrokeColor] = true
		}
		if oval.FillColor != "" {
			deps.colors[oval.FillColor] = true
		}
	}

	// Analyze polygons
	for i := range sp.InnerSpread.Polygons {
		polygon := &sp.InnerSpread.Polygons[i]
		if polygon.AppliedObjectStyle != "" {
			deps.objectStyles[polygon.AppliedObjectStyle] = true
		}

		// Track stroke and fill colors
		if polygon.StrokeColor != "" {
			deps.colors[polygon.StrokeColor] = true
		}
		if polygon.FillColor != "" {
			deps.colors[polygon.FillColor] = true
		}
	}

	// Analyze graphic lines
	for i := range sp.InnerSpread.GraphicLines {
		line := &sp.InnerSpread.GraphicLines[i]
		if line.AppliedObjectStyle != "" {
			deps.objectStyles[line.AppliedObjectStyle] = true
		}
		// Track stroke and fill colors
		if line.StrokeColor != "" {
			deps.colors[line.StrokeColor] = true
		}
		if line.FillColor != "" {
			deps.colors[line.FillColor] = true
		}
	}

	// Analyze groups
	for i := range sp.InnerSpread.Groups {
		group := &sp.InnerSpread.Groups[i]
		if group.AppliedObjectStyle != "" {
			deps.objectStyles[group.AppliedObjectStyle] = true
		}
	}

	return nil
}

// findOrphanedFonts identifies fonts that are defined but not used.
func (rm *ResourceManager) findOrphanedFonts(deps *dependencySet, result *OrphanedResources) error {
	// Get all defined fonts
	fonts, err := rm.pkg.Fonts()
	if err != nil {
		// If there's no Fonts.xml file, that's okay - no fonts to clean up
		if errors.Is(err, common.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("failed to get fonts: %w", err)
	}

	// Find fonts that are defined but not used
	for _, fontFamily := range fonts.FontFamilies {
		if !deps.fonts[fontFamily.Name] {
			result.Fonts = append(result.Fonts, fontFamily.Name)
		}
	}

	return nil
}

// findOrphanedStyles identifies paragraph and character styles that are defined but not used.
func (rm *ResourceManager) findOrphanedStyles(deps *dependencySet, result *OrphanedResources) error {
	// Get all defined styles
	styles, err := rm.pkg.Styles()
	if err != nil {
		// If there's no Styles.xml file, that's okay - no styles to clean up
		if errors.Is(err, common.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("failed to get styles: %w", err)
	}

	// Find orphaned paragraph styles
	if styles.RootParagraphStyleGroup != nil {
		for _, ps := range styles.RootParagraphStyleGroup.ParagraphStyles {
			if !deps.paragraphStyles[ps.Self] {
				result.ParagraphStyles = append(result.ParagraphStyles, ps.Self)
			}
		}
	}

	// Find orphaned character styles
	if styles.RootCharacterStyleGroup != nil {
		for _, cs := range styles.RootCharacterStyleGroup.CharacterStyles {
			if !deps.characterStyles[cs.Self] {
				result.CharacterStyles = append(result.CharacterStyles, cs.Self)
			}
		}
	}

	return nil
}

// findOrphanedObjectStyles identifies object styles that are defined but not used.
func (rm *ResourceManager) findOrphanedObjectStyles(deps *dependencySet, result *OrphanedResources) error {
	// Get all defined styles
	styles, err := rm.pkg.Styles()
	if err != nil {
		// If there's no Styles.xml file, that's okay - no object styles to clean up
		if errors.Is(err, common.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("failed to get styles: %w", err)
	}

	// Find orphaned object styles
	if styles.RootObjectStyleGroup != nil {
		for _, os := range styles.RootObjectStyleGroup.ObjectStyles {
			if !deps.objectStyles[os.Self] {
				result.ObjectStyles = append(result.ObjectStyles, os.Self)
			}
		}
	}

	return nil
}

// findOrphanedColors identifies colors that are defined but not used.
func (rm *ResourceManager) findOrphanedColors(deps *dependencySet, result *OrphanedResources) error {
	// Get all defined colors
	graphics, err := rm.pkg.Graphics()
	if err != nil {
		// If there's no Graphic.xml file, that's okay - no colors to clean up
		if errors.Is(err, common.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("failed to get graphics: %w", err)
	}

	// Find colors that are defined but not used
	for _, color := range graphics.Colors {
		colorRef := "Color/" + color.Self
		if !deps.colors[colorRef] {
			result.Colors = append(result.Colors, color.Self)
		}
	}

	return nil
}

// findOrphanedSwatches identifies swatches that are defined but not used.
func (rm *ResourceManager) findOrphanedSwatches(deps *dependencySet, result *OrphanedResources) error {
	// Get all defined swatches
	graphics, err := rm.pkg.Graphics()
	if err != nil {
		// If there's no Graphic.xml file, that's okay - no swatches to clean up
		if errors.Is(err, common.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("failed to get graphics: %w", err)
	}

	// Find swatches that are defined but not used
	// Note: We check both "Swatch/" and "Color/" prefixes as swatches can be referenced both ways
	for _, swatch := range graphics.Swatches {
		swatchRef1 := "Swatch/" + swatch.Self
		swatchRef2 := "Color/" + swatch.Self
		if !deps.swatches[swatchRef1] && !deps.colors[swatchRef2] {
			result.Swatches = append(result.Swatches, swatch.Self)
		}
	}

	return nil
}

// findOrphanedLayers identifies layers that exist but have no page items on them.
// NOTE: Layer tracking is currently disabled as layers are not stored in SpreadElement.
// This will be re-enabled once we understand where layers are stored in IDML.
func (rm *ResourceManager) findOrphanedLayers(deps *dependencySet, result *OrphanedResources) error {
	// Layer tracking not yet implemented.
	// See: https://app.clickup.com/t/86c6nxkk2
	return nil
}

// CleanupOrphans removes orphaned resources according to the options.
// This method first identifies orphaned resources using FindOrphans(),
// then removes them according to the CleanupOptions configuration.
//
// If DryRun is true, no resources are actually removed - the method just
// returns what would be removed.
//
// Returns a CleanupResult containing details about what was removed (or
// would be removed in DryRun mode).
func (rm *ResourceManager) CleanupOrphans(opts CleanupOptions) (*CleanupResult, error) {
	// Step 1: Find all orphaned resources
	orphans, err := rm.FindOrphans()
	if err != nil {
		return nil, common.WrapError("idml", "cleanup orphans", fmt.Errorf("failed to find orphans: %w", err))
	}

	result := &CleanupResult{}

	// Step 2: If dry run, just return what would be removed
	if opts.DryRun {
		if opts.RemoveOrphanedFonts {
			result.RemovedFonts = orphans.Fonts
		}
		if opts.RemoveOrphanedParagraphStyles || opts.RemoveOrphanedCharacterStyles {
			result.RemovedParagraphStyles = orphans.ParagraphStyles
			result.RemovedCharacterStyles = orphans.CharacterStyles
		}
		if opts.RemoveOrphanedObjectStyles {
			result.RemovedObjectStyles = orphans.ObjectStyles
		}
		if opts.RemoveOrphanedColors {
			result.RemovedColors = orphans.Colors
		}
		if opts.RemoveOrphanedSwatches {
			result.RemovedSwatches = orphans.Swatches
		}
		if opts.RemoveOrphanedLayers {
			result.RemovedLayers = orphans.Layers
		}
		return result, nil
	}

	// Step 3: Remove orphaned fonts if requested
	if opts.RemoveOrphanedFonts {
		if err := rm.removeOrphanedFonts(orphans.Fonts, result); err != nil {
			return result, common.WrapError("idml", "remove orphaned fonts", err)
		}
	}

	// Step 4: Remove orphaned styles if requested
	if opts.RemoveOrphanedParagraphStyles || opts.RemoveOrphanedCharacterStyles {
		if err := rm.removeOrphanedStyles(orphans, opts, result); err != nil {
			return result, common.WrapError("idml", "remove orphaned styles", err)
		}
	}

	// Step 5: Remove orphaned object styles if requested
	if opts.RemoveOrphanedObjectStyles {
		if err := rm.removeOrphanedObjectStyles(orphans.ObjectStyles, result); err != nil {
			return result, common.WrapError("idml", "remove orphaned object styles", err)
		}
	}

	// Step 6: Remove orphaned colors if requested
	if opts.RemoveOrphanedColors {
		if err := rm.removeOrphanedColors(orphans.Colors, result); err != nil {
			return result, common.WrapError("idml", "remove orphaned colors", err)
		}
	}

	// Step 7: Remove orphaned swatches if requested
	if opts.RemoveOrphanedSwatches {
		if err := rm.removeOrphanedSwatches(orphans.Swatches, result); err != nil {
			return result, common.WrapError("idml", "remove orphaned swatches", err)
		}
	}

	// Step 8: Remove orphaned layers if requested
	if opts.RemoveOrphanedLayers {
		if err := rm.removeOrphanedLayers(orphans.Layers, result); err != nil {
			return result, common.WrapError("idml", "remove orphaned layers", err)
		}
	}

	return result, nil
}

// removeOrphanedFonts removes the specified fonts from the Fonts.xml file.
func (rm *ResourceManager) removeOrphanedFonts(fontNames []string, result *CleanupResult) error {
	if len(fontNames) == 0 {
		return nil // Nothing to do
	}

	// Get the fonts file
	fonts, err := rm.pkg.Fonts()
	if err != nil {
		return fmt.Errorf("failed to get fonts: %w", err)
	}

	// Create a set for fast lookup
	fontsToRemove := make(map[string]bool)
	for _, name := range fontNames {
		fontsToRemove[name] = true
	}

	// Filter out orphaned fonts
	filtered := make([]resources.FontFamily, 0, len(fonts.FontFamilies))
	for _, ff := range fonts.FontFamilies {
		if !fontsToRemove[ff.Name] {
			filtered = append(filtered, ff)
		} else {
			result.RemovedFonts = append(result.RemovedFonts, ff.Name)
		}
	}

	fonts.FontFamilies = filtered

	// Update the cached fonts
	rm.pkg.SetFonts(fonts)

	// Marshal and update the file entry
	data, err := resources.MarshalFontsFile(fonts)
	if err != nil {
		return fmt.Errorf("failed to marshal fonts: %w", err)
	}

	// Update the file entry, preserving the header if it exists
	if entry, exists := rm.pkg.files[PathFonts]; exists {
		entry.data = data
	} else {
		rm.pkg.files[PathFonts] = &fileEntry{data: data}
	}

	return nil
}

// removeOrphanedStyles removes the specified styles from the Styles.xml file.
func (rm *ResourceManager) removeOrphanedStyles(orphans *OrphanedResources, opts CleanupOptions, result *CleanupResult) error {
	if len(orphans.ParagraphStyles) == 0 && len(orphans.CharacterStyles) == 0 {
		return nil // Nothing to do
	}

	// Get the styles file
	styles, err := rm.pkg.Styles()
	if err != nil {
		return fmt.Errorf("failed to get styles: %w", err)
	}

	// Remove orphaned paragraph styles if requested
	if opts.RemoveOrphanedParagraphStyles && len(orphans.ParagraphStyles) > 0 {
		stylesToRemove := make(map[string]bool)
		for _, id := range orphans.ParagraphStyles {
			stylesToRemove[id] = true
		}

		if styles.RootParagraphStyleGroup != nil {
			filtered := make([]resources.ParagraphStyle, 0, len(styles.RootParagraphStyleGroup.ParagraphStyles))
			for _, ps := range styles.RootParagraphStyleGroup.ParagraphStyles {
				if !stylesToRemove[ps.Self] {
					filtered = append(filtered, ps)
				} else {
					result.RemovedParagraphStyles = append(result.RemovedParagraphStyles, ps.Self)
				}
			}
			styles.RootParagraphStyleGroup.ParagraphStyles = filtered
		}
	}

	// Remove orphaned character styles if requested
	if opts.RemoveOrphanedCharacterStyles && len(orphans.CharacterStyles) > 0 {
		stylesToRemove := make(map[string]bool)
		for _, id := range orphans.CharacterStyles {
			stylesToRemove[id] = true
		}

		if styles.RootCharacterStyleGroup != nil {
			filtered := make([]resources.CharacterStyle, 0, len(styles.RootCharacterStyleGroup.CharacterStyles))
			for _, cs := range styles.RootCharacterStyleGroup.CharacterStyles {
				if !stylesToRemove[cs.Self] {
					filtered = append(filtered, cs)
				} else {
					result.RemovedCharacterStyles = append(result.RemovedCharacterStyles, cs.Self)
				}
			}
			styles.RootCharacterStyleGroup.CharacterStyles = filtered
		}
	}

	// Update the cached styles
	rm.pkg.SetStyles(styles)

	// Marshal and update the file entry
	data, err := resources.MarshalStylesFile(styles)
	if err != nil {
		return fmt.Errorf("failed to marshal styles: %w", err)
	}

	// Update the file entry, preserving the header if it exists
	if entry, exists := rm.pkg.files[PathStyles]; exists {
		entry.data = data
	} else {
		rm.pkg.files[PathStyles] = &fileEntry{data: data}
	}

	return nil
}

// removeOrphanedObjectStyles removes the specified object styles from the Styles.xml file.
func (rm *ResourceManager) removeOrphanedObjectStyles(styleIDs []string, result *CleanupResult) error {
	if len(styleIDs) == 0 {
		return nil // Nothing to do
	}

	// Get the styles file
	styles, err := rm.pkg.Styles()
	if err != nil {
		return fmt.Errorf("failed to get styles: %w", err)
	}

	// Create a set for fast lookup
	stylesToRemove := make(map[string]bool)
	for _, id := range styleIDs {
		stylesToRemove[id] = true
	}

	// Filter out orphaned object styles
	if styles.RootObjectStyleGroup != nil {
		filtered := make([]resources.ObjectStyle, 0, len(styles.RootObjectStyleGroup.ObjectStyles))
		for _, os := range styles.RootObjectStyleGroup.ObjectStyles {
			if !stylesToRemove[os.Self] {
				filtered = append(filtered, os)
			} else {
				result.RemovedObjectStyles = append(result.RemovedObjectStyles, os.Self)
			}
		}
		styles.RootObjectStyleGroup.ObjectStyles = filtered
	}

	// Update the cached styles
	rm.pkg.SetStyles(styles)

	// Marshal and update the file entry
	data, err := resources.MarshalStylesFile(styles)
	if err != nil {
		return fmt.Errorf("failed to marshal styles: %w", err)
	}

	// Update the file entry
	if entry, exists := rm.pkg.files[PathStyles]; exists {
		entry.data = data
	} else {
		rm.pkg.files[PathStyles] = &fileEntry{data: data}
	}

	return nil
}

// removeOrphanedColors removes the specified colors from the Graphic.xml file.
func (rm *ResourceManager) removeOrphanedColors(colorIDs []string, result *CleanupResult) error {
	if len(colorIDs) == 0 {
		return nil // Nothing to do
	}

	// Get the graphics file
	graphics, err := rm.pkg.Graphics()
	if err != nil {
		return fmt.Errorf("failed to get graphics: %w", err)
	}

	// Create a set for fast lookup
	colorsToRemove := make(map[string]bool)
	for _, id := range colorIDs {
		colorsToRemove[id] = true
	}

	// Filter out orphaned colors
	filtered := make([]resources.Color, 0, len(graphics.Colors))
	for _, color := range graphics.Colors {
		if !colorsToRemove[color.Self] {
			filtered = append(filtered, color)
		} else {
			result.RemovedColors = append(result.RemovedColors, color.Self)
		}
	}
	graphics.Colors = filtered

	// Update the cached graphics
	rm.pkg.SetGraphics(graphics)

	// Marshal and update the file entry
	data, err := resources.MarshalGraphicFile(graphics)
	if err != nil {
		return fmt.Errorf("failed to marshal graphics: %w", err)
	}

	// Update the file entry
	if entry, exists := rm.pkg.files[PathGraphic]; exists {
		entry.data = data
	} else {
		rm.pkg.files[PathGraphic] = &fileEntry{data: data}
	}

	return nil
}

// removeOrphanedSwatches removes the specified swatches from the Graphic.xml file.
func (rm *ResourceManager) removeOrphanedSwatches(swatchIDs []string, result *CleanupResult) error {
	if len(swatchIDs) == 0 {
		return nil // Nothing to do
	}

	// Get the graphics file
	graphics, err := rm.pkg.Graphics()
	if err != nil {
		return fmt.Errorf("failed to get graphics: %w", err)
	}

	// Create a set for fast lookup
	swatchesToRemove := make(map[string]bool)
	for _, id := range swatchIDs {
		swatchesToRemove[id] = true
	}

	// Filter out orphaned swatches
	filtered := make([]resources.Swatch, 0, len(graphics.Swatches))
	for _, swatch := range graphics.Swatches {
		if !swatchesToRemove[swatch.Self] {
			filtered = append(filtered, swatch)
		} else {
			result.RemovedSwatches = append(result.RemovedSwatches, swatch.Self)
		}
	}
	graphics.Swatches = filtered

	// Update the cached graphics
	rm.pkg.SetGraphics(graphics)

	// Marshal and update the file entry
	data, err := resources.MarshalGraphicFile(graphics)
	if err != nil {
		return fmt.Errorf("failed to marshal graphics: %w", err)
	}

	// Update the file entry
	if entry, exists := rm.pkg.files[PathGraphic]; exists {
		entry.data = data
	} else {
		rm.pkg.files[PathGraphic] = &fileEntry{data: data}
	}

	return nil
}

// removeOrphanedLayers removes the specified layers from all spread files.
// NOTE: Layer removal is currently disabled as layers are not stored in SpreadElement.
// This will be re-enabled once we understand where layers are stored in IDML.
func (rm *ResourceManager) removeOrphanedLayers(layerIDs []string, result *CleanupResult) error {
	// Layer removal not yet implemented.
	// See: https://app.clickup.com/t/86c6nxkk2
	return nil
}

// ============================================================================
// Color Extraction from Style Definitions
// ============================================================================

// findParagraphStyleByID recursively searches for a paragraph style by ID.
// It searches through the paragraph style group hierarchy, including nested groups.
func (rm *ResourceManager) findParagraphStyleByID(group *resources.ParagraphStyleGroup, styleID string) *resources.ParagraphStyle {
	if group == nil {
		return nil
	}

	// Search in current group's styles
	for i := range group.ParagraphStyles {
		if group.ParagraphStyles[i].Self == styleID {
			return &group.ParagraphStyles[i]
		}
	}

	// Search in nested groups
	for i := range group.NestedGroups {
		if style := rm.findParagraphStyleByID(&group.NestedGroups[i], styleID); style != nil {
			return style
		}
	}

	return nil
}

// extractColorsFromParagraphStyles looks up used paragraph styles and extracts their colors.
// Colors are defined as FillColor attributes on paragraph style definitions in Styles.xml.
// This is called after analyzing stories to collect the full set of color dependencies.
func (rm *ResourceManager) extractColorsFromParagraphStyles(deps *dependencySet) error {
	// Get Styles file
	styles, err := rm.pkg.Styles()
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			return nil // No styles file, nothing to extract
		}
		return fmt.Errorf("failed to get styles: %w", err)
	}

	// For each used paragraph style, extract its colors
	for styleID := range deps.paragraphStyles {
		// Skip built-in styles
		if strings.HasPrefix(styleID, "$ID/") {
			continue
		}

		// Find the style definition
		ps := rm.findParagraphStyleByID(styles.RootParagraphStyleGroup, styleID)
		if ps == nil {
			continue // Style not found
		}

		// Extract FillColor if present
		if ps.FillColor != "" && ps.FillColor != "Text Color" {
			deps.colors[ps.FillColor] = true
		}

		// TODO: Extract rule colors from Properties if present
		// This requires adding GetRuleAboveColor() and GetRuleBelowColor() helpers
		// to common.Properties, similar to GetAppliedFont()
	}

	return nil
}

// extractColorsFromCharacterStyles looks up used character styles and extracts their colors.
// Colors are defined as FillColor and StrokeColor attributes on character style definitions.
func (rm *ResourceManager) extractColorsFromCharacterStyles(deps *dependencySet) error {
	// Get Styles file
	styles, err := rm.pkg.Styles()
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("failed to get styles: %w", err)
	}

	// For each used character style, extract its colors
	for styleID := range deps.characterStyles {
		// Skip built-in styles
		if strings.HasPrefix(styleID, "$ID/") {
			continue
		}

		// Find the style definition (reuse existing helper from font detection)
		cs := rm.findCharacterStyleByID(styles.RootCharacterStyleGroup, styleID)
		if cs == nil {
			continue
		}

		// Extract FillColor if present
		if cs.FillColor != "" && cs.FillColor != "Text Color" {
			deps.colors[cs.FillColor] = true
		}

		// Extract StrokeColor if present
		if cs.StrokeColor != "" && cs.StrokeColor != "Swatch/None" {
			deps.colors[cs.StrokeColor] = true
		}
	}

	return nil
}
