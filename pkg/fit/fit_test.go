package fit

import (
	"encoding/xml"
	"math"
	"strconv"
	"testing"

	"github.com/dimelords/idmllib/v3/pkg/common"
	"github.com/dimelords/idmllib/v3/pkg/resources"
	"github.com/dimelords/idmllib/v3/pkg/spread"
)

// With a nil Measurer every character is half the type size, so at 12pt a
// character is 6pt. That makes the expectations below exact rather than
// approximate.

func TestEvaluate(t *testing.T) {
	base := Constraints{FrameWidth: 120, FrameHeight: 40, PointSize: 12, Leading: 14.4, MaxLines: 1}

	tests := []struct {
		name string
		text string
		c    Constraints
		want Failure // empty means it fits
	}{
		{"fits on one line", "kort text", base, ""},
		{
			"needs a second line it does not have",
			"en betydligt langre rubrik an vad som far plats", base, FailTooManyLines,
		},
		{
			// 23 characters at 6pt is 138pt in a 120pt frame. No break
			// helps; Swedish compounds hit this routinely.
			"a single compound wider than the frame",
			"Funktionshinderomsorgen", base, FailWordTooWide,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Evaluate(tt.text, tt.c, nil)
			if tt.want == "" {
				if !got.Fits {
					t.Errorf("Fits = false, want true: %s", got.Reason)
				}
				return
			}
			if got.Fits {
				t.Fatalf("Fits = true, want failure %q", tt.want)
			}
			if got.Why != tt.want {
				t.Errorf("Why = %q, want %q (%s)", got.Why, tt.want, got.Reason)
			}
			if got.Reason == "" {
				t.Error("no Reason given; the caller cannot act on this")
			}
		})
	}
}

// TestEvaluate_ReportsEstimation is the honesty check. A verdict measured
// with an average character width can name the wrong word or reject text
// that fits, so a caller has to be able to tell the two apart.
func TestEvaluate_ReportsEstimation(t *testing.T) {
	c := Constraints{FrameWidth: 120, FrameHeight: 40, PointSize: 12, Leading: 14.4, MaxLines: 1}

	if got := Evaluate("kort", c, nil); !got.Estimated {
		t.Error("Estimated = false without a Measurer")
	}
	if got := Evaluate("kort", c, fixedWidth(3)); got.Estimated {
		t.Error("Estimated = true although a Measurer was supplied")
	}
}

func TestEvaluate_Fold(t *testing.T) {
	// A broadsheet headline frame: 1491pt across a spread, spine at its
	// midpoint. At 72pt a character is 36pt with the nil measurer.
	c := Constraints{
		FrameWidth: 1491, FrameHeight: 80, PointSize: 72, Leading: 72,
		MaxLines: 1, SpansFold: true, FoldX: 745.5,
	}

	t.Run("a word on the spine is rejected", func(t *testing.T) {
		got := Evaluate("aaaaaaaaaaaaaaaaaaaaaaaaaa bb", c, nil)
		if got.Fits {
			t.Fatal("accepted a word the spine runs through")
		}
		if got.Why != FailWordOnFold {
			t.Fatalf("Why = %q, want %q", got.Why, FailWordOnFold)
		}
		if got.Word == "" {
			t.Error("no Word named; the wording cannot be corrected blind")
		}
	})

	t.Run("a break on the spine is accepted", func(t *testing.T) {
		// 20 characters is 720pt, then a space spanning 720..756, which
		// contains the fold at 745.5.
		if got := Evaluate("aaaaaaaaaaaaaaaaaaaa bbbbbbbbbbbbbbbbbbbb", c, nil); !got.Fits {
			t.Errorf("rejected a break landing on the fold: %s", got.Reason)
		}
	})

	t.Run("text ending before the spine is fine", func(t *testing.T) {
		if got := Evaluate("kort", c, nil); !got.Fits {
			t.Errorf("rejected text that never reaches the fold: %s", got.Reason)
		}
	})

	t.Run("a gutter at the spine lifts the constraint", func(t *testing.T) {
		// The same text, in a frame split into columns with a gutter. The
		// layout breaks at the fold by itself, so the wording is free -
		// without this, every headline on a correctly built template
		// would be rejected for a problem the template had solved.
		protected := c
		protected.FoldProtected = true
		if got := Evaluate("aaaaaaaaaaaaaaaaaaaaaaaaaa bb", protected, nil); !got.Fits {
			t.Errorf("rejected text in a gutter-protected frame: %s", got.Reason)
		}
	})
}

