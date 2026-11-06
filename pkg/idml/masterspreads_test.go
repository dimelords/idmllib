package idml

import (
	"strings"
	"testing"

	"github.com/dimelords/idmllib/v3/pkg/spread"
)

func newSportMaster(t *testing.T, self string) *spread.Spread {
	t.Helper()
	ms, err := spread.NewMasterSpread(spread.MasterSpreadOptions{
		Self: self, NamePrefix: "S", BaseName: "Sport",
		PageWidth: 595.275590551, PageHeight: 841.889763778, Facing: true,
		Top: 36, Bottom: 36, Inside: 30, Outside: 40, Columns: 5, Gutter: 12,
	})
	if err != nil {
		t.Fatal(err)
	}
	return ms
}

// TestAddMasterSpreadRegistersAndRoundTrips checks that a parent page added through the API is referenced from
// designmap.xml as an <idPkg:MasterSpread> (not a Spread) and reads back with its pages and grid.
func TestAddMasterSpreadRegistersAndRoundTrips(t *testing.T) {
	pkg := loadPlainIDML(t)
	filename := MasterSpreadPath("umsport")
	if err := pkg.AddMasterSpread(filename, newSportMaster(t, "umsport")); err != nil {
		t.Fatalf("AddMasterSpread: %v", err)
	}

	pkg2 := reread(t, pkg, "addmaster.idml")
	designmap := string(mustFileData(t, pkg2, PathDesignmap))
	if !strings.Contains(designmap, `<idPkg:MasterSpread src="`+filename+`"`) {
		t.Errorf("designmap.xml has no idPkg:MasterSpread element for %s", filename)
	}
	if strings.Contains(designmap, `<idPkg:Spread src="`+filename+`"`) {
		t.Errorf("the master spread was registered as a Spread")
	}
	masters, err := pkg2.MasterSpreads()
	if err != nil {
		t.Fatal(err)
	}
	if len(masters) != 2 {
		t.Errorf("%d master spreads, want the original one plus the new one", len(masters))
	}
	ms, err := pkg2.MasterSpread(filename)
	if err != nil {
		t.Fatal(err)
	}
	if !ms.IsMaster() || len(ms.Pages) != 2 {
		t.Fatalf("IsMaster %v, %d pages", ms.IsMaster(), len(ms.Pages))
	}
	if name, _, _ := ms.MasterName(); name != "S-Sport" {
		t.Errorf("name %q", name)
	}
	mp := ms.Pages[1].MarginPreference
	if mp == nil || mp.ColumnCount != "5" || mp.Left != "30" || mp.Right != "40" {
		t.Errorf("margin preference %+v", mp)
	}
}

func TestAddMasterSpreadRejects(t *testing.T) {
	pkg := loadPlainIDML(t)
	if err := pkg.AddMasterSpread(MasterSpreadPath("ua"), newSportMaster(t, "ua")); err != nil {
		t.Fatal(err)
	}
	if err := pkg.AddMasterSpread(MasterSpreadPath("ua"), newSportMaster(t, "ub")); err == nil {
		t.Error("adding to an existing file should fail")
	}
	if err := pkg.AddMasterSpread(MasterSpreadPath("uz"), newSportMaster(t, "ubb")); err == nil {
		t.Error("reusing the Self of plain.idml's master spread should fail")
	}
	notMaster := newTestSpread("unm")
	if err := pkg.AddMasterSpread(MasterSpreadPath("unm"), notMaster); err == nil {
		t.Error("a regular spread should be refused")
	}
	if err := pkg.AddMasterSpread(SpreadPath("ux"), newSportMaster(t, "ux")); err == nil {
		t.Error("a Spreads/ path should be refused")
	}
}

func TestApplyMaster(t *testing.T) {
	pkg := loadPlainIDML(t)
	if err := pkg.AddMasterSpread(MasterSpreadPath("umsport"), newSportMaster(t, "umsport")); err != nil {
		t.Fatal(err)
	}
	const sp = "Spreads/Spread_ud3.xml"
	if err := pkg.ApplyMaster(sp, "ud8", "umsport"); err != nil {
		t.Fatalf("ApplyMaster: %v", err)
	}
	if err := pkg.ApplyMaster(sp, "ud8", "unknown"); err == nil {
		t.Error("an unknown master spread should fail")
	}
	if err := pkg.ApplyMaster(sp, "unknown", "umsport"); err == nil {
		t.Error("an unknown page should fail")
	}
	got, err := reread(t, pkg, "apply.idml").Spread(sp)
	if err != nil {
		t.Fatal(err)
	}
	if got.Pages[0].AppliedMaster != "umsport" {
		t.Errorf("AppliedMaster = %q, want umsport", got.Pages[0].AppliedMaster)
	}
}

func TestRemoveMasterSpread(t *testing.T) {
	pkg := loadPlainIDML(t)
	const filename = "MasterSpreads/MasterSpread_ubb.xml" // applied to page ud8
	err := pkg.RemoveMasterSpread(filename, false)
	if err == nil || !strings.Contains(err.Error(), "ud8") {
		t.Fatalf("removing an applied master without detach: err %v, want one naming page ud8", err)
	}
	if err := pkg.RemoveMasterSpread(filename, true); err != nil {
		t.Fatalf("RemoveMasterSpread(detach): %v", err)
	}

	pkg2 := reread(t, pkg, "removemaster.idml")
	if pkg2.hasFile(filename) {
		t.Errorf("%s is still in the package", filename)
	}
	if strings.Contains(string(mustFileData(t, pkg2, PathDesignmap)), filename) {
		t.Errorf("designmap.xml still references %s", filename)
	}
	sp, err := pkg2.Spread("Spreads/Spread_ud3.xml")
	if err != nil {
		t.Fatal(err)
	}
	if sp.Pages[0].AppliedMaster != "n" {
		t.Errorf("page AppliedMaster = %q, want n", sp.Pages[0].AppliedMaster)
	}
}

func TestRemoveSpreadMovesSection(t *testing.T) {
	pkg := loadPlainIDML(t)
	const original = "Spreads/Spread_ud3.xml"
	if err := pkg.RemoveSpread(original); err == nil {
		t.Fatal("removing the only spread should fail")
	}
	if err := pkg.AddSpread(SpreadPath("u2"), newTestSpread("u2")); err != nil {
		t.Fatal(err)
	}
	if err := pkg.RemoveSpread(original); err != nil {
		t.Fatalf("RemoveSpread: %v", err)
	}

	pkg2 := reread(t, pkg, "removespread.idml")
	if pkg2.hasFile(original) {
		t.Errorf("%s is still in the package", original)
	}
	doc, err := pkg2.Document()
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Spreads) != 1 || doc.Spreads[0].Src != SpreadPath("u2") {
		t.Errorf("spreads %+v, want only %s", doc.Spreads, SpreadPath("u2"))
	}
	if len(doc.Sections) != 1 || doc.Sections[0].PageStart != "up_u2" {
		t.Errorf("sections %+v, want one starting on page up_u2", doc.Sections)
	}
}
