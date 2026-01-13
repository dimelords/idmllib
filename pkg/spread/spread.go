package spread

import (
	"encoding/xml"

	"github.com/dimelords/idmllib/v2/pkg/common"
)

// PageItemBase contains the common attributes shared by all page items.
// This struct is embedded in specific page item types to reduce duplication
// and provide consistent interface implementation.
type PageItemBase struct {
	// Core identification
	Self string `xml:"Self,attr"`
	Name string `xml:"Name,attr,omitempty"`

	// Layer and visibility
	ItemLayer string `xml:"ItemLayer,attr,omitempty"`
	Visible   string `xml:"Visible,attr,omitempty"` // "true" or "false"

	// Geometry and transform
	GeometricBounds string `xml:"GeometricBounds,attr,omitempty"` // "y1 x1 y2 x2" format
	ItemTransform   string `xml:"ItemTransform,attr,omitempty"`   // 6-value transform matrix
}

// GetSelf returns the unique identifier for this page item.
func (p *PageItemBase) GetSelf() string {
	return p.Self
}

// GetItemLayer returns the layer ID this page item is on.
func (p *PageItemBase) GetItemLayer() string {
	return p.ItemLayer
}

// GetGeometricBounds returns the bounding box in "y1 x1 y2 x2" format.
func (p *PageItemBase) GetGeometricBounds() string {
	return p.GeometricBounds
}

// GetItemTransform returns the 6-value transformation matrix.
func (p *PageItemBase) GetItemTransform() string {
	return p.ItemTransform
}

// GetVisible returns the visibility state ("true" or "false").
func (p *PageItemBase) GetVisible() string {
	return p.Visible
}

// GetName returns the display name of the page item.
func (p *PageItemBase) GetName() string {
	return p.Name
}

// Spread represents a spread (page layout) in an IDML document.
// Spreads contain pages, guides, and page items like text frames and images.
//
// The root element is <idPkg:Spread> with the idPkg namespace.
//
// DESIGN DECISION: Dual Structure Approach
// This type uses a dual structure to handle IDML's namespace wrapper pattern:
// - Outer Spread: Handles the <idPkg:Spread> wrapper with namespace and DOMVersion
// - Inner SpreadElement: Contains the actual <Spread> content with page items
// This separation allows clean XML marshaling while providing convenient access methods.
// The alternative would be complex custom marshaling for every access method.
type Spread struct {
	// XMLName is not set directly - we handle it manually in MarshalXML/UnmarshalXML
	XMLName xml.Name `xml:"-"`

	// DOMVersion is the InDesign DOM version (e.g., "20.4")
	DOMVersion string `xml:"DOMVersion,attr"`

	// Inner spread element (the actual Spread, not the wrapper)
	// DESIGN DECISION: Embedded struct provides direct access to content
	// while maintaining the namespace wrapper structure for XML compatibility.
	InnerSpread SpreadElement `xml:"-"`

	// rawXML field removed as it was unused
}

// TextFrames returns all text frames in this spread.
// This is a convenience method to access text frames without navigating through InnerSpread.
func (s *Spread) TextFrames() []SpreadTextFrame {
	return s.InnerSpread.TextFrames
}

// Pages returns all pages in this spread.
// This is a convenience method to access pages without navigating through InnerSpread.
func (s *Spread) Pages() []Page {
	return s.InnerSpread.Pages
}

// Rectangles returns all rectangles in this spread.
// This is a convenience method to access rectangles without navigating through InnerSpread.
func (s *Spread) Rectangles() []Rectangle {
	return s.InnerSpread.Rectangles
}

// Images returns all images in this spread.
// This is a convenience method to access images without navigating through InnerSpread.
func (s *Spread) Images() []Image {
	return s.InnerSpread.Images
}

// Ovals returns all ovals in this spread.
// This is a convenience method to access ovals without navigating through InnerSpread.
func (s *Spread) Ovals() []Oval {
	return s.InnerSpread.Ovals
}

// Polygons returns all polygons in this spread.
// This is a convenience method to access polygons without navigating through InnerSpread.
func (s *Spread) Polygons() []Polygon {
	return s.InnerSpread.Polygons
}

