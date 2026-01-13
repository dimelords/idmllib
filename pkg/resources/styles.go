package resources

import (
	"encoding/xml"
	"strconv"

	"github.com/dimelords/idmllib/v2/pkg/common"
)

// StylesFile represents the Resources/Styles.xml file containing style definitions.
//
// The root element is <idPkg:Styles> with the idPkg namespace.
type StylesFile struct {
	// XMLName is not set directly - we handle it manually in MarshalXML/UnmarshalXML
	XMLName xml.Name `xml:"-"`

	// DOMVersion is the InDesign DOM version (e.g., "20.4")
	DOMVersion string `xml:"DOMVersion,attr"`

	// RootCharacterStyleGroup contains all character style definitions
	RootCharacterStyleGroup *CharacterStyleGroup `xml:"RootCharacterStyleGroup,omitempty"`

	// RootParagraphStyleGroup contains all paragraph style definitions
	RootParagraphStyleGroup *ParagraphStyleGroup `xml:"RootParagraphStyleGroup,omitempty"`

	// RootCellStyleGroup contains table cell style definitions
	RootCellStyleGroup *CellStyleGroup `xml:"RootCellStyleGroup,omitempty"`

	// RootTableStyleGroup contains table style definitions
	RootTableStyleGroup *TableStyleGroup `xml:"RootTableStyleGroup,omitempty"`

	// RootObjectStyleGroup contains object style definitions
	RootObjectStyleGroup *ObjectStyleGroup `xml:"RootObjectStyleGroup,omitempty"`

	// TOCStyles contain table of contents style definitions
	TOCStyles []TOCStyle `xml:"TOCStyle,omitempty"`

	// Catch-all for other elements
	OtherElements []common.RawXMLElement `xml:",any"`
}

// CharacterStyleGroup represents a group of character styles.
// Supports nested groups for hierarchical organization (e.g., "Naviga:Standard").
// XMLName will be "RootCharacterStyleGroup" for root or "CharacterStyleGroup" for nested.
type CharacterStyleGroup struct {
	XMLName         xml.Name               // Will be set during unmarshal
	Self            string                 `xml:"Self,attr"`
	Name            string                 `xml:"Name,attr,omitempty"`
	CharacterStyles []CharacterStyle       `xml:"CharacterStyle,omitempty"`
	NestedGroups    []CharacterStyleGroup  `xml:"CharacterStyleGroup,omitempty"` // Nested groups
	OtherElements   []common.RawXMLElement `xml:",any"`
}

// CharacterStyle represents a character style definition.
// Character styles apply formatting to selected text within a paragraph.
type CharacterStyle struct {
	Self                     string `xml:"Self,attr"`
	Name                     string `xml:"Name,attr"`
	Imported                 string `xml:"Imported,attr,omitempty"`                 // "true" or "false"
	SplitDocument            string `xml:"SplitDocument,attr,omitempty"`            // "true" or "false"
	EmitCss                  string `xml:"EmitCss,attr,omitempty"`                  // "true" or "false"
	StyleUniqueId            string `xml:"StyleUniqueId,attr,omitempty"`            // UUID
	IncludeClass             string `xml:"IncludeClass,attr,omitempty"`             // "true" or "false"
	ExtendedKeyboardShortcut string `xml:"ExtendedKeyboardShortcut,attr,omitempty"` // Keyboard shortcut

	// Basic text formatting attributes
	FontStyle   string `xml:"FontStyle,attr,omitempty"`   // Font style name
	PointSize   string `xml:"PointSize,attr,omitempty"`   // Font size
	FillColor   string `xml:"FillColor,attr,omitempty"`   // Fill color reference
	StrokeColor string `xml:"StrokeColor,attr,omitempty"` // Stroke color reference
	Underline   string `xml:"Underline,attr,omitempty"`   // "true" or "false"
	StrikeThru  string `xml:"StrikeThru,attr,omitempty"`  // "true" or "false"

	// Properties contain additional style settings
	Properties *common.Properties `xml:"Properties,omitempty"`

	// Catch-all for all other attributes and elements
	OtherElements []common.RawXMLElement `xml:",any"`
}

// GetAppliedFont returns the applied font from Properties, or empty string if not set.
// This is a convenience method to access the font without checking Properties nil.
func (cs *CharacterStyle) GetAppliedFont() string {
	if cs.Properties == nil {
		return ""
	}
	return cs.Properties.GetAppliedFont()
}

