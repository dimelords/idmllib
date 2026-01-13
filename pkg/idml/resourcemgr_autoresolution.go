package idml

import (
	"fmt"

	"github.com/dimelords/idmllib/v2/pkg/common"
	"github.com/dimelords/idmllib/v2/pkg/resources"
)

// ============================================================================
// Phase 3: Auto-Resolution - Adding Missing Resources
// ============================================================================

// AddMissingResources automatically adds missing resources to the package.
// This is useful when importing content from another document or when
// programmatically creating content that references resources.
//
// The method first identifies all missing resources, then creates default
// versions of each missing resource type according to the ValidationOptions.
//
// Returns an error if resource creation fails.
func (rm *ResourceManager) AddMissingResources(opts ValidationOptions) error {
	// Find all missing resources
	missing, err := rm.FindMissingResources()
	if err != nil {
		return common.WrapError("idml", "add missing resources", fmt.Errorf("failed to find missing resources: %w", err))
	}

	if !missing.HasMissing() {
		return nil // Nothing to do
	}

	// Add missing fonts if requested
	if opts.EnsureFontsExist && len(missing.Fonts) > 0 {
		if err := rm.addMissingFonts(missing.Fonts); err != nil {
			return common.WrapError("idml", "add missing fonts", err)
		}
	}

	// Add missing paragraph styles if requested
	if opts.EnsureStylesExist && len(missing.ParagraphStyles) > 0 {
		if err := rm.addMissingParagraphStyles(missing.ParagraphStyles); err != nil {
			return common.WrapError("idml", "add missing paragraph styles", err)
		}
	}

	// Add missing character styles if requested
	if opts.EnsureStylesExist && len(missing.CharacterStyles) > 0 {
		if err := rm.addMissingCharacterStyles(missing.CharacterStyles); err != nil {
			return common.WrapError("idml", "add missing character styles", err)
		}
	}

	// Add missing object styles if requested
	if opts.EnsureStylesExist && len(missing.ObjectStyles) > 0 {
		if err := rm.addMissingObjectStyles(missing.ObjectStyles); err != nil {
			return common.WrapError("idml", "add missing object styles", err)
		}
	}

	// Add missing colors if requested
	if opts.EnsureColorsExist && len(missing.Colors) > 0 {
		if err := rm.addMissingColors(missing.Colors); err != nil {
			return common.WrapError("idml", "add missing colors", err)
		}
	}

	// Add missing swatches if requested
	if opts.EnsureColorsExist && len(missing.Swatches) > 0 {
		if err := rm.addMissingSwatches(missing.Swatches); err != nil {
			return common.WrapError("idml", "add missing swatches", err)
		}
	}

	// Add missing layers if requested
	if opts.EnsureLayersExist && len(missing.Layers) > 0 {
		if err := rm.addMissingLayers(missing.Layers); err != nil {
			return common.WrapError("idml", "add missing layers", err)
		}
	}

	return nil
}

// addMissingFonts adds default font definitions for missing fonts.
func (rm *ResourceManager) addMissingFonts(fontFamilies map[string][]string) error {
	// Get or create the fonts file
	fonts, err := rm.getOrCreateFontsFile()
	if err != nil {
		return err
	}

	// Add each missing font
	for fontFamily := range fontFamilies {
		newFont := createDefaultFontFamily(fontFamily)
		fonts.FontFamilies = append(fonts.FontFamilies, newFont)
	}

	// Update and marshal the fonts file
	return rm.updateFontsFile(fonts)
}

// getOrCreateFontsFile gets the existing fonts file or creates a new one.
func (rm *ResourceManager) getOrCreateFontsFile() (*resources.FontsFile, error) {
	fonts, err := rm.pkg.Fonts()
	if err != nil {
		// If Fonts.xml doesn't exist, create a new one
		fonts = &resources.FontsFile{
			DOMVersion:   "20.4",
			FontFamilies: []resources.FontFamily{},
		}
	}
	return fonts, nil
}

