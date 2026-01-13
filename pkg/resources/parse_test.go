package resources

import (
	"errors"
	"strings"
	"testing"

	"github.com/dimelords/idmllib/v2/pkg/common"
)

// TestParseFontsFile_InvalidXML tests error handling for invalid XML.
func TestParseFontsFile_InvalidXML(t *testing.T) {
	tests := []struct {
		name string
		xml  string
	}{
		{
			name: "malformed XML",
			xml:  `<FontFamily><unclosed>`,
		},
		{
			name: "empty data",
			xml:  ``,
		},
		{
			name: "invalid structure",
			xml:  `<?xml version="1.0"?><InvalidRoot/>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseFontsFile([]byte(tt.xml))
			if err == nil {
				t.Error("ParseFontsFile() should return error for invalid XML")
			}

			// Verify it's our custom error type
			var commonErr *common.Error
			if !errors.As(err, &commonErr) {
				t.Errorf("ParseFontsFile() error type = %T, want *common.Error", err)
			}
		})
	}
}

// TestParseGraphicFile_InvalidXML tests error handling for invalid XML.
func TestParseGraphicFile_InvalidXML(t *testing.T) {
	tests := []struct {
		name string
		xml  string
	}{
		{
			name: "malformed XML",
			xml:  `<Color><unclosed>`,
		},
		{
			name: "empty data",
			xml:  ``,
		},
		{
			name: "invalid structure",
			xml:  `<?xml version="1.0"?><InvalidRoot/>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseGraphicFile([]byte(tt.xml))
			if err == nil {
				t.Error("ParseGraphicFile() should return error for invalid XML")
			}

			// Verify it's our custom error type
			var commonErr *common.Error
			if !errors.As(err, &commonErr) {
				t.Errorf("ParseGraphicFile() error type = %T, want *common.Error", err)
			}
		})
	}
}

