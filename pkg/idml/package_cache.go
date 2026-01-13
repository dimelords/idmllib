package idml

import (
	"github.com/dimelords/idmllib/v2/pkg/document"
	"github.com/dimelords/idmllib/v2/pkg/resources"
	"github.com/dimelords/idmllib/v2/pkg/spread"
	"github.com/dimelords/idmllib/v2/pkg/story"
)

// Cache management methods for Package struct.
// These methods provide centralized cache operations and invalidation.

// clearCache clears all cached parsed objects.
// This forces re-parsing from raw file data on next access.
// Useful when file data has been modified externally.
func (p *Package) clearCache() {
	// Clear document cache
	p.document = nil
	p.documentMetadata = nil

	// Clear story cache
	p.stories = make(map[string]*story.Story)

	// Clear spread cache
	p.spreads = make(map[string]*spread.Spread)

	// Clear resource cache
	p.resources = make(map[string]*ResourceFile)

	// Clear typed resource cache
	p.fonts = nil
	p.graphics = nil
	p.styles = nil

	// Clear metadata cache
	p.metadata = make(map[string]*MetadataFile)

	// Clear index cache
	p.indexState = itemIndexState{}
}

// invalidateCache invalidates cached objects for a specific file path.
// This is more efficient than clearing all caches when only one file changes.
func (p *Package) invalidateCache(path string) {
	switch path {
	case PathDesignmap:
		p.document = nil
		p.documentMetadata = nil

	case PathFonts:
		p.fonts = nil
		// Also clear generic resource cache for this file
		delete(p.resources, path)

	case PathGraphic:
		p.graphics = nil
		// Also clear generic resource cache for this file
		delete(p.resources, path)

	case PathStyles:
		p.styles = nil
		// Also clear generic resource cache for this file
		delete(p.resources, path)

	default:
		// Handle story files
		if IsStoryPath(path) {
			delete(p.stories, path)
			// Invalidate index since stories changed
			p.invalidateIndex()
			return
		}

		// Handle spread files
		if IsSpreadPath(path) {
			delete(p.spreads, path)
			// Invalidate index since spreads changed
			p.invalidateIndex()
			return
		}

		// Handle resource files
		if IsResourcePath(path) {
			delete(p.resources, path)
			return
		}

		// Handle metadata files
		if IsMetaInfPath(path) || IsXMLPath(path) {
			delete(p.metadata, path)
			return
		}
	}
}

// invalidateStoryCache clears all cached story objects.
// Useful when multiple stories have been modified.
func (p *Package) invalidateStoryCache() {
	p.stories = make(map[string]*story.Story)
	p.invalidateIndex() // Stories affect the index
}

// invalidateSpreadCache clears all cached spread objects.
// Useful when multiple spreads have been modified.
func (p *Package) invalidateSpreadCache() {
	p.spreads = make(map[string]*spread.Spread)
	p.invalidateIndex() // Spreads affect the index
}

// invalidateResourceCache clears all cached resource objects.
// This includes both generic resources and typed resources.
func (p *Package) invalidateResourceCache() {
	p.resources = make(map[string]*ResourceFile)
	p.fonts = nil
	p.graphics = nil
	p.styles = nil
}

// invalidateMetadataCache clears all cached metadata objects.
func (p *Package) invalidateMetadataCache() {
	p.metadata = make(map[string]*MetadataFile)
}

// invalidateIndex clears the page item index.
// The index will be rebuilt on next access to selection methods.
func (p *Package) invalidateIndex() {
	p.indexState = itemIndexState{}
}

// getCacheStats returns statistics about cached objects.
// Useful for debugging and monitoring cache usage.
func (p *Package) getCacheStats() CacheStats {
	stats := CacheStats{}

	// Document cache
	if p.document != nil {
		stats.DocumentCached = true
	}

	// Story cache
	stats.StoriesCached = len(p.stories)

	// Spread cache
	stats.SpreadsCached = len(p.spreads)

	// Resource cache
	stats.ResourcesCached = len(p.resources)

	// Typed resource cache
	if p.fonts != nil {
		stats.FontsCached = true
	}
	if p.graphics != nil {
		stats.GraphicsCached = true
	}
	if p.styles != nil {
		stats.StylesCached = true
	}

	// Metadata cache
	stats.MetadataCached = len(p.metadata)

	// Index cache
	if p.indexState.index != nil {
		stats.IndexCached = true
		stats.IndexedItems = p.ItemCount()
	}

	return stats
}

// CacheStats provides information about cached objects in a Package.
type CacheStats struct {
	DocumentCached  bool // Whether document is cached
	StoriesCached   int  // Number of cached stories
	SpreadsCached   int  // Number of cached spreads
	ResourcesCached int  // Number of cached generic resources
	FontsCached     bool // Whether typed fonts are cached
	GraphicsCached  bool // Whether typed graphics are cached
	StylesCached    bool // Whether typed styles are cached
	MetadataCached  int  // Number of cached metadata files
	IndexCached     bool // Whether page item index is cached
	IndexedItems    int  // Number of items in the index
}

// ensureCacheInitialized ensures all cache maps are initialized.
// This is called by methods that need to write to cache maps.
func (p *Package) ensureCacheInitialized() {
	if p.stories == nil {
		p.stories = make(map[string]*story.Story)
	}
	if p.spreads == nil {
		p.spreads = make(map[string]*spread.Spread)
	}
	if p.resources == nil {
		p.resources = make(map[string]*ResourceFile)
	}
	if p.metadata == nil {
		p.metadata = make(map[string]*MetadataFile)
	}
}

// cacheStory stores a parsed story in the cache.
func (p *Package) cacheStory(filename string, st *story.Story) {
	p.ensureCacheInitialized()
	p.stories[filename] = st
}

// cacheSpread stores a parsed spread in the cache.
func (p *Package) cacheSpread(filename string, sp *spread.Spread) {
	p.ensureCacheInitialized()
	p.spreads[filename] = sp
}

// cacheResource stores a parsed resource in the cache.
func (p *Package) cacheResource(filename string, resource *ResourceFile) {
	p.ensureCacheInitialized()
	p.resources[filename] = resource
}

// cacheMetadata stores a parsed metadata file in the cache.
func (p *Package) cacheMetadata(path string, metadata *MetadataFile) {
	p.ensureCacheInitialized()
	p.metadata[path] = metadata
}

// cacheDocument stores a parsed document in the cache.
func (p *Package) cacheDocument(doc *document.Document, docMeta *document.DocumentWithMetadata) {
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
func (p *Package) getCachedStory(filename string) (*story.Story, bool) {
	if p.stories == nil {
		return nil, false
	}
	st, exists := p.stories[filename]
	return st, exists
}

// getCachedSpread retrieves a cached spread if it exists.
func (p *Package) getCachedSpread(filename string) (*spread.Spread, bool) {
	if p.spreads == nil {
		return nil, false
	}
	sp, exists := p.spreads[filename]
	return sp, exists
}

// getCachedResource retrieves a cached resource if it exists.
func (p *Package) getCachedResource(filename string) (*ResourceFile, bool) {
	if p.resources == nil {
		return nil, false
	}
	resource, exists := p.resources[filename]
	return resource, exists
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
