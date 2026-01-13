package idml

import (
	"github.com/dimelords/idmllib/pkg/common"
	"github.com/dimelords/idmllib/pkg/spread"
	"github.com/dimelords/idmllib/pkg/story"
)

// Helper functions for validation and common operations
// ============================================================================

// validateStoryDependencies validates story dependencies if validation options require it.
// Returns error if validation fails and FailOnMissing is true.
func (p *Package) validateStoryDependencies(st *story.Story, opts ValidationOptions, filename string) error {
	if !opts.EnsureStylesExist && !opts.EnsureFontsExist && !opts.EnsureColorsExist {
		return nil // No validation requested
	}

	rm := NewResourceManager(p)

	// Create a temporary dependency set for this story
	deps := newDependencySet()
	if err := rm.analyzeStory(st, deps); err != nil {
		return common.WrapErrorWithPath("idml", "analyze story dependencies", filename, err)
	}

	// Check for missing resources
	missing, err := rm.findMissingResourcesForDeps(deps)
	if err != nil {
		return common.WrapErrorWithPath("idml", "find missing resources", filename, err)
	}

	if missing.HasMissing() {
		if opts.AutoAddMissing {
			// Add missing resources
			if err := rm.AddMissingResources(opts); err != nil {
				return common.WrapErrorWithPath("idml", "add missing resources", filename, err)
			}
		} else if opts.FailOnMissing {
			return common.WrapErrorWithPath("idml", "validate story", filename, NewMissingResourcesError(missing))
		}
	}

	return nil
}

// validateTextFrameDependencies validates text frame dependencies if validation options require it.
func (p *Package) validateTextFrameDependencies(tf *spread.SpreadTextFrame, opts ValidationOptions, spreadFilename string) error {
	if !opts.EnsureStylesExist && !opts.EnsureFontsExist && !opts.EnsureColorsExist {
		return nil // No validation requested
	}

	rm := NewResourceManager(p)

	// Check if the referenced story exists (if ParentStory is set)
	if tf.ParentStory != "" {
		if _, err := p.Story(tf.ParentStory); err != nil {
			if opts.FailOnMissing {
				return common.WrapErrorWithPath("idml", "validate text frame", spreadFilename, err)
			}
		}
	}

	// Check object style if specified
	if tf.AppliedObjectStyle != "" && opts.EnsureStylesExist {
		return p.validateObjectStyle(tf.AppliedObjectStyle, opts, spreadFilename, rm)
	}

	return nil
}

// validateRectangleDependencies validates rectangle dependencies if validation options require it.
func (p *Package) validateRectangleDependencies(rect *spread.Rectangle, opts ValidationOptions, spreadFilename string) error {
	if !opts.EnsureStylesExist && !opts.EnsureColorsExist {
		return nil // No validation requested
	}

	rm := NewResourceManager(p)

	// Check object style if specified
	if rect.AppliedObjectStyle != "" && opts.EnsureStylesExist {
		return p.validateObjectStyle(rect.AppliedObjectStyle, opts, spreadFilename, rm)
	}

	return nil
}

// validateObjectStyle validates that an object style exists or adds it if AutoAddMissing is enabled.
func (p *Package) validateObjectStyle(styleID string, opts ValidationOptions, context string, rm *ResourceManager) error {
	// Delegate to ResourceManager for style validation
	missing, err := rm.findMissingObjectStyle(styleID)
	if err != nil {
		return common.WrapErrorWithPath("idml", "validate object style", context, err)
	}

	if missing {
		if opts.FailOnMissing && !opts.AutoAddMissing {
			return common.WrapErrorWithPath("idml", "validate object style", context, common.ErrMissingDependency)
		}

		if opts.AutoAddMissing {
			// Auto-add the missing object style
			if err := rm.addMissingObjectStyle(styleID); err != nil {
				return common.WrapErrorWithPath("idml", "validate object style", context, err)
			}
		}
	}

	return nil
}

// marshalAndUpdateSpread marshals a spread and updates it in the package.
func (p *Package) marshalAndUpdateSpread(spreadFilename string, sp *spread.Spread) error {
	data, err := spread.MarshalSpread(sp)
	if err != nil {
		return common.WrapErrorWithPath("idml", "marshal spread", spreadFilename, err)
	}

	// Update files map
	p.setFileData(spreadFilename, data)

	// Update cached spread
	p.cacheSpread(spreadFilename, sp)

	return nil
}

