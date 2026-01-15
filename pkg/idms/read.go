package idms

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/dimelords/idmllib/v2/pkg/common"
	"github.com/dimelords/idmllib/v2/pkg/document"
)

// Read reads an IDMS file from the given path.
func Read(path string) (*Package, error) {
	// #nosec G304 - This is a library function; file path is intentionally provided by caller
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, common.WrapErrorWithPath("idms", "read", path, err)
	}
	return Parse(data)
}

// Parse parses IDMS XML data into a Package.
func Parse(data []byte) (*Package, error) {
	pkg := &Package{
		rawXML: data,
	}

	// Step 1: Extract XMP metadata (it's embedded in the Document element)
	pkg.XMPMetadata = extractXMPMetadata(string(data))

	// Step 2: Parse XML with decoder to extract PIs and Document
	decoder := xml.NewDecoder(bytes.NewReader(data))

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, common.WrapError("idms", "parse", err)
		}

		switch t := token.(type) {
		case xml.ProcInst:
			// Handle processing instructions
			if t.Target == "xml" {
				// XML declaration
				pkg.XMLDeclaration = fmt.Sprintf("<?xml %s?>", string(t.Inst))
			} else if t.Target == "aid" {
				// AID processing instruction
				// Trim trailing whitespace (original file may have varying spacing before ?>)
				inst := strings.TrimRight(string(t.Inst), " \t")
				pkg.AIDProcessingInstructions = append(pkg.AIDProcessingInstructions, document.ProcessingInstruction{
					Target: t.Target,
					Inst:   inst,
				})
			}

		case xml.StartElement:
			// Parse the Document element
			if t.Name.Local == "Document" {
				// Extract the Document XML
				var docBuf bytes.Buffer
				encoder := xml.NewEncoder(&docBuf)
				if err := encoder.EncodeToken(token); err != nil {
					return nil, common.WrapError("idms", "parse", fmt.Errorf("encode token: %w", err))
				}

				// Read until the closing Document tag
				depth := 1
				for depth > 0 {
					t, err := decoder.Token()
					if err != nil {
						return nil, common.WrapError("idms", "parse", fmt.Errorf("parse Document element: %w", err))
					}
					if err := encoder.EncodeToken(t); err != nil {
						return nil, common.WrapError("idms", "parse", fmt.Errorf("encode token: %w", err))
					}

					switch t.(type) {
					case xml.StartElement:
						depth++
					case xml.EndElement:
						depth--
					}
				}
				encoder.Flush()

				// Remove XMP metadata from Document XML before parsing
				docXML := docBuf.Bytes()
				if pkg.XMPMetadata != "" {
					xmpPattern := regexp.MustCompile(`(?s)<\?xpacket begin.*?<\?xpacket end[^>]*\?>`)
					docXML = xmpPattern.ReplaceAll(docXML, []byte(""))
				}

				// Parse the Document using document.ParseDocument
				doc, err := document.ParseDocument(docXML)
				if err != nil {
					return nil, common.WrapError("idms", "parse", fmt.Errorf("parse IDMS Document: %w", err))
				}
				pkg.Document = doc

				// We're done after parsing the Document
				break
			}
		}
	}

	return pkg, nil
}

// extractXMPMetadata extracts the XMP packet from IDMS XML.
// Returns empty string if no XMP packet is found.
func extractXMPMetadata(xmlContent string) string {
	// Match from <?xpacket begin to <?xpacket end
	xmpPattern := regexp.MustCompile(`(?s)<\?xpacket begin.*?<\?xpacket end[^>]*\?>`)
	return xmpPattern.FindString(xmlContent)
}
