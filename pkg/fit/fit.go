// Package fit answers whether a piece of text will physically fit a text
// frame, and when it will not, which word is in the way.
//
// Everything here is derived from the document: the frame's own path
// geometry, the applied paragraph style's size and leading followed
// through BasedOn, the frame's column settings, and where it sits on a
// spread. What the package deliberately does not do is read font files.
// An IDML does not contain its fonts, only their names, so measuring text
// needs something this library has no business owning - hence Measurer,
// which the caller supplies.
//
// That split is not arbitrary. The subtle part is here: style inheritance,
// leading against type size, the difference between a frame that spans a
// spread's spine and one that breaks at a gutter. Glyph advance widths are
// not subtle, and keeping them out means idmllib stays free of any
// runtime dependency.
package fit

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/dimelords/idmllib/v3/pkg/common"
	"github.com/dimelords/idmllib/v3/pkg/resources"
	"github.com/dimelords/idmllib/v3/pkg/spread"
	"github.com/dimelords/idmllib/v3/pkg/story"
)

// Measurer reports the advance width of text, in points, when set at
// sizePt. Implementations read the actual font; see the package comment
// for why that lives outside idmllib.
//
// A nil Measurer is allowed and falls back to half the type size per
// character. That estimate is rough enough that a near-boundary verdict
// from it should not be acted on, and Constraints.Estimated says so.
type Measurer interface {
	Width(s string, sizePt float64) float64
}

// Ascender is a Measurer that also knows how far its font rises above the
// baseline, in points at sizePt. It is optional: a Measurer that does not
// implement it still measures widths.
//
// It matters for line count. A frame's first baseline normally sits at the
// font's ascent below the top, not a full leading, so a display headline
// often has room for one more line than height divided by leading says.
// Without an Ascender every line is charged a full leading, which can
// only undercount: a yes is still a yes.
//
// InDesign's ascent is the OS/2 table's typographic ascender
// (sTypoAscender), not hhea's ascender, which is often far larger.
type Ascender interface {
	Ascent(sizePt float64) float64
}

// Constraints describe what a frame will accept.
type Constraints struct {
	// FrameWidth and FrameHeight are the frame's size in points.
	FrameWidth  float64 `json:"frameWidth"`
	FrameHeight float64 `json:"frameHeight"`

	// PointSize and Leading come from the applied style, followed through
	// BasedOn. A style with no leading of its own is on InDesign's
	// automatic leading, 120% of the type size.
	PointSize float64 `json:"pointSize"`
	Leading   float64 `json:"leading"`

	// FontFamily and FontStyle name the face the text is set in, so a
	// caller can find the file to measure with.
	FontFamily string `json:"fontFamily,omitempty"`
	FontStyle  string `json:"fontStyle,omitempty"`

	// MaxLines is how many lines fit in the frame at this leading. Call
	// Measured to work it out with the font's ascent; Evaluate and Budget
	// do so themselves.
	MaxLines int `json:"maxLines"`

	// WordSpacing is the style's desired word spacing as a fraction of the
	// font's own space: 0.9 for DesiredWordSpacing="90". Zero means the
	// default, 100%. House headline styles commonly tighten it, and a
	// width measured with full spaces runs a line a few percent wide.
	WordSpacing float64 `json:"wordSpacing,omitempty"`
	// LetterSpacing is the style's desired letter spacing: the space added
	// after every character, as a fraction of the font's space. -0.05 for
	// DesiredLetterSpacing="-5"; zero for none. Together with WordSpacing
	// it is what a headline style uses to set its lines tight, ragged or
	// not.
	LetterSpacing float64 `json:"letterSpacing,omitempty"`

	// FirstBaseline is the frame's FirstBaselineOffset: "AscentOffset"
	// when the frame says nothing, which is InDesign's default.
	FirstBaseline string `json:"firstBaseline,omitempty"`
	// MinFirstBaseline is the frame's MinimumFirstBaselineOffset in
	// points. It is the whole offset for "FixedLineHeight", and a floor
	// under the others.
	MinFirstBaseline float64 `json:"minFirstBaseline,omitempty"`

	// widthScale scales every width measured, for Evaluate's close-call
	// check. Zero means unscaled.
	widthScale float64

	// FoldX is where a spread's spine crosses the frame, measured from
	// its left edge, and is zero when it does not. See FoldProtected: a
	// crossing frame is only constrained when it cannot break there by
	// itself.
	FoldX float64 `json:"foldX,omitempty"`
	// SpansFold reports whether the frame crosses the spine at all.
	SpansFold bool `json:"spansFold"`
	// FoldProtected reports that the frame crosses the spine but is split
	// into columns with a gutter, so the layout breaks at the fold on its
	// own and the wording is unconstrained by it.
	//
	// Detecting this matters. Without it every headline on a template
	// that had already solved the problem would be warned about it.
	FoldProtected bool `json:"foldProtected"`
}

