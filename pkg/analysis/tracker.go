// Package analysis provides tools for analyzing IDML documents and tracking dependencies.
// This is used for IDMS export to collect all resources needed for selected page items.
package analysis

import (
	"github.com/dimelords/idmllib/pkg/idml"
	"github.com/dimelords/idmllib/pkg/spread"
	"github.com/dimelords/idmllib/pkg/story"
)

// DependencySet tracks all dependencies for a set of selected page items.
// This includes stories, styles, colors, fonts, and other resources needed
// to export the selection as a standalone IDMS snippet.
type DependencySet struct {
	// Stories tracks referenced story files by their filename
	// Key: story filename (e.g., "Stories/Story_u1d8.xml")
	Stories map[string]bool

	// ParagraphStyles tracks referenced paragraph style IDs
	// Key: paragraph style ID (e.g., "ParagraphStyle/$ID/[No paragraph style]")
	ParagraphStyles map[string]bool

	// CharacterStyles tracks referenced character style IDs
	// Key: character style ID (e.g., "CharacterStyle/$ID/[No character style]")
	CharacterStyles map[string]bool

	// ObjectStyles tracks referenced object style IDs
	// Key: object style ID (e.g., "ObjectStyle/$ID/[Normal]")
	ObjectStyles map[string]bool

	// Colors tracks referenced color IDs
	// Key: color ID (e.g., "Color/Black")
	Colors map[string]bool

	// Swatches tracks referenced swatch IDs
	// Key: swatch ID
	Swatches map[string]bool

	// Fonts tracks referenced font families
	// Key: font family name (e.g., "Minion Pro")
	Fonts map[string]bool

	// Layers tracks referenced layer IDs
	// Key: layer ID
	Layers map[string]bool

	// Links tracks referenced external file links (for images)
	// Key: link ID or URI
	Links map[string]bool

	// ColorSpaces tracks referenced color spaces (RGB, CMYK, Lab, etc.)
	// Key: color space name
	ColorSpaces map[string]bool
}

// NewDependencySet creates a new empty DependencySet.
func NewDependencySet() *DependencySet {
	return &DependencySet{
		Stories:         make(map[string]bool),
		ParagraphStyles: make(map[string]bool),
		CharacterStyles: make(map[string]bool),
		ObjectStyles:    make(map[string]bool),
		Colors:          make(map[string]bool),
		Swatches:        make(map[string]bool),
		Fonts:           make(map[string]bool),
		Layers:          make(map[string]bool),
		Links:           make(map[string]bool),
		ColorSpaces:     make(map[string]bool),
	}
}

// DependencyTracker analyzes IDML elements and tracks their dependencies.
type DependencyTracker struct {
	// deps is the dependency set being populated
	deps *DependencySet

	// pkg is the IDML package being analyzed
	pkg *idml.Package
}

// NewDependencyTracker creates a new DependencyTracker for the given package.
func NewDependencyTracker(pkg *idml.Package) *DependencyTracker {
	return &DependencyTracker{
		deps: NewDependencySet(),
		pkg:  pkg,
	}
}

// Dependencies returns the collected dependency set.
func (dt *DependencyTracker) Dependencies() *DependencySet {
	return dt.deps
}

// AnalyzeTextFrame analyzes a text frame and tracks all its dependencies.
// This includes:
// - The parent story
// - Object style applied to the frame
// - Layer the frame is on
func (dt *DependencyTracker) AnalyzeTextFrame(tf *spread.SpreadTextFrame) error {
	// Track the parent story
	if tf.ParentStory != "" {
		// Story references are typically in the format "u1d8"
		// We need to convert this to the full filename
		storyFilename := "Stories/Story_" + tf.ParentStory + ".xml"
		dt.deps.Stories[storyFilename] = true

		// Analyze the story content to find style dependencies
		story, err := dt.pkg.Story(storyFilename)
		if err == nil {
			if err := dt.AnalyzeStory(story); err != nil {
				// Don't fail - just skip this story
				// The story might not exist in the package
			}
		}
	}

	// Track the applied object style
	if tf.AppliedObjectStyle != "" {
		dt.deps.ObjectStyles[tf.AppliedObjectStyle] = true
	}

	// Track the layer
	if tf.ItemLayer != "" {
		dt.deps.Layers[tf.ItemLayer] = true
	}

	return nil
}

