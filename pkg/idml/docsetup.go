package idml

import (
	"encoding/xml"
	"strconv"
	"strings"

	"github.com/dimelords/idmllib/v3/pkg/common"
	"github.com/dimelords/idmllib/v3/pkg/spread"
)

// DocumentSetup is the document-wide page setup from <DocumentPreference> in Resources/Preferences.xml. Lengths
// are in points.
type DocumentSetup struct {
	PageWidth, PageHeight float64
	FacingPages           bool
	PagesPerDocument      int
	// Binding is "LeftToRight" or "RightToLeft".
	Binding string
	// Bleed and Slug are top, bottom, inside (left on single-sided documents) and outside (right).
	Bleed, Slug [4]float64
}

// BaselineGrid is the document baseline grid from <GridPreference>. RelativeTo is "TopOfPage" or "TopOfMargin".
type BaselineGrid struct {
	Start, Division float64
	RelativeTo      string
}

var (
	bleedAttrs = [4]string{"DocumentBleedTopOffset", "DocumentBleedBottomOffset",
		"DocumentBleedInsideOrLeftOffset", "DocumentBleedOutsideOrRightOffset"}
	slugAttrs = [4]string{"SlugTopOffset", "SlugBottomOffset", "SlugInsideOrLeftOffset", "SlugRightOrOutsideOffset"}
)

const baselineRelativeSuffix = "OfBaselineGridRelativeOption"

// DocumentSetup reads the document's page setup.
func (p *Package) DocumentSetup() (DocumentSetup, error) {
	el, err := p.preferenceElement("DocumentPreference", false)
	if err != nil {
		return DocumentSetup{}, err
	}
	a := toAttrMap(el.Attrs)
	ds := DocumentSetup{
		PageWidth:        a.float("PageWidth"),
		PageHeight:       a.float("PageHeight"),
		FacingPages:      a["FacingPages"] == "true",
		PagesPerDocument: int(a.float("PagesPerDocument")),
		Binding:          a["PageBinding"],
	}
	for i := range 4 {
		ds.Bleed[i] = a.float(bleedAttrs[i])
		ds.Slug[i] = a.float(slugAttrs[i])
	}
	return ds, nil
}

// SetDocumentSetup writes the document's page setup. Attributes it doesn't model are kept. It changes the
// preferences only: pages that already exist keep their size (set it with spread.PageGeometricBounds and
// spread.PageItemTransform).
func (p *Package) SetDocumentSetup(ds DocumentSetup) error {
	if ds.PageWidth <= 0 || ds.PageHeight <= 0 {
		return common.Errorf("idml", "set document setup", PathPreferences, "page size %v × %v: want positive", ds.PageWidth, ds.PageHeight)
	}
	if ds.Binding != "" && ds.Binding != "LeftToRight" && ds.Binding != "RightToLeft" {
		return common.Errorf("idml", "set document setup", PathPreferences, "binding %q: want LeftToRight or RightToLeft", ds.Binding)
	}
	el, err := p.preferenceElement("DocumentPreference", true)
	if err != nil {
		return err
	}
	set := func(name, value string) { el.Attrs = setXMLAttr(el.Attrs, name, value) }
	set("PageWidth", formatPt(ds.PageWidth))
	set("PageHeight", formatPt(ds.PageHeight))
	set("FacingPages", strconv.FormatBool(ds.FacingPages))
	set("PagesPerDocument", strconv.Itoa(max(ds.PagesPerDocument, 1)))
	if ds.Binding != "" {
		set("PageBinding", ds.Binding)
	}
	for i := range 4 {
		set(bleedAttrs[i], formatPt(ds.Bleed[i]))
		set(slugAttrs[i], formatPt(ds.Slug[i]))
	}
	set("DocumentBleedUniformSize", strconv.FormatBool(uniform(ds.Bleed)))
	set("DocumentSlugUniformSize", strconv.FormatBool(uniform(ds.Slug)))
	p.markDirty(PathPreferences)
	return nil
}

// BaselineGrid reads the document baseline grid.
func (p *Package) BaselineGrid() (BaselineGrid, error) {
	el, err := p.preferenceElement("GridPreference", false)
	if err != nil {
		return BaselineGrid{}, err
	}
	a := toAttrMap(el.Attrs)
	return BaselineGrid{
		Start:      a.float("BaselineStart"),
		Division:   a.float("BaselineDivision"),
		RelativeTo: strings.TrimSuffix(a["BaselineGridRelativeOption"], baselineRelativeSuffix),
	}, nil
}

