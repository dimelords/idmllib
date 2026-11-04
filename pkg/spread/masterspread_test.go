package spread

import "testing"

func TestNewMasterSpreadFacing(t *testing.T) {
	ms, err := NewMasterSpread(MasterSpreadOptions{
		Self: "umstd", NamePrefix: "A", BaseName: "Standard",
		PageWidth: broadsheetWidth, PageHeight: broadsheetHeight, Facing: true,
		Top: broadsheetTop, Bottom: broadsheetTop, Inside: broadsheetInside, Outside: broadsheetInside,
		Columns: 6, Gutter: broadsheetGutter,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !ms.IsMaster() || ms.PageCount != "2" || len(ms.Pages) != 2 {
		t.Fatalf("IsMaster %v, PageCount %q, %d pages", ms.IsMaster(), ms.PageCount, len(ms.Pages))
	}
	if name, prefix, base := ms.MasterName(); name != "A-Standard" || prefix != "A" || base != "Standard" {
		t.Errorf("MasterName = %q %q %q", name, prefix, base)
	}
	left, right := ms.Pages[0], ms.Pages[1]
	if left.Self != "umstdp1" || right.Self != "umstdp2" || left.Name != "A" || left.AppliedMaster != "n" {
		t.Errorf("pages %q %q, name %q, applied %q", left.Self, right.Self, left.Name, left.AppliedMaster)
	}
	sameFloats(t, "left ItemTransform", left.ItemTransform, "1 0 0 1 -771.0236220472441 -564.0944881889764")
	sameFloats(t, "right ItemTransform", right.ItemTransform, "1 0 0 1 0 -564.0944881889764")
	sameFloats(t, "ColumnsPositions", right.MarginPreference.ColumnsPositions, broadsheetColumns)

	ms.SetMasterName("S", "Sport")
	if name, _, _ := ms.MasterName(); name != "S-Sport" || ms.Pages[0].Name != "S" {
		t.Errorf("after rename: %q, page name %q", name, ms.Pages[0].Name)
	}
}

func TestNewMasterSpreadSingleAndErrors(t *testing.T) {
	ms, err := NewMasterSpread(MasterSpreadOptions{
		Self: "um1", NamePrefix: "A", BaseName: "Master", PageWidth: 612, PageHeight: 792,
		Top: 36, Bottom: 36, Inside: 36, Outside: 36, PageSelfs: []string{"upg"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(ms.Pages) != 1 || ms.Pages[0].Self != "upg" || ms.Pages[0].MarginPreference.ColumnCount != "1" {
		t.Errorf("single page master: %+v", ms.Pages)
	}
	sameFloats(t, "ItemTransform", ms.Pages[0].ItemTransform, "1 0 0 1 -306 -396")

	for name, opts := range map[string]MasterSpreadOptions{
		"no self":   {NamePrefix: "A", BaseName: "M", PageWidth: 1, PageHeight: 1},
		"no size":   {Self: "u", NamePrefix: "A", BaseName: "M"},
		"page ids":  {Self: "u", NamePrefix: "A", BaseName: "M", PageWidth: 100, PageHeight: 100, Facing: true, PageSelfs: []string{"x"}},
		"too small": {Self: "u", NamePrefix: "A", BaseName: "M", PageWidth: 60, PageHeight: 100, Inside: 30, Outside: 30},
	} {
		if _, err := NewMasterSpread(opts); err == nil {
			t.Errorf("%s: want an error", name)
		}
	}
}
