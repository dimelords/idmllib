package idml

import (
	"testing"

	"github.com/dimelords/idmllib/v3/pkg/common"
)

// compoundpaths.idml was written by InDesign 2026: a ring (an oval with a smaller one cut out, two paths), the
// letters BOB converted to outlines (one polygon of eight paths: B and O have counters), and a group of a ring and
// a plain rectangle.
const compoundPathsFixture = "../../testdata/compoundpaths.idml"

func pathCounts(g *common.PathGeometry) []int {
	var out []int
	for _, p := range g.Paths() {
		out = append(out, len(p.PathPointArray.PathPoints))
	}
	return out
}

func TestReadCompoundPaths(t *testing.T) {
	pkg, err := Read(compoundPathsFixture)
	if err != nil {
		t.Fatal(err)
	}
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatal(err)
	}
	var ring, letters, grouped []int
	for _, sp := range spreads {
		for i := range sp.Ovals {
			ring = pathCounts(sp.Ovals[i].Properties.PathGeometry)
		}
		for i := range sp.Polygons {
			letters = pathCounts(sp.Polygons[i].Properties.PathGeometry)
		}
		for i := range sp.Groups {
			for j := range sp.Groups[i].Ovals {
				grouped = pathCounts(sp.Groups[i].Ovals[j].Properties.PathGeometry)
			}
		}
	}
	if len(ring) != 2 || ring[0] != 4 || ring[1] != 4 {
		t.Errorf("ring paths %v, want two of 4 points", ring)
	}
	total := 0
	for _, n := range letters {
		total += n
	}
	if len(letters) != 8 || total != 58 {
		t.Errorf("letters paths %v (%d points), want 8 paths of 58 points", letters, total)
	}
	if len(grouped) != 2 {
		t.Errorf("grouped ring paths %v, want 2", grouped)
	}
}
