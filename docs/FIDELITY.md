# Roundtrip fidelity

The core promise of this library is that parsing an IDML file and writing it
back does not change its meaning. This document describes how that is
achieved, what the guarantees are, and how to keep them when adding types.

## The problem

`encoding/xml` maps elements to Go structs. Used naively it loses data in
three ways, all of which InDesign accepts silently:

1. **Unknown attributes are dropped.** A struct that declares 14 of a
   CharacterStyle's 36 attributes writes back only those 14. In practice this
   stripped capitalization, language and kerning from styles.
2. **Sibling order is lost.** Children stored in one slice per element kind
   are written grouped by field. On a spread, child order is stacking order;
   in a story it is text order.
3. **Reading was destructive.** `Write` re-marshaled every cached object, so
   calling `Styles()` to inspect a document was enough to strip the file.

## How it is solved

### Attribute catch-all

Every element struct has an `OtherAttrs []xml.Attr` field tagged
`xml:",any,attr"`. `encoding/xml` fills it with attributes that have no typed
field and writes them back. Types with custom marshalers use
`common.UnmarshalAttrs` / `common.MarshalAttrs`, which follow the same tags
reflectively.

Limitation: a typed attribute declared `omitempty` cannot distinguish an empty
value (`OverrideList=""`) from an absent one. Such attributes are written back
absent. InDesign treats both the same for all attributes seen so far.

### Ordered children

Types with more than one kind of child record the document order of their
children while unmarshaling in a `childOrder common.ChildOrder` field and
replay it when marshaling (`common.EncodeChildren`). Children appended after
parsing are written after the recorded ones, grouped by kind; removed
children are skipped. For a spread this means newly added items land on top
of the stack.

Two implementations exist:

- **Generic:** `common.UnmarshalOrdered` / `common.MarshalOrdered` walk the
  struct's `xml` tags reflectively. Most types use them through two-line
  delegating methods in each package's `zz_ordered.go`.
- **Hand-written:** `spread.Spread`, `story.Story`, `ParagraphStyleRange`,
  `Document`, `StylesFile`, `Properties` and the nestable style groups, where
  element naming or namespaces need special handling.

To add a type: give it a `childOrder` field and two delegating methods in
`zz_ordered.go`, following the existing entries. Supported child field shapes
are `*T`, `[]T`, `T`, `string`, the `xml:",any"` raw catch-all and a
`xml:",chardata"` string.

### Modification tracking

`Package` tracks which files were changed through the API. `Write`
re-marshals only those; every other file is written back byte for byte.
The `Add*`, `Update*`, `Remove*` and `Set*` methods mark their targets. Code
that edits an object obtained from a getter must call
`Package.MarkModified(path)` for that file, otherwise the edit is not written.

## Tests

- `pkg/idml/fidelity_test.go` parses every typed file of every IDML in
  `testdata`, forces a re-marshal, and compares each XML file with the
  original using order-sensitive structural comparison
  (`xmlutil.StrictCompareOptions`). It also checks that reading without
  modifying leaves every file byte-identical, and that master spreads
  roundtrip.
- `pkg/idms/fidelity_test.go` does the same for every IDMS snippet.
- `pkg/common/attrs_test.go` covers the helpers.

When a fidelity test fails, the parser is wrong, not the fixture. Do not
"fix" it by relaxing the comparison.

## Verifying in InDesign

The XML comparison proves structural identity; InDesign proves meaning. A
useful manual check is to open the original and the roundtripped file with
ExtendScript and compare properties, for example:

```javascript
var cs = doc.characterStyles.itemByName("Some style");
cs.capitalization; cs.appliedLanguage.name;
```

Run it with `osascript -e 'tell application "Adobe InDesign 2026" to do script (POSIX file "check.jsx") language javascript'`.

## ItemTransform: absent is not identity

Every page item InDesign writes carries an `ItemTransform` attribute. Across
the test corpus that is 157 of 157 items, including `<Page>`, with no
exceptions. The Go fields are `omitempty`, so an item constructed in code and
never given one writes no attribute at all.

That matters in two opposite directions, measured against a real InDesign
open and a real Scribus import:

- **Scribus's IDML importer silently skips any page item without
  ItemTransform.** The item is neither drawn nor returned by the scripter's
  `getAllObjects`, and nothing errors. A generated document imports as a
  blank page.
- **InDesign treats an absent transform differently from an identity one.**
  Adding `ItemTransform="1 0 0 1 0 0"` to frames whose PathGeometry is in
  absolute page coordinates moved them off the page: a frame requested at
  (40, 120) to (540, 320) was placed at (635, 541) to (1135, 741).

So the library does not fill this in. There is no default that is correct for
both readers, and the right value depends on the coordinate convention of
whoever built the geometry. A generator that wants its output to survive
Scribus must emit the transform that matches its own convention.
