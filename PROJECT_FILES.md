# idmlbuild - Project File Documentation

This document describes the purpose and responsibilities of each file in the idmlbuild project.

## Overview

idmlbuild is a Go library for reading, writing, and manipulating Adobe InDesign IDML (InDesign Markup Language) files. It provides a clean, type-safe API for working with IDML's ZIP-based XML format, including support for IDMS snippet export.

---

## Root Directory

### Configuration Files

| File | Purpose |
|------|---------|
| `go.mod` | Go module definition with dependencies including etree (XML), bubbletea (TUI), lipgloss (styling), goldie (golden testing), and go-cmp (comparison) |
| `go.sum` | Go module checksums for dependency verification |
| `.go-version` | Specifies Go version for goenv (1.23.12) |
| `.gitignore` | Git ignore patterns for build artifacts, IDE files, and temporary files |

### Documentation

| File | Purpose |
|------|---------|
| `README.md` | Main project documentation with installation, usage examples, API reference, and development guide |
| `ARCHITECTURE.md` | Detailed architecture documentation covering design patterns, data models, testing strategy, and development guidelines |
| `CHANGELOG.md` | Version history and release notes |
| `claude.md` | Project context and instructions for AI assistance |
| `PROJECT_FILES.md` | This file - comprehensive documentation of all project files |

### Build Artifacts

| File/Directory | Purpose |
|----------------|---------|
| `bin/` | Compiled CLI binary output directory |
| `export.idms` | Sample exported IDMS snippet file |
| `idms.test` | Compiled test binary |

---

## `/pkg` - Public API Packages

The package structure follows a domain-driven design that mirrors the IDML file structure.

### `/pkg/idml` - Main API Package

The coordinator package that orchestrates all domain packages and provides the public API.

| File | Purpose |
|------|---------|
| `doc.go` | Package documentation |
| `package.go` | **Package** struct - main entry point for working with IDML files. Manages file caching, lazy parsing of documents/stories/spreads/resources, and coordinates all domain packages |
| `package_io.go` | File I/O operations extracted from package.go - handles reading, writing, and managing file data within the package |
| `package_cache.go` | Cache management operations extracted from package.go - handles cache invalidation and cleanup |
| `package_modifications.go` | Methods for adding/updating/removing stories and resources in a Package |
| `read.go` | **Read()** and **ReadWithOptions()** - Opens IDML files from disk with ZIP bomb protection, path validation, and compression ratio checks |
| `write.go` | **Write()** - Saves IDML packages to disk, handling mimetype requirements (must be first and uncompressed), marshaling modified content, and preserving file order |
| `errors.go` | Custom error types (`Error`, `ErrNotFound`) with operation context |
| `paths.go` | Path constants and helper functions (`PathDesignmap`, `PathMimetype`, `IsStoryPath()`, `IsSpreadPath()`, etc.) |
| `index.go` | Item index for O(1) page item lookups by Self ID |
| `selection.go` | **Selection** struct for programmatic element selection (used for IDMS export) |
| `interfaces.go` | **PageItem** interface definition for polymorphic page item operations |
| `metadata.go` | **MetadataFile** and **ResourceFile** types for generic XML preservation |
| `resources.go` | Generic resource file parsing and marshaling |
| `resourcemgr.go` | **ResourceManager** - finds orphaned resources, validates references |
| `resourcemgr_autoresolution.go` | Auto-resolution of missing style references |
| `resourcemgr_cleanup.go` | Cleanup methods for removing orphaned resources |
| `resourcemgr_validation.go` | Validation methods for resource consistency |
| `style_hierarchy.go` | Style hierarchy traversal utilities |
| `templates.go` | Built-in document templates for creating new IDML files |
| `fonts.go` | Font-related convenience methods for accessing font information |