// CloseMargin is how far, as a fraction of a line's width, this package's
// measurement can be from what a composer sets. A verdict that flips when
// every width is this much wider or narrower is reported as Close.
//
// The figure is measured, not chosen: on a real newspaper headline,
// kerning the package does not apply moved a line 0.5%, and the letter
// spacing and glyph scaling a composer may squeeze a ragged line by, which
// it does not model either, another 2.2%. Together they turned "needs five
// lines" into the four InDesign sets.
const CloseMargin = 0.03

// Failure says why a candidate does not fit.
type Failure string

const (
	// FailTooManyLines means the text needs more lines than the frame has.
	FailTooManyLines Failure = "too_many_lines"
	// FailWordTooWide means one word is wider than the frame. No line
	// break helps; the word itself has to change.
	FailWordTooWide Failure = "word_too_wide"
	// FailWordOnFold means the spine of a spread runs through a word. The
	// text fits; the break is in the wrong place. It is reported as a
	// Warning on a result that fits, never as Why: InDesign sets such a
	// line without complaint, and real pages ship with a word across the
	// fold, so whether to reword is an editor's call, not a verdict.
	FailWordOnFold Failure = "word_on_fold"
)

// Warning is something about a candidate that fits but may still want a
// person's eye.
type Warning struct {
	Why Failure `json:"why"`
	// Word names the word concerned.
	Word string `json:"word,omitempty"`
	// Reason states the warning in a sentence.
	Reason string `json:"reason"`
}

// Result is the verdict on one candidate.
type Result struct {
	Fits bool    `json:"fits"`
	Why  Failure `json:"why,omitempty"`
	// Reason states the failure in a sentence, for whoever proposed the
	// text.
	Reason string `json:"reason,omitempty"`
	// Lines is the text as it would wrap.
	Lines []string `json:"lines,omitempty"`
	// Word names the word at fault, for FailWordTooWide. A word on the
	// fold is named in its Warning instead.
	Word string `json:"word,omitempty"`
	// OverflowPt is how far past the frame the offending word runs.
	OverflowPt float64 `json:"overflowPt,omitempty"`
	// Estimated reports that no Measurer was supplied, so the widths
	// behind this verdict are a crude average and a close call should not
	// be trusted.
	Estimated bool `json:"estimated,omitempty"`
	// Warnings are reported on text that fits: a word the spine of a
	// spread runs through.
	Warnings []Warning `json:"warnings,omitempty"`
	// Close reports that the verdict would flip if the text set CloseMargin
	// wider or narrower than measured. Neither answer is safe then: a
	// composer squeezing letter spacing, kerning, or justification can
	// decide it either way, so render the text and let the layout engine
	// decide.
	Close bool `json:"close,omitempty"`
}