// createDefaultFontFamily creates a default font family definition.
func createDefaultFontFamily(fontFamily string) resources.FontFamily {
	return resources.FontFamily{
		Self: "FontFamily/" + fontFamily,
		Name: fontFamily,
		// Add minimal Font entry - InDesign will use system font
		Fonts: []resources.Font{
			{
				Self:           "Font/" + fontFamily + "$Regular",
				Name:           "Regular",
				FontFamily:     fontFamily,
				PostScriptName: fontFamily,
				FontStyleName:  "Regular",
				Status:         "Installed",
				FontType:       "TrueType",
			},
		},
	}
}

// updateFontsFile marshals and updates the fonts file in the package.
func (rm *ResourceManager) updateFontsFile(fonts *resources.FontsFile) error {
	// Update the cached fonts
	rm.pkg.SetFonts(fonts)

	// Marshal and update the file entry
	data, err := resources.MarshalFontsFile(fonts)
	if err != nil {
		return common.WrapError("idml", "add missing fonts", fmt.Errorf("failed to marshal fonts: %w", err))
	}

	return rm.updateFileEntry(PathFonts, data)
}

// addMissingParagraphStyles adds default paragraph style definitions.
func (rm *ResourceManager) addMissingParagraphStyles(styleIDs map[string][]string) error {
	// Get or create the styles file
	styles, err := rm.getOrCreateStylesFile()
	if err != nil {
		return err
	}

	// Ensure the root paragraph style group exists
	if styles.RootParagraphStyleGroup == nil {
		styles.RootParagraphStyleGroup = &resources.ParagraphStyleGroup{
			ParagraphStyles: []resources.ParagraphStyle{},
		}
	}

	// Add each missing style
	for styleID := range styleIDs {
		newStyle := createDefaultParagraphStyle(styleID)
		styles.RootParagraphStyleGroup.ParagraphStyles = append(
			styles.RootParagraphStyleGroup.ParagraphStyles,
			newStyle,
		)
	}

	// Update and marshal the styles file
	return rm.updateStylesFile(styles)
}

// addMissingCharacterStyles adds default character style definitions.
func (rm *ResourceManager) addMissingCharacterStyles(styleIDs map[string][]string) error {
	// Get or create the styles file
	styles, err := rm.getOrCreateStylesFile()
	if err != nil {
		return err
	}

	// Ensure the root character style group exists
	if styles.RootCharacterStyleGroup == nil {
		styles.RootCharacterStyleGroup = &resources.CharacterStyleGroup{
			CharacterStyles: []resources.CharacterStyle{},
		}
	}

	// Add each missing style
	for styleID := range styleIDs {
		newStyle := createDefaultCharacterStyle(styleID)
		styles.RootCharacterStyleGroup.CharacterStyles = append(
			styles.RootCharacterStyleGroup.CharacterStyles,
			newStyle,
		)
	}

	// Update and marshal the styles file
	return rm.updateStylesFile(styles)
}

// addMissingObjectStyles adds default object style definitions.
func (rm *ResourceManager) addMissingObjectStyles(styleIDs map[string][]string) error {
	// Get or create the styles file
	styles, err := rm.getOrCreateStylesFile()
	if err != nil {
		return err
	}

	// Ensure the root object style group exists
	if styles.RootObjectStyleGroup == nil {
		styles.RootObjectStyleGroup = &resources.ObjectStyleGroup{
			ObjectStyles: []resources.ObjectStyle{},
		}
	}

	// Add each missing style
	for styleID := range styleIDs {
		newStyle := createDefaultObjectStyle(styleID)
		styles.RootObjectStyleGroup.ObjectStyles = append(
			styles.RootObjectStyleGroup.ObjectStyles,
			newStyle,
		)
	}

	// Update and marshal the styles file
	return rm.updateStylesFile(styles)
}

// getOrCreateStylesFile gets the existing styles file or creates a new one.
func (rm *ResourceManager) getOrCreateStylesFile() (*resources.StylesFile, error) {
	styles, err := rm.pkg.Styles()
	if err != nil {
		// If Styles.xml doesn't exist, create a new one
		styles = &resources.StylesFile{
			DOMVersion: "20.4",
			RootParagraphStyleGroup: &resources.ParagraphStyleGroup{
				ParagraphStyles: []resources.ParagraphStyle{},
			},
			RootCharacterStyleGroup: &resources.CharacterStyleGroup{
				CharacterStyles: []resources.CharacterStyle{},
			},
		}
	}
	return styles, nil
}

