package idml

import (
	"strings"
)

// Modification tracking.
//
// Package caches parsed objects on first access. Reading a file through a
// getter such as Story or Styles must not change what Write produces, so only
// files that were modified through the API are re-marshaled on Write. All
// other files are written back exactly as they were read.
//
// The Add*, Update*, Remove* and Set* methods mark their targets modified.
// Code that edits an object through the pointer returned by a getter must call
// MarkModified for the file it changed, otherwise the change is not written.

// MarkModified records that the cached object for path has been changed and
// must be re-marshaled by Write. Use it after editing an object obtained from a
// getter such as Document, Story, Spread, Styles, Fonts or Graphics.
//
// Paths are the IDML package paths, for example "designmap.xml",
// "Stories/Story_u123.xml" or "Resources/Styles.xml".
func (p *Package) MarkModified(path string) {
	p.markDirty(path)
}

// IsModified reports whether path has been marked modified since it was read.
func (p *Package) IsModified(path string) bool {
	return p.isDirty(path)
}

func (p *Package) markDirty(path string) {
	if p.dirty == nil {
		p.dirty = make(map[string]bool)
	}
	p.dirty[path] = true
}

func (p *Package) isDirty(path string) bool {
	return p.dirty[path]
}

func (p *Package) clearDirty(path string) {
	delete(p.dirty, path)
}

// IsMasterSpreadPath checks if a path is in the MasterSpreads directory.
func IsMasterSpreadPath(path string) bool {
	return strings.HasPrefix(path, PrefixMasterSpreads) && strings.HasSuffix(path, ExtXML)
}
