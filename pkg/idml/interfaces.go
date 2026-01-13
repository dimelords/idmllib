package idml

import (
	"github.com/dimelords/idmllib/v2/pkg/resources"
	"github.com/dimelords/idmllib/v2/pkg/spread"
	"github.com/dimelords/idmllib/v2/pkg/story"
)

// PackageReader provides read-only access to IDML content.
// This interface enables testing and mocking of Package dependencies.
type PackageReader interface {
	// Stories returns all parsed Story files from the Stories/ directory.
	Stories() (map[string]*story.Story, error)

	// Story returns a parsed Story from the Stories/ directory.
	Story(filename string) (*story.Story, error)

	// Spreads returns all parsed Spread files from the Spreads/ directory.
	Spreads() (map[string]*spread.Spread, error)

	// Spread returns a parsed Spread from the Spreads/ directory.
	Spread(filename string) (*spread.Spread, error)

	// Fonts returns the typed Fonts.xml file.
	Fonts() (*resources.FontsFile, error)

	// Styles returns the typed Styles.xml file.
	Styles() (*resources.StylesFile, error)

	// Graphics returns the typed Graphic.xml file.
	Graphics() (*resources.GraphicFile, error)
}

// PackageWriter provides write access to IDML content.
// This interface enables testing and mocking of Package modifications.
type PackageWriter interface {
	// SetFonts updates the cached fonts file.
	// The file will be marshaled when Write() is called.
	SetFonts(fonts *resources.FontsFile)

	// SetStyles updates the cached styles file.
	// The file will be marshaled when Write() is called.
	SetStyles(styles *resources.StylesFile)

	// SetGraphics updates the cached graphics file.
	// The file will be marshaled when Write() is called.
	SetGraphics(graphics *resources.GraphicFile)
}

// PackageAccessor combines read and write access to IDML content.
// This is the primary interface used by ResourceManager and other
// components that need both read and write access.
type PackageAccessor interface {
	PackageReader
	PackageWriter
}

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

// Ensure Package implements PackageAccessor at compile time.
var _ PackageAccessor = (*Package)(nil)