// marshalAndUpdateStory marshals a story and updates it in the package.
func (p *Package) marshalAndUpdateStory(filename string, st *story.Story) error {
	data, err := story.MarshalStory(st)
	if err != nil {
		return common.WrapErrorWithPath("idml", "marshal story", filename, err)
	}

	// Update files map
	p.setFileData(filename, data)

	// Update cached story
	p.cacheStory(filename, st)

	return nil
}

// removeItemFromSpread removes an item from a spread by ID and type.
// Returns true if the item was found and removed, false otherwise.
func (p *Package) removeItemFromSpread(sp *spread.Spread, itemID string, itemType string) bool {
	switch itemType {
	case "textframe":
		for i, tf := range sp.InnerSpread.TextFrames {
			if tf.Self == itemID {
				sp.InnerSpread.TextFrames = append(sp.InnerSpread.TextFrames[:i], sp.InnerSpread.TextFrames[i+1:]...)
				return true
			}
		}
	case "rectangle":
		for i, rect := range sp.InnerSpread.Rectangles {
			if rect.Self == itemID {
				sp.InnerSpread.Rectangles = append(sp.InnerSpread.Rectangles[:i], sp.InnerSpread.Rectangles[i+1:]...)
				return true
			}
		}
	}
	return false
}

// Epic 2: High-Level Modification API - Story Operations
// ============================================================================

// RemoveStory removes a story from the package and optionally cleans up orphaned resources.
//
// This is a high-level API that:
//  1. Validates the story exists
//  2. Removes it from all internal maps
//  3. Updates the fileOrder slice
//  4. Optionally runs cleanup to remove orphaned resources
//
// Parameters:
//   - filename: The story filename (e.g., "Stories/Story_u1d8.xml")
//   - cleanup: If true, automatically remove orphaned resources after deletion
//
// Returns:
//   - CleanupResult: Details about resources that were removed (if cleanup=true)
//   - error: Error if story doesn't exist or cleanup fails
//
// Example:
//
//	pkg, _ := idml.Read("document.idml")
//	result, err := pkg.RemoveStory("Stories/Story_u1d8.xml", true)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Removed %d orphaned resources\n", result.Count())
func (p *Package) RemoveStory(filename string, cleanup bool) (*CleanupResult, error) {
	// Step 1: Validate story exists
	if err := p.validateStoryExists(filename); err != nil {
		return nil, err
	}

	// Step 2: Remove from caches and files
	p.removeStoryFromPackage(filename)

	// Step 3: Cleanup orphaned resources if requested
	return p.performCleanupIfRequested(cleanup)
}

// validateStoryExists checks if a story file exists in the package.
func (p *Package) validateStoryExists(filename string) error {
	if !p.hasFile(filename) {
		return common.WrapErrorWithPath("idml", "remove story", filename, common.ErrNotFound)
	}
	return nil
}

// removeStoryFromPackage removes the story from internal caches and files.
func (p *Package) removeStoryFromPackage(filename string) {
	p.invalidateCache(filename)
	p.removeFile(filename)
}

// performCleanupIfRequested performs resource cleanup if requested.
func (p *Package) performCleanupIfRequested(cleanup bool) (*CleanupResult, error) {
	if cleanup {
		rm := NewResourceManager(p)
		result, err := rm.CleanupOrphans(DefaultCleanupOptions())
		if err != nil {
			return nil, common.WrapError("idml", "cleanup after remove story", err)
		}
		return result, nil
	}
	// No cleanup performed
	return &CleanupResult{}, nil
}

// AddStory adds a new story to the package with optional validation.
//
// This is a high-level API that:
//  1. Validates the story doesn't already exist
//  2. Analyzes story dependencies (styles, fonts, colors)
//  3. Validates all resources exist (if validation enabled)
//  4. Auto-adds missing resources (if AutoAddMissing enabled)
//  5. Adds story to the package
//
// Parameters:
//   - filename: The story filename (e.g., "Stories/Story_u1d8.xml")
//   - story: The Story struct to add
//   - opts: Validation options controlling dependency checking
//
// Returns:
//   - error: Error if story already exists, validation fails, or marshal fails
//
// Example:
//
//	pkg, _ := idml.Read("document.idml")
//	newStory := &idml.Story{ /* ... */ }
//	opts := idml.ValidationOptions{
//	    EnsureStylesExist: true,
//	    AutoAddMissing:    true,
//	}
//	err := pkg.AddStory("Stories/Story_new.xml", newStory, opts)
func (p *Package) AddStory(filename string, st *story.Story, opts ValidationOptions) error {
	// Step 1: Validate story doesn't already exist
	if err := p.validateStoryDoesNotExist(filename); err != nil {
		return err
	}

	// Step 2: Validate dependencies if requested
	if err := p.validateStoryDependencies(st, opts, filename); err != nil {
		return err
	}

	// Step 3: Marshal and add the story
	return p.marshalAndUpdateStory(filename, st)
}

