// Package spread provides types and functions for working with IDML spreads and page items.
//
// The spread package contains all types related to spread layout in IDML files, including:
//   - Spread: The root spread container with page layout information
//   - SpreadElement: The actual spread content with pages and page items
//   - Page: Individual pages within a spread
//   - SpreadTextFrame: Text frames on spreads
//   - Rectangle: Rectangular frames (can contain text, images, or be empty)
//   - Image: Linked images in frames
//   - GraphicLine: Vector line elements
//   - Oval, Polygon, Group: Other page item types
//   - Parsing functions: ParseSpread, MarshalSpread
//
// # Architecture
//
//   - pkg/common: Shared types used across all packages
//   - pkg/document: Document metadata
//   - pkg/spread: Spread and page layout types
//   - pkg/story: Text content and story types
//   - pkg/resources: Styles, fonts, graphics
//
// # Usage
//
// Parse a spread XML file:
//
//	data, _ := os.ReadFile("Spreads/Spread_u210.xml")
//	spread, err := spread.ParseSpread(data)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("Pages:", len(spread.InnerSpread.Pages))
//	fmt.Println("Text frames:", len(spread.InnerSpread.TextFrames))
//	fmt.Println("Rectangles:", len(spread.InnerSpread.Rectangles))
//
// Marshal back to XML:
//
//	xmlData, err := spread.MarshalSpread(spread)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	os.WriteFile("output.xml", xmlData, 0644)
//
// # Namespace Handling
//
// Spreads use the idPkg namespace wrapper:
//
//	<?xml version="1.0" encoding="UTF-8"?>
//	<idPkg:Spread xmlns:idPkg="..." DOMVersion="20.4">
//	  <Spread Self="u210" ...>
//	    <Page Self="u211" .../>
//	    <TextFrame Self="uf3" .../>
//	  </Spread>
//	</idPkg:Spread>
//
// The custom UnmarshalXML/MarshalXML methods handle this wrapper correctly.
package spread
