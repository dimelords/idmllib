// Package idms provides functionality for working with Adobe InDesign IDMS (snippet) files.
// This file implements the IDMS exporter that can create minimal snippets from selected page items.
package idms

import (
	"encoding/xml"
	"fmt"

	"github.com/dimelords/idmllib/v2/pkg/analysis"
	"github.com/dimelords/idmllib/v2/pkg/common"
	"github.com/dimelords/idmllib/v2/pkg/document"
	"github.com/dimelords/idmllib/v2/pkg/idml"
	"github.com/dimelords/idmllib/v2/pkg/resources"
	"github.com/dimelords/idmllib/v2/pkg/story"
)

// Exporter builds IDMS snippet files from selected page items in an IDML document.
// It analyzes dependencies and extracts only the minimal set of resources needed.
type Exporter struct {
	// pkg is the source IDML package to export from
	pkg *idml.Package

	// deps tracks all dependencies for the selection
	deps *analysis.DependencySet

	// tracker is used to analyze element dependencies
	tracker *analysis.DependencyTracker
}

// NewExporter creates a new IDMS exporter for the given IDML package.
func NewExporter(pkg *idml.Package) *Exporter {
	return &Exporter{
		pkg:     pkg,
		deps:    analysis.NewDependencySet(),
		tracker: analysis.NewDependencyTracker(pkg),
	}
}

// ExportSelection exports the given selection as an IDMS package.
// It analyzes all dependencies and creates a minimal standalone snippet.
//
// Returns:
//   - *Package: The IDMS package containing the selected items and their dependencies
//   - error: Any error encountered during export
//
// The export process:
//  1. Validates the selection is not empty
//  2. Analyzes dependencies for all selected items
//  3. Extracts required resources (stories, styles, colors, fonts, etc.)
//  4. Builds a minimal IDMS package with only the needed resources
func (e *Exporter) ExportSelection(sel *idml.Selection) (*Package, error) {
	// Validate selection
	if sel == nil {
		return nil, common.WrapError("idms", "export selection", fmt.Errorf("selection cannot be nil"))
	}

	if sel.IsEmpty() {
		return nil, common.WrapError("idms", "export selection", fmt.Errorf("selection is empty"))
	}

	// Phase 4.2 - Analyze dependencies for all selected items
	if err := e.analyzeDependencies(sel); err != nil {
		return nil, common.WrapError("idms", "export selection", fmt.Errorf("failed to analyze dependencies: %w", err))
	}

	// Phase 4.2 - Extract referenced resources
	resources, err := e.extractAllResources()
	if err != nil {
		return nil, common.WrapError("idms", "export selection", fmt.Errorf("failed to extract resources: %w", err))
	}

	// Phase 4.3 - Build minimal IDMS package
	pkg, err := e.buildMinimalPackage(sel, resources)
	if err != nil {
		return nil, common.WrapError("idms", "export selection", fmt.Errorf("failed to build package: %w", err))
	}

	return pkg, nil
}

// analyzeDependencies analyzes all selected items and collects their dependencies.
// This is the foundation for Phase 4.2 - Resource Extraction.
func (e *Exporter) analyzeDependencies(sel *idml.Selection) error {
	// Analyze text frames
	for _, tf := range sel.TextFrames {
		if err := e.tracker.AnalyzeTextFrame(tf); err != nil {
			return fmt.Errorf("failed to analyze text frame %s: %w", tf.Self, err)
		}
	}

	// Analyze rectangles (often contain images)
	for _, rect := range sel.Rectangles {
		if err := e.tracker.AnalyzeRectangle(rect); err != nil {
			return fmt.Errorf("failed to analyze rectangle %s: %w", rect.Self, err)
		}
	}

	// Analyze ovals
	for _, oval := range sel.Ovals {
		if err := e.tracker.AnalyzeOval(oval); err != nil {
			return fmt.Errorf("failed to analyze oval %s: %w", oval.Self, err)
		}
	}

	// Analyze polygons
	for _, polygon := range sel.Polygons {
		if err := e.tracker.AnalyzePolygon(polygon); err != nil {
			return fmt.Errorf("failed to analyze polygon %s: %w", polygon.Self, err)
		}
	}

	// Analyze graphic lines
	for _, line := range sel.GraphicLines {
		if err := e.tracker.AnalyzeGraphicLine(line); err != nil {
			return fmt.Errorf("failed to analyze graphic line %s: %w", line.Self, err)
		}
	}

	// Analyze groups
	for _, group := range sel.Groups {
		if err := e.tracker.AnalyzeGroup(group); err != nil {
			return fmt.Errorf("failed to analyze group %s: %w", group.Self, err)
		}
	}

	// Store the collected dependencies
	e.deps = e.tracker.Dependencies()

	return nil
}