// SetBaselineGrid writes the document baseline grid. Attributes it doesn't model are kept.
func (p *Package) SetBaselineGrid(bg BaselineGrid) error {
	if bg.Division <= 0 {
		return common.Errorf("idml", "set baseline grid", PathPreferences, "division %v: want positive", bg.Division)
	}
	rel := bg.RelativeTo
	if rel == "" {
		rel = "TopOfPage"
	}
	if rel != "TopOfPage" && rel != "TopOfMargin" {
		return common.Errorf("idml", "set baseline grid", PathPreferences, "relative to %q: want TopOfPage or TopOfMargin", bg.RelativeTo)
	}
	el, err := p.preferenceElement("GridPreference", true)
	if err != nil {
		return err
	}
	el.Attrs = setXMLAttr(el.Attrs, "BaselineStart", formatPt(bg.Start))
	el.Attrs = setXMLAttr(el.Attrs, "BaselineDivision", formatPt(bg.Division))
	el.Attrs = setXMLAttr(el.Attrs, "BaselineGridRelativeOption", rel+baselineRelativeSuffix)
	p.markDirty(PathPreferences)
	return nil
}

// DefaultMargins reads the document's default <MarginPreference>, which new pages get.
func (p *Package) DefaultMargins() (spread.MarginPreference, error) {
	el, err := p.preferenceElement("MarginPreference", false)
	if err != nil {
		return spread.MarginPreference{}, err
	}
	a := toAttrMap(el.Attrs)
	return spread.MarginPreference{
		ColumnCount: a["ColumnCount"], ColumnGutter: a["ColumnGutter"],
		Top: a["Top"], Bottom: a["Bottom"], Left: a["Left"], Right: a["Right"],
		ColumnDirection: a["ColumnDirection"], ColumnsPositions: a["ColumnsPositions"],
	}, nil
}

// SetDefaultMargins writes the document's default <MarginPreference>. Empty fields are left as they are.
func (p *Package) SetDefaultMargins(m spread.MarginPreference) error {
	el, err := p.preferenceElement("MarginPreference", true)
	if err != nil {
		return err
	}
	for name, value := range map[string]string{
		"ColumnCount": m.ColumnCount, "ColumnGutter": m.ColumnGutter, "Top": m.Top, "Bottom": m.Bottom,
		"Left": m.Left, "Right": m.Right, "ColumnDirection": m.ColumnDirection, "ColumnsPositions": m.ColumnsPositions,
	} {
		if value != "" {
			el.Attrs = setXMLAttr(el.Attrs, name, value)
		}
	}
	p.markDirty(PathPreferences)
	return nil
}

// preferenceElement returns a top-level element of Preferences.xml, creating it when create is set and it is
// missing.
func (p *Package) preferenceElement(local string, create bool) (*common.RawXMLElement, error) {
	prefs, err := p.Preferences()
	if err != nil {
		return nil, err
	}
	for i := range prefs.Elements {
		if prefs.Elements[i].XMLName.Local == local {
			return &prefs.Elements[i], nil
		}
	}
	if !create {
		return nil, common.Errorf("idml", "read preferences", PathPreferences, "no <%s>", local)
	}
	prefs.Elements = append(prefs.Elements, common.RawXMLElement{XMLName: xml.Name{Local: local}})
	return &prefs.Elements[len(prefs.Elements)-1], nil
}

type attrMap map[string]string

func (a attrMap) float(name string) float64 {
	f, err := strconv.ParseFloat(strings.TrimSpace(a[name]), 64)
	if err != nil {
		return 0
	}
	return f
}

func setXMLAttr(list []xml.Attr, name, value string) []xml.Attr {
	for i := range list {
		if list[i].Name.Local == name {
			list[i].Value = value
			return list
		}
	}
	return append(list, xml.Attr{Name: xml.Name{Local: name}, Value: value})
}

func formatPt(f float64) string {
	if f == 0 {
		return "0"
	}
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func uniform(v [4]float64) bool { return v[0] == v[1] && v[1] == v[2] && v[2] == v[3] }

func toAttrMap(list []xml.Attr) attrMap {
	m := make(attrMap, len(list))
	for _, a := range list {
		m[a.Name.Local] = a.Value
	}
	return m
}
