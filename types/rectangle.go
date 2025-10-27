//revive:disable:var-naming
package types

import "encoding/xml"

// Rectangle represents a rectangle shape in an InDesign document
type Rectangle struct {
	XMLName                     xml.Name `xml:"Rectangle"`
	Self                        string   `xml:"Self,attr"`
	ContentType                 string   `xml:"ContentType,attr"`
	StoryTitle                  string   `xml:"StoryTitle,attr"`
	OverriddenPageItemProps     string   `xml:"OverriddenPageItemProps,attr"`
	Visible                     string   `xml:"Visible,attr"`
	Name                        string   `xml:"Name,attr"`
	HorizontalLayoutConstraints string   `xml:"HorizontalLayoutConstraints,attr"`
	VerticalLayoutConstraints   string   `xml:"VerticalLayoutConstraints,attr"`
	GradientFillStart           string   `xml:"GradientFillStart,attr"`
	GradientFillLength          string   `xml:"GradientFillLength,attr"`
	GradientFillAngle           string   `xml:"GradientFillAngle,attr"`
	GradientStrokeStart         string   `xml:"GradientStrokeStart,attr"`
	GradientStrokeLength        string   `xml:"GradientStrokeLength,attr"`
	GradientStrokeAngle         string   `xml:"GradientStrokeAngle,attr"`
	ItemLayer                   string   `xml:"ItemLayer,attr"`
	Locked                      string   `xml:"Locked,attr"`
	LocalDisplaySetting         string   `xml:"LocalDisplaySetting,attr"`
	GradientFillHiliteLength    string   `xml:"GradientFillHiliteLength,attr"`
	GradientFillHiliteAngle     string   `xml:"GradientFillHiliteAngle,attr"`
	GradientStrokeHiliteLength  string   `xml:"GradientStrokeHiliteLength,attr"`
	GradientStrokeHiliteAngle   string   `xml:"GradientStrokeHiliteAngle,attr"`
}
