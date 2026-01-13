// Package resources provides types and functions for working with IDML resource files.
//
// Resource files contain the styling and formatting definitions used throughout
// an IDML document. This includes fonts, colors, paragraph styles, character styles,
// object styles, and graphics settings.
//
// # Resource Files
//
// IDML packages contain several resource files in the Resources/ directory:
//   - Fonts.xml: Font family and font definitions
//   - Styles.xml: Paragraph, character, and object style definitions
//   - Graphic.xml: Colors, swatches, gradients, and stroke styles
//   - Preferences.xml: Document-level preferences and settings
//   - Tags.xml: XML tagging definitions (optional)
//
// # Key Types
//
// ## Font Resources
//   - FontsFile: Root container for font definitions
//   - FontFamily: Groups fonts by family name (e.g., "Minion Pro")
//   - Font: Individual font with style, metrics, and installation status
//   - CompositeFont: Multi-script font definitions (primarily CJK)
//
// ## Style Resources
//   - StylesFile: Root container for all style definitions
//   - ParagraphStyleGroup: Hierarchical paragraph style organization
//   - CharacterStyleGroup: Hierarchical character style organization
//   - ObjectStyleGroup: Hierarchical object style organization
//   - ParagraphStyle, CharacterStyle, ObjectStyle: Individual style definitions
//
// ## Graphic Resources
//   - GraphicFile: Root container for colors and graphics settings
//   - Color: Color definitions (RGB, CMYK, Lab, Spot)
//   - Swatch: Color swatches and tints
//   - Gradient: Gradient definitions
//   - StrokeStyle: Custom stroke patterns and styles
//
// # Usage
//
// Parse a fonts resource file:
//
//	data, _ := os.ReadFile("Resources/Fonts.xml")
//	fonts, err := resources.ParseFontsFile(data)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Access font families
//	for _, family := range fonts.FontFamilies {
//	    fmt.Printf("Family: %s\n", family.Name)
//	    for _, font := range family.Fonts {
//	        fmt.Printf("  Font: %s (%s)\n", font.FontStyleName, font.Status)
//	    }
//	}
//
// Parse a styles resource file:
//
//	data, _ := os.ReadFile("Resources/Styles.xml")
//	styles, err := resources.ParseStylesFile(data)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Access paragraph styles
//	for _, style := range styles.RootParagraphStyleGroup.ParagraphStyles {
//	    fmt.Printf("Paragraph Style: %s\n", style.Name)
//	}
//
// Marshal back to XML:
//
//	xmlData, err := resources.MarshalFontsFile(fonts)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	os.WriteFile("output.xml", xmlData, 0644)
//
// # Namespace Handling
//
// Resource files use the idPkg namespace wrapper:
//
//	<?xml version="1.0" encoding="UTF-8"?>
//	<idPkg:Fonts xmlns:idPkg="..." DOMVersion="20.4">
//	  <FontFamily Self="..." Name="Minion Pro">
//	    <Font Self="..." FontStyleName="Regular" .../>
//	  </FontFamily>
//	</idPkg:Fonts>
//
// The custom UnmarshalXML/MarshalXML methods handle this wrapper correctly.
//
// # Style Hierarchies
//
// Styles support inheritance through BasedOn relationships:
//   - Paragraph styles can be based on other paragraph styles
//   - Character styles can be based on other character styles
//   - Object styles can be based on other object styles
//
// The pkg/analysis package provides tools for resolving these hierarchies
// when exporting IDMS snippets.
//
// # Font Status
//
// Fonts have status indicators:
//   - "Installed": Font is available on the system
//   - "Substituted": Font was substituted with a similar font
//   - "NotAvailable": Font is missing and needs to be installed
//
// # Color Spaces
//
// Colors support multiple color spaces:
//   - RGB: Red, Green, Blue (screen colors)
//   - CMYK: Cyan, Magenta, Yellow, Black (print colors)
//   - Lab: Lightness, A, B (device-independent colors)
//   - Spot: Named spot colors for special inks
//
// # Architecture
//
// This package is part of the domain-specific architecture:
//   - pkg/common: Shared types and utilities
//   - pkg/document: Document metadata and structure
//   - pkg/spread: Page layout and page items
//   - pkg/story: Text content and formatting
//   - pkg/resources: Styling and resource definitions (this package)
//   - pkg/analysis: Dependency tracking and analysis
//   - pkg/idms: IDMS export functionality
package resources
