package idml

import (
	"fmt"

	"github.com/dimelords/idmllib/v2/pkg/common"
	"github.com/dimelords/idmllib/v2/pkg/spread"
)

// selectPageItemByID is a generic helper that searches for a page item by ID across all spreads.
// It reduces code duplication in SelectXxxByID methods.
// NOTE: This is the fallback linear search. SelectXxxByID methods now use the index for O(1) lookups.
// selectPageItemByID is a generic function to select page items by ID
// This function is currently unused but kept for potential future use
// nolint:unused
func selectPageItemByID[T PageItem](p *Package, itemType string, getter func(*spread.SpreadElement) []T, getSelf func(*T) string, id string) (*T, error) {
	spreads, err := p.Spreads()
	if err != nil {
		return nil, common.WrapError("idml", "select "+itemType, fmt.Errorf("failed to load spreads: %w", err))
	}

	for _, sp := range spreads {
		items := getter(&sp.InnerSpread)
		for i := range items {
			if getSelf(&items[i]) == id {
				return &items[i], nil
			}
		}
	}

	return nil, common.WrapError("idml", "select "+itemType, fmt.Errorf("%s with ID '%s' not found", itemType, id))
}

// Selection represents a collection of selected page items from an IDML document.
// This is typically used for exporting subsets of a document (e.g., to IDMS snippets).
type Selection struct {
	// TextFrames contains selected text frames
	TextFrames []*spread.SpreadTextFrame

	// Rectangles contains selected rectangle frames (often containing images)
	Rectangles []*spread.Rectangle

	// Ovals contains selected oval frames
	Ovals []*spread.Oval

	// Polygons contains selected polygon frames
	Polygons []*spread.Polygon

	// GraphicLines contains selected graphic lines
	GraphicLines []*spread.GraphicLine

	// Groups contains selected groups
	Groups []*spread.Group
}

// NewSelection creates a new empty Selection.
func NewSelection() *Selection {
	return &Selection{
		TextFrames:   []*spread.SpreadTextFrame{},
		Rectangles:   []*spread.Rectangle{},
		Ovals:        []*spread.Oval{},
		Polygons:     []*spread.Polygon{},
		GraphicLines: []*spread.GraphicLine{},
		Groups:       []*spread.Group{},
	}
}

// IsEmpty returns true if the selection contains no elements.
func (s *Selection) IsEmpty() bool {
	return len(s.TextFrames) == 0 &&
		len(s.Rectangles) == 0 &&
		len(s.Ovals) == 0 &&
		len(s.Polygons) == 0 &&
		len(s.GraphicLines) == 0 &&
		len(s.Groups) == 0
}

// Count returns the total number of selected elements.
func (s *Selection) Count() int {
	return len(s.TextFrames) +
		len(s.Rectangles) +
		len(s.Ovals) +
		len(s.Polygons) +
		len(s.GraphicLines) +
		len(s.Groups)
}

// AddPageItem adds any page item to the selection using the PageItem interface.
// This method determines the concrete type and adds it to the appropriate slice.
func (s *Selection) AddPageItem(item PageItem) {
	switch v := item.(type) {
	case *spread.SpreadTextFrame:
		s.AddTextFrame(v)
	case *spread.Rectangle:
		s.AddRectangle(v)
	case *spread.Oval:
		s.AddOval(v)
	case *spread.Polygon:
		s.AddPolygon(v)
	case *spread.GraphicLine:
		s.AddGraphicLine(v)
	case *spread.Group:
		s.AddGroup(v)
	}
}

// GetAllPageItems returns all selected page items as a slice of PageItem interface.
// This enables polymorphic operations on the entire selection.
func (s *Selection) GetAllPageItems() []PageItem {
	var items []PageItem

	for i := range s.TextFrames {
		items = append(items, s.TextFrames[i])
	}
	for i := range s.Rectangles {
		items = append(items, s.Rectangles[i])
	}
	for i := range s.Ovals {
		items = append(items, s.Ovals[i])
	}
	for i := range s.Polygons {
		items = append(items, s.Polygons[i])
	}
	for i := range s.GraphicLines {
		items = append(items, s.GraphicLines[i])
	}
	for i := range s.Groups {
		items = append(items, s.Groups[i])
	}

	return items
}

// AddTextFrame adds a text frame to the selection.
func (s *Selection) AddTextFrame(tf *spread.SpreadTextFrame) {
	s.TextFrames = append(s.TextFrames, tf)
}

// AddRectangle adds a rectangle to the selection.
func (s *Selection) AddRectangle(rect *spread.Rectangle) {
	s.Rectangles = append(s.Rectangles, rect)
}

// AddOval adds an oval to the selection.
func (s *Selection) AddOval(oval *spread.Oval) {
	s.Ovals = append(s.Ovals, oval)
}

// AddPolygon adds a polygon to the selection.
func (s *Selection) AddPolygon(polygon *spread.Polygon) {
	s.Polygons = append(s.Polygons, polygon)
}

