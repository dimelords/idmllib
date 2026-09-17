package idms

import (
	"bytes"
	"fmt"
	"io"

	"github.com/dimelords/idmllib/v3/pkg/common"
	"github.com/dimelords/idmllib/v3/pkg/document"
)

// Marshal serializes an IDMS Package to XML bytes.
func Marshal(pkg *Package) ([]byte, error) {
	if err := pkg.Validate(); err != nil {
		return nil, common.WrapError("idms", "validate", err)
	}

	var buf bytes.Buffer

	// 1. XML Declaration
	buf.WriteString(pkg.XMLDeclaration)
	buf.WriteString("\n")

	// 2. AID Processing Instructions
	for _, pi := range pkg.AIDProcessingInstructions {
		buf.WriteString("<?")
		buf.WriteString(pi.Target)
		buf.WriteString(" ")
		buf.WriteString(pi.Inst)
		buf.WriteString(" ?>") // Space before ?> to match InDesign format
		buf.WriteString("\n")
	}

	// 3. Document Element
	docXML, err := document.MarshalDocument(pkg.Document)
	if err != nil {
		return nil, common.WrapError("idms", "marshal", fmt.Errorf("marshal document: %w", err))
	}

	// Remove the XML declaration from docXML (it has its own)
	docXMLStr := string(docXML)
	if len(docXMLStr) > 0 && docXMLStr[0] == '<' && docXMLStr[1] == '?' {
		// Find the end of the XML declaration
		if idx := bytes.Index(docXML, []byte("?>")); idx != -1 {
			docXML = docXML[idx+2:]
			// Trim leading whitespace
			docXML = bytes.TrimLeft(docXML, " \t\n\r")
		}
	}

	// 4. Insert XMP Metadata BEFORE closing </Document> tag (if present)
	if pkg.XMPMetadata != "" {
		// Find the closing </Document> tag
		closingTag := []byte("</Document>")
		idx := bytes.LastIndex(docXML, closingTag)
		if idx != -1 {
			// Insert XMP before </Document>
			buf.Write(docXML[:idx])
			buf.WriteString("\n")
			buf.WriteString(pkg.XMPMetadata)
			buf.WriteString("\n")
			buf.Write(docXML[idx:]) // Write </Document>
		} else {
			// Fallback: if </Document> not found, write as before
			buf.Write(docXML)
			buf.WriteString("\n")
			buf.WriteString(pkg.XMPMetadata)
		}
	} else {
		buf.Write(docXML)
	}

	return buf.Bytes(), nil
}

// Write marshals the snippet and writes it to path. The file is written to a
// temporary file next to path and renamed into place, so a failure never
// leaves a truncated snippet behind. The file is created with mode 0600.
func Write(pkg *Package, path string) error {
	data, err := Marshal(pkg)
	if err != nil {
		return common.WrapErrorWithPath("idms", "marshal", path, err)
	}
	err = common.WriteFileAtomic(path, 0o600, func(w io.Writer) error {
		_, err := w.Write(data)
		return err
	})
	if err != nil {
		return common.WrapErrorWithPath("idms", "write", path, err)
	}
	return nil
}