// createDefaultParagraphStyle creates a default paragraph style.
func createDefaultParagraphStyle(styleID string) resources.ParagraphStyle {
	return resources.ParagraphStyle{
		Self: styleID,
		Name: extractStyleName(styleID),
		// InDesign will use default formatting
		// Leave other fields empty - they will use defaults
	}
}

// createDefaultCharacterStyle creates a default character style.
func createDefaultCharacterStyle(styleID string) resources.CharacterStyle {
	return resources.CharacterStyle{
		Self: styleID,
		Name: extractStyleName(styleID),
		// InDesign will use default formatting
	}
}

// createDefaultObjectStyle creates a default object style.
func createDefaultObjectStyle(styleID string) resources.ObjectStyle {
	return resources.ObjectStyle{
		Self: styleID,
		Name: extractStyleName(styleID),
		// InDesign will use default formatting
	}
}

// updateStylesFile marshals and updates the styles file in the package.
func (rm *ResourceManager) updateStylesFile(styles *resources.StylesFile) error {
	// Update the cached styles
	rm.pkg.SetStyles(styles)

	// Marshal and update the file entry
	data, err := resources.MarshalStylesFile(styles)
	if err != nil {
		return common.WrapError("idml", "update styles", fmt.Errorf("failed to marshal styles: %w", err))
	}

	return rm.updateFileEntry(PathStyles, data)
}

// addMissingColors adds default color definitions.
func (rm *ResourceManager) addMissingColors(colorRefs map[string][]string) error {
	// Get or create the graphics file
	graphics, err := rm.getOrCreateGraphicsFile()
	if err != nil {
		return err
	}

	// Add each missing color
	for colorRef := range colorRefs {
		newColor := createDefaultColor(colorRef)
		graphics.Colors = append(graphics.Colors, newColor)
	}

	// Update and marshal the graphics file
	return rm.updateGraphicsFile(graphics)
}

// addMissingSwatches adds default swatch definitions.
func (rm *ResourceManager) addMissingSwatches(swatchRefs map[string][]string) error {
	// Get or create the graphics file
	graphics, err := rm.getOrCreateGraphicsFile()
	if err != nil {
		return err
	}

	// Add each missing swatch
	for swatchRef := range swatchRefs {
		newSwatch := createDefaultSwatch(swatchRef)
		graphics.Swatches = append(graphics.Swatches, newSwatch)
	}

	// Update and marshal the graphics file
	return rm.updateGraphicsFile(graphics)
}

// getOrCreateGraphicsFile gets the existing graphics file or creates a new one.
func (rm *ResourceManager) getOrCreateGraphicsFile() (*resources.GraphicFile, error) {
	graphics, err := rm.pkg.Graphics()
	if err != nil {
		// If Graphic.xml doesn't exist, create a new one
		graphics = &resources.GraphicFile{
			DOMVersion: "1.0",
			Colors:     []resources.Color{},
			Swatches:   []resources.Swatch{},
		}
	}
	return graphics, nil
}

// createDefaultColor creates a default color definition.
func createDefaultColor(colorRef string) resources.Color {
	// Extract color ID (remove "Color/" prefix)
	colorID := colorRef
	if len(colorRef) > 6 && colorRef[:6] == "Color/" {
		colorID = colorRef[6:]
	}

	// Create a default color (black in CMYK)
	return resources.Color{
		Self:       colorID,
		Model:      "Process",
		Space:      "CMYK",
		ColorValue: "0 0 0 100", // Black in CMYK as string
		Name:       colorID,
	}
}

// createDefaultSwatch creates a default swatch definition.
func createDefaultSwatch(swatchRef string) resources.Swatch {
	// Extract swatch ID (remove "Swatch/" or "Color/" prefix)
	swatchID := swatchRef
	if len(swatchRef) > 7 && swatchRef[:7] == "Swatch/" {
		swatchID = swatchRef[7:]
	} else if len(swatchRef) > 6 && swatchRef[:6] == "Color/" {
		swatchID = swatchRef[6:]
	}

	// Create a default swatch
	return resources.Swatch{
		Self: swatchID,
		Name: swatchID,
	}
}

