package resources

import (
	"encoding/xml"

	"github.com/dimelords/idmllib/pkg/common"
)

// GraphicFile represents the Resources/Graphic.xml file containing colors, swatches, and graphic resources.
//
// The root element is <idPkg:Graphic> with the idPkg namespace.
type GraphicFile struct {
	// XMLName is not set directly - we handle it manually in MarshalXML/UnmarshalXML
	XMLName xml.Name `xml:"-"`

	// DOMVersion is the InDesign DOM version (e.g., "20.4")
	DOMVersion string `xml:"DOMVersion,attr"`

	// Colors define color swatches (CMYK, RGB, LAB, etc.)
	Colors []Color `xml:"Color,omitempty"`

	// Inks define printing ink properties
	Inks []Ink `xml:"Ink,omitempty"`

	// Gradients define gradient fill definitions
	Gradients []Gradient `xml:"Gradient,omitempty"`

	// Swatches define named swatch references
	Swatches []Swatch `xml:"Swatch,omitempty"`

	// PastedSmoothShades define pasted smooth shade definitions
	PastedSmoothShades []PastedSmoothShade `xml:"PastedSmoothShade,omitempty"`

	// StrokeStyles define stroke/line style definitions
	StrokeStyles []StrokeStyle `xml:"StrokeStyle,omitempty"`

	// Catch-all for other elements we haven't explicitly modeled
	OtherElements []common.RawXMLElement `xml:",any"`
}

// Color represents a color swatch definition.
// Supports Process colors (CMYK, RGB, LAB) and Spot colors.
type Color struct {
	Self                      string `xml:"Self,attr"`
	Model                     string `xml:"Model,attr"`                         // "Process" or "Registration"
	Space                     string `xml:"Space,attr"`                         // "CMYK", "RGB", "LAB"
	ColorValue                string `xml:"ColorValue,attr"`                    // Space-separated values (e.g., "0 0 100 0")
	ColorOverride             string `xml:"ColorOverride,attr,omitempty"`       // "Normal", "Specialblack", "Hiddenreserved", etc.
	ConvertToHsb              string `xml:"ConvertToHsb,attr,omitempty"`        // "true" or "false"
	AlternateSpace            string `xml:"AlternateSpace,attr,omitempty"`      // "NoAlternateColor" or alternate color space
	AlternateColorValue       string `xml:"AlternateColorValue,attr,omitempty"` // Alternate color values
	Name                      string `xml:"Name,attr"`
	ColorEditable             string `xml:"ColorEditable,attr,omitempty"`  // "true" or "false"
	ColorRemovable            string `xml:"ColorRemovable,attr,omitempty"` // "true" or "false"
	Visible                   string `xml:"Visible,attr,omitempty"`        // "true" or "false"
	SwatchCreatorID           string `xml:"SwatchCreatorID,attr,omitempty"`
	SwatchColorGroupReference string `xml:"SwatchColorGroupReference,attr,omitempty"`
}

// Ink represents a printing ink definition.
// Controls ink behavior in color separation and printing.
type Ink struct {
	Self             string `xml:"Self,attr"`
	Name             string `xml:"Name,attr"`
	Angle            string `xml:"Angle,attr,omitempty"`            // Screen angle in degrees
	ConvertToProcess string `xml:"ConvertToProcess,attr,omitempty"` // "true" or "false"
	Frequency        string `xml:"Frequency,attr,omitempty"`        // Screen frequency in lpi
	NeutralDensity   string `xml:"NeutralDensity,attr,omitempty"`   // Optical density value
	PrintInk         string `xml:"PrintInk,attr,omitempty"`         // "true" or "false"
	TrapOrder        string `xml:"TrapOrder,attr,omitempty"`        // Trap order number
	InkType          string `xml:"InkType,attr,omitempty"`          // "Normal", "Transparent", "Opaque"
}

// Gradient represents a gradient fill definition.
// Supports Linear and Radial gradient types with multiple stops.
type Gradient struct {
	Self                      string         `xml:"Self,attr"`
	Type                      string         `xml:"Type,attr"` // "Linear" or "Radial"
	Name                      string         `xml:"Name,attr"`
	ColorEditable             string         `xml:"ColorEditable,attr,omitempty"`  // "true" or "false"
	ColorRemovable            string         `xml:"ColorRemovable,attr,omitempty"` // "true" or "false"
	Visible                   string         `xml:"Visible,attr,omitempty"`        // "true" or "false"
	SwatchCreatorID           string         `xml:"SwatchCreatorID,attr,omitempty"`
	SwatchColorGroupReference string         `xml:"SwatchColorGroupReference,attr,omitempty"`
	GradientStops             []GradientStop `xml:"GradientStop,omitempty"`
}

// GradientStop represents a color stop in a gradient.
type GradientStop struct {
	Self      string `xml:"Self,attr"`
	StopColor string `xml:"StopColor,attr"`          // Reference to a Color (e.g., "Color/Black")
	Location  string `xml:"Location,attr"`           // Position 0-100
	Midpoint  string `xml:"Midpoint,attr,omitempty"` // Midpoint position (0-100)
}

// Swatch represents a named swatch reference.
// Typically references "None" or other swatches.
type Swatch struct {
	Self                      string `xml:"Self,attr"`
	Name                      string `xml:"Name,attr"`
	ColorEditable             string `xml:"ColorEditable,attr,omitempty"`  // "true" or "false"
	ColorRemovable            string `xml:"ColorRemovable,attr,omitempty"` // "true" or "false"
	Visible                   string `xml:"Visible,attr,omitempty"`        // "true" or "false"
	SwatchCreatorID           string `xml:"SwatchCreatorID,attr,omitempty"`
	SwatchColorGroupReference string `xml:"SwatchColorGroupReference,attr,omitempty"`
}

// PastedSmoothShade represents a pasted smooth shade definition.
// Contains embedded shade data in Contents property.
type PastedSmoothShade struct {
	Self                      string             `xml:"Self,attr"`
	ContentsVersion           string             `xml:"ContentsVersion,attr,omitempty"`
	ContentsType              string             `xml:"ContentsType,attr,omitempty"` // "ConstantShade", etc.
	SpotColorList             string             `xml:"SpotColorList,attr,omitempty"`
	ContentsEncoding          string             `xml:"ContentsEncoding,attr,omitempty"` // "Ascii64Encoding"
	ContentsMatrix            string             `xml:"ContentsMatrix,attr,omitempty"`   // Transform matrix
	Name                      string             `xml:"Name,attr"`
	ColorEditable             string             `xml:"ColorEditable,attr,omitempty"`  // "true" or "false"
	ColorRemovable            string             `xml:"ColorRemovable,attr,omitempty"` // "true" or "false"
	Visible                   string             `xml:"Visible,attr,omitempty"`        // "true" or "false"
	SwatchCreatorID           string             `xml:"SwatchCreatorID,attr,omitempty"`
	SwatchColorGroupReference string             `xml:"SwatchColorGroupReference,attr,omitempty"`
	Properties                *common.Properties `xml:"Properties,omitempty"` // Contains <Contents> CDATA
}

// StrokeStyle represents a stroke/line style definition.
// Defines patterns for lines and strokes (solid, dashed, dotted, etc.).
type StrokeStyle struct {
	Self string `xml:"Self,attr"`
	Name string `xml:"Name,attr"`
	// Additional stroke properties would go here
	OtherElements []common.RawXMLElement `xml:",any"`
}