// TestParseStylesFile_InvalidXML tests error handling for invalid XML.
func TestParseStylesFile_InvalidXML(t *testing.T) {
	tests := []struct {
		name string
		xml  string
	}{
		{
			name: "malformed XML",
			xml:  `<ParagraphStyle><unclosed>`,
		},
		{
			name: "empty data",
			xml:  ``,
		},
		{
			name: "invalid structure",
			xml:  `<?xml version="1.0"?><InvalidRoot/>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseStylesFile([]byte(tt.xml))
			if err == nil {
				t.Error("ParseStylesFile() should return error for invalid XML")
			}

			// Verify it's our custom error type
			var commonErr *common.Error
			if !errors.As(err, &commonErr) {
				t.Errorf("ParseStylesFile() error type = %T, want *common.Error", err)
			}
		})
	}
}

// TestMarshalFontsFile_BasicStructure tests marshaling a basic fonts structure.
func TestMarshalFontsFile_BasicStructure(t *testing.T) {
	fonts := &FontsFile{
		DOMVersion: "15.0",
		FontFamilies: []FontFamily{
			{
				Self: "FontFamily/Test",
				Name: "Test Font",
				Fonts: []Font{
					{
						Self:                "Font/Test%3aRegular",
						FontFamily:          "Test Font",
						Name:                "Test Font\tRegular",
						PostScriptName:      "TestFont-Regular",
						FontStyleName:       "Regular",
						FontType:            "OpenType",
						WritingScript:       "0",
						FullName:            "Test Font Regular",
						FullNameNative:      "Test Font Regular",
						FontStyleNameNative: "Regular",
						PlatformName:        "TestFont-Regular",
						Version:             "1.000",
					},
				},
			},
		},
	}

	data, err := MarshalFontsFile(fonts)
	if err != nil {
		t.Fatalf("MarshalFontsFile() error = %v", err)
	}

	// Verify XML declaration exists
	if !strings.HasPrefix(string(data), `<?xml version="1.0"`) {
		t.Error("MarshalFontsFile() should start with XML declaration")
	}

	// Verify we can parse it back
	parsed, err := ParseFontsFile(data)
	if err != nil {
		t.Fatalf("ParseFontsFile() error after marshaling = %v", err)
	}

	if parsed.DOMVersion != fonts.DOMVersion {
		t.Errorf("DOMVersion mismatch after roundtrip: %s != %s", parsed.DOMVersion, fonts.DOMVersion)
	}

	if len(parsed.FontFamilies) != len(fonts.FontFamilies) {
		t.Errorf("FontFamilies count mismatch: %d != %d", len(parsed.FontFamilies), len(fonts.FontFamilies))
	}
}

// TestMarshalGraphicFile_BasicStructure tests marshaling a basic graphic structure.
func TestMarshalGraphicFile_BasicStructure(t *testing.T) {
	graphic := &GraphicFile{
		DOMVersion: "15.0",
		Colors: []Color{
			{
				Self:       "Color/Test",
				Name:       "Test Color",
				Model:      "Process",
				Space:      "RGB",
				ColorValue: "255 0 0",
			},
		},
		Swatches: []Swatch{
			{
				Self: "Swatch/None",
				Name: "None",
			},
		},
	}

	data, err := MarshalGraphicFile(graphic)
	if err != nil {
		t.Fatalf("MarshalGraphicFile() error = %v", err)
	}

	// Verify XML declaration exists
	if !strings.HasPrefix(string(data), `<?xml version="1.0"`) {
		t.Error("MarshalGraphicFile() should start with XML declaration")
	}

	// Verify we can parse it back
	parsed, err := ParseGraphicFile(data)
	if err != nil {
		t.Fatalf("ParseGraphicFile() error after marshaling = %v", err)
	}

	if parsed.DOMVersion != graphic.DOMVersion {
		t.Errorf("DOMVersion mismatch after roundtrip: %s != %s", parsed.DOMVersion, graphic.DOMVersion)
	}

	if len(parsed.Colors) != len(graphic.Colors) {
		t.Errorf("Colors count mismatch: %d != %d", len(parsed.Colors), len(graphic.Colors))
	}
}

// TestMarshalStylesFile_BasicStructure tests marshaling a basic styles structure.
func TestMarshalStylesFile_BasicStructure(t *testing.T) {
	styles := &StylesFile{
		DOMVersion: "15.0",
		RootParagraphStyleGroup: &ParagraphStyleGroup{
			Self: "ParagraphStyleGroup/$ID/[Root]",
			ParagraphStyles: []ParagraphStyle{
				{
					Self: "ParagraphStyle/Test",
					Name: "Test Style",
				},
			},
		},
		RootCharacterStyleGroup: &CharacterStyleGroup{
			Self: "CharacterStyleGroup/$ID/[Root]",
			CharacterStyles: []CharacterStyle{
				{
					Self: "CharacterStyle/Test",
					Name: "Test Style",
				},
			},
		},
	}

	data, err := MarshalStylesFile(styles)
	if err != nil {
		t.Fatalf("MarshalStylesFile() error = %v", err)
	}

	// Verify XML declaration exists
	if !strings.HasPrefix(string(data), `<?xml version="1.0"`) {
		t.Error("MarshalStylesFile() should start with XML declaration")
	}

	// Verify we can parse it back
	parsed, err := ParseStylesFile(data)
	if err != nil {
		t.Fatalf("ParseStylesFile() error after marshaling = %v", err)
	}

	if parsed.DOMVersion != styles.DOMVersion {
		t.Errorf("DOMVersion mismatch after roundtrip: %s != %s", parsed.DOMVersion, styles.DOMVersion)
	}

	if parsed.RootParagraphStyleGroup == nil {
		t.Error("RootParagraphStyleGroup is nil after roundtrip")
	}

	if parsed.RootCharacterStyleGroup == nil {
		t.Error("RootCharacterStyleGroup is nil after roundtrip")
	}
}
