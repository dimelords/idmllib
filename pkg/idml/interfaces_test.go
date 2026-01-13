package idml

import (
	"errors"
	"testing"

	"github.com/dimelords/idmllib/pkg/resources"
	"github.com/dimelords/idmllib/pkg/spread"
	"github.com/dimelords/idmllib/pkg/story"
)

func TestMockPackage_ImplementsPackageAccessor(t *testing.T) {
	// This test verifies at compile time that MockPackage implements PackageAccessor
	var _ PackageAccessor = (*MockPackage)(nil)
}

func TestMockPackage_Stories(t *testing.T) {
	mock := NewMockPackage()
	mock.MockStories["Stories/Story_u1.xml"] = &story.Story{}

	stories, err := mock.Stories()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stories) != 1 {
		t.Errorf("expected 1 story, got %d", len(stories))
	}
}

func TestMockPackage_StoriesError(t *testing.T) {
	mock := NewMockPackage()
	mock.StoriesErr = errors.New("test error")

	_, err := mock.Stories()
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestMockPackage_Story(t *testing.T) {
	mock := NewMockPackage()
	mock.MockStories["Stories/Story_u1.xml"] = &story.Story{}

	st, err := mock.Story("Stories/Story_u1.xml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if st == nil {
		t.Error("expected story, got nil")
	}
}

func TestMockPackage_StoryNotFound(t *testing.T) {
	mock := NewMockPackage()

	_, err := mock.Story("Stories/Story_notfound.xml")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestMockPackage_Spreads(t *testing.T) {
	mock := NewMockPackage()
	mock.MockSpreads["Spreads/Spread_u1.xml"] = &spread.Spread{}

	spreads, err := mock.Spreads()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(spreads) != 1 {
		t.Errorf("expected 1 spread, got %d", len(spreads))
	}
}

func TestMockPackage_Fonts(t *testing.T) {
	mock := NewMockPackage()
	mock.MockFonts = &resources.FontsFile{}

	fonts, err := mock.Fonts()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fonts == nil {
		t.Error("expected fonts, got nil")
	}
}

func TestMockPackage_FontsNotFound(t *testing.T) {
	mock := NewMockPackage()

	_, err := mock.Fonts()
	if err == nil {
		t.Error("expected error for missing fonts, got nil")
	}
}

func TestMockPackage_SetFonts(t *testing.T) {
	mock := NewMockPackage()
	fonts := &resources.FontsFile{}

	mock.SetFonts(fonts)

	if mock.MockFonts != fonts {
		t.Error("SetFonts did not update MockFonts")
	}
}

func TestMockPackage_SetStyles(t *testing.T) {
	mock := NewMockPackage()
	styles := &resources.StylesFile{}

	mock.SetStyles(styles)

	if mock.MockStyles != styles {
		t.Error("SetStyles did not update MockStyles")
	}
}

func TestMockPackage_SetGraphics(t *testing.T) {
	mock := NewMockPackage()
	graphics := &resources.GraphicFile{}

	mock.SetGraphics(graphics)

	if mock.MockGraphics != graphics {
		t.Error("SetGraphics did not update MockGraphics")
	}
}

func TestPackage_ImplementsPackageAccessor(t *testing.T) {
	// This test verifies at compile time that Package implements PackageAccessor
	var _ PackageAccessor = (*Package)(nil)
}

func TestPackage_SettersWork(t *testing.T) {
	pkg := New()

	fonts := &resources.FontsFile{}
	styles := &resources.StylesFile{}
	graphics := &resources.GraphicFile{}

	pkg.SetFonts(fonts)
	pkg.SetStyles(styles)
	pkg.SetGraphics(graphics)

	// Verify the setters work by retrieving the values
	retrievedFonts, err := pkg.Fonts()
	if err != nil {
		t.Fatalf("Failed to get fonts: %v", err)
	}
	if retrievedFonts != fonts {
		t.Error("SetFonts did not update fonts cache")
	}

	retrievedStyles, err := pkg.Styles()
	if err != nil {
		t.Fatalf("Failed to get styles: %v", err)
	}
	if retrievedStyles != styles {
		t.Error("SetStyles did not update styles cache")
	}

	retrievedGraphics, err := pkg.Graphics()
	if err != nil {
		t.Fatalf("Failed to get graphics: %v", err)
	}
	if retrievedGraphics != graphics {
		t.Error("SetGraphics did not update graphics cache")
	}
}