// Evaluate reports whether text fits, changing nothing.
//
// It does not hyphenate, so text it rejects as too wide may still be set
// with a hyphen by InDesign or Scribus, and the two engines can disagree
// at the boundary in any case. Treat a pass as "worth rendering" and the
// render as the decision.
func Evaluate(text string, c Constraints, m Measurer) Result {
	c = c.Measured(m)
	res := evaluate(text, c, m)
	if c.PointSize <= 0 || c.FrameWidth <= 0 {
		return res
	}
	// Set at the far edge of the margin in the verdict's own direction: a
	// pass is close when slightly wider text fails, a failure when
	// slightly narrower text passes.
	edge := c
	edge.widthScale = 1 + CloseMargin
	if !res.Fits {
		edge.widthScale = 1 - CloseMargin
	}
	res.Close = evaluate(text, edge, m).Fits != res.Fits
	return res
}

func evaluate(text string, c Constraints, m Measurer) Result {
	if c.PointSize <= 0 || c.FrameWidth <= 0 {
		return Result{Fits: true, Estimated: m == nil}
	}

	lines, tooWide := wrap(text, c, m)
	res := Result{Lines: lines, Estimated: m == nil}

	if tooWide != "" {
		res.Why, res.Word = FailWordTooWide, tooWide
		res.OverflowPt = measure(tooWide, c, m) - c.FrameWidth
		res.Reason = fmt.Sprintf("the word %q overflows the frame by %.0fpt at %.0fpt type; "+
			"no line break can fix it", tooWide, res.OverflowPt, c.PointSize)
		return res
	}
	if c.MaxLines > 0 && len(lines) > c.MaxLines {
		res.Why = FailTooManyLines
		res.Reason = fmt.Sprintf("needs %d lines but only %d fit at %.0fpt type",
			len(lines), c.MaxLines, c.PointSize)
		return res
	}
	res.Fits = true
	if c.FoldX > 0 && !c.FoldProtected {
		if word, gap := onFold(lines, c, m); word != "" {
			res.Warnings = append(res.Warnings, Warning{
				Why:  FailWordOnFold,
				Word: word,
				Reason: fmt.Sprintf("the spine of the spread runs through the word %q; the nearest "+
					"word break is %.0fpt away, so rewording would be needed to put a space on the fold",
					word, math.Abs(gap)),
			})
		}
	}
	return res
}

// Budget estimates how many characters of representative text fit. It is
// a hint for whoever writes the text and deliberately not a test: real
// character widths vary several-fold, so a length safe for one string is
// wrong for the next. Evaluate decides.
//
// sample should look like the language being written.
func Budget(sample string, c Constraints, m Measurer) int {
	c = c.Measured(m)
	if sample == "" || c.FrameWidth <= 0 {
		return 0
	}
	avg := measure(sample, c, m) / float64(len([]rune(sample)))
	if avg <= 0 {
		return 0
	}
	lines := c.MaxLines
	if lines < 1 {
		lines = 1
	}
	return int(c.FrameWidth * float64(lines) / avg)
}

// wrap greedily breaks text into lines, reporting the first word that
// cannot fit on a line of its own. Such a word is still placed, since
// there is nowhere narrower to put it.
func wrap(text string, c Constraints, m Measurer) (lines []string, tooWide string) {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil, ""
	}
	var cur string
	for _, w := range words {
		if cur == "" {
			if tooWide == "" && measure(w, c, m) > c.FrameWidth {
				tooWide = w
			}
			cur = w
			continue
		}
		if measure(cur+" "+w, c, m) <= c.FrameWidth {
			cur += " " + w
			continue
		}
		lines = append(lines, cur)
		if tooWide == "" && measure(w, c, m) > c.FrameWidth {
			tooWide = w
		}
		cur = w
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	return lines, tooWide
}