// updateGraphicsFile marshals and updates the graphics file in the package.
func (rm *ResourceManager) updateGraphicsFile(graphics *resources.GraphicFile) error {
	// Update the cached graphics
	rm.pkg.SetGraphics(graphics)

	// Marshal and update the file entry
	data, err := resources.MarshalGraphicFile(graphics)
	if err != nil {
		return common.WrapError("idml", "update graphics", fmt.Errorf("failed to marshal graphics: %w", err))
	}

	return rm.updateFileEntry(PathGraphic, data)
}

// updateFileEntry is a helper method to update or create a file entry in the package.
func (rm *ResourceManager) updateFileEntry(path string, data []byte) error {
	// Update the file entry
	if entry, exists := rm.pkg.files[path]; exists {
		entry.data = data
	} else {
		rm.pkg.files[path] = &fileEntry{data: data}
		// Add to file order if new
		rm.pkg.fileOrder = append(rm.pkg.fileOrder, path)
	}
	return nil
}

// addMissingLayers adds default layer definitions to spreads.
// NOTE: Layer creation is currently disabled as layers are not stored in SpreadElement.
// This will be re-enabled once we understand where layers are stored in IDML.
func (rm *ResourceManager) addMissingLayers(layerIDs map[string][]string) error {
	// Layer creation not yet implemented.
	// See: https://app.clickup.com/t/86c6nxkk2
	return nil
}

// Helper functions for extracting names from IDs

// extractStyleName extracts a display name from a style ID.
// Example: "ParagraphStyle/$ID/MyStyle" -> "MyStyle"
func extractStyleName(styleID string) string {
	// Look for the last "/" in the ID
	for i := len(styleID) - 1; i >= 0; i-- {
		if styleID[i] == '/' {
			return styleID[i+1:]
		}
	}
	// If no "/" found, return the whole ID
	return styleID
}

// extractLayerName extracts a display name from a layer ID.
// Example: "u123" -> "Layer u123"
// extractLayerName extracts the layer name from a layer ID
// This function is currently unused but kept for potential future use
// nolint:unused
func extractLayerName(layerID string) string {
	return "Layer " + layerID
}

// addMissingObjectStyle adds a default object style with the given ID to the styles file.
// This is a helper method for TextFrame and Rectangle operations that need to ensure
// object styles exist.
func (rm *ResourceManager) addMissingObjectStyle(styleID string) error {
	styles, err := rm.pkg.Styles()
	if err != nil {
		return err
	}

	// Ensure RootObjectStyleGroup exists
	if styles.RootObjectStyleGroup == nil {
		styles.RootObjectStyleGroup = &resources.ObjectStyleGroup{
			Self:         "RootObjectStyleGroup",
			ObjectStyles: []resources.ObjectStyle{},
		}
	}

	// Extract a display name from the styleID
	styleName := extractStyleName(styleID)

	// Create a minimal default object style
	newStyle := resources.ObjectStyle{
		Self: styleID,
		Name: styleName,
		// Add other essential fields with defaults if needed
	}

	// Add to the styles file
	styles.RootObjectStyleGroup.ObjectStyles = append(styles.RootObjectStyleGroup.ObjectStyles, newStyle)

	// Update the cached styles
	rm.pkg.SetStyles(styles)

	return nil
}

// findMissingObjectStyle checks if an object style exists in the styles file.
// Returns true if the style is missing, false if it exists.
func (rm *ResourceManager) findMissingObjectStyle(styleID string) (bool, error) {
	styles, err := rm.pkg.Styles()
	if err != nil {
		return true, err // If we can't load styles, assume it's missing
	}

	// Check if object style exists
	if styles.RootObjectStyleGroup != nil {
		for _, style := range styles.RootObjectStyleGroup.ObjectStyles {
			if style.Self == styleID {
				return false, nil // Found it
			}
		}
	}

	return true, nil // Not found
}