// Dependencies returns the collected dependency set from the last export operation.
// This is useful for debugging and testing.
func (e *Exporter) Dependencies() *analysis.DependencySet {
	return e.deps
}

// ============================================================================
// Phase 4.2: Resource Extraction
// ============================================================================

// extractReferencedStories extracts story files referenced in dependencies.
// Returns a map of story filename to Story struct.
func (e *Exporter) extractReferencedStories() (map[string]*story.Story, error) {
	stories := make(map[string]*story.Story)

	for storyFile := range e.deps.Stories {
		story, err := e.pkg.Story(storyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to get story %s: %w", storyFile, err)
		}
		stories[storyFile] = story
	}

	return stories, nil
}

// extractReferencedStyles extracts all referenced style definitions.
// This includes ParagraphStyles, CharacterStyles, and ObjectStyles with their inheritance chains.
// Returns the complete Styles resource file with only referenced styles.
func (e *Exporter) extractReferencedStyles() (*resources.StylesFile, error) {
	// Get the full styles file from source package
	srcStyles, err := e.pkg.Styles()
	if err != nil {
		return nil, fmt.Errorf("failed to get source styles: %w", err)
	}

	// Create new styles file with only referenced styles
	extracted := &resources.StylesFile{
		DOMVersion: srcStyles.DOMVersion,
		RootParagraphStyleGroup: &resources.ParagraphStyleGroup{
			Self:    "u79", // Required Self attribute for InDesign compatibility
			XMLName: xml.Name{Local: "RootParagraphStyleGroup"},
		},
		RootCharacterStyleGroup: &resources.CharacterStyleGroup{
			Self:    "u7a", // Required Self attribute for InDesign compatibility
			XMLName: xml.Name{Local: "RootCharacterStyleGroup"},
		},
		RootObjectStyleGroup: &resources.ObjectStyleGroup{
			Self:    "u8a", // Required Self attribute for InDesign compatibility
			XMLName: xml.Name{Local: "RootObjectStyleGroup"},
		},
	}

	// Extract paragraph styles with their nested group structure
	e.extractParagraphStylesWithGroups(srcStyles.RootParagraphStyleGroup, extracted.RootParagraphStyleGroup)

	// Extract character styles with their nested group structure
	e.extractCharacterStylesWithGroups(srcStyles.RootCharacterStyleGroup, extracted.RootCharacterStyleGroup)

	// Extract object styles with their nested group structure
	e.extractObjectStylesWithGroups(srcStyles.RootObjectStyleGroup, extracted.RootObjectStyleGroup)

	return extracted, nil
}

// findParagraphStyle recursively searches for a paragraph style in nested groups.
// findParagraphStyle finds a paragraph style in a style group
// This function is currently unused but kept for potential future use
// nolint:unused
func (e *Exporter) findParagraphStyle(group *resources.ParagraphStyleGroup, styleID string) *resources.ParagraphStyle {
	if group == nil {
		return nil
	}

	// Search in current group
	for i := range group.ParagraphStyles {
		if group.ParagraphStyles[i].Self == styleID {
			return &group.ParagraphStyles[i]
		}
	}

	// Search in nested groups
	for i := range group.NestedGroups {
		if style := e.findParagraphStyle(&group.NestedGroups[i], styleID); style != nil {
			return style
		}
	}

	return nil
}

// findCharacterStyle recursively searches for a character style in nested groups.
// findCharacterStyle finds a character style in a style group
// This function is currently unused but kept for potential future use
// nolint:unused
func (e *Exporter) findCharacterStyle(group *resources.CharacterStyleGroup, styleID string) *resources.CharacterStyle {
	if group == nil {
		return nil
	}

	// Search in current group
	for i := range group.CharacterStyles {
		if group.CharacterStyles[i].Self == styleID {
			return &group.CharacterStyles[i]
		}
	}

	// Search in nested groups
	for i := range group.NestedGroups {
		if style := e.findCharacterStyle(&group.NestedGroups[i], styleID); style != nil {
			return style
		}
	}

	return nil
}

// findObjectStyle recursively searches for an object style in nested groups.
// findObjectStyle finds an object style in a style group
// This function is currently unused but kept for potential future use
// nolint:unused
func (e *Exporter) findObjectStyle(group *resources.ObjectStyleGroup, styleID string) *resources.ObjectStyle {
	if group == nil {
		return nil
	}

	// Search in current group
	for i := range group.ObjectStyles {
		if group.ObjectStyles[i].Self == styleID {
			return &group.ObjectStyles[i]
		}
	}

	// Search in nested groups
	for i := range group.NestedGroups {
		if style := e.findObjectStyle(&group.NestedGroups[i], styleID); style != nil {
			return style
		}
	}

	return nil
}

