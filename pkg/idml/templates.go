package idml

import (
	"archive/zip"
	"bytes"
	_ "embed"
	"fmt"
	"sync"
	"text/template"

	"github.com/dimelords/idmllib/pkg/common"
)

// Template initialization with sync.Once for thread-safe lazy loading
var (
	designmapTmpl     *template.Template
	designmapTmplErr  error
	designmapTmplOnce sync.Once

	masterspreadTmpl     *template.Template
	masterspreadTmplErr  error
	masterspreadTmplOnce sync.Once
)

// getDesignmapTemplate returns the parsed designmap template.
// The template is parsed once and cached for subsequent calls.
func getDesignmapTemplate() (*template.Template, error) {
	designmapTmplOnce.Do(func() {
		designmapTmpl, designmapTmplErr = template.New("designmap").Parse(string(minimalDesignMap))
	})
	return designmapTmpl, designmapTmplErr
}

// getMasterspreadTemplate returns the parsed masterspread template.
// The template is parsed once and cached for subsequent calls.
func getMasterspreadTemplate() (*template.Template, error) {
	masterspreadTmplOnce.Do(func() {
		masterspreadTmpl, masterspreadTmplErr = template.New("masterspread").Parse(string(minimalMasterSpread))
	})
	return masterspreadTmpl, masterspreadTmplErr
}

// Template files embedded at compile time.
// These provide minimal valid structures for creating IDML documents from scratch.

//go:embed templates/minimal/mimetype
var minimalMimetype []byte

//go:embed templates/minimal/container.xml
var minimalContainer []byte

//go:embed templates/minimal/designmap.xml
var minimalDesignMap []byte

//go:embed templates/minimal/Preferences.xml
var minimalPreferences []byte

//go:embed templates/minimal/MasterSpread_ub4.xml
var minimalMasterSpread []byte

//go:embed templates/minimal/Graphic.xml
var minimalGraphic []byte

//go:embed templates/minimal/Fonts.xml
var minimalFonts []byte

//go:embed templates/minimal/Styles.xml
var minimalStyles []byte

//go:embed templates/minimal/Tags.xml
var minimalTags []byte

// DocumentPreset defines standard page sizes and configurations.
type DocumentPreset string

const (
	// PresetLetterUS creates a US Letter document (8.5" × 11" / 612 × 792 pt)
	PresetLetterUS DocumentPreset = "letter-us"

	// PresetA4 creates an A4 document (210 × 297 mm / 595.276 × 841.89 pt)
	PresetA4 DocumentPreset = "a4"

	// PresetTabloid creates a Tabloid document (11" × 17" / 792 × 1224 pt)
	PresetTabloid DocumentPreset = "tabloid"

	// PresetLegal creates a Legal document (8,5" × 11" / 612 × 792 mm)
	PresetLegalUS DocumentPreset = "legal"

	// PresetCustom allows custom dimensions
	PresetCustom DocumentPreset = "custom"
)

// PageDimensions holds page width and height in points.
type PageDimensions struct {
	Width  float64 // Width in points (1 point = 1/72 inch)
	Height float64 // Height in points
}

// StandardPresets provides common page dimensions.
var StandardPresets = map[DocumentPreset]PageDimensions{
	PresetLetterUS: {Width: 612, Height: 792},        // 8.5" × 11"
	PresetA4:       {Width: 595.276, Height: 841.89}, // 210mm × 297mm
	PresetTabloid:  {Width: 792, Height: 1224},       // 11" × 17"
	PresetLegalUS:  {Width: 612, Height: 1008},       // 8.5" × 14"
}

// TemplateOptions configures document creation from templates.
type TemplateOptions struct {
	// DOMVersion specifies the InDesign DOM version (e.g., "20.4")
	// If empty, defaults to "20.4"
	DOMVersion string

	// Preset specifies a standard document size
	// Default: PresetLetterUS
	Preset DocumentPreset

	// CustomDimensions allows custom page sizes when Preset is PresetCustom
	CustomDimensions *PageDimensions

	// Orientation determines page layout
	// "Portrait" (default) or "Landscape"
	Orientation string

	// Margins in points (default: 36 points = 0.5 inches)
	Margins struct {
		Top    float64
		Bottom float64
		Left   float64
		Right  float64
	}

	// ColumnCount for text columns (default: 1)
	ColumnCount int

	// ColumnGutter spacing between columns in points (default: 12)
	ColumnGutter float64
}

