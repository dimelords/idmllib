package idml

import (
	"sync"
	"testing"
)

func TestItemIndex_LazyBuild(t *testing.T) {
	// Load a package without triggering the index
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("failed to load IDML: %v", err)
	}

	// Index should not be built yet
	if pkg.indexState.index != nil {
		t.Error("index should be nil before first access")
	}

	// First, build the index and get a valid ID
	err = pkg.ensureItemIndex()
	if err != nil {
		t.Fatalf("failed to build index: %v", err)
	}

	// Find a valid text frame ID
	var validID string
	for id := range pkg.indexState.index.textFrames {
		validID = id
		break
	}
	if validID == "" {
		t.Fatal("no text frames found in test document")
	}
	t.Logf("Using valid text frame ID: %s", validID)

	// Load fresh package for actual test
	pkg2, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("failed to load IDML: %v", err)
	}

	// Index should not be built yet
	if pkg2.indexState.index != nil {
		t.Error("index should be nil before first access")
	}

	// Trigger index build via SelectTextFrameByID
	tf, err := pkg2.SelectTextFrameByID(validID)
	if err != nil {
		t.Fatalf("SelectTextFrameByID failed: %v", err)
	}

	// Verify we got the right item
	if tf.Self != validID {
		t.Errorf("expected ID %s, got %s", validID, tf.Self)
	}

	// Now index should be built
	if pkg2.indexState.index == nil {
		t.Error("index should be built after SelectTextFrameByID")
	}

	// Verify index contains expected items
	if pkg2.ItemCount() == 0 {
		t.Error("index should contain items")
	}
}

func TestItemIndex_ItemCounts(t *testing.T) {
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("failed to load IDML: %v", err)
	}

	// Before index is built, counts should be 0
	if pkg.ItemCount() != 0 {
		t.Error("ItemCount should be 0 before index is built")
	}
	if pkg.TextFrameCount() != 0 {
		t.Error("TextFrameCount should be 0 before index is built")
	}
	if pkg.RectangleCount() != 0 {
		t.Error("RectangleCount should be 0 before index is built")
	}

	// Trigger index build
	_ = pkg.ensureItemIndex()

	// Now counts should be positive
	if pkg.ItemCount() == 0 {
		t.Error("ItemCount should be positive after index build")
	}
	t.Logf("Total items indexed: %d", pkg.ItemCount())
	t.Logf("  TextFrames: %d", pkg.TextFrameCount())
	t.Logf("  Rectangles: %d", pkg.RectangleCount())
}

func TestItemIndex_ConcurrentAccess(t *testing.T) {
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("failed to load IDML: %v", err)
	}

	// First get a valid ID
	err = pkg.ensureItemIndex()
	if err != nil {
		t.Fatalf("failed to build index: %v", err)
	}
	var validID string
	for id := range pkg.indexState.index.textFrames {
		validID = id
		break
	}
	if validID == "" {
		t.Fatal("no text frames found")
	}

	// Load fresh package for concurrent test
	pkg2, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("failed to load IDML: %v", err)
	}

	// Run concurrent selections
	var wg sync.WaitGroup
	errors := make(chan error, 100)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := pkg2.SelectTextFrameByID(validID)
			if err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("concurrent access failed: %v", err)
	}
}

func TestItemIndex_AllTypes(t *testing.T) {
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("failed to load IDML: %v", err)
	}

	// Build index
	err = pkg.ensureItemIndex()
	if err != nil {
		t.Fatalf("failed to build index: %v", err)
	}

	// Check that index has entries for each type
	idx := pkg.indexState.index

	t.Logf("Index contents:")
	t.Logf("  TextFrames: %d", len(idx.textFrames))
	t.Logf("  Rectangles: %d", len(idx.rectangles))
	t.Logf("  Ovals: %d", len(idx.ovals))
	t.Logf("  Polygons: %d", len(idx.polygons))
	t.Logf("  GraphicLines: %d", len(idx.graphicLines))
	t.Logf("  Groups: %d", len(idx.groups))

	// Verify each map lookup returns correct pointer
	for id, tf := range idx.textFrames {
		if tf.Self != id {
			t.Errorf("textFrame ID mismatch: map key %q != Self %q", id, tf.Self)
		}
	}

	for id, rect := range idx.rectangles {
		if rect.Self != id {
			t.Errorf("rectangle ID mismatch: map key %q != Self %q", id, rect.Self)
		}
	}
}

func TestItemIndex_NotFoundErrors(t *testing.T) {
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("failed to load IDML: %v", err)
	}

	testCases := []struct {
		name   string
		lookup func() error
	}{
		{"TextFrame", func() error { _, err := pkg.SelectTextFrameByID("nonexistent"); return err }},
		{"Rectangle", func() error { _, err := pkg.SelectRectangleByID("nonexistent"); return err }},
		{"Oval", func() error { _, err := pkg.SelectOvalByID("nonexistent"); return err }},
		{"Polygon", func() error { _, err := pkg.SelectPolygonByID("nonexistent"); return err }},
		{"GraphicLine", func() error { _, err := pkg.SelectGraphicLineByID("nonexistent"); return err }},
		{"Group", func() error { _, err := pkg.SelectGroupByID("nonexistent"); return err }},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.lookup()
			if err == nil {
				t.Errorf("%s: expected error for nonexistent ID", tc.name)
			}
		})
	}
}

func BenchmarkSelectTextFrameByID_WithIndex(b *testing.B) {
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		b.Fatalf("failed to load IDML: %v", err)
	}

	// Pre-build index and get valid ID
	_ = pkg.ensureItemIndex()
	var validID string
	for id := range pkg.indexState.index.textFrames {
		validID = id
		break
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = pkg.SelectTextFrameByID(validID)
	}
}

func BenchmarkSelectByIDs_WithIndex(b *testing.B) {
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		b.Fatalf("failed to load IDML: %v", err)
	}

	// Pre-build index and collect valid IDs
	_ = pkg.ensureItemIndex()
	var ids []string
	count := 0
	for id := range pkg.indexState.index.textFrames {
		ids = append(ids, id)
		count++
		if count >= 4 {
			break
		}
	}
	for id := range pkg.indexState.index.rectangles {
		ids = append(ids, id)
		count++
		if count >= 8 {
			break
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = pkg.SelectByIDs(ids...)
	}
}
