package resources

import (
	"archive/zip"
	"bytes"
	"testing"
)

// TestGraphicFileRoundtrip tests parsing and marshaling of Graphic.xml
func TestGraphicFileRoundtrip(t *testing.T) {
	// Open test IDML file
	r, err := zip.OpenReader("../../testdata/plain.idml")
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer r.Close()

	// Find and read Graphic.xml
	var graphicData []byte
	for _, f := range r.File {
		if f.Name == "Resources/Graphic.xml" {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("Failed to open Graphic.xml: %v", err)
			}
			defer rc.Close()

			buf := new(bytes.Buffer)
			if _, err := buf.ReadFrom(rc); err != nil {
				t.Fatalf("Failed to read Graphic.xml: %v", err)
			}
			graphicData = buf.Bytes()
			break
		}
	}

	if len(graphicData) == 0 {
		t.Fatal("Graphic.xml not found in test file")
	}

	// Parse Graphic.xml
	graphic, err := ParseGraphicFile(graphicData)
	if err != nil {
		t.Fatalf("Failed to parse Graphic.xml: %v", err)
	}

	// Validate parsed data
	if graphic.DOMVersion == "" {
		t.Error("DOMVersion is empty")
	}

	if len(graphic.Colors) == 0 {
		t.Error("No colors parsed")
	}

	if len(graphic.Inks) == 0 {
		t.Error("No inks parsed")
	}

	// Marshal back to XML
	marshaledData, err := MarshalGraphicFile(graphic)
	if err != nil {
		t.Fatalf("Failed to marshal Graphic.xml: %v", err)
	}

	// Parse marshaled data to verify round-trip
	graphic2, err := ParseGraphicFile(marshaledData)
	if err != nil {
		t.Fatalf("Failed to parse marshaled Graphic.xml: %v", err)
	}

	// Compare key fields
	if graphic.DOMVersion != graphic2.DOMVersion {
		t.Errorf("DOMVersion mismatch: %s != %s", graphic.DOMVersion, graphic2.DOMVersion)
	}

	if len(graphic.Colors) != len(graphic2.Colors) {
		t.Errorf("Color count mismatch: %d != %d", len(graphic.Colors), len(graphic2.Colors))
	}

	if len(graphic.Inks) != len(graphic2.Inks) {
		t.Errorf("Ink count mismatch: %d != %d", len(graphic.Inks), len(graphic2.Inks))
	}

	t.Logf("Parsed %d colors, %d inks, %d gradients, %d swatches",
		len(graphic.Colors), len(graphic.Inks), len(graphic.Gradients), len(graphic.Swatches))
}

// TestFontsFileRoundtrip tests parsing and marshaling of Fonts.xml
func TestFontsFileRoundtrip(t *testing.T) {
	// Open test IDML file
	r, err := zip.OpenReader("../../testdata/plain.idml")
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer r.Close()

	// Find and read Fonts.xml
	var fontsData []byte
	for _, f := range r.File {
		if f.Name == "Resources/Fonts.xml" {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("Failed to open Fonts.xml: %v", err)
			}
			defer rc.Close()

			buf := new(bytes.Buffer)
			if _, err := buf.ReadFrom(rc); err != nil {
				t.Fatalf("Failed to read Fonts.xml: %v", err)
			}
			fontsData = buf.Bytes()
			break
		}
	}

	if len(fontsData) == 0 {
		t.Fatal("Fonts.xml not found in test file")
	}

	// Parse Fonts.xml
	fonts, err := ParseFontsFile(fontsData)
	if err != nil {
		t.Fatalf("Failed to parse Fonts.xml: %v", err)
	}

	// Validate parsed data
	if fonts.DOMVersion == "" {
		t.Error("DOMVersion is empty")
	}

	if len(fonts.FontFamilies) == 0 {
		t.Error("No font families parsed")
	}

	// Check first font family
	if len(fonts.FontFamilies) > 0 {
		family := fonts.FontFamilies[0]
		if family.Name == "" {
			t.Error("Font family name is empty")
		}
		if len(family.Fonts) == 0 {
			t.Error("Font family has no fonts")
		}
	}

	// Marshal back to XML
	marshaledData, err := MarshalFontsFile(fonts)
	if err != nil {
		t.Fatalf("Failed to marshal Fonts.xml: %v", err)
	}

	// Parse marshaled data to verify round-trip
	fonts2, err := ParseFontsFile(marshaledData)
	if err != nil {
		t.Fatalf("Failed to parse marshaled Fonts.xml: %v", err)
	}

	// Compare key fields
	if fonts.DOMVersion != fonts2.DOMVersion {
		t.Errorf("DOMVersion mismatch: %s != %s", fonts.DOMVersion, fonts2.DOMVersion)
	}

	if len(fonts.FontFamilies) != len(fonts2.FontFamilies) {
		t.Errorf("Font family count mismatch: %d != %d", len(fonts.FontFamilies), len(fonts2.FontFamilies))
	}

	t.Logf("Parsed %d font families", len(fonts.FontFamilies))
	for _, family := range fonts.FontFamilies {
		t.Logf("  - %s (%d fonts)", family.Name, len(family.Fonts))
	}
}

