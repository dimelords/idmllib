package idml

import (
	"archive/zip"
	"encoding/xml"
	"io"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"
)

// fixtures are the real InDesign exports the font checks are measured
// against. example_big.idml is left out: it is 92MB and adds no font
// reference shape the others lack.
var fixtures = []string{
	"../../testdata/plain.idml",
	"../../testdata/example.idml",
	"../../testdata/features.idml",
	"../../testdata/tripple.idml",
}

// collectedFonts runs the dependency analysis and returns the font
// families it found, sorted.
func collectedFonts(t *testing.T, path string) []string {
	t.Helper()

	pkg, err := Read(path)
	if err != nil {
		t.Fatalf("Read(%s): %v", path, err)
	}
	deps, err := NewResourceManager(pkg).analyzeDependencies()
	if err != nil {
		t.Fatalf("analyzeDependencies(%s): %v", path, err)
	}

	got := make([]string, 0, len(deps.fonts))
	for family := range deps.fonts {
		got = append(got, family)
	}
	sort.Strings(got)
	return got
}

// referencedFontsInArchive is an independent ground truth: it walks the
// raw XML of every file in the archive and reports each font family
// named by a font-bearing element or attribute.
//
// Resources/Fonts.xml is skipped because it declares families rather
// than referencing them, and META-INF/metadata.xml because its XMP
// <stFnt:*> entries describe the file instead of pointing at a font the
// document must ship.
//
// The names below are spelled out rather than taken from fontElements
// and fontAttributes on purpose. Sharing those maps with the code under
// test would make this agree with the implementation by construction:
// dropping an entry would change both sides at once and the comparison
// would still pass. Written out, the test fails when they diverge.
func referencedFontsInArchive(t *testing.T, path string) []string {
	wantElements := map[string]bool{"AppliedFont": true, "BulletsFont": true}
	wantAttributes := map[string]bool{"WatermarkFontFamily": true}

	t.Helper()

	zr, err := zip.OpenReader(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer func() { _ = zr.Close() }()

	found := map[string]bool{}
	for _, f := range zr.File {
		if !strings.HasSuffix(f.Name, ".xml") ||
			f.Name == "Resources/Fonts.xml" || f.Name == "META-INF/metadata.xml" {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			t.Fatalf("open %s in %s: %v", f.Name, path, err)
		}
		data, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			t.Fatalf("read %s in %s: %v", f.Name, path, err)
		}

		dec := xml.NewDecoder(strings.NewReader(string(data)))
		for {
			tok, err := dec.Token()
			if err != nil {
				break
			}
			start, ok := tok.(xml.StartElement)
			if !ok {
				continue
			}
			for _, attr := range start.Attr {
				if wantAttributes[attr.Name.Local] {
					addFamily(found, attr.Value)
				}
			}
			if wantElements[start.Name.Local] {
				var text string
				if err := dec.DecodeElement(&text, &start); err == nil {
					addFamily(found, text)
				}
			}
		}
	}

	out := make([]string, 0, len(found))
	for family := range found {
		out = append(out, family)
	}
	sort.Strings(out)
	return out
}

func addFamily(set map[string]bool, family string) {
	family = strings.TrimSpace(family)
	if family != "" && !strings.HasPrefix(family, "$ID/") {
		set[family] = true
	}
}

// TestExtractFonts_MatchesArchive is the guard against the typed model
// drifting away from the file format. The dependency analysis reaches
// fonts through Go structs, so a family held in a field nobody walks is
// invisible to it; scanning the archive's XML cannot miss one that way.
// The two must agree on every fixture.
//
// This matters in both directions. A reference the analysis misses makes
// a font in use look orphaned, and FindOrphans would then offer to
// delete it. One it invents makes ValidateReferences cry wolf.
func TestExtractFonts_MatchesArchive(t *testing.T) {
	for _, path := range fixtures {
		t.Run(filepath.Base(path), func(t *testing.T) {
			got := collectedFonts(t, path)
			want := referencedFontsInArchive(t, path)

			if slices.Equal(got, want) {
				return
			}
			for _, family := range want {
				if !slices.Contains(got, family) {
					t.Errorf("font %q is referenced in the archive but the analysis missed it "+
						"(FindOrphans would offer to delete a font in use)", family)
				}
			}
			for _, family := range got {
				if !slices.Contains(want, family) {
					t.Errorf("analysis reported font %q, which nothing in the archive references", family)
				}
			}
		})
	}
}