// GetPointSize returns the point size as a float64, or 0 if not set or invalid.
// This is a convenience method that handles the string-to-float conversion.
func (cs *CharacterStyle) GetPointSize() float64 {
	if cs.PointSize == "" {
		return 0
	}
	size, _ := strconv.ParseFloat(cs.PointSize, 64)
	return size
}

// ParagraphStyleGroup represents a group of paragraph styles.
// Supports nested groups for hierarchical organization (e.g., "Naviga:Standard").
// XMLName will be "RootParagraphStyleGroup" for root or "ParagraphStyleGroup" for nested.
type ParagraphStyleGroup struct {
	XMLName         xml.Name               // Will be set during unmarshal
	Self            string                 `xml:"Self,attr"`
	Name            string                 `xml:"Name,attr,omitempty"`
	ParagraphStyles []ParagraphStyle       `xml:"ParagraphStyle,omitempty"`
	NestedGroups    []ParagraphStyleGroup  `xml:"ParagraphStyleGroup,omitempty"` // Nested groups
	OtherElements   []common.RawXMLElement `xml:",any"`
}

// ParagraphStyle represents a paragraph style definition.
// Paragraph styles apply comprehensive formatting to entire paragraphs.
// Note: Due to the extremely large number of attributes (100+), we use OtherElements
// to capture all attributes while explicitly modeling only the most critical ones.
type ParagraphStyle struct {
	Self                     string `xml:"Self,attr"`
	Name                     string `xml:"Name,attr"`
	Imported                 string `xml:"Imported,attr,omitempty"`
	NextStyle                string `xml:"NextStyle,attr,omitempty"` // Reference to next paragraph style
	SplitDocument            string `xml:"SplitDocument,attr,omitempty"`
	EmitCss                  string `xml:"EmitCss,attr,omitempty"`
	StyleUniqueId            string `xml:"StyleUniqueId,attr,omitempty"`
	IncludeClass             string `xml:"IncludeClass,attr,omitempty"`
	ExtendedKeyboardShortcut string `xml:"ExtendedKeyboardShortcut,attr,omitempty"`
	KeyboardShortcut         string `xml:"KeyboardShortcut,attr,omitempty"`

	// Common paragraph formatting
	FontStyle       string `xml:"FontStyle,attr,omitempty"`
	PointSize       string `xml:"PointSize,attr,omitempty"`
	FillColor       string `xml:"FillColor,attr,omitempty"`
	Justification   string `xml:"Justification,attr,omitempty"` // "LeftAlign", "CenterAlign", "RightAlign", etc.
	SpaceBefore     string `xml:"SpaceBefore,attr,omitempty"`
	SpaceAfter      string `xml:"SpaceAfter,attr,omitempty"`
	LeftIndent      string `xml:"LeftIndent,attr,omitempty"`
	RightIndent     string `xml:"RightIndent,attr,omitempty"`
	FirstLineIndent string `xml:"FirstLineIndent,attr,omitempty"`

	// Text adjustment parameters (for accurate typography calculations)
	Tracking            string `xml:"Tracking,attr,omitempty"`            // Letter spacing (-25 = tight, 0 = normal, 25 = loose)
	KerningMethod       string `xml:"KerningMethod,attr,omitempty"`       // "$ID/Optical" or "$ID/Metrics"
	MinimumWordSpacing  string `xml:"MinimumWordSpacing,attr,omitempty"`  // Percentage (90 = 90%)
	MaximumWordSpacing  string `xml:"MaximumWordSpacing,attr,omitempty"`  // Percentage (110 = 110%)
	MinimumGlyphScaling string `xml:"MinimumGlyphScaling,attr,omitempty"` // Percentage (97 = 97%)
	MaximumGlyphScaling string `xml:"MaximumGlyphScaling,attr,omitempty"` // Percentage (103 = 103%)

	// Properties contain additional style settings (AppliedFont, Leading, TabList, etc.)
	Properties *common.Properties `xml:"Properties,omitempty"`

	// Catch-all for the many other attributes (100+ attributes total)
	OtherElements []common.RawXMLElement `xml:",any"`
}