// AddGraphicLine adds a graphic line to the selection.
func (s *Selection) AddGraphicLine(line *spread.GraphicLine) {
	s.GraphicLines = append(s.GraphicLines, line)
}

// AddGroup adds a group to the selection.
func (s *Selection) AddGroup(group *spread.Group) {
	s.Groups = append(s.Groups, group)
}

// SelectPageItemByID finds and returns any page item by its Self ID using the PageItem interface.
// Uses an internal index for O(1) lookup performance.
// Returns an error if the page item is not found.
func (p *Package) SelectPageItemByID(id string) (PageItem, error) {
	if err := p.ensureItemIndex(); err != nil {
		return nil, common.WrapError("idml", "select page item", fmt.Errorf("failed to build index: %w", err))
	}

	idx := p.indexState.index

	// Check each type of page item
	if tf, ok := idx.textFrames[id]; ok {
		return tf, nil
	}
	if rect, ok := idx.rectangles[id]; ok {
		return rect, nil
	}
	if oval, ok := idx.ovals[id]; ok {
		return oval, nil
	}
	if poly, ok := idx.polygons[id]; ok {
		return poly, nil
	}
	if line, ok := idx.graphicLines[id]; ok {
		return line, nil
	}
	if group, ok := idx.groups[id]; ok {
		return group, nil
	}

	return nil, common.WrapError("idml", "select page item", fmt.Errorf("page item with ID '%s' not found", id))
}

// SelectTextFrameByID finds and returns a text frame by its Self ID.
// Uses an internal index for O(1) lookup performance.
// Returns an error if the text frame is not found.
func (p *Package) SelectTextFrameByID(id string) (*spread.SpreadTextFrame, error) {
	if err := p.ensureItemIndex(); err != nil {
		return nil, common.WrapError("idml", "select text frame", fmt.Errorf("failed to build index: %w", err))
	}

	if tf, ok := p.indexState.index.textFrames[id]; ok {
		return tf, nil
	}

	return nil, common.WrapError("idml", "select text frame", fmt.Errorf("text frame with ID '%s' not found", id))
}

// SelectRectangleByID finds and returns a rectangle by its Self ID.
// Uses an internal index for O(1) lookup performance.
// Returns an error if the rectangle is not found.
func (p *Package) SelectRectangleByID(id string) (*spread.Rectangle, error) {
	if err := p.ensureItemIndex(); err != nil {
		return nil, common.WrapError("idml", "select rectangle", fmt.Errorf("failed to build index: %w", err))
	}

	if rect, ok := p.indexState.index.rectangles[id]; ok {
		return rect, nil
	}

	return nil, common.WrapError("idml", "select rectangle", fmt.Errorf("rectangle with ID '%s' not found", id))
}

// SelectOvalByID finds and returns an oval by its Self ID.
// Uses an internal index for O(1) lookup performance.
// Returns an error if the oval is not found.
func (p *Package) SelectOvalByID(id string) (*spread.Oval, error) {
	if err := p.ensureItemIndex(); err != nil {
		return nil, common.WrapError("idml", "select oval", fmt.Errorf("failed to build index: %w", err))
	}

	if oval, ok := p.indexState.index.ovals[id]; ok {
		return oval, nil
	}

	return nil, common.WrapError("idml", "select oval", fmt.Errorf("oval with ID '%s' not found", id))
}

// SelectPolygonByID finds and returns a polygon by its Self ID.
// Uses an internal index for O(1) lookup performance.
// Returns an error if the polygon is not found.
func (p *Package) SelectPolygonByID(id string) (*spread.Polygon, error) {
	if err := p.ensureItemIndex(); err != nil {
		return nil, common.WrapError("idml", "select polygon", fmt.Errorf("failed to build index: %w", err))
	}

	if poly, ok := p.indexState.index.polygons[id]; ok {
		return poly, nil
	}

	return nil, common.WrapError("idml", "select polygon", fmt.Errorf("polygon with ID '%s' not found", id))
}

// SelectGraphicLineByID finds and returns a graphic line by its Self ID.
// Uses an internal index for O(1) lookup performance.
// Returns an error if the graphic line is not found.
func (p *Package) SelectGraphicLineByID(id string) (*spread.GraphicLine, error) {
	if err := p.ensureItemIndex(); err != nil {
		return nil, common.WrapError("idml", "select graphic line", fmt.Errorf("failed to build index: %w", err))
	}

	if line, ok := p.indexState.index.graphicLines[id]; ok {
		return line, nil
	}

	return nil, common.WrapError("idml", "select graphic line", fmt.Errorf("graphic line with ID '%s' not found", id))
}

// SelectGroupByID finds and returns a group by its Self ID.
// Uses an internal index for O(1) lookup performance.
// Returns an error if the group is not found.
func (p *Package) SelectGroupByID(id string) (*spread.Group, error) {
	if err := p.ensureItemIndex(); err != nil {
		return nil, common.WrapError("idml", "select group", fmt.Errorf("failed to build index: %w", err))
	}

	if group, ok := p.indexState.index.groups[id]; ok {
		return group, nil
	}

	return nil, common.WrapError("idml", "select group", fmt.Errorf("group with ID '%s' not found", id))
}

