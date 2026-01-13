package idml

import (
	"testing"

	"github.com/dimelords/idmllib/pkg/spread"
)

// TestNewSelection_CreatesEmptySelection tests creating a new empty selection
func TestNewSelection_CreatesEmptySelection(t *testing.T) {
	sel := NewSelection()

	if sel == nil {
		t.Fatal("NewSelection() returned nil")
	}

	if !sel.IsEmpty() {
		t.Error("New selection should be empty")
	}

	if sel.Count() != 0 {
		t.Errorf("New selection count should be 0, got %d", sel.Count())
	}
}

// TestSelectionAddElements_AddsElementsCorrectly tests adding elements to a selection
func TestSelectionAddElements_AddsElementsCorrectly(t *testing.T) {
	sel := NewSelection()

	// Add a text frame
	tf := &spread.SpreadTextFrame{
		PageItemBase: spread.PageItemBase{Self: "tf1"},
	}
	sel.AddTextFrame(tf)

	if len(sel.TextFrames) != 1 {
		t.Errorf("Expected 1 text frame, got %d", len(sel.TextFrames))
	}

	if sel.Count() != 1 {
		t.Errorf("Expected count 1, got %d", sel.Count())
	}

	// Add a rectangle
	rect := &spread.Rectangle{
		PageItemBase: spread.PageItemBase{Self: "rect1"},
	}
	sel.AddRectangle(rect)

	if len(sel.Rectangles) != 1 {
		t.Errorf("Expected 1 rectangle, got %d", len(sel.Rectangles))
	}

	if sel.Count() != 2 {
		t.Errorf("Expected count 2, got %d", sel.Count())
	}

	if sel.IsEmpty() {
		t.Error("Selection should not be empty")
	}
}

// TestSelectTextFrameByID_FindsTextFrame tests selecting a text frame by ID
func TestSelectTextFrameByID_FindsTextFrame(t *testing.T) {
	// Load test IDML file
	pkg, err := Read("../../testdata/plain.idml")
	if err != nil {
		t.Fatalf("Failed to read IDML: %v", err)
	}

	// Get all spreads to find a text frame ID
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Failed to get spreads: %v", err)
	}

	if len(spreads) == 0 {
		t.Skip("No spreads in test file")
	}

	// Find a text frame ID from the first spread
	var testID string
	for _, spread := range spreads {
		if len(spread.InnerSpread.TextFrames) > 0 {
			testID = spread.InnerSpread.TextFrames[0].Self
			break
		}
	}

	if testID == "" {
		t.Skip("No text frames found in test file")
	}

	// Test selecting by ID
	tf, err := pkg.SelectTextFrameByID(testID)
	if err != nil {
		t.Fatalf("SelectTextFrameByID() error: %v", err)
	}

	if tf == nil {
		t.Fatal("SelectTextFrameByID() returned nil")
	}

	if tf.Self != testID {
		t.Errorf("Expected ID '%s', got '%s'", testID, tf.Self)
	}

	t.Logf("✅ Successfully selected text frame: %s", testID)
}

// TestSelectTextFrameByID_NotFound tests error handling for non-existent ID
func TestSelectTextFrameByID_NotFound(t *testing.T) {
	pkg, err := Read("../../testdata/plain.idml")
	if err != nil {
		t.Fatalf("Failed to read IDML: %v", err)
	}

	// Try to select a non-existent text frame
	tf, err := pkg.SelectTextFrameByID("nonexistent_id")
	if err == nil {
		t.Error("Expected error for non-existent ID, got nil")
	}

	if tf != nil {
		t.Error("Expected nil text frame for non-existent ID")
	}

	t.Logf("✅ Correctly returned error: %v", err)
}

