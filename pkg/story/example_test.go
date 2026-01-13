package story_test

import (
	"fmt"
	"log"

	"github.com/dimelords/idmllib/v2/pkg/story"
)

// Example demonstrates basic usage of the story package.
func Example() {
	// Parse a story XML file
	data := []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<idPkg:Story xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging" DOMVersion="20.4">
	<Story Self="u1d8" TrackChanges="false" StoryTitle="$ID/" AppliedTOCStyle="n">
		<StoryPreference OpticalMarginAlignment="false" OpticalMarginSize="12" />
		<ParagraphStyleRange AppliedParagraphStyle="ParagraphStyle/$ID/NormalParagraphStyle">
			<CharacterStyleRange AppliedCharacterStyle="CharacterStyle/$ID/[No character style]">
				<Content>Hello, World!</Content>
			</CharacterStyleRange>
		</ParagraphStyleRange>
	</Story>
</idPkg:Story>`)

	st, err := story.ParseStory(data)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Story Self: %s\n", st.StoryElement.Self)
	fmt.Printf("Number of paragraph style ranges: %d\n", len(st.StoryElement.ParagraphStyleRanges))

	if len(st.StoryElement.ParagraphStyleRanges) > 0 {
		psr := st.StoryElement.ParagraphStyleRanges[0]
		fmt.Printf("Applied paragraph style: %s\n", psr.AppliedParagraphStyle)

		if len(psr.CharacterStyleRanges) > 0 {
			csr := psr.CharacterStyleRanges[0]
			if len(csr.Children) > 0 && csr.Children[0].Content != nil {
				fmt.Printf("First content: %s\n", csr.Children[0].Content.Text)
			}
		}
	}

	// Output:
	// Story Self: u1d8
	// Number of paragraph style ranges: 1
	// Applied paragraph style: ParagraphStyle/$ID/NormalParagraphStyle
	// First content: Hello, World!
}

// ExampleStory_ExtractText demonstrates extracting all text content from a story.
func ExampleStory_ExtractText() {
	data := []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<idPkg:Story xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging" DOMVersion="20.4">
	<Story Self="u1d8">
		<ParagraphStyleRange AppliedParagraphStyle="ParagraphStyle/$ID/NormalParagraphStyle">
			<CharacterStyleRange AppliedCharacterStyle="CharacterStyle/$ID/[No character style]">
				<Content>First paragraph.</Content>
			</CharacterStyleRange>
		</ParagraphStyleRange>
		<ParagraphStyleRange AppliedParagraphStyle="ParagraphStyle/$ID/NormalParagraphStyle">
			<CharacterStyleRange AppliedCharacterStyle="CharacterStyle/$ID/[No character style]">
				<Content>Second paragraph.</Content>
			</CharacterStyleRange>
		</ParagraphStyleRange>
	</Story>
</idPkg:Story>`)

	st, err := story.ParseStory(data)
	if err != nil {
		log.Fatal(err)
	}

	// Extract all text manually
	var allText string
	for _, psr := range st.StoryElement.ParagraphStyleRanges {
		for _, csr := range psr.CharacterStyleRanges {
			for _, child := range csr.Children {
				if child.Content != nil {
					allText += child.Content.Text
				}
			}
		}
	}

	fmt.Printf("All text: %s\n", allText)

	// Output:
	// All text: First paragraph.Second paragraph.
}