// extractParagraphStylesWithGroups extracts paragraph styles while preserving nested group structure.
// Only extracts groups that contain referenced styles.
func (e *Exporter) extractParagraphStylesWithGroups(srcGroup, dstGroup *resources.ParagraphStyleGroup) {
	if srcGroup == nil || dstGroup == nil {
		return
	}

	// Extract styles at this level
	for i := range srcGroup.ParagraphStyles {
		style := &srcGroup.ParagraphStyles[i]
		if e.deps.ParagraphStyles[style.Self] {
			dstGroup.ParagraphStyles = append(dstGroup.ParagraphStyles, *style)
		}
	}

	// Recursively extract nested groups that contain referenced styles
	for i := range srcGroup.NestedGroups {
		srcNestedGroup := &srcGroup.NestedGroups[i]
		if e.groupContainsParagraphStyle(srcNestedGroup) {
			// Create a copy of the nested group (without styles/nested groups initially)
			dstNestedGroup := resources.ParagraphStyleGroup{
				XMLName: srcNestedGroup.XMLName,
				Self:    srcNestedGroup.Self,
				Name:    srcNestedGroup.Name,
			}
			// Recursively extract styles from this group
			e.extractParagraphStylesWithGroups(srcNestedGroup, &dstNestedGroup)
			dstGroup.NestedGroups = append(dstGroup.NestedGroups, dstNestedGroup)
		}
	}
}

// groupContainsParagraphStyle checks if a group or any of its nested groups contains a referenced paragraph style.
func (e *Exporter) groupContainsParagraphStyle(group *resources.ParagraphStyleGroup) bool {
	if group == nil {
		return false
	}

	// Check styles at this level
	for i := range group.ParagraphStyles {
		if e.deps.ParagraphStyles[group.ParagraphStyles[i].Self] {
			return true
		}
	}

	// Check nested groups
	for i := range group.NestedGroups {
		if e.groupContainsParagraphStyle(&group.NestedGroups[i]) {
			return true
		}
	}

	return false
}

// extractCharacterStylesWithGroups extracts character styles while preserving nested group structure.
// Only extracts groups that contain referenced styles.
func (e *Exporter) extractCharacterStylesWithGroups(srcGroup, dstGroup *resources.CharacterStyleGroup) {
	if srcGroup == nil || dstGroup == nil {
		return
	}

	// Extract styles at this level
	for i := range srcGroup.CharacterStyles {
		style := &srcGroup.CharacterStyles[i]
		if e.deps.CharacterStyles[style.Self] {
			dstGroup.CharacterStyles = append(dstGroup.CharacterStyles, *style)
		}
	}

	// Recursively extract nested groups that contain referenced styles
	for i := range srcGroup.NestedGroups {
		srcNestedGroup := &srcGroup.NestedGroups[i]
		if e.groupContainsCharacterStyle(srcNestedGroup) {
			// Create a copy of the nested group (without styles/nested groups initially)
			dstNestedGroup := resources.CharacterStyleGroup{
				XMLName: srcNestedGroup.XMLName,
				Self:    srcNestedGroup.Self,
				Name:    srcNestedGroup.Name,
			}
			// Recursively extract styles from this group
			e.extractCharacterStylesWithGroups(srcNestedGroup, &dstNestedGroup)
			dstGroup.NestedGroups = append(dstGroup.NestedGroups, dstNestedGroup)
		}
	}
}

// groupContainsCharacterStyle checks if a group or any of its nested groups contains a referenced character style.
func (e *Exporter) groupContainsCharacterStyle(group *resources.CharacterStyleGroup) bool {
	if group == nil {
		return false
	}

	// Check styles at this level
	for i := range group.CharacterStyles {
		if e.deps.CharacterStyles[group.CharacterStyles[i].Self] {
			return true
		}
	}

	// Check nested groups
	for i := range group.NestedGroups {
		if e.groupContainsCharacterStyle(&group.NestedGroups[i]) {
			return true
		}
	}

	return false
}