// AnalyzeStory analyzes a story and tracks all style dependencies.
// This includes:
// - Paragraph styles used in the story
// - Character styles used in the story
// - Fonts referenced by the styles (future enhancement)
// - Colors used in the styles (future enhancement)
func (dt *DependencyTracker) AnalyzeStory(story *story.Story) error {
	// Analyze each paragraph style range
	for _, psr := range story.StoryElement.ParagraphStyleRanges {
		// Track the paragraph style
		if psr.AppliedParagraphStyle != "" {
			dt.deps.ParagraphStyles[psr.AppliedParagraphStyle] = true
		}

		// Analyze each character style range within the paragraph
		for _, csr := range psr.CharacterStyleRanges {
			// Track the character style
			if csr.AppliedCharacterStyle != "" {
				dt.deps.CharacterStyles[csr.AppliedCharacterStyle] = true
			}
		}
	}

	return nil
}

// AnalyzeRectangle analyzes a rectangle and tracks all its dependencies.
// This includes:
// - Object style applied to the rectangle
// - Layer the rectangle is on
// - Image and its dependencies if the rectangle contains an image
//
// Note: Rectangle doesn't have direct StrokeColor/FillColor attributes.
// Colors are inherited from AppliedObjectStyle or defined in Properties element.
func (dt *DependencyTracker) AnalyzeRectangle(rect *spread.Rectangle) error {
	// Track the applied object style
	if rect.AppliedObjectStyle != "" {
		dt.deps.ObjectStyles[rect.AppliedObjectStyle] = true
	}

	// Track the layer
	if rect.ItemLayer != "" {
		dt.deps.Layers[rect.ItemLayer] = true
	}

	// Analyze the image if present
	if rect.Image != nil {
		if err := dt.AnalyzeImage(rect.Image); err != nil {
			return err
		}
	}

	return nil
}

// AnalyzeImage analyzes an image and tracks all its dependencies.
// This includes:
// - Object style applied to the image
// - Link to external file
// - Color space used by the image
func (dt *DependencyTracker) AnalyzeImage(img *spread.Image) error {
	// Track the applied object style on the image itself
	if img.AppliedObjectStyle != "" {
		dt.deps.ObjectStyles[img.AppliedObjectStyle] = true
	}

	// Track the color space
	if img.Space != "" {
		dt.deps.ColorSpaces[img.Space] = true
	}

	// Track the link if present
	if img.Link != nil {
		if img.Link.Self != "" {
			dt.deps.Links[img.Link.Self] = true
		}
		if img.Link.LinkResourceURI != "" {
			dt.deps.Links[img.Link.LinkResourceURI] = true
		}
	}

	return nil
}

// AnalyzeOval analyzes an oval and tracks all its dependencies.
// This includes:
// - Object style applied to the oval
// - Layer the oval is on
// - Image and its dependencies if the oval contains an image
// - Colors used in stroke and fill
func (dt *DependencyTracker) AnalyzeOval(oval *spread.Oval) error {
	// Track the applied object style
	if oval.AppliedObjectStyle != "" {
		dt.deps.ObjectStyles[oval.AppliedObjectStyle] = true
	}

	// Track the layer
	if oval.ItemLayer != "" {
		dt.deps.Layers[oval.ItemLayer] = true
	}

	// Track stroke and fill colors
	if oval.StrokeColor != "" {
		dt.deps.Colors[oval.StrokeColor] = true
	}
	if oval.FillColor != "" {
		dt.deps.Colors[oval.FillColor] = true
	}

	// Analyze the image if present
	if oval.Image != nil {
		if err := dt.AnalyzeImage(oval.Image); err != nil {
			return err
		}
	}

	return nil
}