func TestBudget(t *testing.T) {
	c := Constraints{FrameWidth: 120, PointSize: 12, MaxLines: 1}
	// 10 characters at 6pt each is 60pt, so 120pt holds 20.
	if got, want := Budget("abcdefghij", c, nil), 20; got != want {
		t.Errorf("one line: %d, want %d", got, want)
	}
	// A second line roughly doubles it, which is the single biggest lever
	// on how long a headline may be.
	c.MaxLines = 2
	if got, want := Budget("abcdefghij", c, nil), 40; got != want {
		t.Errorf("two lines: %d, want %d", got, want)
	}
	if got := Budget("", c, nil); got != 0 {
		t.Errorf("empty sample: %d, want 0", got)
	}
}

// TestStyleResolution covers the inheritance that a renderer reading only
// a style's own attributes gets wrong. A house style routinely carries no
// size of its own.
func TestStyleResolution(t *testing.T) {
	styles := &resources.StylesFile{
		RootParagraphStyleGroup: &resources.ParagraphStyleGroup{
			NestedGroups: []resources.ParagraphStyleGroup{{
				Self: "ParagraphStyleGroup/$ID/House",
				// Two groups deep, as real templates organise them.
				NestedGroups: []resources.ParagraphStyleGroup{{
					Self: "ParagraphStyleGroup/$ID/House%3aStandard",
					ParagraphStyles: []resources.ParagraphStyle{{
						Self:       "ParagraphStyle/headline",
						FontStyle:  "Light Semicondensed",
						Properties: props(leading(66), appliedFont("Kepler Std")),
						PointSize:  "72",
					}, {
						// No size of its own: it comes from the root.
						Self:       "ParagraphStyle/dateline",
						Properties: props(basedOn("$ID/[No paragraph style]")),
					}},
				}},
			}},
			ParagraphStyles: []resources.ParagraphStyle{{
				Self:      "ParagraphStyle/$ID/[No paragraph style]",
				PointSize: "12",
			}},
		},
	}

	if got := StyleSize(styles, "ParagraphStyle/headline"); got != 72 {
		t.Errorf("headline size = %v, want 72", got)
	}
	if got := StyleLeading(styles, "ParagraphStyle/headline", 72); got != 66 {
		t.Errorf("headline leading = %v, want 66", got)
	}
	fam, style := StyleFont(styles, "ParagraphStyle/headline")
	if fam != "Kepler Std" || style != "Light Semicondensed" {
		t.Errorf("headline font = %q %q", fam, style)
	}

	// The inherited case, and the reason BasedOn needs its prefix
	// restored: Self carries "ParagraphStyle/" and BasedOn does not.
	if got := StyleSize(styles, "ParagraphStyle/dateline"); got != 12 {
		t.Errorf("dateline size = %v, want 12 inherited through BasedOn", got)
	}
	// No leading anywhere in the chain means automatic leading.
	if got := StyleLeading(styles, "ParagraphStyle/dateline", 12); math.Abs(got-14.4) > 0.001 {
		t.Errorf("dateline leading = %v, want automatic 14.4", got)
	}
	if got := StyleSize(styles, "ParagraphStyle/missing"); got != 0 {
		t.Errorf("unknown style = %v, want 0", got)
	}
}

// TestStyleResolution_SurvivesACycle guards the walk: a damaged document
// must not hang a service.
func TestStyleResolution_SurvivesACycle(t *testing.T) {
	styles := &resources.StylesFile{
		RootParagraphStyleGroup: &resources.ParagraphStyleGroup{
			ParagraphStyles: []resources.ParagraphStyle{
				{Self: "ParagraphStyle/a", Properties: props(basedOn("b"))},
				{Self: "ParagraphStyle/b", Properties: props(basedOn("a"))},
			},
		},
	}
	if got := StyleSize(styles, "ParagraphStyle/a"); got != 0 {
		t.Errorf("got %v, want 0", got)
	}
}

