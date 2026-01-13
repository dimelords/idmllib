package idml

import (
	"fmt"
	"strings"
	"testing"
)

func TestNewFromTemplate_CreatesFromTemplate(t *testing.T) {
	tests := []struct {
		name    string
		opts    *TemplateOptions
		wantErr bool
	}{
		{
			name:    "default options (US Letter)",
			opts:    nil,
			wantErr: false,
		},
		{
			name: "A4 portrait",
			opts: &TemplateOptions{
				DOMVersion:  "20.4",
				Preset:      PresetA4,
				Orientation: "Portrait",
			},
			wantErr: false,
		},
		{
			name: "A4 landscape",
			opts: &TemplateOptions{
				Preset:      PresetA4,
				Orientation: "Landscape",
			},
			wantErr: false,
		},
		{
			name: "tabloid portrait",
			opts: &TemplateOptions{
				Preset:      PresetTabloid,
				Orientation: "Portrait",
			},
			wantErr: false,
		},
		{
			name: "us legal portrait",
			opts: &TemplateOptions{
				Preset:      PresetLegalUS,
				Orientation: "Portrait",
			},
			wantErr: false,
		},
		{
			name: "custom dimensions",
			opts: &TemplateOptions{
				Preset: PresetCustom,
				CustomDimensions: &PageDimensions{
					Width:  720,
					Height: 1080,
				},
			},
			wantErr: false,
		},
		{
			name: "with custom margins",
			opts: &TemplateOptions{
				Preset: PresetLetterUS,
				Margins: struct {
					Top    float64
					Bottom float64
					Left   float64
					Right  float64
				}{
					Top:    72,
					Bottom: 72,
					Left:   72,
					Right:  72,
				},
			},
			wantErr: false,
		},
		{
			name: "two columns",
			opts: &TemplateOptions{
				Preset:       PresetLetterUS,
				ColumnCount:  2,
				ColumnGutter: 24,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkg, err := NewFromTemplate(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFromTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			// Verify package was created
			if pkg == nil {
				t.Fatal("NewFromTemplate() returned nil package")
			}

			// Verify required files exist
			requiredFiles := []string{
				"designmap.xml",
				"MasterSpreads/MasterSpread_ub4.xml",
				"Resources/Preferences.xml",
			}

			for _, filename := range requiredFiles {
				if _, exists := pkg.files[filename]; !exists {
					t.Errorf("%s not found in package", filename)
				}
			}
		})
	}
}

func TestNewFromTemplate_DocumentParsing(t *testing.T) {
	tests := []struct {
		name       string
		opts       *TemplateOptions
		wantWidth  float64
		wantHeight float64
	}{
		{
			name:       "US Letter portrait",
			opts:       &TemplateOptions{Preset: PresetLetterUS, Orientation: "Portrait"},
			wantWidth:  612,
			wantHeight: 792,
		},
		{
			name:       "US Letter landscape",
			opts:       &TemplateOptions{Preset: PresetLetterUS, Orientation: "Landscape"},
			wantWidth:  792,
			wantHeight: 612,
		},
		{
			name:       "A4 portrait",
			opts:       &TemplateOptions{Preset: PresetA4, Orientation: "Portrait"},
			wantWidth:  595.276,
			wantHeight: 841.89,
		},
		{
			name:       "A4 landscape",
			opts:       &TemplateOptions{Preset: PresetA4, Orientation: "Landscape"},
			wantWidth:  841.89,
			wantHeight: 595.276,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkg, err := NewFromTemplate(tt.opts)
			if err != nil {
				t.Fatalf("NewFromTemplate() failed: %v", err)
			}

			// Verify designmap.xml contains correct dimensions
			entry := pkg.files["designmap.xml"]
			content := string(entry.data)

			// Check for width and height in XML
			if !strings.Contains(content, fmt.Sprintf("PageWidth=\"%.3f\"", tt.wantWidth)) {
				t.Errorf("designmap.xml missing PageWidth=%.3f", tt.wantWidth)
			}
			if !strings.Contains(content, fmt.Sprintf("PageHeight=\"%.3f\"", tt.wantHeight)) {
				t.Errorf("designmap.xml missing PageHeight=%.3f", tt.wantHeight)
			}
		})
	}
}

func TestNewFromTemplate_MasterSpread(t *testing.T) {
	pkg, err := NewFromTemplate(nil)
	if err != nil {
		t.Fatalf("NewFromTemplate() failed: %v", err)
	}

	// Verify MasterSpread file exists
	entry, exists := pkg.files["MasterSpreads/MasterSpread_ub4.xml"]
	if !exists {
		t.Fatal("MasterSpread_ub4.xml not found")
	}

	content := string(entry.data)

	// Verify it's valid XML with correct structure
	if !strings.Contains(content, "<?xml") {
		t.Error("MasterSpread missing XML declaration")
	}
	if !strings.Contains(content, "<idPkg:MasterSpread") {
		t.Error("MasterSpread missing root element")
	}
	if !strings.Contains(content, "Self=\"ub4\"") {
		t.Error("MasterSpread missing Self=\"ub4\" attribute")
	}
	if !strings.Contains(content, "<MarginPreference") {
		t.Error("MasterSpread missing MarginPreference")
	}
}