// GraphicLines returns all graphic lines in this spread.
// This is a convenience method to access graphic lines without navigating through InnerSpread.
func (s *Spread) GraphicLines() []GraphicLine {
	return s.InnerSpread.GraphicLines
}

// SpreadElement represents the actual <Spread> element with all attributes and children.
type SpreadElement struct {
	XMLName xml.Name `xml:"Spread"`

	// Core attributes
	Self                    string `xml:"Self,attr"`
	PageTransitionType      string `xml:"PageTransitionType,attr,omitempty"`
	PageTransitionDirection string `xml:"PageTransitionDirection,attr,omitempty"`
	PageTransitionDuration  string `xml:"PageTransitionDuration,attr,omitempty"`
	ShowMasterItems         string `xml:"ShowMasterItems,attr,omitempty"`
	PageCount               string `xml:"PageCount,attr,omitempty"`
	BindingLocation         string `xml:"BindingLocation,attr,omitempty"`
	SpreadHidden            string `xml:"SpreadHidden,attr,omitempty"`
	AllowPageShuffle        string `xml:"AllowPageShuffle,attr,omitempty"`
	ItemTransform           string `xml:"ItemTransform,attr,omitempty"`
	FlattenerOverride       string `xml:"FlattenerOverride,attr,omitempty"`

	// Child elements
	FlattenerPreference *FlattenerPreference `xml:"FlattenerPreference,omitempty"`
	Pages               []Page               `xml:"Page,omitempty"`
	TextFrames          []SpreadTextFrame    `xml:"TextFrame,omitempty"`
	Rectangles          []Rectangle          `xml:"Rectangle,omitempty"`
	Images              []Image              `xml:"Image,omitempty"`
	Ovals               []Oval               `xml:"Oval,omitempty"`
	Polygons            []Polygon            `xml:"Polygon,omitempty"`
	GraphicLines        []GraphicLine        `xml:"GraphicLine,omitempty"`
	Groups              []Group              `xml:"Group,omitempty"`

	// Catch-all for other elements we haven't explicitly modeled
	OtherElements []common.RawXMLElement `xml:",any"`
}

// FlattenerPreference contains settings for transparency flattening.
type FlattenerPreference struct {
	LineArtAndTextResolution    string             `xml:"LineArtAndTextResolution,attr,omitempty"`
	GradientAndMeshResolution   string             `xml:"GradientAndMeshResolution,attr,omitempty"`
	ClipComplexRegions          string             `xml:"ClipComplexRegions,attr,omitempty"`
	ConvertAllStrokesToOutlines string             `xml:"ConvertAllStrokesToOutlines,attr,omitempty"`
	ConvertAllTextToOutlines    string             `xml:"ConvertAllTextToOutlines,attr,omitempty"`
	Properties                  *common.Properties `xml:"Properties,omitempty"`
}

// Page represents a page within a spread.
type Page struct {
	// Core attributes
	Self                   string `xml:"Self,attr"`
	TabOrder               string `xml:"TabOrder,attr,omitempty"`
	AppliedMaster          string `xml:"AppliedMaster,attr,omitempty"`
	OverrideList           string `xml:"OverrideList,attr,omitempty"`
	MasterPageTransform    string `xml:"MasterPageTransform,attr,omitempty"`
	Name                   string `xml:"Name,attr,omitempty"`
	AppliedTrapPreset      string `xml:"AppliedTrapPreset,attr,omitempty"`
	GeometricBounds        string `xml:"GeometricBounds,attr,omitempty"`
	ItemTransform          string `xml:"ItemTransform,attr,omitempty"`
	AppliedAlternateLayout string `xml:"AppliedAlternateLayout,attr,omitempty"`
	LayoutRule             string `xml:"LayoutRule,attr,omitempty"`
	SnapshotBlendingMode   string `xml:"SnapshotBlendingMode,attr,omitempty"`
	OptionalPage           string `xml:"OptionalPage,attr,omitempty"`
	GridStartingPoint      string `xml:"GridStartingPoint,attr,omitempty"`
	UseMasterGrid          string `xml:"UseMasterGrid,attr,omitempty"`

	// Child elements
	Properties          *common.Properties          `xml:"Properties,omitempty"`
	Guides              []Guide                     `xml:"Guide,omitempty"`
	MarginPreference    *MarginPreference           `xml:"MarginPreference,omitempty"`
	GridDataInformation *common.GridDataInformation `xml:"GridDataInformation,omitempty"`

	// Catch-all for other elements
	OtherElements []common.RawXMLElement `xml:",any"`
}

