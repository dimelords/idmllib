// Package types provides type definitions for InDesign IDML and IDMS files.
package types

import (
	"encoding/xml"
	"regexp"
)

// IDMS represents a complete InDesign Snippet XML structure
type IDMS struct {
	XMLName                            xml.Name                            `xml:"Document"`
	DOMVersion                         string                              `xml:"DOMVersion,attr"`
	Self                               string                              `xml:"Self,attr"`
	Colors                             []Color                             `xml:"Color,omitempty"`
	Swatches                           []Swatch                            `xml:"Swatch,omitempty"`
	StrokeStyles                       []StrokeStyle                       `xml:"StrokeStyle,omitempty"`
	RootCharacterStyleGroup            *RootCharacterStyleGroup            `xml:"RootCharacterStyleGroup,omitempty"`
	RootParagraphStyleGroup            *RootParagraphStyleGroup            `xml:"RootParagraphStyleGroup,omitempty"`
	RootObjectStyleGroup               *RootObjectStyleGroup               `xml:"RootObjectStyleGroup,omitempty"`
	TinDocumentDataObject              *TinDocumentDataObject              `xml:"TinDocumentDataObject,omitempty"`
	TransparencyDefaultContainerObject *TransparencyDefaultContainerObject `xml:"TransparencyDefaultContainerObject,omitempty"`
	Layers                             []Layer                             `xml:"Layer,omitempty"`
	Spreads                            []Spread                            `xml:"Spread,omitempty"`
	Stories                            []Story                             `xml:"Story,omitempty"`
	ColorGroups                        []ColorGroup                        `xml:"ColorGroup,omitempty"`
}

// Color represents a color definition
type Color struct {
	Self                      string `xml:"Self,attr"`
	Model                     string `xml:"Model,attr,omitempty"`
	Space                     string `xml:"Space,attr,omitempty"`
	ColorValue                string `xml:"ColorValue,attr,omitempty"`
	ColorOverride             string `xml:"ColorOverride,attr,omitempty"`
	ConvertToHsb              string `xml:"ConvertToHsb,attr,omitempty"`
	AlternateSpace            string `xml:"AlternateSpace,attr,omitempty"`
	AlternateColorValue       string `xml:"AlternateColorValue,attr,omitempty"`
	Name                      string `xml:"Name,attr"`
	ColorEditable             string `xml:"ColorEditable,attr,omitempty"`
	ColorRemovable            string `xml:"ColorRemovable,attr,omitempty"`
	Visible                   string `xml:"Visible,attr,omitempty"`
	SwatchCreatorID           string `xml:"SwatchCreatorID,attr,omitempty"`
	SwatchColorGroupReference string `xml:"SwatchColorGroupReference,attr,omitempty"`
}

// Swatch represents a swatch definition
type Swatch struct {
	Self                      string `xml:"Self,attr"`
	Name                      string `xml:"Name,attr"`
	ColorEditable             string `xml:"ColorEditable,attr,omitempty"`
	ColorRemovable            string `xml:"ColorRemovable,attr,omitempty"`
	Visible                   string `xml:"Visible,attr,omitempty"`
	SwatchCreatorID           string `xml:"SwatchCreatorID,attr,omitempty"`
	SwatchColorGroupReference string `xml:"SwatchColorGroupReference,attr,omitempty"`
}

// StrokeStyle represents a stroke style
type StrokeStyle struct {
	Self string `xml:"Self,attr"`
	Name string `xml:"Name,attr"`
}

// RootCharacterStyleGroup represents the root character style group
type RootCharacterStyleGroup struct {
	Self                 string                `xml:"Self,attr"`
	CharacterStyleGroups []CharacterStyleGroup `xml:"CharacterStyleGroup,omitempty"`
}

