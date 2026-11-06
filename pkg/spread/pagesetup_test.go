package spread

import (
	"math"
	"strconv"
	"strings"
	"testing"
)

// A broadsheet set up in millimetres: 272 × 398 mm, margins 13/13/12/12 mm, 6 columns with a 4 mm gutter. The
// values are as InDesign writes them.
const (
	broadsheetWidth   = 771.0236220472441
	broadsheetHeight  = 1128.1889763779527
	broadsheetTop     = 36.8503937007874
	broadsheetInside  = 34.01574803149606
	broadsheetGutter  = 11.338582677165354
	broadsheetColumns = "0 107.71653543307087 119.05511811023622 226.77165354330708 238.11023622047244 345.82677165354335 " +
		"357.1653543307087 464.88188976377955 476.2204724409449 583.9370078740158 595.2755905511812 702.992125984252"
)

func floats(t *testing.T, s string) []float64 {
	t.Helper()
	var out []float64
	for f := range strings.FieldsSeq(s) {
		v, err := strconv.ParseFloat(f, 64)
		if err != nil {
			t.Fatalf("parse %q: %v", f, err)
		}
		out = append(out, v)
	}
	return out
}

func sameFloats(t *testing.T, what, got, want string) {
	t.Helper()
	g, w := floats(t, got), floats(t, want)
	if len(g) != len(w) {
		t.Fatalf("%s = %q, want %q", what, got, want)
	}
	for i := range g {
		if math.Abs(g[i]-w[i]) > 1e-9 {
			t.Fatalf("%s = %q, want %q", what, got, want)
		}
	}
}

// TestPageItemTransformMatchesInDesign uses the transforms InDesign wrote for the A4 facing parent page in
// testdata/plain.idml.
func TestPageItemTransformMatchesInDesign(t *testing.T) {
	const w, h = 595.275590551, 841.889763778
	for _, tc := range []struct {
		side PageSide
		want string
	}{
		{SideLeft, "1 0 0 1 -595.275590551 -420.944881889"},
		{SideRight, "1 0 0 1 0 -420.944881889"},
		{SideSingle, "1 0 0 1 -297.6377952755 -420.944881889"},
	} {
		sameFloats(t, "ItemTransform", PageItemTransform(w, h, tc.side), tc.want)
	}
	if got := PageGeometricBounds(w, h); got != "0 0 841.889763778 595.275590551" {
		t.Errorf("GeometricBounds = %q", got)
	}
}

func TestSetColumnsWritesPositionsAsInDesign(t *testing.T) {
	var m MarginPreference
	m.SetMargins(broadsheetTop, broadsheetTop, broadsheetInside, broadsheetInside)
	if err := m.SetColumns(6, broadsheetGutter, broadsheetWidth); err != nil {
		t.Fatal(err)
	}
	sameFloats(t, "ColumnsPositions", m.ColumnsPositions, broadsheetColumns)
	if m.ColumnCount != "6" || m.ColumnDirection != "Horizontal" {
		t.Errorf("ColumnCount %q, ColumnDirection %q", m.ColumnCount, m.ColumnDirection)
	}
	sameFloats(t, "ColumnGutter", m.ColumnGutter, "11.338582677165354")
}

func TestSetColumnsUnevenWidths(t *testing.T) {
	var m MarginPreference
	m.SetMargins(36, 36, 36, 36) // text area 400 on a 472 pt page
	if err := m.SetColumns(2, 10, 472, 250, 140); err != nil {
		t.Fatal(err)
	}
	sameFloats(t, "ColumnsPositions", m.ColumnsPositions, "0 250 260 400")

	if err := m.SetColumns(2, 10, 472, 250, 100); err == nil {
		t.Error("widths that don't fill the text area should fail")
	}
	if err := m.SetColumns(3, 10, 472, 250, 140); err == nil {
		t.Error("a width count that doesn't match the column count should fail")
	}
	if err := m.SetColumns(0, 10, 472); err == nil {
		t.Error("zero columns should fail")
	}
	if err := m.SetColumns(50, 10, 472); err == nil {
		t.Error("columns that don't fit should fail")
	}
}
