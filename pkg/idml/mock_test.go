package idml

import (
	"github.com/dimelords/idmllib/pkg/common"
	"github.com/dimelords/idmllib/pkg/resources"
	"github.com/dimelords/idmllib/pkg/spread"
	"github.com/dimelords/idmllib/pkg/story"
)

// MockPackage is a test double for Package that implements PackageAccessor.
// Use it in unit tests to avoid loading real IDML files.
type MockPackage struct {
	// MockStories holds the stories to return from Stories() and Story().
	MockStories map[string]*story.Story

	// MockSpreads holds the spreads to return from Spreads() and Spread().
	MockSpreads map[string]*spread.Spread

	// MockFonts holds the fonts to return from Fonts().
	MockFonts *resources.FontsFile

	// MockStyles holds the styles to return from Styles().
	MockStyles *resources.StylesFile

	// MockGraphics holds the graphics to return from Graphics().
	MockGraphics *resources.GraphicFile

	// Error fields allow simulating errors.
	StoriesErr  error
	SpreadsErr  error
	FontsErr    error
	StylesErr   error
	GraphicsErr error
}

// NewMockPackage creates a new MockPackage with initialized maps.
func NewMockPackage() *MockPackage {
	return &MockPackage{
		MockStories: make(map[string]*story.Story),
		MockSpreads: make(map[string]*spread.Spread),
	}
}

// Stories returns MockStories or StoriesErr.
func (m *MockPackage) Stories() (map[string]*story.Story, error) {
	if m.StoriesErr != nil {
		return nil, m.StoriesErr
	}
	return m.MockStories, nil
}

// Story returns a specific story from MockStories.
func (m *MockPackage) Story(filename string) (*story.Story, error) {
	if m.StoriesErr != nil {
		return nil, m.StoriesErr
	}
	st, ok := m.MockStories[filename]
	if !ok {
		return nil, common.WrapErrorWithPath("idml", "get story", filename, common.ErrNotFound)
	}
	return st, nil
}

// Spreads returns MockSpreads or SpreadsErr.
func (m *MockPackage) Spreads() (map[string]*spread.Spread, error) {
	if m.SpreadsErr != nil {
		return nil, m.SpreadsErr
	}
	return m.MockSpreads, nil
}

// Spread returns a specific spread from MockSpreads.
func (m *MockPackage) Spread(filename string) (*spread.Spread, error) {
	if m.SpreadsErr != nil {
		return nil, m.SpreadsErr
	}
	sp, ok := m.MockSpreads[filename]
	if !ok {
		return nil, common.WrapErrorWithPath("idml", "get spread", filename, common.ErrNotFound)
	}
	return sp, nil
}

// Fonts returns MockFonts or FontsErr.
func (m *MockPackage) Fonts() (*resources.FontsFile, error) {
	if m.FontsErr != nil {
		return nil, m.FontsErr
	}
	if m.MockFonts == nil {
		return nil, common.WrapErrorWithPath("idml", "get fonts", "Resources/Fonts.xml", common.ErrNotFound)
	}
	return m.MockFonts, nil
}

// Styles returns MockStyles or StylesErr.
func (m *MockPackage) Styles() (*resources.StylesFile, error) {
	if m.StylesErr != nil {
		return nil, m.StylesErr
	}
	if m.MockStyles == nil {
		return nil, common.WrapErrorWithPath("idml", "get styles", "Resources/Styles.xml", common.ErrNotFound)
	}
	return m.MockStyles, nil
}

// Graphics returns MockGraphics or GraphicsErr.
func (m *MockPackage) Graphics() (*resources.GraphicFile, error) {
	if m.GraphicsErr != nil {
		return nil, m.GraphicsErr
	}
	if m.MockGraphics == nil {
		return nil, common.WrapErrorWithPath("idml", "get graphics", "Resources/Graphic.xml", common.ErrNotFound)
	}
	return m.MockGraphics, nil
}

// SetFonts updates MockFonts.
func (m *MockPackage) SetFonts(fonts *resources.FontsFile) {
	m.MockFonts = fonts
}

// SetStyles updates MockStyles.
func (m *MockPackage) SetStyles(styles *resources.StylesFile) {
	m.MockStyles = styles
}

// SetGraphics updates MockGraphics.
func (m *MockPackage) SetGraphics(graphics *resources.GraphicFile) {
	m.MockGraphics = graphics
}