// CharacterStyleGroup represents a group of character styles
type CharacterStyleGroup struct {
	Self            string           `xml:"Self,attr"`
	Name            string           `xml:"Name,attr"`
	CharacterStyles []CharacterStyle `xml:"CharacterStyle,omitempty"`
}

// CharacterStyle represents a character style
type CharacterStyle struct {
	Self                     string      `xml:"Self,attr"`
	Imported                 string      `xml:"Imported,attr,omitempty"`
	SplitDocument            string      `xml:"SplitDocument,attr,omitempty"`
	EmitCss                  string      `xml:"EmitCss,attr,omitempty"`       //nolint:revive // XML field must match InDesign schema
	StyleUniqueId            string      `xml:"StyleUniqueId,attr,omitempty"` //nolint:revive // XML field must match InDesign schema
	IncludeClass             string      `xml:"IncludeClass,attr,omitempty"`
	ExtendedKeyboardShortcut string      `xml:"ExtendedKeyboardShortcut,attr,omitempty"`
	KeyboardShortcut         string      `xml:"KeyboardShortcut,attr,omitempty"`
	Name                     string      `xml:"Name,attr"`
	Properties               *Properties `xml:"Properties,omitempty"`
}

// RootParagraphStyleGroup represents the root paragraph style group
type RootParagraphStyleGroup struct {
	Self                 string                `xml:"Self,attr"`
	ParagraphStyleGroups []ParagraphStyleGroup `xml:"ParagraphStyleGroup,omitempty"`
}

// ParagraphStyleGroup represents a group of paragraph styles
type ParagraphStyleGroup struct {
	Self            string                `xml:"Self,attr"`
	Name            string                `xml:"Name,attr"`
	ParagraphStyles []ParagraphStyle      `xml:"ParagraphStyle,omitempty"`
	SubGroups       []ParagraphStyleGroup `xml:"ParagraphStyleGroup,omitempty"`
}

// ParagraphStyle represents a paragraph style
type ParagraphStyle struct {
	Self                     string      `xml:"Self,attr"`
	Name                     string      `xml:"Name,attr"`
	Imported                 string      `xml:"Imported,attr,omitempty"`
	NextStyle                string      `xml:"NextStyle,attr,omitempty"`
	SplitDocument            string      `xml:"SplitDocument,attr,omitempty"`
	EmitCss                  string      `xml:"EmitCss,attr,omitempty"`       //nolint:revive // XML field must match InDesign schema
	StyleUniqueId            string      `xml:"StyleUniqueId,attr,omitempty"` //nolint:revive // XML field must match InDesign schema
	IncludeClass             string      `xml:"IncludeClass,attr,omitempty"`
	ExtendedKeyboardShortcut string      `xml:"ExtendedKeyboardShortcut,attr,omitempty"`
	EmptyNestedStyles        string      `xml:"EmptyNestedStyles,attr,omitempty"`
	EmptyLineStyles          string      `xml:"EmptyLineStyles,attr,omitempty"`
	EmptyGrepStyles          string      `xml:"EmptyGrepStyles,attr,omitempty"`
	KeyboardShortcut         string      `xml:"KeyboardShortcut,attr,omitempty"`
	FontStyle                string      `xml:"FontStyle,attr,omitempty"`
	PointSize                string      `xml:"PointSize,attr,omitempty"`
	KerningMethod            string      `xml:"KerningMethod,attr,omitempty"`
	Tracking                 string      `xml:"Tracking,attr,omitempty"`
	HyphenateLadderLimit     string      `xml:"HyphenateLadderLimit,attr,omitempty"`
	AppliedLanguage          string      `xml:"AppliedLanguage,attr,omitempty"`
	HyphenateAfterFirst      string      `xml:"HyphenateAfterFirst,attr,omitempty"`
	HyphenateBeforeLast      string      `xml:"HyphenateBeforeLast,attr,omitempty"`
	HyphenateWordsLongerThan string      `xml:"HyphenateWordsLongerThan,attr,omitempty"`
	MaximumWordSpacing       string      `xml:"MaximumWordSpacing,attr,omitempty"`
	MinimumWordSpacing       string      `xml:"MinimumWordSpacing,attr,omitempty"`
	MaximumGlyphScaling      string      `xml:"MaximumGlyphScaling,attr,omitempty"`
	MinimumGlyphScaling      string      `xml:"MinimumGlyphScaling,attr,omitempty"`
	HyphenateAcrossColumns   string      `xml:"HyphenateAcrossColumns,attr,omitempty"`
	Properties               *Properties `xml:"Properties,omitempty"`
}

