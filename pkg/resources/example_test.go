package resources_test

import (
	"fmt"
	"log"

	"github.com/dimelords/idmllib/pkg/resources"
)

// Example demonstrates basic usage of the resources package.
func Example() {
	// Parse a Styles.xml file
	data := []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<idPkg:Styles xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging" DOMVersion="16.0">
	<RootParagraphStyleGroup Self="pandg">
		<ParagraphStyle Self="ParagraphStyle/$ID/NormalParagraphStyle" Name="$ID/NormalParagraphStyle" />
		<ParagraphStyle Self="ParagraphStyle/CustomStyle" Name="CustomStyle" />
	</RootParagraphStyleGroup>
	<RootCharacterStyleGroup Self="candg">
		<CharacterStyle Self="CharacterStyle/$ID/[No character style]" Name="$ID/[No character style]" />
	</RootCharacterStyleGroup>
</idPkg:Styles>`)

	styles, err := resources.ParseStylesFile(data)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Number of paragraph styles: %d\n", len(styles.RootParagraphStyleGroup.ParagraphStyles))
	fmt.Printf("Number of character styles: %d\n", len(styles.RootCharacterStyleGroup.CharacterStyles))

	// Output:
	// Number of paragraph styles: 2
	// Number of character styles: 1
}

// ExampleStylesFile_FindParagraphStyle demonstrates finding a specific paragraph style.
func ExampleStylesFile_FindParagraphStyle() {
	data := []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<idPkg:Styles xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging" DOMVersion="16.0">
	<RootParagraphStyleGroup Self="pandg">
		<ParagraphStyle Self="ParagraphStyle/$ID/NormalParagraphStyle" Name="$ID/NormalParagraphStyle" />
		<ParagraphStyle Self="ParagraphStyle/CustomStyle" Name="CustomStyle" />
	</RootParagraphStyleGroup>
</idPkg:Styles>`)

	styles, err := resources.ParseStylesFile(data)
	if err != nil {
		log.Fatal(err)
	}

	style := styles.FindParagraphStyle("ParagraphStyle/CustomStyle")
	if style != nil {
		fmt.Printf("Found style: %s\n", style.Name)
	} else {
		fmt.Println("Style not found")
	}

	// Output:
	// Found style: CustomStyle
}

// ExampleFontsFile demonstrates parsing fonts.
func ExampleFontsFile() {
	data := []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<idPkg:Fonts xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging" DOMVersion="16.0">
	<FontFamily Self="di$ID/Arial" Name="Arial">
		<Font Self="di$ID/Arial	Regular" FontFamily="di$ID/Arial" Name="Regular" PostScriptName="ArialMT" />
		<Font Self="di$ID/Arial	Bold" FontFamily="di$ID/Arial" Name="Bold" PostScriptName="Arial-BoldMT" />
	</FontFamily>
</idPkg:Fonts>`)

	fonts, err := resources.ParseFontsFile(data)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Number of font families: %d\n", len(fonts.FontFamilies))
	if len(fonts.FontFamilies) > 0 {
		family := fonts.FontFamilies[0]
		fmt.Printf("First family name: %s\n", family.Name)
		fmt.Printf("Number of fonts in family: %d\n", len(family.Fonts))
	}

	// Output:
	// Number of font families: 1
	// First family name: Arial
	// Number of fonts in family: 2
}
