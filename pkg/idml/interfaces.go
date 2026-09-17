package idml

// PageItem represents any visual element that can be placed on a spread.
// All page items share common attributes like Self ID, layer, bounds, transform, visibility, and name.
// This interface enables polymorphic operations on different page item types.
type PageItem interface {
	// GetSelf returns the unique identifier for this page item
	GetSelf() string

	// GetItemLayer returns the layer ID this page item is on
	GetItemLayer() string

	// GetGeometricBounds returns the bounding box in "y1 x1 y2 x2" format
	GetGeometricBounds() string

	// GetItemTransform returns the 6-value transformation matrix
	GetItemTransform() string

	// GetVisible returns the visibility state ("true" or "false")
	GetVisible() string

	// GetName returns the display name of the page item
	GetName() string
}
