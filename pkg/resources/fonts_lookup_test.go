package resources_test

import (
	"slices"
	"testing"

	"github.com/dimelords/idmllib/v3/pkg/idml"
	"github.com/dimelords/idmllib/v3/pkg/resources"
)

// testFonts returns the Fonts.xml of the example document, which is the only
// fixture with a rich set of families.
func testFonts(t *testing.T) *resources.FontsFile {
	t.Helper()
	pkg, err := idml.Read("../../testdata/example.idml")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	fonts, err := pkg.Fonts()
	if err != nil {
		t.Fatalf("Fonts(): %v", err)
	}
	return fonts
}

// firstFamilyWithFonts returns a family name that actually has fonts in it,
// so the tests do not depend on a particular document's font list.
func firstFamilyWithFonts(t *testing.T, fonts *resources.FontsFile) (string, string) {
	t.Helper()
	for _, name := range fonts.Families() {
		if styles := fonts.Styles(name); len(styles) > 0 {
			return name, styles[0]
		}
	}
	t.Fatal("no font family with fonts in the fixture")
	return "", ""
}

func TestFontsFileFamilies(t *testing.T) {
	fonts := testFonts(t)
	families := fonts.Families()
	if len(families) == 0 {
		t.Fatal("Families() returned none")
	}
	if len(families) != len(fonts.FontFamilies) {
		t.Errorf("Families() returned %d names for %d families", len(families), len(fonts.FontFamilies))
	}
	// Families are returned in document order.
	for i, name := range families {
		if name != fonts.FontFamilies[i].Name {
			t.Errorf("family %d = %q, want %q", i, name, fonts.FontFamilies[i].Name)
		}
	}
}

func TestFontsFileFamilyLookup(t *testing.T) {
	fonts := testFonts(t)
	want, _ := firstFamilyWithFonts(t, fonts)

	family, ok := fonts.Family(want)
	if !ok {
		t.Fatalf("Family(%q) not found", want)
	}
	if family.Name != want {
		t.Errorf("Family(%q).Name = %q", want, family.Name)
	}
	if _, ok := fonts.Family("No Such Family"); ok {
		t.Error("Family reported an unknown family as present")
	}
}

func TestFontsFileStyles(t *testing.T) {
	fonts := testFonts(t)
	family, style := firstFamilyWithFonts(t, fonts)

	styles := fonts.Styles(family)
	if !slices.Contains(styles, style) {
		t.Errorf("Styles(%q) = %v, want it to contain %q", family, styles, style)
	}
	if got := fonts.Styles("No Such Family"); got != nil {
		t.Errorf("Styles of an unknown family = %v, want nil", got)
	}
}

func TestFontsFileFont(t *testing.T) {
	fonts := testFonts(t)
	family, style := firstFamilyWithFonts(t, fonts)

	font, ok := fonts.Font(family, style)
	if !ok {
		t.Fatalf("Font(%q, %q) not found", family, style)
	}
	if font.FontStyleName != style {
		t.Errorf("Font(%q, %q).FontStyleName = %q", family, style, font.FontStyleName)
	}
	if _, ok := fonts.Font(family, "No Such Style"); ok {
		t.Error("Font reported an unknown style as present")
	}
	if _, ok := fonts.Font("No Such Family", style); ok {
		t.Error("Font reported a style of an unknown family as present")
	}
}

func TestFontsFilePostScriptName(t *testing.T) {
	fonts := testFonts(t)
	family, style := firstFamilyWithFonts(t, fonts)

	name, ok := fonts.PostScriptName(family, style)
	if !ok {
		t.Fatalf("PostScriptName(%q, %q) not found", family, style)
	}
	font, _ := fonts.Font(family, style)
	if name != font.PostScriptName {
		t.Errorf("PostScriptName = %q, want %q", name, font.PostScriptName)
	}
	if _, ok := fonts.PostScriptName(family, "No Such Style"); ok {
		t.Error("PostScriptName reported an unknown style as present")
	}
}