// extractObjectStylesWithGroups extracts object styles while preserving nested group structure.
// Only extracts groups that contain referenced styles.
func (e *Exporter) extractObjectStylesWithGroups(srcGroup, dstGroup *resources.ObjectStyleGroup) {
	if srcGroup == nil || dstGroup == nil {
		return
	}

	// Extract styles at this level
	for i := range srcGroup.ObjectStyles {
		style := &srcGroup.ObjectStyles[i]
		if e.deps.ObjectStyles[style.Self] {
			dstGroup.ObjectStyles = append(dstGroup.ObjectStyles, *style)
		}
	}

	// Recursively extract nested groups that contain referenced styles
	for i := range srcGroup.NestedGroups {
		srcNestedGroup := &srcGroup.NestedGroups[i]
		if e.groupContainsObjectStyle(srcNestedGroup) {
			// Create a copy of the nested group (without styles/nested groups initially)
			dstNestedGroup := resources.ObjectStyleGroup{
				XMLName: srcNestedGroup.XMLName,
				Self:    srcNestedGroup.Self,
				Name:    srcNestedGroup.Name,
			}
			// Recursively extract styles from this group
			e.extractObjectStylesWithGroups(srcNestedGroup, &dstNestedGroup)
			dstGroup.NestedGroups = append(dstGroup.NestedGroups, dstNestedGroup)
		}
	}
}

// groupContainsObjectStyle checks if a group or any of its nested groups contains a referenced object style.
func (e *Exporter) groupContainsObjectStyle(group *resources.ObjectStyleGroup) bool {
	if group == nil {
		return false
	}

	// Check styles at this level
	for i := range group.ObjectStyles {
		if e.deps.ObjectStyles[group.ObjectStyles[i].Self] {
			return true
		}
	}

	// Check nested groups
	for i := range group.NestedGroups {
		if e.groupContainsObjectStyle(&group.NestedGroups[i]) {
			return true
		}
	}

	return false
}

// extractReferencedColors extracts referenced color definitions from Graphics resource.
// Returns a GraphicFile with only referenced colors.
func (e *Exporter) extractReferencedColors() (*resources.GraphicFile, error) {
	// Get source graphics file
	srcGraphics, err := e.pkg.Graphics()
	if err != nil {
		return nil, fmt.Errorf("failed to get source graphics: %w", err)
	}

	// Create new graphics file with only referenced colors
	extracted := &resources.GraphicFile{
		DOMVersion: srcGraphics.DOMVersion,
	}

	// Extract referenced colors
	for colorID := range e.deps.Colors {
		for _, color := range srcGraphics.Colors {
			if color.Self == colorID {
				extracted.Colors = append(extracted.Colors, color)
				break
			}
		}
	}

	// Extract referenced swatches
	for swatchID := range e.deps.Swatches {
		for _, swatch := range srcGraphics.Swatches {
			if swatch.Self == swatchID {
				extracted.Swatches = append(extracted.Swatches, swatch)
				break
			}
		}
	}

	return extracted, nil
}

// extractReferencedLayers extracts referenced layer definitions.
// Currently returns nil as layer extraction will be implemented when needed.
func (e *Exporter) extractReferencedLayers() (map[string]*document.Layer, error) {
	// Layer extraction not yet implemented.
	// See: https://app.clickup.com/t/86c6nxkk2
	return nil, nil
}

// extractResources extracts all referenced resources from the source package.
// This is called during the export process after dependencies are analyzed.
type ExtractedResources struct {
	Stories  map[string]*story.Story
	Styles   *resources.StylesFile
	Graphics *resources.GraphicFile
	Layers   map[string]*document.Layer
}

// extractAllResources extracts all referenced resources based on collected dependencies.
func (e *Exporter) extractAllResources() (*ExtractedResources, error) {
	resources := &ExtractedResources{}

	// Extract stories
	stories, err := e.extractReferencedStories()
	if err != nil {
		return nil, fmt.Errorf("failed to extract stories: %w", err)
	}
	resources.Stories = stories

	// Extract styles (only if we have style dependencies)
	if len(e.deps.ParagraphStyles) > 0 || len(e.deps.CharacterStyles) > 0 || len(e.deps.ObjectStyles) > 0 {
		styles, err := e.extractReferencedStyles()
		if err != nil {
			return nil, fmt.Errorf("failed to extract styles: %w", err)
		}
		resources.Styles = styles
	}

	// Extract graphics (colors, swatches)
	if len(e.deps.Colors) > 0 || len(e.deps.Swatches) > 0 {
		graphics, err := e.extractReferencedColors()
		if err != nil {
			return nil, fmt.Errorf("failed to extract graphics: %w", err)
		}
		resources.Graphics = graphics
	}

	// Extract layers
	layers, err := e.extractReferencedLayers()
	if err != nil {
		return nil, fmt.Errorf("failed to extract layers: %w", err)
	}
	resources.Layers = layers

	return resources, nil
}

// ============================================================================
