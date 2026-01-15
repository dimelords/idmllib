package idml

import (
	"archive/zip"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/dimelords/idmllib/v2/pkg/common"
	"github.com/dimelords/idmllib/v2/pkg/document"
	"github.com/dimelords/idmllib/v2/pkg/resources"
	"github.com/dimelords/idmllib/v2/pkg/spread"
	"github.com/dimelords/idmllib/v2/pkg/story"
)

// marshalCachedObjects marshals all cached objects back to XML data.
// This ensures any modifications to parsed structs are saved.
func (p *Package) marshalCachedObjects() error {
	// Update XMP metadata in META-INF/metadata.xml if it exists and has been modified
	if err := p.updateXMPInMetadataFile(); err != nil {
		return err
	}

	// If document was parsed, marshal it back to XML with preserved metadata
	if p.documentMetadata != nil {
		xmlData, err := document.MarshalDocumentWithMetadata(p.documentMetadata)
		if err != nil {
			return common.WrapErrorWithPath("idml", "marshal document", PathDesignmap, err)
		}
		p.setFileData(PathDesignmap, xmlData)
	}

	// If stories were parsed, marshal them back to XML
	for filename, st := range p.stories {
		xmlData, err := story.MarshalStory(st)
		if err != nil {
			return common.WrapErrorWithPath("idml", "marshal story", filename, err)
		}
		p.setFileData(filename, xmlData)
	}

	// If spreads were parsed, marshal them back to XML
	for filename, sp := range p.spreads {
		xmlData, err := spread.MarshalSpread(sp)
		if err != nil {
			return common.WrapErrorWithPath("idml", "marshal spread", filename, err)
		}
		p.setFileData(filename, xmlData)
	}

	// If resources were parsed, marshal them back to XML
	for filename, resource := range p.resources {
		xmlData, err := MarshalResourceFile(resource)
		if err != nil {
			return common.WrapErrorWithPath("idml", "marshal resource", filename, err)
		}
		p.setFileData(filename, xmlData)
	}

	// If typed fonts were parsed, marshal them back to XML
	if p.fonts != nil {
		xmlData, err := resources.MarshalFontsFile(p.fonts)
		if err != nil {
			return common.WrapErrorWithPath("idml", "marshal fonts", PathFonts, err)
		}
		p.setFileData(PathFonts, xmlData)
	}

	// If typed graphics were parsed, marshal them back to XML
	if p.graphics != nil {
		xmlData, err := resources.MarshalGraphicFile(p.graphics)
		if err != nil {
			return common.WrapErrorWithPath("idml", "marshal graphics", PathGraphic, err)
		}
		p.setFileData(PathGraphic, xmlData)
	}

	// If typed styles were parsed, marshal them back to XML
	if p.styles != nil {
		xmlData, err := resources.MarshalStylesFile(p.styles)
		if err != nil {
			return common.WrapErrorWithPath("idml", "marshal styles", PathStyles, err)
		}
		p.setFileData(PathStyles, xmlData)
	}

	// If metadata files were parsed, marshal them back
	for filename, metadata := range p.metadata {
		data, err := MarshalMetadataFile(metadata)
		if err != nil {
			return common.WrapErrorWithPath("idml", "marshal metadata", filename, err)
		}
		p.setFileData(filename, data)
	}

	return nil
}

// updateXMPInMetadataFile updates the XMP packet in META-INF/metadata.xml.
// This ensures XMP modifications are persisted when writing the package.
func (p *Package) updateXMPInMetadataFile() error {
	// Check if metadata.xml exists
	entry, err := p.getFileEntry("META-INF/metadata.xml")
	if err != nil {
		// If metadata.xml doesn't exist, nothing to update
		return nil
	}

	// Get the current content
	content := string(entry.data)

	// Replace the XMP packet with the updated one
	xmpPattern := regexp.MustCompile(`(?s)<\?xpacket begin.*?<\?xpacket end[^>]*\?>`)
	
	if p.XMPMetadata != "" {
		// Replace existing XMP or add if not present
		if xmpPattern.MatchString(content) {
			content = xmpPattern.ReplaceAllString(content, p.XMPMetadata)
		} else {
			// If no XMP exists, add it before the closing tag
			// Find a suitable insertion point (before </rdf:RDF> or at the end)
			if idx := strings.Index(content, "</rdf:RDF>"); idx != -1 {
				content = content[:idx] + p.XMPMetadata + "\n" + content[idx:]
			} else {
				content = content + "\n" + p.XMPMetadata
			}
		}
	} else {
		// Remove XMP if it's been cleared
		content = xmpPattern.ReplaceAllString(content, "")
	}

	// Update the file data
	p.setFileData("META-INF/metadata.xml", []byte(content))
	return nil
}

