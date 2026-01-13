package idml

import (
	"archive/zip"

	"github.com/dimelords/idmllib/v2/pkg/common"
	"github.com/dimelords/idmllib/v2/pkg/document"
	"github.com/dimelords/idmllib/v2/pkg/resources"
	"github.com/dimelords/idmllib/v2/pkg/spread"
	"github.com/dimelords/idmllib/v2/pkg/story"
)

// fileEntry stores both the content and ZIP metadata for a file.
type fileEntry struct {
	data   []byte
	header *zip.FileHeader
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
	documentMetadata *document.DocumentWithMetadata

	// stories caches parsed Story files from the Stories/ directory.
	// Map key is the story filename (e.g., "Stories/Story_u1d8.xml").
	stories map[string]*story.Story

	// spreads caches parsed Spread files from the Spreads/ directory.
	// Map key is the spread filename (e.g., "Spreads/Spread_u210.xml").
	spreads map[string]*spread.Spread

	// resources caches parsed Resource files from the Resources/ directory.
	// Map key is the resource filename (e.g., "Resources/Graphic.xml").
	// This is the generic preservation-based parser.
	resources map[string]*ResourceFile

	// fonts caches the typed Fonts.xml file (if parsed)
	fonts *resources.FontsFile

	// graphics caches the typed Graphic.xml file (if parsed)
	graphics *resources.GraphicFile

	// styles caches the typed Styles.xml file (if parsed)
	styles *resources.StylesFile

	// metadata caches optional metadata files (META-INF/*, XML/*).
	// Map key is the file path (e.g., "META-INF/container.xml").
	metadata map[string]*MetadataFile

	// indexState holds the item index for O(1) page item lookups.
	// Built lazily on first SelectXxxByID call.
	indexState itemIndexState
}

// New creates a new empty IDML package.
func New() *Package {
	return &Package{
		files:     make(map[string]*fileEntry),
		stories:   make(map[string]*story.Story),
		spreads:   make(map[string]*spread.Spread),
		resources: make(map[string]*ResourceFile),
		metadata:  make(map[string]*MetadataFile),
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
	docMeta, err := document.ParseDocumentWithMetadata(entry.data)
	if err != nil {
		return nil, err
	}

	// Cache for future calls
	p.cacheDocument(docMeta.Document, docMeta)
	return p.document, nil
}

// Story returns a parsed Story from the Stories/ directory.
// The story is parsed on first access and cached.
// Returns an error if the story file doesn't exist or can't be parsed.
func (p *Package) Story(filename string) (*story.Story, error) {
	// Return cached story if available
	if st, cached := p.getCachedStory(filename); cached {
		return st, nil
	}

	// Get story file
	entry, err := p.getFileEntry(filename)
	if err != nil {
		return nil, err
	}

	// Parse the story
	st, err := story.ParseStory(entry.data)
	if err != nil {
		return nil, common.WrapErrorWithPath("idml", "parse story", filename, err)
	}

	// Cache for future calls
	p.cacheStory(filename, st)
	return st, nil
}

// Stories returns all parsed Story files from the Stories/ directory.
// Stories are parsed on first access and cached.
func (p *Package) Stories() (map[string]*story.Story, error) {
	stories := make(map[string]*story.Story)

	// First, add any already-cached stories
	for filename, st := range p.stories {
		stories[filename] = st
	}

	// Then find all story files from p.files that aren't cached yet
	for filename := range p.files {
		if IsStoryPath(filename) {
			// Skip if already in cache
			if _, cached := stories[filename]; cached {
				continue
			}

			st, err := p.Story(filename)
			if err != nil {
				return nil, err
			}
			stories[filename] = st
		}
	}

	return stories, nil
}

// Spread returns a parsed Spread from the Spreads/ directory.
// The spread is parsed on first access and cached.
// Returns an error if the spread file doesn't exist or can't be parsed.
func (p *Package) Spread(filename string) (*spread.Spread, error) {
	// Return cached spread if available
	if sp, cached := p.getCachedSpread(filename); cached {
		return sp, nil
	}

	// Get spread file
	entry, err := p.getFileEntry(filename)
	if err != nil {
		return nil, err
	}

	// Parse the spread
	sp, err := spread.ParseSpread(entry.data)
	if err != nil {
		return nil, common.WrapErrorWithPath("idml", "parse spread", filename, err)
	}

	// Cache for future calls
	p.cacheSpread(filename, sp)
	return sp, nil
}

// Spreads returns all parsed Spread files from the Spreads/ directory.
// Spreads are parsed on first access and cached.
func (p *Package) Spreads() (map[string]*spread.Spread, error) {
	spreads := make(map[string]*spread.Spread)

	// Find all spread files
	for filename := range p.files {
		if IsSpreadPath(filename) {
			sp, err := p.Spread(filename)
			if err != nil {
				return nil, err
			}
			spreads[filename] = sp
		}
	}

	return spreads, nil
}

// Resource returns a parsed Resource file from the Resources/ directory.
// The resource is parsed on first access and cached.
// Returns an error if the resource file doesn't exist or can't be parsed.
func (p *Package) Resource(filename string) (*ResourceFile, error) {
	// Return cached resource if available
	if resource, cached := p.getCachedResource(filename); cached {
		return resource, nil
	}

	// Get resource file
	entry, err := p.getFileEntry(filename)
	if err != nil {
		return nil, err
	}

	// Parse the resource
	resource, err := ParseResourceFile(entry.data)
	if err != nil {
		return nil, common.WrapErrorWithPath("idml", "parse resource", filename, err)
	}

	// Cache for future calls
	p.cacheResource(filename, resource)
	return resource, nil
}

// Resources returns all parsed Resource files from the Resources/ directory.
// Resources are parsed on first access and cached.
func (p *Package) Resources() (map[string]*ResourceFile, error) {
	resources := make(map[string]*ResourceFile)

	// Find all resource files
	for filename := range p.files {
		if IsResourcePath(filename) {
			resource, err := p.Resource(filename)
			if err != nil {
				return nil, err
			}
			resources[filename] = resource
		}
	}

	return resources, nil
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
	mf, err := ParseMetadataFile(path, entry.data)
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
}

// SetStyles updates the cached styles file.
// The file will be marshaled when Write() is called.
func (p *Package) SetStyles(styles *resources.StylesFile) {
	p.cacheStyles(styles)
}

// SetGraphics updates the cached graphics file.
// The file will be marshaled when Write() is called.
func (p *Package) SetGraphics(graphics *resources.GraphicFile) {
	p.cacheGraphics(graphics)
}

// ============================================================================
