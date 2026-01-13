package analysis

import (
	"testing"

	"github.com/dimelords/idmllib/pkg/idml"
	"github.com/dimelords/idmllib/pkg/spread"
	"github.com/dimelords/idmllib/pkg/story"
)

// TestNewDependencySet tests creating a new dependency set
func TestNewDependencySet(t *testing.T) {
	deps := NewDependencySet()

	if deps == nil {
		t.Fatal("NewDependencySet() returned nil")
	}

	if deps.Stories == nil {
		t.Error("Stories map is nil")
	}

	if deps.ParagraphStyles == nil {
		t.Error("ParagraphStyles map is nil")
	}

	if deps.ColorSpaces == nil {
		t.Error("ColorSpaces map is nil")
	}

	if len(deps.Stories) != 0 {
		t.Errorf("Expected empty Stories map, got %d entries", len(deps.Stories))
	}

	if len(deps.ColorSpaces) != 0 {
		t.Errorf("Expected empty ColorSpaces map, got %d entries", len(deps.ColorSpaces))
	}
}

// TestNewDependencyTracker tests creating a new dependency tracker
func TestNewDependencyTracker(t *testing.T) {
	// Load test IDML
	pkg, err := idml.Read("../../testdata/plain.idml")
	if err != nil {
		t.Fatalf("Failed to read IDML: %v", err)
	}

	tracker := NewDependencyTracker(pkg)

	if tracker == nil {
		t.Fatal("NewDependencyTracker() returned nil")
	}

	if tracker.pkg != pkg {
		t.Error("Package not set correctly")
	}

	deps := tracker.Dependencies()
	if deps == nil {
		t.Error("Dependencies() returned nil")
	}
}

// TestAnalyzeTextFrame tests analyzing a text frame
func TestAnalyzeTextFrame(t *testing.T) {
	// Load test IDML
	pkg, err := idml.Read("../../testdata/plain.idml")
	if err != nil {
		t.Fatalf("Failed to read IDML: %v", err)
	}

	tracker := NewDependencyTracker(pkg)

	// Get a text frame to analyze
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Failed to get spreads: %v", err)
	}

	if len(spreads) == 0 {
		t.Skip("No spreads in test file")
	}

	// Find a text frame
	var tf *spread.SpreadTextFrame
	for _, sp := range spreads {
		if len(sp.InnerSpread.TextFrames) > 0 {
			tf = &sp.InnerSpread.TextFrames[0]
			break
		}
	}

	if tf == nil {
		t.Skip("No text frames found")
	}

	// Analyze the text frame
	err = tracker.AnalyzeTextFrame(tf)
	if err != nil {
		t.Fatalf("AnalyzeTextFrame() error: %v", err)
	}

	deps := tracker.Dependencies()

	// Verify that we tracked something
	if tf.ParentStory != "" && len(deps.Stories) == 0 {
		t.Error("Expected to track parent story")
	}

	if tf.AppliedObjectStyle != "" && len(deps.ObjectStyles) == 0 {
		t.Error("Expected to track object style")
	}

	if tf.ItemLayer != "" && len(deps.Layers) == 0 {
		t.Error("Expected to track layer")
	}

	t.Logf("✅ Analyzed text frame '%s'", tf.Self)
	t.Logf("   Stories: %d", len(deps.Stories))
	t.Logf("   Object styles: %d", len(deps.ObjectStyles))
	t.Logf("   Layers: %d", len(deps.Layers))
	t.Logf("   Paragraph styles: %d", len(deps.ParagraphStyles))
	t.Logf("   Character styles: %d", len(deps.CharacterStyles))
}