// SelectAllGraphicsInSpread returns all rectangles containing images in the specified spread.
// The spreadFilename should be in the format "Spreads/Spread_*.xml".
// Returns an error if the spread is not found or can't be loaded.
func (p *Package) SelectAllGraphicsInSpread(spreadFilename string) ([]*spread.Rectangle, error) {
	sp, err := p.Spread(spreadFilename)
	if err != nil {
		return nil, common.WrapErrorWithPath("idml", "select all graphics", spreadFilename, fmt.Errorf("failed to load spread '%s': %w", spreadFilename, err))
	}

	// Collect all rectangles that have images
	var graphics []*spread.Rectangle
	for i := range sp.InnerSpread.Rectangles {
		rect := &sp.InnerSpread.Rectangles[i]
		// Check if this rectangle has an image
		if rect.ContentType == "GraphicType" || rect.Image != nil {
			graphics = append(graphics, rect)
		}
	}

	return graphics, nil
}

// SelectAllTextFramesInSpread returns all text frames in the specified spread.
// The spreadFilename should be in the format "Spreads/Spread_*.xml".
// Returns an error if the spread is not found or can't be loaded.
func (p *Package) SelectAllTextFramesInSpread(spreadFilename string) ([]*spread.SpreadTextFrame, error) {
	sp, err := p.Spread(spreadFilename)
	if err != nil {
		return nil, common.WrapErrorWithPath("idml", "select all text frames", spreadFilename, fmt.Errorf("failed to load spread '%s': %w", spreadFilename, err))
	}

	// Collect all text frames
	textFrames := make([]*spread.SpreadTextFrame, len(sp.InnerSpread.TextFrames))
	for i := range sp.InnerSpread.TextFrames {
		textFrames[i] = &sp.InnerSpread.TextFrames[i]
	}

	return textFrames, nil
}

// SelectPageItemsByIDs creates a Selection containing all page items with the specified IDs.
// Uses an internal index for O(1) lookup per ID.
// Elements that are not found are silently skipped.
// Returns both a Selection (for backward compatibility) and a slice of PageItem interfaces.
func (p *Package) SelectPageItemsByIDs(ids ...string) (*Selection, []PageItem, error) {
	if err := p.ensureItemIndex(); err != nil {
		return nil, nil, common.WrapError("idml", "select page items by IDs", fmt.Errorf("failed to build index: %w", err))
	}

	selection := NewSelection()
	var pageItems []PageItem
	idx := p.indexState.index

	for _, id := range ids {
		// Check each type of page item
		if tf, ok := idx.textFrames[id]; ok {
			selection.AddTextFrame(tf)
			pageItems = append(pageItems, tf)
			continue
		}
		if rect, ok := idx.rectangles[id]; ok {
			selection.AddRectangle(rect)
			pageItems = append(pageItems, rect)
			continue
		}
		if oval, ok := idx.ovals[id]; ok {
			selection.AddOval(oval)
			pageItems = append(pageItems, oval)
			continue
		}
		if poly, ok := idx.polygons[id]; ok {
			selection.AddPolygon(poly)
			pageItems = append(pageItems, poly)
			continue
		}
		if line, ok := idx.graphicLines[id]; ok {
			selection.AddGraphicLine(line)
			pageItems = append(pageItems, line)
			continue
		}
		if group, ok := idx.groups[id]; ok {
			selection.AddGroup(group)
			pageItems = append(pageItems, group)
		}
		// If not found in any, silently skip (as documented)
	}

	return selection, pageItems, nil
}

// SelectByIDs creates a Selection containing all elements with the specified IDs.
// Uses an internal index for O(1) lookup per ID.
// Elements that are not found are silently skipped.
func (p *Package) SelectByIDs(ids ...string) (*Selection, error) {
	if err := p.ensureItemIndex(); err != nil {
		return nil, common.WrapError("idml", "select by IDs", fmt.Errorf("failed to build index: %w", err))
	}

	selection := NewSelection()
	idx := p.indexState.index

	for _, id := range ids {
		// Check each type of page item
		if tf, ok := idx.textFrames[id]; ok {
			selection.AddTextFrame(tf)
			continue
		}
		if rect, ok := idx.rectangles[id]; ok {
			selection.AddRectangle(rect)
			continue
		}
		if oval, ok := idx.ovals[id]; ok {
			selection.AddOval(oval)
			continue
		}
		if poly, ok := idx.polygons[id]; ok {
			selection.AddPolygon(poly)
			continue
		}
		if line, ok := idx.graphicLines[id]; ok {
			selection.AddGraphicLine(line)
			continue
		}
		if group, ok := idx.groups[id]; ok {
			selection.AddGroup(group)
		}
		// If not found in any, silently skip (as documented)
	}

	return selection, nil
}