// validateStoryDoesNotExist checks that a story file doesn't already exist.
func (p *Package) validateStoryDoesNotExist(filename string) error {
	if p.hasFile(filename) {
		return common.WrapErrorWithPath("idml", "add story", filename, common.ErrAlreadyExists)
	}
	return nil
}

// UpdateStory replaces an existing story with a new version, with optional validation.
//
// This is a high-level API that:
//  1. Validates the story exists
//  2. Analyzes new story dependencies
//  3. Validates all resources exist (if validation enabled)
//  4. Auto-adds missing resources (if AutoAddMissing enabled)
//  5. Replaces story in the package (preserves fileOrder position)
//
// Parameters:
//   - filename: The story filename (e.g., "Stories/Story_u1d8.xml")
//   - story: The new Story struct to replace the old one
//   - opts: Validation options controlling dependency checking
//
// Returns:
//   - error: Error if story doesn't exist, validation fails, or marshal fails
//
// Example:
//
//	pkg, _ := idml.Read("document.idml")
//	story, _ := pkg.Story("Stories/Story_u1d8.xml")
//	// Modify story...
//	story.StoryElement.ParagraphStyleRanges[0].AppliedParagraphStyle = "NewStyle"
//	err := pkg.UpdateStory("Stories/Story_u1d8.xml", story, idml.DefaultValidationOptions())
func (p *Package) UpdateStory(filename string, st *story.Story, opts ValidationOptions) error {
	// Step 1: Validate story exists
	if !p.hasFile(filename) {
		return common.WrapErrorWithPath("idml", "update story", filename, common.ErrNotFound)
	}

	// Step 2: Validate dependencies if requested
	if err := p.validateStoryDependencies(st, opts, filename); err != nil {
		return err
	}

	// Step 3: Marshal and update the story
	if err := p.marshalAndUpdateStory(filename, st); err != nil {
		return err
	}

	return nil
}

// ═══════════════════════════════════════════════════════════════════════════
// TextFrame Operations
// ═══════════════════════════════════════════════════════════════════════════

// RemoveTextFrame removes a text frame from a spread and optionally cleans up orphaned resources.
//
// This operation:
//  1. Loads the spread file
//  2. Removes the text frame by ID
//  3. Updates the spread file
//  4. Optionally removes orphaned resources (fonts, styles, colors)
//
// Example:
//
//	result, err := pkg.RemoveTextFrame("Spreads/Spread_u210.xml", "u1d9", true)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Removed %d orphaned resources\n", result.Count())
func (p *Package) RemoveTextFrame(spreadFilename string, textFrameID string, cleanup bool) (*CleanupResult, error) {
	// Step 1: Load the spread
	sp, err := p.loadSpreadForModification(spreadFilename, "remove text frame")
	if err != nil {
		return nil, err
	}

	// Step 2: Remove the text frame
	if err := p.removeTextFrameFromSpread(sp, textFrameID, spreadFilename); err != nil {
		return nil, err
	}

	// Step 3: Marshal and save the spread
	if err := p.marshalAndUpdateSpread(spreadFilename, sp); err != nil {
		return nil, err
	}

	// Step 4: Cleanup orphaned resources if requested
	return p.performCleanupIfRequested(cleanup)
}

// loadSpreadForModification loads a spread file for modification operations.
func (p *Package) loadSpreadForModification(spreadFilename, operation string) (*spread.Spread, error) {
	sp, err := p.Spread(spreadFilename)
	if err != nil {
		return nil, common.WrapErrorWithPath("idml", operation, spreadFilename, err)
	}
	return sp, nil
}

// removeTextFrameFromSpread removes a text frame by ID from the spread.
func (p *Package) removeTextFrameFromSpread(sp *spread.Spread, textFrameID, spreadFilename string) error {
	if !p.removeItemFromSpread(sp, textFrameID, "textframe") {
		return common.WrapErrorWithPath("idml", "remove text frame", spreadFilename, common.ErrNotFound)
	}
	return nil
}

