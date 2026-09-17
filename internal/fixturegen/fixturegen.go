// Package fixturegen builds the IDML documents the test suite reads from
// testdata.
//
// These fixtures are generated rather than captured, and that is the
// point. They used to be real InDesign exports from a newspaper's
// production template, which meant this public repository carried that
// paper's name, its editorial style taxonomy, its licensed typefaces,
// its internal file-server paths and the names of four of its staff.
// None of that was ever what the tests needed: they exercise structure -
// nested style groups, style inheritance, story and spread counts,
// colour and gradient references, roundtrip fidelity - and structure is
// what a generator can produce without borrowing anybody's document.
//
// The shape deliberately matches what the captured fixtures had, so the
// tests keep testing what they tested: 131 paragraph styles across three
// groups, 27 character styles, 22 object styles across two groups, 20
// colours, a swatch and a gradient. Only the names changed.
//
// Regenerate with: go test ./internal/fixturegen -update
package fixturegen

import (
	"encoding/xml"
	"fmt"

	"github.com/dimelords/idmllib/v3/pkg/common"
	"github.com/dimelords/idmllib/v3/pkg/idml"
	"github.com/dimelords/idmllib/v3/pkg/resources"
	"github.com/dimelords/idmllib/v3/pkg/spread"
	"github.com/dimelords/idmllib/v3/pkg/story"
	"github.com/dimelords/idmllib/v3/pkg/xmp"
)

// xmpPacket is the document metadata a real IDML carries in
// META-INF/metadata.xml. The dates are fixed rather than taken from the
// clock so regenerating the fixtures produces byte-identical files; a
// generator whose output changed on every run would make every golden
// diff meaningless.
const xmpPacket = `<?xpacket begin="\ufeff" id="W5M0MpCehiHzreSzNTczkc9d"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
  <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
    <rdf:Description rdf:about=""
        xmlns:xmp="http://ns.adobe.com/xap/1.0/"
        xmlns:dc="http://purl.org/dc/elements/1.1/">
      <xmp:CreatorTool>Adobe InDesign</xmp:CreatorTool>
      <xmp:CreateDate>2024-01-01T00:00:00Z</xmp:CreateDate>
      <xmp:ModifyDate>2024-01-01T00:00:00Z</xmp:ModifyDate>
      <xmp:MetadataDate>2024-01-01T00:00:00Z</xmp:MetadataDate>
      <dc:format>application/x-indesign</dc:format>
    </rdf:Description>
  </rdf:RDF>
</x:xmpmeta>
<?xpacket end="r"?>`

// DOMVersion is the InDesign DOM version the fixtures claim.
const DOMVersion = "20.4"

// Spec describes one fixture document.
type Spec struct {
	// Name is the file name written under testdata.
	Name string
	// Stories and Spreads are the counts to produce. They match the
	// captured fixtures these replaced, because tests count them.
	Stories int
	Spreads int
}

// Specs are the fixtures this package owns.
func Specs() []Spec {
	return []Spec{
		{Name: "example.idml", Stories: 13, Spreads: 1},
		{Name: "tripple.idml", Stories: 29, Spreads: 3},
	}
}

// Build assembles one fixture document.
func Build(s Spec) (*idml.Package, error) {
	opts := idml.DefaultTemplateOptions()
	opts.DOMVersion = DOMVersion
	opts.Preset = idml.PresetA4
	opts.ColumnCount = 3

	pkg, err := idml.NewFromTemplate(opts)
	if err != nil {
		return nil, fmt.Errorf("fixturegen: template: %w", err)
	}

	pkg.SetStyles(buildStyles())
	pkg.SetGraphics(buildGraphics())
	pkg.SetFonts(buildFonts())
	pkg.SetXMP(xmp.Parse(xmpPacket))

	for i := 1; i <= s.Stories; i++ {
		name := fmt.Sprintf("Stories/Story_story%02d.xml", i)
		if err := pkg.AddStory(name, buildStory(i), idml.ValidationOptions{}); err != nil {
			return nil, fmt.Errorf("fixturegen: add story %d: %w", i, err)
		}
	}

	// Stories are dealt out over the spreads so every frame points at a
	// story that exists; a dangling ParentStory would make the validation
	// tests fail for a reason that has nothing to do with what they test.
	per := s.Stories / s.Spreads
	for i := 1; i <= s.Spreads; i++ {
		first := (i-1)*per + 1
		last := i * per
		if i == s.Spreads {
			last = s.Stories
		}
		// Named the way InDesign names them, and the way the tests
		// already reference them. The id is opaque - it says nothing
		// about whose document it once was.
		name := fmt.Sprintf("Spreads/Spread_u%d.xml", 209+i)
		if err := pkg.AddSpread(name, buildSpread(i, first, last)); err != nil {
			return nil, fmt.Errorf("fixturegen: add spread %d: %w", i, err)
		}
	}

	return pkg, nil
}

