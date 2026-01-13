package idml_test

import (
	"fmt"
	"log"
	"os"

	"github.com/dimelords/idmllib/v2/pkg/idml"
)

// ExampleRead demonstrates reading an IDML file.
func ExampleRead() {
	pkg, err := idml.Read("../../testdata/plain.idml")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Files: %d\n", pkg.FileCount())
	// Output: Files: 13
}

// ExampleWrite demonstrates writing an IDML package to a file.
func ExampleWrite() {
	// Read an existing document
	pkg, err := idml.Read("../../testdata/plain.idml")
	if err != nil {
		log.Fatal(err)
	}

	// Create a temporary file for output
	tmpFile, err := os.CreateTemp("", "example_output_*.idml")
	if err != nil {
		log.Fatal(err)
	}
	tmpFile.Close() // Close the file handle, we just need the path

	// Ensure cleanup even if Write fails
	defer os.Remove(tmpFile.Name())

	// Write to the temporary file
	err = idml.Write(pkg, tmpFile.Name())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Written successfully")
	// Output: Written successfully
}

// ExampleNewFromTemplate demonstrates creating a new IDML document from a template.
func ExampleNewFromTemplate() {
	// Create an A4 document
	pkg, err := idml.NewFromTemplate(&idml.TemplateOptions{
		Preset: idml.PresetA4,
	})
	if err != nil {
		log.Fatal(err)
	}

	doc, err := pkg.Document()
	if err != nil {
		log.Fatal(err)
	}

	// Note: Spreads count may vary based on template configuration
	fmt.Printf("Document loaded: %t\n", doc != nil)
	// Output: Document loaded: true
}

// ExampleNewFromTemplate_letterUS demonstrates creating a US Letter-sized document.
func ExampleNewFromTemplate_letterUS() {
	opts := idml.DefaultTemplateOptions()
	opts.Preset = idml.PresetLetterUS
	pageDims := opts.GetDimensions()

	fmt.Printf("Width: %.0f, Height: %.0f\n", pageDims.Width, pageDims.Height)
	// Output: Width: 612, Height: 792
}

// ExamplePackage_Document demonstrates accessing the parsed designmap.xml.
func ExamplePackage_Document() {
	pkg, err := idml.Read("../../testdata/plain.idml")
	if err != nil {
		log.Fatal(err)
	}

	doc, err := pkg.Document()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Has spreads: %t\n", len(doc.Spreads) > 0)
	// Output: Has spreads: true
}

// ExamplePackage_Stories demonstrates accessing all stories in an IDML document.
func ExamplePackage_Stories() {
	pkg, err := idml.Read("../../testdata/plain.idml")
	if err != nil {
		log.Fatal(err)
	}

	stories, err := pkg.Stories()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Story count: %d\n", len(stories))
	// Output: Story count: 1
}

// ExamplePackage_Spreads demonstrates accessing all spreads in an IDML document.
func ExamplePackage_Spreads() {
	pkg, err := idml.Read("../../testdata/plain.idml")
	if err != nil {
		log.Fatal(err)
	}

	spreads, err := pkg.Spreads()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Spread count: %d\n", len(spreads))
	// Output: Spread count: 1
}

// ExamplePackage_Styles demonstrates accessing the Styles.xml resource.
func ExamplePackage_Styles() {
	pkg, err := idml.Read("../../testdata/plain.idml")
	if err != nil {
		log.Fatal(err)
	}

	styles, err := pkg.Styles()
	if err != nil {
		log.Fatal(err)
	}

	paraCount := 0
	if styles.RootParagraphStyleGroup != nil {
		paraCount = len(styles.RootParagraphStyleGroup.ParagraphStyles)
	}
	fmt.Printf("Has paragraph styles: %t\n", paraCount > 0)
	// Output: Has paragraph styles: true
}

// ExamplePackage_SelectTextFrameByID demonstrates finding a text frame by ID.
func ExamplePackage_SelectTextFrameByID() {
	pkg, err := idml.Read("../../testdata/plain.idml")
	if err != nil {
		log.Fatal(err)
	}

	// Get spreads to find a text frame ID
	spreads, err := pkg.Spreads()
	if err != nil {
		log.Fatal(err)
	}

	// Find first text frame
	for _, sp := range spreads {
		if len(sp.InnerSpread.TextFrames) > 0 {
			tfID := sp.InnerSpread.TextFrames[0].Self
			tf, err := pkg.SelectTextFrameByID(tfID)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Found: %t\n", tf != nil)
			return
		}
	}
	// Output: Found: true
}

// ExampleNewSelection demonstrates creating and populating a selection.
func ExampleNewSelection() {
	selection := idml.NewSelection()
	fmt.Printf("Empty: %t\n", selection.IsEmpty())
	// Output: Empty: true
}

// ExampleStoryPath demonstrates generating story file paths.
func ExampleStoryPath() {
	path := idml.StoryPath("u1d8")
	fmt.Println(path)
	// Output: Stories/Story_u1d8.xml
}

// ExampleSpreadPath demonstrates generating spread file paths.
func ExampleSpreadPath() {
	path := idml.SpreadPath("u210")
	fmt.Println(path)
	// Output: Spreads/Spread_u210.xml
}

// ExampleIsStoryPath demonstrates checking if a path is a story file.
func ExampleIsStoryPath() {
	fmt.Println(idml.IsStoryPath("Stories/Story_u1d8.xml"))
	fmt.Println(idml.IsStoryPath("Spreads/Spread_u210.xml"))
	// Output:
	// true
	// false
}

// ExampleNewResourceManager demonstrates creating a resource manager for validation.
func ExampleNewResourceManager() {
	pkg, err := idml.Read("../../testdata/plain.idml")
	if err != nil {
		log.Fatal(err)
	}

	rm := idml.NewResourceManager(pkg)
	missing, err := rm.FindMissingResources()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Has missing resources: %t\n", missing.HasMissing())
	// Output: Has missing resources: false
}
