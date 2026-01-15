# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- New `pkg/xmp` package for XMP (Extensible Metadata Platform) metadata support
- XMP metadata parsing and extraction from IDML and IDMS files
- `XMP()` and `SetXMP()` methods on IDML and IDMS Package types
- XMP timestamp management with `UpdateTimestamps()` method
- XMP thumbnail management with `RemoveThumbnails()` and `AddThumbnail()` methods
- XMP field access operations with `GetField()` and `SetField()` methods
- Automatic XMP metadata persistence when writing IDML/IDMS files

### Changed
- Updated `.golangci.yml` to v2 configuration format for golangci-lint v2.8.0 compatibility
- Updated Go version in linter config from 1.21 to 1.23
- IDML and IDMS packages now automatically extract and preserve XMP metadata during read/write operations

### Deprecated

### Removed

### Fixed
- Fixed `.golangci.yml` configuration errors (version field, output format, deprecated linters)
- Removed deprecated linters: `typecheck`, `gofmt`, `goimports`, `gosimple`

### Security

## [2.0.0] - 2025-01-13

### Added

- Complete IDML read/write support with roundtrip fidelity
- Domain-driven package architecture mirroring IDML file structure
- Story, Spread, and Document parsing with full XML support
- ResourceManager for tracking, validating, and cleaning up resources
- Selection API for programmatically selecting elements by ID
- IDMS snippet export functionality
- DependencyTracker for analyzing element dependencies
- Golden file testing infrastructure
- CLI tool with interactive TUI interface
- Zero external dependencies (Go stdlib only)

### Changed
- **BREAKING**: Updated module path to `github.com/dimelords/idmllib/v2` following Go module versioning semantics for v2.0.0
- Updated all import statements to use v2 module path
- Updated README.md examples and installation instructions for v2

### Package Structure

- `pkg/idml` - Main coordinator and public API (file I/O, caching)
- `pkg/document` - Document/designmap types and parsing
- `pkg/story` - Story types and parsing
- `pkg/spread` - Spread types and page item parsing
- `pkg/resources` - Resource file types (Fonts, Styles, Graphics)
- `pkg/analysis` - Cross-domain dependency tracking
- `pkg/idms` - IDMS snippet export
- `pkg/common` - Shared types (Properties, RawXMLElement, etc.)
- `internal/xmlutil` - XML comparison and formatting utilities
- `internal/testutil` - Test helpers and golden file support

### Testing

- Overall test coverage: 74%
- Critical path coverage: 93.6% (pkg/analysis)
- Comprehensive roundtrip tests
- Golden file validation
- 158/161 tests passing

