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
	"strings"

	"github.com/dimelords/idmllib/v3/pkg/common"
	"github.com/dimelords/idmllib/v3/pkg/document"
	"github.com/dimelords/idmllib/v3/pkg/idml"
	"github.com/dimelords/idmllib/v3/pkg/resources"
	"github.com/dimelords/idmllib/v3/pkg/spread"
	"github.com/dimelords/idmllib/v3/pkg/story"
)

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

	if err := enrichDesignMap(pkg, s); err != nil {
		return nil, err
	}

	pkg.SetStyles(buildStyles())
	pkg.SetGraphics(buildGraphics())
	pkg.SetFonts(buildFonts())

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

// enrichDesignMap adds the designmap children a real document carries and
// the minimal template does not: numbering lists, named grids, sections,
// document users, colour groups. The document package's tests read each
// of these out of testdata/designmap.xml, which used to be lifted
// straight out of a customer's file - the document users were where four
// of their staff were named. Added here rather than in the template,
// because a minimal template should stay minimal.
// labelProperties builds a Properties element carrying an empty label,
// which is what InDesign writes on most designmap children. The parser
// tests check that Properties survives on each of them, so an element
// without one is not a realistic sample.
func labelProperties() *common.Properties {
	return &common.Properties{
		OtherElements: []common.RawXMLElement{{
			XMLName: xml.Name{Local: "Label"},
			Attrs:   []xml.Attr{{Name: xml.Name{Local: "type"}, Value: "list"}},
		}},
	}
}