// CellStyleGroup represents a group of table cell styles.
type CellStyleGroup struct {
	XMLName       xml.Name               `xml:"RootCellStyleGroup"`
	Self          string                 `xml:"Self,attr"`
	CellStyles    []CellStyle            `xml:"CellStyle,omitempty"`
	OtherElements []common.RawXMLElement `xml:",any"`
}

// CellStyle represents a table cell style definition.
type CellStyle struct {
	Self                  string                 `xml:"Self,attr"`
	Name                  string                 `xml:"Name,attr"`
	AppliedParagraphStyle string                 `xml:"AppliedParagraphStyle,attr,omitempty"` // Reference to paragraph style
	Properties            *common.Properties     `xml:"Properties,omitempty"`
	OtherElements         []common.RawXMLElement `xml:",any"`
}

// TableStyleGroup represents a group of table styles.
type TableStyleGroup struct {
	XMLName       xml.Name               `xml:"RootTableStyleGroup"`
	Self          string                 `xml:"Self,attr"`
	TableStyles   []TableStyle           `xml:"TableStyle,omitempty"`
	OtherElements []common.RawXMLElement `xml:",any"`
}

// TableStyle represents a table style definition.
// Contains comprehensive table border, fill, and stroke settings.
type TableStyle struct {
	Self                     string `xml:"Self,attr"`
	Name                     string `xml:"Name,attr"`
	ExtendedKeyboardShortcut string `xml:"ExtendedKeyboardShortcut,attr,omitempty"`
	KeyboardShortcut         string `xml:"KeyboardShortcut,attr,omitempty"`

	// Table spacing
	SpaceBefore string `xml:"SpaceBefore,attr,omitempty"`
	SpaceAfter  string `xml:"SpaceAfter,attr,omitempty"`

	// Border properties (Top, Left, Bottom, Right)
	TopBorderStrokeWeight  string `xml:"TopBorderStrokeWeight,attr,omitempty"`
	TopBorderStrokeColor   string `xml:"TopBorderStrokeColor,attr,omitempty"`
	LeftBorderStrokeWeight string `xml:"LeftBorderStrokeWeight,attr,omitempty"`
	LeftBorderStrokeColor  string `xml:"LeftBorderStrokeColor,attr,omitempty"`
	// ... many more border attributes

	Properties    *common.Properties     `xml:"Properties,omitempty"`
	OtherElements []common.RawXMLElement `xml:",any"`
}

// ObjectStyleGroup represents a group of object styles.
// Supports nested groups for hierarchical organization.
// XMLName will be "RootObjectStyleGroup" for root or "ObjectStyleGroup" for nested.
type ObjectStyleGroup struct {
	XMLName       xml.Name               // Will be set during unmarshal
	Self          string                 `xml:"Self,attr"`
	Name          string                 `xml:"Name,attr,omitempty"`
	ObjectStyles  []ObjectStyle          `xml:"ObjectStyle,omitempty"`
	NestedGroups  []ObjectStyleGroup     `xml:"ObjectStyleGroup,omitempty"` // Nested groups
	OtherElements []common.RawXMLElement `xml:",any"`
}

// ObjectStyle represents an object style definition.
// Applies to frames, text boxes, graphics, and other page objects.
type ObjectStyle struct {
	Self                     string `xml:"Self,attr"`
	Name                     string `xml:"Name,attr"`
	ExtendedKeyboardShortcut string `xml:"ExtendedKeyboardShortcut,attr,omitempty"`
	KeyboardShortcut         string `xml:"KeyboardShortcut,attr,omitempty"`
	AppliedParagraphStyle    string `xml:"AppliedParagraphStyle,attr,omitempty"`
	EmitCss                  string `xml:"EmitCss,attr,omitempty"`
	IncludeClass             string `xml:"IncludeClass,attr,omitempty"`

	// Stroke and fill properties
	FillColor    string `xml:"FillColor,attr,omitempty"`
	FillTint     string `xml:"FillTint,attr,omitempty"`
	StrokeColor  string `xml:"StrokeColor,attr,omitempty"`
	StrokeTint   string `xml:"StrokeTint,attr,omitempty"`
	StrokeWeight string `xml:"StrokeWeight,attr,omitempty"`

	// Corner properties
	TopLeftCornerOption string `xml:"TopLeftCornerOption,attr,omitempty"`
	TopLeftCornerRadius string `xml:"TopLeftCornerRadius,attr,omitempty"`
	CornerRadius        string `xml:"CornerRadius,attr,omitempty"`

	// Child elements
	TransformAttributeOption *TransformAttributeOption `xml:"TransformAttributeOption,omitempty"`
	ObjectExportOption       *ObjectExportOption       `xml:"ObjectExportOption,omitempty"`
	TextFramePreference      *TextFramePreference      `xml:"TextFramePreference,omitempty"`

	Properties    *common.Properties     `xml:"Properties,omitempty"`
	OtherElements []common.RawXMLElement `xml:",any"`
}