// AnalyzePolygon analyzes a polygon and tracks all its dependencies.
// This includes:
// - Object style applied to the polygon
// - Layer the polygon is on
// - Image and its dependencies if the polygon contains an image
// - Colors used in stroke and fill
func (dt *DependencyTracker) AnalyzePolygon(polygon *spread.Polygon) error {
	// Track the applied object style
	if polygon.AppliedObjectStyle != "" {
		dt.deps.ObjectStyles[polygon.AppliedObjectStyle] = true
	}

	// Track the layer
	if polygon.ItemLayer != "" {
		dt.deps.Layers[polygon.ItemLayer] = true
	}

	// Track stroke and fill colors
	if polygon.StrokeColor != "" {
		dt.deps.Colors[polygon.StrokeColor] = true
	}
	if polygon.FillColor != "" {
		dt.deps.Colors[polygon.FillColor] = true
	}

	// Analyze the image if present
	if polygon.Image != nil {
		if err := dt.AnalyzeImage(polygon.Image); err != nil {
			return err
		}
	}

	return nil
}

// AnalyzeGraphicLine analyzes a graphic line and tracks all its dependencies.
// This includes:
// - Object style applied to the line
// - Layer the line is on
// - Colors used in stroke and fill
func (dt *DependencyTracker) AnalyzeGraphicLine(line *spread.GraphicLine) error {
	// Track the applied object style
	if line.AppliedObjectStyle != "" {
		dt.deps.ObjectStyles[line.AppliedObjectStyle] = true
	}

	// Track the layer
	if line.ItemLayer != "" {
		dt.deps.Layers[line.ItemLayer] = true
	}

	// Track stroke color
	if line.StrokeColor != "" {
		dt.deps.Colors[line.StrokeColor] = true
	}

	// Track fill color (rare for lines, but possible)
	if line.FillColor != "" {
		dt.deps.Colors[line.FillColor] = true
	}

	return nil
}

// AnalyzeGroup analyzes a group and tracks all its dependencies.
// This includes:
// - Object style applied to the group
// - Layer the group is on
// Note: Group contents are typically analyzed separately
func (dt *DependencyTracker) AnalyzeGroup(group *spread.Group) error {
	// Track the applied object style
	if group.AppliedObjectStyle != "" {
		dt.deps.ObjectStyles[group.AppliedObjectStyle] = true
	}

	// Track the layer
	if group.ItemLayer != "" {
		dt.deps.Layers[group.ItemLayer] = true
	}

	return nil
}

// AnalyzeSelection analyzes an entire selection and tracks all dependencies.
// This is a convenience method that calls the appropriate Analyze* method
// for each element in the selection.
func (dt *DependencyTracker) AnalyzeSelection(sel *idml.Selection) error {
	// Analyze all text frames
	for _, tf := range sel.TextFrames {
		if err := dt.AnalyzeTextFrame(tf); err != nil {
			return err
		}
	}

	// Analyze all rectangles
	for _, rect := range sel.Rectangles {
		if err := dt.AnalyzeRectangle(rect); err != nil {
			return err
		}
	}

	// Analyze all ovals
	for _, oval := range sel.Ovals {
		if err := dt.AnalyzeOval(oval); err != nil {
			return err
		}
	}

	// Analyze all polygons
	for _, polygon := range sel.Polygons {
		if err := dt.AnalyzePolygon(polygon); err != nil {
			return err
		}
	}

	// Analyze all graphic lines
	for _, line := range sel.GraphicLines {
		if err := dt.AnalyzeGraphicLine(line); err != nil {
			return err
		}
	}

	// Analyze all groups
	for _, group := range sel.Groups {
		if err := dt.AnalyzeGroup(group); err != nil {
			return err
		}
	}

	return nil
}

