package common

import (
	"encoding/xml"
	"strings"
	"testing"
)

const compoundXML = `<PathGeometry X="1">` +
	`<GeometryPathType PathOpen="false"><PathPointArray><PathPointType Anchor="0 0"/><PathPointType Anchor="10 0"/><PathPointType Anchor="10 10"/></PathPointArray></GeometryPathType>` +
	`<Unknown Self="z"/>` +
	`<GeometryPathType PathOpen="false"><PathPointArray><PathPointType Anchor="2 2"/><PathPointType Anchor="8 2"/></PathPointArray></GeometryPathType>` +
	`<GeometryPathType PathOpen="true"><PathPointArray><PathPointType Anchor="4 4"/></PathPointArray></GeometryPathType>` +
	`</PathGeometry>`

// A compound shape's paths are each kept, in order, and written back as they were: encoding/xml used to decode them
// all into one GeometryPathType, appending the points.
func TestPathGeometryKeepsEveryPath(t *testing.T) {
	var g PathGeometry
	if err := xml.Unmarshal([]byte(compoundXML), &g); err != nil {
		t.Fatal(err)
	}
	paths := g.Paths()
	if len(paths) != 3 || len(g.MorePaths) != 2 || len(g.OtherElements) != 1 || len(g.OtherAttrs) != 1 {
		t.Fatalf("%d paths, %d more, %d other, %d attrs", len(paths), len(g.MorePaths), len(g.OtherElements), len(g.OtherAttrs))
	}
	for i, want := range []int{3, 2, 1} {
		if n := len(paths[i].PathPointArray.PathPoints); n != want {
			t.Errorf("path %d has %d points, want %d", i, n, want)
		}
	}
	if paths[2].PathOpen != "true" {
		t.Errorf("third path open %q", paths[2].PathOpen)
	}
	out, err := xml.Marshal(g)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.ReplaceAll(string(out), "></PathPointType>", "/>"); got != strings.ReplaceAll(compoundXML, `<Unknown Self="z"/>`, `<Unknown Self="z"></Unknown>`) {
		t.Errorf("written back as\n%s\nwant\n%s", got, compoundXML)
	}

	// Paths points into the geometry, so a caller can change any path.
	paths[1].PathPointArray.PathPoints[0].Anchor = "3 3"
	if g.MorePaths[0].PathPointArray.PathPoints[0].Anchor != "3 3" {
		t.Error("Paths doesn't point into MorePaths")
	}
}

// One path reads and writes as before; a geometry built in code without a recorded order writes its paths first.
func TestPathGeometrySinglePathAndBuilt(t *testing.T) {
	one := `<PathGeometry><GeometryPathType PathOpen="false"><PathPointArray><PathPointType Anchor="0 0"></PathPointType></PathPointArray></GeometryPathType></PathGeometry>`
	var g PathGeometry
	if err := xml.Unmarshal([]byte(one), &g); err != nil {
		t.Fatal(err)
	}
	if len(g.Paths()) != 1 || g.MorePaths != nil {
		t.Fatalf("paths %d", len(g.Paths()))
	}
	if out, _ := xml.Marshal(g); string(out) != one {
		t.Errorf("written back as %s", out)
	}
	built := PathGeometry{GeometryPathType: &GeometryPathType{PathOpen: "false"}, MorePaths: []GeometryPathType{{PathOpen: "true"}}}
	out, err := xml.Marshal(built)
	if err != nil {
		t.Fatal(err)
	}
	if want := `<PathGeometry><GeometryPathType PathOpen="false"></GeometryPathType><GeometryPathType PathOpen="true"></GeometryPathType></PathGeometry>`; string(out) != want {
		t.Errorf("built geometry written as %s", out)
	}
	var none *PathGeometry
	if none.Paths() != nil {
		t.Error("a nil geometry has paths")
	}
}