func TestFrameConstraints(t *testing.T) {
	styles := &resources.StylesFile{
		RootParagraphStyleGroup: &resources.ParagraphStyleGroup{
			ParagraphStyles: []resources.ParagraphStyle{{
				Self: "ParagraphStyle/headline", PointSize: "72",
				Properties: props(leading(66)),
			}},
		},
	}

	// A headline frame spanning a spread: 1491pt wide, its left edge
	// 745.5pt left of the spine at x=0. The local origin sits inside the
	// frame, as InDesign leaves it, so the transform is the difference
	// rather than the edge itself.
	tf := frame("hl", "1 0 0 1 -588 -500", -157.5, 1333.5, -9, 59.4)

	c, err := FrameConstraints(tf, styles, "ParagraphStyle/headline", 0)
	if err != nil {
		t.Fatalf("FrameConstraints: %v", err)
	}
	if got, want := c.FrameWidth, 1491.0; got != want {
		t.Errorf("FrameWidth = %v, want %v", got, want)
	}
	if c.PointSize != 72 || c.Leading != 66 {
		t.Errorf("size/leading = %v/%v, want 72/66", c.PointSize, c.Leading)
	}
	if c.MaxLines != 1 {
		t.Errorf("MaxLines = %d, want 1", c.MaxLines)
	}
	if !c.SpansFold {
		t.Error("SpansFold = false for a frame straddling the spine")
	}
	if got, want := c.FoldX, 745.5; math.Abs(got-want) > 0.01 {
		t.Errorf("FoldX = %v, want %v", got, want)
	}
	if c.FoldProtected {
		t.Error("FoldProtected = true for a single-column frame")
	}

	t.Run("a single-page spread has no spine", func(t *testing.T) {
		c, err := FrameConstraints(tf, styles, "ParagraphStyle/headline", math.Inf(1))
		if err != nil {
			t.Fatal(err)
		}
		if c.SpansFold || c.FoldX != 0 {
			t.Errorf("SpansFold=%v FoldX=%v, want neither", c.SpansFold, c.FoldX)
		}
	})

	t.Run("columns with a gutter protect the fold", func(t *testing.T) {
		split := frame("hl", "1 0 0 1 -588 -500", -157.5, 1333.5, -9, 59.4)
		split.OtherElements = append(split.OtherElements, common.RawXMLElement{
			XMLName: xml.Name{Local: "TextFramePreference"},
			Attrs: []xml.Attr{
				{Name: xml.Name{Local: "TextColumnCount"}, Value: "2"},
				{Name: xml.Name{Local: "TextColumnGutter"}, Value: "96.4"},
			},
		})
		c, err := FrameConstraints(split, styles, "ParagraphStyle/headline", 0)
		if err != nil {
			t.Fatal(err)
		}
		if !c.SpansFold {
			t.Error("SpansFold = false; the frame still crosses the spine")
		}
		if !c.FoldProtected {
			t.Error("FoldProtected = false although the frame breaks at a gutter")
		}
	})

	t.Run("a frame with no path geometry is an error", func(t *testing.T) {
		if _, err := FrameConstraints(&spread.TextFrame{}, styles, "", 0); err == nil {
			t.Error("accepted a frame with no geometry")
		}
	})
}

// --- helpers ---

type fixedWidth float64

func (f fixedWidth) Width(s string, _ float64) float64 {
	return float64(len([]rune(s))) * float64(f)
}

func props(elems ...common.RawXMLElement) *common.Properties {
	return &common.Properties{OtherElements: elems}
}

func raw(name, text string) common.RawXMLElement {
	return common.RawXMLElement{XMLName: xml.Name{Local: name}, Content: []byte(text)}
}

func leading(pt float64) common.RawXMLElement {
	return raw("Leading", strconv.FormatFloat(pt, 'f', -1, 64))
}

func basedOn(ref string) common.RawXMLElement { return raw("BasedOn", ref) }

func appliedFont(family string) common.RawXMLElement { return raw("AppliedFont", family) }

// frame builds a text frame with a rectangular path in local
// coordinates, where the origin is wherever the designer left it.
func frame(self, transform string, left, right, top, bottom float64) *spread.TextFrame {
	pt := func(x, y float64) common.PathPointType {
		a := strconv.FormatFloat(x, 'f', -1, 64) + " " + strconv.FormatFloat(y, 'f', -1, 64)
		return common.PathPointType{Anchor: a}
	}
	return &spread.TextFrame{
		PageItemBase: spread.PageItemBase{Self: self, ItemTransform: transform},
		Properties: &common.Properties{PathGeometry: &common.PathGeometry{
			GeometryPathType: &common.GeometryPathType{
				PathPointArray: &common.PathPointArray{PathPoints: []common.PathPointType{
					pt(left, top), pt(left, bottom), pt(right, bottom), pt(right, top),
				}},
			},
		}},
	}
}