// Guide represents a ruler guide on a page.
type Guide struct {
	Self                    string             `xml:"Self,attr"`
	OverriddenPageItemProps string             `xml:"OverriddenPageItemProps,attr,omitempty"`
	Orientation             string             `xml:"Orientation,attr,omitempty"`
	Location                string             `xml:"Location,attr,omitempty"`
	FitToPage               string             `xml:"FitToPage,attr,omitempty"`
	ViewThreshold           string             `xml:"ViewThreshold,attr,omitempty"`
	Locked                  string             `xml:"Locked,attr,omitempty"`
	ItemLayer               string             `xml:"ItemLayer,attr,omitempty"`
	PageIndex               string             `xml:"PageIndex,attr,omitempty"`
	GuideType               string             `xml:"GuideType,attr,omitempty"`
	GuideZone               string             `xml:"GuideZone,attr,omitempty"`
	Properties              *common.Properties `xml:"Properties,omitempty"`
}

// MarginPreference contains page margin settings.
type MarginPreference struct {
	ColumnCount      string `xml:"ColumnCount,attr,omitempty"`
	ColumnGutter     string `xml:"ColumnGutter,attr,omitempty"`
	Top              string `xml:"Top,attr,omitempty"`
	Bottom           string `xml:"Bottom,attr,omitempty"`
	Left             string `xml:"Left,attr,omitempty"`
	Right            string `xml:"Right,attr,omitempty"`
	ColumnDirection  string `xml:"ColumnDirection,attr,omitempty"`
	ColumnsPositions string `xml:"ColumnsPositions,attr,omitempty"`
}

type BasicFrame struct {
	Self string `xml:"Self,attr"`
}

// SpreadTextFrame represents a text frame on a spread.
type SpreadTextFrame struct {
	PageItemBase

	// Core attributes
	ParentStory             string `xml:"ParentStory,attr,omitempty"`
	PreviousTextFrame       string `xml:"PreviousTextFrame,attr,omitempty"`
	NextTextFrame           string `xml:"NextTextFrame,attr,omitempty"`
	ContentType             string `xml:"ContentType,attr,omitempty"`
	OverriddenPageItemProps string `xml:"OverriddenPageItemProps,attr,omitempty"`

	// Layout constraints
	HorizontalLayoutConstraints string `xml:"HorizontalLayoutConstraints,attr,omitempty"`
	VerticalLayoutConstraints   string `xml:"VerticalLayoutConstraints,attr,omitempty"`

	// Gradient properties
	GradientFillStart          string `xml:"GradientFillStart,attr,omitempty"`
	GradientFillLength         string `xml:"GradientFillLength,attr,omitempty"`
	GradientFillAngle          string `xml:"GradientFillAngle,attr,omitempty"`
	GradientFillHiliteLength   string `xml:"GradientFillHiliteLength,attr,omitempty"`
	GradientFillHiliteAngle    string `xml:"GradientFillHiliteAngle,attr,omitempty"`
	GradientStrokeStart        string `xml:"GradientStrokeStart,attr,omitempty"`
	GradientStrokeLength       string `xml:"GradientStrokeLength,attr,omitempty"`
	GradientStrokeAngle        string `xml:"GradientStrokeAngle,attr,omitempty"`
	GradientStrokeHiliteLength string `xml:"GradientStrokeHiliteLength,attr,omitempty"`
	GradientStrokeHiliteAngle  string `xml:"GradientStrokeHiliteAngle,attr,omitempty"`

	// Layer and locking
	Locked              string `xml:"Locked,attr,omitempty"`
	LocalDisplaySetting string `xml:"LocalDisplaySetting,attr,omitempty"`

	// Style and transform
	AppliedObjectStyle string `xml:"AppliedObjectStyle,attr,omitempty"`

	// Version tracking
	ParentInterfaceChangeCount      string `xml:"ParentInterfaceChangeCount,attr,omitempty"`
	TargetInterfaceChangeCount      string `xml:"TargetInterfaceChangeCount,attr,omitempty"`
	LastUpdatedInterfaceChangeCount string `xml:"LastUpdatedInterfaceChangeCount,attr,omitempty"`

	// Child elements
	Properties *common.Properties `xml:"Properties,omitempty"`

	// Catch-all for all other attributes and elements
	OtherElements []common.RawXMLElement `xml:",any"`
}

