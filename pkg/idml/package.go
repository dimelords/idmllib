package idml

import (
	"archive/zip"
	"io"

	"github.com/dimelords/idmllib/v2/pkg/common"
	"github.com/dimelords/idmllib/v2/pkg/document"
	"github.com/dimelords/idmllib/v2/pkg/resources"
	"github.com/dimelords/idmllib/v2/pkg/spread"
	"github.com/dimelords/idmllib/v2/pkg/story"
	"github.com/dimelords/idmllib/v2/pkg/xmp"
)

// fileEntry stores both the content and ZIP metadata for a file.
type fileEntry struct {
	data   []byte
	header *zip.FileHeader

	// src points at the entry in the source archive while its bytes have not
	// been read yet; see lazy.go.
	src *zip.File
}

// Package represents an IDML document loaded into memory.
// In Phase 1, files are stored as raw bytes with their ZIP metadata preserved.
// In Phase 2, we add structured access to the designmap.xml Document.
//
// DESIGN DECISION: Lazy Parsing with Caching
// The Package uses a lazy parsing strategy where files are only parsed when first accessed,
// then cached for subsequent calls. This provides several benefits:
// 1. Fast initial load times - only reads ZIP structure, doesn't parse all XML
// 2. Memory efficiency - only parses files that are actually used
// 3. Performance - avoids re-parsing the same file multiple times
// 4. Flexibility - supports both raw byte access and structured object access
// The trade-off is slightly more complex code, but the performance benefits are significant
// for large IDML files where only a subset of content is typically accessed.
type Package struct {
	// files maps filenames to their content and metadata.
	// This is intentionally not exported to maintain encapsulation.
	files map[string]*fileEntry

	// fileOrder preserves the original order of files in the ZIP.
	// DESIGN DECISION: Preserve ZIP file order for byte-perfect roundtrip
	// InDesign is sensitive to file order, particularly the mimetype file which
	// must be first and uncompressed. Maintaining original order ensures compatibility.
	fileOrder []string

	// DESIGN DECISION: Dual-level caching strategy
	// We cache both generic (ResourceFile) and typed (FontsFile, StylesFile) versions
	// of resource files. This allows the same file to be accessed through different
	// APIs without re-parsing, while maintaining type safety where needed.

	// document is the parsed designmap.xml Document (Phase 2).
	// This is parsed on demand and cached.
	document *document.Document

	// documentMetadata stores processing instructions and other metadata
	// that needs to be preserved during marshal/unmarshal.
	documentMetadata *document.File

	// stories caches parsed Story files from the Stories/ directory.
	// Map key is the story filename (e.g., "Stories/Story_u1d8.xml").
	stories map[string]*story.File

	// spreads caches parsed Spread files from the Spreads/ directory.
	// Map key is the spread filename (e.g., "Spreads/Spread_u210.xml").
	spreads map[string]*spread.File

	// resources caches parsed Resource files from the Resources/ directory.
	// Map key is the resource filename (e.g., "Resources/Graphic.xml").
	// fonts caches the typed Fonts.xml file (if parsed)
	fonts *resources.FontsFile

	// graphics caches the typed Graphic.xml file (if parsed)
	graphics *resources.GraphicFile

	// styles caches the typed Styles.xml file (if parsed)
	styles *resources.StylesFile

	// preferences caches the parsed Resources/Preferences.xml.
	preferences *resources.PreferencesFile

	// tags caches the parsed XML/Tags.xml.
	tags *resources.TagsFile

	// metadata caches optional metadata files (META-INF/*, XML/*).
	// Map key is the file path (e.g., "META-INF/container.xml").
	metadata map[string]*MetadataFile

	// XMPMetadata contains the XMP packet extracted from designmap.xml.
	// XMP is Adobe's standard for embedding metadata in documents.
	// Example: <?xpacket begin="" id="..."?>...<x:xmpmeta>...</x:xmpmeta><?xpacket end="r"?>
	XMPMetadata string

	// indexState holds the item index for O(1) page item lookups.
	// Built lazily on first SelectXxxByID call.
	indexState itemIndexState

	// streamingMode strips embedded image payloads from parsed spreads; see
	// streaming.go.
	streamingMode bool

	// streamedContents holds those payloads as sub-slices of the raw bytes.
	streamedContents contentsStore

	// Lazy reading state; see lazy.go.
	lazy     bool
	source   string
	readOpts *ReadOptions
	closer   io.Closer

	// dirty tracks which cached files were modified through the API and must be
	// re-marshaled on Write. Untouched files are written back byte for byte.
	dirty map[string]bool
}

// New creates a new empty IDML package.
func New() *Package {
	return &Package{
		files:    make(map[string]*fileEntry),
		stories:  make(map[string]*story.File),
		spreads:  make(map[string]*spread.File),
		metadata: make(map[string]*MetadataFile),
	}
}

// Files returns a copy of all filenames in the package.
// Useful for inspection and debugging.
func (p *Package) Files() []string {
	names := make([]string, 0, len(p.files))
	for name := range p.files {
		names = append(names, name)
	}
	return names
}

// FileCount returns the number of files in the package.
func (p *Package) FileCount() int {
	return len(p.files)
}

