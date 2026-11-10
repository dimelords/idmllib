package idml

import (
	"sort"

	"github.com/dimelords/idmllib/v3/pkg/common"
	"github.com/dimelords/idmllib/v3/pkg/document"
)

// AddCustomTextVariable adds a custom text variable named name with the given text to designmap.xml, and returns
// it. Place it in a story with story.NewTextVariableRange. It fails when the document already has a variable of
// that name.
func (p *Package) AddCustomTextVariable(name, contents string) (*document.TextVariable, error) {
	doc, err := p.Document()
	if err != nil {
		return nil, err
	}
	v, err := doc.AddTextVariable(document.NewCustomTextVariable(name, contents))
	if err != nil {
		return nil, err
	}
	p.markDirty(PathDesignmap)
	return v, nil
}

// SetCustomTextVariable sets the text of the custom text variable named name, for instance the date of the
// edition being built, and writes it into the ResultText of the variable's instances in every story, so that a
// reader that doesn't recompose the text sees the new value too. It returns how many instances it updated. It
// fails when there is no such variable or it isn't a custom text variable.
func (p *Package) SetCustomTextVariable(name, contents string) (int, error) {
	doc, err := p.Document()
	if err != nil {
		return 0, err
	}
	v := doc.TextVariable(name)
	if v == nil {
		return 0, common.Errorf("idml", "set custom text variable", PathDesignmap, "no text variable %q", name)
	}
	if err := v.SetCustomText(contents); err != nil {
		return 0, err
	}
	p.markDirty(PathDesignmap)

	stories, err := p.Stories()
	if err != nil {
		return 0, err
	}
	paths := make([]string, 0, len(stories))
	for path := range stories {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	n := 0
	for _, path := range paths {
		changed := false
		for _, inst := range stories[path].TextVariableInstances() {
			if inst.AssociatedTextVariable != v.Self {
				continue
			}
			n++
			if inst.ResultText != contents {
				inst.ResultText = contents
				changed = true
			}
		}
		if changed {
			p.markDirty(path)
		}
	}
	return n, nil
}
