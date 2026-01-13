# IDML Build - Architecture Documentation

**Project:** idmlbuild  
**Language:** Go 1.23+  
**Purpose:** Production-grade IDML (InDesign Markup Language) parsing and manipulation library

---

## Table of Contents

1. [Overview](#overview)
2. [Core Architecture](#core-architecture)
3. [Package Structure](#package-structure)
4. [Data Models](#data-models)
5. [Design Patterns](#design-patterns)
6. [Testing Strategy](#testing-strategy)
7. [Development Guidelines](#development-guidelines)

---

## Overview

### What is IDML?

IDML (InDesign Markup Language) is Adobe InDesign's XML-based file format. It's essentially a ZIP archive containing:
- **designmap.xml** - Document manifest and structure
- **Stories/** - Text content XML files
- **Spreads/** - Page layout XML files
- **Resources/** - Styles, fonts, graphics definitions
- **META-INF/** - Package metadata

### Project Goals

1. **Parse** IDML files with 100% fidelity
2. **Manipulate** document structure programmatically
3. **Generate** valid IDML files from Go structs
4. **Preserve** unknown elements (forward compatibility)
5. **Production-ready** code with comprehensive tests

---

## Core Architecture

### Layered Design

```
┌─────────────────────────────────────┐
│   High-Level APIs (Future)         │
│   - Document manipulation           │
│   - Style management                │
│   - Content creation                │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│   pkg/idml (Current Focus)          │
│   - Document struct (designmap.xml) │
│   - Parse/Marshal functions         │
│   - Type-safe element access        │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│   internal/xmlutil                  │
│   - XML utilities                   │
│   - Namespace handling              │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│   Go Standard Library               │
│   - encoding/xml                    │
│   - archive/zip                     │
└─────────────────────────────────────┘
```

### Key Principles

1. **Separation of Concerns**
   - Document parsing (pkg/idml)
   - ZIP handling (read.go, write.go)
   - XML utilities (internal/xmlutil)

2. **Forward Compatibility**
   - Unknown elements preserved via `RawXMLElement`
   - No data loss during roundtrips
   - Future InDesign versions supported

3. **Type Safety**
   - Explicit structs for major elements
   - Compile-time checks
   - IDE autocomplete support

4. **Zero Dependencies**
   - Only Go standard library
   - No external XML/ZIP parsers
   - Minimal maintenance burden

---

## Package Structure

```
idmlbuild/
├── pkg/
│   ├── idml/              # Core coordinator + public API
│   │   ├── package.go     # Package coordinator (file I/O, caching)
│   │   ├── read.go        # ZIP reading
│   │   ├── write.go       # ZIP writing
│   │   ├── selection.go   # Selection API for IDMS export
│   │   ├── resourcemgr*.go  # Resource management
│   │   ├── errors.go      # Error types
│   │   └── *_test.go      # Tests
│   │
│   ├── common/            # Shared types across all domains
│   │   └── types.go       # RawXMLElement, Properties, PathGeometry
│   │
│   ├── document/          # Document types (designmap.xml)
│   │   ├── document.go    # Document + 30 types (Language, Layer, etc.)
│   │   ├── metadata.go    # ProcessingInstruction, DocumentWithMetadata
│   │   ├── parse.go       # ParseDocument*(), MarshalDocument*()
│   │   └── designmap.go   # Legacy Designmap types (deprecated)
│   │
│   ├── spread/            # Spread types (Spreads/*.xml)
│   │   ├── spread.go      # Spread, SpreadElement, Page
│   │   ├── pageitems.go   # TextFrame, Rectangle, Oval, Image, etc.
│   │   ├── graphicline.go # GraphicLine
│   │   └── parse.go       # ParseSpread(), MarshalSpread()
│   │
│   ├── story/             # Story types (Stories/*.xml)
│   │   ├── story.go       # Story, StoryElement, ParagraphStyleRange
│   │   └── parse.go       # ParseStory(), MarshalStory()
│   │
│   ├── resources/         # Resource types (Resources/*.xml)
│   │   ├── graphics.go    # GraphicFile, Color, Swatch, Gradient, etc.
│   │   ├── fonts.go       # FontsFile, FontFamily, Font, etc.
│   │   ├── styles.go      # StylesFile, CharacterStyle, ParagraphStyle
│   │   ├── parse_*.go     # Parsing functions for each resource type
│   │   └── errors.go      # Resource-specific errors
│   │
│   ├── analysis/          # Dependency tracking for IDMS export
│   │   └── tracker.go     # DependencyTracker, DependencySet
│   │
│   └── idms/              # IDMS snippet export functionality
│       ├── package.go     # IDMS Package (single XML file)
│       ├── exporter.go    # Exporter, selection → IDMS
│       └── *.go           # IDMS-specific I/O and marshaling
│
├── internal/
│   ├── xmlutil/           # XML utilities
│   └── testutil/          # Test helpers
│
├── cmd/cli/               # Interactive CLI tool with Bubbletea TUI
│
├── testdata/              # Test fixtures
│   ├── *.idml             # Sample IDML files
│   ├── *.xml              # Sample XML files
│   └── golden/            # Expected outputs
│
└── docs/
    ├── ARCHITECTURE.md    # This file
    └── EPIC*.md           # Epic documentation
```

**Architecture After Epic 5 Refactoring:**

The package structure follows a domain-driven design that mirrors the IDML file structure:
- **pkg/common**: Shared types used across all domains
- **pkg/document**: designmap.xml types and parsing
- **pkg/spread**: Spread XML types and page items
- **pkg/story**: Story XML types and text content
- **pkg/resources**: Resource XML types (styles, fonts, graphics)
- **pkg/idml**: Coordinator package that orchestrates all domains
- **pkg/analysis**: Cross-domain dependency analysis
- **pkg/idms**: IDMS export using types from all packages

**Package Design:**
- Each package provides domain-specific types and parsing functions
- pkg/idml coordinates all packages and provides the main public API
- Direct imports from domain packages (document, story, spread, resources)
- Zero circular dependencies through the common package pattern

### File Responsibilities

**document.go** (2,463 lines)
- Document struct definition (50+ attributes, 14+ element types)
- 35+ supporting structs (Language, Layer, Section, etc.)
- All element type definitions
- Core data model

**document_parse.go**
- `ParseDocument(data []byte) (*Document, error)`
- `MarshalDocument(doc *Document) ([]byte, error)`
- Namespace handling
- XML parsing/generation

**read.go**
- `OpenIDML(path string) (*Package, error)`
- ZIP file reading
- File extraction
- MIME type handling

**write.go**
- `WriteIDML(pkg *Package, path string) error`
- ZIP file writing
- Compression
- File ordering (mimetype first, uncompressed)

**errors.go**
- Custom error types
- Error wrapping
- Context-aware errors

---

## Data Models

### Core Struct: Document

The `Document` struct represents designmap.xml - the manifest file in an IDML package.

```go
type Document struct {
    // 50+ attributes (identity, layout, colors, etc.)
    XMLName    xml.Name
    Xmlns      string
    DOMVersion string
    Self       string
    // ... many more ...

    // Explicit child elements (14 categories, 35+ types)
    Properties      *Properties
    Languages       []Language
    GraphicResource *ResourceRef
    // ... all major elements ...

    // Catch-all for unknown elements
    OtherElements []RawXMLElement
}
```

### Element Categories

**Implemented (95%+ of typical documents):**

1. **Metadata** - Properties, Labels, KeyValuePairs
2. **Languages** - Localization settings
3. **Resources** - Graphic, Fonts, Styles, Preferences, Tags
4. **Content** - MasterSpreads, Spreads, Stories, BackingStory
5. **Layout** - Layers, Sections
6. **Typography** - NumberingList, NamedGrid, GridDataInformation
7. **Users** - DocumentUser (collaboration)
8. **Colors** - ColorGroup, ColorGroupSwatch
9. **Bullets** - ABullet definitions
10. **Workflow** - Assignment (InCopy)
11. **Variables** - TextVariable (10 types)

**Not Yet Implemented (rare elements):**

- KinsokuTable, MojikumiTable (CJK typography)
- CrossReferenceFormat
- ConditionalTextPreference
- IndexingSortOption
- Various export preferences

All unimplemented elements are preserved via `OtherElements []RawXMLElement`.

---

## Design Patterns

### 1. RawXMLElement Pattern

**Problem:** Need to preserve unknown XML elements without explicit structs.

**Solution:**
```go
type RawXMLElement struct {
    XMLName xml.Name
    Attrs   []xml.Attr `xml:",any,attr"`
    Content []byte     `xml:",innerxml"`
}

type Document struct {
    // Explicit elements
    Languages []Language
    
    // Catch-all
    OtherElements []RawXMLElement `xml:",any"`
}
```

**Benefits:**
- No data loss during roundtrips
- Forward compatible with future InDesign versions
- Easy to migrate elements from catch-all to explicit

### 2. Resource Reference Pattern

**Problem:** IDML uses namespaced references to external XML files.

**Solution:**
```go
type ResourceRef struct {
    XMLName xml.Name
    Src     string `xml:"src,attr"`
}

// In Document:
GraphicResource *ResourceRef `xml:"http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging Graphic,omitempty"`
```

**Benefits:**
- Type-safe references
- Consistent handling of all resource types
- Easy to add new resource types

### 3. Preference Pattern

**Problem:** TextVariables have different preference types based on variable type.

**Solution:**
```go
type TextVariable struct {
    VariableType string
    
    // Only one of these will be populated
    DatePreference       *DateVariablePreference
    FileNamePreference   *FileNameVariablePreference
    PageNumberPreference *PageNumberVariablePreference
    // ...
    
    OtherElements []RawXMLElement `xml:",any"`
}
```

**Benefits:**
- Type-safe preference access
- Compiler catches invalid combinations
- Extensible for new preference types

### 4. Properties Container Pattern

**Problem:** Many elements have a Properties child with various content.

**Solution:**
```go
type Properties struct {
    XMLName       xml.Name `xml:"Properties"`
    Label         *Label
    OtherElements []RawXMLElement `xml:",any"`
}

// Used by: Document, Layer, Section, Assignment, ABullet, etc.
```

**Benefits:**
- Consistent pattern across all elements
- Handles known properties (Label) + unknown
- Easy to extend

---

## Testing Strategy

### Test Coverage

**30+ test functions covering:**

1. **Parsing Tests** - Verify struct population
2. **Roundtrip Tests** - Parse → Marshal → Parse equality
3. **Golden File Tests** - Byte-perfect ZIP preservation
4. **XML Structural Tests** - XML tree comparison
5. **Element-Specific Tests** - Each element type validated

### Test Data

**testdata/ contains:**
- `plain.idml` - Minimal valid IDML
- `example.idml` - Complex real-world IDML
- `designmap.xml` - Full manifest XML
- `designmap_minimal.xml` - Minimal manifest
- Individual XML files for specific tests

### Running Tests

```bash
# All tests
go test ./pkg/idml/

# Verbose output
go test -v ./pkg/idml/

# Specific test
go test -v -run TestDocumentRoundtrip ./pkg/idml/

# Coverage
go test -cover ./pkg/idml/
```

### Golden File Testing

Golden files ensure byte-perfect roundtrips:

```go
func TestGoldenRoundtrip(t *testing.T) {
    // Read original IDML
    original := readIDML("plain.idml")
    
    // Parse and re-write
    pkg, _ := idml.OpenIDML("plain.idml")
    idml.WriteIDML(pkg, "output.idml")
    
    // Compare byte-for-byte
    assertIdentical(t, original, output)
}
```

---

## Development Guidelines

### Adding New Elements

**1. Research the element structure:**
```bash
grep -A 10 "ElementName" testdata/*.xml
```

**2. Define the struct:**
```go
type NewElement struct {
    XMLName       xml.Name `xml:"NewElement"`
    Self          string   `xml:"Self,attr"`
    Name          string   `xml:"Name,attr"`
    // ... attributes ...
    OtherElements []RawXMLElement `xml:",any"`
}
```

**3. Add to Document:**
```go
type Document struct {
    // ... existing fields ...
    
    // Step N: New Elements
    NewElements []NewElement `xml:"NewElement,omitempty"`
    
    OtherElements []RawXMLElement `xml:",any"`
}
```

**4. Update catch-all comment:**
Remove element from OtherElements documentation.

**5. Write tests:**
```go
func TestDocumentNewElements(t *testing.T) { /* ... */ }
func TestDocumentNewElementsRoundtrip(t *testing.T) { /* ... */ }
```

**6. Run tests:**
```bash
go test ./pkg/idml/
```

### Code Style

**Struct Tags:**
```go
// Attributes
Self string `xml:"Self,attr"`
Name string `xml:"Name,attr,omitempty"` // Optional

// Child elements
Languages []Language `xml:"Language,omitempty"` // Multiple
Properties *Properties `xml:"Properties,omitempty"` // Optional single

// Namespaced elements
Graphic *ResourceRef `xml:"http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging Graphic,omitempty"`

// Catch-all
OtherElements []RawXMLElement `xml:",any"`
```

**Documentation:**
```go
// MyElement represents a document element for XYZ.
// This element controls ABC and is used for DEF.
type MyElement struct {
    // Identification
    Self string `xml:"Self,attr"` // Unique identifier (e.g., "abc123")
    Name string `xml:"Name,attr"` // Display name
    
    // Configuration
    Enabled string `xml:"Enabled,attr,omitempty"` // Enable feature ("true"/"false")
    
    // Child elements
    Properties *Properties `xml:"Properties,omitempty"` // Optional properties
    
    // Catch-all
    OtherElements []RawXMLElement `xml:",any"`
}
```

**Naming Conventions:**
- Structs: PascalCase (e.g., `DocumentUser`)
- Fields: PascalCase (e.g., `PageNumberStart`)
- Functions: PascalCase for public, camelCase for private
- Test functions: `TestSubject_Scenario` or `TestSubject`

### Error Handling

```go
// Wrap errors with context
if err := doSomething(); err != nil {
    return nil, &Error{
        Op:  "operation name",
        Err: err,
    }
}

// Custom error type
type Error struct {
    Op  string // Operation that failed
    Err error  // Underlying error
}
```

### Performance Considerations

1. **Lazy parsing** - Only parse what's needed
2. **Streaming** - Use io.Reader/Writer where possible
3. **Memory efficiency** - Avoid duplicate data storage
4. **Minimal allocations** - Reuse buffers when appropriate

---

## Future Roadmap

### Phase 3: Content Parsing (Next)

**Stories:**
- Parse Story XML files
- Text content and formatting
- Paragraph/character styles

**Spreads:**
- Parse Spread XML files
- Page layout and frames
- Positioned elements

**Styles:**
- Parse Styles.xml
- Character styles
- Paragraph styles
- Object styles

### Phase 4: High-Level APIs

**Document manipulation:**
```go
// Create new document
doc := idml.NewDocument()

// Add story
story := doc.AddStory("My Story")
story.AddParagraph("Hello, world!")

// Add page
spread := doc.AddSpread()
page := spread.AddPage()

// Save
doc.SaveAs("output.idml")
```

**Content extraction:**
```go
// Get all text
text := doc.ExtractAllText()

// Find specific content
results := doc.Search("keyword")
```

**Style management:**
```go
// List styles
styles := doc.ListParagraphStyles()

// Apply style
paragraph.ApplyStyle("Heading 1")

// Create style
doc.CreateParagraphStyle("Custom", styleOpts)
```

---

## Performance Metrics

**Current (Phase 2):**
- Parse designmap.xml: ~0.5ms (9KB file)
- Full IDML open: ~20ms (includes ZIP extraction)
- Memory: ~2MB for typical document
- Test suite: <0.4s for 30+ tests

**Goals (Phase 3):**
- Full document parse: <100ms
- Memory: <10MB for typical document
- Streaming support for large files

---

## API Stability

**Current Status:** Alpha

- pkg/idml API is relatively stable
- Document struct may evolve
- Breaking changes possible until v1.0

**Versioning:**
- Follow semantic versioning
- Major version bump for breaking changes
- Minor version for new features
- Patch version for bug fixes

---

## Contributing

### Setup

```bash
# Clone
git clone https://github.com/dimelords/idmlbuild
cd idmlbuild

# Test
go test ./...

# Lint (if golangci-lint installed)
golangci-lint run
```

### Pull Request Process

1. Write tests first
2. Implement feature
3. Run full test suite
4. Update documentation
5. Submit PR with clear description

### Code Review Checklist

- [ ] Tests pass
- [ ] Documentation updated
- [ ] No breaking changes (or justified)
- [ ] Code follows style guide
- [ ] Error handling included
- [ ] Performance considered

---

## References

### IDML Specification

- [Adobe InDesign Interchange (INX) & Markup (IDML)](https://www.adobe.com/devnet/indesign/sdk.html)
- [IDML Cookbook](https://wwwimages2.adobe.com/content/dam/acom/en/devnet/indesign/sdk/cs6/idml/idml-cookbook.pdf)
- [IDML Specification](https://wwwimages2.adobe.com/content/dam/acom/en/devnet/indesign/sdk/cs6/idml/idml-specification.pdf)

### Go Resources

- [encoding/xml package](https://pkg.go.dev/encoding/xml)
- [archive/zip package](https://pkg.go.dev/archive/zip)
- [Testing in Go](https://go.dev/doc/tutorial/add-a-test)

---

## License

[Your License Here]

---

**Last Updated:** 2025-11-24
**Version:** Epic 5 Complete - Domain-driven Architecture
**Status:** Production-ready for full IDML/IDMS operations with clean package structure
