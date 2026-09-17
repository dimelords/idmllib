package document

import (
	"github.com/dimelords/idmllib/v2/internal/xmlutil"
)

// ProcessingInstruction represents an XML processing instruction like <?aid ...?>
type ProcessingInstruction struct {
	Target string // e.g., "aid"
	Inst   string // e.g., 'style="50" type="document" ...'
}

// File wraps Document with additional metadata that doesn't
// fit into the standard xml.Unmarshal/Marshal flow.
type File struct {
	*Document
	XMLDeclaration         string // e.g., '<?xml version="1.0" encoding="UTF-8" standalone="yes"?>'
	ProcessingInstructions []ProcessingInstruction
}

// ParseFile parses designmap.xml and preserves processing instructions.
func ParseFile(data []byte) (*File, error) {
	// First, parse the document normally
	doc, err := ParseDocument(data)
	if err != nil {
		return nil, err
	}

	// Extract metadata using shared utilities
	metadata, err := xmlutil.ParseWithMetadata(data, &struct{}{}) // We already parsed doc, just need metadata
	if err != nil {
		return nil, err
	}

	result := &File{
		Document:               doc,
		XMLDeclaration:         metadata.XMLDeclaration,
		ProcessingInstructions: make([]ProcessingInstruction, len(metadata.ProcessingInstructions)),
	}

	// Convert from xmlutil.ProcessingInstruction to document.ProcessingInstruction
	for i, pi := range metadata.ProcessingInstructions {
		result.ProcessingInstructions[i] = ProcessingInstruction{
			Target: pi.Target,
			Inst:   pi.Inst,
		}
	}

	return result, nil
}

// MarshalFile marshals a Document back to XML with preserved metadata.
func MarshalFile(docMeta *File) ([]byte, error) {
	// Convert document.ProcessingInstruction to xmlutil.ProcessingInstruction
	metadata := &xmlutil.Metadata{
		XMLDeclaration:         docMeta.XMLDeclaration,
		ProcessingInstructions: make([]xmlutil.ProcessingInstruction, len(docMeta.ProcessingInstructions)),
		NamespaceDeclarations:  map[string]string{"idPkg": "http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging"},
	}

	for i, pi := range docMeta.ProcessingInstructions {
		metadata.ProcessingInstructions[i] = xmlutil.ProcessingInstruction{
			Target: pi.Target,
			Inst:   pi.Inst,
		}
	}

	// Use shared utility to marshal with metadata
	return xmlutil.MarshalWithMetadata(docMeta.Document, metadata)
}
