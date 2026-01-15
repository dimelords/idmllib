// Package xmp provides XMP (Extensible Metadata Platform) metadata support
// for IDML and IDMS files. XMP is Adobe's standard for embedding metadata
// in documents.
package xmp

import (
	"errors"
	"fmt"
	"regexp"
	"time"
)

// Metadata represents XMP metadata that can be embedded in IDML or IDMS files.
// XMP is Adobe's standard for embedding metadata in documents.
type Metadata struct {
	raw string // The complete XMP packet including processing instructions
}

// Parse creates an XMP Metadata instance from a raw XMP string.
// The XMP string typically includes <?xpacket begin...?> and <?xpacket end...?>
// processing instructions wrapping the <x:xmpmeta> element.
//
// Parse accepts any string, including empty strings, and returns a non-nil
// *Metadata instance in all cases.
func Parse(xmpString string) *Metadata {
	return &Metadata{raw: xmpString}
}

// String returns the XMP metadata as a string, including processing instructions.
// This allows the modified XMP to be retrieved after operations.
func (x *Metadata) String() string {
	return x.raw
}

// IsEmpty returns true if there is no XMP metadata.
// Used to check if XMP metadata exists before operations.
func (x *Metadata) IsEmpty() bool {
	return x.raw == ""
}

var (
	modifyDatePattern   = regexp.MustCompile(`<xmp:ModifyDate>[^<]+</xmp:ModifyDate>`)
	metadataDatePattern = regexp.MustCompile(`<xmp:MetadataDate>[^<]+</xmp:MetadataDate>`)
	thumbnailPattern    = regexp.MustCompile(`(?s)\s*<xmp:Thumbnails>.*?</xmp:Thumbnails>`)
)

// UpdateTimestamps updates xmp:ModifyDate and xmp:MetadataDate to the current time.
// The xmp:CreateDate field is preserved unchanged.
// Returns an error if no XMP metadata exists.
func (x *Metadata) UpdateTimestamps() error {
	if x.raw == "" {
		return errors.New("no XMP metadata")
	}

	// Get current timestamp in ISO 8601 format with timezone
	currentTime := time.Now().Format("2006-01-02T15:04:05-07:00")

	// Update xmp:ModifyDate
	x.raw = modifyDatePattern.ReplaceAllString(x.raw, "<xmp:ModifyDate>"+currentTime+"</xmp:ModifyDate>")

	// Update xmp:MetadataDate
	x.raw = metadataDatePattern.ReplaceAllString(x.raw, "<xmp:MetadataDate>"+currentTime+"</xmp:MetadataDate>")

	return nil
}

// RemoveThumbnails removes the xmp:Thumbnails section from the metadata.
// This is useful when content has been modified and thumbnails are outdated.
// Removing thumbnails can reduce file size by 20-30KB.
// Returns an error if no XMP metadata exists.
func (x *Metadata) RemoveThumbnails() error {
	if x.raw == "" {
		return errors.New("no XMP metadata")
	}

	// Remove the entire xmp:Thumbnails section including nested content and surrounding whitespace
	x.raw = thumbnailPattern.ReplaceAllString(x.raw, "")

	return nil
}

// AddThumbnail adds a thumbnail to the XMP metadata.
// The thumbnailData should be a base64-encoded JPEG image.
// The thumbnail will be added with default dimensions of 512x512.
// Returns an error if no XMP metadata exists.
func (x *Metadata) AddThumbnail(thumbnailData string, width, height int) error {
	if x.raw == "" {
		return errors.New("no XMP metadata")
	}

	// Check if thumbnails already exist and remove them first
	if thumbnailPattern.MatchString(x.raw) {
		x.raw = thumbnailPattern.ReplaceAllString(x.raw, "")
	}

	// Build the thumbnail XML structure
	thumbnailXML := fmt.Sprintf(`<xmp:Thumbnails>
            <rdf:Alt>
               <rdf:li rdf:parseType="Resource">
                  <xmpGImg:format>JPEG</xmpGImg:format>
                  <xmpGImg:width>%d</xmpGImg:width>
                  <xmpGImg:height>%d</xmpGImg:height>
                  <xmpGImg:image>%s</xmpGImg:image>
               </rdf:li>
            </rdf:Alt>
         </xmp:Thumbnails>`, width, height, thumbnailData)

	// Find the position to insert the thumbnail (after xmp:ModifyDate or xmp:MetadataDate)
	insertPattern := regexp.MustCompile(`(</xmp:ModifyDate>|</xmp:MetadataDate>)`)
	matches := insertPattern.FindStringIndex(x.raw)
	
	if matches == nil {
		return errors.New("could not find insertion point for thumbnail")
	}

	// Insert the thumbnail after the matched tag
	insertPos := matches[1]
	x.raw = x.raw[:insertPos] + "\n         " + thumbnailXML + x.raw[insertPos:]

	return nil
}

// GetField retrieves the value of a specific XMP field.
// The fieldName should include the namespace prefix (e.g., "xmp:CreateDate").
// Returns an error if the field is not found or if no XMP metadata exists.
func (x *Metadata) GetField(fieldName string) (string, error) {
	if x.raw == "" {
		return "", errors.New("no XMP metadata")
	}

	// Build regex pattern to extract field value
	// Use regexp.QuoteMeta to escape special characters in field name
	pattern := regexp.MustCompile(fmt.Sprintf(`<%s>([^<]+)</%s>`, 
		regexp.QuoteMeta(fieldName), regexp.QuoteMeta(fieldName)))
	
	matches := pattern.FindStringSubmatch(x.raw)
	if matches == nil {
		return "", fmt.Errorf("field %s not found", fieldName)
	}

	// Return the captured group (field value)
	return matches[1], nil
}

// SetField sets the value of a specific XMP field.
// The fieldName should include the namespace prefix (e.g., "xmp:CreatorTool").
// Returns an error if the field is not found or if no XMP metadata exists.
func (x *Metadata) SetField(fieldName, value string) error {
	if x.raw == "" {
		return errors.New("no XMP metadata")
	}

	// Build regex pattern to find and replace field value
	// Use regexp.QuoteMeta to escape special characters in field name
	pattern := regexp.MustCompile(fmt.Sprintf(`<%s>[^<]+</%s>`, 
		regexp.QuoteMeta(fieldName), regexp.QuoteMeta(fieldName)))
	
	// Check if field exists
	if !pattern.MatchString(x.raw) {
		return fmt.Errorf("field %s not found", fieldName)
	}

	// Replace field value
	replacement := fmt.Sprintf("<%s>%s</%s>", fieldName, value, fieldName)
	x.raw = pattern.ReplaceAllString(x.raw, replacement)

	return nil
}
