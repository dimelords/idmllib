//revive:disable:var-naming
package types

import "encoding/xml"

// IDPkgSpread represents the root element containing a Spread
type IDPkgSpread struct {
	Spread Spread `xml:"Spread"`
}

// Spread represents a spread in an InDesign document
type Spread struct {
	Self      string `xml:"Self,attr"`
	PageCount int    `xml:"PageCount,attr"`

	Pages      []Page      `xml:"Page"`
	TextFrames []TextFrame `xml:"TextFrame"`
	Rectangles []Rectangle `xml:"Rectangle"`
}

// TextFrame - the key struct you're interested in
type TextFrame struct {
	XMLName                         xml.Name `xml:"TextFrame"`
	Self                            string   `xml:"Self,attr"`
	ParentStory                     string   `xml:"ParentStory,attr,omitempty"`
	PreviousTextFrame               string   `xml:"PreviousTextFrame,attr,omitempty"`
	NextTextFrame                   string   `xml:"NextTextFrame,attr,omitempty"`
	ContentType                     string   `xml:"ContentType,attr,omitempty"`
	OverriddenPageItemProps         string   `xml:"OverriddenPageItemProps,attr,omitempty"`
	Visible                         string   `xml:"Visible,attr,omitempty"`
	Name                            string   `xml:"Name,attr,omitempty"`
	HorizontalLayoutConstraints     string   `xml:"HorizontalLayoutConstraints,attr,omitempty"`
	VerticalLayoutConstraints       string   `xml:"VerticalLayoutConstraints,attr,omitempty"`
	GradientFillStart               string   `xml:"GradientFillStart,attr,omitempty"`
	GradientFillLength              string   `xml:"GradientFillLength,attr,omitempty"`
	GradientFillAngle               string   `xml:"GradientFillAngle,attr,omitempty"`
	GradientStrokeStart             string   `xml:"GradientStrokeStart,attr,omitempty"`
	GradientStrokeLength            string   `xml:"GradientStrokeLength,attr,omitempty"`
	GradientStrokeAngle             string   `xml:"GradientStrokeAngle,attr,omitempty"`
	ItemLayer                       string   `xml:"ItemLayer,attr,omitempty"`
	Locked                          string   `xml:"Locked,attr,omitempty"`
	LocalDisplaySetting             string   `xml:"LocalDisplaySetting,attr,omitempty"`
	GradientFillHiliteLength        string   `xml:"GradientFillHiliteLength,attr,omitempty"`
	GradientFillHiliteAngle         string   `xml:"GradientFillHiliteAngle,attr,omitempty"`
	GradientStrokeHiliteLength      string   `xml:"GradientStrokeHiliteLength,attr,omitempty"`
	GradientStrokeHiliteAngle       string   `xml:"GradientStrokeHiliteAngle,attr,omitempty"`
	AppliedObjectStyle              string   `xml:"AppliedObjectStyle,attr,omitempty"`
	ItemTransform                   string   `xml:"ItemTransform,attr,omitempty"`
	ParentInterfaceChangeCount      string   `xml:"ParentInterfaceChangeCount,attr,omitempty"`
	TargetInterfaceChangeCount      string   `xml:"TargetInterfaceChangeCount,attr,omitempty"`
	LastUpdatedInterfaceChangeCount string   `xml:"LastUpdatedInterfaceChangeCount,attr,omitempty"`
	InnerXML                        string   `xml:",innerxml"`
	// Optional: get Label data
	Label Label `xml:"Properties>Label"`
}
