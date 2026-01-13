package resources

import (
	"encoding/xml"

	"github.com/dimelords/idmllib/pkg/common"
)

// FontsFile represents the Resources/Fonts.xml file containing font definitions.
//
// The root element is <idPkg:Fonts> with the idPkg namespace.
type FontsFile struct {
	// XMLName is not set directly - we handle it manually in MarshalXML/UnmarshalXML
	XMLName xml.Name `xml:"-"`

	// DOMVersion is the InDesign DOM version (e.g., "20.4")
	DOMVersion string `xml:"DOMVersion,attr"`

	// FontFamilies contain font family groups with individual font definitions
	FontFamilies []FontFamily `xml:"FontFamily,omitempty"`

	// CompositeFonts define composite font settings (primarily for Asian typography)
	CompositeFonts []CompositeFont `xml:"CompositeFont,omitempty"`

	// Catch-all for other elements we haven't explicitly modeled
	OtherElements []common.RawXMLElement `xml:",any"`
}

// FontFamily represents a font family group (e.g., "Minion Pro", "Myriad Pro").
// Contains multiple Font entries for different weights and styles.
type FontFamily struct {
	Self  string `xml:"Self,attr"`
	Name  string `xml:"Name,attr"`
	Fonts []Font `xml:"Font,omitempty"`
}

// Font represents an individual font definition within a font family.
// Contains detailed font metrics, style information, and installation status.
type Font struct {
	Self                string `xml:"Self,attr"`
	FontFamily          string `xml:"FontFamily,attr"`
	Name                string `xml:"Name,attr"`
	PostScriptName      string `xml:"PostScriptName,attr"`
	Status              string `xml:"Status,attr"`        // "Installed", "Substituted", "NotAvailable"
	FontStyleName       string `xml:"FontStyleName,attr"` // "Regular", "Bold", "Italic", etc.
	FontType            string `xml:"FontType,attr"`      // "OpenTypeCFF", "OpenTypeCID", "TrueType", etc.
	WritingScript       string `xml:"WritingScript,attr"` // "0" for Latin, "1" for CJK, etc.
	FullName            string `xml:"FullName,attr"`
	FullNameNative      string `xml:"FullNameNative,attr"`
	FontStyleNameNative string `xml:"FontStyleNameNative,attr"`
	PlatformName        string `xml:"PlatformName,attr"`
	Version             string `xml:"Version,attr"`
	TypekitID           string `xml:"TypekitID,attr,omitempty"` // Adobe Typekit/Fonts ID
}

// CompositeFont represents a composite font definition (primarily for CJK typography).
// Allows mixing of different fonts for different character ranges.
type CompositeFont struct {
	Self                 string                 `xml:"Self,attr"`
	Name                 string                 `xml:"Name,attr"`
	CompositeFontEntries []CompositeFontEntry   `xml:"CompositeFontEntry,omitempty"`
	OtherElements        []common.RawXMLElement `xml:",any"`
}

// CompositeFontEntry represents a single entry in a composite font.
// Defines which font to use for a specific character range or script.
type CompositeFontEntry struct {
	Self             string             `xml:"Self,attr"`
	Name             string             `xml:"Name,attr"`                       // Character range name (e.g., "$ID/Kanji", "$ID/Kana")
	FontStyle        string             `xml:"FontStyle,attr"`                  // Font style reference
	RelativeSize     string             `xml:"RelativeSize,attr,omitempty"`     // Relative size percentage
	HorizontalScale  string             `xml:"HorizontalScale,attr,omitempty"`  // Horizontal scale percentage
	VerticalScale    string             `xml:"VerticalScale,attr,omitempty"`    // Vertical scale percentage
	CustomCharacters string             `xml:"CustomCharacters,attr,omitempty"` // Custom character list
	Locked           string             `xml:"Locked,attr,omitempty"`           // "true" or "false"
	ScaleOption      string             `xml:"ScaleOption,attr,omitempty"`      // "true" or "false"
	BaselineShift    string             `xml:"BaselineShift,attr,omitempty"`    // Baseline shift value
	Properties       *common.Properties `xml:"Properties,omitempty"`            // Contains <AppliedFont>
}
