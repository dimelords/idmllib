package spread

import (
	"encoding/xml"

	"github.com/dimelords/idmllib/v2/pkg/common"
)

// Rectangle represents a rectangular page item in a spread.
// Rectangles can contain text, graphics, or be empty frames.
type Rectangle struct {
	PageItemBase

	// Content and display
	ContentType             string `xml:"ContentType,attr,omitempty"` // "TextType", "GraphicType", "Unassigned"
	StoryTitle              string `xml:"StoryTitle,attr,omitempty"`
	OverriddenPageItemProps string `xml:"OverriddenPageItemProps,attr,omitempty"`

	// Layout constraints
	HorizontalLayoutConstraints string `xml:"HorizontalLayoutConstraints,attr,omitempty"` // e.g., "FlexibleDimension FixedDimension FlexibleDimension"
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
	Locked              string `xml:"Locked,attr,omitempty"`              // "true" or "false"
	LocalDisplaySetting string `xml:"LocalDisplaySetting,attr,omitempty"` // "Default", etc.

	// Style and transform
	AppliedObjectStyle string `xml:"AppliedObjectStyle,attr,omitempty"`

	// Version tracking (for complex documents)
	ParentInterfaceChangeCount      string `xml:"ParentInterfaceChangeCount,attr,omitempty"`
	TargetInterfaceChangeCount      string `xml:"TargetInterfaceChangeCount,attr,omitempty"`
	LastUpdatedInterfaceChangeCount string `xml:"LastUpdatedInterfaceChangeCount,attr,omitempty"`

	// Child elements
	Properties         *common.Properties  `xml:"Properties,omitempty"`
	FrameFittingOption *FrameFittingOption `xml:"FrameFittingOption,omitempty"`
	ObjectExportOption *ObjectExportOption `xml:"ObjectExportOption,omitempty"`
	TextWrapPreference *TextWrapPreference `xml:"TextWrapPreference,omitempty"`
	InCopyExportOption *InCopyExportOption `xml:"InCopyExportOption,omitempty"`
	Image              *Image              `xml:"Image,omitempty"`
	PDF                *PDF                `xml:"PDF,omitempty"`

	// Catch-all for other elements
	OtherElements []common.RawXMLElement `xml:",any"`
}

// FrameFittingOption controls how content fits within a frame.
// Critical for image rectangles.
type FrameFittingOption struct {
	AutoFit             string `xml:"AutoFit,attr,omitempty"` // "true" or "false"
	LeftCrop            string `xml:"LeftCrop,attr,omitempty"`
	TopCrop             string `xml:"TopCrop,attr,omitempty"`
	RightCrop           string `xml:"RightCrop,attr,omitempty"`
	BottomCrop          string `xml:"BottomCrop,attr,omitempty"`
	FittingOnEmptyFrame string `xml:"FittingOnEmptyFrame,attr,omitempty"` // "None", "FitContentProportionally", etc.
	FittingAlignment    string `xml:"FittingAlignment,attr,omitempty"`    // "TopLeftAnchor", "CenterAnchor", etc.
}

// FrameContentBase contains common attributes shared by Image and PDF content.
// These represent content placed within a frame (Rectangle, Oval, Polygon).
type FrameContentBase struct {
	// Core identification
	Self string `xml:"Self,attr"`
	Name string `xml:"Name,attr,omitempty"`

	// Display and style
	OverriddenPageItemProps string `xml:"OverriddenPageItemProps,attr,omitempty"`
	LocalDisplaySetting     string `xml:"LocalDisplaySetting,attr,omitempty"`
	ImageTypeName           string `xml:"ImageTypeName,attr,omitempty"` // e.g., "$ID/Portable Network Graphics (PNG)" or "$ID/Adobe Portable Document Format (PDF)"
	AppliedObjectStyle      string `xml:"AppliedObjectStyle,attr,omitempty"`
	Visible                 string `xml:"Visible,attr,omitempty"`

	// Layout constraints
	HorizontalLayoutConstraints string `xml:"HorizontalLayoutConstraints,attr,omitempty"`
	VerticalLayoutConstraints   string `xml:"VerticalLayoutConstraints,attr,omitempty"`

	// Transform (position, rotation, scale)
	ItemTransform string `xml:"ItemTransform,attr,omitempty"`

	// Version tracking
	ParentInterfaceChangeCount      string `xml:"ParentInterfaceChangeCount,attr,omitempty"`
	TargetInterfaceChangeCount      string `xml:"TargetInterfaceChangeCount,attr,omitempty"`
	LastUpdatedInterfaceChangeCount string `xml:"LastUpdatedInterfaceChangeCount,attr,omitempty"`
}

