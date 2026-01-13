package spread_test

import (
	"fmt"
	"log"

	"github.com/dimelords/idmllib/pkg/spread"
)

// Example demonstrates basic usage of the spread package.
func Example() {
	// Parse a spread XML file
	data := []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<idPkg:Spread xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging" DOMVersion="16.0">
<Spread Self="u210" FlattenerOverride="Default" ItemTransform="1 0 0 1 0 0">
	<TextFrame Self="u1d8" ParentStory="u1d8" ItemTransform="1 0 0 1 72 72" GeometricBounds="72 72 144 216" />
</Spread>
</idPkg:Spread>`)

	sp, err := spread.ParseSpread(data)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Spread Self: %s\n", sp.InnerSpread.Self)
	fmt.Printf("Number of text frames: %d\n", len(sp.InnerSpread.TextFrames))

	if len(sp.InnerSpread.TextFrames) > 0 {
		tf := sp.InnerSpread.TextFrames[0]
		fmt.Printf("First text frame Self: %s\n", tf.Self)
		fmt.Printf("First text frame ParentStory: %s\n", tf.ParentStory)
	}

	// Output:
	// Spread Self: u210
	// Number of text frames: 1
	// First text frame Self: u1d8
	// First text frame ParentStory: u1d8
}

// ExampleSpread_pageItems demonstrates getting all page items from a spread.
func ExampleSpread_pageItems() {
	data := []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<idPkg:Spread xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging" DOMVersion="16.0">
<Spread Self="u210">
	<TextFrame Self="tf1" />
	<Rectangle Self="rect1" />
	<Oval Self="oval1" />
</Spread>
</idPkg:Spread>`)

	sp, err := spread.ParseSpread(data)
	if err != nil {
		log.Fatal(err)
	}

	// Count all page items
	totalItems := len(sp.InnerSpread.TextFrames) +
		len(sp.InnerSpread.Rectangles) +
		len(sp.InnerSpread.Ovals) +
		len(sp.InnerSpread.Polygons) +
		len(sp.InnerSpread.GraphicLines) +
		len(sp.InnerSpread.Groups) +
		len(sp.InnerSpread.Images)

	fmt.Printf("Total page items: %d\n", totalItems)

	// Show individual items
	for _, tf := range sp.InnerSpread.TextFrames {
		fmt.Printf("TextFrame Self: %s\n", tf.Self)
	}
	for _, rect := range sp.InnerSpread.Rectangles {
		fmt.Printf("Rectangle Self: %s\n", rect.Self)
	}
	for _, oval := range sp.InnerSpread.Ovals {
		fmt.Printf("Oval Self: %s\n", oval.Self)
	}

	// Output:
	// Total page items: 3
	// TextFrame Self: tf1
	// Rectangle Self: rect1
	// Oval Self: oval1
}