// TestValidateReferences_NoFalseFontPositives checks that real InDesign
// exports come back clean. Every family these files name is declared in
// their own Fonts.xml, so any font finding here is a bug in the check.
func TestValidateReferences_NoFalseFontPositives(t *testing.T) {
	for _, path := range fixtures {
		t.Run(filepath.Base(path), func(t *testing.T) {
			pkg, err := Read(path)
			if err != nil {
				t.Fatalf("Read: %v", err)
			}
			errs, err := NewResourceManager(pkg).ValidateReferences()
			if err != nil {
				t.Fatalf("ValidateReferences: %v", err)
			}
			for _, e := range errs {
				if e.ResourceType == "Font" {
					t.Errorf("false positive: %s", e.Message)
				}
			}
		})
	}
}

// TestValidateReferences_ReportsUndeclaredFont pins the failure this
// check exists for: a style names a font that Resources/Fonts.xml never
// declares. InDesign opens such a file and renders it correctly, because
// it resolves the family from the system rather than the inventory, so
// nothing upstream catches it - but Scribus's IDML importer needs the
// declaration and silently substitutes its own default face without it.
// That is how a generator shipped previews in the wrong font
// (see docs/FIDELITY.md).
func TestValidateReferences_ReportsUndeclaredFont(t *testing.T) {
	pkg, err := Read("../../testdata/plain.idml")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	fonts, err := pkg.Fonts()
	if err != nil {
		t.Fatalf("Fonts: %v", err)
	}
	// Drop every declaration while leaving the styles that name them
	// untouched - exactly the shape of an empty Fonts.xml.
	fonts.FontFamilies = nil
	pkg.SetFonts(fonts)

	errs, err := NewResourceManager(pkg).ValidateReferences()
	if err != nil {
		t.Fatalf("ValidateReferences: %v", err)
	}

	reported := map[string]bool{}
	for _, e := range errs {
		if e.ResourceType == "Font" {
			reported[e.ResourceID] = true
		}
	}

	// Everything the document still references is now undeclared, so
	// every one of those families must be reported.
	referenced := referencedFontsInArchive(t, "../../testdata/plain.idml")
	if len(referenced) == 0 {
		t.Fatal("fixture references no fonts, nothing to detect")
	}
	for _, family := range referenced {
		if !reported[family] {
			t.Errorf("font %q is referenced but undeclared, and was not reported", family)
		}
	}
}

// TestFindOrphans_KeepsFontsInUse guards the destructive direction: a
// font the document still renders with must never be offered up for
// removal. Before the analysis collected fonts at all, the used-font set
// was always empty and FindOrphans reported every declared family as
// orphaned - advice that would have stripped the document bare.
func TestFindOrphans_KeepsFontsInUse(t *testing.T) {
	for _, path := range fixtures {
		t.Run(filepath.Base(path), func(t *testing.T) {
			orphans, err := NewResourceManagerFor(t, path).FindOrphans()
			if err != nil {
				t.Fatalf("FindOrphans: %v", err)
			}
			inUse := referencedFontsInArchive(t, path)
			for _, family := range orphans.Fonts {
				if slices.Contains(inUse, family) {
					t.Errorf("FindOrphans reported %q as unused, but the document references it", family)
				}
			}
			if len(orphans.Fonts) == len(inUse) && len(inUse) > 0 {
				t.Errorf("every font reported orphaned (%d), which is the old always-empty-deps bug",
					len(orphans.Fonts))
			}
		})
	}
}

// NewResourceManagerFor opens a fixture and wraps it, failing the test
// on error.
func NewResourceManagerFor(t *testing.T, path string) *ResourceManager {
	t.Helper()
	pkg, err := Read(path)
	if err != nil {
		t.Fatalf("Read(%s): %v", path, err)
	}
	return NewResourceManager(pkg)
}
