package idml

import (
	"archive/zip"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/dimelords/idmllib/v3/pkg/common"
	"github.com/dimelords/idmllib/v3/pkg/document"
	"github.com/dimelords/idmllib/v3/pkg/resources"
	"github.com/dimelords/idmllib/v3/pkg/spread"
	"github.com/dimelords/idmllib/v3/pkg/story"
)

// marshalCachedObjects marshals all cached objects back to XML data.
// This ensures any modifications to parsed structs are saved.
func (p *Package) marshalCachedObjects() error {
	// Update XMP metadata in META-INF/metadata.xml if it exists and has been modified
	if err := p.updateXMPInMetadataFile(); err != nil {
		return err
	}

	// If document was parsed, marshal it back to XML with preserved metadata
	if p.documentMetadata != nil && p.isDirty(PathDesignmap) {
		xmlData, err := document.MarshalFile(p.documentMetadata)
		if err != nil {
			return common.WrapErrorWithPath("idml", "marshal document", PathDesignmap, err)
		}
		p.setFileData(PathDesignmap, xmlData)
	}

	// If stories were parsed, marshal them back to XML
	for filename, st := range p.stories {
		if !p.isDirty(filename) {
			continue
		}
		xmlData, err := story.MarshalStory(st)
		if err != nil {
			return common.WrapErrorWithPath("idml", "marshal story", filename, err)
		}
		p.setFileData(filename, xmlData)
	}

	// If spreads were parsed, marshal them back to XML
	for filename, sp := range p.spreads {
		if !p.isDirty(filename) {
			continue
		}
		xmlData, err := spread.MarshalSpread(sp)
		if err != nil {
			return common.WrapErrorWithPath("idml", "marshal spread", filename, err)
		}
		p.setFileData(filename, p.restoreSpreadData(filename, xmlData))
	}

	// If typed fonts were parsed, marshal them back to XML
	if p.fonts != nil && p.isDirty(PathFonts) {
		xmlData, err := resources.MarshalFontsFile(p.fonts)
		if err != nil {
			return common.WrapErrorWithPath("idml", "marshal fonts", PathFonts, err)
		}
		p.setFileData(PathFonts, xmlData)
	}

	// If typed graphics were parsed, marshal them back to XML
	if p.graphics != nil && p.isDirty(PathGraphic) {
		xmlData, err := resources.MarshalGraphicFile(p.graphics)
		if err != nil {
			return common.WrapErrorWithPath("idml", "marshal graphics", PathGraphic, err)
		}
		p.setFileData(PathGraphic, xmlData)
	}

	// If typed styles were parsed, marshal them back to XML
	if p.styles != nil && p.isDirty(PathStyles) {
		xmlData, err := resources.MarshalStylesFile(p.styles)
		if err != nil {
			return common.WrapErrorWithPath("idml", "marshal styles", PathStyles, err)
		}
		p.setFileData(PathStyles, xmlData)
	}

	if p.preferences != nil && p.isDirty(PathPreferences) {
		xmlData, err := resources.MarshalPreferencesFile(p.preferences)
		if err != nil {
			return common.WrapErrorWithPath("idml", "marshal preferences", PathPreferences, err)
		}
		p.setFileData(PathPreferences, xmlData)
	}

	if p.tags != nil && p.isDirty(PathTags) {
		xmlData, err := resources.MarshalTagsFile(p.tags)
		if err != nil {
			return common.WrapErrorWithPath("idml", "marshal tags", PathTags, err)
		}
		p.setFileData(PathTags, xmlData)
	}

	// If metadata files were parsed, marshal them back
	for filename, metadata := range p.metadata {
		data, err := marshalMetadataFile(metadata)
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
		return nil //nolint:nilerr // absent metadata.xml means there is nothing to update
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

		entry, ok := pkg.files[name]
		if !ok || entry == nil {
			continue // Skip missing files
		}

		if err := writeEntry(w, name, entry); err != nil {
			return err
		}
	}

	return nil
}

// Write writes an IDML package to a file.
//
// The function marshals every object marked modified back to XML, then writes
// the ZIP with the mimetype entry first and uncompressed, as the IDML
// specification requires. The archive is written to a temporary file next to
// path and renamed into place, so a failure never leaves a truncated document
// behind. The file is created with mode 0600.
func Write(pkg *Package, path string) error {
	if err := pkg.marshalCachedObjects(); err != nil {
		return err
	}
	err := common.WriteFileAtomic(path, 0o600, func(w io.Writer) error {
		zw := zip.NewWriter(w)
		if err := writeZipFiles(zw, pkg); err != nil {
			return err
		}
		return zw.Close()
	})
	if err != nil {
		return common.WrapErrorWithPath("idml", "write", path, err)
	}
	return nil
}

// defaultHeader is used for files added through the API, which have no header
// from a source archive.
func defaultHeader(name string) *zip.FileHeader {
	return &zip.FileHeader{
		Name:     name,
		Method:   zip.Deflate, // everything except mimetype is compressed
		Modified: time.Now(),
	}
}