// onFold reports the first word the spine runs through, across every
// line, and how far the nearest word gap sits from it. The fold is
// vertical, so every line is checked, and a line ending before it is no
// violation.
func onFold(lines []string, c Constraints, m Measurer) (word string, gap float64) {
	space := measure(" ", c, m)
	nearest := math.Inf(1)

	for _, line := range lines {
		x := 0.0
		words := strings.Fields(line)
		for i, w := range words {
			end := x + measure(w, c, m)
			if x < c.FoldX && c.FoldX < end {
				for _, g := range []float64{x, end + space/2} {
					if d := g - c.FoldX; math.Abs(d) < math.Abs(nearest) {
						nearest = d
					}
				}
				if word == "" {
					word = w
				}
			}
			x = end
			if i < len(words)-1 {
				x += space
			}
		}
	}
	if math.IsInf(nearest, 1) {
		nearest = 0
	}
	return word, nearest
}

// measure is the width of s as the frame would set it: the font's widths,
// with every space at the style's word spacing and every character followed
// by its letter spacing.
func measure(s string, c Constraints, m Measurer) float64 {
	w := rawWidth(s, c, m)
	if c.WordSpacing > 0 && c.WordSpacing != 1 {
		if n := strings.Count(s, " "); n > 0 {
			w -= float64(n) * (1 - c.WordSpacing) * rawWidth(" ", c, m)
		}
	}
	if c.LetterSpacing != 0 && s != "" {
		w += float64(len([]rune(s))) * c.LetterSpacing * rawWidth(" ", c, m)
	}
	if c.widthScale > 0 {
		w *= c.widthScale
	}
	return w
}

func rawWidth(s string, c Constraints, m Measurer) float64 {
	if m != nil {
		return m.Width(s, c.PointSize)
	}
	return float64(len([]rune(s))) * c.PointSize * 0.5
}

// FrameConstraints works out what a frame will accept, from its geometry,
// its applied paragraph style and its position on the spread.
//
// spineX is where the spine sits in the spread's own coordinates. For a
// two-page spread InDesign places the left page at negative x, so the
// spine is at 0; pass math.Inf(1) for a single-page spread, which has
// none.
func FrameConstraints(tf *spread.TextFrame, styles *resources.StylesFile, appliedStyle string, spineX float64) (Constraints, error) {
	size := StyleSize(styles, appliedStyle)
	family, style := StyleFont(styles, appliedStyle)
	c, err := frameConstraints(tf, size, StyleLeading(styles, appliedStyle, size), family, style, spineX)
	if err != nil {
		return Constraints{}, err
	}
	c.WordSpacing = StyleWordSpacing(styles, appliedStyle)
	c.LetterSpacing = StyleLetterSpacing(styles, appliedStyle)
	return c, nil
}

// StoryFrameConstraints is FrameConstraints for the story a frame holds,
// and the one to use when the story is at hand.
//
// A paragraph style is only where formatting starts. Headlines in
// particular are routinely sized by hand: the text carries PointSize and
// Leading as local formatting, on its CharacterStyleRange or its
// ParagraphStyleRange, and the style underneath says something else
// entirely. A 68pt headline on 64.55pt leading, set in a style that says
// 69pt on automatic leading, gets one line fewer from the style than the
// frame really holds, and a headline InDesign sets comfortably is
// rejected.
//
// Local formatting is read from the first character range that holds
// text, then its paragraph range, then the applied style through BasedOn.
// Word and letter spacing are paragraph attributes and are read from the
// paragraph range and the style.
// Character styles are not followed; size set through one is still read
// from the paragraph style.
func StoryFrameConstraints(tf *spread.TextFrame, styles *resources.StylesFile, st *story.Story, spineX float64) (Constraints, error) {
	var paragraphs []*story.ParagraphStyleRange
	if st != nil {
		// Paragraphs, not ParagraphStyleRanges: a story tagged with XML
		// structure keeps its paragraphs inside XMLElements.
		paragraphs = st.Paragraphs()
	}
	if len(paragraphs) == 0 {
		return FrameConstraints(tf, styles, "", spineX)
	}
	psr := paragraphs[0]
	applied := psr.AppliedParagraphStyle

	local := paragraphFormatting(psr)
	if csr := firstTextRange(psr); csr != nil {
		local = local.over(characterFormatting(csr))
	}

	size := StyleSize(styles, applied)
	if local.size > 0 {
		size = local.size
	}
	family, style := StyleFont(styles, applied)
	if local.family != "" {
		family = local.family
	}
	if local.style != "" {
		style = local.style
	}

	var lead float64
	switch {
	case local.autoLeading:
		lead = size * 1.2
	case local.leading > 0:
		lead = local.leading
	default:
		// The style's own leading, or automatic leading at the size the
		// text is really set in, not the style's.
		lead = StyleLeading(styles, applied, size)
	}
	c, err := frameConstraints(tf, size, lead, family, style, spineX)
	if err != nil {
		return Constraints{}, err
	}
	c.WordSpacing = StyleWordSpacing(styles, applied)
	if local.wordSpacing > 0 {
		c.WordSpacing = local.wordSpacing
	}
	c.LetterSpacing = StyleLetterSpacing(styles, applied)
	if local.letterSpacingSet {
		c.LetterSpacing = local.letterSpacing
	}
	return c, nil
}