// Document returns the parsed designmap.xml Document.
// The document is parsed on first access and cached.
// Returns an error if designmap.xml doesn't exist or can't be parsed.
func (p *Package) Document() (*document.Document, error) {
	// Return cached document if available
	if doc, cached := p.getCachedDocument(); cached {
		return doc, nil
	}

	// Get designmap.xml file
	entry, err := p.getFileEntry(PathDesignmap)
	if err != nil {
		return nil, err
	}

	// Parse the document with metadata (processing instructions, etc.)
	docMeta, err := document.ParseFile(entry.data)
	if err != nil {
		return nil, err
	}

	// Cache for future calls
	p.cacheDocument(docMeta.Document, docMeta)
	return p.document, nil
}

// MetadataFile returns a metadata file by path.
// Metadata files are optional and include:
//   - META-INF/container.xml
//   - META-INF/metadata.xml
//   - XML/Tags.xml
//   - XML/BackingStory.xml
//
// The file is parsed on first access and cached.
// Returns ErrNotFound if the file doesn't exist.
func (p *Package) MetadataFile(path string) (*MetadataFile, error) {
	// Return cached if available
	if mf, cached := p.getCachedMetadata(path); cached {
		return mf, nil
	}

	// Get file
	entry, err := p.getFileEntry(path)
	if err != nil {
		return nil, err
	}

	// Parse the metadata file
	mf, err := parseMetadataFile(path, entry.data)
	if err != nil {
		return nil, common.WrapErrorWithPath("idml", "parse metadata", path, err)
	}

	// Cache for future calls
	p.cacheMetadata(path, mf)
	return mf, nil
}

// MetadataFiles returns all metadata files in the package.
// This includes all files in META-INF/ and XML/ directories.
func (p *Package) MetadataFiles() (map[string]*MetadataFile, error) {
	metadata := make(map[string]*MetadataFile)

	// Find all metadata files (META-INF/* and XML/*)
	for filename := range p.files {
		if IsMetaInfPath(filename) || IsXMLPath(filename) {
			mf, err := p.MetadataFile(filename)
			if err != nil {
				return nil, err
			}
			metadata[filename] = mf
		}
	}

	return metadata, nil
}

// Fonts returns the typed Fonts.xml file.
// The file is parsed on first access and cached.
// Returns an error if the file doesn't exist or can't be parsed.
func (p *Package) Fonts() (*resources.FontsFile, error) {
	// Return cached if available
	if fonts, cached := p.getCachedFonts(); cached {
		return fonts, nil
	}

	// Get Fonts.xml file
	entry, err := p.getFileEntry(PathFonts)
	if err != nil {
		return nil, err
	}

	// Parse the fonts file
	fonts, err := resources.ParseFontsFile(entry.data)
	if err != nil {
		return nil, common.WrapErrorWithPath("idml", "parse fonts", PathFonts, err)
	}

	// Cache for future calls
	p.cacheFonts(fonts)
	return fonts, nil
}

// Graphics returns the typed Graphic.xml file.
// The file is parsed on first access and cached.
// Returns an error if the file doesn't exist or can't be parsed.
func (p *Package) Graphics() (*resources.GraphicFile, error) {
	// Return cached if available
	if graphics, cached := p.getCachedGraphics(); cached {
		return graphics, nil
	}

	// Get Graphic.xml file
	entry, err := p.getFileEntry(PathGraphic)
	if err != nil {
		return nil, err
	}

	// Parse the graphics file
	graphics, err := resources.ParseGraphicFile(entry.data)
	if err != nil {
		return nil, common.WrapErrorWithPath("idml", "parse graphics", PathGraphic, err)
	}

	// Cache for future calls
	p.cacheGraphics(graphics)
	return graphics, nil
}

// Styles returns the typed Styles.xml file.
// The file is parsed on first access and cached.
// Returns an error if the file doesn't exist or can't be parsed.
func (p *Package) Styles() (*resources.StylesFile, error) {
	// Return cached if available
	if styles, cached := p.getCachedStyles(); cached {
		return styles, nil
	}

	// Get Styles.xml file
	entry, err := p.getFileEntry(PathStyles)
	if err != nil {
		return nil, err
	}

	// Parse the styles file
	styles, err := resources.ParseStylesFile(entry.data)
	if err != nil {
		return nil, common.WrapErrorWithPath("idml", "parse styles", PathStyles, err)
	}

	// Cache for future calls
	p.cacheStyles(styles)
	return styles, nil
}

// SetFonts updates the cached fonts file.
// The file will be marshaled when Write() is called.
func (p *Package) SetFonts(fonts *resources.FontsFile) {
	p.cacheFonts(fonts)
	p.markDirty(PathFonts)
}

// SetStyles updates the cached styles file.
// The file will be marshaled when Write() is called.
func (p *Package) SetStyles(styles *resources.StylesFile) {
	p.cacheStyles(styles)
	p.markDirty(PathStyles)
}

// SetGraphics updates the cached graphics file.
// The file will be marshaled when Write() is called.
func (p *Package) SetGraphics(graphics *resources.GraphicFile) {
	p.cacheGraphics(graphics)
	p.markDirty(PathGraphic)
}

// XMP returns an XMP accessor for the package metadata.
// This allows reading and modifying XMP metadata in a type-safe way.
// Returns an XMP Metadata instance that can be used to update timestamps,
// remove thumbnails, or modify specific fields.
func (p *Package) XMP() *xmp.Metadata {
	return xmp.Parse(p.XMPMetadata)
}

// SetXMP updates the package XMP metadata.
// This should be called after modifying XMP metadata to persist changes.
func (p *Package) SetXMP(x *xmp.Metadata) {
	p.XMPMetadata = x.String()
}

// ============================================================================
