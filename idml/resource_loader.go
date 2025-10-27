package idml

import (
	"encoding/xml"

	"github.com/dimelords/idmllib/types"
)

// loadStyles reads character and paragraph styles from Resources/Styles.xml
func (p *Package) loadStyles(idms *types.IDMS) error {
	stylesData, err := p.readFileFromIDML("Resources/Styles.xml")
	if err != nil {
		return &FileNotFoundError{FileName: "Resources/Styles.xml"}
	}

	// Parse the full styles document to get complete style hierarchies
	var doc struct {
		XMLName                 xml.Name                       `xml:"Styles"`
		RootCharacterStyleGroup *types.RootCharacterStyleGroup `xml:"RootCharacterStyleGroup"`
		RootParagraphStyleGroup *types.RootParagraphStyleGroup `xml:"RootParagraphStyleGroup"`
		RootObjectStyleGroup    *types.RootObjectStyleGroup    `xml:"RootObjectStyleGroup"`
	}

	if err := xml.Unmarshal(stylesData, &doc); err != nil {
		return &ParseError{FileName: "Resources/Styles.xml", Err: err}
	}

	// Assign the parsed style groups directly
	// This preserves the complete structure including all nested groups and styles
	if doc.RootCharacterStyleGroup != nil {
		idms.RootCharacterStyleGroup = doc.RootCharacterStyleGroup
	}

	if doc.RootParagraphStyleGroup != nil {
		idms.RootParagraphStyleGroup = doc.RootParagraphStyleGroup
	}

	if doc.RootObjectStyleGroup != nil {
		idms.RootObjectStyleGroup = doc.RootObjectStyleGroup
	}

	return nil
}

// loadColorsAndSwatches reads colors and swatches from Resources/Graphic.xml
func (p *Package) loadColorsAndSwatches(idms *types.IDMS) error {
	graphicData, err := p.readFileFromIDML("Resources/Graphic.xml")
	if err != nil {
		return &FileNotFoundError{FileName: "Resources/Graphic.xml"}
	}

	// Parse colors and swatches from Graphic.xml
	var doc struct {
		Colors   []types.Color  `xml:"Color"`
		Swatches []types.Swatch `xml:"Swatch"`
	}

	if err := xml.Unmarshal(graphicData, &doc); err != nil {
		return &ParseError{FileName: "Resources/Graphic.xml", Err: err}
	}

	idms.Colors = doc.Colors
	idms.Swatches = doc.Swatches

	return nil
}

// loadTransparencyDefaults loads transparency settings from Resources/Preferences.xml
func (p *Package) loadTransparencyDefaults(idms *types.IDMS) error {
	prefsData, err := p.readFileFromIDML("Resources/Preferences.xml")
	if err != nil {
		return &FileNotFoundError{FileName: "Resources/Preferences.xml"}
	}

	// Parse transparency container - store as raw XML since structure is complex
	var doc struct {
		TransparencyDefaultContainerObject struct {
			InnerXML string `xml:",innerxml"`
		} `xml:"TransparencyDefaultContainerObject"`
	}

	if err := xml.Unmarshal(prefsData, &doc); err != nil {
		return &ParseError{FileName: "Resources/Preferences.xml", Err: err}
	}

	idms.TransparencyDefaultContainerObject = &types.TransparencyDefaultContainerObject{
		InnerXML: doc.TransparencyDefaultContainerObject.InnerXML,
	}

	return nil
}

// loadColorGroups reads color groups from designmap.xml
func (p *Package) loadColorGroups(idms *types.IDMS) error {
	designmapData, err := p.readFileFromIDML("designmap.xml")
	if err != nil {
		return &FileNotFoundError{FileName: "designmap.xml"}
	}

	// Parse color groups from designmap
	var doc struct {
		ColorGroups []types.ColorGroup `xml:"ColorGroup"`
	}

	if err := xml.Unmarshal(designmapData, &doc); err != nil {
		return &ParseError{FileName: "designmap.xml", Err: err}
	}

	idms.ColorGroups = doc.ColorGroups

	return nil
}

// loadLayers extracts layer definitions from the IDML designmap.xml
func (p *Package) loadLayers(idms *types.IDMS) error {
	designmapData, err := p.readFileFromIDML("designmap.xml")
	if err != nil {
		return &FileNotFoundError{FileName: "designmap.xml"}
	}

	// Parse the designmap to find layer definitions
	// Layers are direct children of the Document element
	var designMapXML struct {
		Layers []types.Layer `xml:"Layer"`
	}

	if err := xml.Unmarshal(designmapData, &designMapXML); err != nil {
		return &ParseError{FileName: "designmap.xml", Err: err}
	}

	if len(designMapXML.Layers) > 0 {
		idms.Layers = designMapXML.Layers
	}

	return nil
}

// LoadStyles implements ResourceLoader interface
func (p *Package) LoadStyles(idms *types.IDMS) error {
	return p.loadStyles(idms)
}

// LoadColorsAndSwatches implements ResourceLoader interface
func (p *Package) LoadColorsAndSwatches(idms *types.IDMS) error {
	return p.loadColorsAndSwatches(idms)
}

// LoadLayers implements ResourceLoader interface
func (p *Package) LoadLayers(idms *types.IDMS) error {
	return p.loadLayers(idms)
}

// LoadColorGroups implements ResourceLoader interface
func (p *Package) LoadColorGroups(idms *types.IDMS) error {
	return p.loadColorGroups(idms)
}

// LoadTransparencyDefaults implements ResourceLoader interface
func (p *Package) LoadTransparencyDefaults(idms *types.IDMS) error {
	return p.loadTransparencyDefaults(idms)
}

// GetSpread implements Reader interface - retrieves spreads for a story
func (p *Package) GetSpread(storyID string) ([]types.Spread, error) {
	// First verify the story exists
	if _, err := p.GetStory(storyID); err != nil {
		return nil, err
	}

	// Get spreads for this story
	spreads, err := p.getSpreadsForStory(&types.Story{Self: storyID})
	if err != nil {
		return nil, &SpreadNotFoundError{StoryID: storyID}
	}

	return spreads, nil
}