func frameConstraints(tf *spread.TextFrame, size, lead float64, family, style string, spineX float64) (Constraints, error) {
	minX, maxX, minY, maxY, ok := frameBounds(tf)
	if !ok {
		return Constraints{}, fmt.Errorf("fit: text frame %s has no usable path geometry", tf.Self)
	}
	c := Constraints{
		FrameWidth:  maxX - minX,
		FrameHeight: maxY - minY,
		PointSize:   size,
		Leading:     lead,
		FontFamily:  family,
		FontStyle:   style,
	}
	c.FirstBaseline = textFramePreference(tf, "FirstBaselineOffset")
	if c.FirstBaseline == "" {
		c.FirstBaseline = "AscentOffset"
	}
	if v, err := strconv.ParseFloat(textFramePreference(tf, "MinimumFirstBaselineOffset"), 64); err == nil && v > 0 {
		c.MinFirstBaseline = v
	}
	c.MaxLines = c.maxLines(c.Leading)
	if tx, _, err := transformOffset(tf.ItemTransform); err == nil && !math.IsInf(spineX, 1) {
		left, right := tx+minX, tx+maxX
		if left < spineX && spineX < right {
			c.SpansFold = true
			c.FoldX = spineX - left
			c.FoldProtected = HasColumnGutter(tf)
		}
	}
	return c, nil
}

// localFormatting is what one story range sets for itself. Zero values
// mean "not set here".
type localFormatting struct {
	size        float64
	leading     float64
	autoLeading bool
	family      string
	style       string
	wordSpacing float64
	// letterSpacing is set when letterSpacingSet is: zero is a real value,
	// "no letter spacing", that overrides a style's.
	letterSpacing    float64
	letterSpacingSet bool
}

// over returns l with every value inner sets replacing l's own: inner is
// the more local of the two.
func (l localFormatting) over(inner localFormatting) localFormatting {
	if inner.size > 0 {
		l.size = inner.size
	}
	if inner.leading > 0 || inner.autoLeading {
		l.leading, l.autoLeading = inner.leading, inner.autoLeading
	}
	if inner.family != "" {
		l.family = inner.family
	}
	if inner.style != "" {
		l.style = inner.style
	}
	if inner.wordSpacing > 0 {
		l.wordSpacing = inner.wordSpacing
	}
	if inner.letterSpacingSet {
		l.letterSpacing, l.letterSpacingSet = inner.letterSpacing, true
	}
	return l
}

func paragraphFormatting(psr *story.ParagraphStyleRange) localFormatting {
	var props *common.RawXMLElement
	for i := range psr.OtherElements {
		if psr.OtherElements[i].XMLName.Local == "Properties" {
			props = &psr.OtherElements[i]
		}
	}
	return readFormatting(psr.OtherAttrs, props)
}

func characterFormatting(csr *story.CharacterStyleRange) localFormatting {
	var props *common.RawXMLElement
	for _, child := range csr.Children {
		if child.Other != nil && child.Other.XMLName.Local == "Properties" {
			props = child.Other
		}
	}
	return readFormatting(csr.OtherAttrs, props)
}