**Test Files:**
| File | Purpose |
|------|---------|
| `*_test.go` | Unit tests for each corresponding source file |
| `*_properties_test.go` | Property-based tests for error handling, interface consistency, and roundtrip operations |
| `golden_test.go` | Golden file tests for byte-perfect roundtrip verification |
| `mock_test.go` | Mock implementations for testing |
| `example_test.go` | Runnable documentation examples |

### `/pkg/common` - Shared Types

Common types used across all domain packages to avoid circular dependencies.

| File | Purpose |
|------|---------|
| `doc.go` | Package documentation |
| `types.go` | **RawXMLElement** - catch-all for unknown XML (forward compatibility), **Properties** - key-value metadata container, **PathGeometry** - geometric path definitions, **GridDataInformation** - grid layout configuration |
| `errors.go` | **Error** - unified error type with package, operation, and path context. Sentinel errors (**ErrNotFound**, **ErrInvalidFormat**, etc.) and helper functions for error wrapping and checking |

**Test Files:**
| File | Purpose |
|------|---------|
| `*_test.go` | Unit tests for each corresponding source file |
| `*_properties_test.go` | Property-based tests for error construction consistency and Go error conventions |

### `/pkg/document` - Document Types

Types for InDesign document structure (designmap.xml).

| File | Purpose |
|------|---------|
| `doc.go` | Package documentation |
| `document.go` | **Document** struct (460+ lines) - represents designmap.xml with 50+ attributes. Contains types for **Language**, **Layer**, **Section**, **NumberingList**, **NamedGrid**, **ColorGroup**, **ABullet**, **Assignment**, **TextVariable**, **ResourceRef**, and inline IDMS resources |
| `document_test.go` | Unit tests for Document parsing and marshaling |
| `designmap.go` | Legacy Designmap types (deprecated, kept for backward compatibility) |
| `designmap_test.go` | Tests for legacy designmap functionality |
| `metadata.go` | **ProcessingInstruction**, **DocumentWithMetadata** - preserves XML processing instructions (<?xml ...?>) during roundtrip |
| `parse.go` | **ParseDocument()**, **ParseDocumentWithMetadata()**, **MarshalDocument()**, **MarshalDocumentWithMetadata()** - document serialization functions |

### `/pkg/spread` - Spread Types

