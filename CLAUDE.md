# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**idmlbuild** is a Go library for reading, writing, and manipulating Adobe InDesign IDML (InDesign Markup Language) files. IDML files are ZIP archives containing XML documents that define InDesign layouts, text, styles, and graphics.

- **Language**: Go 1.25+ for the library (go.mod; downstream services build on 1.25), Go 1.26 for cmd/. .go-version is the local dev toolchain. CI tests 1.25, 1.26 and 1.27.
- **Type**: Library (not an application)
- **Repository**: github.com/dimelords/idmlbuild

## Core Architecture

### Package Structure

```
pkg/
├── idml/          # Core coordinator + I/O + public API
│   ├── package.go      # Package coordinator with file I/O
│   ├── read.go         # Read IDML from ZIP
│   ├── write.go        # Write IDML to ZIP
│   ├── selection.go    # Selection API for IDMS export
│   └── resourcemgr*.go # Resource management
│
├── common/        # Shared types used across all domains
│   └── types.go        # RawXMLElement, Properties, PathGeometry, etc.
│
├── document/      # Document types (designmap.xml)
│   ├── document.go     # Document + 30 types (Language, Layer, etc.)
│   ├── metadata.go     # ProcessingInstruction, File (designmap wrapper)
│   ├── parse.go        # ParseDocument*() functions
│   └── designmap.go    # Legacy Designmap types (deprecated)
│
├── spread/        # Spread types (Spreads/*.xml)
│   ├── spread.go       # File (wrapper), Spread, Page
│   ├── pageitems.go    # TextFrame, Rectangle, Oval, Image, etc.
│   ├── graphicline.go  # GraphicLine
│   ├── parse.go        # ParseSpread(), MarshalSpread()
│   └── group.go        # Group, Button, State
│
├── story/         # Story types (Stories/*.xml)
│   ├── story.go        # File (wrapper), Story, ParagraphStyleRange, etc.
│   └── parse.go        # ParseStory(), MarshalStory()
│
├── resources/     # Resource types (Resources/*.xml)
│   ├── graphics.go     # GraphicFile, Color, Swatch, Gradient, etc.
│   ├── fonts.go        # FontsFile, FontFamily, Font, etc.
│   ├── styles.go       # StylesFile, CharacterStyle, ParagraphStyle, etc.
│   ├── parse_*.go      # Parsing functions for each resource type
│   └── errors.go       # Resource-specific errors
│
├── analysis/      # Dependency tracking for IDMS export
│   └── tracker.go      # DependencyTracker, DependencySet
│
├── idms/          # IDMS snippet export functionality
│   ├── package.go      # IDMS Package (single XML file)
│   ├── exporter.go     # Exporter, selection → IDMS
│   └── read.go/write.go # IDMS I/O
│
internal/
├── xmlutil/       # XML comparison and formatting utilities
└── testutil/      # Test helpers and golden file testing

cmd/cli/           # Interactive CLI tool with Bubbletea TUI
```

**Key architectural decisions (Epic 5 refactoring):**
- **Domain-driven packages**: Each package mirrors IDML file structure
- **Clear separation**: Common types in pkg/common, domain types in their packages
- **Coordinator pattern**: pkg/idml/Package orchestrates all domains
- **Zero circular dependencies**: Through the common package pattern

### Key Concepts

**IDML Structure**: ZIP archive with specific requirements:
- `mimetype` file MUST be first and uncompressed (`zip.Store`)
- `designmap.xml` is the main manifest
- Content in `Stories/*.xml`, `Spreads/*.xml`, `Resources/*.xml`
- All XML uses `http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging` namespace

**Central Types**:
- `idml.Package` (pkg/idml/package.go): Main coordinator holding all IDML files
  - Files cached on-demand with getters: `Document()`, `Story()`, `Spread()`, `Styles()`, etc.
  - Modification methods: `AddStory()`, `UpdateStory()`, `RemoveStory()`, etc.
  - Returns types from domain packages (document.Document, story.Story, etc.)
