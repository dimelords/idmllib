package idml

import (
	"github.com/dimelords/idmllib/v2/pkg/common"
	"github.com/dimelords/idmllib/v2/pkg/document"
	"github.com/dimelords/idmllib/v2/pkg/spread"
	"github.com/dimelords/idmllib/v2/pkg/story"
)

// Typed getters. Each file is parsed on first access and cached; later calls
// return the same pointer. Edits made through a returned pointer are written
// by Write only after MarkModified(path) or the corresponding Update method.
//
// The plain getters return the element callers work with (the <Story>, the
// <Spread>, the <Document>). The *File getters return the idPkg wrapper that
// also carries the DOMVersion and, for designmap.xml, the XML declaration and
// processing instructions.

// DocumentFile returns the designmap.xml wrapper around Document.
func (p *Package) DocumentFile() (*document.File, error) {
	if _, err := p.Document(); err != nil {
		return nil, err
	}
	return p.documentMetadata, nil
}

// StoryFile returns the parsed story file at filename, e.g. "Stories/Story_u123.xml".
func (p *Package) StoryFile(filename string) (*story.File, error) {
	if st, cached := p.getCachedStory(filename); cached {
		return st, nil
	}
	entry, err := p.getFileEntry(filename)
	if err != nil {
		return nil, err
	}
	st, err := story.ParseStory(entry.data)
	if err != nil {
		return nil, common.WrapErrorWithPath("idml", "parse story", filename, err)
	}
	p.cacheStory(filename, st)
	return st, nil
}

// Story returns the <Story> element of the story file at filename.
func (p *Package) Story(filename string) (*story.Story, error) {
	f, err := p.StoryFile(filename)
	if err != nil {
		return nil, err
	}
	return &f.Story, nil
}

// Stories returns every story in the package, keyed by path. Stories held only
// in the cache, without a backing file yet, are included.
func (p *Package) Stories() (map[string]*story.Story, error) {
	stories := make(map[string]*story.Story)
	for filename, st := range p.stories {
		stories[filename] = &st.Story
	}
	for filename := range p.files {
		if !IsStoryPath(filename) {
			continue
		}
		if _, done := stories[filename]; done {
			continue
		}
		st, err := p.Story(filename)
		if err != nil {
			return nil, err
		}
		stories[filename] = st
	}
	return stories, nil
}

// SpreadFile returns the parsed spread or master spread file at filename.
func (p *Package) SpreadFile(filename string) (*spread.File, error) {
	if sp, cached := p.getCachedSpread(filename); cached {
		return sp, nil
	}
	entry, err := p.getFileEntry(filename)
	if err != nil {
		return nil, err
	}
	sp, err := spread.ParseSpread(p.prepareSpreadData(filename, entry.data))
	if err != nil {
		return nil, common.WrapErrorWithPath("idml", "parse spread", filename, err)
	}
	p.cacheSpread(filename, sp)
	return sp, nil
}

// Spread returns the <Spread> (or <MasterSpread>) element of the file at filename.
func (p *Package) Spread(filename string) (*spread.Spread, error) {
	f, err := p.SpreadFile(filename)
	if err != nil {
		return nil, err
	}
	return &f.Spread, nil
}

// Spreads returns every regular spread in the package, keyed by path.
// Master spreads are returned by MasterSpreads.
func (p *Package) Spreads() (map[string]*spread.Spread, error) {
	return p.spreadsWithPrefix(IsSpreadPath)
}

// MasterSpread returns the <MasterSpread> element of the file at filename,
// e.g. "MasterSpreads/MasterSpread_ub4.xml". The result reports IsMaster().
func (p *Package) MasterSpread(filename string) (*spread.Spread, error) {
	return p.Spread(filename)
}

// MasterSpreads returns every master spread in the package, keyed by path.
func (p *Package) MasterSpreads() (map[string]*spread.Spread, error) {
	return p.spreadsWithPrefix(IsMasterSpreadPath)
}

func (p *Package) spreadsWithPrefix(match func(string) bool) (map[string]*spread.Spread, error) {
	spreads := make(map[string]*spread.Spread)
	for filename := range p.files {
		if !match(filename) {
			continue
		}
		sp, err := p.Spread(filename)
		if err != nil {
			return nil, err
		}
		spreads[filename] = sp
	}
	return spreads, nil
}

// domVersion returns the document's DOM version, used when a new file wrapper
// has to be created for content added through the API.
func (p *Package) domVersion() string {
	if doc, err := p.Document(); err == nil && doc.DOMVersion != "" {
		return doc.DOMVersion
	}
	return "20.4"
}

// marshalAndUpdateSpread writes sp into the spread file at spreadFilename and
// re-marshals it. sp is normally the pointer returned by Spread, in which case
// the cached file already holds it.
func (p *Package) marshalAndUpdateSpread(spreadFilename string, sp *spread.Spread) error {
	var f *spread.File
	if p.hasFile(spreadFilename) {
		existing, err := p.SpreadFile(spreadFilename)
		if err != nil {
			return common.WrapErrorWithPath("idml", "marshal spread", spreadFilename, err)
		}
		f = existing
		if sp != &f.Spread {
			f.Spread = *sp
		}
	} else {
		f = &spread.File{DOMVersion: p.domVersion(), Spread: *sp}
		p.cacheSpread(spreadFilename, f)
	}
	data, err := spread.MarshalSpread(f)
	if err != nil {
		return common.WrapErrorWithPath("idml", "marshal spread", spreadFilename, err)
	}
	p.setFileData(spreadFilename, p.restoreSpreadData(spreadFilename, data))
	p.markDirty(spreadFilename)
	return nil
}

// marshalAndUpdateStory writes st into the story file at filename, creating
// the file wrapper for a new story, and re-marshals it.
func (p *Package) marshalAndUpdateStory(filename string, st *story.Story) error {
	var f *story.File
	if p.hasFile(filename) {
		existing, err := p.StoryFile(filename)
		if err != nil {
			return common.WrapErrorWithPath("idml", "marshal story", filename, err)
		}
		f = existing
		if st != &f.Story {
			f.Story = *st
		}
	} else {
		f = &story.File{DOMVersion: p.domVersion(), Story: *st}
		p.cacheStory(filename, f)
	}
	data, err := story.MarshalStory(f)
	if err != nil {
		return common.WrapErrorWithPath("idml", "marshal story", filename, err)
	}
	p.setFileData(filename, data)
	p.markDirty(filename)
	return nil
}
