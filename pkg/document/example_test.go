package document_test

import (
	"fmt"
	"log"

	"github.com/dimelords/idmllib/v2/pkg/document"
)

// Example demonstrates basic usage of the document package.
func Example() {
	// Parse a designmap.xml file
	data := []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Document xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging" DOMVersion="16.0" Self="d">
	<idPkg:Spread src="Spreads/Spread_u210.xml" />
	<idPkg:Story src="Stories/Story_u1d8.xml" />
</Document>`)

	doc, err := document.ParseDocument(data)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Document Self: %s\n", doc.Self)
	fmt.Printf("Number of spreads: %d\n", len(doc.Spreads))
	fmt.Printf("Number of stories: %d\n", len(doc.Stories))

	// Output:
	// Document Self: d
	// Number of spreads: 1
	// Number of stories: 1
}

// ExampleDocument_spreads demonstrates getting spread filenames.
func ExampleDocument_spreads() {
	data := []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Document xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging" DOMVersion="16.0" Self="d">
	<idPkg:Spread src="Spreads/Spread_u210.xml" />
	<idPkg:Spread src="Spreads/Spread_u220.xml" />
</Document>`)

	doc, err := document.ParseDocument(data)
	if err != nil {
		log.Fatal(err)
	}

	for _, spread := range doc.Spreads {
		fmt.Println(spread.Src)
	}

	// Output:
	// Spreads/Spread_u210.xml
	// Spreads/Spread_u220.xml
}
