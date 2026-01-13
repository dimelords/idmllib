<div align="center">
  <img src="assets/idmllib-icon.png" alt="idmllib" width="200"/>
  
  # IDML Library

  [![Go Reference](https://pkg.go.dev/badge/github.com/dimelords/idmllib.svg)](https://pkg.go.dev/github.com/dimelords/idmllib)
  [![Go Report Card](https://goreportcard.com/badge/github.com/dimelords/idmllib)](https://goreportcard.com/report/github.com/dimelords/idmllib)
  [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
  [![CI](https://github.com/dimelords/idmllib/actions/workflows/ci.yml/badge.svg)](https://github.com/dimelords/idmllib/actions/workflows/ci.yml)
  [![Security](https://github.com/dimelords/idmllib/actions/workflows/ci.yml/badge.svg?label=Security&logo=security)](https://github.com/dimelords/idmllib/security)
  
  **A Go library for reading, writing, and manipulating Adobe InDesign IDML (InDesign Markup Language) files.**
  
  Parse IDML documents and export TextFrames as InDesign Snippets (IDMS) with intelligent filtering of styles, colors, and resources.
</div>

## Status

✅ **Production Ready** - Full parsing, modification API, and IDMS export

### Current Capabilities

- ✅ Read IDML files (ZIP archive handling)
- ✅ Parse `designmap.xml` with full Document structure
- ✅ Parse Stories, Spreads, and Resources (Styles, Fonts, Graphics)
- ✅ Marshal all types back to XML with perfect roundtrip
- ✅ Content modification API (add/update/remove stories and resources)
- ✅ Dependency tracking and resource management
- ✅ Selection API for programmatic element access
- ✅ IDMS snippet export functionality
- ✅ Domain-driven architecture for maintainability

## Overview

IDML is Adobe InDesign's XML-based file format. This library provides a clean, type-safe API for working with IDML files in Go.

### Package Structure

```
pkg/
├── idml/          # Main API - Package coordinator and backward-compatible types
├── common/        # Shared types (Properties, PathGeometry, etc.)
├── document/      # Document structure (designmap.xml)
├── spread/        # Page layouts (Spreads/*.xml)
├── story/         # Text content (Stories/*.xml)
├── resources/     # Styles, fonts, and graphics (Resources/*.xml)
├── analysis/      # Dependency tracking
└── idms/          # IDMS snippet export
```

### Features

- ✅ **Epic 1: Foundation** - Core IDML read/write with ZIP handling
- ✅ **Epic 2: Modification API** - Full document manipulation capabilities
- ✅ **Epic 3: IDMS Export** - Generate InDesign snippets from selections
- ✅ **Epic 5: Architecture** - Domain-driven package structure
- 📋 **Future: High-Level APIs** - Simplified document creation and editing

## Installation

```bash
go get github.com/dimelords/idmllib
```

## Usage

### Reading and Inspecting IDML Files

```go
package main

import (
    "log"
    "github.com/dimelords/idmllib/pkg/idml"
)

func main() {
    // Read an IDML file
    pkg, err := idml.Read("document.idml")
    if err != nil {
        log.Fatal(err)
    }

    // Access parsed document structure
    doc, err := pkg.Document()
    if err != nil {
        log.Fatal(err)
    }

    // Inspect document properties
    log.Printf("Document version: %s", doc.Version)
    log.Printf("Stories: %d", len(doc.Stories))
    log.Printf("Spreads: %d", len(doc.Spreads))

    // Access a story
    story, err := pkg.Story("Stories/Story_u123.xml")
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Story has %d paragraph ranges",
        len(story.StoryElement.ParagraphStyleRanges))

    // Access a spread
    spread, err := pkg.Spread("Spreads/Spread_ue6.xml")
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Spread has %d text frames", len(spread.InnerSpread.TextFrames))
}
```

### Modifying Content

```go
package main

import (
    "log"
    "github.com/dimelords/idmllib/pkg/idml"
    "github.com/dimelords/idmllib/pkg/story"
)

func main() {
    pkg, err := idml.Read("document.idml")
    if err != nil {
        log.Fatal(err)
    }

    // Create a new story
    newStory := &story.Story{}
    newStory.StoryElement.Self = "u1234"
    newStory.StoryElement.ParagraphStyleRanges = []story.ParagraphStyleRange{
        {
            AppliedParagraphStyle: "ParagraphStyle/$ID/NormalParagraphStyle",
            CharacterStyleRanges: []story.CharacterStyleRange{
                story.NewCharacterStyleRange(
                    "CharacterStyle/$ID/[No character style]",
                    "Hello, World!",
                ),
            },
        },
    }

    // Add the story to the package
    err = pkg.AddStory("Stories/Story_u1234.xml", newStory)
    if err != nil {
        log.Fatal(err)
    }

    // Save the modified document
    err = idml.Write(pkg, "output.idml")
    if err != nil {
        log.Fatal(err)
    }
}
```

### Resource Management

```go
package main

import (
    "log"
    "github.com/dimelords/idmllib/pkg/idml"
)

func main() {
    pkg, err := idml.Read("document.idml")
    if err != nil {
        log.Fatal(err)
    }

    // Create a resource manager
    rm := idml.NewResourceManager(pkg)

    // Find orphaned resources (unused styles, colors, etc.)
    report := rm.FindOrphans()
    log.Printf("Found %d orphaned styles", len(report.OrphanedStyles))
    log.Printf("Found %d orphaned colors", len(report.OrphanedColors))

    // Clean up orphaned resources
    cleanupReport := rm.CleanupOrphans()
    log.Printf("Removed %d unused resources", cleanupReport.TotalRemoved)

    // Save the cleaned document
    err = idml.Write(pkg, "cleaned.idml")
    if err != nil {
        log.Fatal(err)
    }
}
```

### Exporting IDMS Snippets

```go
package main

import (
    "log"
    "github.com/dimelords/idmllib/pkg/idml"
    "github.com/dimelords/idmllib/pkg/idms"
)

func main() {
    // Read source IDML document
    pkg, err := idml.Read("document.idml")
    if err != nil {
        log.Fatal(err)
    }

    // Select elements to export
    sel := idml.NewSelection()
    textFrame, _ := pkg.SelectTextFrameByID("u1e6")
    sel.AddTextFrame(textFrame)

    // Export as IDMS snippet
    exporter := idms.NewExporter(pkg)
    snippet, err := exporter.ExportSelection(sel)
    if err != nil {
        log.Fatal(err)
    }

    // Write the snippet
    err = snippet.Write("snippet.idms")
    if err != nil {
        log.Fatal(err)
    }
}
```

## CLI Tool

The project includes an interactive CLI tool for exploring and manipulating IDML files:

```bash
# Build the CLI
go build -o bin/idmllib ./cmd/cli

# Run interactively
./bin/idmllib
```

Features:
- Browse document structure
- Inspect stories, spreads, and resources
- Export IDMS snippets
- Analyze dependencies
- Resource management

## Development

### Requirements

- Go 1.23 or later
- golangci-lint (for code quality checks)

### Code Quality and Linting

This project uses [golangci-lint](https://golangci-lint.run/) for comprehensive code quality checks.

#### Installation

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Or using homebrew on macOS
brew install golangci-lint

# Or using the install script
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2
```

#### Running Linting

```bash
# Run all linting checks
golangci-lint run

# Run with specific timeout
golangci-lint run --timeout=10m

# Run on specific files or directories
golangci-lint run ./pkg/idml/

# Fix auto-fixable issues
golangci-lint run --fix
```

#### Linting Configuration

The project uses a comprehensive `.golangci.yml` configuration that includes:

- **Error checking**: errcheck, gosec, staticcheck
- **Code quality**: gosimple, ineffassign, unused
- **Formatting**: gofmt, goimports, misspell
- **Type safety**: typecheck, govet

The configuration is designed to be strict but practical for existing codebases.

#### Continuous Integration

The project includes a comprehensive CI/CD pipeline (`.github/workflows/ci.yml`) that:

- **Runs on every push and pull request** to main/develop branches
- **Linting**: Fails the build if any linting violations are found
- **Testing**: Runs tests on multiple Go versions (1.22, 1.23) with race detection
- **Coverage**: Generates coverage reports and uploads to Codecov
- **Security**: Runs Gosec security scanning
- **Build**: Ensures all binaries build successfully

The pipeline is configured to **fail fast** - if linting or tests fail, subsequent jobs are skipped.

### Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with race detection and coverage
go test -v -race -coverprofile=coverage.out ./...

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# Run specific package tests
go test ./pkg/idml -v

# Update golden files when intentionally changing output
UPDATE_GOLDEN=1 go test ./pkg/idml
```

#### Test Cleanup Patterns

This project follows strict test cleanup patterns to keep the repository clean:

**Automatic Cleanup**:
```go
func TestExample(t *testing.T) {
    // Use t.TempDir() for automatic cleanup
    tempDir := t.TempDir()
    outputPath := filepath.Join(tempDir, "output.idml")
    
    // Test logic here...
    // No manual cleanup needed - t.TempDir() handles it
}
```

**Debug Mode with Cleanup**:
```go
func TestExampleWithDebug(t *testing.T) {
    var outputPath string
    
    if *preserveTestOutput {
        outputPath = "debug_output.idml"
        t.Cleanup(func() {
            if !t.Failed() {
                os.Remove(outputPath)
            }
        })
    } else {
        tempDir := t.TempDir()
        outputPath = filepath.Join(tempDir, "output.idml")
    }
    
    // Test logic...
}
```

**Test Artifact Guidelines**:
- Never commit test output files to the repository
- Use `t.TempDir()` for temporary files that should be cleaned up automatically
- Place persistent test data in `testdata/` directories
- Use debug flags sparingly and ensure cleanup after debugging

### Test Coverage

- Overall: ~74%
- pkg/analysis: 93.6%
- pkg/idml: 61.1%
- pkg/idms: 70.6%
- internal/xmlutil: 79.5%

### Project Structure

```
idmllib/
├── pkg/
│   ├── idml/          # Main API and coordinator
│   ├── common/        # Shared types
│   ├── document/      # Document structure
│   ├── spread/        # Page layouts
│   ├── story/         # Text content
│   ├── resources/     # Styles, fonts, graphics
│   ├── analysis/      # Dependency tracking
│   └── idms/          # IDMS export
├── internal/
│   ├── xmlutil/       # XML utilities
│   └── testutil/      # Test helpers
├── cmd/
│   ├── cli/           # Interactive CLI tool
│   └── debug-export/  # Debug utilities
├── docs/              # Documentation
└── testdata/          # Test IDML files
```

## Architecture

This library uses a domain-driven architecture where packages mirror the IDML file structure:

- **pkg/common**: Shared types used across domains
- **pkg/document**: designmap.xml types
- **pkg/spread**: Spreads/*.xml types (page items, layouts)
- **pkg/story**: Stories/*.xml types (text content)
- **pkg/resources**: Resources/*.xml types (styles, fonts, graphics)
- **pkg/idml**: Main coordinator and public API

See [ARCHITECTURE.md](ARCHITECTURE.md) for detailed documentation.

## Resources

- [Adobe IDML Specification](https://www.adobe.com/devnet/indesign/sdk.html)
- [Project Documentation](claude.md)
- [Architecture Documentation](ARCHITECTURE.md)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## Author

Fredrik Gustafsson ([@dimelords](https://github.com/dimelords))

## Changelog

### v0.3.0 (2025-01-13)
- ✅ Go 1.23 compatibility with downgraded dependencies
- ✅ Epic 5: Domain-driven architecture refactoring
- ✅ 8 focused packages for better organization
- ✅ Clean separation of concerns with zero circular dependencies
- ✅ All tests passing, zero regressions
- ✅ Comprehensive linting and CI/CD pipeline

### v0.2.0 (2025-11-16)
- ✅ Epic 2: Full modification API
- ✅ Resource management and dependency tracking
- ✅ Selection API for programmatic element access
- ✅ IDMS export foundation

### v0.1.0
- ✅ Initial release with basic IDML read/write
- ✅ Document, Story, and Spread parsing
- ✅ Full roundtrip capability