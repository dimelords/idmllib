# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

### Changed

### Deprecated

### Removed

### Fixed

### Security

## [3.2.1] - 2026-10-22

### Added
- Typed hyperlinks in `document.Document`: `HyperlinkURLDestinations`, `HyperlinkPageItemSources` and `Hyperlinks`, which were kept as raw `OtherElements`. `Hyperlink.Destination` returns the destination named in its Properties, `Document.HyperlinkURL` resolves a hyperlink to its URL (by that destination, else by `DestinationUniqueKey`), and `Document.PageItemHyperlinks` / `Document.PageItemURLs` return the hyperlinks and URLs of a page item, such as an ad frame that links to a web page. Roundtrip is unchanged: child order is recorded and replayed, and the empty `EpubAriaRole` and `HypherlinkAltText` attributes InDesign writes stay in `OtherAttrs` so they are not dropped

## [3.2.0] - 2026-10-15

### Added
- `fit.StoryFrameConstraints`: frame constraints that read the story's own local formatting (`PointSize`, `Leading` including `Auto`, `AppliedFont`, `FontStyle` on its first text-bearing `CharacterStyleRange`, then its `ParagraphStyleRange`) ahead of the applied paragraph style. `FrameConstraints` reads the style only, so a hand-sized headline got the style's size and leading and too few lines
- `fit.Ascender`, an optional interface a `Measurer` implements to report its font's ascent, and `Constraints.Measured`, which works out `MaxLines` with it. A frame's first baseline sits at the ascent by default (`FirstBaselineOffset`, now on `Constraints` with `MinFirstBaseline`), so the first line no longer costs a full leading and a display headline gets the line InDesign gives it. `Evaluate` and `Budget` measure for themselves; without an `Ascender` the count is unchanged
- `Constraints.WordSpacing` and `fit.StyleWordSpacing`: the style's `DesiredWordSpacing`, followed through `BasedOn` and overridable on the story's `ParagraphStyleRange`, is applied to every space `fit` measures. Headline styles commonly set 90%, and full-width spaces ran their lines a few percent wide
- `Constraints.LetterSpacing` and `fit.StyleLetterSpacing`: the style's `DesiredLetterSpacing`, resolved like word spacing (an explicit `0` is a value and stops the `BasedOn` walk), is added after every character `fit` measures. A real headline style's `-5` was the last 2% between this package's widths and the lines InDesign sets
- `fit.Result.Close` and `fit.CloseMargin`: `Evaluate` reports when its verdict would flip with every width 3% wider or narrower, the margin measured on a real headline between this package's widths and what a composer sets (kerning, and letter spacing and glyph scaling squeezed on ragged lines). A close verdict should be settled by rendering

### Changed
- `fit.Evaluate` no longer rejects a candidate because the spine of a spread runs through a word. Text that otherwise fits now has `Fits: true`, with a `fit.Warning` (`Why: word_on_fold`, the word and a reason) in the new `Result.Warnings`. InDesign sets such a line, and real pages ship with a word across the fold, so rewording is an editor's call. Callers that treated `Why == FailWordOnFold` as a rejection should read `Warnings` instead

## [2.2.0] - 2026-01-20

### Added
- PDF support for Rectangle page items with `PDF` struct and `PDFAttribute`
- `FrameContentBase` shared structure for Image and PDF content types
- Support for extracting PDF links from `Rectangle/PDF/Link@LinkResourceURI`
- PDF-specific color policy attributes: `GrayVectorPolicy`, `RGBVectorPolicy`, `CMYKVectorPolicy`
- PDF page selection and cropping attributes via `PDFAttribute`

### Changed
- Refactored `Image` and `PDF` structs to use shared `FrameContentBase` (eliminates 13 duplicated attributes)
- `Rectangle` now supports both `Image` and `PDF` child elements
- Updated test fixtures to reflect new attribute ordering in marshaled XML

### Fixed
- Fixed test compilation errors after FrameContentBase refactoring
- Updated golden test files for new XML attribute ordering

## [2.1.0] - 2025-01-15

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

