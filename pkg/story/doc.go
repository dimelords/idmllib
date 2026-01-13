// Package story provides types and functions for working with InDesign Story XML files.
//
// Stories are the primary text containers in InDesign documents. They contain paragraphs,
// character formatting, and all text content.
//
// # Architecture
//
// Story files use a double-wrapped structure with idPkg namespace:
//
//	<idPkg:Story DOMVersion="...">
//	  <Story Self="...">
//	    <ParagraphStyleRange AppliedParagraphStyle="...">
//	      <CharacterStyleRange AppliedCharacterStyle="...">
//	        <Content>Text here</Content>
//	        <Br/>
//	      </CharacterStyleRange>
//	    </ParagraphStyleRange>
//	  </Story>
//	</idPkg:Story>
//
// # Key Types
//
//   - Story: The outer wrapper with DOMVersion and namespace
//   - StoryElement: The inner <Story> element containing actual content
//   - ParagraphStyleRange: Groups characters by paragraph style
//   - CharacterStyleRange: Groups text by character style with custom marshal/unmarshal
//   - Content: Actual text content
//   - Br: Line break element
//
// # Usage
//
// Parse a story XML file:
//
//	data, err := os.ReadFile("Stories/Story_u12a.xml")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	story, err := story.ParseStory(data)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Access story content
//	for _, psr := range story.StoryElement.ParagraphStyleRanges {
//	    for _, csr := range psr.CharacterStyleRanges {
//	        contents := csr.GetContent()
//	        for _, content := range contents {
//	            fmt.Println(content.Text)
//	        }
//	    }
//	}
//
// Marshal a story back to XML:
//
//	data, err := story.MarshalStory(&story)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	os.WriteFile("output.xml", data, 0644)
//
// # Custom Marshaling
//
// CharacterStyleRange uses custom UnmarshalXML/MarshalXML to preserve the exact
// order of Content and Br elements, which is critical for InDesign compatibility.
// The Children field stores mixed content in order.
//
// # Backward Compatibility
//
// Helper methods GetContent(), SetContent(), and AddContent() provide backward
// compatibility for code that doesn't need to deal with mixed content ordering.
package story
