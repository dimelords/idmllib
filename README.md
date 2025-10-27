# IDML Library

[![Go Reference](https://pkg.go.dev/badge/github.com/dimelords/idmllib.svg)](https://pkg.go.dev/github.com/dimelords/idmllib)
[![Go Report Card](https://goreportcard.com/badge/github.com/dimelords/idmllib)](https://goreportcard.com/report/github.com/dimelords/idmllib)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go](https://github.com/dimelords/idmllib/actions/workflows/test.yml/badge.svg)](https://github.com/dimelords/idmllib/actions/workflows/test.yml)

A Go library and CLI tool for working with InDesign Markup Language (IDML) files. Parse IDML documents and export TextFrames as InDesign Snippets (IDMS) with intelligent filtering of styles, colors, and resources.

## Features

- 📖 **Parse IDML** - Read and parse InDesign IDML files
- 📤 **Export IDMS** - Extract TextFrames as InDesign Snippet (IDMS) files
- 🎨 **Smart Filtering** - Include only used styles, colors, and resources
- 🔗 **Dependency Resolution** - Automatically include base styles (BasedOn, NextStyle)
- ✨ **InDesign Compatible** - Output matches InDesign's native snippet format

## Architecture

The library is organized into three main packages:

- **`idml`** - Parses IDML files and provides read-only access to stories and spreads
- **`idms`** - Exports IDMS snippets with intelligent filtering using predicates
- **`types`** - Shared data structures for IDML/IDMS documents

This separation allows for clean, flexible usage where IDML handles reading and IDMS handles export.

## Installation

### As a Library

```bash
go get github.com/dimelords/idmllib
```

### As a CLI Tool

```bash
go install github.com/dimelords/idmllib/cmd/idmllib@latest
```

Or build from source:

```bash
git clone https://github.com/dimelords/idmllib.git
cd idmllib
go build -o bin/idmllib ./cmd/idmllib
```

## Quick Start

### Library Usage

```go
package main

import (
    "log"
    
    "github.com/dimelords/idmllib/idml"
    "github.com/dimelords/idmllib/idms"
    "github.com/dimelords/idmllib/types"
)

func main() {
    // Open IDML file (read-only)
    pkg, err := idml.Open("document.idml")
    if err != nil {
        log.Fatal(err)
    }
    defer pkg.Close()

    // List all stories
    for _, story := range pkg.Stories {
        log.Printf("Story: %s\n", story.Self)
    }

    // Export using predicate (e.g., export specific TextFrame)
    exporter := idms.NewExporter(pkg)
    
    // Export TextFrame with ID "u123"
    predicate := func(tf *types.TextFrame) bool {
        return tf.Self == "u123"
    }
    
    err = exporter.ExportXML("output.idms", predicate)
    if err != nil {
        log.Fatal(err)
    }
    
    // Or export all TextFrames for a specific story
    storyID := "u222"
    storyPredicate := func(tf *types.TextFrame) bool {
        return tf.ParentStory == storyID
    }
    
    err = exporter.ExportXML("story.idms", storyPredicate)
    if err != nil {
        log.Fatal(err)
    }
}
```

### CLI Usage

#### List all stories

```bash
idmllib -idml document.idml -list
```

Output shows:
- Story ID (e.g., `u222`)
- Story self reference

#### Export a TextFrame as IDMS snippet

```bash
idmllib -idml document.idml -textframe u123 -output textframe.idms
```

The exported IDMS file includes:
- TextFrame content with all formatting
- Only the styles actually used (with dependency resolution)
- Only the colors and swatches referenced
- Only the TextFrames linked to the story
- Proper ColorGroup structure

## Command-Line Options

```
-idml string
    Path to IDML file (required)
    
-list
    List all stories in the IDML file
    
-textframe string
    TextFrame ID to export (e.g., "u123")
    
-output string
    Output IDMS file path (required with -textframe)
```

## How It Works

### IDML Structure

IDML files are ZIP archives containing:
- `designmap.xml` - Document structure and story index
- `Stories/` - Text content (Story_*.xml files)
- `Resources/Styles.xml` - Character, paragraph, and object styles
- `Resources/Graphic.xml` - Colors and swatches
- `Spreads/` - Page layouts and TextFrames

### IDMS Export Process

1. **Parse IDML** - Extract story content and structure
2. **Apply Predicate** - Select TextFrames based on custom logic
3. **Analyze Dependencies** - Find all used styles, colors, and layers
4. **Resolve Relationships** - Include base styles and referenced resources
5. **Filter Resources** - Remove unused styles, colors, and swatches
6. **Build IDMS** - Create InDesign-compatible snippet with minimal data

## API Documentation

Full API documentation is available at [pkg.go.dev](https://pkg.go.dev/github.com/dimelords/idmllib).

### Core Packages

#### idml Package
```go
// Open and parse an IDML file
pkg, err := idml.Open("document.idml")

// Access stories and spreads
stories := pkg.Stories
spreads := pkg.Spreads

// Get specific story
story, err := pkg.GetStory("u222")
```

#### idms Package
```go
// Create exporter with IDML package as reader
exporter := idms.NewExporter(pkg)

// Export with custom predicate
predicate := func(tf *types.TextFrame) bool {
    return tf.Self == "u123" || tf.ParentStory == "u222"
}
err := exporter.ExportXML("output.idms", predicate)

// Or use ExportStoryXML convenience method
err := exporter.ExportStoryXML("u222", "story.idms")
```

## Testing

Run the test suite:

```bash
go test ./...
```


With coverage:

```bash
go test -cover ./...
```

Run specific package tests:

```bash
# Test IDML parsing
go test -v ./idml

# Test IDMS export and filtering
go test -v ./idms

# Test filter logic
go test -v ./idms/filter
```

Run specific tests:

```bash
go test -v -run TestStyleFiltering ./idms
go test -v -run TestOpen ./idml
```

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Write tests for your changes
4. Ensure tests pass (`go test ./...`)
5. Commit your changes (`git commit -am 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## Package Structure

```
github.com/dimelords/idmllib/
├── types/              # Shared data structures
│   ├── idms.go        # IDMS document types
│   ├── story.go       # Story types
│   ├── spread.go      # Spread types
│   └── predicates.go  # Predicate types
├── idml/              # IDML parsing (read-only)
│   ├── package.go     # Package struct
│   ├── parser.go      # IDML parser
│   ├── reader.go      # Reader methods
│   └── resource_loader.go
├── idms/              # IDMS export with filtering
│   ├── exporter.go    # Export logic
│   └── filter/        # Filtering logic
│       ├── styles.go
│       ├── colors.go
│       └── dependencies.go
├── cmd/idmllib/       # CLI tool
└── testdata/          # Shared test fixtures
```

## Known Limitations

- Graphics and images are not included in exports (by design)
- Complex nested style groups may require additional testing
- Right-to-left languages have limited testing

## Troubleshooting

#### "WARN XML file not found file=XML/Mapping.xml"
This is normal for documents without XML structure tagging. Can be safely ignored.

#### "WARN Graphics directory not found or empty"
This is normal for text-only documents. Graphics are not included in IDMS exports.

#### Export is missing some styles
Verify the styles are actually applied in the TextFrames. The filtering removes unused styles to minimize file size.

#### No TextFrames matched the predicate
Your predicate function returned false for all TextFrames. Check that:
- The TextFrame ID is correct
- The predicate logic matches your intent
- Use `-list` to see available stories and their IDs

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Author

Fredrik Gustafsson

## Acknowledgments

- InDesign IDML specification
- Go community for excellent XML handling libraries