// TestStylesFileRoundtrip tests parsing and marshaling of Styles.xml
func TestStylesFileRoundtrip(t *testing.T) {
	// Open test IDML file
	r, err := zip.OpenReader("../../testdata/plain.idml")
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer r.Close()

	// Find and read Styles.xml
	var stylesData []byte
	for _, f := range r.File {
		if f.Name == "Resources/Styles.xml" {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("Failed to open Styles.xml: %v", err)
			}
			defer rc.Close()

			buf := new(bytes.Buffer)
			if _, err := buf.ReadFrom(rc); err != nil {
				t.Fatalf("Failed to read Styles.xml: %v", err)
			}
			stylesData = buf.Bytes()
			break
		}
	}

	if len(stylesData) == 0 {
		t.Fatal("Styles.xml not found in test file")
	}

	// Parse Styles.xml
	styles, err := ParseStylesFile(stylesData)
	if err != nil {
		t.Fatalf("Failed to parse Styles.xml: %v", err)
	}

	// Validate parsed data
	if styles.DOMVersion == "" {
		t.Error("DOMVersion is empty")
	}

	if styles.RootParagraphStyleGroup == nil {
		t.Error("No paragraph style group parsed")
	}

	if styles.RootCharacterStyleGroup == nil {
		t.Error("No character style group parsed")
	}

	// Check paragraph styles
	if styles.RootParagraphStyleGroup != nil && len(styles.RootParagraphStyleGroup.ParagraphStyles) > 0 {
		pStyle := styles.RootParagraphStyleGroup.ParagraphStyles[0]
		if pStyle.Name == "" {
			t.Error("Paragraph style name is empty")
		}
	}

	// Marshal back to XML
	marshaledData, err := MarshalStylesFile(styles)
	if err != nil {
		t.Fatalf("Failed to marshal Styles.xml: %v", err)
	}

	// Parse marshaled data to verify round-trip
	styles2, err := ParseStylesFile(marshaledData)
	if err != nil {
		t.Fatalf("Failed to parse marshaled Styles.xml: %v", err)
	}

	// Compare key fields
	if styles.DOMVersion != styles2.DOMVersion {
		t.Errorf("DOMVersion mismatch: %s != %s", styles.DOMVersion, styles2.DOMVersion)
	}

	if styles.RootParagraphStyleGroup != nil && styles2.RootParagraphStyleGroup != nil {
		count1 := len(styles.RootParagraphStyleGroup.ParagraphStyles)
		count2 := len(styles2.RootParagraphStyleGroup.ParagraphStyles)
		if count1 != count2 {
			t.Errorf("Paragraph style count mismatch: %d != %d", count1, count2)
		}
		t.Logf("Parsed %d paragraph styles", count1)
	}

	if styles.RootCharacterStyleGroup != nil && styles2.RootCharacterStyleGroup != nil {
		count1 := len(styles.RootCharacterStyleGroup.CharacterStyles)
		count2 := len(styles2.RootCharacterStyleGroup.CharacterStyles)
		if count1 != count2 {
			t.Errorf("Character style count mismatch: %d != %d", count1, count2)
		}
		t.Logf("Parsed %d character styles", count1)
	}
}
