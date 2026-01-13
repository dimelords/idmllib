package xmlutil

import (
	"bytes"
	"encoding/xml"
	"regexp"
	"strings"

	"github.com/dimelords/idmllib/v2/pkg/common"
)

// ProcessingInstruction represents an XML processing instruction like <?aid ...?>
type ProcessingInstruction struct {
	Target string // e.g., "aid"
	Inst   string // e.g., 'style="50" type="document" ...'
}

// Metadata holds XML metadata that needs to be preserved during parsing/marshaling.
type Metadata struct {
	XMLDeclaration         string                  // e.g., 'version="1.0" encoding="UTF-8" standalone="yes"'
	ProcessingInstructions []ProcessingInstruction // Processing instructions like <?aid ...?>
	NamespaceDeclarations  map[string]string       // Namespace prefix to URI mappings
}

// ParseWithMetadata parses XML data into the provided interface while preserving metadata.
// Returns the parsed data and extracted metadata (processing instructions, etc.).
func ParseWithMetadata(data []byte, v interface{}) (*Metadata, error) {
	// Add nil checks for input parameters
	if data == nil {
		return nil, common.Errorf("xmlutil", "parse with metadata", "", "input data is nil")
	}

	if len(data) == 0 {
		return nil, common.Errorf("xmlutil", "parse with metadata", "", "input data is empty")
	}

	if v == nil {
		return nil, common.Errorf("xmlutil", "parse with metadata", "", "target interface is nil")
	}

	// Parse the data normally first
	if err := xml.Unmarshal(data, v); err != nil {
		return nil, common.WrapError("xmlutil", "parse with metadata", err)
	}

	// Extract metadata
	metadata := &Metadata{
		NamespaceDeclarations: make(map[string]string),
	}

	// Extract XML declaration
	xmlDeclRegex := regexp.MustCompile(`<\?xml\s+([^?]*)\?>`)
	if match := xmlDeclRegex.FindSubmatch(data); match != nil {
		metadata.XMLDeclaration = string(match[1])
	}

	// Extract processing instructions (excluding xml declaration)
	piRegex := regexp.MustCompile(`<\?(\w+)\s+([^?]+)\?>`)
	matches := piRegex.FindAllSubmatch(data, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			target := string(match[1])
			// Skip xml declaration (already captured)
			if target == "xml" {
				continue
			}
			// Trim trailing whitespace from instruction
			inst := strings.TrimRight(string(match[2]), " \t")
			metadata.ProcessingInstructions = append(metadata.ProcessingInstructions, ProcessingInstruction{
				Target: target,
				Inst:   inst,
			})
		}
	}

	// Extract namespace declarations from root element
	rootElemRegex := regexp.MustCompile(`<[^>]*xmlns:([^=]+)="([^"]*)"[^>]*>`)
	nsMatches := rootElemRegex.FindAllSubmatch(data, -1)
	for _, match := range nsMatches {
		if len(match) >= 3 {
			prefix := string(match[1])
			uri := string(match[2])
			metadata.NamespaceDeclarations[prefix] = uri
		}
	}

	return metadata, nil
}

// MarshalWithMetadata marshals the provided interface to XML with preserved metadata.
// The metadata includes processing instructions and namespace declarations.
func MarshalWithMetadata(v interface{}, metadata *Metadata) ([]byte, error) {
	var buf bytes.Buffer

	// Add XML declaration (use preserved if available, otherwise use default)
	if metadata != nil && metadata.XMLDeclaration != "" {
		buf.WriteString(`<?xml `)
		buf.WriteString(metadata.XMLDeclaration)
		buf.WriteString(`?>`)
	} else {
		buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	}
	buf.WriteByte('\n')

	// Add processing instructions
	if metadata != nil {
		for _, pi := range metadata.ProcessingInstructions {
			buf.WriteString(`<?`)
			buf.WriteString(pi.Target)
			buf.WriteByte(' ')
			buf.WriteString(pi.Inst)
			buf.WriteString(` ?>`)
			buf.WriteByte('\n')
		}
	}

	// Marshal the content (WITHOUT xml.Header prefix since we added our own)
	contentXML, err := xml.MarshalIndent(v, "", "\t")
	if err != nil {
		return nil, common.WrapError("xmlutil", "marshal with metadata", err)
	}

	// Apply namespace prefix fixes if needed
	if metadata != nil && len(metadata.NamespaceDeclarations) > 0 {
		contentXML = fixNamespacePrefixes(contentXML, metadata.NamespaceDeclarations)
	}

	// Convert empty elements to self-closing tags (IDML format)
	contentXML = CompactEmptyElements(contentXML)

	buf.Write(contentXML)
	return buf.Bytes(), nil
}

// MarshalWithHeader marshals the provided interface to XML with standard XML header.
// This is a convenience function for cases where no special metadata is needed.
func MarshalWithHeader(v interface{}) ([]byte, error) {
	return MarshalWithMetadata(v, nil)
}