// DefaultTemplateOptions returns sensible defaults for US Letter portrait.
func DefaultTemplateOptions() *TemplateOptions {
	opts := &TemplateOptions{
		DOMVersion:   "20.4",
		Preset:       PresetLetterUS,
		Orientation:  "Portrait",
		ColumnCount:  1,
		ColumnGutter: 12,
	}
	opts.Margins.Top = 36
	opts.Margins.Bottom = 36
	opts.Margins.Left = 36
	opts.Margins.Right = 36
	return opts
}

// GetDimensions returns the page dimensions based on preset and orientation.
func (opts *TemplateOptions) GetDimensions() PageDimensions {
	var dims PageDimensions

	if opts.Preset == PresetCustom && opts.CustomDimensions != nil {
		dims = *opts.CustomDimensions
	} else {
		dims = StandardPresets[opts.Preset]
	}

	// Swap dimensions for landscape
	if opts.Orientation == "Landscape" {
		dims.Width, dims.Height = dims.Height, dims.Width
	}

	return dims
}

// NewFromTemplate creates a new IDML package from embedded templates.
// This is useful for creating IDML documents from scratch programmatically.
//
// Example:
//
//	// Create default US Letter portrait
//	pkg, err := idml.NewFromTemplate(nil)
//
//	// Create A4 landscape
//	pkg, err := idml.NewFromTemplate(&idml.TemplateOptions{
//	    Preset:      idml.PresetA4,
//	    Orientation: "Landscape",
//	})
//
//	// Create custom size
//	pkg, err := idml.NewFromTemplate(&idml.TemplateOptions{
//	    Preset: idml.PresetCustom,
//	    CustomDimensions: &idml.PageDimensions{
//	        Width:  720,  // 10 inches
//	        Height: 1080, // 15 inches
//	    },
//	})
//
// The created package will have:
//   - designmap.xml with configured document structure
//   - MasterSpreads/MasterSpread_ub4.xml with master page
//   - Resources/Preferences.xml with sensible defaults
//   - Required directory structure
func NewFromTemplate(opts *TemplateOptions) (*Package, error) {
	if opts == nil {
		opts = DefaultTemplateOptions()
	}

	// Ensure defaults
	if opts.DOMVersion == "" {
		opts.DOMVersion = "20.4"
	}
	if opts.Orientation == "" {
		opts.Orientation = "Portrait"
	}
	if opts.ColumnCount <= 0 {
		opts.ColumnCount = 1
	}
	if opts.ColumnGutter <= 0 {
		opts.ColumnGutter = 12
	}

	pkg := New()

	// Get page dimensions
	dims := opts.GetDimensions()

	// Generate customized designmap.xml
	designmap, err := generateDesignMap(opts, dims)
	if err != nil {
		return nil, common.WrapErrorWithPath("idml", "create from template", PathDesignmap, err)
	}
	if err := pkg.addFileFromTemplate(PathDesignmap, designmap); err != nil {
		return nil, common.WrapErrorWithPath("idml", "create from template", PathDesignmap, err)
	}

	// Generate customized MasterSpread
	masterSpread, err := generateMasterSpread(opts, dims)
	if err != nil {
		return nil, common.WrapErrorWithPath("idml", "create from template", PathMasterSpread, err)
	}
	if err := pkg.addFileFromTemplate(PathMasterSpread, masterSpread); err != nil {
		return nil, common.WrapErrorWithPath("idml", "create from template", PathMasterSpread, err)
	}

	// Add minimal Preferences.xml
	if err := pkg.addFileFromTemplate(PathPreferences, minimalPreferences); err != nil {
		return nil, common.WrapErrorWithPath("idml", "create from template", PathPreferences, err)
	}

	// CRITICAL: Add mimetype file (must be first and uncompressed)
	if err := pkg.addFileFromTemplate(PathMimetype, minimalMimetype); err != nil {
		return nil, common.WrapErrorWithPath("idml", "create from template", PathMimetype, err)
	}

	// Add META-INF files
	if err := pkg.addFileFromTemplate(PathContainer, minimalContainer); err != nil {
		return nil, common.WrapErrorWithPath("idml", "create from template", PathContainer, err)
	}

	// Add required resource files
	if err := pkg.addFileFromTemplate(PathGraphic, minimalGraphic); err != nil {
		return nil, common.WrapErrorWithPath("idml", "create from template", PathGraphic, err)
	}

	if err := pkg.addFileFromTemplate(PathFonts, minimalFonts); err != nil {
		return nil, common.WrapErrorWithPath("idml", "create from template", PathFonts, err)
	}

	if err := pkg.addFileFromTemplate(PathStyles, minimalStyles); err != nil {
		return nil, common.WrapErrorWithPath("idml", "create from template", PathStyles, err)
	}

	// Add XML/Tags.xml
	if err := pkg.addFileFromTemplate(PathTags, minimalTags); err != nil {
		return nil, common.WrapErrorWithPath("idml", "create from template", PathTags, err)
	}

	return pkg, nil
}

