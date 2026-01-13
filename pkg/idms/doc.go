// Package idms provides functionality for exporting IDML content as IDMS snippets.
//
// IDMS (InDesign Snippet) is a single XML file format that contains selected
// page items and all their dependencies (styles, fonts, colors, etc.) in a
// self-contained format. This allows sharing design elements between documents
// or creating reusable content libraries.
//
// # Key Types
//
//   - Exporter: Main exporter that converts IDML selections to IDMS format
//   - ExportOptions: Configuration options for the export process
//   - Package: Represents an IDMS package with inline resources
//
// # Usage
//
// Export a selection to IDMS:
//
//	// Create a selection of page items
//	selection := idml.NewSelection()
//	selection.AddTextFrame(textFrame)
//	selection.AddRectangle(rectangle)
//
//	// Create an exporter
//	exporter := idms.NewExporter(pkg)
//
//	// Export to IDMS
//	idmsData, err := exporter.Export(selection, nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Write to file
//	err = os.WriteFile("snippet.idms", idmsData, 0644)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Export Process
//
// The IDMS export process involves several steps:
//
//  1. **Dependency Analysis**: Use pkg/analysis to find all resources needed
//     by the selected page items (styles, fonts, colors, stories, etc.)
//
//  2. **Resource Collection**: Gather all identified resources from the source
//     IDML package's resource files (Fonts.xml, Styles.xml, Graphic.xml)
//
//  3. **Content Extraction**: Extract the selected page items and any referenced
//     stories with their complete content
//
//  4. **Inline Assembly**: Combine everything into a single Document structure
//     with inline resources instead of external file references
//
//  5. **XML Generation**: Marshal the complete structure to XML with proper
//     IDMS formatting and namespace declarations
//
// # IDMS vs IDML Structure
//
// IDML uses external files:
//
//	designmap.xml -> References Resources/Styles.xml
//	Spreads/Spread_u210.xml -> Contains page items
//	Stories/Story_u1d8.xml -> Contains text content
//
// IDMS embeds everything inline:
//
//	snippet.idms -> Contains Document with inline styles, page items, and stories
//
// # Supported Page Items
//
// The exporter supports all major InDesign page item types:
//   - TextFrame: Text containers with story content
//   - Rectangle: Rectangular frames (text, image, or empty)
//   - Oval: Elliptical/circular frames
//   - Polygon: Multi-sided shapes
//   - GraphicLine: Vector lines and paths
//   - Group: Grouped collections of page items
//   - Image: Linked images within frames
//
// # Resource Dependencies
//
// The exporter automatically includes all necessary resources:
//   - Paragraph and character styles used in text
//   - Object styles applied to page items
//   - Colors and swatches used in fills and strokes
//   - Font families referenced by styles
//   - Layers containing the page items
//   - External image links (metadata only)
//
// # Error Handling
//
// The package uses the common error handling patterns:
//   - Validation errors for invalid selections
//   - Resource errors for missing dependencies
//   - Export errors for XML generation issues
//
// All errors include context about the operation and affected resources.
package idms
