package idml

import (
	"github.com/dimelords/idmllib/v2/pkg/spread"
)

// Selection represents a collection of selected page items from an IDML document.
// This is typically used for exporting subsets of a document (e.g., to IDMS snippets).
type Selection struct {
	// TextFrames contains selected text frames
	TextFrames []*spread.TextFrame

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
		TextFrames:   []*spread.TextFrame{},
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
	case *spread.TextFrame:
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
func (s *Selection) AddTextFrame(tf *spread.TextFrame) {
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
