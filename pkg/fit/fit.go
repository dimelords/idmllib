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
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/dimelords/idmllib/v3/pkg/common"
	"github.com/dimelords/idmllib/v3/pkg/resources"
	"github.com/dimelords/idmllib/v3/pkg/spread"
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

	// MaxLines is how many lines fit in the frame at this leading.
	MaxLines int `json:"maxLines"`

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
	// that had already solved the problem would be rejected for it.
	FoldProtected bool `json:"foldProtected"`
}

// Failure says why a candidate does not fit.
type Failure string

const (
	// FailTooManyLines means the text needs more lines than the frame has.
	FailTooManyLines Failure = "too_many_lines"
	// FailWordTooWide means one word is wider than the frame. No line
	// break helps; the word itself has to change.
	FailWordTooWide Failure = "word_too_wide"
	// FailWordOnFold means the spine of a spread runs through a word. The
	// text fits; the break is in the wrong place.
	FailWordOnFold Failure = "word_on_fold"
)

// Result is the verdict on one candidate.
type Result struct {
	Fits bool    `json:"fits"`
	Why  Failure `json:"why,omitempty"`
	// Reason states the failure in a sentence, for whoever proposed the
	// text.
	Reason string `json:"reason,omitempty"`
	// Lines is the text as it would wrap.
	Lines []string `json:"lines,omitempty"`
	// Word names the word at fault, for FailWordTooWide and
	// FailWordOnFold.
	Word string `json:"word,omitempty"`
	// OverflowPt is how far past the frame the offending word runs.
	OverflowPt float64 `json:"overflowPt,omitempty"`
	// Estimated reports that no Measurer was supplied, so the widths
	// behind this verdict are a crude average and a close call should not
	// be trusted.
	Estimated bool `json:"estimated,omitempty"`
}

// Evaluate reports whether text fits, changing nothing.
//
// It does not hyphenate, so text it rejects as too wide may still be set
// with a hyphen by InDesign or Scribus, and the two engines can disagree
// at the boundary in any case. Treat a pass as "worth rendering" and the
// render as the decision.
func Evaluate(text string, c Constraints, m Measurer) Result {
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
	if c.FoldX > 0 && !c.FoldProtected {
		if word, gap := onFold(lines, c, m); word != "" {
			res.Why, res.Word = FailWordOnFold, word
			res.Reason = fmt.Sprintf("the spine of the spread runs through the word %q; the nearest "+
				"word break is %.0fpt away, so the wording has to change to put a space on the fold",
				word, math.Abs(gap))
			return res
		}
	}

	res.Fits = true
	return res
}

// Budget estimates how many characters of representative text fit. It is
// a hint for whoever writes the text and deliberately not a test: real
// character widths vary several-fold, so a length safe for one string is
// wrong for the next. Evaluate decides.
//
// sample should look like the language being written.
func Budget(sample string, c Constraints, m Measurer) int {
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

func measure(s string, c Constraints, m Measurer) float64 {
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
	minX, maxX, minY, maxY, ok := frameBounds(tf)
	if !ok {
		return Constraints{}, fmt.Errorf("fit: text frame %s has no usable path geometry", tf.Self)
	}

	c := Constraints{
		FrameWidth:  maxX - minX,
		FrameHeight: maxY - minY,
	}
	c.PointSize = StyleSize(styles, appliedStyle)
	c.Leading = StyleLeading(styles, appliedStyle, c.PointSize)
	c.FontFamily, c.FontStyle = StyleFont(styles, appliedStyle)

	if c.Leading > 0 {
		c.MaxLines = int(c.FrameHeight / c.Leading)
	}
	if c.MaxLines < 1 {
		c.MaxLines = 1
	}

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
