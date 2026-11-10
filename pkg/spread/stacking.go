package spread

import (
	"slices"

	"github.com/dimelords/idmllib/v3/pkg/common"
)

// Stacking order is the order of a spread's page items in its XML, bottom to top. A parsed spread keeps it (see
// docs/FIDELITY.md), but items appended through the typed slices are written after the recorded ones, grouped by
// kind: text frames, then rectangles, images, ovals, polygons, lines, groups and buttons. A text frame appended
// after a filled rectangle therefore ends up below it. StackingOrder and SetStackingOrder read and set the order
// across kinds.

// stackKinds are the typed page item kinds that can be restacked, in the order MarshalXML lists them.
var stackKinds = []string{"TextFrame", "Rectangle", "Image", "Oval", "Polygon", "GraphicLine", "Group", "Button"}

// childKinds are all child kinds in the order MarshalXML lists them, with their current counts.
func (s *Spread) childKinds() []common.ChildKind {
	fp := 0
	if s.FlattenerPreference != nil {
		fp = 1
	}
	return []common.ChildKind{
		{Name: "FlattenerPreference", Len: fp},
		{Name: "Page", Len: len(s.Pages)},
		{Name: "TextFrame", Len: len(s.TextFrames)},
		{Name: "Rectangle", Len: len(s.Rectangles)},
		{Name: "Image", Len: len(s.Images)},
		{Name: "Oval", Len: len(s.Ovals)},
		{Name: "Polygon", Len: len(s.Polygons)},
		{Name: "GraphicLine", Len: len(s.GraphicLines)},
		{Name: "Group", Len: len(s.Groups)},
		{Name: "Button", Len: len(s.Buttons)},
		{Name: kindOther, Len: len(s.OtherElements)},
	}
}

// childRef is one child of the spread: its kind and its index in that kind's slice.
type childRef struct {
	kind string
	i    int
}

// children returns the spread's children in the order MarshalXML writes them: the recorded order first, then the
// children appended since, grouped by kind. It follows common.EncodeChildren.
func (s *Spread) children() []childRef {
	kinds := s.childKinds()
	length := map[string]int{}
	for _, k := range kinds {
		length[k.Name] = k.Len
	}
	cursor := map[string]int{}
	var out []childRef
	for _, k := range s.childOrder {
		if n, ok := length[k]; ok && cursor[k] < n {
			out = append(out, childRef{k, cursor[k]})
			cursor[k]++
		}
	}
	for _, k := range kinds {
		for i := cursor[k.Name]; i < k.Len; i++ {
			out = append(out, childRef{k.Name, i})
		}
	}
	return out
}

func isStackKind(kind string) bool { return slices.Contains(stackKinds, kind) }

// self returns the Self of a typed page item child.
func (s *Spread) self(c childRef) string {
	switch c.kind {
	case "TextFrame":
		return s.TextFrames[c.i].Self
	case "Rectangle":
		return s.Rectangles[c.i].Self
	case "Image":
		return s.Images[c.i].Self
	case "Oval":
		return s.Ovals[c.i].Self
	case "Polygon":
		return s.Polygons[c.i].Self
	case "GraphicLine":
		return s.GraphicLines[c.i].Self
	case "Group":
		return s.Groups[c.i].Self
	case "Button":
		return s.Buttons[c.i].Self
	}
	return ""
}

// StackingOrder returns the Selfs of the spread's top-level page items, bottom to top, as the spread would be
// written now. Items kept as raw elements (page item types not modeled yet, such as EPS) are not listed; they keep
// their place when the spread is restacked. Items inside a group are the group's.
func (s *Spread) StackingOrder() []string {
	var out []string
	for _, c := range s.children() {
		if isStackKind(c.kind) {
			out = append(out, s.self(c))
		}
	}
	return out
}

// SetStackingOrder restacks the spread's top-level page items: selfs names every item StackingOrder lists, once,
// bottom to top. The items fill the places page items have in the spread in that order, so pages, the flattener
// preference and raw elements stay where they are. Nothing changes when selfs is not such a list.
//
// Call it after appending items, before writing. Edits made through a getter pointer still need
// Package.MarkModified for the spread's file.
func (s *Spread) SetStackingOrder(selfs []string) error {
	children := s.children()
	where := map[string]childRef{}
	for _, c := range children {
		if !isStackKind(c.kind) {
			continue
		}
		self := s.self(c)
		if _, dup := where[self]; dup {
			return common.Errorf("spread", "set stacking order", s.Self, "two page items are %s", self)
		}
		where[self] = c
	}
	seen := map[string]bool{}
	for _, self := range selfs {
		if _, ok := where[self]; !ok {
			return common.Errorf("spread", "set stacking order", s.Self, "%s is not a top-level page item of the spread", self)
		}
		if seen[self] {
			return common.Errorf("spread", "set stacking order", s.Self, "%s is named twice", self)
		}
		seen[self] = true
	}
	for self := range where {
		if !seen[self] {
			return common.Errorf("spread", "set stacking order", s.Self, "%s is missing from the order", self)
		}
	}

	// Rebuild each kind's slice in the new order, and the recorded order with the page items' places refilled.
	var (
		textFrames   []TextFrame
		rectangles   []Rectangle
		images       []Image
		ovals        []Oval
		polygons     []Polygon
		graphicLines []GraphicLine
		groups       []Group
		buttons      []Button
	)
	for _, self := range selfs {
		c := where[self]
		switch c.kind {
		case "TextFrame":
			textFrames = append(textFrames, s.TextFrames[c.i])
		case "Rectangle":
			rectangles = append(rectangles, s.Rectangles[c.i])
		case "Image":
			images = append(images, s.Images[c.i])
		case "Oval":
			ovals = append(ovals, s.Ovals[c.i])
		case "Polygon":
			polygons = append(polygons, s.Polygons[c.i])
		case "GraphicLine":
			graphicLines = append(graphicLines, s.GraphicLines[c.i])
		case "Group":
			groups = append(groups, s.Groups[c.i])
		case "Button":
			buttons = append(buttons, s.Buttons[c.i])
		}
	}
	var order common.ChildOrder
	next := 0
	for _, c := range children {
		if isStackKind(c.kind) {
			order.Record(where[selfs[next]].kind)
			next++
			continue
		}
		order.Record(c.kind)
	}
	s.TextFrames, s.Rectangles, s.Images, s.Ovals = textFrames, rectangles, images, ovals
	s.Polygons, s.GraphicLines, s.Groups, s.Buttons = polygons, graphicLines, groups, buttons
	s.childOrder = order
	return nil
}