func readFormatting(attrs []xml.Attr, props *common.RawXMLElement) localFormatting {
	var l localFormatting
	for _, a := range attrs {
		switch a.Name.Local {
		case "PointSize":
			if v, err := strconv.ParseFloat(a.Value, 64); err == nil && v > 0 {
				l.size = v
			}
		case "FontStyle":
			l.style = a.Value
		case "DesiredWordSpacing":
			l.wordSpacing = percent(a.Value)
		case "DesiredLetterSpacing":
			if v, err := strconv.ParseFloat(a.Value, 64); err == nil {
				l.letterSpacing, l.letterSpacingSet = v/100, true
			}
		}
	}
	if props == nil {
		return l
	}
	switch v := rawChildText(props, "Leading"); {
	case v == "Auto":
		l.autoLeading = true
	case v != "":
		if pt, err := strconv.ParseFloat(v, 64); err == nil && pt > 0 {
			l.leading = pt
		}
	}
	l.family = rawChildText(props, "AppliedFont")
	return l
}

// firstTextRange is the first character range of a paragraph that holds
// text: the formatting a reader sees. A paragraph can open with an empty
// range left behind by editing.
func firstTextRange(psr *story.ParagraphStyleRange) *story.CharacterStyleRange {
	for i := range psr.CharacterStyleRanges {
		for _, child := range psr.CharacterStyleRanges[i].Children {
			if child.Content != nil && strings.TrimSpace(child.Content.Text) != "" {
				return &psr.CharacterStyleRanges[i]
			}
		}
	}
	return nil
}

// rawChildText is the text of a direct child of an element held as raw
// XML, such as the Properties of a story range.
func rawChildText(el *common.RawXMLElement, name string) string {
	d := xml.NewDecoder(bytes.NewReader(el.Content))
	depth := 0
	for {
		tok, err := d.Token()
		if err != nil {
			return ""
		}
		switch t := tok.(type) {
		case xml.StartElement:
			depth++
			if depth == 1 && t.Name.Local == name {
				var text string
				if err := d.DecodeElement(&text, &t); err != nil {
					return ""
				}
				return strings.TrimSpace(text)
			}
		case xml.EndElement:
			depth--
		}
	}
}

// Measured returns c with MaxLines worked out for the font m measures
// with. When m is an Ascender and the frame's first baseline sits at the
// ascent, the first line is charged the ascent rather than a full
// leading; otherwise c comes back unchanged. It is idempotent, so calling
// it on constraints that have already been measured does no harm.
func (c Constraints) Measured(m Measurer) Constraints {
	a, ok := m.(Ascender)
	if !ok || c.PointSize <= 0 || c.FirstBaseline != "AscentOffset" {
		return c
	}
	if ascent := a.Ascent(c.PointSize); ascent > 0 {
		c.MaxLines = c.maxLines(ascent)
	}
	return c
}

// maxLines is how many lines the frame holds when its first baseline sits
// first points below the top: that line, plus every further leading that
// still fits. The same rule InDesign follows, and the one naviga-scribus-
// rust's renderer uses.
//
// FirstBaseline values this package does not model (cap height, x-height)
// are charged a full leading, which can only undercount.
func (c Constraints) maxLines(first float64) int {
	if c.Leading <= 0 {
		return 1
	}
	if c.FirstBaseline == "FixedLineHeight" && c.MinFirstBaseline > 0 {
		first = c.MinFirstBaseline
	}
	first = math.Max(first, c.MinFirstBaseline)
	// The epsilon keeps a frame drawn to an exact multiple of the leading
	// from losing its last line to floating-point noise in the geometry.
	n := int(math.Floor((c.FrameHeight-first+1e-6)/c.Leading)) + 1
	return max(n, 1)
}