// MarshalIndentWithHeader marshals the provided interface to XML with custom indentation and header.
func MarshalIndentWithHeader(v interface{}, prefix, indent string) ([]byte, error) {
	var buf bytes.Buffer

	// Add XML declaration
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	buf.WriteByte('\n')

	// Marshal with custom indentation
	contentXML, err := xml.MarshalIndent(v, prefix, indent)
	if err != nil {
		return nil, common.WrapError("xmlutil", "marshal indent with header", err)
	}

	// Convert empty elements to self-closing tags
	contentXML = CompactEmptyElements(contentXML)

	buf.Write(contentXML)
	buf.WriteByte('\n')

	return buf.Bytes(), nil
}

// fixNamespacePrefixes converts full namespace declarations back to prefixed format.
// This handles the common IDML pattern where Go's xml.Marshal adds xmlns="..."
// to each element, but IDML expects prefixed elements like idPkg:ElementName.
func fixNamespacePrefixes(xmlData []byte, namespaceDeclarations map[string]string) []byte {
	xmlStr := string(xmlData)

	// CRITICAL FIX: Remove the broken xmlns:_xmlns="xmlns" declaration
	// and fix _xmlns:prefix back to xmlns:prefix
	// This is a Go XML marshaler bug when using xml:"xmlns:prefix,attr"
	xmlStr = regexp.MustCompile(`xmlns:_xmlns="xmlns"\s+`).ReplaceAllString(xmlStr, "")

	for prefix := range namespaceDeclarations {
		xmlStr = strings.ReplaceAll(xmlStr, "_xmlns:"+prefix, "xmlns:"+prefix)
	}

	// For each namespace declaration, fix elements that use that namespace
	for prefix, nsURL := range namespaceDeclarations {
		// Fix opening tags with namespace
		re1 := regexp.MustCompile(`<(\w+)\s+xmlns="` + regexp.QuoteMeta(nsURL) + `"`)
		xmlStr = re1.ReplaceAllString(xmlStr, `<`+prefix+`:$1`)

		// Fix self-closing tags with namespace
		re2 := regexp.MustCompile(`<(\w+)\s+xmlns="` + regexp.QuoteMeta(nsURL) + `"\s+`)
		xmlStr = re2.ReplaceAllString(xmlStr, `<`+prefix+`:$1 `)

		// Fix closing tags - this is more complex as we need to match prefixed opening tags
		// For IDML, handle common elements
		if prefix == "idPkg" {
			idmlElements := []string{
				"Graphic", "Fonts", "Styles", "Preferences", "Tags",
				"MasterSpread", "Spread", "Story", "BackingStory",
			}

			for _, elem := range idmlElements {
				xmlStr = strings.ReplaceAll(xmlStr, "</"+elem+">", "</"+prefix+":"+elem+">")
			}
		}
	}

	return []byte(xmlStr)
}

// ExtractProcessingInstructions extracts processing instructions from XML data.
// This is useful when you only need the processing instructions without parsing the full document.
func ExtractProcessingInstructions(data []byte) ([]ProcessingInstruction, error) {
	var instructions []ProcessingInstruction

	// Extract processing instructions (excluding xml declaration)
	piRegex := regexp.MustCompile(`<\?(\w+)\s+([^?]+)\?>`)
	matches := piRegex.FindAllSubmatch(data, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			target := string(match[1])
			// Skip xml declaration
			if target == "xml" {
				continue
			}
			// Trim trailing whitespace from instruction
			inst := strings.TrimRight(string(match[2]), " \t")
			instructions = append(instructions, ProcessingInstruction{
				Target: target,
				Inst:   inst,
			})
		}
	}

	return instructions, nil
}

// PreserveProcessingInstructions adds processing instructions to XML data.
// This is useful for adding processing instructions to already-marshaled XML.
func PreserveProcessingInstructions(xmlData []byte, instructions []ProcessingInstruction) []byte {
	var buf bytes.Buffer

	// Find the XML declaration and preserve it
	xmlDeclRegex := regexp.MustCompile(`<\?xml[^?]*\?>\s*`)
	xmlDecl := xmlDeclRegex.Find(xmlData)
	if xmlDecl != nil {
		buf.Write(xmlDecl)
		// Remove declaration from original data
		xmlData = xmlDeclRegex.ReplaceAll(xmlData, []byte{})
	}

	// Add processing instructions
	for _, pi := range instructions {
		buf.WriteString(`<?`)
		buf.WriteString(pi.Target)
		buf.WriteByte(' ')
		buf.WriteString(pi.Inst)
		buf.WriteString(` ?>`)
		buf.WriteByte('\n')
	}

	// Add the rest of the XML
	buf.Write(bytes.TrimLeft(xmlData, " \t\n\r"))

	return buf.Bytes()
}
