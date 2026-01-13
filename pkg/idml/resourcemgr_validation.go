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

// ============================================================================
// Phase 2: Validation - Missing Resource Detection
// ============================================================================

// ValidateReferences checks if all referenced resources exist in the package.
// This method analyzes the entire document to find all resource references,
// then verifies that each referenced resource is defined.
//
// Returns a slice of ValidationError for each missing resource, or an error
// if the analysis fails. An empty slice means all references are valid.
func (rm *ResourceManager) ValidateReferences() ([]ValidationError, error) {
	missing, err := rm.FindMissingResources()
	if err != nil {
		return nil, common.WrapError("idml", "validate references", fmt.Errorf("failed to find missing resources: %w", err))
	}

	if !missing.HasMissing() {
		return []ValidationError{}, nil
	}

	var errors []ValidationError

	// Convert missing fonts to validation errors
	for fontFamily, usedBy := range missing.Fonts {
		errors = append(errors, ValidationError{
			ResourceType: "Font",
			ResourceID:   fontFamily,
			UsedBy:       usedBy,
			Message:      fmt.Sprintf("Font '%s' is referenced but not defined in Fonts.xml", fontFamily),
		})
	}

	// Convert missing paragraph styles to validation errors
	for styleID, usedBy := range missing.ParagraphStyles {
		errors = append(errors, ValidationError{
			ResourceType: "ParagraphStyle",
			ResourceID:   styleID,
			UsedBy:       usedBy,
			Message:      fmt.Sprintf("Paragraph style '%s' is referenced but not defined in Styles.xml", styleID),
		})
	}

	// Convert missing character styles to validation errors
	for styleID, usedBy := range missing.CharacterStyles {
		errors = append(errors, ValidationError{
			ResourceType: "CharacterStyle",
			ResourceID:   styleID,
			UsedBy:       usedBy,
			Message:      fmt.Sprintf("Character style '%s' is referenced but not defined in Styles.xml", styleID),
		})
	}

	// Convert missing object styles to validation errors
	for styleID, usedBy := range missing.ObjectStyles {
		errors = append(errors, ValidationError{
			ResourceType: "ObjectStyle",
			ResourceID:   styleID,
			UsedBy:       usedBy,
			Message:      fmt.Sprintf("Object style '%s' is referenced but not defined in Styles.xml", styleID),
		})
	}

	// Convert missing colors to validation errors
	for colorID, usedBy := range missing.Colors {
		errors = append(errors, ValidationError{
			ResourceType: "Color",
			ResourceID:   colorID,
			UsedBy:       usedBy,
			Message:      fmt.Sprintf("Color '%s' is referenced but not defined in Graphic.xml", colorID),
		})
	}

	// Convert missing swatches to validation errors
	for swatchID, usedBy := range missing.Swatches {
		errors = append(errors, ValidationError{
			ResourceType: "Swatch",
			ResourceID:   swatchID,
			UsedBy:       usedBy,
			Message:      fmt.Sprintf("Swatch '%s' is referenced but not defined in Graphic.xml", swatchID),
		})
	}

	// Convert missing layers to validation errors
	for layerID, usedBy := range missing.Layers {
		errors = append(errors, ValidationError{
			ResourceType: "Layer",
			ResourceID:   layerID,
			UsedBy:       usedBy,
			Message:      fmt.Sprintf("Layer '%s' is referenced but not defined in any Spread", layerID),
		})
	}

	return errors, nil
}

// FindMissingResources identifies resources that are referenced but don't exist.
// This is the inverse of FindOrphans() - instead of finding defined-but-unused
// resources, this finds used-but-not-defined resources.
//
// Returns a MissingResources struct containing all missing resources with
// information about where they are used.
func (rm *ResourceManager) FindMissingResources() (*MissingResources, error) {
	result := &MissingResources{
		Fonts:           make(map[string][]string),
		ParagraphStyles: make(map[string][]string),
		CharacterStyles: make(map[string][]string),
		ObjectStyles:    make(map[string][]string),
		Colors:          make(map[string][]string),
		Swatches:        make(map[string][]string),
		Layers:          make(map[string][]string),
	}

	// Step 1: Analyze the entire document to find what's used
	deps, err := rm.analyzeDependencies()
	if err != nil {
		return nil, common.WrapError("idml", "find missing resources", fmt.Errorf("failed to analyze content: %w", err))
	}

	// Step 2: Check if used resources exist
	if err := rm.findMissingFonts(deps, result); err != nil {
		return nil, common.WrapError("idml", "find missing fonts", err)
	}

	if err := rm.findMissingStyles(deps, result); err != nil {
		return nil, common.WrapError("idml", "find missing styles", err)
	}

	if err := rm.findMissingObjectStyles(deps, result); err != nil {
		return nil, common.WrapError("idml", "find missing object styles", err)
	}

	if err := rm.findMissingColors(deps, result); err != nil {
		return nil, common.WrapError("idml", "find missing colors", err)
	}

	if err := rm.findMissingSwatches(deps, result); err != nil {
		return nil, common.WrapError("idml", "find missing swatches", err)
	}

	if err := rm.findMissingLayers(deps, result); err != nil {
		return nil, common.WrapError("idml", "find missing layers", err)
	}

	return result, nil
}