// RootObjectStyleGroup represents the root object style group
type RootObjectStyleGroup struct {
	Self              string             `xml:"Self,attr"`
	ObjectStyleGroups []ObjectStyleGroup `xml:"ObjectStyleGroup,omitempty"`
}

// ObjectStyleGroup represents a group of object styles
type ObjectStyleGroup struct {
	Self         string             `xml:"Self,attr"`
	Name         string             `xml:"Name,attr"`
	ObjectStyles []ObjectStyle      `xml:"ObjectStyle,omitempty"`
	SubGroups    []ObjectStyleGroup `xml:"ObjectStyleGroup,omitempty"`
}

// ObjectStyle represents an object style with all its complex attributes
type ObjectStyle struct {
	Self                               string      `xml:"Self,attr"`
	Name                               string      `xml:"Name,attr"`
	EnableTransformAttributes          string      `xml:"EnableTransformAttributes,attr,omitempty"`
	TopLeftCornerOption                string      `xml:"TopLeftCornerOption,attr,omitempty"`
	TopRightCornerOption               string      `xml:"TopRightCornerOption,attr,omitempty"`
	BottomLeftCornerOption             string      `xml:"BottomLeftCornerOption,attr,omitempty"`
	BottomRightCornerOption            string      `xml:"BottomRightCornerOption,attr,omitempty"`
	TopLeftCornerRadius                string      `xml:"TopLeftCornerRadius,attr,omitempty"`
	TopRightCornerRadius               string      `xml:"TopRightCornerRadius,attr,omitempty"`
	BottomLeftCornerRadius             string      `xml:"BottomLeftCornerRadius,attr,omitempty"`
	BottomRightCornerRadius            string      `xml:"BottomRightCornerRadius,attr,omitempty"`
	EmitCss                            string      `xml:"EmitCss,attr,omitempty"` //nolint:revive // XML field must match InDesign schema
	IncludeClass                       string      `xml:"IncludeClass,attr,omitempty"`
	EnableTextFrameAutoSizingOptions   string      `xml:"EnableTextFrameAutoSizingOptions,attr,omitempty"`
	ExtendedKeyboardShortcut           string      `xml:"ExtendedKeyboardShortcut,attr,omitempty"`
	EnableTextFrameColumnRuleOptions   string      `xml:"EnableTextFrameColumnRuleOptions,attr,omitempty"`
	EnableExportTagging                string      `xml:"EnableExportTagging,attr,omitempty"`
	EnableObjectExportAltTextOptions   string      `xml:"EnableObjectExportAltTextOptions,attr,omitempty"`
	EnableObjectExportTaggedPdfOptions string      `xml:"EnableObjectExportTaggedPdfOptions,attr,omitempty"`
	EnableObjectExportEpubOptions      string      `xml:"EnableObjectExportEpubOptions,attr,omitempty"`
	AppliedParagraphStyle              string      `xml:"AppliedParagraphStyle,attr,omitempty"`
	ApplyNextParagraphStyle            string      `xml:"ApplyNextParagraphStyle,attr,omitempty"`
	EnableFill                         string      `xml:"EnableFill,attr,omitempty"`
	EnableStroke                       string      `xml:"EnableStroke,attr,omitempty"`
	EnableParagraphStyle               string      `xml:"EnableParagraphStyle,attr,omitempty"`
	EnableTextFrameGeneralOptions      string      `xml:"EnableTextFrameGeneralOptions,attr,omitempty"`
	EnableTextFrameBaselineOptions     string      `xml:"EnableTextFrameBaselineOptions,attr,omitempty"`
	EnableStoryOptions                 string      `xml:"EnableStoryOptions,attr,omitempty"`
	EnableTextWrapAndOthers            string      `xml:"EnableTextWrapAndOthers,attr,omitempty"`
	EnableAnchoredObjectOptions        string      `xml:"EnableAnchoredObjectOptions,attr,omitempty"`
	CornerRadius                       string      `xml:"CornerRadius,attr,omitempty"`
	FillColor                          string      `xml:"FillColor,attr,omitempty"`
	FillTint                           string      `xml:"FillTint,attr,omitempty"`
	StrokeWeight                       string      `xml:"StrokeWeight,attr,omitempty"`
	MiterLimit                         string      `xml:"MiterLimit,attr,omitempty"`
	EndCap                             string      `xml:"EndCap,attr,omitempty"`
	EndJoin                            string      `xml:"EndJoin,attr,omitempty"`
	StrokeType                         string      `xml:"StrokeType,attr,omitempty"`
	LeftLineEnd                        string      `xml:"LeftLineEnd,attr,omitempty"`
	RightLineEnd                       string      `xml:"RightLineEnd,attr,omitempty"`
	StrokeColor                        string      `xml:"StrokeColor,attr,omitempty"`
	StrokeTint                         string      `xml:"StrokeTint,attr,omitempty"`
	GapColor                           string      `xml:"GapColor,attr,omitempty"`
	GapTint                            string      `xml:"GapTint,attr,omitempty"`
	StrokeAlignment                    string      `xml:"StrokeAlignment,attr,omitempty"`
	Nonprinting                        string      `xml:"Nonprinting,attr,omitempty"`
	GradientFillAngle                  string      `xml:"GradientFillAngle,attr,omitempty"`
	GradientStrokeAngle                string      `xml:"GradientStrokeAngle,attr,omitempty"`
	AppliedNamedGrid                   string      `xml:"AppliedNamedGrid,attr,omitempty"`
	KeyboardShortcut                   string      `xml:"KeyboardShortcut,attr,omitempty"`
	EnableFrameFittingOptions          string      `xml:"EnableFrameFittingOptions,attr,omitempty"`
	CornerOption                       string      `xml:"CornerOption,attr,omitempty"`
	EnableStrokeAndCornerOptions       string      `xml:"EnableStrokeAndCornerOptions,attr,omitempty"`
	ArrowHeadAlignment                 string      `xml:"ArrowHeadAlignment,attr,omitempty"`
	LeftArrowHeadScale                 string      `xml:"LeftArrowHeadScale,attr,omitempty"`
	RightArrowHeadScale                string      `xml:"RightArrowHeadScale,attr,omitempty"`
	EnableTextFrameFootnoteOptions     string      `xml:"EnableTextFrameFootnoteOptions,attr,omitempty"`
	Properties                         *Properties `xml:"Properties,omitempty"`
	InnerXML                           string      `xml:",innerxml"`
}