// paragraphStyleNames returns the leaf names for one group, generated so
// the count is exact and the names say nothing about anybody.
func paragraphStyleNames(prefix string, n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = fmt.Sprintf("%s-%02d", prefix, i+1)
	}
	return out
}

func buildStyles() *resources.StylesFile {
	// 131 paragraph styles over three groups, matching the captured
	// fixture: a flat set at the root plus two nested groups, because the
	// hierarchy tests walk exactly that shape.
	root := &resources.ParagraphStyleGroup{
		Self: "u1",
		Name: "$ID/",
		// InDesign's own built-ins. Documents reference them constantly
		// and the style-resolution tests look them up by these exact
		// names, so a fixture without them is not a realistic document.
		ParagraphStyles: []resources.ParagraphStyle{{
			Self: "ParagraphStyle/$ID/[No paragraph style]",
			Name: "$ID/[No paragraph style]",
		}, {
			Self: "ParagraphStyle/$ID/NormalParagraphStyle",
			Name: "$ID/NormalParagraphStyle",
		}},
	}
	for i, n := range paragraphStyleNames("body", 61) {
		// Every style after the first is based on the one before it, so
		// the hierarchy tests have a real chain to walk rather than a
		// flat list that happens to parse.
		basedOn := "ParagraphStyle/$ID/NormalParagraphStyle"
		if i > 0 {
			basedOn = "ParagraphStyle/body-" + fmt.Sprintf("%02d", i)
		}
		root.ParagraphStyles = append(root.ParagraphStyles, resources.ParagraphStyle{
			Self:       "ParagraphStyle/" + n,
			Name:       n,
			PointSize:  fmt.Sprintf("%d", 9+i%4),
			FillColor:  "Color/Text",
			Properties: basedOnProperty(basedOn),
		})
	}
	for _, g := range []struct {
		self, name, prefix string
		count              int
	}{
		{"pgroup1", "Editorial", "heading", 40},
		{"pgroup2", "Captions", "caption", 30},
	} {
		grp := resources.ParagraphStyleGroup{Self: g.self, Name: g.name}
		for i, n := range paragraphStyleNames(g.prefix, g.count) {
			grp.ParagraphStyles = append(grp.ParagraphStyles, resources.ParagraphStyle{
				Self:      fmt.Sprintf("ParagraphStyle/%s%%3a%s", g.name, n),
				Name:      n,
				PointSize: fmt.Sprintf("%d", 8+i%6),
				FillColor: "Color/Text",
			})
		}
		root.NestedGroups = append(root.NestedGroups, grp)
	}

	chars := &resources.CharacterStyleGroup{Self: "u2", Name: "$ID/",
		CharacterStyles: []resources.CharacterStyle{{
			Self: "CharacterStyle/$ID/[No character style]",
			Name: "$ID/[No character style]",
		}},
	}
	for i, n := range paragraphStyleNames("emphasis", 27) {
		chars.CharacterStyles = append(chars.CharacterStyles, resources.CharacterStyle{
			Self:      "CharacterStyle/" + n,
			Name:      n,
			FontStyle: []string{"Regular", "Italic", "Bold"}[i%3],
		})
	}

	objRoot := &resources.ObjectStyleGroup{Self: "u3", Name: "$ID/",
		ObjectStyles: []resources.ObjectStyle{{
			Self: "ObjectStyle/$ID/[None]",
			Name: "$ID/[None]",
		}},
	}
	for _, n := range paragraphStyleNames("frame", 12) {
		objRoot.ObjectStyles = append(objRoot.ObjectStyles, resources.ObjectStyle{
			Self: "ObjectStyle/" + n, Name: n,
		})
	}
	objNested := resources.ObjectStyleGroup{Self: "ogroup1", Name: "Images"}
	for _, n := range paragraphStyleNames("image", 10) {
		objNested.ObjectStyles = append(objNested.ObjectStyles, resources.ObjectStyle{
			Self: "ObjectStyle/Images%3a" + n, Name: n,
		})
	}
	objRoot.NestedGroups = append(objRoot.NestedGroups, objNested)

	return &resources.StylesFile{
		DOMVersion:              DOMVersion,
		RootParagraphStyleGroup: root,
		RootCharacterStyleGroup: chars,
		RootObjectStyleGroup:    objRoot,
	}
}

// basedOnProperty builds the Properties element that carries a style's
// parent. BasedOn is not an attribute: InDesign writes it as a child of
// Properties, which is why it does not appear on the struct.
func basedOnProperty(parent string) *common.Properties {
	return &common.Properties{
		OtherElements: []common.RawXMLElement{{
			XMLName: xml.Name{Local: "BasedOn"},
			Attrs:   []xml.Attr{{Name: xml.Name{Local: "type"}, Value: "object"}},
			Content: []byte(parent),
		}},
	}
}

