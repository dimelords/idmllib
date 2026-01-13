// resourcemgr.go defines the ResourceManager type and related types for
// managing IDML document resources (Epic 2: Resource Management API).
package idml

import "fmt"

// ResourceManager analyzes and manages resources in an IDML package.
// It provides functionality for:
//   - Detecting orphaned resources (fonts, styles, colors)
//   - Cleaning up unused resources
//   - Validating resource references
//   - Auto-resolution of missing resources
//
// The ResourceManager analyzes the document content directly to determine
// which resources are actually used.
type ResourceManager struct {
	pkg *Package
}

// NewResourceManager creates a new ResourceManager for the given package.
// The manager can be reused for multiple operations on the same package.
func NewResourceManager(pkg *Package) *ResourceManager {
	return &ResourceManager{
		pkg: pkg,
	}
}

// dependencySet tracks all dependencies found during analysis.
// This is an internal type used by the ResourceManager.
type dependencySet struct {
	fonts           map[string]bool
	paragraphStyles map[string]bool
	characterStyles map[string]bool
	objectStyles    map[string]bool
	colors          map[string]bool
	swatches        map[string]bool
	layers          map[string]bool
}

// newDependencySet creates a new empty dependency set.
func newDependencySet() *dependencySet {
	return &dependencySet{
		fonts:           make(map[string]bool),
		paragraphStyles: make(map[string]bool),
		characterStyles: make(map[string]bool),
		objectStyles:    make(map[string]bool),
		colors:          make(map[string]bool),
		swatches:        make(map[string]bool),
		layers:          make(map[string]bool),
	}
}

// CleanupOptions configures what to clean up when removing elements.
// By default, most cleanup operations are enabled for safety.
// Use DefaultCleanupOptions() to get a safe default configuration.
type CleanupOptions struct {
	// RemoveOrphanedFonts removes fonts that are no longer used in any story
	RemoveOrphanedFonts bool

	// RemoveOrphanedParagraphStyles removes paragraph styles not used in any story
	RemoveOrphanedParagraphStyles bool

	// RemoveOrphanedCharacterStyles removes character styles not used in any story
	RemoveOrphanedCharacterStyles bool

	// RemoveOrphanedObjectStyles removes object styles not used on any page item
	// DEFAULT: false - object styles often serve as templates
	RemoveOrphanedObjectStyles bool

	// RemoveOrphanedColors removes colors not used in any element
	// DEFAULT: false - colors are part of the color library
	RemoveOrphanedColors bool

	// RemoveOrphanedSwatches removes swatches not used in any element
	// DEFAULT: false - swatches are part of the swatch library
	RemoveOrphanedSwatches bool

	// RemoveOrphanedLayers removes empty layers with no page items
	// DEFAULT: false - layers are structural and users often want to keep them
	RemoveOrphanedLayers bool

	// DryRun doesn't actually remove anything, just reports what would be removed
	// Useful for previewing cleanup operations before committing
	DryRun bool
}

// DefaultCleanupOptions returns safe default cleanup options.
// These defaults are conservative and only remove obviously unused resources.
func DefaultCleanupOptions() CleanupOptions {
	return CleanupOptions{
		RemoveOrphanedFonts:           true,
		RemoveOrphanedParagraphStyles: true,
		RemoveOrphanedCharacterStyles: true,
		RemoveOrphanedObjectStyles:    false, // Keep - often used as templates
		RemoveOrphanedColors:          false, // Keep - part of color library
		RemoveOrphanedSwatches:        false, // Keep - part of swatch library
		RemoveOrphanedLayers:          false, // Keep - layers are structural
		DryRun:                        false,
	}
}

// ValidationOptions configures validation when adding elements.
// These options control what gets validated and how missing resources are handled.
type ValidationOptions struct {
	// EnsureStylesExist verifies that paragraph and character styles exist
	EnsureStylesExist bool

	// EnsureFontsExist verifies that fonts referenced in styles exist
	EnsureFontsExist bool

	// EnsureColorsExist verifies that colors referenced in elements exist
	EnsureColorsExist bool

	// EnsureLayersExist verifies that layers referenced by page items exist
	EnsureLayersExist bool

	// AutoAddMissing automatically adds missing resources with default values
	// When false, missing resources will either fail (if FailOnMissing=true) or be ignored
	AutoAddMissing bool

	// FailOnMissing returns an error if resources are missing (when AutoAdd=false)
	// When false, validation warnings are logged but operation continues
	FailOnMissing bool
}

// Common ValidationOptions presets for convenience.
var (
	// NoValidation disables all validation checks.
	// Use this when you're confident the resources are valid or when performance is critical.
	NoValidation = ValidationOptions{}

	// FullValidation enables all validation checks but does not auto-add missing resources.
	// Use this when you want to ensure all resources exist but let the operation fail if they don't.
	FullValidation = ValidationOptions{
		EnsureStylesExist: true,
		EnsureFontsExist:  true,
		EnsureColorsExist: true,
		EnsureLayersExist: true,
		FailOnMissing:     true,
	}

	// AutoResolve enables all validation checks and automatically adds missing resources.
	// Use this when you want to ensure the document is valid and fix it automatically.
	AutoResolve = ValidationOptions{
		EnsureStylesExist: true,
		EnsureFontsExist:  true,
		EnsureColorsExist: true,
		EnsureLayersExist: true,
		AutoAddMissing:    true,
	}

	// StylesOnly validates only paragraph and character styles.
	// Use this when modifying story content but colors/layers don't matter.
	StylesOnly = ValidationOptions{
		EnsureStylesExist: true,
		FailOnMissing:     true,
	}
)