// generateDesignMap creates a customized designmap.xml based on options.
func generateDesignMap(opts *TemplateOptions, dims PageDimensions) ([]byte, error) {
	tmpl, err := getDesignmapTemplate()
	if err != nil {
		return nil, common.WrapError("idml", "generate design map", fmt.Errorf("failed to parse designmap template: %w", err))
	}

	data := struct {
		DOMVersion   string
		PageWidth    float64
		PageHeight   float64
		Orientation  string
		ColumnCount  int
		ColumnGutter float64
	}{
		DOMVersion:   opts.DOMVersion,
		PageWidth:    dims.Width,
		PageHeight:   dims.Height,
		Orientation:  opts.Orientation,
		ColumnCount:  opts.ColumnCount,
		ColumnGutter: opts.ColumnGutter,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, common.WrapError("idml", "generate design map", fmt.Errorf("failed to execute designmap template: %w", err))
	}

	return buf.Bytes(), nil
}

// generateMasterSpread creates a customized MasterSpread based on options.
func generateMasterSpread(opts *TemplateOptions, dims PageDimensions) ([]byte, error) {
	tmpl, err := getMasterspreadTemplate()
	if err != nil {
		return nil, common.WrapError("idml", "generate master spread", fmt.Errorf("failed to parse masterspread template: %w", err))
	}

	data := struct {
		DOMVersion   string
		PageWidth    float64
		PageHeight   float64
		CenterX      float64
		CenterY      float64
		ColumnCount  int
		ColumnGutter float64
	}{
		DOMVersion:   opts.DOMVersion,
		PageWidth:    dims.Width,
		PageHeight:   dims.Height,
		CenterX:      dims.Width / 2,
		CenterY:      dims.Height / 2,
		ColumnCount:  opts.ColumnCount,
		ColumnGutter: opts.ColumnGutter,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, common.WrapError("idml", "generate master spread", fmt.Errorf("failed to execute masterspread template: %w", err))
	}

	return buf.Bytes(), nil
}

// addFileFromTemplate adds a file to the package from template data.
// This is a helper for NewFromTemplate.
func (p *Package) addFileFromTemplate(path string, data []byte) error {
	// Create a copy of the data to avoid sharing the embedded slice
	fileCopy := make([]byte, len(data))
	copy(fileCopy, data)

	// Create a basic ZIP file header
	header := &zip.FileHeader{
		Name:   path,
		Method: zip.Deflate,
	}
	header.SetMode(0644)

	// Add to files map with proper metadata
	p.files[path] = &fileEntry{
		data:   fileCopy,
		header: header,
	}

	// Preserve file order
	p.fileOrder = append(p.fileOrder, path)

	return nil
}
