package spread

import (
	"encoding/xml"
	"fmt"
	"strconv"

	"github.com/dimelords/idmllib/v3/pkg/common"
)

// MasterSpreadOptions describes a parent page (master spread) to build with NewMasterSpread.
type MasterSpreadOptions struct {
	// Self is the master spread's ID. It must be unique in the document.
	Self string
	// NamePrefix and BaseName make the parent page's name, "A-Master": InDesign shows the prefix on the pages
	// that use it.
	NamePrefix, BaseName string
	// PageWidth and PageHeight are the page size in points.
	PageWidth, PageHeight float64
	// Facing builds a left and a right page instead of one single-sided page.
	Facing bool
	// Top, Bottom, Inside and Outside are the margins in points (see MarginPreference.SetMargins).
	Top, Bottom, Inside, Outside float64
	// Columns and Gutter set the column grid. ColumnWidths is optional, for columns of different widths.
	Columns      int
	Gutter       float64
	ColumnWidths []float64
	// PageSelfs optionally sets the pages' IDs; by default they are Self + "p1" (and "p2" on facing spreads).
	PageSelfs []string
}

// NewMasterSpread builds a parent page with its pages, page transforms, margins and columns, ready for
// idml.Package.AddMasterSpread.
func NewMasterSpread(opts MasterSpreadOptions) (*Spread, error) {
	if opts.Self == "" || opts.NamePrefix == "" || opts.BaseName == "" {
		return nil, common.Errorf("spread", "new master spread", "", "Self, NamePrefix and BaseName are required")
	}
	if opts.PageWidth <= 0 || opts.PageHeight <= 0 {
		return nil, common.Errorf("spread", "new master spread", "", "page size %v × %v: want positive", opts.PageWidth, opts.PageHeight)
	}
	sides := []PageSide{SideSingle}
	if opts.Facing {
		sides = []PageSide{SideLeft, SideRight}
	}
	if opts.PageSelfs != nil && len(opts.PageSelfs) != len(sides) {
		return nil, common.Errorf("spread", "new master spread", "", "%d page IDs for %d pages", len(opts.PageSelfs), len(sides))
	}

	ms := &Spread{
		XMLName:         xml.Name{Local: "MasterSpread"},
		Self:            opts.Self,
		ItemTransform:   "1 0 0 1 0 0",
		ShowMasterItems: "true",
		PageCount:       strconv.Itoa(len(sides)),
		OtherAttrs: []xml.Attr{
			{Name: xml.Name{Local: "OverriddenPageItemProps"}, Value: ""},
			{Name: xml.Name{Local: "Name"}, Value: opts.NamePrefix + "-" + opts.BaseName},
			{Name: xml.Name{Local: "NamePrefix"}, Value: opts.NamePrefix},
			{Name: xml.Name{Local: "BaseName"}, Value: opts.BaseName},
			{Name: xml.Name{Local: "PrimaryTextFrame"}, Value: "n"},
		},
	}
	for i, side := range sides {
		self := fmt.Sprintf("%sp%d", opts.Self, i+1)
		if opts.PageSelfs != nil {
			self = opts.PageSelfs[i]
		}
		mp := &MarginPreference{}
		mp.SetMargins(opts.Top, opts.Bottom, opts.Inside, opts.Outside)
		if err := mp.SetColumns(max(opts.Columns, 1), opts.Gutter, opts.PageWidth, opts.ColumnWidths...); err != nil {
			return nil, common.WrapError("spread", "new master spread", err)
		}
		ms.Pages = append(ms.Pages, Page{
			Self:                self,
			TabOrder:            "",
			AppliedMaster:       "n",
			OverrideList:        "",
			MasterPageTransform: "1 0 0 1 0 0",
			Name:                opts.NamePrefix,
			AppliedTrapPreset:   "TrapPreset/$ID/kDefaultTrapStyleName",
			GeometricBounds:     PageGeometricBounds(opts.PageWidth, opts.PageHeight),
			ItemTransform:       PageItemTransform(opts.PageWidth, opts.PageHeight, side),
			GridStartingPoint:   "TopOutside",
			UseMasterGrid:       "true",
			MarginPreference:    mp,
		})
	}
	return ms, nil
}

// MasterName returns a master spread's Name, NamePrefix and BaseName ("A-Master", "A", "Master").
func (s *Spread) MasterName() (name, prefix, base string) {
	for _, a := range s.OtherAttrs {
		switch a.Name.Local {
		case "Name":
			name = a.Value
		case "NamePrefix":
			prefix = a.Value
		case "BaseName":
			base = a.Value
		}
	}
	return name, prefix, base
}

// SetMasterName renames a master spread. The pages' names are set to the new prefix, as InDesign does.
func (s *Spread) SetMasterName(prefix, base string) {
	s.OtherAttrs = setAttr(s.OtherAttrs, "Name", prefix+"-"+base)
	s.OtherAttrs = setAttr(s.OtherAttrs, "NamePrefix", prefix)
	s.OtherAttrs = setAttr(s.OtherAttrs, "BaseName", base)
	for i := range s.Pages {
		s.Pages[i].Name = prefix
	}
}

func setAttr(attrs []xml.Attr, name, value string) []xml.Attr {
	for i := range attrs {
		if attrs[i].Name.Local == name {
			attrs[i].Value = value
			return attrs
		}
	}
	return append(attrs, xml.Attr{Name: xml.Name{Local: name}, Value: value})
}
