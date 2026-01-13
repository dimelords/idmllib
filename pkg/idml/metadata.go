package idml

// MetadataFile represents optional metadata files in an IDML package.
// These files include:
//   - META-INF/container.xml: Package metadata (OASIS container specification)
//   - META-INF/metadata.xml: XMP metadata (Dublin Core, creation dates, etc.)
//   - XML/Tags.xml: XML tag definitions for structured content
//   - XML/BackingStory.xml: Default story for XML content
//
// We use a preservation strategy (storing raw XML) to ensure perfect roundtrips
// without needing to model every complex XMP/RDF structure.
type MetadataFile struct {
	// Filename is the path within the IDML package (e.g., "META-INF/container.xml")
	Filename string

	// RawContent stores the complete file content as-is
	// This preserves all XML structure, namespaces, and formatting
	RawContent []byte
}

// ParseMetadataFile parses a metadata file using the preservation strategy.
// The file content is stored as-is for perfect roundtrip fidelity.
func ParseMetadataFile(filename string, data []byte) (*MetadataFile, error) {
	return &MetadataFile{
		Filename:   filename,
		RawContent: data,
	}, nil
}

// MarshalMetadataFile marshals a metadata file back to its original form.
// Since we use the preservation strategy, this simply returns the raw content.
func MarshalMetadataFile(mf *MetadataFile) ([]byte, error) {
	return mf.RawContent, nil
}