// Oval represents an elliptical or circular frame on a spread.
// Ovals can contain images, text, or be empty decorative elements.
type Oval struct {
	PageItemBase

	// Content and display
	ContentType string `xml:"ContentType,attr,omitempty"` // "Unassigned", "GraphicType", "TextType"

	// Layer and locking
	LockState string `xml:"LockState,attr,omitempty"` // "None", etc.
	Locked    string `xml:"Locked,attr,omitempty"`    // "true" or "false"

	// Stroke properties
	StrokeWeight string `xml:"StrokeWeight,attr,omitempty"` // Border width in points
	StrokeType   string `xml:"StrokeType,attr,omitempty"`   // "Solid", "Dashed", etc.
	StrokeColor  string `xml:"StrokeColor,attr,omitempty"`  // Color swatch reference
	StrokeTint   string `xml:"StrokeTint,attr,omitempty"`   // Tint percentage

	// Fill properties
	FillColor string `xml:"FillColor,attr,omitempty"`
	FillTint  string `xml:"FillTint,attr,omitempty"`

	// Applied styles
	AppliedObjectStyle string `xml:"AppliedObjectStyle,attr,omitempty"`

	// Display settings
	OverriddenPageItemProps string `xml:"OverriddenPageItemProps,attr,omitempty"`
	LocalDisplaySetting     string `xml:"LocalDisplaySetting,attr,omitempty"`

	// Child elements
	Properties         *common.Properties  `xml:"Properties,omitempty"`
	TextWrapPreference *TextWrapPreference `xml:"TextWrapPreference,omitempty"`
	Image              *Image              `xml:"Image,omitempty"` // If oval contains an image

	// Catch-all for other elements
	OtherElements []common.RawXMLElement `xml:",any"`
}

// Polygon represents a multi-sided shape on a spread.
// Polygons can be regular (equal sides) or irregular, and can contain images or text.
type Polygon struct {
	PageItemBase

	// Content and display
	ContentType string `xml:"ContentType,attr,omitempty"` // "Unassigned", "GraphicType", "TextType"

	// Layer and locking
	LockState string `xml:"LockState,attr,omitempty"` // "None", etc.
	Locked    string `xml:"Locked,attr,omitempty"`    // "true" or "false"

	// Stroke properties
	StrokeWeight string `xml:"StrokeWeight,attr,omitempty"` // Border width in points
	StrokeType   string `xml:"StrokeType,attr,omitempty"`   // "Solid", "Dashed", etc.
	StrokeColor  string `xml:"StrokeColor,attr,omitempty"`  // Color swatch reference
	StrokeTint   string `xml:"StrokeTint,attr,omitempty"`   // Tint percentage

	// Fill properties
	FillColor string `xml:"FillColor,attr,omitempty"`
	FillTint  string `xml:"FillTint,attr,omitempty"`

	// Applied styles
	AppliedObjectStyle string `xml:"AppliedObjectStyle,attr,omitempty"`

	// Display settings
	OverriddenPageItemProps string `xml:"OverriddenPageItemProps,attr,omitempty"`
	LocalDisplaySetting     string `xml:"LocalDisplaySetting,attr,omitempty"`

	// Child elements
	Properties         *common.Properties  `xml:"Properties,omitempty"`
	TextWrapPreference *TextWrapPreference `xml:"TextWrapPreference,omitempty"`
	Image              *Image              `xml:"Image,omitempty"` // If polygon contains an image

	// Catch-all for other elements
	OtherElements []common.RawXMLElement `xml:",any"`
}