// Properties is a generic container for nested properties
type Properties struct {
	InnerXML string `xml:",innerxml"`
}

// TinDocumentDataObject represents document data
type TinDocumentDataObject struct {
	Properties *Properties `xml:"Properties,omitempty"`
}

// TransparencyDefaultContainerObject contains transparency settings
type TransparencyDefaultContainerObject struct {
	InnerXML string `xml:",innerxml"`
}

// TextFrame represents a text frame with all its attributes

// StoryElement represents a story
type StoryElement struct {
	Self             string `xml:"Self,attr"`
	UserText         string `xml:"UserText,attr,omitempty"`
	IsEndnoteStory   string `xml:"IsEndnoteStory,attr,omitempty"`
	AppliedTOCStyle  string `xml:"AppliedTOCStyle,attr,omitempty"`
	TrackChanges     string `xml:"TrackChanges,attr,omitempty"`
	StoryTitle       string `xml:"StoryTitle,attr,omitempty"`
	AppliedNamedGrid string `xml:"AppliedNamedGrid,attr,omitempty"`
	InnerXML         string `xml:",innerxml"`
}

// ColorGroup represents a color group
type ColorGroup struct {
	Self               string             `xml:"Self,attr"`
	Name               string             `xml:"Name,attr"`
	IsRootColorGroup   string             `xml:"IsRootColorGroup,attr,omitempty"`
	ColorGroupSwatches []ColorGroupSwatch `xml:"ColorGroupSwatch,omitempty"`
}