// Image represents an image placed in a frame (typically Rectangle).
// Images are linked to external files.
type Image struct {
	FrameContentBase

	// Image-specific properties
	Space                string `xml:"Space,attr,omitempty"`
	ActualPpi            string `xml:"ActualPpi,attr,omitempty"`            // "72 72" format
	EffectivePpi         string `xml:"EffectivePpi,attr,omitempty"`         // "394 394" format
	ImageRenderingIntent string `xml:"ImageRenderingIntent,attr,omitempty"` // "UseColorSettings", etc.

	// Gradient properties
	GradientFillStart        string `xml:"GradientFillStart,attr,omitempty"`
	GradientFillLength       string `xml:"GradientFillLength,attr,omitempty"`
	GradientFillAngle        string `xml:"GradientFillAngle,attr,omitempty"`
	GradientFillHiliteLength string `xml:"GradientFillHiliteLength,attr,omitempty"`
	GradientFillHiliteAngle  string `xml:"GradientFillHiliteAngle,attr,omitempty"`

	// Child elements
	Properties           *common.Properties    `xml:"Properties,omitempty"`
	ClippingPathSettings *ClippingPathSettings `xml:"ClippingPathSettings,omitempty"`
	ImageIOPreference    *ImageIOPreference    `xml:"ImageIOPreference,omitempty"`
	TextWrapPreference   *TextWrapPreference   `xml:"TextWrapPreference,omitempty"`
	Link                 *Link                 `xml:"Link,omitempty"`

	// Catch-all for other elements
	OtherElements []common.RawXMLElement `xml:",any"`
}

// Link represents a link to an external file (image, etc.).
type Link struct {
	Self                       string `xml:"Self,attr"`
	AssetURL                   string `xml:"AssetURL,attr,omitempty"`
	AssetID                    string `xml:"AssetID,attr,omitempty"`
	LinkResourceURI            string `xml:"LinkResourceURI,attr,omitempty"` // "file:/path/to/image.jpg"
	LinkResourceFormat         string `xml:"LinkResourceFormat,attr,omitempty"`
	StoredState                string `xml:"StoredState,attr,omitempty"` // "Normal", "Modified", "Missing"
	LinkClassID                string `xml:"LinkClassID,attr,omitempty"`
	LinkClientID               string `xml:"LinkClientID,attr,omitempty"`
	LinkResourceModified       string `xml:"LinkResourceModified,attr,omitempty"` // "true" or "false"
	LinkObjectModified         string `xml:"LinkObjectModified,attr,omitempty"`
	ShowInUI                   string `xml:"ShowInUI,attr,omitempty"`
	CanEmbed                   string `xml:"CanEmbed,attr,omitempty"`
	CanUnembed                 string `xml:"CanUnembed,attr,omitempty"`
	CanPackage                 string `xml:"CanPackage,attr,omitempty"`
	ImportPolicy               string `xml:"ImportPolicy,attr,omitempty"` // "NoAutoImport"
	ExportPolicy               string `xml:"ExportPolicy,attr,omitempty"`
	LinkImportStamp            string `xml:"LinkImportStamp,attr,omitempty"`
	LinkImportModificationTime string `xml:"LinkImportModificationTime,attr,omitempty"`
	LinkImportTime             string `xml:"LinkImportTime,attr,omitempty"`
	LinkResourceSize           string `xml:"LinkResourceSize,attr,omitempty"`
	RenditionData              string `xml:"RenditionData,attr,omitempty"` // "Actual"
}

