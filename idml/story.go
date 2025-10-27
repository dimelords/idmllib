package idml

import (
	"encoding/xml"
	"strings"

	"github.com/dimelords/idmllib/types"
)

// readStory reads and parses a single story file
func (p *Package) readStory(data []byte) (types.Story, error) {
	// Parse the outer wrapper to get the inner Story
	var idPkgStory types.IDPkgStory
	if err := xml.Unmarshal(data, &idPkgStory); err != nil {
		return types.Story{}, &ParseError{FileName: "Story", Err: err}
	}

	return idPkgStory.Story, nil
}

// GetStory retrieves a specific story by ID
func (p *Package) GetStory(id string) (*types.Story, error) {
	for i := range p.Stories {
		if p.Stories[i].Self == id {
			return &p.Stories[i], nil
		}
	}
	return nil, &StoryNotFoundError{StoryID: id}
}

func isStory(name string) bool {
	return strings.HasPrefix(name, "Stories/Story_") && strings.HasSuffix(name, ".xml")
}