// ColorGroupSwatch represents a swatch reference in a color group
type ColorGroupSwatch struct {
	Self          string `xml:"Self,attr"`
	SwatchItemRef string `xml:"SwatchItemRef,attr"`
}

// GetSelf returns the style's Self identifier
func (cs CharacterStyle) GetSelf() string {
	return cs.Self
}

// GetBasedOn extracts the BasedOn reference from Properties.InnerXML
func (cs CharacterStyle) GetBasedOn() string {
	if cs.Properties != nil && cs.Properties.InnerXML != "" {
		basedOnPattern := regexp.MustCompile(`<BasedOn[^>]*>([^<]+)</BasedOn>`)
		if matches := basedOnPattern.FindStringSubmatch(cs.Properties.InnerXML); len(matches) > 1 {
			return matches[1]
		}
	}
	return ""
}

// GetNextStyle returns empty string as CharacterStyle doesn't have NextStyle
func (cs CharacterStyle) GetNextStyle() string {
	return ""
}

// IsSystemStyle checks if this is a system/default style
func (cs CharacterStyle) IsSystemStyle(basedOn string) bool {
	return basedOn == "n" || basedOn == "$ID/[No character style]"
}

// GetSelf returns the style's Self identifier
func (ps ParagraphStyle) GetSelf() string {
	return ps.Self
}

// GetBasedOn extracts the BasedOn reference from Properties.InnerXML
func (ps ParagraphStyle) GetBasedOn() string {
	if ps.Properties != nil && ps.Properties.InnerXML != "" {
		basedOnPattern := regexp.MustCompile(`<BasedOn[^>]*>([^<]+)</BasedOn>`)
		if matches := basedOnPattern.FindStringSubmatch(ps.Properties.InnerXML); len(matches) > 1 {
			return matches[1]
		}
	}
	return ""
}

// GetNextStyle returns the NextStyle attribute
func (ps ParagraphStyle) GetNextStyle() string {
	return ps.NextStyle
}

// IsSystemStyle checks if this is a system/default style
func (ps ParagraphStyle) IsSystemStyle(ref string) bool {
	return ref == "n" || ref == "$ID/[No paragraph style]"
}

// GetSelf returns the style's Self identifier
func (os ObjectStyle) GetSelf() string {
	return os.Self
}

// GetBasedOn extracts the BasedOn reference from InnerXML
func (os ObjectStyle) GetBasedOn() string {
	if os.InnerXML != "" {
		basedOnPattern := regexp.MustCompile(`<BasedOn[^>]*>([^<]+)</BasedOn>`)
		if matches := basedOnPattern.FindStringSubmatch(os.InnerXML); len(matches) > 1 {
			return matches[1]
		}
	}
	return ""
}

// GetNextStyle returns empty string as ObjectStyle doesn't have NextStyle
func (os ObjectStyle) GetNextStyle() string {
	return ""
}

// IsSystemStyle checks if this is a system/default style
func (os ObjectStyle) IsSystemStyle(basedOn string) bool {
	return basedOn == "n" || basedOn == "$ID/[None]"
}