// GraphicLine represents a line or path on a spread.
// GraphicLines are vector drawing elements with stroke properties and optional arrowheads.
type GraphicLine struct {
	PageItemBase

	// Content and display
	ContentType string `xml:"ContentType,attr,omitempty"` // "Unassigned", "GraphicType"

	// Layer and locking
	LockState string `xml:"LockState,attr,omitempty"` // "None", etc.
	Locked    string `xml:"Locked,attr,omitempty"`    // "true" or "false"

	// Stroke properties
	StrokeWeight string `xml:"StrokeWeight,attr,omitempty"` // Line width in points
	StrokeType   string `xml:"StrokeType,attr,omitempty"`   // "Solid", "Dashed", etc.
	StrokeColor  string `xml:"StrokeColor,attr,omitempty"`  // Color swatch reference (e.g., "Color/Black")
	StrokeTint   string `xml:"StrokeTint,attr,omitempty"`   // Tint percentage

	// Fill properties (lines typically don't have fill, but included for completeness)
	FillColor string `xml:"FillColor,attr,omitempty"`
	FillTint  string `xml:"FillTint,attr,omitempty"`

	// Line cap and join
	EndCap     string `xml:"EndCap,attr,omitempty"`     // "ButtEndCap", "RoundEndCap", "ProjectingEndCap"
	EndJoin    string `xml:"EndJoin,attr,omitempty"`    // "MiterEndJoin", "RoundEndJoin", "BevelEndJoin"
	MiterLimit string `xml:"MiterLimit,attr,omitempty"` // Miter limit for sharp corners

	// Arrowheads/line ends
	LeftLineEnd  string `xml:"LeftLineEnd,attr,omitempty"`  // "None", "SimpleArrow", etc.
	RightLineEnd string `xml:"RightLineEnd,attr,omitempty"` // "None", "SimpleArrow", etc.

	// Applied styles
	AppliedObjectStyle string `xml:"AppliedObjectStyle,attr,omitempty"`

	// Display settings
	OverriddenPageItemProps string `xml:"OverriddenPageItemProps,attr,omitempty"`
	LocalDisplaySetting     string `xml:"LocalDisplaySetting,attr,omitempty"` // "Default", etc.

	// Gradient properties (for gradient strokes and fills)
	GradientFillStart          string `xml:"GradientFillStart,attr,omitempty"`
	GradientFillLength         string `xml:"GradientFillLength,attr,omitempty"`
	GradientFillAngle          string `xml:"GradientFillAngle,attr,omitempty"`
	GradientFillHiliteLength   string `xml:"GradientFillHiliteLength,attr,omitempty"`
	GradientFillHiliteAngle    string `xml:"GradientFillHiliteAngle,attr,omitempty"`
	GradientStrokeStart        string `xml:"GradientStrokeStart,attr,omitempty"`
	GradientStrokeLength       string `xml:"GradientStrokeLength,attr,omitempty"`
	GradientStrokeAngle        string `xml:"GradientStrokeAngle,attr,omitempty"`
	GradientStrokeHiliteLength string `xml:"GradientStrokeHiliteLength,attr,omitempty"`
	GradientStrokeHiliteAngle  string `xml:"GradientStrokeHiliteAngle,attr,omitempty"`

	// Layout constraints
	HorizontalLayoutConstraints string `xml:"HorizontalLayoutConstraints,attr,omitempty"`
	VerticalLayoutConstraints   string `xml:"VerticalLayoutConstraints,attr,omitempty"`

	// Version tracking (for complex documents)
	ParentInterfaceChangeCount      string `xml:"ParentInterfaceChangeCount,attr,omitempty"`
	TargetInterfaceChangeCount      string `xml:"TargetInterfaceChangeCount,attr,omitempty"`
	LastUpdatedInterfaceChangeCount string `xml:"LastUpdatedInterfaceChangeCount,attr,omitempty"`

	// Child elements
	PathGeometry       *common.PathGeometry `xml:"PathGeometry,omitempty"`
	Properties         *common.Properties   `xml:"Properties,omitempty"`
	TextWrapPreference *TextWrapPreference  `xml:"TextWrapPreference,omitempty"`
	ObjectExportOption *ObjectExportOption  `xml:"ObjectExportOption,omitempty"`

	// Catch-all for other elements
	OtherElements []common.RawXMLElement `xml:",any"`
}

// Group represents a collection of page items grouped together.
type Group struct {
	PageItemBase
	AppliedObjectStyle string                 `xml:"AppliedObjectStyle,attr,omitempty"`
	OtherElements      []common.RawXMLElement `xml:",any"`
}
