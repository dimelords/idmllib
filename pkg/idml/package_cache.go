package idml

import (
	"github.com/dimelords/idmllib/v3/pkg/document"
	"github.com/dimelords/idmllib/v3/pkg/resources"
	"github.com/dimelords/idmllib/v3/pkg/spread"
	"github.com/dimelords/idmllib/v3/pkg/story"
)

// Cache management methods for Package struct.
// These methods provide centralized cache operations and invalidation.

// invalidateCache invalidates cached objects for a specific file path.
// This is more efficient than clearing all caches when only one file changes.
func (p *Package) invalidateCache(path string) {
	p.clearDirty(path)
	switch path {
	case PathDesignmap:
		p.document = nil
		p.documentMetadata = nil

	case PathFonts:
		p.fonts = nil
		// Also clear generic resource cache for this file

	case PathGraphic:
		p.graphics = nil
		// Also clear generic resource cache for this file

	case PathStyles:
		p.styles = nil
	case PathPreferences:
		p.preferences = nil
	case PathTags:
		p.tags = nil
		// Also clear generic resource cache for this file

	default:
		// Handle story files
		if IsStoryPath(path) {
			delete(p.stories, path)
			// Invalidate index since stories changed
			p.invalidateIndex()
			return
		}

		// Handle spread files
		if IsSpreadPath(path) || IsMasterSpreadPath(path) {
			delete(p.spreads, path)
			// Invalidate index since spreads changed
			p.invalidateIndex()
			return
		}

		// Handle resource files
		if IsResourcePath(path) {
			return
		}

		// Handle metadata files
		if IsMetaInfPath(path) || IsXMLPath(path) {
			delete(p.metadata, path)
			return
		}
	}
}

// invalidateIndex clears the page item index.
// The index will be rebuilt on next access to selection methods.
func (p *Package) invalidateIndex() {
	p.indexState = itemIndexState{}
}

// ensureCacheInitialized ensures all cache maps are initialized.
// This is called by methods that need to write to cache maps.
func (p *Package) ensureCacheInitialized() {
	if p.stories == nil {
		p.stories = make(map[string]*story.File)
	}
	if p.spreads == nil {
		p.spreads = make(map[string]*spread.File)
	}
	if p.metadata == nil {
		p.metadata = make(map[string]*MetadataFile)
	}
}

// cacheStory stores a parsed story in the cache.
func (p *Package) cacheStory(filename string, st *story.File) {
	p.ensureCacheInitialized()
	p.stories[filename] = st
}

// cacheSpread stores a parsed spread in the cache.
func (p *Package) cacheSpread(filename string, sp *spread.File) {
	p.ensureCacheInitialized()
	p.spreads[filename] = sp
}

// cacheMetadata stores a parsed metadata file in the cache.
func (p *Package) cacheMetadata(path string, metadata *MetadataFile) {
	p.ensureCacheInitialized()
	p.metadata[path] = metadata
}

// cacheDocument stores a parsed document in the cache.
func (p *Package) cacheDocument(doc *document.Document, docMeta *document.File) {
	p.document = doc
	p.documentMetadata = docMeta
}

// cacheFonts stores parsed fonts in the cache.
func (p *Package) cacheFonts(fonts *resources.FontsFile) {
	p.fonts = fonts
}

// cacheGraphics stores parsed graphics in the cache.
func (p *Package) cacheGraphics(graphics *resources.GraphicFile) {
	p.graphics = graphics
}

// cacheStyles stores parsed styles in the cache.
func (p *Package) cacheStyles(styles *resources.StylesFile) {
	p.styles = styles
}

// getCachedStory retrieves a cached story if it exists.
func (p *Package) getCachedStory(filename string) (*story.File, bool) {
	if p.stories == nil {
		return nil, false
	}
	st, exists := p.stories[filename]
	return st, exists
}

// getCachedSpread retrieves a cached spread if it exists.
func (p *Package) getCachedSpread(filename string) (*spread.File, bool) {
	if p.spreads == nil {
		return nil, false
	}
	sp, exists := p.spreads[filename]
	return sp, exists
}

// getCachedMetadata retrieves a cached metadata file if it exists.
func (p *Package) getCachedMetadata(path string) (*MetadataFile, bool) {
	if p.metadata == nil {
		return nil, false
	}
	metadata, exists := p.metadata[path]
	return metadata, exists
}

// getCachedDocument retrieves the cached document if it exists.
func (p *Package) getCachedDocument() (*document.Document, bool) {
	return p.document, p.document != nil
}

// getCachedFonts retrieves cached fonts if they exist.
func (p *Package) getCachedFonts() (*resources.FontsFile, bool) {
	return p.fonts, p.fonts != nil
}

// getCachedGraphics retrieves cached graphics if they exist.
func (p *Package) getCachedGraphics() (*resources.GraphicFile, bool) {
	return p.graphics, p.graphics != nil
}

// getCachedStyles retrieves cached styles if they exist.
func (p *Package) getCachedStyles() (*resources.StylesFile, bool) {
	return p.styles, p.styles != nil
}
