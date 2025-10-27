//revive:disable:var-naming
package types

import "encoding/xml"

// Page represents a page in an InDesign spread
type Page struct {
	Self                   string              `xml:"Self,attr"`
	AppliedAlternateLayout string              `xml:"AppliedAlternateLayout,attr"`
	OverrideList           string              `xml:"OverrideList,attr"`
	GeometricBounds        string              `xml:"GeometricBounds,attr"`
	ItemTransform          string              `xml:"ItemTransform,attr"`
	Name                   string              `xml:"Name,attr"`
	AppliedTrapPreset      string              `xml:"AppliedTrapPreset,attr"`
	TabOrder               string              `xml:"TabOrder,attr"`
	GridStartingPoint      string              `xml:"GridStartingPoint,attr"`
	UseMasterGrid          string              `xml:"UseMasterGrid,attr"`
	AppliedMaster          string              `xml:"AppliedMaster,attr"`
	MasterPageTransform    string              `xml:"MasterPageTransform,attr"`
	SnapshotBlendingMode   string              `xml:"SnapshotBlendingMode,attr"`
	PageColor              string              `xml:"PageColor,attr"`
	UsePrimaryTextFrame    string              `xml:"UsePrimaryTextFrame,attr"`
	LayoutRule             string              `xml:"LayoutRule,attr"`
	Properties             PageProperties      `xml:"Properties"`
	MarginPreference       MarginPreference    `xml:"MarginPreference"`
	GridDataInformation    GridDataInformation `xml:"GridDataInformation"`
}

// PageProperties contains property elements for a page
type PageProperties struct {
	Applied          string     `xml:"Applied"`
	CustomLayoutType Descriptor `xml:"CustomLayoutType>EnumValue>Descriptor"`
}

// ValueWithType represents a value with a type attribute
type ValueWithType struct {
	Type  string `xml:"type,attr"`
	Value string `xml:",chardata"`
}

// Descriptor represents a descriptor with name, type, and value
type Descriptor struct {
	Name  string `xml:"Name,attr"`
	Type  string `xml:"type,attr"`
	Value string `xml:",chardata"`
}

// ListItem represents an item in a list
type ListItem struct {
	Type  string `xml:"type,attr"`
	Value string `xml:",chardata"`
}

// Guide represents a guide on a page
type Guide struct {
	XMLName       xml.Name        `xml:"Guide"`
	Self          string          `xml:"Self,attr"`
	ItemLayer     string          `xml:"ItemLayer,attr"`
	Locked        string          `xml:"Locked,attr"`
	GuideType     string          `xml:"GuideType,attr"`
	Location      string          `xml:"Location,attr"`
	Orientation   string          `xml:"Orientation,attr"`
	GuideZone     string          `xml:"GuideZone,attr"`
	FitToPage     string          `xml:"FitToPage,attr"`
	ViewThreshold string          `xml:"ViewThreshold,attr"`
	GuideColor    string          `xml:"GuideColor,attr"`
	Properties    GuideProperties `xml:"Properties"`
}

// GuideProperties contains property elements for a guide
type GuideProperties struct {
	Label Label `xml:"Label"`
}

// MarginPreference contains margin preferences for a page
type MarginPreference struct {
	ColumnCount      string `xml:"ColumnCount,attr"`
	ColumnGutter     string `xml:"ColumnGutter,attr"`
	Top              string `xml:"Top,attr"`
	Bottom           string `xml:"Bottom,attr"`
	Left             string `xml:"Left,attr"`
	Right            string `xml:"Right,attr"`
	ColumnDirection  string `xml:"ColumnDirection,attr"`
	ColumnsPositions string `xml:"ColumnsPositions,attr"`
}

// GridDataInformation contains grid data information for a page
type GridDataInformation struct {
	FontStyle              string         `xml:"FontStyle,attr"`
	PointSize              string         `xml:"PointSize,attr"`
	CharacterAki           string         `xml:"CharacterAki,attr"`
	LineAki                string         `xml:"LineAki,attr"`
	HorizontalScale        string         `xml:"HorizontalScale,attr"`
	VerticalScale          string         `xml:"VerticalScale,attr"`
	LineAlignment          string         `xml:"LineAlignment,attr"`
	GridAlignment          string         `xml:"GridAlignment,attr"`
	CharacterAlignment     string         `xml:"CharacterAlignment,attr"`
	GridView               string         `xml:"GridView,attr"`
	CharacterCountLocation string         `xml:"CharacterCountLocation,attr"`
	CharacterCountSize     string         `xml:"CharacterCountSize,attr"`
	Properties             GridProperties `xml:"Properties"`
}

// GridProperties contains property elements for grid data
type GridProperties struct {
	AppliedFont string `xml:"AppliedFont"`
}