// TestSelectRectangleByID_FindsRectangle tests selecting a rectangle by ID
func TestSelectRectangleByID_FindsRectangle(t *testing.T) {
	// Load test IDML file
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read IDML: %v", err)
	}

	// Get all spreads to find a rectangle ID
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Failed to get spreads: %v", err)
	}

	if len(spreads) == 0 {
		t.Skip("No spreads in test file")
	}

	// Find a rectangle ID from the first spread
	var testID string
	for _, spread := range spreads {
		if len(spread.InnerSpread.Rectangles) > 0 {
			testID = spread.InnerSpread.Rectangles[0].Self
			break
		}
	}

	if testID == "" {
		t.Skip("No rectangles found in test file")
	}

	// Test selecting by ID
	rect, err := pkg.SelectRectangleByID(testID)
	if err != nil {
		t.Fatalf("SelectRectangleByID() error: %v", err)
	}

	if rect == nil {
		t.Fatal("SelectRectangleByID() returned nil")
	}

	if rect.Self != testID {
		t.Errorf("Expected ID '%s', got '%s'", testID, rect.Self)
	}

	t.Logf("✅ Successfully selected rectangle: %s", testID)
}

// TestSelectRectangleByID_NotFound tests error handling for non-existent rectangle
func TestSelectRectangleByID_NotFound(t *testing.T) {
	pkg, err := Read("../../testdata/plain.idml")
	if err != nil {
		t.Fatalf("Failed to read IDML: %v", err)
	}

	// Try to select a non-existent rectangle
	rect, err := pkg.SelectRectangleByID("nonexistent_rect")
	if err == nil {
		t.Error("Expected error for non-existent ID, got nil")
	}

	if rect != nil {
		t.Error("Expected nil rectangle for non-existent ID")
	}

	t.Logf("✅ Correctly returned error: %v", err)
}

// TestSelectAllGraphicsInSpread_FindsAllGraphics tests selecting all graphics in a spread
func TestSelectAllGraphicsInSpread_FindsAllGraphics(t *testing.T) {
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read IDML: %v", err)
	}

	// Get all spreads to find one with graphics
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Failed to get spreads: %v", err)
	}

	if len(spreads) == 0 {
		t.Skip("No spreads in test file")
	}

	// Find a spread with rectangles
	var spreadFilename string
	for filename, spread := range spreads {
		if len(spread.InnerSpread.Rectangles) > 0 {
			spreadFilename = filename
			break
		}
	}

	if spreadFilename == "" {
		t.Skip("No spreads with rectangles found")
	}

	// Test selecting all graphics
	graphics, err := pkg.SelectAllGraphicsInSpread(spreadFilename)
	if err != nil {
		t.Fatalf("SelectAllGraphicsInSpread() error: %v", err)
	}

	if graphics == nil {
		t.Fatal("SelectAllGraphicsInSpread() returned nil")
	}

	t.Logf("✅ Found %d graphics in spread %s", len(graphics), spreadFilename)

	// Verify each graphic is actually a graphic type
	for i, graphic := range graphics {
		if graphic.ContentType != "GraphicType" && graphic.Image == nil {
			t.Errorf("Graphics[%d] is not a graphic type (ContentType=%s, Image=%v)",
				i, graphic.ContentType, graphic.Image != nil)
		}
	}
}

// TestSelectAllTextFramesInSpread_FindsAllTextFrames tests selecting all text frames in a spread
func TestSelectAllTextFramesInSpread_FindsAllTextFrames(t *testing.T) {
	pkg, err := Read("../../testdata/plain.idml")
	if err != nil {
		t.Fatalf("Failed to read IDML: %v", err)
	}

	// Get all spreads
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Failed to get spreads: %v", err)
	}

	if len(spreads) == 0 {
		t.Skip("No spreads in test file")
	}

	// Find a spread with text frames
	var spreadFilename string
	var expectedCount int
	for filename, spread := range spreads {
		if len(spread.InnerSpread.TextFrames) > 0 {
			spreadFilename = filename
			expectedCount = len(spread.InnerSpread.TextFrames)
			break
		}
	}

	if spreadFilename == "" {
		t.Skip("No spreads with text frames found")
	}

	// Test selecting all text frames
	textFrames, err := pkg.SelectAllTextFramesInSpread(spreadFilename)
	if err != nil {
		t.Fatalf("SelectAllTextFramesInSpread() error: %v", err)
	}

	if textFrames == nil {
		t.Fatal("SelectAllTextFramesInSpread() returned nil")
	}

	if len(textFrames) != expectedCount {
		t.Errorf("Expected %d text frames, got %d", expectedCount, len(textFrames))
	}

	t.Logf("✅ Found %d text frames in spread %s", len(textFrames), spreadFilename)
}