// TestAnalyzeStory tests analyzing a story
func TestAnalyzeStory(t *testing.T) {
	// Load test IDML
	pkg, err := idml.Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read IDML: %v", err)
	}

	tracker := NewDependencyTracker(pkg)

	// Get a story to analyze
	stories, err := pkg.Stories()
	if err != nil {
		t.Fatalf("Failed to get stories: %v", err)
	}

	if len(stories) == 0 {
		t.Skip("No stories in test file")
	}

	// Analyze the first story
	var storyFilename string
	var st *story.Story
	for filename, s := range stories {
		storyFilename = filename
		st = s
		break
	}

	err = tracker.AnalyzeStory(st)
	if err != nil {
		t.Fatalf("AnalyzeStory() error: %v", err)
	}

	deps := tracker.Dependencies()

	// Verify that we tracked styles
	if len(st.StoryElement.ParagraphStyleRanges) > 0 {
		if len(deps.ParagraphStyles) == 0 {
			t.Error("Expected to track paragraph styles")
		}
		if len(deps.CharacterStyles) == 0 {
			t.Error("Expected to track character styles")
		}
	}

	t.Logf("✅ Analyzed story '%s'", storyFilename)
	t.Logf("   Paragraph styles: %d", len(deps.ParagraphStyles))
	t.Logf("   Character styles: %d", len(deps.CharacterStyles))

	// List the styles found
	if len(deps.ParagraphStyles) > 0 {
		t.Log("   Paragraph style IDs:")
		for styleID := range deps.ParagraphStyles {
			t.Logf("     - %s", styleID)
		}
	}
}

// TestAnalyzeRectangle tests analyzing a rectangle
func TestAnalyzeRectangle(t *testing.T) {
	// Load test IDML with graphics
	pkg, err := idml.Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read IDML: %v", err)
	}

	tracker := NewDependencyTracker(pkg)

	// Get a rectangle to analyze
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Failed to get spreads: %v", err)
	}

	if len(spreads) == 0 {
		t.Skip("No spreads in test file")
	}

	// Find a rectangle
	var rect *spread.Rectangle
	for _, sp := range spreads {
		if len(sp.InnerSpread.Rectangles) > 0 {
			rect = &sp.InnerSpread.Rectangles[0]
			break
		}
	}

	if rect == nil {
		t.Skip("No rectangles found")
	}

	// Analyze the rectangle
	err = tracker.AnalyzeRectangle(rect)
	if err != nil {
		t.Fatalf("AnalyzeRectangle() error: %v", err)
	}

	deps := tracker.Dependencies()

	// Verify that we tracked something
	if rect.AppliedObjectStyle != "" && len(deps.ObjectStyles) == 0 {
		t.Error("Expected to track object style")
	}

	if rect.ItemLayer != "" && len(deps.Layers) == 0 {
		t.Error("Expected to track layer")
	}

	hasImage := rect.Image != nil && rect.Image.Link != nil
	if hasImage && len(deps.Links) == 0 {
		t.Error("Expected to track image link")
	}

	t.Logf("✅ Analyzed rectangle '%s'", rect.Self)
	t.Logf("   Object styles: %d", len(deps.ObjectStyles))
	t.Logf("   Layers: %d", len(deps.Layers))
	t.Logf("   Links: %d", len(deps.Links))
	t.Logf("   Color spaces: %d", len(deps.ColorSpaces))
	t.Logf("   Has image: %v", hasImage)
}

