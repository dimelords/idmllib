package idml_test

import (
	"testing"

	"github.com/dimelords/idmllib/pkg/idml"
)

func TestPackage_GetFontPostScriptName(t *testing.T) {
	// Use a test document with known fonts
	pkg, err := idml.Read("../../testdata/example.idml")
	if err != nil {
		t.Skipf("Test file not available: %v", err)
		return
	}

	tests := []struct {
		name       string
		family     string
		style      string
		wantPS     string
		wantErr    bool
		skipReason string
	}{
		{
			name:   "existing font",
			family: "Minion Pro",
			style:  "Regular",
			wantPS: "MinionPro-Regular",
		},
		{
			name:    "non-existent family",
			family:  "NonExistentFont",
			style:   "Regular",
			wantErr: true,
		},
		{
			name:    "non-existent style",
			family:  "Minion Pro",
			style:   "NonExistentStyle",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipReason != "" {
				t.Skip(tt.skipReason)
			}

			psName, err := pkg.GetFontPostScriptName(tt.family, tt.style)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.wantPS != "" && psName != tt.wantPS {
				t.Errorf("PostScript name = %q, want %q", psName, tt.wantPS)
			}
		})
	}
}

func TestPackage_GetFontByStyle(t *testing.T) {
	pkg, err := idml.Read("../../testdata/example.idml")
	if err != nil {
		t.Skipf("Test file not available: %v", err)
		return
	}

	font, err := pkg.GetFontByStyle("Minion Pro", "Regular")
	if err != nil {
		t.Skipf("Font not found in test document: %v", err)
		return
	}

	if font.FontFamily != "Minion Pro" {
		t.Errorf("FontFamily = %q, want %q", font.FontFamily, "Minion Pro")
	}

	if font.FontStyleName != "Regular" {
		t.Errorf("FontStyleName = %q, want %q", font.FontStyleName, "Regular")
	}

	if font.PostScriptName == "" {
		t.Error("PostScriptName should not be empty")
	}
}

func TestPackage_ListFontFamilies(t *testing.T) {
	pkg, err := idml.Read("../../testdata/example.idml")
	if err != nil {
		t.Skipf("Test file not available: %v", err)
		return
	}

	families, err := pkg.ListFontFamilies()
	if err != nil {
		t.Fatalf("ListFontFamilies error: %v", err)
	}

	if len(families) == 0 {
		t.Error("Expected at least one font family")
	}
}

func TestPackage_ListFontStyles(t *testing.T) {
	pkg, err := idml.Read("../../testdata/example.idml")
	if err != nil {
		t.Skipf("Test file not available: %v", err)
		return
	}

	// First get available families
	families, err := pkg.ListFontFamilies()
	if err != nil || len(families) == 0 {
		t.Skip("No font families available in test document")
		return
	}

	// Test listing styles for the first family
	styles, err := pkg.ListFontStyles(families[0])
	if err != nil {
		t.Fatalf("ListFontStyles error: %v", err)
	}

	if len(styles) == 0 {
		t.Errorf("Expected at least one font style for family %q", families[0])
	}

	// Test non-existent family
	_, err = pkg.ListFontStyles("NonExistentFamily")
	if err == nil {
		t.Error("Expected error for non-existent family")
	}
}