// TestSelectByIDs_SelectsMultipleElements tests selecting multiple elements by their IDs
func TestSelectByIDs_SelectsMultipleElements(t *testing.T) {
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read IDML: %v", err)
	}

	// Get all spreads to collect some IDs
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Failed to get spreads: %v", err)
	}

	if len(spreads) == 0 {
		t.Skip("No spreads in test file")
	}

	// Collect IDs from various elements
	var ids []string
	for _, spread := range spreads {
		// Add up to 2 text frame IDs
		for i := 0; i < len(spread.InnerSpread.TextFrames) && i < 2; i++ {
			ids = append(ids, spread.InnerSpread.TextFrames[i].Self)
		}

		// Add up to 2 rectangle IDs
		for i := 0; i < len(spread.InnerSpread.Rectangles) && i < 2; i++ {
			ids = append(ids, spread.InnerSpread.Rectangles[i].Self)
		}

		if len(ids) >= 4 {
			break
		}
	}

	if len(ids) == 0 {
		t.Skip("No elements found to select")
	}

	// Test selecting by IDs
	selection, err := pkg.SelectByIDs(ids...)
	if err != nil {
		t.Fatalf("SelectByIDs() error: %v", err)
	}

	if selection == nil {
		t.Fatal("SelectByIDs() returned nil")
	}

	if selection.IsEmpty() {
		t.Error("Selection should not be empty")
	}

	totalCount := selection.Count()
	if totalCount == 0 {
		t.Error("Selection count should be > 0")
	}

	t.Logf("✅ Selected %d elements from %d IDs", totalCount, len(ids))
	t.Logf("   Text frames: %d", len(selection.TextFrames))
	t.Logf("   Rectangles: %d", len(selection.Rectangles))
}

// TestSelectByIDs_Empty tests selecting with no IDs
func TestSelectByIDs_Empty(t *testing.T) {
	pkg, err := Read("../../testdata/plain.idml")
	if err != nil {
		t.Fatalf("Failed to read IDML: %v", err)
	}

	// Test selecting with no IDs
	selection, err := pkg.SelectByIDs()
	if err != nil {
		t.Fatalf("SelectByIDs() error: %v", err)
	}

	if selection == nil {
		t.Fatal("SelectByIDs() returned nil")
	}

	if !selection.IsEmpty() {
		t.Error("Selection should be empty")
	}

	if selection.Count() != 0 {
		t.Errorf("Selection count should be 0, got %d", selection.Count())
	}

	t.Log("✅ Empty ID list returns empty selection")
}

// TestSelectByIDs_NonExistent tests selecting with non-existent IDs
func TestSelectByIDs_NonExistent(t *testing.T) {
	pkg, err := Read("../../testdata/plain.idml")
	if err != nil {
		t.Fatalf("Failed to read IDML: %v", err)
	}

	// Test selecting with non-existent IDs (should be silently skipped)
	selection, err := pkg.SelectByIDs("fake1", "fake2", "fake3")
	if err != nil {
		t.Fatalf("SelectByIDs() error: %v", err)
	}

	if selection == nil {
		t.Fatal("SelectByIDs() returned nil")
	}

	if !selection.IsEmpty() {
		t.Error("Selection should be empty for non-existent IDs")
	}

	t.Log("✅ Non-existent IDs are silently skipped")
}

