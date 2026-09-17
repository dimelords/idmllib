package idml

import (
	"sync"

	"github.com/dimelords/idmllib/v3/pkg/spread"
)

// itemIndex provides O(1) lookup for page items by their Self ID.
// It caches pointers to all page items across all spreads.
// The index is built lazily on first access.
type itemIndex struct {
	textFrames   map[string]*spread.TextFrame
	rectangles   map[string]*spread.Rectangle
	ovals        map[string]*spread.Oval
	polygons     map[string]*spread.Polygon
	graphicLines map[string]*spread.GraphicLine
	groups       map[string]*spread.Group
}

// newItemIndex creates a new empty item index with initialized maps.
func newItemIndex() *itemIndex {
	return &itemIndex{
		textFrames:   make(map[string]*spread.TextFrame),
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
//
// The sync.Once only serializes the build itself, so read-only selection on a
// package that is not being modified can run from several goroutines. It does
// not make Package safe for concurrent use in general: getters populate caches
// and the modification API rebuilds this index, and neither is synchronized.
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
		for i := range sp.TextFrames {
			tf := &sp.TextFrames[i]
			p.indexState.index.textFrames[tf.Self] = tf
		}

		// Index rectangles
		for i := range sp.Rectangles {
			rect := &sp.Rectangles[i]
			p.indexState.index.rectangles[rect.Self] = rect
		}

		// Index ovals
		for i := range sp.Ovals {
			oval := &sp.Ovals[i]
			p.indexState.index.ovals[oval.Self] = oval
		}

		// Index polygons
		for i := range sp.Polygons {
			poly := &sp.Polygons[i]
			p.indexState.index.polygons[poly.Self] = poly
		}

		// Index graphic lines
		for i := range sp.GraphicLines {
			line := &sp.GraphicLines[i]
			p.indexState.index.graphicLines[line.Self] = line
		}

		// Index groups
		for i := range sp.Groups {
			group := &sp.Groups[i]
			p.indexState.index.groups[group.Self] = group
		}
	}

	return nil
}