// findMissingFonts checks if all used fonts exist in the Fonts.xml file.
func (rm *ResourceManager) findMissingFonts(deps *dependencySet, result *MissingResources) error {
	// Add validation for parameters
	if deps == nil {
		return common.Errorf("idml", "find missing fonts", "", "dependency set is nil")
	}

	if result == nil {
		return common.Errorf("idml", "find missing fonts", "", "result is nil")
	}

	// If no fonts are used, nothing to check
	if len(deps.fonts) == 0 {
		return nil
	}

	// Get defined fonts
	fonts, err := rm.pkg.Fonts()
	if err != nil {
		// If there's no Fonts.xml but fonts are used, all fonts are missing
		if errors.Is(err, common.ErrNotFound) {
			for fontFamily := range deps.fonts {
				// Validate font family name
				if fontFamily == "" {
					continue // Skip empty font names
				}
				result.Fonts[fontFamily] = []string{"(used in stories)"}
			}
			return nil
		}
		return fmt.Errorf("failed to get fonts: %w", err)
	}

	// Build set of defined fonts
	definedFonts := make(map[string]bool)
	if fonts != nil {
		for _, fontFamily := range fonts.FontFamilies {
			if fontFamily.Name != "" { // Validate font family name
				definedFonts[fontFamily.Name] = true
			}
		}
	}

	// Check each used font
	for fontFamily := range deps.fonts {
		// Validate font family name
		if fontFamily == "" {
			continue // Skip empty font names
		}

		if !definedFonts[fontFamily] {
			// Font is used but not defined - find where it's used
			usedBy := rm.findFontUsage(fontFamily)
			result.Fonts[fontFamily] = usedBy
		}
	}

	return nil
}

// findMissingStyles checks if all used styles exist in the Styles.xml file.
func (rm *ResourceManager) findMissingStyles(deps *dependencySet, result *MissingResources) error {
	// Add validation for parameters
	if deps == nil {
		return common.Errorf("idml", "find missing styles", "", "dependency set is nil")
	}

	if result == nil {
		return common.Errorf("idml", "find missing styles", "", "result is nil")
	}

	// If no styles are used, nothing to check
	if len(deps.paragraphStyles) == 0 && len(deps.characterStyles) == 0 {
		return nil
	}

	// Get defined styles
	styles, err := rm.pkg.Styles()
	if err != nil {
		// If there's no Styles.xml but styles are used, all styles are missing
		if errors.Is(err, common.ErrNotFound) {
			if len(deps.paragraphStyles) > 0 {
				for styleID := range deps.paragraphStyles {
					// Validate style ID
					if styleID == "" {
						continue // Skip empty style IDs
					}
					result.ParagraphStyles[styleID] = []string{"(used in stories)"}
				}
			}
			if len(deps.characterStyles) > 0 {
				for styleID := range deps.characterStyles {
					// Validate style ID
					if styleID == "" {
						continue // Skip empty style IDs
					}
					result.CharacterStyles[styleID] = []string{"(used in stories)"}
				}
			}
			return nil
		}
		return fmt.Errorf("failed to get styles: %w", err)
	}

	// Build sets of defined styles
	definedParagraphStyles := make(map[string]bool)
	definedCharacterStyles := make(map[string]bool)

	if styles != nil && styles.RootParagraphStyleGroup != nil {
		for _, ps := range styles.RootParagraphStyleGroup.ParagraphStyles {
			if ps.Self != "" { // Validate style ID
				definedParagraphStyles[ps.Self] = true
			}
		}
	}

	if styles != nil && styles.RootCharacterStyleGroup != nil {
		for _, cs := range styles.RootCharacterStyleGroup.CharacterStyles {
			if cs.Self != "" { // Validate style ID
				definedCharacterStyles[cs.Self] = true
			}
		}
	}

	// Check each used paragraph style
	for styleID := range deps.paragraphStyles {
		// Validate style ID
		if styleID == "" {
			continue // Skip empty style IDs
		}

		if !definedParagraphStyles[styleID] {
			usedBy := rm.findParagraphStyleUsage(styleID)
			result.ParagraphStyles[styleID] = usedBy
		}
	}

	// Check each used character style
	for styleID := range deps.characterStyles {
		// Validate style ID
		if styleID == "" {
			continue // Skip empty style IDs
		}

		if !definedCharacterStyles[styleID] {
			usedBy := rm.findCharacterStyleUsage(styleID)
			result.CharacterStyles[styleID] = usedBy
		}
	}

	return nil
}