// TestSelectionAddAllTypes_AddsAllElementTypes tests adding all types of elements
func TestSelectionAddAllTypes_AddsAllElementTypes(t *testing.T) {
	sel := NewSelection()

	// Add oval
	oval := &spread.Oval{
		PageItemBase: spread.PageItemBase{Self: "oval1"},
	}
	sel.AddOval(oval)
	if len(sel.Ovals) != 1 {
		t.Errorf("Expected 1 oval, got %d", len(sel.Ovals))
	}

	// Add polygon
	polygon := &spread.Polygon{
		PageItemBase: spread.PageItemBase{Self: "poly1"},
	}
	sel.AddPolygon(polygon)
	if len(sel.Polygons) != 1 {
		t.Errorf("Expected 1 polygon, got %d", len(sel.Polygons))
	}

	// Add graphic line
	line := &spread.GraphicLine{
		PageItemBase: spread.PageItemBase{Self: "line1"},
	}
	sel.AddGraphicLine(line)
	if len(sel.GraphicLines) != 1 {
		t.Errorf("Expected 1 graphic line, got %d", len(sel.GraphicLines))
	}

	// Add group
	group := &spread.Group{
		PageItemBase: spread.PageItemBase{Self: "group1"},
	}
	sel.AddGroup(group)
	if len(sel.Groups) != 1 {
		t.Errorf("Expected 1 group, got %d", len(sel.Groups))
	}

	// Verify total count
	if sel.Count() != 4 {
		t.Errorf("Expected count 4, got %d", sel.Count())
	}

	if sel.IsEmpty() {
		t.Error("Selection should not be empty")
	}
}

// TestPackageFiles_ReturnsFileList tests the Files() method
func TestPackageFiles_ReturnsFileList(t *testing.T) {
	pkg, err := Read("../../testdata/plain.idml")
	if err != nil {
		t.Fatalf("Failed to read IDML: %v", err)
	}

	files := pkg.Files()
	if len(files) == 0 {
		t.Error("Files() should return non-empty list")
	}

	// Verify expected files exist
	expectedFiles := map[string]bool{
		"mimetype":            false,
		"designmap.xml":       false,
		"Resources/Fonts.xml": false,
	}

	for _, f := range files {
		if _, ok := expectedFiles[f]; ok {
			expectedFiles[f] = true
		}
	}

	for f, found := range expectedFiles {
		if !found {
			t.Errorf("Expected file not found: %s", f)
		}
	}

	// Verify file count matches FileCount()
	if len(files) != pkg.FileCount() {
		t.Errorf("Files() count %d != FileCount() %d", len(files), pkg.FileCount())
	}
}

// TestSelectOvalByID_FindsOval tests selecting an oval by ID
func TestSelectOvalByID_FindsOval(t *testing.T) {
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read IDML: %v", err)
	}

	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Failed to get spreads: %v", err)
	}

	var testID string
	for _, spread := range spreads {
		if len(spread.InnerSpread.Ovals) > 0 {
			testID = spread.InnerSpread.Ovals[0].Self
			break
		}
	}

	if testID == "" {
		t.Skip("No ovals found in test file")
	}

	oval, err := pkg.SelectOvalByID(testID)
	if err != nil {
		t.Fatalf("SelectOvalByID() error: %v", err)
	}

	if oval == nil {
		t.Fatal("SelectOvalByID() returned nil")
	}

	if oval.Self != testID {
		t.Errorf("Expected ID '%s', got '%s'", testID, oval.Self)
	}

	t.Logf("✅ Successfully selected oval: %s", testID)
}

// TestSelectOvalByID_NotFound tests error handling for non-existent oval
func TestSelectOvalByID_NotFound(t *testing.T) {
	pkg, err := Read("../../testdata/plain.idml")
	if err != nil {
		t.Fatalf("Failed to read IDML: %v", err)
	}

	_, err = pkg.SelectOvalByID("NonExistentOvalID")
	if err == nil {
		t.Error("SelectOvalByID() should return error for non-existent ID")
	}
}

