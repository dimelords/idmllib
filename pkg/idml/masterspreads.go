package idml

import (
	"encoding/xml"
	"slices"
	"strings"

	"github.com/dimelords/idmllib/v3/pkg/common"
	"github.com/dimelords/idmllib/v3/pkg/document"
	"github.com/dimelords/idmllib/v3/pkg/spread"
)

// AddMasterSpread adds a parent page (master spread) and registers it as an <idPkg:MasterSpread> in
// designmap.xml, without which InDesign ignores the file. ms must be a master spread: build one with
// spread.NewMasterSpread, or read one with MasterSpread and change it. Its Self and its pages' Selfs must not
// be used by another spread.
//
// Example:
//
//	ms, _ := spread.NewMasterSpread(spread.MasterSpreadOptions{
//	    Self: "umsport", NamePrefix: "S", BaseName: "Sport",
//	    PageWidth: 771.02, PageHeight: 1128.19, Facing: true,
//	    Top: 36.85, Bottom: 36.85, Inside: 34.02, Outside: 34.02,
//	    Columns: 5, Gutter: 11.34,
//	})
//	err := pkg.AddMasterSpread(idml.MasterSpreadPath("umsport"), ms)
func (p *Package) AddMasterSpread(filename string, ms *spread.Spread) error {
	if !IsMasterSpreadPath(filename) {
		return common.Errorf("idml", "add master spread", filename, "not a MasterSpreads/ path")
	}
	if ms == nil || !ms.IsMaster() {
		return common.Errorf("idml", "add master spread", filename, "not a master spread (build one with spread.NewMasterSpread)")
	}
	if p.hasFile(filename) {
		return common.WrapErrorWithPath("idml", "add master spread", filename, common.ErrAlreadyExists)
	}
	if err := p.checkSelfsUnused(ms); err != nil {
		return common.WrapErrorWithPath("idml", "add master spread", filename, err)
	}
	if err := p.marshalAndUpdateSpread(filename, ms); err != nil {
		return err
	}
	return p.registerMasterSpread(filename)
}

// RemoveMasterSpread removes a parent page and its designmap reference. Pages that apply it would point at
// nothing, so by default the call fails and names them; with detach, they get no parent page instead
// (AppliedMaster="n"), as when "[None]" is applied in InDesign.
func (p *Package) RemoveMasterSpread(filename string, detach bool) error {
	if !IsMasterSpreadPath(filename) {
		return common.Errorf("idml", "remove master spread", filename, "not a MasterSpreads/ path")
	}
	ms, err := p.MasterSpread(filename)
	if err != nil {
		return err
	}
	users, err := p.pagesApplying(ms.Self, filename)
	if err != nil {
		return err
	}
	if len(users) > 0 && !detach {
		var names []string
		for _, u := range users {
			names = append(names, u.page.Self)
		}
		return common.Errorf("idml", "remove master spread", filename,
			"applied to %d page(s) (%s); pass detach to remove it from them", len(users), strings.Join(names, ", "))
	}
	for _, u := range users {
		u.page.AppliedMaster = "n"
		p.markDirty(u.file)
	}
	p.dropFile(filename)
	return p.unregisterResource(filename, func(doc *document.Document) *[]document.ResourceRef { return &doc.MasterSpreads })
}

// RemoveSpread removes a document spread and its designmap reference. A section that starts on one of its pages
// moves to the first page of the next spread, or of the previous one when the spread was last; if that page
// already starts a section, the moved section is dropped. The last spread can't be removed: a document needs a
// page. Stories in the spread's text frames are kept; remove them with RemoveStory when nothing else uses them.
func (p *Package) RemoveSpread(filename string) error {
	if !IsSpreadPath(filename) {
		return common.Errorf("idml", "remove spread", filename, "not a Spreads/ path")
	}
	sp, err := p.Spread(filename)
	if err != nil {
		return err
	}
	doc, err := p.Document()
	if err != nil {
		return err
	}
	i := slices.IndexFunc(doc.Spreads, func(r document.ResourceRef) bool { return r.Src == filename })
	if i >= 0 && len(doc.Spreads) == 1 {
		return common.Errorf("idml", "remove spread", filename, "the document's last spread can't be removed")
	}

	removed := map[string]bool{}
	for _, pg := range sp.Pages {
		removed[pg.Self] = true
	}
	// The page sections move to: first page of the next spread, else of the previous one.
	target := ""
	for _, j := range []int{i + 1, i - 1} {
		if i < 0 || j < 0 || j >= len(doc.Spreads) {
			continue
		}
		other, err := p.Spread(doc.Spreads[j].Src)
		if err != nil {
			return err
		}
		if len(other.Pages) > 0 {
			target = other.Pages[0].Self
			break
		}
	}
	startsAt := map[string]bool{}
	for _, s := range doc.Sections {
		if !removed[s.PageStart] {
			startsAt[s.PageStart] = true
		}
	}
	var sections []document.Section
	for _, s := range doc.Sections {
		if removed[s.PageStart] {
			if target == "" || startsAt[target] {
				continue
			}
			s.PageStart = target
			startsAt[target] = true
		}
		sections = append(sections, s)
	}
	doc.Sections = sections
	p.markDirty(PathDesignmap)

	p.dropFile(filename)
	return p.unregisterResource(filename, func(doc *document.Document) *[]document.ResourceRef { return &doc.Spreads })
}

