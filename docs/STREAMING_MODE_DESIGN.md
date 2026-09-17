# Streaming mode

**Status:** implemented 2026-09-16. This document describes what was built and
the measurements behind it. It replaces the earlier proposal, which was drafted
before modification tracking existed and assumed a more complex design.

## The problem

A spread that contains embedded images is dominated by their payloads. In the
test corpus, `testdata/example_big.idml` has a single 121 MB spread that is
**99.9% base64 image data**, held in five `<Contents>` elements:

```xml
<Image Self="u26a" ...><Contents><![CDATA[/9j/2wBD...]]></Contents></Image>
```

Parsing that spread materializes every payload a second time inside the parsed
document, because `<Contents>` is not modeled and lands in the raw-element
catch-all. Measured on the fixture:

| Stage | Held heap |
| --- | --- |
| After `Read` | 117 MB |
| After `Spreads()` | 277 MB |

So holding a parsed package costs roughly twice the file size, which is the
difference between fitting and not fitting in a small Lambda.

## What streaming mode does

`ReadOptions.StreamingMode` keeps each payload exactly once, in the raw file
bytes the package already holds, and gives the parsed spread an empty
`<Contents>` instead.

- **Read** is unchanged: the raw bytes are stored either way.
- **Parse** strips the payloads. The stripped copy is tiny (121 MB becomes
  about 100 KB) and each payload is recorded as a **sub-slice of the raw
  bytes**, so it costs no additional memory.
- **Write** restores them:
  - a spread that was not modified is written back from its raw bytes, so the
    payloads are preserved exactly and no work happens at all;
  - a modified spread is re-marshaled and each empty `<Contents>` is filled in
    from the payload of the page item that owns it.

Payloads are matched by the `Self` attribute of the owning page item, not by
position, so reordering, adding or removing images is safe. A `<Contents>` that
already holds data is never overwritten, so a payload the caller set explicitly
wins.

## Measurements

Held heap after `Read` plus `Spreads()` on `example_big.idml`:

| Mode | Held heap | Allocated per op | Time per op |
| --- | --- | --- | --- |
| Normal | 277 MB | 610 MB | 1.98 s |
| Streaming | 117 MB | 275 MB | 2.39 s |

Streaming holds **58% less** memory and allocates **55% less**, at about **21%
more CPU** for the byte scan over the payloads. Reading without parsing costs
the same in both modes, which `BenchmarkReadOnly_Big` documents.

## Trade-offs

- `Contents` is empty in the parsed document. Use `Package.ImageContents` to
  read a payload, or read without streaming mode.
- The saving applies to embedded images only. A document with linked images or
  no images sees no benefit and a small scanning cost.
- The mode is opt-in and changes no behavior that is observable in output.

## Correctness

The byte-level strip and restore are covered by unit tests in
`internal/xmlutil/contents_test.go`, including owner matching under reordering
and the refusal to overwrite explicit content.

`pkg/idml/streaming_test.go` requires that streaming and normal mode produce
**identical archives**, both for an untouched package and after every spread is
marked modified, across the whole fixture corpus. It also asserts the memory
saving so the feature cannot silently stop working.

Verified in InDesign 2026 on 2026-09-16: after a streamed read, an edit and a
write, the document still reports 5 embedded images at identical dimensions,
all five payloads are byte-for-byte identical to the source, and the edit is
present.

## Lazy reading

`ReadOptions.Lazy` goes further: an archive entry is read the first time it is
actually used, and one that is never used is never decompressed. On write, an
entry that was neither loaded nor modified is copied straight across in its
compressed form with `zip.Writer.Copy`, so it is not decompressed and not
recompressed.

Limits are enforced exactly as before, because they are checked against the
sizes declared in the archive directory, which lazy reading still validates for
every entry up front.

Measured on `example_big.idml` (92 MB on disk, 122 MB uncompressed):

| Workload | Eager | Lazy |
| --- | --- | --- |
| Open and read the document structure (held heap) | 114 MB | 0.8 MB |
| Read then write back unchanged (time) | 2.90 s | 0.19 s |
| Read then write back unchanged (allocated) | 125 MB | 4.2 MB |

The package holds the source archive open, so the file must stay readable and
unchanged, and `Package.Close` releases it. Close is safe on any package, and
`Package.LoadAll` materializes everything if you want to drop the source early.
Entries loaded before Close stay usable; reading one that was not gives a clear
error rather than empty data.

`pkg/idml/lazy_test.go` requires that lazy and eager reads write identical
archive contents across the whole corpus, untouched and after modification.