// TestAnalyzeImage tests analyzing an image
func TestAnalyzeImage(t *testing.T) {
	// Load test IDML with graphics
	pkg, err := idml.Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read IDML: %v", err)
	}

	tracker := NewDependencyTracker(pkg)

	// Get an image to analyze - find a rectangle with an image
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Failed to get spreads: %v", err)
	}

	if len(spreads) == 0 {
		t.Skip("No spreads in test file")
	}

	// Find a rectangle with an image
	var img *spread.Image
	for _, sp := range spreads {
		for i := range sp.InnerSpread.Rectangles {
			if sp.InnerSpread.Rectangles[i].Image != nil {
				img = sp.InnerSpread.Rectangles[i].Image
				break
			}
		}
		if img != nil {
			break
		}
	}

	if img == nil {
		t.Skip("No images found in rectangles")
	}

	// Analyze the image
	err = tracker.AnalyzeImage(img)
	if err != nil {
		t.Fatalf("AnalyzeImage() error: %v", err)
	}

	deps := tracker.Dependencies()

	// Verify that we tracked the image's properties
	if img.AppliedObjectStyle != "" && len(deps.ObjectStyles) == 0 {
		t.Error("Expected to track image object style")
	}

	if img.Space != "" && len(deps.ColorSpaces) == 0 {
		t.Error("Expected to track color space")
	}

	if img.Link != nil {
		if img.Link.Self != "" && len(deps.Links) == 0 {
			t.Error("Expected to track link self")
		}
		if img.Link.LinkResourceURI != "" && len(deps.Links) == 0 {
			t.Error("Expected to track link URI")
		}
	}

	t.Logf("✅ Analyzed image '%s'", img.Self)
	t.Logf("   Object styles: %d", len(deps.ObjectStyles))
	t.Logf("   Color spaces: %d", len(deps.ColorSpaces))
	t.Logf("   Links: %d", len(deps.Links))

	if img.Space != "" {
		t.Logf("   Color space: %s", img.Space)
	}
	if img.Link != nil && img.Link.LinkResourceURI != "" {
		t.Logf("   Link URI: %s", img.Link.LinkResourceURI)
	}
}

// TestAnalyzeImage_ColorSpaceTracking tests that color spaces are tracked correctly
func TestAnalyzeImage_ColorSpaceTracking(t *testing.T) {
	pkg := idml.New()
	tracker := NewDependencyTracker(pkg)

	// Create test images with different color spaces
	testCases := []struct {
		name       string
		colorSpace string
	}{
		{"RGB image", "RGB"},
		{"CMYK image", "CMYK"},
		{"Lab image", "Lab"},
		{"Grayscale image", "DeviceGray"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			img := &spread.Image{
				Self:  "test_image_" + tc.colorSpace,
				Space: tc.colorSpace,
			}

			err := tracker.AnalyzeImage(img)
			if err != nil {
				t.Fatalf("AnalyzeImage() error: %v", err)
			}

			deps := tracker.Dependencies()
			if len(deps.ColorSpaces) == 0 {
				t.Error("Expected to track color space")
			}

			if !deps.ColorSpaces[tc.colorSpace] {
				t.Errorf("Expected color space '%s' to be tracked", tc.colorSpace)
			}
		})
	}

	// Verify all color spaces were tracked
	deps := tracker.Dependencies()
	if len(deps.ColorSpaces) != len(testCases) {
		t.Errorf("Expected %d color spaces, got %d", len(testCases), len(deps.ColorSpaces))
	}

	t.Logf("✅ Tracked %d color spaces correctly", len(deps.ColorSpaces))
}

// TestAnalyzeRectangle_WithImage tests analyzing a rectangle that contains an image
func TestAnalyzeRectangle_WithImage(t *testing.T) {
	pkg := idml.New()
	tracker := NewDependencyTracker(pkg)

	// Create a rectangle with an image
	rect := &spread.Rectangle{
		PageItemBase: spread.PageItemBase{
			Self:      "test_rect",
			ItemLayer: "u123",
		},
		AppliedObjectStyle: "ObjectStyle/$ID/TestStyle",
		Image: &spread.Image{
			Self:               "test_image",
			Space:              "CMYK",
			AppliedObjectStyle: "ObjectStyle/$ID/ImageStyle",
			Link: &spread.Link{
				Self:            "test_link",
				LinkResourceURI: "file://test.jpg",
			},
		},
	}

	err := tracker.AnalyzeRectangle(rect)
	if err != nil {
		t.Fatalf("AnalyzeRectangle() error: %v", err)
	}

	deps := tracker.Dependencies()

	// Verify rectangle properties were tracked
	if !deps.ObjectStyles["ObjectStyle/$ID/TestStyle"] {
		t.Error("Expected to track rectangle object style")
	}

	if !deps.Layers["u123"] {
		t.Error("Expected to track layer")
	}

	// Verify image properties were tracked
	if !deps.ObjectStyles["ObjectStyle/$ID/ImageStyle"] {
		t.Error("Expected to track image object style")
	}

	if !deps.ColorSpaces["CMYK"] {
		t.Error("Expected to track CMYK color space")
	}

	if !deps.Links["test_link"] {
		t.Error("Expected to track link self")
	}

	if !deps.Links["file://test.jpg"] {
		t.Error("Expected to track link URI")
	}

	t.Logf("✅ Rectangle with image tracked all dependencies")
	t.Logf("   Object styles: %d (includes both rectangle and image styles)", len(deps.ObjectStyles))
	t.Logf("   Layers: %d", len(deps.Layers))
	t.Logf("   Links: %d", len(deps.Links))
	t.Logf("   Color spaces: %d", len(deps.ColorSpaces))
}