// textFramePreference is an attribute of a frame's TextFramePreference,
// or "" when the frame does not set it.
func textFramePreference(tf *spread.TextFrame, name string) string {
	for i := range tf.OtherElements {
		e := &tf.OtherElements[i]
		if e.XMLName.Local != "TextFramePreference" {
			continue
		}
		for _, a := range e.Attrs {
			if a.Name.Local == name {
				return a.Value
			}
		}
	}
	return ""
}

// HasColumnGutter reports whether a frame is split into two or more
// columns with a gutter between them, which is how a template keeps a
// spread's spine clear of text.
func HasColumnGutter(tf *spread.TextFrame) bool {
	for i := range tf.OtherElements {
		e := &tf.OtherElements[i]
		if e.XMLName.Local != "TextFramePreference" {
			continue
		}
		var count, gutter float64
		for _, a := range e.Attrs {
			switch a.Name.Local {
			case "TextColumnCount":
				count, _ = strconv.ParseFloat(a.Value, 64)
			case "TextColumnGutter":
				gutter, _ = strconv.ParseFloat(a.Value, 64)
			}
		}
		if count >= 2 && gutter > 0 {
			return true
		}
	}
	return false
}

// frameBounds reads a frame's extent from its path geometry. Real
// InDesign frames carry their shape there rather than in
// GeometricBounds, and the local origin is wherever the designer left it,
// so both extremes have to be found rather than assumed.
func frameBounds(tf *spread.TextFrame) (minX, maxX, minY, maxY float64, ok bool) {
	if tf.Properties == nil || tf.Properties.PathGeometry == nil {
		return 0, 0, 0, 0, false
	}
	gp := tf.Properties.PathGeometry.GeometryPathType
	if gp == nil || gp.PathPointArray == nil {
		return 0, 0, 0, 0, false
	}

	first := true
	for _, p := range gp.PathPointArray.PathPoints {
		parts := strings.Fields(p.Anchor)
		if len(parts) != 2 {
			continue
		}
		x, err1 := strconv.ParseFloat(parts[0], 64)
		y, err2 := strconv.ParseFloat(parts[1], 64)
		if err1 != nil || err2 != nil {
			continue
		}
		if first {
			minX, maxX, minY, maxY, first = x, x, y, y, false
			continue
		}
		minX, maxX = math.Min(minX, x), math.Max(maxX, x)
		minY, maxY = math.Min(minY, y), math.Max(maxY, y)
	}
	return minX, maxX, minY, maxY, !first
}

// StyleSize is the type size a paragraph style sets, followed through
// BasedOn.
//
// Following the chain is the whole point: a house style routinely carries
// no size of its own and inherits it from the root, and a reader that
// stops at the style itself sees nothing.
func StyleSize(styles *resources.StylesFile, self string) float64 {
	var out float64
	walkStyles(styles, self, func(ps *resources.ParagraphStyle) bool {
		if v, err := strconv.ParseFloat(ps.PointSize, 64); err == nil && v > 0 {
			out = v
			return true
		}
		return false
	})
	return out
}

// StyleLeading is the leading a paragraph style sets, followed through
// BasedOn. A style with no leading of its own is on InDesign's automatic
// leading, 120% of the type size.
func StyleLeading(styles *resources.StylesFile, self string, size float64) float64 {
	var out float64
	walkStyles(styles, self, func(ps *resources.ParagraphStyle) bool {
		if v, err := strconv.ParseFloat(propertyText(ps.Properties, "Leading"), 64); err == nil && v > 0 {
			out = v
			return true
		}
		return false
	})
	if out == 0 && size > 0 {
		out = size * 1.2
	}
	return out
}

// StyleFont is the family and style a paragraph style sets, followed
// through BasedOn.
func StyleFont(styles *resources.StylesFile, self string) (family, style string) {
	walkStyles(styles, self, func(ps *resources.ParagraphStyle) bool {
		if f := ps.Properties.GetAppliedFont(); f != "" {
			family, style = f, ps.FontStyle
			return true
		}
		return false
	})
	return family, style
}