// writeZipFiles writes all files to the ZIP archive in the correct order.
func writeZipFiles(w *zip.Writer, pkg *Package) error {
	// Add validation for parameters
	if w == nil {
		return common.Errorf("idml", "write zip files", "", "zip writer is nil")
	}

	if pkg == nil {
		return common.Errorf("idml", "write zip files", "", "package is nil")
	}

	// CRITICAL: Write mimetype first and uncompressed
	if entry, err := pkg.getFileEntry(PathMimetype); err == nil {
		// Validate entry before using
		if entry == nil || entry.header == nil {
			return common.WrapErrorWithPath("idml", "write", PathMimetype, common.Errorf("idml", "write", PathMimetype, "invalid mimetype entry"))
		}

		// Always use Store method for mimetype (CRITICAL requirement)
		// Create a copy of the header to avoid modifying the original
		header := *entry.header
		header.Method = zip.Store // Force uncompressed

		mimeWriter, err := w.CreateHeader(&header)
		if err != nil {
			return common.WrapErrorWithPath("idml", "write", PathMimetype, err)
		}

		// Write mimetype content
		if _, err := mimeWriter.Write(entry.data); err != nil {
			return common.WrapErrorWithPath("idml", "write", PathMimetype, err)
		}
	}

	// Write all other files in original order
	for _, name := range pkg.fileOrder {
		if name == PathMimetype {
			continue // Already written
		}

		entry, err := pkg.getFileEntry(name)
		if err != nil {
			continue // Skip missing files
		}

		// Create header if it doesn't exist (e.g., for newly added files)
		if entry.header == nil {
			entry.header = &zip.FileHeader{
				Name:     name,
				Method:   zip.Deflate, // Use compression for all files except mimetype
				Modified: time.Now(),
			}
		}

		// Use the original FileHeader to preserve compression and metadata
		fileWriter, err := w.CreateHeader(entry.header)
		if err != nil {
			return common.WrapErrorWithPath("idml", "write", name, err)
		}

		if _, err := fileWriter.Write(entry.data); err != nil {
			return common.WrapErrorWithPath("idml", "write", name, err)
		}
	}

	return nil
}

// Write writes an IDML package to a file.
//
// The function:
// 1. Marshals the Document struct back to designmap.xml (if modified)
// 2. Writes mimetype first and uncompressed (CRITICAL IDML requirement)
// 3. Writes all other files in original order
//
// CRITICAL: The mimetype file MUST be written first and MUST be uncompressed.
// This is required by the IDML specification. InDesign will reject files
// that don't follow this requirement.
func Write(pkg *Package, path string) error {
	// Step 1: Marshal all cached objects back to XML
	if err := pkg.marshalCachedObjects(); err != nil {
		return err
	}

	// Step 2: Create the output file
	// #nosec G304 - This is a library function; file path is intentionally provided by caller
	f, err := os.Create(path)
	if err != nil {
		return common.WrapErrorWithPath("idml", "write", path, err)
	}
	defer f.Close()

	// Step 3: Create ZIP writer and write all files
	w := zip.NewWriter(f)
	defer w.Close()

	if err := writeZipFiles(w, pkg); err != nil {
		return err
	}

	// Step 4: Close the ZIP writer (important!)
	if err := w.Close(); err != nil {
		return common.WrapErrorWithPath("idml", "write", path, err)
	}

	return nil
}