// AddTextFrame adds a text frame to a spread with optional validation.
//
// This operation:
//  1. Loads the spread file
//  2. Validates the text frame's dependencies (if requested)
//  3. Adds the text frame to the spread
//  4. Updates the spread file
//
// Example:
//
//	tf := &idml.SpreadTextFrame{
//	    BasicFrame: idml.BasicFrame{Self: "u1d9"},
//	    ParentStory: "Stories/Story_u1d8.xml",
//	    ItemTransform: "1 0 0 1 72 72",
//	}
//	opts := idml.ValidationOptions{
//	    EnsureStylesExist: true,
//	    AutoAddMissing:    true,
//	}
//	err := pkg.AddTextFrame("Spreads/Spread_u210.xml", tf, opts)
func (p *Package) AddTextFrame(spreadFilename string, tf *spread.SpreadTextFrame, opts ValidationOptions) error {
	// Step 1: Load the spread
	sp, err := p.loadSpreadForModification(spreadFilename, "add text frame")
	if err != nil {
		return err
	}

	// Step 2: Check if text frame ID already exists
	if err := p.validateTextFrameDoesNotExist(sp, tf.Self, spreadFilename); err != nil {
		return err
	}

	// Step 3: Validate dependencies if requested
	if err := p.validateTextFrameDependencies(tf, opts, spreadFilename); err != nil {
		return err
	}

	// Step 4: Add the text frame
	p.addTextFrameToSpread(sp, tf)

	// Step 5: Marshal and save the spread
	return p.marshalAndUpdateSpread(spreadFilename, sp)
}

// validateTextFrameDoesNotExist checks that a text frame ID doesn't already exist.
func (p *Package) validateTextFrameDoesNotExist(sp *spread.Spread, textFrameID, spreadFilename string) error {
	for _, existing := range sp.InnerSpread.TextFrames {
		if existing.Self == textFrameID {
			return common.WrapErrorWithPath("idml", "add text frame", spreadFilename, common.ErrAlreadyExists)
		}
	}
	return nil
}

// addTextFrameToSpread adds a text frame to the spread.
func (p *Package) addTextFrameToSpread(sp *spread.Spread, tf *spread.SpreadTextFrame) {
	sp.InnerSpread.TextFrames = append(sp.InnerSpread.TextFrames, *tf)
}

// UpdateTextFrame updates a text frame in a spread with optional validation.
//
// This operation:
//  1. Loads the spread file
//  2. Finds the text frame by ID
//  3. Validates the updated text frame's dependencies (if requested)
//  4. Updates the text frame
//  5. Saves the spread file
//
// Example:
//
//	// Load and modify text frame
//	tf := &idml.SpreadTextFrame{
//	    BasicFrame: idml.BasicFrame{Self: "u1d9"},
//	    ParentStory: "Stories/Story_u1d8.xml",
//	    ItemTransform: "1 0 0 1 144 144", // Move it
//	}
//	err := pkg.UpdateTextFrame("Spreads/Spread_u210.xml", "u1d9", tf, idml.DefaultValidationOptions())
func (p *Package) UpdateTextFrame(spreadFilename string, textFrameID string, tf *spread.SpreadTextFrame, opts ValidationOptions) error {
	// Step 1: Load the spread
	sp, err := p.loadSpreadForModification(spreadFilename, "update text frame")
	if err != nil {
		return err
	}

	// Step 2: Find and update the text frame
	if err := p.findAndUpdateTextFrame(sp, textFrameID, tf, opts, spreadFilename); err != nil {
		return err
	}

	// Step 3: Marshal and save the spread
	return p.marshalAndUpdateSpread(spreadFilename, sp)
}

// findAndUpdateTextFrame finds a text frame by ID and updates it.
func (p *Package) findAndUpdateTextFrame(sp *spread.Spread, textFrameID string, tf *spread.SpreadTextFrame, opts ValidationOptions, spreadFilename string) error {
	for i, existing := range sp.InnerSpread.TextFrames {
		if existing.Self == textFrameID {
			// Validate dependencies if requested
			if err := p.validateTextFrameDependencies(tf, opts, spreadFilename); err != nil {
				return err
			}

			// Update the text frame
			sp.InnerSpread.TextFrames[i] = *tf
			return nil
		}
	}

	return common.WrapErrorWithPath("idml", "update text frame", spreadFilename, common.ErrNotFound)
}

// ═══════════════════════════════════════════════════════════════════════════
// Rectangle Operations
// ═══════════════════════════════════════════════════════════════════════════