// findFontUsage finds all stories that use a specific font.
// This is a helper method for detailed error reporting.
func (rm *ResourceManager) findFontUsage(fontFamily string) []string {
	var usedBy []string

	stories, err := rm.pkg.Stories()
	if err != nil {
		return usedBy
	}

	// Check each story for font usage
	for filename, st := range stories {
		if rm.storyUsesFont(st, fontFamily) {
			usedBy = append(usedBy, filename)
		}
	}

	return usedBy
}

// findCharacterStyleByID recursively searches for a character style by ID.
// It searches through the character style group hierarchy, including nested groups.
func (rm *ResourceManager) findCharacterStyleByID(group *resources.CharacterStyleGroup, styleID string) *resources.CharacterStyle {
	if group == nil {
		return nil
	}

	// Search in current group's styles
	for i := range group.CharacterStyles {
		if group.CharacterStyles[i].Self == styleID {
			return &group.CharacterStyles[i]
		}
	}

	// Search in nested groups
	for i := range group.NestedGroups {
		if style := rm.findCharacterStyleByID(&group.NestedGroups[i], styleID); style != nil {
			return style
		}
	}

	return nil
}

// getFontFromCharacterStyle extracts the AppliedFont from a character style,
// following the BasedOn chain if necessary to resolve inherited fonts.
// Returns the font family name or empty string if not found.
func (rm *ResourceManager) getFontFromCharacterStyle(styleID string) (string, error) {
	// Skip built-in styles - they don't have custom fonts we track
	if strings.HasPrefix(styleID, "$ID/") {
		return "", nil
	}

	// Get the Styles file
	styles, err := rm.pkg.Styles()
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			return "", nil
		}
		return "", fmt.Errorf("failed to get styles: %w", err)
	}

	// Track visited styles to prevent circular references
	visited := make(map[string]bool)
	current := styleID

	// Walk the BasedOn chain
	for {
		// Check for circular reference
		if visited[current] {
			break
		}
		visited[current] = true

		// Find the style definition
		characterStyle := rm.findCharacterStyleByID(styles.RootCharacterStyleGroup, current)
		if characterStyle == nil {
			// Style not found
			break
		}

		// Try to get AppliedFont from this style
		if characterStyle.Properties != nil {
			font := characterStyle.Properties.GetAppliedFont()
			if font != "" {
				return font, nil
			}

			// No font here, check parent style
			basedOn := characterStyle.Properties.GetBasedOn()
			if basedOn == "" {
				// No parent - end of chain
				break
			}

			// Stop if parent is built-in
			if strings.HasPrefix(basedOn, "$ID/") {
				break
			}

			// Continue with parent
			current = basedOn
		} else {
			// No properties - can't continue
			break
		}
	}

	return "", nil
}

// storyUsesFont checks if a story uses a specific font.
// It examines all character style ranges in the story and checks if any of their
// applied character styles reference the given font family (following BasedOn inheritance).
func (rm *ResourceManager) storyUsesFont(st *story.Story, fontFamily string) bool {
	// Iterate through all paragraph style ranges
	for _, psr := range st.StoryElement.ParagraphStyleRanges {
		// Check each character style range within the paragraph
		for _, csr := range psr.CharacterStyleRanges {
			// Get the font from this character style (with inheritance)
			font, err := rm.getFontFromCharacterStyle(csr.AppliedCharacterStyle)
			if err != nil {
				// Log error but continue checking other styles
				continue
			}

			// Compare font family names (case-sensitive as per IDML spec)
			if font == fontFamily {
				return true
			}
		}
	}

	return false
}

// findParagraphStyleUsage finds all stories that use a specific paragraph style.
func (rm *ResourceManager) findParagraphStyleUsage(styleID string) []string {
	var usedBy []string

	stories, err := rm.pkg.Stories()
	if err != nil {
		return usedBy
	}

	for filename, story := range stories {
		// Check each paragraph style range
		for _, psr := range story.StoryElement.ParagraphStyleRanges {
			if psr.AppliedParagraphStyle == styleID {
				usedBy = append(usedBy, filename)
				break // Only add the filename once
			}
		}
	}

	return usedBy
}