// TestAnalyzeSelection tests analyzing an entire selection
func TestAnalyzeSelection(t *testing.T) {
	// Load test IDML
	pkg, err := idml.Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read IDML: %v", err)
	}

	tracker := NewDependencyTracker(pkg)

	// Create a selection with multiple elements
	selection := idml.NewSelection()

	// Add some text frames
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Failed to get spreads: %v", err)
	}

	for _, sp := range spreads {
		// Add up to 2 text frames
		for i := 0; i < len(sp.InnerSpread.TextFrames) && i < 2; i++ {
			selection.AddTextFrame(&sp.InnerSpread.TextFrames[i])
		}

		// Add up to 2 rectangles
		for i := 0; i < len(sp.InnerSpread.Rectangles) && i < 2; i++ {
			selection.AddRectangle(&sp.InnerSpread.Rectangles[i])
		}

		if selection.Count() >= 4 {
			break
		}
	}

	if selection.IsEmpty() {
		t.Skip("No elements to analyze")
	}

	// Analyze the selection
	err = tracker.AnalyzeSelection(selection)
	if err != nil {
		t.Fatalf("AnalyzeSelection() error: %v", err)
	}

	deps := tracker.Dependencies()
	summary := tracker.Summary()

	t.Logf("✅ Analyzed selection with %d elements", selection.Count())
	t.Logf("   Text frames: %d", len(selection.TextFrames))
	t.Logf("   Rectangles: %d", len(selection.Rectangles))
	t.Logf("")
	t.Logf("   Dependencies found:")
	t.Logf("   - Stories: %d", summary.StoriesCount)
	t.Logf("   - Paragraph styles: %d", summary.ParagraphStylesCount)
	t.Logf("   - Character styles: %d", summary.CharacterStylesCount)
	t.Logf("   - Object styles: %d", summary.ObjectStylesCount)
	t.Logf("   - Layers: %d", summary.LayersCount)
	t.Logf("   - Links: %d", summary.LinksCount)
	t.Logf("   - Color spaces: %d", summary.ColorSpacesCount)

	// Verify we found some dependencies
	if len(selection.TextFrames) > 0 {
		if len(deps.Stories) == 0 {
			t.Error("Expected to find story dependencies for text frames")
		}
	}

	if selection.Count() > 0 {
		if len(deps.ObjectStyles) == 0 && len(deps.Layers) == 0 {
			t.Error("Expected to find at least some object styles or layers")
		}
	}
}