- `document.Document` (pkg/document/document.go): Main document structure (designmap.xml)
- `spread.Spread` (pkg/spread/spread.go): the <Spread> element; `spread.File` is the file wrapper
- `story.Story` (pkg/story/story.go): Story with text content (Stories/*.xml)
- `resources.StylesFile` (pkg/resources/styles.go): Style definitions (Resources/Styles.xml)
- `common.RawXMLElement` (pkg/common/types.go): Catch-all for forward compatibility

**Parsing Strategy**:
- Each domain package has its own Parse*/Marshal* functions
- Custom `MarshalXML`/`UnmarshalXML` for namespace handling (idPkg: prefix)
- `common.RawXMLElement` catch-all for forward compatibility
- Every element struct has an `OtherAttrs []xml.Attr` (`xml:",any,attr"`) catch-all; types with several child kinds record and replay child order (`common.ChildOrder`, `zz_ordered.go`). See docs/FIDELITY.md before touching any parser.
- `Write` only re-marshals files marked modified; edits made through a getter pointer need `Package.MarkModified(path)`.
- Master spreads: `Package.MasterSpreads()` / `MasterSpread(path)`; the Spread type handles both root names.
- Parsing functions:
  - `document.ParseDocument()` - Parse designmap.xml
  - `spread.ParseSpread()` - Parse Spread_*.xml and MasterSpread_*.xml files
  - `story.ParseStory()` - Parse Story_*.xml files
  - `resources.ParseStylesFile()` - Parse Styles.xml
  - `resources.ParseFontsFile()` - Parse Fonts.xml
  - `resources.ParseGraphicFile()` - Parse Graphic.xml

**IDMS Export** (pkg/idms/):
- `Exporter` builds minimal snippets from selections
- `analysis.DependencyTracker` collects required resources
- IDMS embeds resources inline instead of separate files
- Uses types from document/, spread/, story/, resources/ packages

**Import Guidelines**:
- Import specific packages as needed for your use case
  - `pkg/idml` for main API (Read, Write, Package operations)
  - `pkg/document` for Document types
  - `pkg/spread` for page item types
  - `pkg/story` for text content types
  - `pkg/resources` for styles, colors, fonts
  - `pkg/common` for shared types

### Critical Implementation Rules

1. **Mimetype Handling**:
```go
// MUST be first in ZIP and uncompressed
mimeWriter, _ := w.CreateHeader(&zip.FileHeader{
    Name:   "mimetype",
    Method: zip.Store,  // No compression!
})
```

2. **XML Namespace Marshaling**:
```go
// Use custom MarshalXML for idPkg prefix
start.Name = xml.Name{Local: "idPkg:Story"}
start.Attr = append(start.Attr, xml.Attr{
    Name:  xml.Name{Local: "xmlns:idPkg"},
    Value: "http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging",
})
```

3. **Library Design**:
- Use `internal/` for implementation details
- Keep public API minimal and stable
- Never export internal types through public methods

## Common Development Tasks

### Building and Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./pkg/idml -v

# Run single test
go test ./pkg/idml -run TestRoundtripWithParsing -v

# Update golden files (when XML output changes intentionally; run the fidelity tests first)
go test ./pkg/idml ./pkg/idms -update

# Fidelity tests: parse -> marshal must be structurally identical for every fixture
go test ./pkg/idml ./pkg/idms -run Fidelity
```

### CLI Tool

```bash
# Build CLI
(cd cmd && go build -o ../bin/idmlbuild ./cli)

# Run interactive TUI
./bin/idmlbuild

# Debug export tool
(cd cmd && go run ./debug-export)
```

### Test Files

- `testdata/plain.idml` - Simple single-page document
- `testdata/example.idml` - Complex multi-page document with graphics
- `testdata/golden/*.xml.golden` - Expected XML output for validation

## Testing Strategy

**Golden File Testing**: Use `github.com/sebdah/goldie/v2` for XML roundtrip validation
- Files stored in `testdata/golden/`
- Update with `UPDATE_GOLDEN=1 go test`

**Roundtrip Tests**: Core validation pattern
```go
// Read → Parse → Write → Read → Compare
pkg1, _ := idml.Read("test.idml")
idml.Write(pkg1, "output.idml")
pkg2, _ := idml.Read("output.idml")
// Compare pkg1 vs pkg2 or use golden files
```

**Test Utilities** (internal/testutil/):
- `CompareZIPStructure()` - Compare file lists
- `CompareXMLSemantically()` - Structural XML comparison
- Test helpers available but prefer golden file tests

## Development Workflow

### Baby Steps Approach
1. Make it work (correctness first)
2. Make it right (clean code)
3. Make it fast (optimize only if needed)

### For New Features
1. Check existing CLAUDE.md for context
2. Write tests first (TDD preferred)
3. Implement minimal working version
4. Validate with InDesign (can it open the file?)
5. Add golden file tests
6. Update documentation

### XML Handling Patterns

**Preserve unknown elements**:
```go
type Element struct {
    KnownField string
    OtherElements []RawXMLElement `xml:",any"`  // Forward compatibility
}
```

**Namespace handling**:
```go
// For parsing
XMLName xml.Name `xml:"http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging Story"`

// For marshaling (use custom MarshalXML)
// See pkg/idml/story.go or spread.go for examples
```

## Releases

The changelog is the source of truth. There is no `Release X.Y.Z` commit and no
hand-written tag.

1. Add a `## [X.Y.Z] - YYYY-MM-DD` section to CHANGELOG.md in the same PR as the
   change it describes.
2. Merge to main.
3. `.github/workflows/release.yml` runs after CI. On a green run on main it tags
   `vX.Y.Z` and publishes a GitHub release with that section as the notes.

- `## [Unreleased]` is skipped, and a version that already has a tag ends the run
  without doing anything, so re-runs and unrelated pushes are safe.
- The major must match the module path: at `/v3` only `3.y.z` may be tagged. A
  mismatched tag looks fine on GitHub and is invisible to `go get`, so the
  workflow fails instead of creating it.
- For a Go library the tag is the whole release; nothing else is published.

## Common Pitfalls

1. **Namespace bugs**: NEVER create `xmlns:_xmlns="xmlns"` attributes
   - Guarded in internal/xmlutil/metadata.go, which strips the broken declaration after marshalling - use that pattern

2. **File order**: ZIP file order matters for byte-identical roundtrips
   - Use `Package.fileOrder` to preserve original order

2b. **Never drop attributes or reorder children**: new structs must include the `OtherAttrs` catch-all and, if they have more than one child kind, a `childOrder` with delegating methods in `zz_ordered.go`. The fidelity tests will fail otherwise.

3. **IDMS vs IDML**: Different resource strategies
   - IDML: Resources in separate files (ResourceRef)
   - IDMS: Resources embedded inline in Document

4. **Premature optimization**: Focus on correctness, not performance initially

## Project Status (as of 2025-11-20)

**Epic 2 Complete**: Full parsing with roundtrip support
- ✅ Document (designmap.xml) parsing
- ✅ Story XML parsing with custom Marshal/Unmarshal
- ✅ Spread XML parsing
- ✅ Resources (Styles, Fonts, Graphics) parsing
- ✅ Golden file test infrastructure
- ✅ High-level modification API (Add/Update/Remove operations)

**Epic 3 Complete**: IDMS Export
- ✅ Selection API (SelectTextFrameByID, etc.)
- ✅ DependencyTracker with full dependency resolution
- ✅ Graphics, styles, fonts dependency tracking
- ✅ Complete exporter implementation with inline resources
- ✅ Roundtrip validation

**Epic 5 Complete**: Architecture Refactoring
- ✅ Phase 1: Common types extraction (pkg/common)
- ✅ Phase 2: Document types extraction (pkg/document)
- ✅ Phase 3: Spread types extraction (pkg/spread)
- ✅ Phase 4: Story types extraction (pkg/story)
- ✅ Phase 5: Resources types extraction (pkg/resources)
- ✅ Phase 6: External packages updated (pkg/analysis, pkg/idms)
- ✅ Phase 7: Package.go cleanup
- ✅ Phase 8: Documentation updates

**Architecture**: Domain-driven package structure
- Clean separation by IDML file structure
- Direct imports from domain packages
- Zero circular dependencies through common package pattern

**Future**:
- Performance optimization (parallel parsing, string interning)
- Context support for cancellation
- Full InDesign compatibility validation
- Additional IDML features (tables, TOC, etc.)

## Resources

- [Adobe IDML Specification](https://www.adobe.com/devnet/indesign/sdk.html)
- Project docs in `/docs/` directory
- Test coverage: ~74% (target: maintain or improve)

## Notes for AI Assistants

- This is a **library**, not an application - use `internal/` appropriately
- Test everything before moving forward
- InDesign validation is critical - files must open without errors
- XML formatting and namespaces are fragile - test thoroughly
- Use existing patterns (Story/Spread parsers) for consistency
- Keep claude.md up to date
- Use Context7 to check up-to-date docs when needed for implementing new libraries or adding new features.
## API shape (2026-09-16)

Each IDML file has an element type and a file wrapper:

| File | Element (what you edit) | Wrapper (DOMVersion, declaration) |
| --- | --- | --- |
| designmap.xml | `document.Document` | `document.File` |
| Stories/*.xml | `story.Story` | `story.File` |
| Spreads/*.xml, MasterSpreads/*.xml | `spread.Spread` | `spread.File` |

`Package.Story`, `Package.Spread`, `Package.MasterSpread` and `Package.Document`
return elements. `Package.StoryFile`, `SpreadFile` and `DocumentFile` return the
wrappers. Typed resource getters: `Styles`, `Fonts`, `Graphics`, `Preferences`.
Raw bytes for any file: `Package.FileData(path)`.

Typed resource getters also include `Tags` for XML/Tags.xml. For image-heavy

`ReadOptions.Lazy` reads archive entries on demand and copies untouched ones
straight through on write; call `Package.Close` when done.
documents, `ReadOptions.StreamingMode` roughly halves held memory; see
docs/STREAMING_MODE_DESIGN.md.

There is no generic `ResourceFile` any more; every resource file has a typed
form. See docs/ARCHITECTURE_PLAN.md for what is still planned.