func TestNewFromTemplate_ResourceParsing(t *testing.T) {
	pkg, err := NewFromTemplate(&TemplateOptions{
		DOMVersion: "20.4",
		Preset:     PresetLetterUS,
	})
	if err != nil {
		t.Fatalf("NewFromTemplate() failed: %v", err)
	}

	// Verify Preferences.xml can be parsed
	prefs, err := pkg.Resource("Resources/Preferences.xml")
	if err != nil {
		t.Fatalf("Resource() failed: %v", err)
	}

	if prefs == nil {
		t.Fatal("Resource() returned nil")
	}

	// Verify resource properties
	if prefs.ResourceType != "Preferences" {
		t.Errorf("ResourceType = %q, want %q", prefs.ResourceType, "Preferences")
	}

	if prefs.DOMVersion == "" {
		t.Error("Resource DOMVersion is empty")
	}

	if len(prefs.RawContent) == 0 {
		t.Error("Resource RawContent is empty")
	}
}

func TestNewFromTemplate_Roundtrip(t *testing.T) {
	tests := []struct {
		name string
		opts *TemplateOptions
	}{
		{
			name: "US Letter",
			opts: &TemplateOptions{Preset: PresetLetterUS},
		},
		{
			name: "A4",
			opts: &TemplateOptions{Preset: PresetA4},
		},
		{
			name: "Tabloid landscape",
			opts: &TemplateOptions{Preset: PresetTabloid, Orientation: "Landscape"},
		},
		{
			name: "US Legal landscape",
			opts: &TemplateOptions{Preset: PresetLegalUS, Orientation: "Landscape"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create package from template
			pkg, err := NewFromTemplate(tt.opts)
			if err != nil {
				t.Fatalf("NewFromTemplate() failed: %v", err)
			}

			// Write to temporary file
			tmpPath := t.TempDir() + "/template_test.idml"
			if err := Write(pkg, tmpPath); err != nil {
				t.Fatalf("Write() failed: %v", err)
			}

			// Read back
			pkg2, err := Read(tmpPath)
			if err != nil {
				t.Fatalf("Read() failed: %v", err)
			}

			// Verify files exist
			if len(pkg2.files) == 0 {
				t.Error("Read package has no files")
			}

			// Verify document can be parsed
			doc, err := pkg2.Document()
			if err != nil {
				t.Fatalf("Document() failed after roundtrip: %v", err)
			}

			if doc == nil {
				t.Fatal("Document() returned nil after roundtrip")
			}
		})
	}
}

func TestDefaultTemplateOptions_HasSensibleDefaults(t *testing.T) {
	opts := DefaultTemplateOptions()

	if opts == nil {
		t.Fatal("DefaultTemplateOptions() returned nil")
	}

	if opts.DOMVersion == "" {
		t.Error("Default DOMVersion is empty")
	}

	if opts.Preset != PresetLetterUS {
		t.Errorf("Default Preset = %v, want %v", opts.Preset, PresetLetterUS)
	}

	if opts.Orientation != "Portrait" {
		t.Errorf("Default Orientation = %v, want Portrait", opts.Orientation)
	}

	if opts.ColumnCount != 1 {
		t.Errorf("Default ColumnCount = %d, want 1", opts.ColumnCount)
	}
}

func TestTemplateEmbedding_EmbeddedTemplatesExist(t *testing.T) {
	// Verify embedded templates are not empty
	if len(minimalDesignMap) == 0 {
		t.Error("minimalDesignMap is empty")
	}

	if len(minimalPreferences) == 0 {
		t.Error("minimalPreferences is empty")
	}

	if len(minimalMasterSpread) == 0 {
		t.Error("minimalMasterSpread is empty")
	}

	// Verify they contain XML declaration
	if string(minimalDesignMap[:5]) != "<?xml" {
		t.Error("minimalDesignMap does not start with XML declaration")
	}

	if string(minimalPreferences[:5]) != "<?xml" {
		t.Error("minimalPreferences does not start with XML declaration")
	}

	if string(minimalMasterSpread[:5]) != "<?xml" {
		t.Error("minimalMasterSpread does not start with XML declaration")
	}
}

func TestGetDimensions_CalculatesDimensions(t *testing.T) {
	tests := []struct {
		name       string
		opts       *TemplateOptions
		wantWidth  float64
		wantHeight float64
	}{
		{
			name:       "US Letter portrait",
			opts:       &TemplateOptions{Preset: PresetLetterUS, Orientation: "Portrait"},
			wantWidth:  612,
			wantHeight: 792,
		},
		{
			name:       "US Letter landscape",
			opts:       &TemplateOptions{Preset: PresetLetterUS, Orientation: "Landscape"},
			wantWidth:  792,
			wantHeight: 612,
		},
		{
			name: "custom dimensions",
			opts: &TemplateOptions{
				Preset:           PresetCustom,
				CustomDimensions: &PageDimensions{Width: 720, Height: 1080},
			},
			wantWidth:  720,
			wantHeight: 1080,
		},
		{
			name: "custom dimensions landscape",
			opts: &TemplateOptions{
				Preset:           PresetCustom,
				CustomDimensions: &PageDimensions{Width: 720, Height: 1080},
				Orientation:      "Landscape",
			},
			wantWidth:  1080,
			wantHeight: 720,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dims := tt.opts.GetDimensions()

			if dims.Width != tt.wantWidth {
				t.Errorf("Width = %.3f, want %.3f", dims.Width, tt.wantWidth)
			}

			if dims.Height != tt.wantHeight {
				t.Errorf("Height = %.3f, want %.3f", dims.Height, tt.wantHeight)
			}
		})
	}
}

func TestStandardPresets_AllPresetsDefined(t *testing.T) {
	// Verify all standard presets are defined
	presets := []DocumentPreset{PresetLetterUS, PresetA4, PresetTabloid, PresetLegalUS}

	for _, preset := range presets {
		dims, exists := StandardPresets[preset]
		if !exists {
			t.Errorf("StandardPresets missing %v", preset)
			continue
		}

		if dims.Width <= 0 || dims.Height <= 0 {
			t.Errorf("Invalid dimensions for %v: %+v", preset, dims)
		}
	}
}
