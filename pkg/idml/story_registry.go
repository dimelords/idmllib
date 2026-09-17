package idml

import (
	"encoding/xml"
	"path"
	"slices"
	"strings"

	"github.com/dimelords/idmllib/v2/pkg/document"
)

// idPkgNamespace is the IDML packaging namespace used by designmap.xml
// resource references.
const idPkgNamespace = "http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging"

// registerStory makes a story file visible to InDesign by adding its
// <idPkg:Story> reference and its ID to the StoryList attribute of the
// Document. Both are idempotent.
func (p *Package) registerStory(filename, storyID string) error {
	if !p.hasFile(PathDesignmap) {
		return nil // a bare package under construction; nothing to register in yet
	}
	doc, err := p.Document()
	if err != nil {
		return err
	}
	hasRef := slices.ContainsFunc(doc.Stories, func(r document.ResourceRef) bool { return r.Src == filename })
	if !hasRef {
		doc.Stories = append(doc.Stories, document.ResourceRef{
			XMLName: xml.Name{Space: idPkgNamespace, Local: "Story"},
			Src:     filename,
		})
	}
	if storyID != "" {
		ids := strings.Fields(doc.StoryList)
		if !slices.Contains(ids, storyID) {
			doc.StoryList = strings.Join(append(ids, storyID), " ")
		}
	}
	p.markDirty(PathDesignmap)
	return nil
}

// unregisterStory removes a story file's reference and StoryList entry from
// the Document.
func (p *Package) unregisterStory(filename string) error {
	if !p.hasFile(PathDesignmap) {
		return nil
	}
	doc, err := p.Document()
	if err != nil {
		return err
	}
	doc.Stories = slices.DeleteFunc(doc.Stories, func(r document.ResourceRef) bool { return r.Src == filename })
	if id := storyIDFromPath(filename); id != "" {
		ids := slices.DeleteFunc(strings.Fields(doc.StoryList), func(s string) bool { return s == id })
		doc.StoryList = strings.Join(ids, " ")
	}
	p.markDirty(PathDesignmap)
	return nil
}

// storyIDFromPath derives the story ID from its file name:
// "Stories/Story_u123.xml" yields "u123".
func storyIDFromPath(filename string) string {
	base := strings.TrimSuffix(path.Base(filename), ExtXML)
	return strings.TrimPrefix(base, "Story_")
}