// DefaultValidationOptions returns safe default validation options.
// These defaults ensure document integrity by validating all resource references.
func DefaultValidationOptions() ValidationOptions {
	return ValidationOptions{
		EnsureStylesExist: true,
		EnsureFontsExist:  true,
		EnsureColorsExist: true,
		EnsureLayersExist: true,
		AutoAddMissing:    false, // Conservative - require explicit opt-in
		FailOnMissing:     true,  // Fail fast on missing dependencies
	}
}

// OrphanedResources contains resources that are defined but not used.
// This is the result of FindOrphans() and represents resources that can
// potentially be removed from the document.
type OrphanedResources struct {
	// Fonts contains font family names that are defined but not used
	Fonts []string

	// ParagraphStyles contains paragraph style IDs that are defined but not used
	ParagraphStyles []string

	// CharacterStyles contains character style IDs that are defined but not used
	CharacterStyles []string

	// ObjectStyles contains object style IDs that are defined but not used
	ObjectStyles []string

	// Colors contains color IDs that are defined but not used
	Colors []string

	// Swatches contains swatch IDs that are defined but not used
	Swatches []string

	// Layers contains layer IDs that exist but have no page items on them
	Layers []string
}

// HasOrphans returns true if there are any orphaned resources.
func (or *OrphanedResources) HasOrphans() bool {
	return len(or.Fonts) > 0 ||
		len(or.ParagraphStyles) > 0 ||
		len(or.CharacterStyles) > 0 ||
		len(or.ObjectStyles) > 0 ||
		len(or.Colors) > 0 ||
		len(or.Swatches) > 0 ||
		len(or.Layers) > 0
}

// Count returns the total number of orphaned resources across all types.
func (or *OrphanedResources) Count() int {
	return len(or.Fonts) +
		len(or.ParagraphStyles) +
		len(or.CharacterStyles) +
		len(or.ObjectStyles) +
		len(or.Colors) +
		len(or.Swatches) +
		len(or.Layers)
}

// CleanupResult contains information about what was cleaned up.
// This is returned by CleanupOrphans() to provide detailed feedback
// about the cleanup operation.
type CleanupResult struct {
	// RemovedFonts lists font family names that were removed
	RemovedFonts []string

	// RemovedParagraphStyles lists paragraph style IDs that were removed
	RemovedParagraphStyles []string

	// RemovedCharacterStyles lists character style IDs that were removed
	RemovedCharacterStyles []string

	// RemovedObjectStyles lists object style IDs that were removed
	RemovedObjectStyles []string

	// RemovedColors lists color IDs that were removed
	RemovedColors []string

	// RemovedSwatches lists swatch IDs that were removed
	RemovedSwatches []string

	// RemovedLayers lists layer IDs that were removed
	RemovedLayers []string
}

// Count returns the total number of resources removed across all types.
func (cr *CleanupResult) Count() int {
	return len(cr.RemovedFonts) +
		len(cr.RemovedParagraphStyles) +
		len(cr.RemovedCharacterStyles) +
		len(cr.RemovedObjectStyles) +
		len(cr.RemovedColors) +
		len(cr.RemovedSwatches) +
		len(cr.RemovedLayers)
}

// MissingResources contains resources that are referenced but not defined.
// Each map key is a resource ID, and the value is a list of element IDs
// or filenames that reference it.
type MissingResources struct {
	// Fonts maps font family names to elements using them
	Fonts map[string][]string

	// ParagraphStyles maps style IDs to story filenames using them
	ParagraphStyles map[string][]string

	// CharacterStyles maps style IDs to story filenames using them
	CharacterStyles map[string][]string

	// ObjectStyles maps style IDs to element IDs using them
	ObjectStyles map[string][]string

	// Colors maps color IDs to element IDs using them
	Colors map[string][]string

	// Swatches maps swatch IDs to element IDs using them
	Swatches map[string][]string

	// Layers maps layer IDs to element IDs on them
	Layers map[string][]string
}

// HasMissing returns true if there are any missing resources.
func (mr *MissingResources) HasMissing() bool {
	return len(mr.Fonts) > 0 ||
		len(mr.ParagraphStyles) > 0 ||
		len(mr.CharacterStyles) > 0 ||
		len(mr.ObjectStyles) > 0 ||
		len(mr.Colors) > 0 ||
		len(mr.Swatches) > 0 ||
		len(mr.Layers) > 0
}

// ValidationError represents a single validation error for a missing resource.
type ValidationError struct {
	// ResourceType describes what kind of resource is missing (e.g., "Font", "ParagraphStyle")
	ResourceType string

	// ResourceID is the identifier of the missing resource
	ResourceID string

	// UsedBy lists element IDs or filenames that reference this resource
	UsedBy []string

	// Message is a human-readable error message
	Message string
}

// Error implements the error interface for ValidationError.
func (ve *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s (used by %d elements)", ve.ResourceType, ve.ResourceID, len(ve.UsedBy))
}

// FindOrphans identifies all orphaned resources in the package.
// An orphaned resource is one that is defined in the package but not
// actually used by any element.
//
// This method scans the entire document to build a complete picture of
// what resources are defined and what resources are used. The difference
// between these two sets represents the orphaned resources.
//
// Returns an OrphanedResources struct containing all orphaned resources,
// or an error if the analysis fails.