// StyleWordSpacing is the desired word spacing a paragraph style sets, as
// a fraction, followed through BasedOn. Zero means none in the chain, which
// is InDesign's default of 100%.
//
// It is the desired value, not the minimum, for every alignment. Ragged
// text is set at the desired spacing; justified text can tighten below it,
// but by how much is the composer's choice, so assuming the minimum would
// accept lines InDesign does not set.
func StyleWordSpacing(styles *resources.StylesFile, self string) float64 {
	var out float64
	walkStyles(styles, self, func(ps *resources.ParagraphStyle) bool {
		for _, a := range ps.OtherAttrs {
			if a.Name.Local == "DesiredWordSpacing" {
				out = percent(a.Value)
				return out > 0
			}
		}
		return false
	})
	return out
}

// StyleLetterSpacing is the desired letter spacing a paragraph style sets,
// as a fraction of the font's space, followed through BasedOn. Unlike word
// spacing, zero is a value a style sets - the root style usually does - so
// the first style in the chain that sets it at all wins.
//
// It is the desired value for the same reason as StyleWordSpacing: ragged
// text is set at it, and how far a composer goes past it is its own
// choice.
func StyleLetterSpacing(styles *resources.StylesFile, self string) float64 {
	var out float64
	walkStyles(styles, self, func(ps *resources.ParagraphStyle) bool {
		for _, a := range ps.OtherAttrs {
			if a.Name.Local == "DesiredLetterSpacing" {
				v, err := strconv.ParseFloat(a.Value, 64)
				if err != nil {
					return false
				}
				out = v / 100
				return true
			}
		}
		return false
	})
	return out
}

// percent reads an IDML percentage, "90" for 90%, as a fraction. Zero for
// anything unreadable or not positive.
func percent(s string) float64 {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil || v <= 0 {
		return 0
	}
	return v / 100
}

// walkStyles follows a BasedOn chain, calling fn until it returns true.
// The walk is bounded: a cycle in a damaged document must not hang the
// caller.
func walkStyles(styles *resources.StylesFile, self string, fn func(*resources.ParagraphStyle) bool) {
	if styles == nil {
		return
	}
	seen := map[string]bool{}
	for range 16 {
		if self == "" || seen[self] {
			return
		}
		seen[self] = true

		ps := FindParagraphStyle(styles.RootParagraphStyleGroup, self)
		if ps == nil || fn(ps) {
			return
		}
		// BasedOn is written without the element prefix that Self
		// carries, so following it by Self needs the prefix restored.
		next := propertyText(ps.Properties, "BasedOn")
		if next != "" && !strings.HasPrefix(next, "ParagraphStyle/") {
			next = "ParagraphStyle/" + next
		}
		self = next
	}
}

// FindParagraphStyle looks a style up by Self, at any depth in the group
// tree.
func FindParagraphStyle(g *resources.ParagraphStyleGroup, self string) *resources.ParagraphStyle {
	if g == nil {
		return nil
	}
	for i := range g.ParagraphStyles {
		if g.ParagraphStyles[i].Self == self {
			return &g.ParagraphStyles[i]
		}
	}
	for i := range g.NestedGroups {
		if found := FindParagraphStyle(&g.NestedGroups[i], self); found != nil {
			return found
		}
	}
	return nil
}

func propertyText(props *common.Properties, name string) string {
	if props == nil {
		return ""
	}
	for i := range props.OtherElements {
		if props.OtherElements[i].XMLName.Local == name {
			return strings.TrimSpace(string(props.OtherElements[i].Content))
		}
	}
	return ""
}

func transformOffset(s string) (x, y float64, err error) {
	parts := strings.Fields(s)
	if len(parts) != 6 {
		return 0, 0, fmt.Errorf("fit: transform %q has %d values, want 6", s, len(parts))
	}
	if x, err = strconv.ParseFloat(parts[4], 64); err != nil {
		return 0, 0, err
	}
	y, err = strconv.ParseFloat(parts[5], 64)
	return x, y, err
}