// ApplyMaster applies a parent page to a page of a spread. masterSelf is the master spread's Self, or "n" for
// no parent page.
func (p *Package) ApplyMaster(spreadFile, pageSelf, masterSelf string) error {
	if masterSelf != "n" {
		found := false
		masters, err := p.MasterSpreads()
		if err != nil {
			return err
		}
		for _, ms := range masters {
			if ms.Self == masterSelf {
				found = true
				break
			}
		}
		if !found {
			return common.Errorf("idml", "apply master", spreadFile, "no master spread %q", masterSelf)
		}
	}
	sp, err := p.Spread(spreadFile)
	if err != nil {
		return err
	}
	for i := range sp.Pages {
		if sp.Pages[i].Self == pageSelf {
			sp.Pages[i].AppliedMaster = masterSelf
			p.markDirty(spreadFile)
			return nil
		}
	}
	return common.Errorf("idml", "apply master", spreadFile, "no page %q", pageSelf)
}

// pageRef is a page found in one of the package's spread files.
type pageRef struct {
	file string
	page *spread.Page
}

// pagesApplying returns the pages, in spreads and other master spreads, that apply master spread self.
func (p *Package) pagesApplying(self, except string) ([]pageRef, error) {
	var refs []pageRef
	for _, filename := range slices.Sorted(slices.Values(p.Files())) {
		if filename == except || !isAnySpreadPath(filename) {
			continue
		}
		sp, err := p.Spread(filename)
		if err != nil {
			return nil, err
		}
		for i := range sp.Pages {
			if sp.Pages[i].AppliedMaster == self {
				refs = append(refs, pageRef{filename, &sp.Pages[i]})
			}
		}
	}
	return refs, nil
}

// checkSelfsUnused fails when a spread's Self or one of its pages' Selfs is already used by another spread.
func (p *Package) checkSelfsUnused(sp *spread.Spread) error {
	want := map[string]bool{sp.Self: true}
	for _, pg := range sp.Pages {
		want[pg.Self] = true
	}
	for _, filename := range p.Files() {
		if !isAnySpreadPath(filename) {
			continue
		}
		other, err := p.Spread(filename)
		if err != nil {
			return err
		}
		if want[other.Self] {
			return common.Errorf("idml", "check ids", filename, "id %q is already used", other.Self)
		}
		for _, pg := range other.Pages {
			if want[pg.Self] {
				return common.Errorf("idml", "check ids", filename, "page id %q is already used", pg.Self)
			}
		}
	}
	return nil
}

// registerMasterSpread adds a master spread's <idPkg:MasterSpread> reference to the Document, after the
// existing ones, as InDesign orders them.
func (p *Package) registerMasterSpread(filename string) error {
	if !p.hasFile(PathDesignmap) {
		return nil
	}
	doc, err := p.Document()
	if err != nil {
		return err
	}
	if !slices.ContainsFunc(doc.MasterSpreads, func(r document.ResourceRef) bool { return r.Src == filename }) {
		doc.MasterSpreads = append(doc.MasterSpreads, document.ResourceRef{
			XMLName: xml.Name{Space: idPkgNamespace, Local: "MasterSpread"},
			Src:     filename,
		})
	}
	p.markDirty(PathDesignmap)
	return nil
}

// unregisterResource removes filename from one of the Document's reference lists.
func (p *Package) unregisterResource(filename string, list func(*document.Document) *[]document.ResourceRef) error {
	if !p.hasFile(PathDesignmap) {
		return nil
	}
	doc, err := p.Document()
	if err != nil {
		return err
	}
	refs := list(doc)
	*refs = slices.DeleteFunc(*refs, func(r document.ResourceRef) bool { return r.Src == filename })
	p.markDirty(PathDesignmap)
	return nil
}

// dropFile removes a file, its cached parse and its modified mark, so Write neither writes nor re-creates it.
func (p *Package) dropFile(filename string) {
	p.invalidateCache(filename)
	p.clearDirty(filename)
	p.removeFile(filename)
}

// isAnySpreadPath reports whether path is a spread or a master spread file.
func isAnySpreadPath(path string) bool { return IsSpreadPath(path) || IsMasterSpreadPath(path) }