// TestSelectPolygonByID_FindsPolygon tests selecting a polygon by ID
func TestSelectPolygonByID_FindsPolygon(t *testing.T) {
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read IDML: %v", err)
	}

	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Failed to get spreads: %v", err)
	}

	var testID string
	for _, spread := range spreads {
		if len(spread.InnerSpread.Polygons) > 0 {
			testID = spread.InnerSpread.Polygons[0].Self
			break
		}
	}

	if testID == "" {
		t.Skip("No polygons found in test file")
	}

	poly, err := pkg.SelectPolygonByID(testID)
	if err != nil {
		t.Fatalf("SelectPolygonByID() error: %v", err)
	}

	if poly == nil {
		t.Fatal("SelectPolygonByID() returned nil")
	}

	if poly.Self != testID {
		t.Errorf("Expected ID '%s', got '%s'", testID, poly.Self)
	}

	t.Logf("✅ Successfully selected polygon: %s", testID)
}

// TestSelectPolygonByID_NotFound tests error handling for non-existent polygon
func TestSelectPolygonByID_NotFound(t *testing.T) {
	pkg, err := Read("../../testdata/plain.idml")
	if err != nil {
		t.Fatalf("Failed to read IDML: %v", err)
	}

	_, err = pkg.SelectPolygonByID("NonExistentPolygonID")
	if err == nil {
		t.Error("SelectPolygonByID() should return error for non-existent ID")
	}
}

// TestSelectGraphicLineByID_FindsGraphicLine tests selecting a graphic line by ID
func TestSelectGraphicLineByID_FindsGraphicLine(t *testing.T) {
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read IDML: %v", err)
	}

	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Failed to get spreads: %v", err)
	}

	var testID string
	for _, spread := range spreads {
		if len(spread.InnerSpread.GraphicLines) > 0 {
			testID = spread.InnerSpread.GraphicLines[0].Self
			break
		}
	}

	if testID == "" {
		t.Skip("No graphic lines found in test file")
	}

	line, err := pkg.SelectGraphicLineByID(testID)
	if err != nil {
		t.Fatalf("SelectGraphicLineByID() error: %v", err)
	}

	if line == nil {
		t.Fatal("SelectGraphicLineByID() returned nil")
	}

	if line.Self != testID {
		t.Errorf("Expected ID '%s', got '%s'", testID, line.Self)
	}

	t.Logf("✅ Successfully selected graphic line: %s", testID)
}

// TestSelectGraphicLineByID_NotFound tests error handling for non-existent graphic line
func TestSelectGraphicLineByID_NotFound(t *testing.T) {
	pkg, err := Read("../../testdata/plain.idml")
	if err != nil {
		t.Fatalf("Failed to read IDML: %v", err)
	}

	_, err = pkg.SelectGraphicLineByID("NonExistentLineID")
	if err == nil {
		t.Error("SelectGraphicLineByID() should return error for non-existent ID")
	}
}

// TestSelectGroupByID_FindsGroup tests selecting a group by ID
func TestSelectGroupByID_FindsGroup(t *testing.T) {
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Failed to read IDML: %v", err)
	}

	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Failed to get spreads: %v", err)
	}

	var testID string
	for _, spread := range spreads {
		if len(spread.InnerSpread.Groups) > 0 {
			testID = spread.InnerSpread.Groups[0].Self
			break
		}
	}

	if testID == "" {
		t.Skip("No groups found in test file")
	}

	group, err := pkg.SelectGroupByID(testID)
	if err != nil {
		t.Fatalf("SelectGroupByID() error: %v", err)
	}

	if group == nil {
		t.Fatal("SelectGroupByID() returned nil")
	}

	if group.Self != testID {
		t.Errorf("Expected ID '%s', got '%s'", testID, group.Self)
	}

	t.Logf("✅ Successfully selected group: %s", testID)
}

// TestSelectGroupByID_NotFound tests error handling for non-existent group
func TestSelectGroupByID_NotFound(t *testing.T) {
	pkg, err := Read("../../testdata/plain.idml")
	if err != nil {
		t.Fatalf("Failed to read IDML: %v", err)
	}

	_, err = pkg.SelectGroupByID("NonExistentGroupID")
	if err == nil {
		t.Error("SelectGroupByID() should return error for non-existent ID")
	}
}
