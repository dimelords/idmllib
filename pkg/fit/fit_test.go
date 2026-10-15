package fit

import (
	"encoding/xml"
	"math"
	"strconv"
	"testing"

	"github.com/dimelords/idmllib/v3/pkg/common"
	"github.com/dimelords/idmllib/v3/pkg/resources"
	"github.com/dimelords/idmllib/v3/pkg/spread"
	"github.com/dimelords/idmllib/v3/pkg/story"
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

	t.Run("a word on the spine fits, with a warning", func(t *testing.T) {
		got := Evaluate("aaaaaaaaaaaaaaaaaaaaaaaaaa bb", c, nil)
		if !got.Fits || got.Why != "" {
			t.Fatalf("fits=%v why=%q: a word on the fold is a warning, not a failure", got.Fits, got.Why)
		}
		if len(got.Warnings) != 1 || got.Warnings[0].Why != FailWordOnFold {
			t.Fatalf("Warnings = %+v, want one %q", got.Warnings, FailWordOnFold)
		}
		if got.Warnings[0].Word == "" || got.Warnings[0].Reason == "" {
			t.Error("no Word or Reason; the wording cannot be corrected blind")
		}
	})

	t.Run("a failure carries no fold warning", func(t *testing.T) {
		twoLines := c
		twoLines.FrameWidth = 1000
		twoLines.FoldX = 500
		if got := Evaluate("aaaaaaaaaaaaaaaaaaaaaaaaaa bbbbbbbbbbbbbbbbbbbbbbbbbb", twoLines, nil); got.Fits || len(got.Warnings) != 0 {
			t.Errorf("fits=%v warnings=%+v, want a failure and no warnings", got.Fits, got.Warnings)
		}
	})

	t.Run("a break on the spine is accepted", func(t *testing.T) {
		// 20 characters is 720pt, then a space spanning 720..756, which
		// contains the fold at 745.5.
		if got := Evaluate("aaaaaaaaaaaaaaaaaaaa bbbbbbbbbbbbbbbbbbbb", c, nil); !got.Fits || len(got.Warnings) != 0 {
			t.Errorf("a break landing on the fold: fits=%v warnings=%+v", got.Fits, got.Warnings)
		}
	})

	t.Run("text ending before the spine is fine", func(t *testing.T) {
		if got := Evaluate("kort", c, nil); !got.Fits || len(got.Warnings) != 0 {
			t.Errorf("text that never reaches the fold: fits=%v warnings=%+v", got.Fits, got.Warnings)
		}
	})

	t.Run("a gutter at the spine lifts the constraint", func(t *testing.T) {
		// The same text, in a frame split into columns with a gutter. The
		// layout breaks at the fold by itself, so the wording is free -
		// without this, every headline on a correctly built template
		// would be rejected for a problem the template had solved.
		protected := c
		protected.FoldProtected = true
		if got := Evaluate("aaaaaaaaaaaaaaaaaaaaaaaaaa bb", protected, nil); !got.Fits || len(got.Warnings) != 0 {
			t.Errorf("text in a gutter-protected frame: fits=%v warnings=%+v", got.Fits, got.Warnings)
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
						Properties: props(leading(66), appliedFont("Myriad Pro")),
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
	if fam != "Myriad Pro" || style != "Light Semicondensed" {
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

// TestStoryFrameConstraints covers text sized by hand. The first case is
// the shape of a real newspaper headline: 68pt on 64.55pt leading set
// locally on the text, in a style that says 69pt on automatic leading.
// Read from the style alone the frame loses a line.
//
// The frame is 270pt rather than the real headline's 256pt: with no
// Ascender the first line is charged a full leading, where InDesign
// charges it the font's ascent, so at 256pt even the right leading gives
// three lines
// where InDesign sets four without an Ascender. TestMeasured covers the
// first line; this test is about the leading.
func TestStoryFrameConstraints(t *testing.T) {
	styles := &resources.StylesFile{
		RootParagraphStyleGroup: &resources.ParagraphStyleGroup{
			ParagraphStyles: []resources.ParagraphStyle{{
				Self: "ParagraphStyle/headline", PointSize: "69", FontStyle: "Black",
				Properties: props(appliedFont("Sole Serif B Headline")),
			}, {
				Self: "ParagraphStyle/fixed", PointSize: "20",
				Properties: props(leading(30)),
			}},
		},
	}
	tf := frame("hl", "1 0 0 1 0 0", 0, 427.5, 0, 270)

	tests := []struct {
		name        string
		paragraph   string // inner XML of the ParagraphStyleRange
		style       string
		size, lead  float64
		lines       int
		fontStyle   string
		fromStyleOK bool // FrameConstraints alone already agrees
	}{{
		name:  "local size and leading on the character range",
		style: "ParagraphStyle/headline",
		paragraph: `<CharacterStyleRange AppliedCharacterStyle="CharacterStyle/$ID/[No character style]" PointSize="68">
				<Properties><Leading type="unit">64.55</Leading></Properties>
				<Content>Den typiska snattaren är en vuxen som är ute efter mat</Content>
			</CharacterStyleRange>`,
		size: 68, lead: 64.55, lines: 4, // 270/64.55; the style alone gives 270/82.8 = 3
		fontStyle: "Black",
	}, {
		name:  "local size alone moves automatic leading with it",
		style: "ParagraphStyle/headline",
		paragraph: `<CharacterStyleRange AppliedCharacterStyle="CharacterStyle/$ID/[No character style]" PointSize="50">
				<Content>Rubrik</Content>
			</CharacterStyleRange>`,
		size: 50, lead: 60, lines: 4, fontStyle: "Black",
	}, {
		name:  "local automatic leading replaces the style's fixed leading",
		style: "ParagraphStyle/fixed",
		paragraph: `<CharacterStyleRange AppliedCharacterStyle="CharacterStyle/$ID/[No character style]">
				<Properties><Leading type="enumeration">Auto</Leading></Properties>
				<Content>Rubrik</Content>
			</CharacterStyleRange>`,
		size: 20, lead: 24, lines: 11,
	}, {
		name:  "the character range is more local than the paragraph range",
		style: "ParagraphStyle/headline",
		paragraph: `<Properties><Leading type="unit">80</Leading></Properties>
			<CharacterStyleRange AppliedCharacterStyle="CharacterStyle/$ID/[No character style]" PointSize="60" FontStyle="Bold">
				<Content>Rubrik</Content>
			</CharacterStyleRange>`,
		size: 60, lead: 80, lines: 3, fontStyle: "Bold",
	}, {
		name:  "an empty opening range is not what the reader sees",
		style: "ParagraphStyle/headline",
		paragraph: `<CharacterStyleRange AppliedCharacterStyle="CharacterStyle/$ID/[No character style]" PointSize="12">
				<Content></Content>
			</CharacterStyleRange>
			<CharacterStyleRange AppliedCharacterStyle="CharacterStyle/$ID/[No character style]" PointSize="40">
				<Properties><Leading type="unit">42</Leading></Properties>
				<Content>Rubrik</Content>
			</CharacterStyleRange>`,
		size: 40, lead: 42, lines: 6, fontStyle: "Black",
	}, {
		name:  "no local formatting is the style",
		style: "ParagraphStyle/fixed",
		paragraph: `<CharacterStyleRange AppliedCharacterStyle="CharacterStyle/$ID/[No character style]">
				<Content>Rubrik</Content>
			</CharacterStyleRange>`,
		size: 20, lead: 30, lines: 9, fromStyleOK: true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := parseStory(t, tt.style, tt.paragraph)
			c, err := StoryFrameConstraints(tf, styles, st, math.Inf(1))
			if err != nil {
				t.Fatal(err)
			}
			if c.PointSize != tt.size || math.Abs(c.Leading-tt.lead) > 0.001 || c.MaxLines != tt.lines {
				t.Errorf("size/leading/lines = %v/%v/%d, want %v/%v/%d",
					c.PointSize, c.Leading, c.MaxLines, tt.size, tt.lead, tt.lines)
			}
			if c.FontStyle != tt.fontStyle {
				t.Errorf("FontStyle = %q, want %q", c.FontStyle, tt.fontStyle)
			}

			fromStyle, err := FrameConstraints(tf, styles, tt.style, math.Inf(1))
			if err != nil {
				t.Fatal(err)
			}
			if agrees := fromStyle == c; agrees != tt.fromStyleOK {
				t.Errorf("FrameConstraints agrees = %v, want %v", agrees, tt.fromStyleOK)
			}
		})
	}

	t.Run("no story is the frame alone", func(t *testing.T) {
		c, err := StoryFrameConstraints(tf, styles, nil, math.Inf(1))
		if err != nil {
			t.Fatal(err)
		}
		if c.FrameWidth != 427.5 || c.PointSize != 0 {
			t.Errorf("got %+v", c)
		}
	})
}

func parseStory(t *testing.T, style, paragraph string) *story.Story {
	t.Helper()
	data := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<idPkg:Story xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging" DOMVersion="20.0">
	<Story Self="u1">
		<ParagraphStyleRange AppliedParagraphStyle="` + style + `">
			` + paragraph + `
		</ParagraphStyleRange>
	</Story>
</idPkg:Story>`
	f, err := story.ParseStory([]byte(data))
	if err != nil {
		t.Fatalf("ParseStory: %v", err)
	}
	return &f.Story
}

// TestMeasured covers the first line. The real headline: 68pt on 64.55pt
// leading in a 256pt frame. InDesign sets four lines, because the first
// baseline sits at the ascent, not a full leading, below the top.
func TestMeasured(t *testing.T) {
	headline := func(prefs ...xml.Attr) Constraints {
		tf := frame("hl", "1 0 0 1 0 0", 0, 427.5, 0, 256)
		if len(prefs) > 0 {
			tf.OtherElements = append(tf.OtherElements, common.RawXMLElement{
				XMLName: xml.Name{Local: "TextFramePreference"}, Attrs: prefs,
			})
		}
		c, err := frameConstraints(tf, 68, 64.55, "", "", math.Inf(1))
		if err != nil {
			t.Fatal(err)
		}
		return c
	}
	pref := func(name, value string) xml.Attr { return xml.Attr{Name: xml.Name{Local: name}, Value: value} }
	// 0.75em: 51pt at 68pt, a typical display face's typographic ascender.
	asc := withAscent{fixedWidth(25), 0.75}

	tests := []struct {
		name  string
		c     Constraints
		m     Measurer
		lines int
	}{
		{"the default first baseline is the ascent", headline(), asc, 4},
		{"without an Ascender every line is a full leading", headline(), fixedWidth(25), 3},
		{"without a Measurer every line is a full leading", headline(), nil, 3},
		{"a leading offset is a full leading", headline(pref("FirstBaselineOffset", "LeadingOffset")), asc, 3},
		{"an explicit ascent offset", headline(pref("FirstBaselineOffset", "AscentOffset")), asc, 4},
		{"a fixed offset is the minimum", headline(pref("FirstBaselineOffset", "FixedLineHeight"), pref("MinimumFirstBaselineOffset", "20")), nil, 4},
		{"the minimum is a floor under the ascent", headline(pref("MinimumFirstBaselineOffset", "70")), asc, 3},
		{"an unmodelled offset is a full leading", headline(pref("FirstBaselineOffset", "CapHeightOffset")), asc, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.c.Measured(tt.m)
			if got.MaxLines != tt.lines {
				t.Errorf("MaxLines = %d, want %d", got.MaxLines, tt.lines)
			}
			if again := got.Measured(tt.m); again != got {
				t.Errorf("Measured is not idempotent: %+v then %+v", got, again)
			}
		})
	}

	t.Run("Evaluate measures for itself", func(t *testing.T) {
		// 427.5pt at 25pt a character is 17 characters a line: four lines.
		text := "Den typiska snattaren är en vuxen som är ute efter mat"
		if r := Evaluate(text, headline(), asc); !r.Fits || len(r.Lines) != 4 {
			t.Errorf("with the ascent: fits=%v in %d lines (%s), want four that fit", r.Fits, len(r.Lines), r.Reason)
		}
		if r := Evaluate(text, headline(), fixedWidth(25)); r.Fits || r.Why != FailTooManyLines {
			t.Errorf("without it: fits=%v why=%q, want too_many_lines", r.Fits, r.Why)
		}
	})
}

// withAscent is a Measurer that is also an Ascender, rising em times the
// type size above the baseline.
type withAscent struct {
	fixedWidth
	em float64
}

func (w withAscent) Ascent(sizePt float64) float64 { return w.em * sizePt }

// TestWordSpacing covers styles that tighten their spaces. The real
// headline's style sets DesiredWordSpacing="90", and measured with full
// spaces its second line runs a few percent past a frame InDesign sets it
// in.
func TestWordSpacing(t *testing.T) {
	ws := func(v string) []xml.Attr {
		return []xml.Attr{{Name: xml.Name{Local: "DesiredWordSpacing"}, Value: v}}
	}
	styles := &resources.StylesFile{
		RootParagraphStyleGroup: &resources.ParagraphStyleGroup{
			ParagraphStyles: []resources.ParagraphStyle{{
				Self: "ParagraphStyle/$ID/[No paragraph style]", PointSize: "12", OtherAttrs: ws("100"),
			}, {
				Self: "ParagraphStyle/headline", PointSize: "20",
				Properties: props(basedOn("$ID/[No paragraph style]")), OtherAttrs: ws("90"),
			}, {
				// Its own spacing comes from the root.
				Self:       "ParagraphStyle/inherits",
				Properties: props(basedOn("$ID/[No paragraph style]")),
			}, {
				Self:       "ParagraphStyle/unreadable",
				OtherAttrs: ws("tight"),
			}},
		},
	}

	t.Run("style resolution", func(t *testing.T) {
		for style, want := range map[string]float64{
			"ParagraphStyle/headline":   0.9,
			"ParagraphStyle/inherits":   1,
			"ParagraphStyle/unreadable": 0,
			"ParagraphStyle/missing":    0,
		} {
			if got := StyleWordSpacing(styles, style); got != want {
				t.Errorf("%s: %v, want %v", style, got, want)
			}
		}
	})

	t.Run("constraints carry it", func(t *testing.T) {
		tf := frame("hl", "1 0 0 1 0 0", 0, 300, 0, 100)
		c, err := FrameConstraints(tf, styles, "ParagraphStyle/headline", math.Inf(1))
		if err != nil {
			t.Fatal(err)
		}
		if c.WordSpacing != 0.9 {
			t.Errorf("FrameConstraints WordSpacing = %v, want 0.9", c.WordSpacing)
		}
		st := parseStory(t, "ParagraphStyle/headline", `<CharacterStyleRange AppliedCharacterStyle="CharacterStyle/$ID/[No character style]">
				<Content>Rubrik</Content>
			</CharacterStyleRange>`)
		st.ParagraphStyleRanges[0].OtherAttrs = append(st.ParagraphStyleRanges[0].OtherAttrs, ws("80")...)
		c, err = StoryFrameConstraints(tf, styles, st, math.Inf(1))
		if err != nil {
			t.Fatal(err)
		}
		if c.WordSpacing != 0.8 {
			t.Errorf("local DesiredWordSpacing: WordSpacing = %v, want 0.8", c.WordSpacing)
		}
	})

	t.Run("spaces are measured at it", func(t *testing.T) {
		c := Constraints{PointSize: 12, WordSpacing: 0.5}
		// Five characters at 10pt, the space at half: 45pt.
		if got := measure("ab cd", c, fixedWidth(10)); got != 45 {
			t.Errorf("measure = %v, want 45", got)
		}
		if got := measure("abcd", c, fixedWidth(10)); got != 40 {
			t.Errorf("a word without spaces = %v, want 40", got)
		}
	})

	t.Run("a line that fits only at the style's spacing", func(t *testing.T) {
		// 14 characters at 10pt with four spaces: 140pt with full spaces,
		// 136pt at 90%. The frame is 137pt.
		text := "aa bb cc dd ee"
		c := Constraints{FrameWidth: 137, FrameHeight: 20, PointSize: 12, Leading: 14.4, MaxLines: 1}
		if r := Evaluate(text, c, fixedWidth(10)); r.Fits {
			t.Errorf("full spaces: fits, want too_many_lines")
		}
		c.WordSpacing = 0.9
		if r := Evaluate(text, c, fixedWidth(10)); !r.Fits {
			t.Errorf("90%% spaces: %s, want it to fit", r.Reason)
		}
	})
}

// TestEvaluate_Close covers verdicts too near the edge to act on. The
// text is 140pt at 10pt a character; CloseMargin moves it 4.2pt either way.
func TestEvaluate_Close(t *testing.T) {
	text := "aa bb cc dd ee"
	tests := []struct {
		name        string
		width       float64
		fits, close bool
	}{
		{"a clear pass", 200, true, false},
		{"a pass that slightly wider text would fail", 142, true, true},
		{"a failure that slightly narrower text would pass", 137, false, true},
		{"a clear failure", 100, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Constraints{FrameWidth: tt.width, FrameHeight: 20, PointSize: 12, Leading: 14.4, MaxLines: 1}
			r := Evaluate(text, c, fixedWidth(10))
			if r.Fits != tt.fits || r.Close != tt.close {
				t.Errorf("fits=%v close=%v, want fits=%v close=%v (%s)", r.Fits, r.Close, tt.fits, tt.close, r.Reason)
			}
			// The check must not leak into the verdict's own lines.
			if len(r.Lines) != 1 && tt.fits {
				t.Errorf("lines = %q", r.Lines)
			}
		})
	}

	t.Run("budget is unaffected", func(t *testing.T) {
		c := Constraints{FrameWidth: 137, FrameHeight: 20, PointSize: 12, Leading: 14.4, MaxLines: 1}
		if got := Budget("abcdefghij", c, fixedWidth(10)); got != 13 {
			t.Errorf("Budget = %d, want 13", got)
		}
	})
}

// TestLetterSpacing covers styles that tighten every letter. The real
// headline's style sets DesiredLetterSpacing="-5" on top of 90% word
// spacing; with both, its tightest line comes within kerning of the frame
// InDesign sets it in, and without the letter spacing it is 2% over.
func TestLetterSpacing(t *testing.T) {
	ls := func(v string) []xml.Attr {
		return []xml.Attr{{Name: xml.Name{Local: "DesiredLetterSpacing"}, Value: v}}
	}
	styles := &resources.StylesFile{
		RootParagraphStyleGroup: &resources.ParagraphStyleGroup{
			ParagraphStyles: []resources.ParagraphStyle{{
				Self: "ParagraphStyle/$ID/[No paragraph style]", PointSize: "12", OtherAttrs: ls("0"),
			}, {
				Self: "ParagraphStyle/headline", PointSize: "20",
				Properties: props(basedOn("$ID/[No paragraph style]")), OtherAttrs: ls("-5"),
			}, {
				Self:       "ParagraphStyle/inherits",
				Properties: props(basedOn("ParagraphStyle/headline")),
			}, {
				// An explicit zero is a value: it stops the walk.
				Self:       "ParagraphStyle/explicit-zero",
				Properties: props(basedOn("ParagraphStyle/headline")), OtherAttrs: ls("0"),
			}},
		},
	}

	t.Run("style resolution", func(t *testing.T) {
		for style, want := range map[string]float64{
			"ParagraphStyle/headline":      -0.05,
			"ParagraphStyle/inherits":      -0.05,
			"ParagraphStyle/explicit-zero": 0,
			"ParagraphStyle/missing":       0,
		} {
			if got := StyleLetterSpacing(styles, style); got != want {
				t.Errorf("%s: %v, want %v", style, got, want)
			}
		}
	})

	t.Run("constraints carry it, and a local zero overrides the style", func(t *testing.T) {
		tf := frame("hl", "1 0 0 1 0 0", 0, 300, 0, 100)
		c, err := FrameConstraints(tf, styles, "ParagraphStyle/headline", math.Inf(1))
		if err != nil {
			t.Fatal(err)
		}
		if c.LetterSpacing != -0.05 {
			t.Errorf("FrameConstraints LetterSpacing = %v, want -0.05", c.LetterSpacing)
		}
		st := parseStory(t, "ParagraphStyle/headline", `<CharacterStyleRange AppliedCharacterStyle="CharacterStyle/$ID/[No character style]">
				<Content>Rubrik</Content>
			</CharacterStyleRange>`)
		st.ParagraphStyleRanges[0].OtherAttrs = append(st.ParagraphStyleRanges[0].OtherAttrs, ls("0")...)
		c, err = StoryFrameConstraints(tf, styles, st, math.Inf(1))
		if err != nil {
			t.Fatal(err)
		}
		if c.LetterSpacing != 0 {
			t.Errorf("local DesiredLetterSpacing=0: LetterSpacing = %v, want 0", c.LetterSpacing)
		}
	})

	t.Run("every character is followed by it", func(t *testing.T) {
		// Five characters at 10pt, each tightened by 10% of the 10pt space.
		c := Constraints{PointSize: 12, LetterSpacing: -0.1}
		if got := measure("ab cd", c, fixedWidth(10)); got != 45 {
			t.Errorf("measure = %v, want 45", got)
		}
		// With 50% word spacing as well: the space shrinks by 5pt first.
		c.WordSpacing = 0.5
		if got := measure("ab cd", c, fixedWidth(10)); got != 40 {
			t.Errorf("with word spacing = %v, want 40", got)
		}
	})

	t.Run("a line that fits only with it", func(t *testing.T) {
		// 14 characters at 10pt: 140pt, 136pt at 90% words, 129pt with
		// -5% letters on top. The frame is 132pt.
		text := "aa bb cc dd ee"
		c := Constraints{FrameWidth: 132, FrameHeight: 20, PointSize: 12, Leading: 14.4, MaxLines: 1, WordSpacing: 0.9}
		if r := Evaluate(text, c, fixedWidth(10)); r.Fits {
			t.Error("word spacing alone: fits, want it not to")
		}
		c.LetterSpacing = -0.05
		if r := Evaluate(text, c, fixedWidth(10)); !r.Fits {
			t.Errorf("with letter spacing: %s, want it to fit", r.Reason)
		}
	})
}

// TestStoryFrameConstraints_Tagged covers a story tagged with XML
// structure, whose paragraphs sit inside an XMLElement rather than directly
// in the story.
func TestStoryFrameConstraints_Tagged(t *testing.T) {
	styles := &resources.StylesFile{
		RootParagraphStyleGroup: &resources.ParagraphStyleGroup{
			ParagraphStyles: []resources.ParagraphStyle{{Self: "ParagraphStyle/headline", PointSize: "69"}},
		},
	}
	data := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<idPkg:Story xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging" DOMVersion="20.0">
	<Story Self="u1">
		<XMLElement Self="di2" MarkupTag="XMLTag/headline">
			<ParagraphStyleRange AppliedParagraphStyle="ParagraphStyle/headline">
				<CharacterStyleRange AppliedCharacterStyle="CharacterStyle/$ID/[No character style]" PointSize="68">
					<Properties><Leading type="unit">64.55</Leading></Properties>
					<Content>Rubrik</Content>
				</CharacterStyleRange>
			</ParagraphStyleRange>
		</XMLElement>
	</Story>
</idPkg:Story>`
	f, err := story.ParseStory([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	c, err := StoryFrameConstraints(frame("hl", "1 0 0 1 0 0", 0, 427.5, 0, 270), styles, &f.Story, math.Inf(1))
	if err != nil {
		t.Fatal(err)
	}
	if c.PointSize != 68 || c.Leading != 64.55 {
		t.Errorf("size/leading = %v/%v, want 68/64.55 from inside the XMLElement", c.PointSize, c.Leading)
	}
}
