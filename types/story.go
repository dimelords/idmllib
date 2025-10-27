//revive:disable:var-naming
package types

import "encoding/xml"

// IDPkgStory represents the root element containing a Story
type IDPkgStory struct {
	Story Story `xml:"Story"`
}

// Story represents a text flow in InDesign
type Story struct {
	XMLName              xml.Name              `xml:"Story"`
	Self                 string                `xml:"Self,attr"`
	UserText             string                `xml:"UserText,attr"`
	IsEndnoteStory       string                `xml:"IsEndnoteStory,attr"`
	AppliedTOCStyle      string                `xml:"AppliedTOCStyle,attr"`
	TrackChanges         string                `xml:"TrackChanges,attr"`
	StoryTitle           string                `xml:"StoryTitle,attr"`
	AppliedNamedGrid     string                `xml:"AppliedNamedGrid,attr"`
	StoryPreference      StoryPreference       `xml:"StoryPreference"`
	InCopyExportOption   InCopyExportOption    `xml:"InCopyExportOption"`
	ParagraphStyleRanges []ParagraphStyleRange `xml:"ParagraphStyleRange"`
}

// StoryPreference contains preferences for a story
type StoryPreference struct {
	OpticalMarginAlignment string `xml:"OpticalMarginAlignment,attr"`
	OpticalMarginSize      string `xml:"OpticalMarginSize,attr"`
	FrameType              string `xml:"FrameType,attr"`
	StoryOrientation       string `xml:"StoryOrientation,attr"`
	StoryDirection         string `xml:"StoryDirection,attr"`
}

// ParagraphStyleRange represents a range of text with a specific paragraph style
type ParagraphStyleRange struct {
	AppliedParagraphStyle string                `xml:"AppliedParagraphStyle,attr"`
	CharacterStyleRanges  []CharacterStyleRange `xml:"CharacterStyleRange"`
}

// CharacterStyleRange represents a range of text with a specific character style
type CharacterStyleRange struct {
	AppliedCharacterStyle string `xml:"AppliedCharacterStyle,attr"`
	Content               string `xml:"Content"`
}
