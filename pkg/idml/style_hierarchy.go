package idml

import (
	"bytes"
	"encoding/xml"
	"strings"
)

// StyleInfo represents basic information about a style and its parent.
type StyleInfo struct {
	Self    string // The style ID (e.g., "ParagraphStyle/MyStyle")
	BasedOn string // The parent style ID this is based on
}

// ParseStylesForHierarchy parses the Styles resource bytes to extract style hierarchy information.
// This is used by the DependencyTracker to resolve BasedOn relationships.
func ParseStylesForHierarchy(data []byte) ([]StyleInfo, error) {
	var styles []StyleInfo

	// Parse the raw content to extract style definitions
	decoder := xml.NewDecoder(bytes.NewReader(data))

	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}

		switch t := tok.(type) {
		case xml.StartElement:
			var styleInfo StyleInfo
			var foundStyle bool

			// Check if this is a style element
			switch t.Name.Local {
			case "ParagraphStyle", "CharacterStyle", "ObjectStyle":
				foundStyle = true
			}

			if !foundStyle {
				continue
			}

			// Extract Self attribute
			for _, attr := range t.Attr {
				if attr.Name.Local == "Self" {
					styleInfo.Self = attr.Value
					break
				}
			}

			// If no Self attribute, skip this style
			if styleInfo.Self == "" {
				continue
			}

			// Look for BasedOn in Properties
			basedOn := parseBasedOnFromProperties(decoder)
			styleInfo.BasedOn = basedOn

			styles = append(styles, styleInfo)
		}
	}

	return styles, nil
}

// parseBasedOnFromProperties reads the Properties element and extracts the BasedOn value.
func parseBasedOnFromProperties(decoder *xml.Decoder) string {
	depth := 0
	var basedOn string

	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}

		switch t := tok.(type) {
		case xml.StartElement:
			depth++
			if t.Name.Local == "BasedOn" {
				// Read the BasedOn value
				for {
					contentTok, err := decoder.Token()
					if err != nil {
						break
					}

					switch content := contentTok.(type) {
					case xml.CharData:
						basedOn = string(content)
					case xml.EndElement:
						if content.Name.Local == "BasedOn" {
							goto done
						}
					}
				}
			}

		case xml.EndElement:
			depth--
			// Exit when we've closed all elements we opened
			if depth < 0 {
				goto done
			}
		}
	}

done:
	return strings.TrimSpace(basedOn)
}
