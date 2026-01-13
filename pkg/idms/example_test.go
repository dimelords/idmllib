package idms_test

import (
	"fmt"
	"log"

	"github.com/dimelords/idmllib/v2/pkg/idml"
	"github.com/dimelords/idmllib/v2/pkg/idms"
)

// ExampleRead demonstrates reading an IDMS snippet file.
func ExampleRead() {
	pkg, err := idms.Read("../../testdata/Snippet_31F27A2D0.idms")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Loaded: %t\n", pkg != nil)
	// Output: Loaded: true
}

// ExampleParse demonstrates parsing IDMS data from bytes.
func ExampleParse() {
	// Read file content first
	pkg, err := idms.Read("../../testdata/Snippet_31F27A2D0.idms")
	if err != nil {
		log.Fatal(err)
	}

	// Marshal back to bytes and parse
	data, err := idms.Marshal(pkg)
	if err != nil {
		log.Fatal(err)
	}

	parsed, err := idms.Parse(data)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Parsed: %t\n", parsed != nil)
	// Output: Parsed: true
}

// ExampleNewExporter demonstrates creating an IDMS exporter from an IDML package.
func ExampleNewExporter() {
	// Read source IDML document
	pkg, err := idml.Read("../../testdata/plain.idml")
	if err != nil {
		log.Fatal(err)
	}

	// Create exporter
	exporter := idms.NewExporter(pkg)
	fmt.Printf("Exporter created: %t\n", exporter != nil)
	// Output: Exporter created: true
}

// ExampleExporter_ExportSelection demonstrates exporting a selection to IDMS.
func ExampleExporter_ExportSelection() {
	// Read source IDML document
	pkg, err := idml.Read("../../testdata/plain.idml")
	if err != nil {
		log.Fatal(err)
	}

	// Get a text frame to export
	spreads, err := pkg.Spreads()
	if err != nil {
		log.Fatal(err)
	}

	// Create selection with first text frame
	selection := idml.NewSelection()
	for _, sp := range spreads {
		if len(sp.InnerSpread.TextFrames) > 0 {
			selection.AddTextFrame(&sp.InnerSpread.TextFrames[0])
			break
		}
	}

	// Export selection
	exporter := idms.NewExporter(pkg)
	snippet, err := exporter.ExportSelection(selection)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Exported: %t\n", snippet != nil)
	// Output: Exported: true
}

// ExampleMarshal demonstrates marshaling an IDMS package to bytes.
func ExampleMarshal() {
	pkg, err := idms.Read("../../testdata/Snippet_31F27A2D0.idms")
	if err != nil {
		log.Fatal(err)
	}

	data, err := idms.Marshal(pkg)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Has data: %t\n", len(data) > 0)
	// Output: Has data: true
}
