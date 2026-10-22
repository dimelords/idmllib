package document

import (
	"encoding/xml"
	"strings"

	"github.com/dimelords/idmllib/v3/pkg/common"
)

// Hyperlink links a source in the document to a destination. The source is a
// HyperlinkPageItemSource for a linked page item, or a HyperlinkTextSource in
// a story for linked text.
//
// EpubAriaRole and HypherlinkAltText (InDesign's spelling) are usually written
// empty, so they are left to OtherAttrs, which keeps an empty attribute; a
// typed omitempty field would drop it.
type Hyperlink struct {
	XMLName xml.Name `xml:"Hyperlink"`

	Self   string `xml:"Self,attr"`             // Unique identifier (e.g., "u106")
	Name   string `xml:"Name,attr"`             // Display name
	Source string `xml:"Source,attr,omitempty"` // Self of the hyperlink source

	// Appearance of the link area in exported PDFs
	Visible     string `xml:"Visible,attr,omitempty"`     // "true"/"false"
	Highlight   string `xml:"Highlight,attr,omitempty"`   // e.g., "None", "Invert"
	Width       string `xml:"Width,attr,omitempty"`       // e.g., "Thin"
	BorderStyle string `xml:"BorderStyle,attr,omitempty"` // e.g., "Solid"

	Hidden               string `xml:"Hidden,attr,omitempty"`               // "true"/"false"
	DestinationUniqueKey string `xml:"DestinationUniqueKey,attr,omitempty"` // Key of the destination

	// Properties holds BorderColor and Destination, the Self of the
	// destination as text (e.g., "HyperlinkURLDestination/URL 1").
	Properties *common.Properties `xml:"Properties,omitempty"`

	// Catch-all for other Hyperlink children
	OtherElements []common.RawXMLElement `xml:",any"`

	// OtherAttrs preserves attributes not modeled by a typed field, so
	// nothing is lost when the element is written back.
	OtherAttrs []xml.Attr `xml:",any,attr"`

	// childOrder records the document order of children so MarshalXML can
	// replay them; see common.ChildOrder.
	childOrder common.ChildOrder
}

// Destination returns the Self of the hyperlink's destination, from its
// Properties, or "" when it names none.
func (h Hyperlink) Destination() string {
	if h.Properties == nil {
		return ""
	}

	for _, el := range h.Properties.OtherElements {
		if el.XMLName.Local == "Destination" {
			return strings.TrimSpace(string(el.Content))
		}
	}

	return ""
}

// HyperlinkPageItemSource makes a page item, such as a rectangle, the source
// of a hyperlink.
type HyperlinkPageItemSource struct {
	XMLName xml.Name `xml:"HyperlinkPageItemSource"`

	Self           string `xml:"Self,attr"`                     // Unique identifier (e.g., "u78e1")
	Name           string `xml:"Name,attr"`                     // Display name (e.g., "Hyperlink 1")
	SourcePageItem string `xml:"SourcePageItem,attr,omitempty"` // Self of the linked page item
	Hidden         string `xml:"Hidden,attr,omitempty"`         // "true"/"false"

	// Catch-all for HyperlinkPageItemSource children
	OtherElements []common.RawXMLElement `xml:",any"`

	// OtherAttrs preserves attributes not modeled by a typed field, so
	// nothing is lost when the element is written back.
	OtherAttrs []xml.Attr `xml:",any,attr"`

	// childOrder records the document order of children so MarshalXML can
	// replay them; see common.ChildOrder.
	childOrder common.ChildOrder
}

// HyperlinkURLDestination is a hyperlink destination outside the document.
type HyperlinkURLDestination struct {
	XMLName xml.Name `xml:"HyperlinkURLDestination"`

	Self                 string `xml:"Self,attr"`                           // e.g., "HyperlinkURLDestination/URL 1"
	Name                 string `xml:"Name,attr"`                           // Display name
	DestinationURL       string `xml:"DestinationURL,attr,omitempty"`       // The URL the link opens
	Hidden               string `xml:"Hidden,attr,omitempty"`               // "true"/"false"
	DestinationUniqueKey string `xml:"DestinationUniqueKey,attr,omitempty"` // Key hyperlinks refer to

	// Catch-all for HyperlinkURLDestination children
	OtherElements []common.RawXMLElement `xml:",any"`

	// OtherAttrs preserves attributes not modeled by a typed field, so
	// nothing is lost when the element is written back.
	OtherAttrs []xml.Attr `xml:",any,attr"`

	// childOrder records the document order of children so MarshalXML can
	// replay them; see common.ChildOrder.
	childOrder common.ChildOrder
}

// PageItemHyperlinks returns the hyperlinks whose source is the page item
// with the given Self, through a HyperlinkPageItemSource.
func (d *Document) PageItemHyperlinks(pageItemSelf string) []Hyperlink {
	sources := make(map[string]bool)

	for _, src := range d.HyperlinkPageItemSources {
		if src.SourcePageItem == pageItemSelf {
			sources[src.Self] = true
		}
	}

	var links []Hyperlink

	for _, h := range d.Hyperlinks {
		if sources[h.Source] {
			links = append(links, h)
		}
	}

	return links
}

// HyperlinkURL returns the URL a hyperlink opens, or false when its
// destination is not a HyperlinkURLDestination. The destination is found by
// the Self in the hyperlink's Properties and, when that names none, by
// DestinationUniqueKey.
func (d *Document) HyperlinkURL(h Hyperlink) (string, bool) {
	if dest := h.Destination(); dest != "" {
		for _, u := range d.HyperlinkURLDestinations {
			if u.Self == dest {
				return u.DestinationURL, true
			}
		}

		return "", false
	}

	if h.DestinationUniqueKey == "" {
		return "", false
	}

	for _, u := range d.HyperlinkURLDestinations {
		if u.DestinationUniqueKey == h.DestinationUniqueKey {
			return u.DestinationURL, true
		}
	}

	return "", false
}

// PageItemURLs returns the URLs that the page item with the given Self links
// to, in the order of its hyperlinks.
func (d *Document) PageItemURLs(pageItemSelf string) []string {
	var urls []string

	for _, h := range d.PageItemHyperlinks(pageItemSelf) {
		if u, ok := d.HyperlinkURL(h); ok {
			urls = append(urls, u)
		}
	}

	return urls
}
