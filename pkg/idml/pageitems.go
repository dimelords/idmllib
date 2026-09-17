package idml

import (
	"github.com/dimelords/idmllib/v2/pkg/common"
)

// Looking up page items.
//
// Every page item carries a Self attribute that is unique within the
// document, and the package builds an index of them on first use. One method
// answers "what is this id?", and a generic helper answers "give me this id
// as a text frame".

// PageItemByID returns the page item with the given Self attribute, whatever
// kind it is. Use PageItemOfType, or a type assertion on the result, when a
// concrete type is needed.
func (p *Package) PageItemByID(id string) (PageItem, error) {
	if err := p.ensureItemIndex(); err != nil {
		return nil, common.WrapErrorWithPath("idml", "find page item", id, err)
	}
	idx := p.indexState.index
	if item, ok := idx.textFrames[id]; ok {
		return item, nil
	}
	if item, ok := idx.rectangles[id]; ok {
		return item, nil
	}
	if item, ok := idx.ovals[id]; ok {
		return item, nil
	}
	if item, ok := idx.polygons[id]; ok {
		return item, nil
	}
	if item, ok := idx.graphicLines[id]; ok {
		return item, nil
	}
	if item, ok := idx.groups[id]; ok {
		return item, nil
	}
	return nil, common.WrapErrorWithPath("idml", "find page item", id, common.ErrNotFound)
}

// PageItemOfType returns the page item with the given Self attribute as a
// concrete type:
//
//	tf, err := idml.PageItemOfType[spread.TextFrame](pkg, "u123")
//
// It reports an error when no item has that id, or when the item is of a
// different kind.
func PageItemOfType[T any](p *Package, id string) (*T, error) {
	item, err := p.PageItemByID(id)
	if err != nil {
		return nil, err
	}
	typed, ok := any(item).(*T)
	if !ok {
		return nil, common.Errorf("idml", "find page item", id,
			"page item is a %T, not the requested type", item)
	}
	return typed, nil
}

// SelectByIDs builds a Selection from page item ids, which is what the IDMS
// exporter consumes. Unknown ids are reported as an error.
func (p *Package) SelectByIDs(ids ...string) (*Selection, error) {
	selection := NewSelection()
	for _, id := range ids {
		item, err := p.PageItemByID(id)
		if err != nil {
			return nil, err
		}
		selection.AddPageItem(item)
	}
	return selection, nil
}