// findCharacterStyleUsage finds all stories that use a specific character style.
func (rm *ResourceManager) findCharacterStyleUsage(styleID string) []string {
	var usedBy []string

	stories, err := rm.pkg.Stories()
	if err != nil {
		return usedBy
	}

	for filename, story := range stories {
		// Check each paragraph style range
		for _, psr := range story.StoryElement.ParagraphStyleRanges {
			// Check each character style range within the paragraph
			for _, csr := range psr.CharacterStyleRanges {
				if csr.AppliedCharacterStyle == styleID {
					usedBy = append(usedBy, filename)
					break
				}
			}
		}
	}

	return usedBy
}

// findMissingObjectStyles checks if all used object styles exist in the Styles.xml file.
func (rm *ResourceManager) findMissingObjectStyles(deps *dependencySet, result *MissingResources) error {
	// If no object styles are used, nothing to check
	if len(deps.objectStyles) == 0 {
		return nil
	}

	// Get defined styles
	styles, err := rm.pkg.Styles()
	if err != nil {
		// If there's no Styles.xml but object styles are used, all are missing
		if errors.Is(err, common.ErrNotFound) {
			for styleID := range deps.objectStyles {
				result.ObjectStyles[styleID] = []string{"(used in spreads)"}
			}
			return nil
		}
		return fmt.Errorf("failed to get styles: %w", err)
	}

	// Build set of defined object styles
	definedObjectStyles := make(map[string]bool)
	if styles.RootObjectStyleGroup != nil {
		for _, os := range styles.RootObjectStyleGroup.ObjectStyles {
			definedObjectStyles[os.Self] = true
		}
	}

	// Check each used object style
	for styleID := range deps.objectStyles {
		if !definedObjectStyles[styleID] {
			usedBy := rm.findObjectStyleUsage(styleID)
			result.ObjectStyles[styleID] = usedBy
		}
	}

	return nil
}

// findMissingColors checks if all used colors exist in the Graphic.xml file.
func (rm *ResourceManager) findMissingColors(deps *dependencySet, result *MissingResources) error {
	// If no colors are used, nothing to check
	if len(deps.colors) == 0 {
		return nil
	}

	// Get defined colors
	graphics, err := rm.pkg.Graphics()
	if err != nil {
		// If there's no Graphic.xml but colors are used, all are missing
		if errors.Is(err, common.ErrNotFound) {
			for colorRef := range deps.colors {
				result.Colors[colorRef] = []string{"(used in content)"}
			}
			return nil
		}
		return fmt.Errorf("failed to get graphics: %w", err)
	}

	// Build set of defined colors (with "Color/" prefix)
	definedColors := make(map[string]bool)
	for _, color := range graphics.Colors {
		definedColors["Color/"+color.Self] = true
	}

	// Check each used color
	for colorRef := range deps.colors {
		if !definedColors[colorRef] {
			usedBy := rm.findColorUsage(colorRef)
			result.Colors[colorRef] = usedBy
		}
	}

	return nil
}

// findMissingSwatches checks if all used swatches exist in the Graphic.xml file.
func (rm *ResourceManager) findMissingSwatches(deps *dependencySet, result *MissingResources) error {
	// If no swatches are used, nothing to check
	if len(deps.swatches) == 0 {
		return nil
	}

	// Get defined swatches
	graphics, err := rm.pkg.Graphics()
	if err != nil {
		// If there's no Graphic.xml but swatches are used, all are missing
		if errors.Is(err, common.ErrNotFound) {
			for swatchRef := range deps.swatches {
				result.Swatches[swatchRef] = []string{"(used in content)"}
			}
			return nil
		}
		return fmt.Errorf("failed to get graphics: %w", err)
	}

	// Build set of defined swatches (with both "Swatch/" and "Color/" prefixes)
	definedSwatches := make(map[string]bool)
	for _, swatch := range graphics.Swatches {
		definedSwatches["Swatch/"+swatch.Self] = true
		definedSwatches["Color/"+swatch.Self] = true
	}

	// Check each used swatch
	for swatchRef := range deps.swatches {
		if !definedSwatches[swatchRef] {
			usedBy := rm.findSwatchUsage(swatchRef)
			result.Swatches[swatchRef] = usedBy
		}
	}

	return nil
}

// findMissingLayers checks if all used layers exist in the spread files.
// NOTE: Layer validation is currently disabled as layers are not stored in SpreadElement.
// This will be re-enabled once we understand where layers are stored in IDML.
func (rm *ResourceManager) findMissingLayers(deps *dependencySet, result *MissingResources) error {
	// Layer validation not yet implemented.
	// See: https://app.clickup.com/t/86c6nxkk2
	return nil
}

