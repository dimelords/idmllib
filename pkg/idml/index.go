package idml

import (
	"sync"

	"github.com/dimelords/idmllib/pkg/spread"
)

// itemIndex provides O(1) lookup for page items by their Self ID.
// It caches pointers to all page items across all spreads.
// The index is built lazily on first access.
type itemIndex struct {
	textFrames   map[string]*spread.SpreadTextFrame
	rectangles   map[string]*spread.Rectangle
	ovals        map[string]*spread.Oval
	polygons     map[string]*spread.Polygon
	graphicLines map[string]*spread.GraphicLine
	groups       map[string]*spread.Group
}

// newItemIndex creates a new empty item index with initialized maps.
func newItemIndex() *itemIndex {
	return &itemIndex{
		textFrames:   make(map[string]*spread.SpreadTextFrame),
		rectangles:   make(map[string]*spread.Rectangle),
		ovals:        make(map[string]*spread.Oval),
		polygons:     make(map[string]*spread.Polygon),
		graphicLines: make(map[string]*spread.GraphicLine),
		groups:       make(map[string]*spread.Group),
	}
}

// itemIndexState holds the index state for a Package.
// This is embedded in the Package struct.
type itemIndexState struct {
	index *itemIndex
	once  sync.Once
	err   error
}

// ensureItemIndex builds the item index if it hasn't been built yet.
// Uses sync.Once for thread-safe lazy initialization.
func (p *Package) ensureItemIndex() error {
	p.indexState.once.Do(func() {
		p.indexState.index = newItemIndex()
		p.indexState.err = p.buildItemIndex()
	})

	return p.indexState.err
}

// buildItemIndex populates the index with all page items from all spreads.
func (p *Package) buildItemIndex() error {
	spreads, err := p.Spreads()
	if err != nil {
		return err
	}

	for _, sp := range spreads {
		// Index text frames
		for i := range sp.InnerSpread.TextFrames {
			tf := &sp.InnerSpread.TextFrames[i]
			p.indexState.index.textFrames[tf.Self] = tf
		}

		// Index rectangles
		for i := range sp.InnerSpread.Rectangles {
			rect := &sp.InnerSpread.Rectangles[i]
			p.indexState.index.rectangles[rect.Self] = rect
		}

		// Index ovals
		for i := range sp.InnerSpread.Ovals {
			oval := &sp.InnerSpread.Ovals[i]
			p.indexState.index.ovals[oval.Self] = oval
		}

		// Index polygons
		for i := range sp.InnerSpread.Polygons {
			poly := &sp.InnerSpread.Polygons[i]
			p.indexState.index.polygons[poly.Self] = poly
		}

		// Index graphic lines
		for i := range sp.InnerSpread.GraphicLines {
			line := &sp.InnerSpread.GraphicLines[i]
			p.indexState.index.graphicLines[line.Self] = line
		}

		// Index groups
		for i := range sp.InnerSpread.Groups {
			group := &sp.InnerSpread.Groups[i]
			p.indexState.index.groups[group.Self] = group
		}
	}

	return nil
}

// ItemCount returns the total number of indexed items.
// Returns 0 if the index hasn't been built.
func (p *Package) ItemCount() int {
	if p.indexState.index == nil {
		return 0
	}
	return len(p.indexState.index.textFrames) +
		len(p.indexState.index.rectangles) +
		len(p.indexState.index.ovals) +
		len(p.indexState.index.polygons) +
		len(p.indexState.index.graphicLines) +
		len(p.indexState.index.groups)
}

// TextFrameCount returns the number of indexed text frames.
func (p *Package) TextFrameCount() int {
	if p.indexState.index == nil {
		return 0
	}
	return len(p.indexState.index.textFrames)
}

// RectangleCount returns the number of indexed rectangles.
func (p *Package) RectangleCount() int {
	if p.indexState.index == nil {
		return 0
	}
	return len(p.indexState.index.rectangles)
}