// TestSummary tests the dependency summary
func TestSummary(t *testing.T) {
	// Create a tracker with some dependencies
	deps := NewDependencySet()
	deps.Stories["Stories/Story_u1d8.xml"] = true
	deps.Stories["Stories/Story_u1d9.xml"] = true
	deps.ParagraphStyles["ParagraphStyle/$ID/[Normal]"] = true
	deps.CharacterStyles["CharacterStyle/$ID/[None]"] = true
	deps.ObjectStyles["ObjectStyle/$ID/[Basic Text Frame]"] = true
	deps.Layers["u123"] = true
	deps.Links["file://test.jpg"] = true
	deps.ColorSpaces["CMYK"] = true
	deps.ColorSpaces["RGB"] = true

	tracker := &DependencyTracker{
		deps: deps,
	}

	summary := tracker.Summary()

	if summary.StoriesCount != 2 {
		t.Errorf("Expected 2 stories, got %d", summary.StoriesCount)
	}

	if summary.ParagraphStylesCount != 1 {
		t.Errorf("Expected 1 paragraph style, got %d", summary.ParagraphStylesCount)
	}

	if summary.CharacterStylesCount != 1 {
		t.Errorf("Expected 1 character style, got %d", summary.CharacterStylesCount)
	}

	if summary.ObjectStylesCount != 1 {
		t.Errorf("Expected 1 object style, got %d", summary.ObjectStylesCount)
	}

	if summary.LayersCount != 1 {
		t.Errorf("Expected 1 layer, got %d", summary.LayersCount)
	}

	if summary.LinksCount != 1 {
		t.Errorf("Expected 1 link, got %d", summary.LinksCount)
	}

	if summary.ColorSpacesCount != 2 {
		t.Errorf("Expected 2 color spaces, got %d", summary.ColorSpacesCount)
	}

	t.Log("✅ Summary correctly counts all dependency types including color spaces")
}

// TestAnalyzeSelection_Empty tests analyzing an empty selection
func TestAnalyzeSelection_Empty(t *testing.T) {
	pkg, err := idml.Read("../../testdata/plain.idml")
	if err != nil {
		t.Fatalf("Failed to read IDML: %v", err)
	}

	tracker := NewDependencyTracker(pkg)
	selection := idml.NewSelection()

	err = tracker.AnalyzeSelection(selection)
	if err != nil {
		t.Fatalf("AnalyzeSelection() error: %v", err)
	}

	summary := tracker.Summary()

	if summary.StoriesCount != 0 {
		t.Error("Expected no dependencies for empty selection")
	}

	if summary.ColorSpacesCount != 0 {
		t.Error("Expected no color space dependencies for empty selection")
	}

	t.Log("✅ Empty selection produces no dependencies")
}

// TestAnalyzeOval tests analyzing an oval element
func TestAnalyzeOval(t *testing.T) {
	pkg, err := idml.Read("../../testdata/plain.idml")
	if err != nil {
		t.Fatalf("Failed to read IDML: %v", err)
	}

	tracker := NewDependencyTracker(pkg)

	// Create test oval with various dependencies
	oval := &spread.Oval{
		PageItemBase: spread.PageItemBase{
			Self:      "oval_1",
			ItemLayer: "Layer1",
		},
		AppliedObjectStyle: "ObjectStyle/$ID/TestStyle",
		StrokeColor:        "Color/Black",
		FillColor:          "Color/Red",
	}

	err = tracker.AnalyzeOval(oval)
	if err != nil {
		t.Fatalf("AnalyzeOval() error: %v", err)
	}

	deps := tracker.Dependencies()

	// Verify object style was tracked
	if !deps.ObjectStyles["ObjectStyle/$ID/TestStyle"] {
		t.Error("Object style not tracked")
	}

	// Verify layer was tracked
	if !deps.Layers["Layer1"] {
		t.Error("Layer not tracked")
	}

	// Verify colors were tracked
	if !deps.Colors["Color/Black"] {
		t.Error("Stroke color not tracked")
	}
	if !deps.Colors["Color/Red"] {
		t.Error("Fill color not tracked")
	}

	t.Log("✅ Oval dependencies correctly tracked")
}

