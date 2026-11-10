# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Custom text variables, which hold text a document sets itself, such as a publication's name or an edition's date, and were only reachable as raw elements. `TextVariable.CustomTextPreference` (`CustomTextVariablePreference`) types InDesign's `<Properties><Contents type="string">`; `TextVariable.CustomText` / `SetCustomText` read and set the text, and `document.NewCustomTextVariable` makes one with the Self InDesign gives it. `Document.TextVariable` finds a variable by name; `Document.AddTextVariable` adds one, written right after the existing variables in a designmap that was read, where InDesign keeps them. `Package.AddCustomTextVariable` adds one to the document; `Package.SetCustomTextVariable` sets its text and the `ResultText` of its instances in every story, for readers that don't recompose (InDesign takes the text from designmap.xml)
- Text variable instances in stories. `story.TextVariableInstance` is a typed `CharacterChild`, where it used to be kept as a raw element; its `ResultText` is part of `ExtractText`. `story.NewTextVariableRange` builds the character style range InDesign writes for one, marked `PageNumberType="TextVariable"`, and `Story.TextVariableInstances` lists a story's instances
- `testdata/textvariables.idml`, written by InDesign 2026: two custom text variables placed on a page and on a parent page
- Embedded PDFs. `spread.NewPDFRectangle` builds a graphic frame with a PDF placed in it and the PDF file inside the IDML, the shape InDesign writes once a PDF's link is embedded: the bytes base64-encoded in the PDF's `Contents`, its `GraphicBounds`, a `PDFAttribute` (page, crop, transparent background) and a `Link` with `StoredState="Embedded"`. The document opens with the graphic in place and no file next to it. The PDF's scale and position in the frame are the caller's, as for `NewImageRectangle`; data that isn't a PDF, a size or scale that isn't positive and a negative page number are refused. `PDF.EmbeddedContents` decodes an embedded PDF back to its file (InDesign's several CDATA sections included) and reports a PDF that is only linked; `PDF.GraphicBounds` reads the placed box

### Changed

### Deprecated

### Removed

### Fixed

### Security

## [3.4.0] - 2026-11-09

### Added
- Stacking order across page item kinds. `Spread.StackingOrder` returns the Selfs of a spread's top-level page items, bottom to top, as the spread would be written; `Spread.SetStackingOrder` restacks them from a list naming each of them once. Items appended through the typed slices are written grouped by kind (text frames first), so until now a text frame added after a filled rectangle could only end up under it. Pages, the flattener preference and items kept as raw elements stay where they are; the typed items fill the places page items have, in the new order. A list that leaves an item out, names one twice or names something else is refused and changes nothing. Works for master spreads too

## [3.3.0] - 2026-11-05

### Added
- Typed document setup in `Resources/Preferences.xml`, which was only reachable as raw elements. `Package.DocumentSetup` / `SetDocumentSetup` read and write the page size, facing pages, pages per document, binding, bleed and slug (`DocumentSetup`; bleed and slug as top, bottom, inside, outside, with the `UniformSize` flags kept in step). `Package.BaselineGrid` / `SetBaselineGrid` do the baseline grid's start, division and whether it counts from the top of the page or the top margin. `Package.DefaultMargins` / `SetDefaultMargins` do the `MarginPreference` new pages get. Attributes these types don't model are kept, and a missing element is created
- Page geometry helpers in `spread`: `PageGeometricBounds`, and `PageItemTransform` for a left, right or single page (`PageSide`), placed in the spread the way InDesign writes it. `MarginPreference.SetMargins` takes top, bottom, inside and outside, and `MarginPreference.SetColumns` writes `ColumnCount`, `ColumnGutter` and `ColumnsPositions` from the page width: equal columns, or explicit widths that must add up to the text area
- Parent pages. `spread.NewMasterSpread` builds a master spread from `MasterSpreadOptions` (name, page size, one page or a facing pair, margins, columns); `Spread.MasterName` / `SetMasterName` read and set its name. `Package.AddMasterSpread` adds it and registers it as an `<idPkg:MasterSpread>` in `designmap.xml`. `AddSpread` registers an `<idPkg:Spread>`, which InDesign ignores for a master spread. `Package.ApplyMaster` applies a parent page to a document page and checks that both exist
- `Package.RemoveMasterSpread` removes a parent page and its designmap entry. Pages that apply it make the call fail, or with `detach` get `AppliedMaster="n"`. `Package.RemoveSpread` removes a document spread; a section starting on one of its pages moves to the next spread's first page, or the previous one's. The last spread can't be removed
- `Package.ResetXMLStructure` empties the document's XML structure down to its root element and returns how many elements it removed. A page from an XML workflow keeps tagged elements in `XML/BackingStory.xml` that point at stories and page items; once those are removed, the references point at nothing

## [3.2.4] - 2026-11-05

### Changed
- `internal/xmlutil/compare.go` moved to `internal/xmlcompare`. It compares two XML documents for
  the fidelity and roundtrip tests and is called only from tests, but `xmlutil` is reached from
  `pkg/resources`, so its `beevik/etree` and `google/go-cmp` imports sat in the build of every
  program that uses this library. `go list -deps ./pkg/...` is now the standard library only; both
  dependencies stay in `go.mod`, where the tests still need them

## [3.2.3] - 2026-11-05

### Fixed
- `NewFromTemplate` now applies `TemplateOptions.Margins`. The option was validated and then ignored: the master spread template hardcoded `Top="36" Bottom="36" Left="36" Right="36"`, so a caller asking for different margins silently got half an inch. A side left at zero still gets 36 points, so a caller that never set margins keeps the document it gets today
- `NewFromTemplate` writes `DocumentPreference` and `ViewPreference` into `Resources/Preferences.xml` instead of `designmap.xml`. That is where InDesign keeps the page setup; the old placement matched this repository's own hand-built `testdata/example.idml` rather than a real export. Documents built from the template open in InDesign 2026 and read back the page size, margins, column count and gutter they were given

## [3.2.2] - 2026-10-27

### Fixed
- `ReadOptions.StreamingMode` no longer breaks a spread where one page item has a second `<Contents>`, such as an embedded EPS whose `MetadataPacketPreference` holds its XMP packet. The second payload is kept inline by design, but the bytes from the previous stripped payload up to it were dropped, so parsing failed with "element `<Contents>` closed by `</Properties>`". Writing a modified spread had the same flaw: an empty `<Contents>` with no payload of its own was dropped with the bytes before it
- In streaming mode a payload is now also kept inline when its owner has an empty `<Contents>` as well, or when the payload itself holds a page item start tag. Restoring matches payloads by owner and could otherwise fill the wrong element, duplicating the payload

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
- ~~Zero external dependencies (Go stdlib only)~~ — not true when written. Importing `pkg/idml`
  pulled in `beevik/etree` and `google/go-cmp`, which reached the public import path through
  `internal/xmlutil`, and stayed there until the comparison helpers moved to
  `internal/xmlcompare`. Corrected 2026-11-05

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