// ResolveStyleHierarchies walks through all collected style dependencies and adds their parent styles.
// This ensures that when exporting an IDMS, all styles in the inheritance chain are included.
// For example, if a paragraph style is based on another style, both must be included.
//
// This method handles:
// - Paragraph style inheritance (BasedOn relationships)
// - Character style inheritance (BasedOn relationships)
// - Object style inheritance (BasedOn relationships)
// - Circular reference detection (to prevent infinite loops)
// - Multi-level inheritance (grandparent styles, etc.)
func (dt *DependencyTracker) ResolveStyleHierarchies() error {
	// Get the Styles resource file
	stylesResource, err := dt.pkg.Resource("Resources/Styles.xml")
	if err != nil {
		// If no Styles file, nothing to resolve
		return nil
	}

	// Parse style hierarchy information
	styleInfos, err := idml.ParseStylesForHierarchy(stylesResource.RawContent)
	if err != nil {
		return err
	}

	// Build style hierarchy maps for quick lookups
	styleParents := make(map[string]string) // styleID -> parentStyleID
	for _, info := range styleInfos {
		if info.BasedOn != "" {
			styleParents[info.Self] = info.BasedOn
		}
	}

	// Resolve paragraph style hierarchies
	paragraphStylesToResolve := make([]string, 0, len(dt.deps.ParagraphStyles))
	for styleID := range dt.deps.ParagraphStyles {
		paragraphStylesToResolve = append(paragraphStylesToResolve, styleID)
	}
	for _, styleID := range paragraphStylesToResolve {
		if err := dt.resolveStyleChain(styleID, styleParents, dt.deps.ParagraphStyles); err != nil {
			return err
		}
	}

	// Resolve character style hierarchies
	characterStylesToResolve := make([]string, 0, len(dt.deps.CharacterStyles))
	for styleID := range dt.deps.CharacterStyles {
		characterStylesToResolve = append(characterStylesToResolve, styleID)
	}
	for _, styleID := range characterStylesToResolve {
		if err := dt.resolveStyleChain(styleID, styleParents, dt.deps.CharacterStyles); err != nil {
			return err
		}
	}

	// Resolve object style hierarchies
	objectStylesToResolve := make([]string, 0, len(dt.deps.ObjectStyles))
	for styleID := range dt.deps.ObjectStyles {
		objectStylesToResolve = append(objectStylesToResolve, styleID)
	}
	for _, styleID := range objectStylesToResolve {
		if err := dt.resolveStyleChain(styleID, styleParents, dt.deps.ObjectStyles); err != nil {
			return err
		}
	}

	return nil
}

// resolveStyleChain recursively walks up the style hierarchy and adds all parent styles.
// It handles circular references by tracking visited styles.
func (dt *DependencyTracker) resolveStyleChain(styleID string, styleParents map[string]string, targetMap map[string]bool) error {
	// Track visited styles to detect circular references
	visited := make(map[string]bool)
	current := styleID

	for {
		// Check if we've seen this style before (circular reference)
		if visited[current] {
			// Circular reference detected - stop here
			break
		}
		visited[current] = true

		// Get the parent style
		parent, hasParent := styleParents[current]
		if !hasParent {
			// No parent - we've reached the top of the hierarchy
			break
		}

		// Check if the parent is a built-in InDesign style (starts with $ID/)
		// These are always available and don't need to be included in dependencies
		if len(parent) > 4 && parent[:4] == "$ID/" {
			// Built-in style - stop here
			break
		}

		// Add the parent style to dependencies
		targetMap[parent] = true

		// Move up to the next parent
		current = parent
	}

	return nil
}

// Summary returns a summary of the dependencies found.
func (dt *DependencyTracker) Summary() DependencySummary {
	return DependencySummary{
		StoriesCount:         len(dt.deps.Stories),
		ParagraphStylesCount: len(dt.deps.ParagraphStyles),
		CharacterStylesCount: len(dt.deps.CharacterStyles),
		ObjectStylesCount:    len(dt.deps.ObjectStyles),
		ColorsCount:          len(dt.deps.Colors),
		SwatchesCount:        len(dt.deps.Swatches),
		FontsCount:           len(dt.deps.Fonts),
		LayersCount:          len(dt.deps.Layers),
		LinksCount:           len(dt.deps.Links),
		ColorSpacesCount:     len(dt.deps.ColorSpaces),
	}
}

// DependencySummary provides a count of each type of dependency.
type DependencySummary struct {
	StoriesCount         int
	ParagraphStylesCount int
	CharacterStylesCount int
	ObjectStylesCount    int
	ColorsCount          int
	SwatchesCount        int
	FontsCount           int
	LayersCount          int
	LinksCount           int
	ColorSpacesCount     int
}