// TestAnalyzeOval_WithImage tests analyzing an oval that contains an image
func TestAnalyzeOval_WithImage(t *testing.T) {
	pkg, err := idml.Read("../../testdata/plain.idml")
	if err != nil {
		t.Fatalf("Failed to read IDML: %v", err)
	}

	tracker := NewDependencyTracker(pkg)

	oval := &spread.Oval{
		PageItemBase: spread.PageItemBase{
			Self:      "oval_2",
			ItemLayer: "OvalImageLayer",
		},
		AppliedObjectStyle: "ObjectStyle/$ID/ImageStyle",
		Image: &spread.Image{
			Self:               "image_1",
			AppliedObjectStyle: "ObjectStyle/$ID/ImageObjStyle",
		},
	}

	err = tracker.AnalyzeOval(oval)
	if err != nil {
		t.Fatalf("AnalyzeOval() error: %v", err)
	}

	deps := tracker.Dependencies()

	// Verify oval layer was tracked
	if !deps.Layers["OvalImageLayer"] {
		t.Error("Oval layer not tracked")
	}

	// Verify image object style was tracked
	if !deps.ObjectStyles["ObjectStyle/$ID/ImageObjStyle"] {
		t.Error("Image object style not tracked")
	}

	t.Log("✅ Oval with image correctly tracked")
}

// TestAnalyzePolygon tests analyzing a polygon element
func TestAnalyzePolygon(t *testing.T) {
	pkg, err := idml.Read("../../testdata/plain.idml")
	if err != nil {
		t.Fatalf("Failed to read IDML: %v", err)
	}

	tracker := NewDependencyTracker(pkg)

	polygon := &spread.Polygon{
		PageItemBase: spread.PageItemBase{
			Self:      "polygon_1",
			ItemLayer: "PolyLayer",
		},
		AppliedObjectStyle: "ObjectStyle/$ID/PolygonStyle",
		StrokeColor:        "Color/Blue",
		FillColor:          "Color/Green",
	}

	err = tracker.AnalyzePolygon(polygon)
	if err != nil {
		t.Fatalf("AnalyzePolygon() error: %v", err)
	}

	deps := tracker.Dependencies()

	if !deps.ObjectStyles["ObjectStyle/$ID/PolygonStyle"] {
		t.Error("Object style not tracked")
	}
	if !deps.Layers["PolyLayer"] {
		t.Error("Layer not tracked")
	}
	if !deps.Colors["Color/Blue"] {
		t.Error("Stroke color not tracked")
	}
	if !deps.Colors["Color/Green"] {
		t.Error("Fill color not tracked")
	}

	t.Log("✅ Polygon dependencies correctly tracked")
}

// TestAnalyzePolygon_WithImage tests analyzing a polygon that contains an image
func TestAnalyzePolygon_WithImage(t *testing.T) {
	pkg, err := idml.Read("../../testdata/plain.idml")
	if err != nil {
		t.Fatalf("Failed to read IDML: %v", err)
	}

	tracker := NewDependencyTracker(pkg)

	polygon := &spread.Polygon{
		PageItemBase: spread.PageItemBase{
			Self:      "polygon_2",
			ItemLayer: "PolygonLayer",
		},
		Image: &spread.Image{
			Self:               "image_2",
			AppliedObjectStyle: "ObjectStyle/$ID/PolyImageStyle",
		},
	}

	err = tracker.AnalyzePolygon(polygon)
	if err != nil {
		t.Fatalf("AnalyzePolygon() error: %v", err)
	}

	deps := tracker.Dependencies()

	if !deps.Layers["PolygonLayer"] {
		t.Error("Polygon layer not tracked")
	}

	if !deps.ObjectStyles["ObjectStyle/$ID/PolyImageStyle"] {
		t.Error("Image object style not tracked")
	}

	t.Log("✅ Polygon with image correctly tracked")
}

