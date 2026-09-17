package spread

import (
	"encoding/xml"

	"github.com/dimelords/idmllib/v3/pkg/common"
)

// Group is a collection of page items grouped together. Its children are
// page items in stacking order; nested groups are allowed.
type Group struct {
	PageItemBase
	AppliedObjectStyle string `xml:"AppliedObjectStyle,attr,omitempty"`

	Properties   *common.Properties `xml:"Properties,omitempty"`
	TextFrames   []TextFrame        `xml:"TextFrame,omitempty"`
	Rectangles   []Rectangle        `xml:"Rectangle,omitempty"`
	Ovals        []Oval             `xml:"Oval,omitempty"`
	Polygons     []Polygon          `xml:"Polygon,omitempty"`
	GraphicLines []GraphicLine      `xml:"GraphicLine,omitempty"`
	Images       []Image            `xml:"Image,omitempty"`
	Groups       []Group            `xml:"Group,omitempty"`
	Buttons      []Button           `xml:"Button,omitempty"`

	TextWrapPreference *TextWrapPreference `xml:"TextWrapPreference,omitempty"`
	ObjectExportOption *ObjectExportOption `xml:"ObjectExportOption,omitempty"`

	// OtherElements preserves children not modeled above.
	OtherElements []common.RawXMLElement `xml:",any"`
	// OtherAttrs preserves attributes not modeled by a typed field.
	OtherAttrs []xml.Attr `xml:",any,attr"`

	childOrder common.ChildOrder
}

// Button is an interactive button. Its appearance states hold ordinary page
// items (usually a Group per state).
type Button struct {
	PageItemBase
	AppliedObjectStyle   string `xml:"AppliedObjectStyle,attr,omitempty"`
	Description          string `xml:"Description,attr,omitempty"`
	VisibilityInPdf      string `xml:"VisibilityInPdf,attr,omitempty"`
	PrintableInPdf       string `xml:"PrintableInPdf,attr,omitempty"`
	HiddenUntilTriggered string `xml:"HiddenUntilTriggered,attr,omitempty"`

	Properties         *common.Properties  `xml:"Properties,omitempty"`
	TextWrapPreference *TextWrapPreference `xml:"TextWrapPreference,omitempty"`
	States             []State             `xml:"State,omitempty"`

	OtherElements []common.RawXMLElement `xml:",any"`
	OtherAttrs    []xml.Attr             `xml:",any,attr"`

	childOrder common.ChildOrder
}

// State is one appearance state of a Button or MultiStateObject.
type State struct {
	Self    string `xml:"Self,attr,omitempty"`
	Name    string `xml:"Name,attr,omitempty"`
	Active  string `xml:"Active,attr,omitempty"`
	Enabled string `xml:"Enabled,attr,omitempty"`

	Properties   *common.Properties `xml:"Properties,omitempty"`
	TextFrames   []TextFrame        `xml:"TextFrame,omitempty"`
	Rectangles   []Rectangle        `xml:"Rectangle,omitempty"`
	Ovals        []Oval             `xml:"Oval,omitempty"`
	Polygons     []Polygon          `xml:"Polygon,omitempty"`
	GraphicLines []GraphicLine      `xml:"GraphicLine,omitempty"`
	Images       []Image            `xml:"Image,omitempty"`
	Groups       []Group            `xml:"Group,omitempty"`

	OtherElements []common.RawXMLElement `xml:",any"`
	OtherAttrs    []xml.Attr             `xml:",any,attr"`

	childOrder common.ChildOrder
}
