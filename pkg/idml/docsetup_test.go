package idml

import (
	"math"
	"strings"
	"testing"

	"github.com/dimelords/idmllib/v3/pkg/spread"
)

// reread writes a package and reads it back, as InDesign would get it.
func reread(t *testing.T, pkg *Package, name string) *Package {
	t.Helper()
	pkg2, err := Read(writeTestIDML(t, pkg, name))
	if err != nil {
		t.Fatal(err)
	}
	return pkg2
}

func TestDocumentSetupReadsPlain(t *testing.T) {
	ds, err := loadPlainIDML(t).DocumentSetup()
	if err != nil {
		t.Fatal(err)
	}
	if ds.PageWidth != 612 || ds.PageHeight != 792 || !ds.FacingPages || ds.PagesPerDocument != 1 {
		t.Errorf("setup %+v, want 612 × 792, facing, 1 page", ds)
	}
}

// TestSetDocumentSetupRoundTrips writes a broadsheet setup (272 × 398 mm, facing, 3 mm bleed outside) and checks
// it reads back, and that attributes the type doesn't model survive.
func TestSetDocumentSetupRoundTrips(t *testing.T) {
	pkg := loadPlainIDML(t)
	before, err := pkg.preferenceElement("DocumentPreference", false)
	if err != nil {
		t.Fatal(err)
	}
	attrsBefore := len(before.Attrs)

	want := DocumentSetup{
		PageWidth: 771.0236220472441, PageHeight: 1128.1889763779527, FacingPages: true, PagesPerDocument: 1,
		Binding: "LeftToRight", Bleed: [4]float64{0, 0, 0, 8.503937007874017},
	}
	if err := pkg.SetDocumentSetup(want); err != nil {
		t.Fatal(err)
	}
	pkg2 := reread(t, pkg, "docsetup.idml")
	got, err := pkg2.DocumentSetup()
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("read back %+v, want %+v", got, want)
	}
	after, err := pkg2.preferenceElement("DocumentPreference", false)
	if err != nil {
		t.Fatal(err)
	}
	if len(after.Attrs) < attrsBefore {
		t.Errorf("%d attributes after, %d before: unmodelled attributes were lost", len(after.Attrs), attrsBefore)
	}
	if a := toAttrMap(after.Attrs); a["DocumentBleedUniformSize"] != "false" {
		t.Errorf("DocumentBleedUniformSize = %q, want false for an uneven bleed", a["DocumentBleedUniformSize"])
	}
}

func TestSetDocumentSetupRejects(t *testing.T) {
	pkg := loadPlainIDML(t)
	if err := pkg.SetDocumentSetup(DocumentSetup{PageWidth: 0, PageHeight: 100}); err == nil {
		t.Error("a zero page width should fail")
	}
	if err := pkg.SetDocumentSetup(DocumentSetup{PageWidth: 100, PageHeight: 100, Binding: "Sideways"}); err == nil {
		t.Error("an unknown binding should fail")
	}
}

// TestSetDocumentSetupCreatesElement covers a package whose Preferences.xml has no DocumentPreference.
func TestSetDocumentSetupCreatesElement(t *testing.T) {
	pkg, err := NewFromTemplate(nil)
	if err != nil {
		t.Fatal(err)
	}
	prefs, err := pkg.Preferences()
	if err != nil {
		t.Fatal(err)
	}
	kept := prefs.Elements[:0]
	for _, el := range prefs.Elements {
		if el.XMLName.Local != "DocumentPreference" {
			kept = append(kept, el)
		}
	}
	prefs.Elements = kept
	if _, err := pkg.DocumentSetup(); err == nil {
		t.Fatal("the test package still has a DocumentPreference")
	}
	want := DocumentSetup{PageWidth: 500, PageHeight: 700, FacingPages: true, PagesPerDocument: 1}
	if err := pkg.SetDocumentSetup(want); err != nil {
		t.Fatal(err)
	}
	got, err := reread(t, pkg, "created.idml").DocumentSetup()
	if err != nil {
		t.Fatal(err)
	}
	if got.PageWidth != 500 || got.PageHeight != 700 || !got.FacingPages {
		t.Errorf("read back %+v", got)
	}
}

func TestBaselineGridRoundTrips(t *testing.T) {
	pkg := loadPlainIDML(t)
	want := BaselineGrid{Start: 38.976, Division: 10.63, RelativeTo: "TopOfMargin"}
	if err := pkg.SetBaselineGrid(want); err != nil {
		t.Fatal(err)
	}
	pkg2 := reread(t, pkg, "baseline.idml")
	got, err := pkg2.BaselineGrid()
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("read back %+v, want %+v", got, want)
	}
	prefs := string(mustFileData(t, pkg2, PathPreferences))
	if !strings.Contains(prefs, `BaselineGridRelativeOption="TopOfMarginOfBaselineGridRelativeOption"`) {
		t.Error("BaselineGridRelativeOption not written as InDesign names it")
	}
	if err := pkg.SetBaselineGrid(BaselineGrid{Division: 10, RelativeTo: "Middle"}); err == nil {
		t.Error("an unknown RelativeTo should fail")
	}
	if err := pkg.SetBaselineGrid(BaselineGrid{Division: 0}); err == nil {
		t.Error("a zero division should fail")
	}
}

func TestDefaultMarginsRoundTrip(t *testing.T) {
	pkg := loadPlainIDML(t)
	var m spread.MarginPreference
	m.SetMargins(36.8503937007874, 36.8503937007874, 34.01574803149606, 34.01574803149606)
	if err := m.SetColumns(6, 11.338582677165354, 771.0236220472441); err != nil {
		t.Fatal(err)
	}
	if err := pkg.SetDefaultMargins(m); err != nil {
		t.Fatal(err)
	}
	got, err := reread(t, pkg, "margins.idml").DefaultMargins()
	if err != nil {
		t.Fatal(err)
	}
	if got.ColumnCount != "6" || got.Top != m.Top || got.ColumnsPositions != m.ColumnsPositions {
		t.Errorf("read back %+v, want %+v", got, m)
	}
	if math.Abs(toAttrMap(nil).float("x")) != 0 {
		t.Error("float of a missing attribute should be 0")
	}
}
