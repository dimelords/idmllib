package types

import "encoding/xml"

// Layer represents a layer in an InDesign document
type Layer struct {
	XMLName    xml.Name `xml:"Layer"`
	Self       string   `xml:"Self,attr"`
	Name       string   `xml:"Name,attr"`
	Visible    string   `xml:"Visible,attr"`
	Locked     string   `xml:"Locked,attr"`
	IgnoreWrap string   `xml:"IgnoreWrap,attr"`
	ShowGuides string   `xml:"ShowGuides,attr"`
	LockGuides string   `xml:"LockGuides,attr"`
	UI         string   `xml:"UI,attr"`
	Expendable string   `xml:"Expendable,attr"`
	Printable  string   `xml:"Printable,attr"`
	Properties struct {
		LayerColor LayerColor `xml:"LayerColor"`
	} `xml:"Properties"`
}

// LayerColor represents the color of a layer
type LayerColor struct {
	Type  string `xml:"type,attr"`
	Value string `xml:",chardata"`
}
