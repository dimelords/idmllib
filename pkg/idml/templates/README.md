# IDML Templates

This directory contains template files for creating IDML documents from scratch.

## Purpose

These templates provide minimal valid XML structures that InDesign will accept. They serve as starting points for programmatically generating IDML documents without requiring an existing document to modify.

## Structure

### `minimal/`

Contains the absolute minimum required files for a valid IDML document:

- **designmap.xml** - Minimal document structure with one page
- **Preferences.xml** - Minimal preferences with sensible defaults

These templates are embedded into the Go binary at compile time using `go:embed` directives, making them available without external file dependencies.

## Usage

### From Go Code

```go
// Create a new IDML document from templates
pkg, err := idml.NewFromTemplate(nil) // uses defaults

// Or with custom options
pkg, err := idml.NewFromTemplate(&idml.TemplateOptions{
    DOMVersion:          "20.4",
    UseMinimalTemplates: true,
})

// Modify the document as needed
doc, _ := pkg.Document()
// ... make changes ...

// Write to file
err = idml.Write(pkg, "output.idml")
```

## Template Files Explained

### Preferences.xml

Contains essential default settings:

- **Print preferences** - Basic print settings (orientation, paper size, etc.)
- **Page item defaults** - Default stroke, fill, corner settings for frames
- **Text preferences** - Typography settings, quote handling
- **Text defaults** - Default font, size, alignment, spacing
- **Frame fitting** - How content fits in frames
- **Story preferences** - Text flow direction and orientation

The minimal version includes only the most essential settings. InDesign will apply its own defaults for any missing preferences.

### designmap.xml

Defines the document structure:

- **Document properties** - Page size, orientation, margins
- **Spreads reference** - Links to spread files (minimal: one spread)
- **Master spreads** - Page templates (minimal: empty master)
- **Languages** - Text language settings
- **View preferences** - Measurement units

## Extending Templates

To add more template variations:

1. Create a new directory (e.g., `templates/standard/`)
2. Add template files to the directory
3. Update `templates.go` with `//go:embed` directives
4. Add options to `TemplateOptions` for selecting templates

Example:

```go
//go:embed templates/standard/Preferences.xml
var standardPreferences []byte

//go:embed templates/standard/Styles.xml
var standardStyles []byte
```

## Validation

All template files should be validated by:

1. Creating an IDML using the templates
2. Opening it in Adobe InDesign
3. Verifying no errors or warnings appear

If InDesign rejects a template, it likely means:
- Missing required elements
- Invalid attribute values
- Incorrect namespace declarations
- Missing referenced resources

## Best Practices

1. **Keep templates minimal** - Only include what's necessary
2. **Use valid defaults** - All attribute values should match InDesign's expectations
3. **Test in InDesign** - Always verify templates open correctly
4. **Document dependencies** - Note which templates require other resources
5. **Version compatibility** - Indicate which InDesign versions support each template

## References

- Adobe InDesign IDML Cookbook (CS5/CS6)
- InDesign SDK Documentation
- IDML format specification: https://www.adobe.com/devnet/indesign/