Types for page layouts (Spreads/*.xml).

| File | Purpose |
|------|---------|
| `doc.go` | Package documentation |
| `spread.go` | **Spread**, **SpreadElement** - spread wrapper and content with dual-structure design for namespace handling. **PageItemBase** - embedded struct for common page item attributes. **Page** - individual page with guides, margins, grids. **FlattenerPreference** - transparency flattening settings |
| `spread_test.go` | Unit tests for spread parsing |
| `page_items.go` | **SpreadTextFrame** - text frame with story reference, geometry, styles. **Rectangle** - rectangular shapes with optional content. **Group** - container for grouped items |
| `graphicline.go` | **GraphicLine** - line/stroke elements with embedded PageItemBase |
| `graphicline_test.go` | Tests for graphic line parsing |
| `oval.go` | **Oval** - elliptical shapes with embedded PageItemBase |
| `oval_test.go` | Tests for oval parsing |
| `polygon.go` | **Polygon** - multi-point shapes with embedded PageItemBase |
| `polygon_test.go` | Tests for polygon parsing |
| `geometry.go` | Geometry calculation utilities for page items |
| `geometry_test.go` | Tests for geometry calculations |
| `text_capacity.go` | Text capacity calculation utilities |
| `parse.go` | **ParseSpread()**, **MarshalSpread()** - spread serialization with namespace handling |

**Test Files:**
| File | Purpose |
|------|---------|
| `*_test.go` | Unit tests for each corresponding source file |
| `example_test.go` | Runnable documentation examples |
| `pathgeometry_test.go` | Tests for path geometry handling |
| `rectangle_test.go` | Tests for rectangle parsing and manipulation |

### `/pkg/story` - Story Types

Types for text content (Stories/*.xml).

| File | Purpose |
|------|---------|
| `doc.go` | Package documentation |
| `story.go` | **Story** - story file wrapper. **StoryElement** - main story content. **ParagraphStyleRange** - paragraphs with style. **CharacterStyleRange** - character-level formatting with content elements. **StoryPreference**, **InCopyExportOption** - story settings |
| `story_test.go` | Unit tests for story parsing |
| `parse.go` | **ParseStory()**, **MarshalStory()** - story serialization with custom Content/Br element handling |

### `/pkg/resources` - Resource Types

Types for styles, fonts, and graphics (Resources/*.xml).

| File | Purpose |
|------|---------|
| `doc.go` | Package documentation |
| `errors.go` | Resource-specific error types |
| `styles.go` | **StylesFile** - Styles.xml wrapper. **CharacterStyleGroup**, **CharacterStyle** - character formatting. **ParagraphStyleGroup**, **ParagraphStyle** - paragraph formatting. **ObjectStyleGroup**, **ObjectStyle** - object formatting. **CellStyleGroup**, **TableStyleGroup**, **TOCStyle** |
| `fonts.go` | **FontsFile** - Fonts.xml wrapper. **FontFamily**, **Font** - font definitions with composite font support |
| `graphics.go` | **GraphicFile** - Graphic.xml wrapper. **Color** - color definitions (RGB, CMYK, Lab, Spot). **Swatch** - color swatches. **Gradient**, **GradientStop** - gradient definitions. **StrokeStyle** - stroke/line styles. **Ink**, **MixedInk** - ink definitions |
| `parse_styles.go` | **ParseStylesFile()**, **MarshalStylesFile()** - styles serialization with namespace handling |
| `parse_fonts.go` | **ParseFontsFile()**, **MarshalFontsFile()** - fonts serialization with namespace handling |
| `parse_graphics.go` | **ParseGraphicFile()**, **MarshalGraphicFile()** - graphics serialization with namespace handling |
| `resources_test.go` | Unit tests for resource parsing |

**Test Files:**
| File | Purpose |
|------|---------|
| `*_test.go` | Unit tests for each corresponding source file |
| `example_test.go` | Runnable documentation examples |
| `parse_test.go` | Tests for parsing functionality across all resource types |

### `/pkg/analysis` - Dependency Tracking

Tools for analyzing IDML documents and tracking dependencies (used for IDMS export).

| File | Purpose |
|------|---------|
| `doc.go` | Package documentation |
| `tracker.go` | **DependencySet** - tracks stories, styles, colors, fonts, layers, links referenced by page items. **DependencyTracker** - analyzes text frames, rectangles, ovals, polygons, graphic lines, groups and collects all their dependencies. **ResolveStyleHierarchies()** - handles style inheritance chains with circular reference detection |
| `tracker_test.go` | Unit tests for dependency tracking |
| `hierarchy_test.go` | Tests for style hierarchy analysis |
| `example_test.go` | Runnable documentation examples |

**Test Files:**
| File | Purpose |
|------|---------|
| `*_test.go` | Unit tests for each corresponding source file |
| `example_test.go` | Runnable documentation examples |

### `/pkg/idms` - IDMS Snippet Export

Functionality for creating and working with InDesign snippet files.

| File | Purpose |
|------|---------|
| `doc.go` | Package documentation |
| `errors.go` | IDMS-specific error types |
| `package.go` | **Package** struct for IDMS files (single XML document containing embedded resources and content) |
| `exporter.go` | **Exporter** - builds IDMS snippets from Selection, orchestrates dependency analysis |
| `exporter_build.go` | Internal methods for building minimal IDMS packages with only needed resources |
| `exporter_test.go` | Unit tests for export functionality |
| `read.go` | **Read()** - reads IDMS files from disk |
| `read_test.go` | Tests for IDMS reading |
| `write.go` | **Write()** - saves IDMS files to disk |
| `templates.go` | IDMS snippet templates |
| `golden_test.go` | Golden file tests for IDMS roundtrip |
| `roundtrip_test.go` | Roundtrip verification tests |
| `graphics_test.go` | Tests for graphics in IDMS export |
| `example_test.go` | Runnable documentation examples |

**Test Files:**
| File | Purpose |
|------|---------|
| `*_test.go` | Unit tests for each corresponding source file |
| `golden_test.go` | Golden file tests for IDMS roundtrip |
| `roundtrip_test.go` | Roundtrip verification tests |
| `graphics_test.go` | Tests for graphics in IDMS export |
| `example_test.go` | Runnable documentation examples |

---

## `/internal` - Internal Packages

### `/internal/xmlutil` - XML Utilities

Consolidated XML parsing and marshaling utilities extracted from domain packages.

| File | Purpose |
|------|---------|
| `compare.go` | **CompareXML()** - structural XML comparison for testing (ignores whitespace, attribute order) |
| `compare_test.go` | Unit tests for XML comparison |
| `compare_detailed_test.go` | Detailed comparison tests |
| `format.go` | **MarshalIndentWithHeader()** - XML formatting utilities with consistent header generation |
| `format_test.go` | Tests for XML formatting |
| `metadata.go` | **ParseWithMetadata()**, **MarshalWithMetadata()** - XML processing with metadata preservation |
| `namespace.go` | **ParseWithNamespace()**, **MarshalWithNamespace()** - consistent namespace handling utilities |

**Test Files:**
| File | Purpose |
|------|---------|
| `*_test.go` | Unit tests for each corresponding source file |

### `/internal/testutil` - Test Helpers

| File | Purpose |
|------|---------|
| `testdata.go` | Test data loading utilities, path helpers |
| `golden.go` | Golden file test utilities |
| `comparison.go` | Test comparison helpers |

---

## `/cmd` - Command-Line Tools

### `/cmd/cli` - Interactive CLI Tool

A Bubbletea-based TUI for exploring and manipulating IDML files.

| File | Purpose |
|------|---------|
| `main.go` | CLI entry point with main menu loop. Routes to create document, roundtrip test, or IDMS export wizards |

### `/cmd/cli/tui` - TUI Components

| File | Purpose |
|------|---------|
| `styles.go` | Lipgloss style definitions for consistent UI appearance |
| `main_menu.go` | **MainMenu** - main menu model with options: Create Document, Roundtrip Test, Export IDMS, Exit |
| `create_document.go` | **CreateDocumentWizard** - multi-step wizard for creating new IDML documents |
| `roundtrip.go` | **RoundtripWizard** - wizard for testing IDML read/write roundtrip |
| `export_idms.go` | **ExportIDMSWizard** - wizard for exporting IDMS snippets from IDML |
| `textframe_selector.go` | **TextFrameSelector** - interactive text frame selection for IDMS export |
| `text_input.go` | Text input component for file paths and names |
| `action_menu.go` | Action menu component for selecting operations |

### `/cmd/debug-export` - Debug Utilities

| File | Purpose |
|------|---------|
| `main.go` | Debug tool for testing IDMS export functionality |

---

## `/docs` - Documentation

| File | Purpose |
|------|---------|
| `ADR-005-PACKAGE-RESTRUCTURING.md` | Architecture Decision Record for Epic 5 package refactoring |
| `CLI_TUI_ARCHITECTURE.md` | Documentation for CLI/TUI architecture |
| `EPIC-5-REFACTORING-ANALYSIS.md` | Analysis document for Epic 5 refactoring |
| `EPIC2_COMPLETION.md` | Epic 2 completion report (Modification API) |
| `EPIC2_PHASE4_COMPLETION.md` | Epic 2 Phase 4 completion details |
| `EPIC2_RESOURCE_MANAGEMENT_SPEC.md` | Specification for resource management |
| `EPIC5_COMPLETION.md` | Epic 5 completion report (Architecture refactoring) |
| `PRESETS.md` | Documentation for document presets |
| `TEMPLATES.md` | Documentation for document templates |
| `TEMPLATE_SYSTEM_COMPLETE.md` | Template system completion report |
| `TESTING.md` | Testing guidelines and strategies |
| `TEST_REPORT.md` | Test coverage and results report |
| `TODO_REVIEW.md` | Review of remaining TODO items |

---

## `/testdata` - Test Fixtures

### Sample IDML Files

| File | Purpose |
|------|---------|
| `example.idml` | Complex real-world IDML file for comprehensive testing |
| `plain.idml` | Minimal valid IDML file for basic tests |
| `tripple.idml` | Multi-spread IDML file for testing |

### Sample XML Files

| File | Purpose |
|------|---------|
| `designmap.xml` | Full designmap.xml for document parsing tests |
| `designmap_minimal.xml` | Minimal designmap for basic tests |
| `Spread_u210.xml` | Sample spread XML file |
| `story_u1d8.xml` | Sample story XML file |

### IDMS Snippet Files

| File | Purpose |
|------|---------|
| `Snippet_*.idms` | Various IDMS snippet files for testing export/import |

### Extracted IDML Structure

`/testdata/example_extracted/` - Unzipped example.idml for reference:
- `META-INF/` - container.xml, metadata.xml
- `MasterSpreads/` - Master page definitions
- `Resources/` - Fonts.xml, Graphic.xml, Preferences.xml, Styles.xml
- `Spreads/` - Page spread definitions
- `Stories/` - Text content XML files
- `XML/` - BackingStory.xml, Tags.xml
- `designmap.xml` - Document manifest
- `mimetype` - MIME type declaration

### Golden Files

`/testdata/golden/` - Expected output files for golden tests:
| File | Purpose |
|------|---------|
| `README.md` | Documentation for golden file usage |
| `example_idml_roundtrip.golden` | Expected output for example.idml roundtrip |
| `plain_idml_roundtrip.golden` | Expected output for plain.idml roundtrip |

---

## `/pkg/idml/templates` - Built-in Templates

| Directory/File | Purpose |
|----------------|---------|
| `README.md` | Template system documentation |
| `minimal/` | Minimal IDML template files |

---

## `/pkg/idms/templates` - IDMS Templates

| File | Purpose |
|------|---------|
| `snippet.xml` | Base IDMS snippet template |

---

## `/pkg/idms/testdata/golden` - IDMS Golden Files

Golden files for IDMS export verification.

---

## Key Architecture Decisions

1. **Domain-Driven Design**: Packages mirror IDML file structure (document, spread, story, resources)

2. **Forward Compatibility**: `RawXMLElement` catch-all preserves unknown XML elements for future InDesign versions

3. **Lazy Parsing with Caching**: Files parsed on-demand and cached for performance, with dual-level caching for generic and typed resource access

4. **Byte-Perfect Roundtrip**: ZIP metadata preserved, file order maintained, mimetype handled correctly

5. **Zero External Dependencies**: Only Go standard library for core functionality (TUI uses charmbracelet)

6. **Resource Reference Pattern**: Type-safe handling of namespaced IDML resource references

7. **Selection-Based Export**: IDMS export uses Selection API to choose specific page items

8. **Dependency Tracking**: Automatic collection of all resources needed for standalone snippets

9. **Unified Error Handling**: Structured error context with package, operation, and path information across all packages

10. **PageItem Interface**: Polymorphic operations on page items using embedded structs and common interface

11. **Separation of Concerns**: File I/O, caching, and coordination responsibilities clearly separated in main Package

12. **XML Utilities Consolidation**: Common XML parsing patterns extracted to internal/xmlutil for consistency

13. **Property-Based Testing**: Correctness properties validated through automated testing with 100+ iterations

14. **Dual Structure Design**: Namespace wrapper handling separated from content structure for clean marshaling