func enrichDesignMap(pkg *idml.Package, spec Spec) error {
	doc, err := pkg.Document()
	if err != nil {
		return fmt.Errorf("fixturegen: document: %w", err)
	}

	doc.Self = "d"
	doc.Name = strings.TrimSuffix(spec.Name, ".idml")
	doc.ZeroPoint = "0 0"
	doc.ActiveLayer = "uba"
	doc.CMYKProfile = "U.S. Web Coated (SWOP) v2"
	doc.RGBProfile = "sRGB IEC61966-2.1"
	doc.Properties = labelProperties()
	doc.BackingStory = &document.ResourceRef{
		XMLName: xml.Name{
			Space: "http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging",
			Local: "BackingStory",
		},
		Src: "XML/BackingStory.xml",
	}

	doc.NumberingLists = append(doc.NumberingLists, document.NumberingList{
		Self:                           "NumberingList/$ID/[Default]",
		Name:                           "$ID/[Default]",
		ContinueNumbersAcrossStories:   "false",
		ContinueNumbersAcrossDocuments: "false",
	})
	doc.NamedGrids = append(doc.NamedGrids, document.NamedGrid{
		Self: "NamedGrid/$ID/[Page Grid]",
		Name: "$ID/[Page Grid]",
		GridDataInformation: &common.GridDataInformation{
			FontStyle:       "Roman",
			PointSize:       "12",
			CharacterAki:    "0",
			LineAki:         "9",
			HorizontalScale: "100",
			VerticalScale:   "100",
			Properties:      labelProperties(),
		},
	})
	doc.Sections = append(doc.Sections, document.Section{
		Self:              "ub4",
		Name:              "A",
		Length:            "2",
		PageNumberStart:   "22",
		SectionPrefix:     "A",
		ContinueNumbering: "false",
		Properties:        labelProperties(),
	})
	// InDesign's own placeholder for an unattributed user. The captured
	// document named four real members of a customer's staff here.
	doc.DocumentUsers = append(doc.DocumentUsers, document.DocumentUser{
		Self:       "dDocumentUser0",
		UserName:   "$ID/Unknown User Name",
		Properties: labelProperties(),
	})
	doc.ColorGroups = append(doc.ColorGroups, document.ColorGroup{
		Self:             "ColorGroup/[Root Color Group]",
		Name:             "[Root Color Group]",
		IsRootColorGroup: "true",
		ColorGroupSwatches: []document.ColorGroupSwatch{{
			Self:          "ColorGroupSwatch/Text",
			SwatchItemRef: "Color/Text",
		}, {
			Self:          "ColorGroupSwatch/Paper",
			SwatchItemRef: "Color/Paper",
		}},
	})
	doc.ABullets = append(doc.ABullets, document.ABullet{
		Self:           "dABullet0",
		CharacterType:  "UnicodeOnly",
		CharacterValue: "8226",
		Properties:     labelProperties(),
	})
	doc.Assignments = append(doc.Assignments, document.Assignment{
		Self:                    "uc9",
		Name:                    "$ID/UnassignedInCopy",
		UserName:                "$ID/Unknown User Name",
		ExportOptions:           "AssignedSpreads",
		IncludeLinksWhenPackage: "true",
		Properties:              labelProperties(),
	})
	doc.TextVariables = append(doc.TextVariables, document.TextVariable{
		Self:                    "dTextVariablenChapter Number",
		Name:                    "Chapter Number",
		VariableType:            "ChapterNumberType",
		ChapterNumberPreference: &document.ChapterNumberVariablePreference{Format: "Current"},
	}, document.TextVariable{
		Self:           "dTextVariablenCreation Date",
		Name:           "Creation Date",
		VariableType:   "CreationDateType",
		DatePreference: &document.DateVariablePreference{Format: "dd/MM/yy"},
	}, document.TextVariable{
		Self:         "dTextVariablenFile Name",
		Name:         "File Name",
		VariableType: "FileNameType",
		FileNamePreference: &document.FileNameVariablePreference{
			IncludePath:      "false",
			IncludeExtension: "false",
		},
	}, document.TextVariable{
		Self:         "dTextVariablenRunning Header",
		Name:         "Running Header",
		VariableType: "MatchParagraphStyleType",
		MatchParagraphStylePreference: &document.MatchParagraphStylePreference{
			SearchStrategy: "FirstOnPage",
		},
	})

	pkg.MarkModified(idml.PathDesignmap)
	return nil
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
		// A chain, not a flat list: the dependency tracker's job is to
		// walk from a style to its ancestors, and it cannot be tested
		// against styles that have none.
		cs := resources.CharacterStyle{
			Self:      "CharacterStyle/" + n,
			Name:      n,
			FontStyle: []string{"Regular", "Italic", "Bold"}[i%3],
		}
		// The first style is left with no parent at all, which is a real
		// thing for an InDesign style to be and gives the resolver a
		// termination case to be tested against. The rest chain to the
		// one before, so walking up from any of them has somewhere to go.
		if i > 0 {
			cs.Properties = basedOnProperty(fmt.Sprintf("CharacterStyle/emphasis-%02d", i))
		}
		chars.CharacterStyles = append(chars.CharacterStyles, cs)
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

// buildImage is one linked image. The link points at a relative name, not
// a path: an absolute one would record where the generator happened to
// run, which is how the captured fixtures came to carry a developer's
// home directory.
func buildImage(spreadNum, i int) *spread.Image {
	return &spread.Image{
		FrameContentBase: spread.FrameContentBase{
			Self:               fmt.Sprintf("image%d_%d", spreadNum, i),
			ImageTypeName:      "$ID/JPEG",
			AppliedObjectStyle: "ObjectStyle/Images%3aimage-01",
		},
		ActualPpi:    "72 72",
		EffectivePpi: "72 72",
		Link: &spread.Link{
			Self:               fmt.Sprintf("link%d_%d", spreadNum, i),
			LinkResourceURI:    fmt.Sprintf("file:placeholder-%d.jpg", i),
			LinkResourceFormat: "$ID/JPEG",
			StoredState:        "Normal",
		},
	}
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
	// A couple of rectangles per spread, one of them holding an image,
	// because the page-item tests look for both and a page of nothing but
	// text frames is not a realistic document either.
	for i := 1; i <= 2; i++ {
		r := spread.Rectangle{
			PageItemBase: spread.PageItemBase{
				Self:          fmt.Sprintf("rect%d_%d", n, i),
				ItemTransform: fmt.Sprintf("1 0 0 1 %d 0", 300*i),
			},
			AppliedObjectStyle: "ObjectStyle/frame-01",
		}
		// Both graphic-typed. An Unassigned rectangle here makes
		// SelectAllGraphicsInSpread return it and its own test then reject
		// it as "not a graphic type" - the selection code and that
		// expectation disagree about whether an empty frame counts. The
		// captured fixture happened to contain no such rectangle, so the
		// disagreement never showed. Worth settling, but not by having
		// the fixture take a side.
		r.ContentType = "GraphicType"
		// The image lives inside its frame, which is where InDesign puts
		// it and where a selection of the frame will find it. Hung off the
		// spread instead, it exports as a snippet with no graphic in it.
		if i == 1 {
			r.Image = buildImage(n, i)
		}
		sp.Rectangles = append(sp.Rectangles, r)
	}

	// One frame per story, stacked down the page.
	y := 0.0
	for s := firstStory; s <= lastStory; s++ {
		sp.TextFrames = append(sp.TextFrames, spread.TextFrame{
			PageItemBase: spread.PageItemBase{
				Self:          fmt.Sprintf("frame%02d", s),
				ItemTransform: fmt.Sprintf("1 0 0 1 0 %.1f", y),
			},
			AppliedObjectStyle: "ObjectStyle/frame-02",
			ParentStory:        fmt.Sprintf("story%02d", s),
		})
		y += 40
	}
	return sp
}