// findObjectStyleUsage finds all page items that use a specific object style.
func (rm *ResourceManager) findObjectStyleUsage(styleID string) []string {
	var usedBy []string

	spreads, err := rm.pkg.Spreads()
	if err != nil {
		return usedBy
	}

	for filename, sp := range spreads {
		// Check text frames
		for _, tf := range sp.InnerSpread.TextFrames {
			if tf.AppliedObjectStyle == styleID {
				usedBy = append(usedBy, filename)
				break
			}
		}

		// Check rectangles
		for _, rect := range sp.InnerSpread.Rectangles {
			if rect.AppliedObjectStyle == styleID {
				usedBy = append(usedBy, filename)
				break
			}
		}
	}

	return usedBy
}

// findColorUsage finds all elements that use a specific color.
// Returns a list of filenames where the color is used (spreads or styles).
func (rm *ResourceManager) findColorUsage(colorRef string) []string {
	var usedBy []string
	seen := make(map[string]bool) // Prevent duplicates

	// Check spreads for page item colors
	spreads, err := rm.pkg.Spreads()
	if err == nil {
		for filename, sp := range spreads {
			if rm.spreadUsesColor(sp, colorRef) && !seen[filename] {
				usedBy = append(usedBy, filename)
				seen[filename] = true
			}
		}
	}

	// Check styles for color usage
	styles, err := rm.pkg.Styles()
	if err == nil {
		// Check paragraph styles
		if rm.paragraphStylesUseColor(styles, colorRef) {
			location := "Resources/Styles.xml (paragraph styles)"
			if !seen[location] {
				usedBy = append(usedBy, location)
				seen[location] = true
			}
		}

		// Check character styles
		if rm.characterStylesUseColor(styles, colorRef) {
			location := "Resources/Styles.xml (character styles)"
			if !seen[location] {
				usedBy = append(usedBy, location)
				seen[location] = true
			}
		}
	}

	return usedBy
}

// spreadUsesColor checks if a spread uses a specific color in any of its page items.
func (rm *ResourceManager) spreadUsesColor(sp *spread.Spread, colorRef string) bool {
	// Check ovals
	for _, oval := range sp.InnerSpread.Ovals {
		if oval.StrokeColor == colorRef || oval.FillColor == colorRef {
			return true
		}
	}

	// Check polygons
	for _, polygon := range sp.InnerSpread.Polygons {
		if polygon.StrokeColor == colorRef || polygon.FillColor == colorRef {
			return true
		}
	}

	// Check graphic lines
	for _, line := range sp.InnerSpread.GraphicLines {
		if line.StrokeColor == colorRef || line.FillColor == colorRef {
			return true
		}
	}

	// TODO: Check rectangles (they use object styles or Properties for colors)
	// TODO: Check text frames (they use object styles or Properties for colors)

	return false
}

// paragraphStylesUseColor checks if any paragraph style uses the specified color.
func (rm *ResourceManager) paragraphStylesUseColor(styles *resources.StylesFile, colorRef string) bool {
	if styles.RootParagraphStyleGroup == nil {
		return false
	}

	return rm.paragraphStyleGroupUsesColor(styles.RootParagraphStyleGroup, colorRef)
}

// paragraphStyleGroupUsesColor recursively checks a paragraph style group for color usage.
func (rm *ResourceManager) paragraphStyleGroupUsesColor(group *resources.ParagraphStyleGroup, colorRef string) bool {
	// Check styles in current group
	for _, ps := range group.ParagraphStyles {
		if ps.FillColor == colorRef {
			return true
		}
		// TODO: Check rule colors in Properties
	}

	// Check nested groups
	for _, nestedGroup := range group.NestedGroups {
		if rm.paragraphStyleGroupUsesColor(&nestedGroup, colorRef) {
			return true
		}
	}

	return false
}

// characterStylesUseColor checks if any character style uses the specified color.
func (rm *ResourceManager) characterStylesUseColor(styles *resources.StylesFile, colorRef string) bool {
	if styles.RootCharacterStyleGroup == nil {
		return false
	}

	return rm.characterStyleGroupUsesColor(styles.RootCharacterStyleGroup, colorRef)
}

// characterStyleGroupUsesColor recursively checks a character style group for color usage.
func (rm *ResourceManager) characterStyleGroupUsesColor(group *resources.CharacterStyleGroup, colorRef string) bool {
	// Check styles in current group
	for _, cs := range group.CharacterStyles {
		if cs.FillColor == colorRef || cs.StrokeColor == colorRef {
			return true
		}
	}

	// Check nested groups
	for _, nestedGroup := range group.NestedGroups {
		if rm.characterStyleGroupUsesColor(&nestedGroup, colorRef) {
			return true
		}
	}

	return false
}