// TransformAttributeOption defines transform reference points for objects.
type TransformAttributeOption struct {
	TransformAttrLeftReference  string `xml:"TransformAttrLeftReference,attr,omitempty"`
	TransformAttrTopReference   string `xml:"TransformAttrTopReference,attr,omitempty"`
	TransformAttrRefAnchorPoint string `xml:"TransformAttrRefAnchorPoint,attr,omitempty"`
}

// ObjectExportOption defines object export settings.
type ObjectExportOption struct {
	AltTextSourceType     string                 `xml:"AltTextSourceType,attr,omitempty"`
	ActualTextSourceType  string                 `xml:"ActualTextSourceType,attr,omitempty"`
	CustomAltText         string                 `xml:"CustomAltText,attr,omitempty"`
	CustomActualText      string                 `xml:"CustomActualText,attr,omitempty"`
	ApplyTagType          string                 `xml:"ApplyTagType,attr,omitempty"`
	ImageConversionType   string                 `xml:"ImageConversionType,attr,omitempty"`
	ImageExportResolution string                 `xml:"ImageExportResolution,attr,omitempty"`
	Properties            *common.Properties     `xml:"Properties,omitempty"`
	OtherElements         []common.RawXMLElement `xml:",any"`
}

// TextFramePreference defines text frame preferences.
type TextFramePreference struct {
	TextColumnCount       string                 `xml:"TextColumnCount,attr,omitempty"`
	TextColumnGutter      string                 `xml:"TextColumnGutter,attr,omitempty"`
	FirstBaselineOffset   string                 `xml:"FirstBaselineOffset,attr,omitempty"`
	VerticalJustification string                 `xml:"VerticalJustification,attr,omitempty"`
	AutoSizingType        string                 `xml:"AutoSizingType,attr,omitempty"`
	OtherElements         []common.RawXMLElement `xml:",any"`
}

// TOCStyle represents a table of contents style definition.
type TOCStyle struct {
	Self                 string                 `xml:"Self,attr"`
	Name                 string                 `xml:"Name,attr"`
	Title                string                 `xml:"Title,attr,omitempty"`
	TitleStyle           string                 `xml:"TitleStyle,attr,omitempty"`           // Reference to paragraph style
	RunIn                string                 `xml:"RunIn,attr,omitempty"`                // "true" or "false"
	IncludeHidden        string                 `xml:"IncludeHidden,attr,omitempty"`        // "true" or "false"
	IncludeBookDocuments string                 `xml:"IncludeBookDocuments,attr,omitempty"` // "true" or "false"
	CreateBookmarks      string                 `xml:"CreateBookmarks,attr,omitempty"`      // "true" or "false"
	OtherElements        []common.RawXMLElement `xml:",any"`
}

// FindParagraphStyle finds a paragraph style by its Self ID.
// It searches through the paragraph style group hierarchy, including nested groups.
func (sf *StylesFile) FindParagraphStyle(styleID string) *ParagraphStyle {
	if sf.RootParagraphStyleGroup == nil {
		return nil
	}
	return sf.findParagraphStyleInGroup(sf.RootParagraphStyleGroup, styleID)
}

// findParagraphStyleInGroup recursively searches for a paragraph style in a group.
func (sf *StylesFile) findParagraphStyleInGroup(group *ParagraphStyleGroup, styleID string) *ParagraphStyle {
	if group == nil {
		return nil
	}

	// Search in direct styles
	for i := range group.ParagraphStyles {
		if group.ParagraphStyles[i].Self == styleID {
			return &group.ParagraphStyles[i]
		}
	}

	// Search in nested groups
	for i := range group.NestedGroups {
		if style := sf.findParagraphStyleInGroup(&group.NestedGroups[i], styleID); style != nil {
			return style
		}
	}

	return nil
}
