<div align="center">
  <img src="assets/idmllib-icon.png" alt="idmllib" width="200"/>

  # idmllib

  [![Go Reference](https://pkg.go.dev/badge/github.com/dimelords/idmllib/v3.svg)](https://pkg.go.dev/github.com/dimelords/idmllib/v3)
  [![CI](https://github.com/dimelords/idmllib/actions/workflows/ci.yml/badge.svg)](https://github.com/dimelords/idmllib/actions/workflows/ci.yml)
  [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

  **Read, write and manipulate Adobe InDesign IDML files in Go.**
</div>

IDML is InDesign's XML-based interchange format: a ZIP archive of XML documents describing a
document's text, layout, styles and graphics. idmllib parses that structure into typed Go values,
lets you modify it, and writes it back in a form InDesign opens without complaint.

Its distinguishing property is fidelity. A file read and written unchanged comes back byte for
byte, and a file you do modify keeps everything you did not touch — unknown attributes, child
order, the long decimals InDesign writes. That matters because IDML is a round-tripping format:
anything the library silently drops is data a designer loses.

## Installation

```bash
go get github.com/dimelords/idmllib/v3
```

Requires Go 1.25 or later. The interactive CLI is a separate module and requires Go 1.26.

## Quick start

```go
package main

import (
	"log"

	"github.com/dimelords/idmllib/v3/pkg/idml"
)

func main() {
	pkg, err := idml.Read("document.idml")
	if err != nil {
		log.Fatal(err)
	}

	doc, err := pkg.Document()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("DOM version: %s", doc.DOMVersion)
	log.Printf("stories: %d, spreads: %d", len(doc.Stories), len(doc.Spreads))

	st, err := pkg.Story("Stories/Story_u123.xml")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("paragraph ranges: %d", len(st.ParagraphStyleRanges))

	sp, err := pkg.Spread("Spreads/Spread_ue6.xml")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("text frames: %d", len(sp.TextFrames))
}
```

## What it does

- **Reads and writes IDML** — ZIP handling, the uncompressed `mimetype` entry, file order, the lot.
- **Parses the whole document** — `designmap.xml`, stories, spreads and master spreads, styles,
  fonts, graphics, preferences and tags, into typed structs.
- **Modifies content** — add, update and remove stories, text frames and rectangles, with optional
  validation that the styles, fonts and colors they reference actually exist.
- **Tracks resources** — find and remove orphaned styles, fonts, colors and swatches; validate that
  every reference in a document resolves.
- **Exports IDMS snippets** — select page items and write an InDesign snippet carrying exactly the
  resources those items depend on, inlined.
- **Keeps memory in hand** — streaming mode and lazy reading for image-heavy documents.

> **Concurrency:** a `Package` is not safe for concurrent use. Read and modify one from a single
> goroutine, or guard it with your own lock. Independent packages can be processed in parallel.

## Usage

### Modifying content

```go
package main

import (
	"log"

	"github.com/dimelords/idmllib/v3/pkg/idml"
	"github.com/dimelords/idmllib/v3/pkg/story"
)

func main() {
	pkg, err := idml.Read("document.idml")
	if err != nil {
		log.Fatal(err)
	}

	newStory := &story.Story{Self: "u1234"}
	newStory.ParagraphStyleRanges = []story.ParagraphStyleRange{{
		AppliedParagraphStyle: "ParagraphStyle/$ID/NormalParagraphStyle",
		CharacterStyleRanges: []story.CharacterStyleRange{
			story.NewCharacterStyleRange(
				"CharacterStyle/$ID/[No character style]",
				[]story.Content{{Text: "Hello, World!"}},
			),
		},
	}}

	if err := pkg.AddStory("Stories/Story_u1234.xml", newStory, idml.DefaultValidationOptions()); err != nil {
		log.Fatal(err)
	}

	if err := idml.Write(pkg, "output.idml"); err != nil {
		log.Fatal(err)
	}
}
```

`AddStory` also registers the story in `designmap.xml`. A story present in the archive but absent
from the manifest is ignored by InDesign, which is a hard failure to debug from the outside.

### Cleaning up unused resources

```go
package main

import (
	"log"

	"github.com/dimelords/idmllib/v3/pkg/idml"
)

func main() {
	pkg, err := idml.Read("document.idml")
	if err != nil {
		log.Fatal(err)
	}

	rm := idml.NewResourceManager(pkg)

	orphans, err := rm.FindOrphans()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("unused paragraph styles: %d", len(orphans.ParagraphStyles))
	log.Printf("unused colors: %d", len(orphans.Colors))

	result, err := rm.CleanupOrphans(idml.CleanupOptions{
		RemoveOrphanedFonts:           true,
		RemoveOrphanedParagraphStyles: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("removed %d fonts, %d paragraph styles",
		len(result.RemovedFonts), len(result.RemovedParagraphStyles))

	if err := idml.Write(pkg, "cleaned.idml"); err != nil {
		log.Fatal(err)
	}
}
```

`CleanupOptions` is opt-in per resource kind, and `DryRun` reports what would go without touching
the document. Object styles are left alone by default, since they are often kept as templates.

### Exporting an IDMS snippet

```go
package main

import (
	"log"

	"github.com/dimelords/idmllib/v3/pkg/idml"
	"github.com/dimelords/idmllib/v3/pkg/idms"
	"github.com/dimelords/idmllib/v3/pkg/spread"
)

func main() {
	pkg, err := idml.Read("document.idml")
	if err != nil {
		log.Fatal(err)
	}

	frame, err := idml.PageItemOfType[spread.TextFrame](pkg, "u1e6")
	if err != nil {
		log.Fatal(err)
	}

	sel := idml.NewSelection()
	sel.AddTextFrame(frame)

	snippet, err := idms.NewExporter(pkg).ExportSelection(sel)
	if err != nil {
		log.Fatal(err)
	}

	if err := idms.Write(snippet, "snippet.idms"); err != nil {
		log.Fatal(err)
	}
}
```

The exporter walks the selection's dependencies — stories, styles and the styles they are based on,
fonts, colors, swatches, graphics — and embeds just those in the snippet. IDMS inlines its
resources rather than referencing separate files, so the snippet is self-contained.

### Large documents

`ReadOptions.Lazy` reads archive entries on demand and copies untouched ones straight through on
write, which turns a structure-only read of a 114 MB document into roughly 0.8 MB of memory. Call
`Package.Close` when finished. `ReadOptions.StreamingMode` roughly halves the memory held for
documents with embedded images; see [docs/STREAMING_MODE_DESIGN.md](docs/STREAMING_MODE_DESIGN.md).

## Packages

| Package | Contents |
| --- | --- |
| `pkg/idml` | The coordinator and public API: `Read`, `Write`, `Package`, selections, resource management |
| `pkg/document` | `designmap.xml` — the document manifest, layers, sections, hyperlinks |
| `pkg/spread` | `Spreads/*.xml` and `MasterSpreads/*.xml` — pages and page items |
| `pkg/story` | `Stories/*.xml` — text content and its formatting ranges |
| `pkg/resources` | `Resources/*.xml` — styles, fonts, graphics, preferences |
| `pkg/common` | Types shared across domains |
| `pkg/analysis` | Dependency tracking |
| `pkg/idms` | IDMS snippet export |
| `pkg/fit` | Text fitting: does this text fit this frame at this style |

The packages mirror the IDML file structure, which keeps the dependency graph acyclic and makes it
obvious where a given piece of the format lives. See [ARCHITECTURE.md](ARCHITECTURE.md).

## CLI

An interactive terminal tool for exploring and exporting from IDML files:

```bash
(cd cmd && go build -o ../bin/idmllib ./cli)
./bin/idmllib
```

It browses the document structure, inspects stories, spreads and resources, exports IDMS snippets
and reports dependencies.

## Development

```bash
go test ./...
go test -race -coverprofile=coverage.out ./...
```

The fidelity tests are the ones that matter most: parsing and re-marshalling every fixture must
produce structurally identical XML.

```bash
go test ./pkg/idml ./pkg/idms -run Fidelity
```

Golden files are refreshed with the `-update` flag, and only deliberately — run the fidelity tests
first.

```bash
go test ./pkg/idml ./pkg/idms -update
```

Linting uses [golangci-lint](https://golangci-lint.run/) v2 with the configuration in
`.golangci.yml`, which enables `govet`, `staticcheck`, `errcheck`, `ineffassign`, `unused`,
`misspell`, `gosec`, `forcetypeassert`, `nilerr` and `modernize`. Both modules are linted:

```bash
golangci-lint run
(cd cmd && golangci-lint run --config=../.golangci.yml)
```

CI runs linting, tests on Go 1.26 and 1.27 with the race detector, a short fuzzing pass over the
parsers, `govulncheck` and `gosec`, and builds the CLI.

Before changing any parser, read [docs/FIDELITY.md](docs/FIDELITY.md). Every element struct carries
an `OtherAttrs` catch-all, and types with several kinds of child record and replay child order; a
new struct that skips either will fail the fidelity tests, and would silently corrupt documents if
it did not.

## Dependencies

The library itself pulls in nothing outside the standard library:

```bash
go list -deps ./pkg/... | grep -v '^github.com/dimelords/idmllib'   # stdlib only
```

`etree`, `go-cmp`, `gopter` and `goldie` appear in `go.mod` because the tests use them. The CLI is
a separate module and has its own dependencies, so it costs a library consumer nothing.

## Documentation

- [Package reference](https://pkg.go.dev/github.com/dimelords/idmllib/v3)
- [ARCHITECTURE.md](ARCHITECTURE.md) — package layout and design decisions
- [docs/FIDELITY.md](docs/FIDELITY.md) — what lossless means here, and how it is enforced
- [docs/STREAMING_MODE_DESIGN.md](docs/STREAMING_MODE_DESIGN.md) — streaming and lazy reading
- [CHANGELOG.md](CHANGELOG.md) — release history
- [Adobe IDML specification](https://www.adobe.com/devnet/indesign/sdk.html)

## Maintenance

idmllib is written and maintained by its author for use in production typesetting systems. Bug
reports, questions and feature requests are welcome as
[issues](https://github.com/dimelords/idmllib/issues) — a reduced IDML file that reproduces the
problem is worth a great deal, since most bugs in this domain are specific to how one document was
built.

The API is stable within v3; anything breaking would arrive as a new major version.

## License

MIT — see [LICENSE](LICENSE).

## Author

Fredrik Gustafsson ([@dimelords](https://github.com/dimelords))
