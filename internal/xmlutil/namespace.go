package xmlutil

import (
	"bytes"
	"encoding/xml"
	"fmt"

	"github.com/dimelords/idmllib/v2/pkg/common"
)

// NamespaceConfig holds configuration for namespace handling.
type NamespaceConfig struct {
	Prefix string // Namespace prefix (e.g., "idPkg")
	URI    string // Namespace URI (e.g., "http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging")
}

// IDMLNamespace returns the standard IDML packaging namespace configuration.
func IDMLNamespace() NamespaceConfig {
	return NamespaceConfig{
		Prefix: "idPkg",
		URI:    "http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging",
	}
}

// ParseWithNamespace parses XML data into the provided interface, handling namespace wrappers.
// This is designed for IDML files that use idPkg namespace wrappers around content.
//
// The function expects XML in the format:
//
//	<idPkg:ElementName xmlns:idPkg="..." DOMVersion="...">
//	  <ElementName>...</ElementName>
//	</idPkg:ElementName>
//
// It extracts the DOMVersion and unmarshals the inner content into the provided interface.
func ParseWithNamespace(data []byte, v interface{}, config NamespaceConfig) (string, error) {
	// Add nil checks for input parameters
	if data == nil {
		return "", common.Errorf("xmlutil", "parse with namespace", "", "input data is nil")
	}

	if len(data) == 0 {
		return "", common.Errorf("xmlutil", "parse with namespace", "", "input data is empty")
	}

	if v == nil {
		return "", common.Errorf("xmlutil", "parse with namespace", "", "target interface is nil")
	}

	decoder := xml.NewDecoder(bytes.NewReader(data))

	var domVersion string

	// Find the wrapper element
	for {
		token, err := decoder.Token()
		if err != nil {
			return "", fmt.Errorf("failed to find namespace wrapper: %w", err)
		}

		if start, ok := token.(xml.StartElement); ok {
			expectedName := config.Prefix + ":" + getElementName(v)
			if start.Name.Local == expectedName ||
				(start.Name.Local == getElementName(v) && start.Name.Space == config.URI) {

				// Extract DOMVersion attribute
				for _, attr := range start.Attr {
					if attr.Name.Local == "DOMVersion" {
						domVersion = attr.Value
						break
					}
				}

				// Parse the content using the decoder positioned at the wrapper
				if err := decoder.DecodeElement(v, &start); err != nil {
					return "", fmt.Errorf("failed to decode element content: %w", err)
				}

				return domVersion, nil
			}
		}
	}
}

// MarshalWithNamespace marshals the provided interface to XML with namespace wrapper.
// This creates XML in the IDML format with idPkg namespace wrapper.
//
// The output format is:
//
//	<idPkg:ElementName xmlns:idPkg="..." DOMVersion="...">
//	  <ElementName>...</ElementName>
//	</idPkg:ElementName>
func MarshalWithNamespace(v interface{}, config NamespaceConfig, domVersion string) ([]byte, error) {
	var buf bytes.Buffer
	encoder := xml.NewEncoder(&buf)

	elementName := getElementName(v)
	wrapperName := config.Prefix + ":" + elementName

	// Create wrapper element
	wrapper := xml.StartElement{
		Name: xml.Name{Local: wrapperName},
		Attr: []xml.Attr{
			{
				Name:  xml.Name{Local: "xmlns:" + config.Prefix},
				Value: config.URI,
			},
			{
				Name:  xml.Name{Local: "DOMVersion"},
				Value: domVersion,
			},
		},
	}

	// Start wrapper element
	if err := encoder.EncodeToken(wrapper); err != nil {
		return nil, fmt.Errorf("failed to encode wrapper start: %w", err)
	}

	// Encode the inner content
	if err := encoder.Encode(v); err != nil {
		return nil, fmt.Errorf("failed to encode inner content: %w", err)
	}

	// End wrapper element
	if err := encoder.EncodeToken(wrapper.End()); err != nil {
		return nil, fmt.Errorf("failed to encode wrapper end: %w", err)
	}

	if err := encoder.Flush(); err != nil {
		return nil, fmt.Errorf("failed to flush encoder: %w", err)
	}

	return buf.Bytes(), nil
}

// ApplyNamespace applies namespace configuration to raw XML data.
// This is useful for transforming XML that doesn't have proper namespace declarations.
func ApplyNamespace(data []byte, config NamespaceConfig) ([]byte, error) {
	// Parse the XML to ensure it's valid
	var temp interface{}
	if err := xml.Unmarshal(data, &temp); err != nil {
		return nil, fmt.Errorf("invalid XML data: %w", err)
	}

	// For now, return the data as-is since this is a complex transformation
	// that would require parsing and rebuilding the entire XML structure.
	// This can be enhanced later if needed.
	return data, nil
}

// getElementName extracts the XML element name from a struct type.
// This uses reflection to find the xml tag or struct name.
func getElementName(v interface{}) string {
	// This is a simplified implementation that assumes the element name
	// matches common IDML patterns. A more sophisticated version would
	// use reflection to examine xml tags.

	switch v.(type) {
	case *interface{}:
		// For generic interfaces, we can't determine the name
		return "Element"
	default:
		// For now, return a generic name. This should be enhanced
		// to use reflection to get the actual struct name or xml tag.
		return "Element"
	}
}

// NamespaceWrapper provides a generic way to handle namespace-wrapped XML.
type NamespaceWrapper struct {
	XMLName    xml.Name `xml:""`
	DOMVersion string   `xml:"DOMVersion,attr"`
	Content    []byte   `xml:",innerxml"`
}

// ParseNamespaceWrapper parses XML with a namespace wrapper and returns the inner content.
func ParseNamespaceWrapper(data []byte) (*NamespaceWrapper, error) {
	var wrapper NamespaceWrapper
	if err := xml.Unmarshal(data, &wrapper); err != nil {
		return nil, common.WrapError("xmlutil", "parse namespace wrapper", err)
	}
	return &wrapper, nil
}

// MarshalNamespaceWrapper creates XML with a namespace wrapper around the provided content.
func MarshalNamespaceWrapper(elementName string, config NamespaceConfig, domVersion string, content []byte) ([]byte, error) {
	wrapper := NamespaceWrapper{
		XMLName:    xml.Name{Local: config.Prefix + ":" + elementName},
		DOMVersion: domVersion,
		Content:    content,
	}

	data, err := xml.Marshal(wrapper)
	if err != nil {
		return nil, common.WrapError("xmlutil", "marshal namespace wrapper", err)
	}

	return data, nil
}