// ClippingPathSettings controls image clipping behavior.
type ClippingPathSettings struct {
	ClippingType           string `xml:"ClippingType,attr,omitempty"` // "None", "PhotoshopPath", etc.
	InvertPath             string `xml:"InvertPath,attr,omitempty"`
	IncludeInsideEdges     string `xml:"IncludeInsideEdges,attr,omitempty"`
	RestrictToFrame        string `xml:"RestrictToFrame,attr,omitempty"`
	UseHighResolutionImage string `xml:"UseHighResolutionImage,attr,omitempty"`
	Threshold              string `xml:"Threshold,attr,omitempty"`
	Tolerance              string `xml:"Tolerance,attr,omitempty"`
	InsetFrame             string `xml:"InsetFrame,attr,omitempty"`
	AppliedPathName        string `xml:"AppliedPathName,attr,omitempty"`
	Index                  string `xml:"Index,attr,omitempty"`
}

// ImageIOPreference controls image import/export settings.
type ImageIOPreference struct {
	ApplyPhotoshopClippingPath string `xml:"ApplyPhotoshopClippingPath,attr,omitempty"`
	AllowAutoEmbedding         string `xml:"AllowAutoEmbedding,attr,omitempty"`
	AlphaChannelName           string `xml:"AlphaChannelName,attr,omitempty"`
}

// ContourOption controls contour wrapping.
type ContourOption struct {
	ContourType        string `xml:"ContourType,attr,omitempty"`
	IncludeInsideEdges string `xml:"IncludeInsideEdges,attr,omitempty"`
	ContourPathName    string `xml:"ContourPathName,attr,omitempty"`
}

// TextWrapPreference controls how text wraps around objects.
type TextWrapPreference struct {
	Inverse               string             `xml:"Inverse,attr,omitempty"`
	ApplyToMasterPageOnly string             `xml:"ApplyToMasterPageOnly,attr,omitempty"`
	TextWrapSide          string             `xml:"TextWrapSide,attr,omitempty"` // "BothSides", "LeftSide", "RightSide"
	TextWrapMode          string             `xml:"TextWrapMode,attr,omitempty"` // "None", "BoundingBoxTextWrap", etc.
	Properties            *common.Properties `xml:"Properties,omitempty"`
	ContourOption         *ContourOption     `xml:"ContourOption,omitempty"`
}

// InCopyExportOption controls InCopy export settings.
type InCopyExportOption struct {
	XMLName               xml.Name `xml:"InCopyExportOption"`
	IncludeGraphicProxies string   `xml:"IncludeGraphicProxies,attr,omitempty"`
	IncludeAllResources   string   `xml:"IncludeAllResources,attr,omitempty"`
}

// ObjectExportOption controls export settings for web/digital publishing.
type ObjectExportOption struct {
	XMLName xml.Name `xml:"ObjectExportOption"`
	// Placeholder for export options - full definition will come in Phase 5
	OtherElements []common.RawXMLElement `xml:",any"`
}

// PDF represents a PDF file placed in a frame (typically Rectangle).
// PDFs can be used for advertisements, imported documents, or graphics.
// Similar to Image but specifically for PDF content.
type PDF struct {
	FrameContentBase

	// PDF-specific color policy settings
	GrayVectorPolicy string `xml:"GrayVectorPolicy,attr,omitempty"` // "IgnoreAll", "HonorAllProfiles"
	RGBVectorPolicy  string `xml:"RGBVectorPolicy,attr,omitempty"`  // "IgnoreAll", "HonorAllProfiles"
	CMYKVectorPolicy string `xml:"CMYKVectorPolicy,attr,omitempty"` // "IgnoreAll", "HonorAllProfiles"

	// Child elements
	Properties         *common.Properties  `xml:"Properties,omitempty"`
	PDFAttribute       *PDFAttribute       `xml:"PDFAttribute,omitempty"`
	Link               *Link               `xml:"Link,omitempty"`
	TextWrapPreference *TextWrapPreference `xml:"TextWrapPreference,omitempty"`

	// Catch-all for other elements
	OtherElements []common.RawXMLElement `xml:",any"`
}

// PDFAttribute contains PDF-specific attributes like page number and crop settings.
type PDFAttribute struct {
	PageNumber            string `xml:"PageNumber,attr,omitempty"`            // "1" (which page of multi-page PDF to display)
	PDFCrop               string `xml:"PDFCrop,attr,omitempty"`               // "CropPDF", "CropContentBox", "CropMediaBox", etc.
	TransparentBackground string `xml:"TransparentBackground,attr,omitempty"` // "true" or "false"
}
