# Architecture plan

Decisions and sequencing for reshaping the public API, recorded before the
work so each step can be judged against the intent. Dates are when the step
landed or was decided.

## Principles

- **Fidelity first.** Every change must keep `TestFidelity*` green. Those
  tests, not the golden files, define correctness (see FIDELITY.md).
- **Structs stay the data model; methods carry the guarantees.** Attributes
  remain strings so values roundtrip verbatim. Typed helpers sit on top.
- **One representation per file.** A package file has exactly one parsed
  form. Raw bytes are always available through `Package.FileData`.
- **Small core.** `pkg/idml` holds Read, Write, the file getters and the
  modification API. Everything that can be written against that public
  surface lives in its own package.

## Steps

### 1. Lossless parsing and modification tracking (done 2026-09-16)

Attribute catch-alls, ordered children, dirty tracking, master spreads,
typed story content, typed groups and buttons, fidelity tests, InDesign
generated fixture. Details in FIDELITY.md.

### 2. One representation per resource file (done 2026-09-16)

- Add `resources.PreferencesFile` and `Package.Preferences()`, so
  Resources/Preferences.xml has a typed form like Fonts, Graphic and Styles.
- Add `Package.FileData(path)` returning a copy of a file's raw bytes, for
  callers that need the source (the style hierarchy parser, diagnostics).
- Remove the generic `ResourceFile`, `Package.Resource`, `Package.Resources`,
  `ParseResourceFile` and `MarshalResourceFile`. They duplicated the typed
  cache and a change through one was invisible to the other.
- Register new stories in designmap.xml. `AddStory` wrote the story file but
  never added the `idPkg:Story` reference or the `StoryList` entry, so
  InDesign never saw the story. Verified in InDesign 2026 on 2026-09-16: with
  the reference the added text appears; without it InDesign creates an empty
  story for the referencing frame and the text is silently lost.
  `RemoveStory` now removes both.

### 3. Module split (done 2026-09-16)

`cmd/` becomes its own Go module (`github.com/dimelords/idmllib/v3/cmd`) with
a `replace` to the library. Library consumers no longer inherit Bubbletea
and Lipgloss in their dependency graph. CI lints and builds both modules.

### 4. Slim `pkg/idml` (deferred)

The plan was to move the resource manager (`resourcemgr_*.go`, about 2 900
lines) to `pkg/resourcemgr`. Measured on 2026-09-16 it is not a clean cut:
`AddStory` and `UpdateStory` validate through the manager's unexported
dependency analysis and auto-add methods, so a separate package would import
`pkg/idml` and be imported by it. The manager also duplicates the dependency
analysis in `pkg/analysis`. The right move is to make `pkg/analysis` the one
analyzer, have the modification API depend on it, and then extract the
manager. That is its own piece of work and is not started here. Selection
stays in `pkg/idml` regardless, since it depends on the internal item index.

### 5. Wrapper naming (done 2026-09-16, breaking)

Today three names express the same idea in three ways: `Spread` (file
wrapper) with `SpreadElement` (the element), `Story` with `StoryElement`, and (before 2026-09-16)
`DocumentWithMetadata` with `Document`. The intended shape is: the element
type carries the natural name (`spread.Spread`, `story.Story`,
`document.Document`) and the file wrapper is `spread.File`, `story.File`,
`document.File`. `Package.Spread`, `Story` and `Document` return the element
callers work with; `SpreadFile`, `StoryFile` and `DocumentFile` return the
wrapper. This is a major-version change.

### 6. Streaming mode (done 2026-09-16)

`ReadOptions.StreamingMode` keeps embedded image payloads only in the raw file
bytes instead of also materializing them into parsed spreads. Held heap on the
large fixture drops from 277 MB to 117 MB. See STREAMING_MODE_DESIGN.md.

### 7. Lazy reading (done 2026-09-16)

`ReadOptions.Lazy` reads each archive entry on first use and copies untouched
entries in compressed form on write. A structure-only read of the large fixture
drops from 114 MB to 0.8 MB held; read-then-write goes from 2.90 s to 0.19 s.

### Later

- Typed accessors over the string attributes (bounds, booleans, style refs).
- Generate the typed attribute lists from InDesign's scripting DOM.
- Automated InDesign-in-the-loop test, opt-in via an environment variable.
- Model the remaining raw page items (EPS, WMF, PICT, ImportedPage, Sound,
  Movie, TextPath, MultiStateObject), document-level hyperlinks, bookmarks
  and conditions, and an API to add spreads, pages, layers and styles.

### 8. Public API cleanup (done 2026-09-17)

`Package` had grown to 62 methods by accretion. It is now 46, without losing
any capability:

- **Page item lookup, 11 methods to 2.** The six per-type selectors
  (`SelectTextFrameByID` and friends) plus `SelectPageItemByID`,
  `SelectPageItemsByIDs` and the two spread-wide helpers collapse into
  `PageItemByID`, which returns the `PageItem` interface, and the generic
  `PageItemOfType[T](pkg, id)` for a concrete type. The spread-wide helpers
  were already redundant: `Spread()` returns the element, so its page items
  are plain fields. Only tests used the removed methods.
- **Font lookups moved to where the data is.** `GetFontByStyle`,
  `GetFontPostScriptName`, `ListFontFamilies` and `ListFontStyles` answer
  questions about Fonts.xml, so they are now `resources.FontsFile` methods:
  `Font`, `PostScriptName`, `Families`, `Styles`, plus `Family`. They return
  `(value, bool)` rather than `(value, error)`, since "not present" is not an
  error condition.
- **Counters removed.** `ItemCount`, `TextFrameCount` and `RectangleCount`
  reported the size of an internal index and returned 0 when it had not been
  built, which made them misleading. The same numbers come from the parsed
  spreads.
- **Unused interfaces removed.** `PackageReader`, `PackageWriter` and
  `PackageAccessor` had no consumers in the library, the CLI or any test
  other than the ones asserting they existed. Go consumers define the
  interfaces they need. `PageItem` stays, since it is a real return type.
- **Internal plumbing unexported:** `ParseMetadataFile`,
  `MarshalMetadataFile` and `NewMissingResourcesError`.
- **`spread.SpreadTextFrame` is now `spread.TextFrame`**, matching Rectangle,
  Oval, Polygon, GraphicLine and Group.

One behavior change: `SelectByIDs` reported unknown ids by silently skipping
them, so an export could quietly omit what was asked for. It now returns an
error naming the offending id.