// findSwatchUsage finds all elements that use a specific swatch.
func (rm *ResourceManager) findSwatchUsage(swatchRef string) []string {
	// Swatches are referenced similarly to colors
	return rm.findColorUsage(swatchRef)
}

// findLayerUsage finds all page items that are on a specific layer.
// NOTE: Layer tracking is currently simplified and will be enhanced when we understand
// the full IDML layer structure.
// findLayerUsage finds where a layer is used in the document
// This function is currently unused but kept for potential future use
// nolint:unused
func (rm *ResourceManager) findLayerUsage(layerID string) []string {
	var usedBy []string

	spreads, err := rm.pkg.Spreads()
	if err != nil {
		return usedBy
	}

	for filename, sp := range spreads {
		// Check text frames
		for _, tf := range sp.InnerSpread.TextFrames {
			if tf.ItemLayer == layerID {
				usedBy = append(usedBy, filename)
				break
			}
		}

		// Check rectangles
		for _, rect := range sp.InnerSpread.Rectangles {
			if rect.ItemLayer == layerID {
				usedBy = append(usedBy, filename)
				break
			}
		}
	}

	return usedBy
}

// findMissingResourcesForDeps checks if resources in the dependency set exist in the package.
// This is used internally by AddStory/UpdateStory to validate dependencies before adding content.
func (rm *ResourceManager) findMissingResourcesForDeps(deps *dependencySet) (*MissingResources, error) {
	result := &MissingResources{
		Fonts:           make(map[string][]string),
		ParagraphStyles: make(map[string][]string),
		CharacterStyles: make(map[string][]string),
		ObjectStyles:    make(map[string][]string),
		Colors:          make(map[string][]string),
		Swatches:        make(map[string][]string),
		Layers:          make(map[string][]string),
	}

	// Check fonts
	if len(deps.fonts) > 0 {
		if err := rm.findMissingFontsInDeps(deps, result); err != nil {
			return nil, err
		}
	}

	// Check paragraph styles
	if len(deps.paragraphStyles) > 0 {
		if err := rm.findMissingParagraphStylesInDeps(deps, result); err != nil {
			return nil, err
		}
	}

	// Check character styles
	if len(deps.characterStyles) > 0 {
		if err := rm.findMissingCharacterStylesInDeps(deps, result); err != nil {
			return nil, err
		}
	}

	// Check object styles
	if len(deps.objectStyles) > 0 {
		if err := rm.findMissingObjectStylesInDeps(deps, result); err != nil {
			return nil, err
		}
	}

	// Check colors
	if len(deps.colors) > 0 {
		if err := rm.findMissingColorsInDeps(deps, result); err != nil {
			return nil, err
		}
	}

	// Check swatches
	if len(deps.swatches) > 0 {
		if err := rm.findMissingSwatchesInDeps(deps, result); err != nil {
			return nil, err
		}
	}

	// Check layers
	if len(deps.layers) > 0 {
		if err := rm.findMissingLayersInDeps(deps, result); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// findMissingFontsInDeps checks if fonts in the dependency set exist in the package.
func (rm *ResourceManager) findMissingFontsInDeps(deps *dependencySet, result *MissingResources) error {
	fonts, err := rm.pkg.Fonts()
	if err != nil {
		// If Fonts.xml doesn't exist, all fonts are missing
		for fontFamily := range deps.fonts {
			result.Fonts[fontFamily] = []string{"<new content>"}
		}
		return nil
	}

	// Build set of defined fonts
	definedFonts := make(map[string]bool)
	for _, fontFamily := range fonts.FontFamilies {
		definedFonts[fontFamily.Name] = true
	}

	// Check each used font
	for fontFamily := range deps.fonts {
		if !definedFonts[fontFamily] {
			result.Fonts[fontFamily] = []string{"<new content>"}
		}
	}

	return nil
}

// findMissingParagraphStylesInDeps checks if paragraph styles in the dependency set exist.
func (rm *ResourceManager) findMissingParagraphStylesInDeps(deps *dependencySet, result *MissingResources) error {
	styles, err := rm.pkg.Styles()
	if err != nil {
		// If Styles.xml doesn't exist, all styles are missing
		for styleID := range deps.paragraphStyles {
			result.ParagraphStyles[styleID] = []string{"<new content>"}
		}
		return nil
	}

	// Build set of defined paragraph styles
	definedStyles := make(map[string]bool)
	if styles.RootParagraphStyleGroup != nil {
		for _, style := range styles.RootParagraphStyleGroup.ParagraphStyles {
			// style.Self already contains the full ID (e.g., "ParagraphStyle/$ID/NormalParagraphStyle")
			definedStyles[style.Self] = true
		}
	}

	// Check each used style
	for styleID := range deps.paragraphStyles {
		if !definedStyles[styleID] {
			result.ParagraphStyles[styleID] = []string{"<new content>"}
		}
	}

	return nil
}

// findMissingCharacterStylesInDeps checks if character styles in the dependency set exist.
func (rm *ResourceManager) findMissingCharacterStylesInDeps(deps *dependencySet, result *MissingResources) error {
	styles, err := rm.pkg.Styles()
	if err != nil {
		// If Styles.xml doesn't exist, all styles are missing
		for styleID := range deps.characterStyles {
			result.CharacterStyles[styleID] = []string{"<new content>"}
		}
		return nil
	}

	// Build set of defined character styles
	definedStyles := make(map[string]bool)
	if styles.RootCharacterStyleGroup != nil {
		for _, style := range styles.RootCharacterStyleGroup.CharacterStyles {
			// style.Self already contains the full ID (e.g., "CharacterStyle/$ID/[No character style]")
			definedStyles[style.Self] = true
		}
	}

	// Check each used style
	for styleID := range deps.characterStyles {
		if !definedStyles[styleID] {
			result.CharacterStyles[styleID] = []string{"<new content>"}
		}
	}

	return nil
}

// findMissingObjectStylesInDeps checks if object styles in the dependency set exist.
func (rm *ResourceManager) findMissingObjectStylesInDeps(deps *dependencySet, result *MissingResources) error {
	styles, err := rm.pkg.Styles()
	if err != nil {
		// If Styles.xml doesn't exist, all styles are missing
		for styleID := range deps.objectStyles {
			result.ObjectStyles[styleID] = []string{"<new content>"}
		}
		return nil
	}

	// Build set of defined object styles
	definedStyles := make(map[string]bool)
	if styles.RootObjectStyleGroup != nil {
		for _, style := range styles.RootObjectStyleGroup.ObjectStyles {
			styleRef := "ObjectStyle/" + style.Self
			definedStyles[styleRef] = true
		}
	}

	// Check each used style
	for styleID := range deps.objectStyles {
		if !definedStyles[styleID] {
			result.ObjectStyles[styleID] = []string{"<new content>"}
		}
	}

	return nil
}

// findMissingColorsInDeps checks if colors in the dependency set exist.
func (rm *ResourceManager) findMissingColorsInDeps(deps *dependencySet, result *MissingResources) error {
	graphics, err := rm.pkg.Graphics()
	if err != nil {
		// If Graphic.xml doesn't exist, all colors are missing
		for colorRef := range deps.colors {
			result.Colors[colorRef] = []string{"<new content>"}
		}
		return nil
	}

	// Build set of defined colors
	definedColors := make(map[string]bool)
	for _, color := range graphics.Colors {
		colorRef := "Color/" + color.Self
		definedColors[colorRef] = true
	}

	// Check each used color
	for colorRef := range deps.colors {
		if !definedColors[colorRef] {
			result.Colors[colorRef] = []string{"<new content>"}
		}
	}

	return nil
}

// findMissingSwatchesInDeps checks if swatches in the dependency set exist.
func (rm *ResourceManager) findMissingSwatchesInDeps(deps *dependencySet, result *MissingResources) error {
	graphics, err := rm.pkg.Graphics()
	if err != nil {
		// If Graphic.xml doesn't exist, all swatches are missing
		for swatchRef := range deps.swatches {
			result.Swatches[swatchRef] = []string{"<new content>"}
		}
		return nil
	}

	// Build set of defined swatches
	definedSwatches := make(map[string]bool)
	for _, swatch := range graphics.Swatches {
		swatchRef := "Color/" + swatch.Self
		definedSwatches[swatchRef] = true
	}

	// Check each used swatch
	for swatchRef := range deps.swatches {
		if !definedSwatches[swatchRef] {
			result.Swatches[swatchRef] = []string{"<new content>"}
		}
	}

	return nil
}

// findMissingLayersInDeps checks if layers in the dependency set exist.
func (rm *ResourceManager) findMissingLayersInDeps(deps *dependencySet, result *MissingResources) error {
	// Layer validation not yet implemented.
	// See: https://app.clickup.com/t/86c6nxkk2
	return nil
}

// ValidateResourceReferences validates that all resource references have valid formats.
// This checks for common issues like malformed IDs, invalid prefixes, and circular references.
func (rm *ResourceManager) ValidateResourceReferences() ([]ValidationError, error) {
	var validationErrors []ValidationError

	// Validate style references
	styleErrors, err := rm.validateStyleReferences()
	if err != nil {
		return nil, common.WrapError("idml", "validate resource references", err)
	}
	validationErrors = append(validationErrors, styleErrors...)

	// Validate font references
	fontErrors, err := rm.validateFontReferences()
	if err != nil {
		return nil, common.WrapError("idml", "validate resource references", err)
	}
	validationErrors = append(validationErrors, fontErrors...)

	// Validate color references
	colorErrors, err := rm.validateColorReferences()
	if err != nil {
		return nil, common.WrapError("idml", "validate resource references", err)
	}
	validationErrors = append(validationErrors, colorErrors...)

	return validationErrors, nil
}

// validateStyleReferences checks for invalid style reference formats and circular dependencies.
func (rm *ResourceManager) validateStyleReferences() ([]ValidationError, error) {
	var validationErrors []ValidationError

	styles, err := rm.pkg.Styles()
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			return validationErrors, nil // No styles file to validate
		}
		return nil, fmt.Errorf("failed to get styles: %w", err)
	}

	if styles == nil {
		return validationErrors, nil
	}

	// Validate paragraph styles
	if styles.RootParagraphStyleGroup != nil {
		for _, ps := range styles.RootParagraphStyleGroup.ParagraphStyles {
			// Check for invalid Self ID format
			if ps.Self == "" {
				validationErrors = append(validationErrors, ValidationError{
					ResourceType: "ParagraphStyle",
					ResourceID:   "<empty>",
					Message:      "Paragraph style has empty Self attribute",
				})
				continue
			}

			// Check for invalid BasedOn references
			if ps.Properties != nil {
				basedOn := ps.Properties.GetBasedOn()
				if basedOn != "" && basedOn == ps.Self {
					validationErrors = append(validationErrors, ValidationError{
						ResourceType: "ParagraphStyle",
						ResourceID:   ps.Self,
						Message:      fmt.Sprintf("Paragraph style '%s' has circular BasedOn reference to itself", ps.Self),
					})
				}
			}
		}
	}

	// Validate character styles
	if styles.RootCharacterStyleGroup != nil {
		for _, cs := range styles.RootCharacterStyleGroup.CharacterStyles {
			// Check for invalid Self ID format
			if cs.Self == "" {
				validationErrors = append(validationErrors, ValidationError{
					ResourceType: "CharacterStyle",
					ResourceID:   "<empty>",
					Message:      "Character style has empty Self attribute",
				})
				continue
			}

			// Check for invalid BasedOn references
			if cs.Properties != nil {
				basedOn := cs.Properties.GetBasedOn()
				if basedOn != "" && basedOn == cs.Self {
					validationErrors = append(validationErrors, ValidationError{
						ResourceType: "CharacterStyle",
						ResourceID:   cs.Self,
						Message:      fmt.Sprintf("Character style '%s' has circular BasedOn reference to itself", cs.Self),
					})
				}
			}
		}
	}

	return validationErrors, nil
}