func buildGraphics() *resources.GraphicFile {
	g := &resources.GraphicFile{DOMVersion: DOMVersion}
	// Twenty colours, in a mix of spaces so the colour-tracking tests see
	// more than one kind.
	specs := []struct {
		name, space, value string
	}{
		{"Text", "CMYK", "0 0 0 100"},
		{"Paper", "CMYK", "0 0 0 0"},
		{"Registration", "CMYK", "100 100 100 100"},
	}
	for i := len(specs); i < 20; i++ {
		specs = append(specs, struct{ name, space, value string }{
			fmt.Sprintf("Accent-%02d", i-2),
			[]string{"CMYK", "RGB"}[i%2],
			[]string{"12 34 56 0", "10 120 200"}[i%2],
		})
	}
	for _, s := range specs {
		g.Colors = append(g.Colors, resources.Color{
			Self:       "Color/" + s.name,
			Name:       s.name,
			Model:      "Process",
			Space:      s.space,
			ColorValue: s.value,
		})
	}
	g.Swatches = append(g.Swatches, resources.Swatch{
		Self: "Swatch/None", Name: "None",
	})
	g.Gradients = append(g.Gradients, resources.Gradient{
		Self: "Gradient/Fade", Name: "Fade", Type: "Linear",
	})
	return g
}

func buildFonts() *resources.FontsFile {
	f := &resources.FontsFile{DOMVersion: DOMVersion}
	// Fonts that ship with InDesign itself, so the fixture names no
	// licensed face belonging to anyone.
	for i, family := range []string{"Minion Pro", "Myriad Pro", "Courier Std"} {
		ff := resources.FontFamily{
			Self: fmt.Sprintf("font-family-%d", i+1),
			Name: family,
		}
		for _, style := range []string{"Regular", "Italic", "Bold"} {
			ff.Fonts = append(ff.Fonts, resources.Font{
				Self:           fmt.Sprintf("%s-%s", ff.Self, style),
				FontFamily:     family,
				Name:           family + " " + style,
				PostScriptName: fmt.Sprintf("%s-%s", family, style),
				FontStyleName:  style,
				Status:         "Installed",
				FontType:       "OpenTypeCFF",
			})
		}
		f.FontFamilies = append(f.FontFamilies, ff)
	}
	return f
}

func buildStory(n int) *story.Story {
	st := &story.Story{
		Self:     fmt.Sprintf("story%02d", n),
		UserText: "true",
	}
	for p := 1; p <= 3; p++ {
		st.ParagraphStyleRanges = append(st.ParagraphStyleRanges, story.ParagraphStyleRange{
			AppliedParagraphStyle: "ParagraphStyle/body-01",
			CharacterStyleRanges: []story.CharacterStyleRange{{
				// CharacterStyleRange and Content are marshaled by hand
				// so their children keep document order, which means the
				// struct tag does not supply a name and it has to be set
				// here. Leaving it off fails with "start tag with no
				// name", a long way from the cause.
				XMLName:               xml.Name{Local: "CharacterStyleRange"},
				AppliedCharacterStyle: "CharacterStyle/$ID/[No character style]",
				Children: []story.CharacterChild{{
					Content: &story.Content{
						XMLName: xml.Name{Local: "Content"},
						Text:    fmt.Sprintf("Story %d, paragraph %d. The quick brown fox jumps over the lazy dog.", n, p),
					},
				}},
			}},
		})
	}
	return st
}

func buildSpread(n, firstStory, lastStory int) *spread.Spread {
	sp := &spread.Spread{
		Self:            fmt.Sprintf("u%d", 209+n),
		PageCount:       "1",
		BindingLocation: "0",
		ItemTransform:   "1 0 0 1 0 0",
	}
	sp.Pages = append(sp.Pages, spread.Page{
		Self:            fmt.Sprintf("upage%d", 209+n),
		Name:            fmt.Sprintf("%d", n),
		GeometricBounds: "0 0 841.89 595.276",
		ItemTransform:   "1 0 0 1 0 0",
	})
	// One frame per story, stacked down the page.
	y := 0.0
	for s := firstStory; s <= lastStory; s++ {
		sp.TextFrames = append(sp.TextFrames, spread.TextFrame{
			PageItemBase: spread.PageItemBase{
				Self:          fmt.Sprintf("frame%02d", s),
				ItemTransform: fmt.Sprintf("1 0 0 1 0 %.1f", y),
			},
			ParentStory: fmt.Sprintf("story%02d", s),
		})
		y += 40
	}
	return sp
}