// TestAnalyzeGraphicLine tests analyzing a graphic line element
func TestAnalyzeGraphicLine(t *testing.T) {
	pkg, err := idml.Read("../../testdata/plain.idml")
	if err != nil {
		t.Fatalf("Failed to read IDML: %v", err)
	}

	tracker := NewDependencyTracker(pkg)

	line := &spread.GraphicLine{
		PageItemBase: spread.PageItemBase{
			Self:      "line_1",
			ItemLayer: "LineLayer",
		},
		AppliedObjectStyle: "ObjectStyle/$ID/LineStyle",
		StrokeColor:        "Color/Yellow",
		FillColor:          "Color/None",
	}

	err = tracker.AnalyzeGraphicLine(line)
	if err != nil {
		t.Fatalf("AnalyzeGraphicLine() error: %v", err)
	}

	deps := tracker.Dependencies()

	if !deps.ObjectStyles["ObjectStyle/$ID/LineStyle"] {
		t.Error("Object style not tracked")
	}
	if !deps.Layers["LineLayer"] {
		t.Error("Layer not tracked")
	}
	if !deps.Colors["Color/Yellow"] {
		t.Error("Stroke color not tracked")
	}
	if !deps.Colors["Color/None"] {
		t.Error("Fill color not tracked")
	}

	t.Log("✅ GraphicLine dependencies correctly tracked")
}

// TestAnalyzeGroup tests analyzing a group element
func TestAnalyzeGroup(t *testing.T) {
	pkg, err := idml.Read("../../testdata/plain.idml")
	if err != nil {
		t.Fatalf("Failed to read IDML: %v", err)
	}

	tracker := NewDependencyTracker(pkg)

	group := &spread.Group{
		PageItemBase: spread.PageItemBase{
			Self:      "group_1",
			ItemLayer: "GroupLayer",
		},
		AppliedObjectStyle: "ObjectStyle/$ID/GroupStyle",
	}

	err = tracker.AnalyzeGroup(group)
	if err != nil {
		t.Fatalf("AnalyzeGroup() error: %v", err)
	}

	deps := tracker.Dependencies()

	if !deps.ObjectStyles["ObjectStyle/$ID/GroupStyle"] {
		t.Error("Object style not tracked")
	}
	if !deps.Layers["GroupLayer"] {
		t.Error("Layer not tracked")
	}

	t.Log("✅ Group dependencies correctly tracked")
}

// TestAnalyzeSelection_WithOvalsPolygonsLines tests full selection with various element types
func TestAnalyzeSelection_WithOvalsPolygonsLines(t *testing.T) {
	pkg, err := idml.Read("../../testdata/plain.idml")
	if err != nil {
		t.Fatalf("Failed to read IDML: %v", err)
	}

	tracker := NewDependencyTracker(pkg)
	selection := idml.NewSelection()

	// Add various element types
	selection.AddOval(&spread.Oval{
		PageItemBase: spread.PageItemBase{
			Self: "oval_sel",
		},
		StrokeColor: "Color/OvalStroke",
	})
	selection.AddPolygon(&spread.Polygon{
		PageItemBase: spread.PageItemBase{
			Self: "poly_sel",
		},
		FillColor: "Color/PolyFill",
	})
	selection.AddGraphicLine(&spread.GraphicLine{
		PageItemBase: spread.PageItemBase{
			Self: "line_sel",
		},
		StrokeColor: "Color/LineStroke",
	})
	selection.AddGroup(&spread.Group{
		PageItemBase: spread.PageItemBase{
			Self:      "group_sel",
			ItemLayer: "GroupSelLayer",
		},
	})

	err = tracker.AnalyzeSelection(selection)
	if err != nil {
		t.Fatalf("AnalyzeSelection() error: %v", err)
	}

	deps := tracker.Dependencies()

	// Verify all colors tracked
	if !deps.Colors["Color/OvalStroke"] {
		t.Error("Oval stroke color not tracked")
	}
	if !deps.Colors["Color/PolyFill"] {
		t.Error("Polygon fill color not tracked")
	}
	if !deps.Colors["Color/LineStroke"] {
		t.Error("Line stroke color not tracked")
	}

	// Verify group layer tracked
	if !deps.Layers["GroupSelLayer"] {
		t.Error("Group layer not tracked")
	}

	t.Log("✅ Selection with ovals, polygons, lines, and groups correctly analyzed")
}
