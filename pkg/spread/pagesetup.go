package spread

import (
	"math"
	"strconv"
	"strings"

	"github.com/dimelords/idmllib/v3/pkg/common"
)

// PageSide says where a page sits in its spread.
type PageSide int

const (
	// SideRight is the right-hand page of a facing-pages spread (odd page numbers in a left-to-right document).
	SideRight PageSide = iota
	// SideLeft is the left-hand page of a facing-pages spread.
	SideLeft
	// SideSingle is a page of a single-sided document.
	SideSingle
)

// columnTolerance is how far, in points, explicit column widths plus gutters may differ from the text area
// width. InDesign stores millimetre values as long decimals, so exact equality is too strict.
const columnTolerance = 0.01

// PageGeometricBounds returns the GeometricBounds of a page: "0 0 height width".
func PageGeometricBounds(pageWidth, pageHeight float64) string {
	return formatFloats(0, 0, pageHeight, pageWidth)
}

// PageItemTransform returns the ItemTransform of a page within its spread, the way InDesign writes it. A spread's
// origin is its centre: on facing pages the binding, so a left page is moved one page width left of it and a
// right page starts at it; a single-sided page is centred on it. Vertically, every page is centred.
func PageItemTransform(pageWidth, pageHeight float64, side PageSide) string {
	tx := 0.0
	switch side {
	case SideLeft:
		tx = -pageWidth
	case SideSingle:
		tx = -pageWidth / 2
	}
	return formatFloats(1, 0, 0, 1, tx, -pageHeight/2)
}

// SetMargins sets the page margins in points. On facing pages Left is the inside margin (towards the binding)
// and Right the outside margin, as in InDesign's object model; on single-sided pages they are left and right.
func (m *MarginPreference) SetMargins(top, bottom, inside, outside float64) {
	m.Top = formatFloat(top)
	m.Bottom = formatFloat(bottom)
	m.Left = formatFloat(inside)
	m.Right = formatFloat(outside)
}

// SetColumns sets the column grid of a page that is pageWidth points wide. Call SetMargins first: the columns
// fill the width between the margins. Without widths the columns are equal; with widths (one per column) they
// may differ, but widths and gutters must add up to the text area width. ColumnsPositions is written as
// InDesign does: start and end of each column, measured from the left margin.
func (m *MarginPreference) SetColumns(count int, gutter, pageWidth float64, widths ...float64) error {
	if count < 1 {
		return common.Errorf("spread", "set columns", "", "column count %d: want at least 1", count)
	}
	if gutter < 0 {
		return common.Errorf("spread", "set columns", "", "negative gutter %v", gutter)
	}
	text := pageWidth - parseFloat(m.Left) - parseFloat(m.Right)
	if text <= 0 {
		return common.Errorf("spread", "set columns", "", "no room for columns: page width %v minus margins %s and %s", pageWidth, m.Left, m.Right)
	}
	if len(widths) == 0 {
		w := (text - float64(count-1)*gutter) / float64(count)
		if w <= 0 {
			return common.Errorf("spread", "set columns", "", "%d columns with gutter %v don't fit in %v points", count, gutter, text)
		}
		widths = make([]float64, count)
		for i := range widths {
			widths[i] = w
		}
	}
	if len(widths) != count {
		return common.Errorf("spread", "set columns", "", "%d column widths for %d columns", len(widths), count)
	}
	span := float64(count-1) * gutter
	for _, w := range widths {
		if w <= 0 {
			return common.Errorf("spread", "set columns", "", "column width %v: want more than 0", w)
		}
		span += w
	}
	if math.Abs(span-text) > columnTolerance {
		return common.Errorf("spread", "set columns", "", "columns and gutters span %v points, the text area is %v", span, text)
	}

	positions := make([]float64, 0, 2*count)
	x := 0.0
	for i, w := range widths {
		if i > 0 {
			x += gutter
		}
		positions = append(positions, x, x+w)
		x += w
	}
	m.ColumnCount = strconv.Itoa(count)
	m.ColumnGutter = formatFloat(gutter)
	m.ColumnsPositions = formatFloats(positions...)
	if m.ColumnDirection == "" {
		m.ColumnDirection = "Horizontal"
	}
	return nil
}

func formatFloat(f float64) string {
	if f == 0 {
		return "0" // avoids "-0"
	}
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func formatFloats(fs ...float64) string {
	parts := make([]string, len(fs))
	for i, f := range fs {
		parts[i] = formatFloat(f)
	}
	return strings.Join(parts, " ")
}

func parseFloat(s string) float64 {
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0
	}
	return f
}
