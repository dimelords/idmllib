package common

import "encoding/xml"

// InCopyExportOption controls how an item is exported to InCopy. It appears
// on stories and on page items; the same element is shared by both.
type InCopyExportOption struct {
	XMLName               xml.Name `xml:"InCopyExportOption"`
	IncludeGraphicProxies string   `xml:"IncludeGraphicProxies,attr,omitempty"` // "true"/"false"
	IncludeAllResources   string   `xml:"IncludeAllResources,attr,omitempty"`   // "true"/"false"

	OtherElements []RawXMLElement `xml:",any"`
	OtherAttrs    []xml.Attr      `xml:",any,attr"`
}

// ObjectExportOption holds accessibility and export settings (alt text, tagged
// PDF, EPUB) for a page item or object style.
type ObjectExportOption struct {
	XMLName               xml.Name `xml:"ObjectExportOption"`
	AltTextSourceType     string   `xml:"AltTextSourceType,attr,omitempty"`
	ActualTextSourceType  string   `xml:"ActualTextSourceType,attr,omitempty"`
	CustomAltText         string   `xml:"CustomAltText,attr,omitempty"`
	CustomActualText      string   `xml:"CustomActualText,attr,omitempty"`
	ApplyTagType          string   `xml:"ApplyTagType,attr,omitempty"`
	ImageConversionType   string   `xml:"ImageConversionType,attr,omitempty"`
	ImageExportResolution string   `xml:"ImageExportResolution,attr,omitempty"`

	Properties    *Properties     `xml:"Properties,omitempty"`
	OtherElements []RawXMLElement `xml:",any"`
	OtherAttrs    []xml.Attr      `xml:",any,attr"`

	childOrder ChildOrder
}
