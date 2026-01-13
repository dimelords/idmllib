// Package idml provides functionality for reading, writing, and manipulating
// Adobe InDesign IDML (InDesign Markup Language) files.
//
// IDML is Adobe InDesign's XML-based file format, structured as a ZIP archive
// containing XML files that define the document structure, content, and styling.
//
// This library focuses on:
//   - Reading IDML files with full fidelity
//   - Writing IDML files that InDesign can open
//   - Preserving all content during roundtrip operations
//   - Providing a clean, type-safe API for document manipulation
//
// Basic usage:
//
//	// Read an IDML file
//	pkg, err := idml.Read("document.idml")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Write it back
//	err = idml.Write(pkg, "output.idml")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// The library is designed to handle IDML files in phases:
//   - Phase 1: Raw roundtrip (read and write without full parsing)
//   - Phase 2: Full parsing with type-safe structures
//   - Phase 3: Content modification API
//   - Phase 4: IDMS snippet export
package idml
