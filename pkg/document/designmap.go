package document

import (
	"encoding/xml"

	"github.com/dimelords/idmllib/pkg/common"
)

// Designmap represents the main Document element in designmap.xml.
// This is a minimal struct that preserves unknown elements for forward compatibility.
//
// DEPRECATED: Phase 1 implementation. Use Document struct in document.go for Phase 2+.
// This struct is kept for backwards compatibility and testing purposes only.
type Designmap struct {
	// XMLName captures the element name and namespace.
	// Note: The default namespace is NOT the same as idPkg namespace.
	XMLName xml.Name `xml:"Document"`

	// Xmlns defines the idPkg namespace prefix.
	// This is required for proper XML namespace handling.
	// Using the full attribute name ensures proper marshaling.
	Xmlns string `xml:"xmlns:idPkg,attr"`

	// Essential attributes that identify the document
	DOMVersion string `xml:"DOMVersion,attr,omitempty"`
	Self       string `xml:"Self,attr,omitempty"`
	Version    string `xml:"Version,attr,omitempty"`

	// Catch-all for all child elements. Phase 2 uses explicit struct fields instead.
	OtherElements []common.RawXMLElement `xml:",any"`
}

// DesignmapMinimal represents an even more minimal struct for testing.
// This version explicitly models some common child elements while still
// preserving unknowns.
//
// DEPRECATED: Phase 1 testing struct. Use Document for Phase 2+.
type DesignmapMinimal struct {
	XMLName    xml.Name `xml:"Document"`
	Xmlns      string   `xml:"xmlns:idPkg,attr,omitempty"`
	DOMVersion string   `xml:"DOMVersion,attr,omitempty"`
	Self       string   `xml:"Self,attr,omitempty"`
	Version    string   `xml:"Version,attr,omitempty"`

	// Properties is a common element in InDesign XML
	Properties *common.Properties `xml:"Properties,omitempty"`

	// Language elements (can have multiple)
	Languages []Language `xml:"Language,omitempty"`

	// Graphic reference using idPkg namespace
	// Note the full namespace in the tag
	Graphic *GraphicRef `xml:"http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging Graphic,omitempty"`

	// Catch-all for all other unknown elements
	OtherElements []common.RawXMLElement `xml:",any"`
}

// GraphicRef represents an idPkg:Graphic reference element.
// DEPRECATED: Phase 1 struct. Use ResourceRef in document.go for Phase 2+.
type GraphicRef struct {
	XMLName xml.Name `xml:"http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging Graphic"`
	Src     string   `xml:"src,attr"`
}

// ParseDesignmap parses a designmap.xml file into a Designmap struct.
// This is a minimal parsing that preserves all content.
//
// Note: Go's encoding/xml doesn't automatically populate namespace declarations
// into struct fields, so we manually extract xmlns:idPkg if present.
func ParseDesignmap(data []byte) (*Designmap, error) {
	var dm Designmap
	if err := xml.Unmarshal(data, &dm); err != nil {
		return nil, err
	}

	// Manually extract xmlns:idPkg namespace if present in the XML
	// This is necessary because encoding/xml doesn't populate namespace declarations
	dataStr := string(data)
	if idx := findSubstringIndex(dataStr, "xmlns:idPkg="); idx != -1 {
		// Find the quote after xmlns:idPkg=
		start := idx + len("xmlns:idPkg=")
		if start < len(dataStr) && (dataStr[start] == '"' || dataStr[start] == '\'') {
			quote := dataStr[start]
			start++
			// Find the closing quote
			end := start
			for end < len(dataStr) && dataStr[end] != byte(quote) {
				end++
			}
			if end < len(dataStr) {
				dm.Xmlns = dataStr[start:end]
			}
		}
	}

	return &dm, nil
}

// findSubstringIndex finds the index of substr in s, returns -1 if not found.
func findSubstringIndex(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// MarshalDesignmap marshals a Designmap struct back to XML bytes.
func MarshalDesignmap(dm *Designmap) ([]byte, error) {
	// Add XML header
	header := []byte(xml.Header)

	data, err := xml.MarshalIndent(dm, "", "\t")
	if err != nil {
		return nil, err
	}

	// Combine header and data
	result := append(header, data...)
	return result, nil
}