// RemoveRectangle removes a rectangle from a spread and optionally cleans up orphaned resources.
//
// This operation:
//  1. Loads the spread file
//  2. Removes the rectangle by ID
//  3. Updates the spread file
//  4. Optionally removes orphaned resources (fonts, styles, colors)
//
// Example:
//
//	result, err := pkg.RemoveRectangle("Spreads/Spread_u210.xml", "u1da", true)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Removed %d orphaned resources\n", result.Count())
func (p *Package) RemoveRectangle(spreadFilename string, rectangleID string, cleanup bool) (*CleanupResult, error) {
	// Step 1: Load the spread
	sp, err := p.Spread(spreadFilename)
	if err != nil {
		return nil, common.WrapErrorWithPath("idml", "remove rectangle", spreadFilename, err)
	}

	// Step 2: Remove the rectangle
	if !p.removeItemFromSpread(sp, rectangleID, "rectangle") {
		return nil, common.WrapErrorWithPath("idml", "remove rectangle", spreadFilename, common.ErrNotFound)
	}

	// Step 3: Marshal and save the spread
	if err := p.marshalAndUpdateSpread(spreadFilename, sp); err != nil {
		return nil, err
	}

	// Step 4: Cleanup orphaned resources if requested
	if cleanup {
		rm := NewResourceManager(p)
		return rm.CleanupOrphans(DefaultCleanupOptions())
	}

	return &CleanupResult{}, nil
}

// AddRectangle adds a rectangle to a spread with optional validation.
//
// This operation:
//  1. Loads the spread file
//  2. Validates the rectangle's dependencies (if requested)
//  3. Adds the rectangle to the spread
//  4. Updates the spread file
//
// Example:
//
//	rect := &idml.Rectangle{
//	    Self: "u1da",
//	    ContentType: "GraphicType",
//	    ItemTransform: "1 0 0 1 72 72",
//	    AppliedObjectStyle: "ObjectStyle/$ID/[Normal Graphics Frame]",
//	}
//	opts := idml.ValidationOptions{
//	    EnsureStylesExist: true,
//	    AutoAddMissing:    true,
//	}
//	err := pkg.AddRectangle("Spreads/Spread_u210.xml", rect, opts)
func (p *Package) AddRectangle(spreadFilename string, rect *spread.Rectangle, opts ValidationOptions) error {
	// Step 1: Load the spread
	sp, err := p.Spread(spreadFilename)
	if err != nil {
		return common.WrapErrorWithPath("idml", "add rectangle", spreadFilename, err)
	}

	// Step 2: Check if rectangle ID already exists
	for _, existing := range sp.InnerSpread.Rectangles {
		if existing.Self == rect.Self {
			return common.WrapErrorWithPath("idml", "add rectangle", spreadFilename, common.ErrAlreadyExists)
		}
	}

	// Step 3: Validate dependencies if requested
	if err := p.validateRectangleDependencies(rect, opts, spreadFilename); err != nil {
		return err
	}

	// Step 4: Add the rectangle
	sp.InnerSpread.Rectangles = append(sp.InnerSpread.Rectangles, *rect)

	// Step 5: Marshal and save the spread
	return p.marshalAndUpdateSpread(spreadFilename, sp)
}

// UpdateRectangle updates a rectangle in a spread with optional validation.
//
// This operation:
//  1. Loads the spread file
//  2. Finds the rectangle by ID
//  3. Validates the updated rectangle's dependencies (if requested)
//  4. Updates the rectangle
//  5. Saves the spread file
//
// Example:
//
//	// Load and modify rectangle
//	rect := &idml.Rectangle{
//	    Self: "u1da",
//	    ContentType: "GraphicType",
//	    ItemTransform: "1 0 0 1 144 144", // Move it
//	}
//	err := pkg.UpdateRectangle("Spreads/Spread_u210.xml", "u1da", rect, idml.DefaultValidationOptions())
func (p *Package) UpdateRectangle(spreadFilename string, rectangleID string, rect *spread.Rectangle, opts ValidationOptions) error {
	// Step 1: Load the spread
	sp, err := p.Spread(spreadFilename)
	if err != nil {
		return common.WrapErrorWithPath("idml", "update rectangle", spreadFilename, err)
	}

	// Step 2: Find and validate the rectangle
	found := false
	for i, existing := range sp.InnerSpread.Rectangles {
		if existing.Self == rectangleID {
			found = true

			// Step 3: Validate dependencies if requested
			if err := p.validateRectangleDependencies(rect, opts, spreadFilename); err != nil {
				return err
			}

			// Step 4: Update the rectangle
			sp.InnerSpread.Rectangles[i] = *rect
			break
		}
	}

	if !found {
		return common.WrapErrorWithPath("idml", "update rectangle", spreadFilename, common.ErrNotFound)
	}

	// Step 5: Marshal and save the spread
	return p.marshalAndUpdateSpread(spreadFilename, sp)
}
