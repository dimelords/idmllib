package idml

import (
	"sync"
	"testing"

	"github.com/dimelords/idmllib/v2/pkg/spread"
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

	// Trigger the index build through a lookup on the fresh package
	tf, err := PageItemOfType[spread.TextFrame](pkg2, validID)
	if err != nil {
		t.Fatalf("PageItemOfType failed: %v", err)
	}

	// Verify we got the right item
	if tf.Self != validID {
		t.Errorf("expected ID %s, got %s", validID, tf.Self)
	}

	// Now index should be built
	if pkg2.indexState.index == nil {
		t.Error("index should be built after a lookup")
	}

	// Verify index contains expected items
	if func() int { n, _, _ := countItems(t, pkg2); return n }() == 0 {
		t.Error("index should contain items")
	}
}

func TestItemIndex_MatchesParsedSpreads(t *testing.T) {
	pkg, err := Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("failed to load IDML: %v", err)
	}
	if err := pkg.ensureItemIndex(); err != nil {
		t.Fatalf("ensureItemIndex: %v", err)
	}

	total, frames, rects := countItems(t, pkg)
	idx := pkg.indexState.index
	if len(idx.textFrames) != frames {
		t.Errorf("index holds %d text frames, spreads hold %d", len(idx.textFrames), frames)
	}
	if len(idx.rectangles) != rects {
		t.Errorf("index holds %d rectangles, spreads hold %d", len(idx.rectangles), rects)
	}
	indexed := len(idx.textFrames) + len(idx.rectangles) + len(idx.ovals) +
		len(idx.polygons) + len(idx.graphicLines) + len(idx.groups)
	if indexed != total {
		t.Errorf("index holds %d items, spreads hold %d", indexed, total)
	}
	if total == 0 {
		t.Error("fixture has no page items")
	}
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

	for range 10 {
		wg.Go(func() {
			_, err := PageItemOfType[spread.TextFrame](pkg2, validID)
			if err != nil {
				errors <- err
			}
		})
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
		{"TextFrame", func() error { _, err := PageItemOfType[spread.TextFrame](pkg, "nonexistent"); return err }},
		{"Rectangle", func() error { _, err := PageItemOfType[spread.Rectangle](pkg, "nonexistent"); return err }},
		{"Oval", func() error { _, err := PageItemOfType[spread.Oval](pkg, "nonexistent"); return err }},
		{"Polygon", func() error { _, err := PageItemOfType[spread.Polygon](pkg, "nonexistent"); return err }},
		{"GraphicLine", func() error { _, err := PageItemOfType[spread.GraphicLine](pkg, "nonexistent"); return err }},
		{"Group", func() error { _, err := PageItemOfType[spread.Group](pkg, "nonexistent"); return err }},
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
		_, _ = PageItemOfType[spread.TextFrame](pkg, validID)
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

// countItems counts indexed page items through the public API, replacing the
// ItemCount/TextFrameCount/RectangleCount helpers that were removed.
func countItems(t *testing.T, pkg *Package) (total, frames, rects int) {
	t.Helper()
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatalf("Spreads(): %v", err)
	}
	for _, sp := range spreads {
		frames += len(sp.TextFrames)
		rects += len(sp.Rectangles)
		total += len(sp.TextFrames) + len(sp.Rectangles) + len(sp.Ovals) +
			len(sp.Polygons) + len(sp.GraphicLines) + len(sp.Groups)
	}
	return total, frames, rects
}