// validateFontReferences checks for invalid font reference formats.
func (rm *ResourceManager) validateFontReferences() ([]ValidationError, error) {
	var validationErrors []ValidationError

	fonts, err := rm.pkg.Fonts()
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			return validationErrors, nil // No fonts file to validate
		}
		return nil, fmt.Errorf("failed to get fonts: %w", err)
	}

	if fonts == nil {
		return validationErrors, nil
	}

	// Validate font family names
	for _, fontFamily := range fonts.FontFamilies {
		if fontFamily.Name == "" {
			validationErrors = append(validationErrors, ValidationError{
				ResourceType: "Font",
				ResourceID:   "<empty>",
				Message:      "Font family has empty Name attribute",
			})
		}
	}

	return validationErrors, nil
}

// validateColorReferences checks for invalid color reference formats.
func (rm *ResourceManager) validateColorReferences() ([]ValidationError, error) {
	var validationErrors []ValidationError

	graphics, err := rm.pkg.Graphics()
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			return validationErrors, nil // No graphics file to validate
		}
		return nil, fmt.Errorf("failed to get graphics: %w", err)
	}

	if graphics == nil {
		return validationErrors, nil
	}

	// Validate color definitions
	for _, color := range graphics.Colors {
		if color.Self == "" {
			validationErrors = append(validationErrors, ValidationError{
				ResourceType: "Color",
				ResourceID:   "<empty>",
				Message:      "Color has empty Self attribute",
			})
		}
	}

	// Validate swatch definitions
	for _, swatch := range graphics.Swatches {
		if swatch.Self == "" {
			validationErrors = append(validationErrors, ValidationError{
				ResourceType: "Swatch",
				ResourceID:   "<empty>",
				Message:      "Swatch has empty Self attribute",
			})
		}
	}

	return validationErrors, nil
}
