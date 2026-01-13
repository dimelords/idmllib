package story

import (
	"encoding/xml"
	"strings"
	"testing"
)

// TestParseStory_InvalidXML tests error handling for invalid XML.
func TestParseStory_InvalidXML(t *testing.T) {
	tests := []struct {
		name string
		xml  string
	}{
		{
			name: "malformed XML",
			xml:  `<Story><unclosed>`,
		},
		{
			name: "empty data",
			xml:  ``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseStory([]byte(tt.xml))
			if err == nil {
				t.Error("ParseStory() should return error for invalid XML")
			}
		})
	}
}

// TestMarshalStory_BasicStructure tests marshaling a basic story structure via roundtrip.
func TestMarshalStory_BasicStructure(t *testing.T) {
	// Use a minimal but valid XML that can be parsed
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<idPkg:Story xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging" DOMVersion="15.0">
	<Story Self="u1d8" AppliedTOCStyle="TOCStyle/$ID/[No TOC Style]">
		<StoryPreference OpticalMarginAlignment="false" OpticalMarginSize="12" />
		<ParagraphStyleRange AppliedParagraphStyle="ParagraphStyle/$ID/NormalParagraphStyle">
			<CharacterStyleRange AppliedCharacterStyle="CharacterStyle/$ID/[No character style]">
				<Content>Test content</Content>
			</CharacterStyleRange>
		</ParagraphStyleRange>
	</Story>
</idPkg:Story>`

	// Parse
	story, err := ParseStory([]byte(xmlData))
	if err != nil {
		t.Fatalf("ParseStory() error = %v", err)
	}

	// Marshal
	data, err := MarshalStory(story)
	if err != nil {
		t.Fatalf("MarshalStory() error = %v", err)
	}

	// Verify XML declaration
	if !strings.HasPrefix(string(data), `<?xml version="1.0"`) {
		t.Error("MarshalStory() should start with XML declaration")
	}

	// Verify namespace
	if !strings.Contains(string(data), `xmlns:idPkg`) {
		t.Error("MarshalStory() should include idPkg namespace")
	}

	// Parse back
	parsed, err := ParseStory(data)
	if err != nil {
		t.Fatalf("ParseStory() after marshal error = %v", err)
	}

	// Verify key fields
	if parsed.DOMVersion != story.DOMVersion {
		t.Errorf("DOMVersion mismatch: %s != %s", parsed.DOMVersion, story.DOMVersion)
	}

	if parsed.StoryElement.Self != story.StoryElement.Self {
		t.Errorf("Story Self mismatch: %s != %s", parsed.StoryElement.Self, story.StoryElement.Self)
	}
}

// TestCharacterStyleRange_WithBr tests unmarshaling CharacterStyleRange with Br elements.
func TestCharacterStyleRange_WithBr(t *testing.T) {
	xmlData := `<CharacterStyleRange AppliedCharacterStyle="CharacterStyle/$ID/[No character style]">
		<Content>Line 1</Content>
		<Br />
		<Content>Line 2</Content>
	</CharacterStyleRange>`

	var csr CharacterStyleRange
	err := unmarshalXML(xmlData, &csr)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	// Verify we have 3 children: Content, Br, Content
	if len(csr.Children) != 3 {
		t.Errorf("Expected 3 children, got %d", len(csr.Children))
	}

	if csr.Children[0].Content == nil {
		t.Error("First child should be Content")
	}

	if csr.Children[1].Br == nil {
		t.Error("Second child should be Br")
	}

	if csr.Children[2].Content == nil {
		t.Error("Third child should be Content")
	}
}

// TestCharacterStyleRange_WithAttributes tests different attribute combinations.
func TestCharacterStyleRange_WithAttributes(t *testing.T) {
	tests := []struct {
		name string
		xml  string
		want CharacterStyleRange
	}{
		{
			name: "with HorizontalScale",
			xml: `<CharacterStyleRange AppliedCharacterStyle="test" HorizontalScale="120">
				<Content>Text</Content>
			</CharacterStyleRange>`,
			want: CharacterStyleRange{
				AppliedCharacterStyle: "test",
				HorizontalScale:       "120",
			},
		},
		{
			name: "with Tracking",
			xml: `<CharacterStyleRange AppliedCharacterStyle="test" Tracking="50">
				<Content>Text</Content>
			</CharacterStyleRange>`,
			want: CharacterStyleRange{
				AppliedCharacterStyle: "test",
				Tracking:              "50",
			},
		},
		{
			name: "with all attributes",
			xml: `<CharacterStyleRange AppliedCharacterStyle="test" HorizontalScale="120" Tracking="50">
				<Content>Text</Content>
			</CharacterStyleRange>`,
			want: CharacterStyleRange{
				AppliedCharacterStyle: "test",
				HorizontalScale:       "120",
				Tracking:              "50",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var csr CharacterStyleRange
			err := unmarshalXML(tt.xml, &csr)
			if err != nil {
				t.Fatalf("Unmarshal error: %v", err)
			}

			if csr.AppliedCharacterStyle != tt.want.AppliedCharacterStyle {
				t.Errorf("AppliedCharacterStyle = %q, want %q", csr.AppliedCharacterStyle, tt.want.AppliedCharacterStyle)
			}

			if csr.HorizontalScale != tt.want.HorizontalScale {
				t.Errorf("HorizontalScale = %q, want %q", csr.HorizontalScale, tt.want.HorizontalScale)
			}

			if csr.Tracking != tt.want.Tracking {
				t.Errorf("Tracking = %q, want %q", csr.Tracking, tt.want.Tracking)
			}
		})
	}
}

// TestCharacterStyleRange_WithUnknownElements tests handling of unknown elements.
func TestCharacterStyleRange_WithUnknownElements(t *testing.T) {
	xmlData := `<CharacterStyleRange AppliedCharacterStyle="test">
		<Content>Text</Content>
		<UnknownElement>Unknown content</UnknownElement>
		<Content>More text</Content>
	</CharacterStyleRange>`

	var csr CharacterStyleRange
	err := unmarshalXML(xmlData, &csr)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	// Should have 3 children
	if len(csr.Children) != 3 {
		t.Fatalf("Expected 3 children, got %d", len(csr.Children))
	}

	// First and third should be Content
	if csr.Children[0].Content == nil {
		t.Error("First child should be Content")
	}

	if csr.Children[2].Content == nil {
		t.Error("Third child should be Content")
	}

	// Second should be Other (unknown element)
	if csr.Children[1].Other == nil {
		t.Error("Second child should be Other (unknown element)")
	}

	if csr.Children[1].Other.XMLName.Local != "UnknownElement" {
		t.Errorf("Unknown element name = %q, want %q", csr.Children[1].Other.XMLName.Local, "UnknownElement")
	}
}

// TestCharacterStyleRange_Marshal tests marshaling CharacterStyleRange.
func TestCharacterStyleRange_Marshal(t *testing.T) {
	csr := CharacterStyleRange{
		XMLName:               xml.Name{Local: "CharacterStyleRange"},
		AppliedCharacterStyle: "CharacterStyle/Test",
		HorizontalScale:       "120",
		Tracking:              "50",
		Children: []CharacterChild{
			{Content: &Content{Text: "Text 1"}},
			{Br: &Br{}},
			{Content: &Content{Text: "Text 2"}},
		},
	}

	data, err := marshalXML(&csr)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	xmlStr := string(data)

	// Verify attributes are present
	if !strings.Contains(xmlStr, `AppliedCharacterStyle="CharacterStyle/Test"`) {
		t.Error("AppliedCharacterStyle attribute missing")
	}

	if !strings.Contains(xmlStr, `HorizontalScale="120"`) {
		t.Error("HorizontalScale attribute missing")
	}

	if !strings.Contains(xmlStr, `Tracking="50"`) {
		t.Error("Tracking attribute missing")
	}

	// Verify children are present
	if !strings.Contains(xmlStr, "Text 1") {
		t.Error("First content missing")
	}

	if !strings.Contains(xmlStr, "Text 2") {
		t.Error("Second content missing")
	}

	// Parse back to verify roundtrip
	var csr2 CharacterStyleRange
	err = unmarshalXML(string(data), &csr2)
	if err != nil {
		t.Fatalf("Unmarshal after marshal error: %v", err)
	}

	if csr2.AppliedCharacterStyle != csr.AppliedCharacterStyle {
		t.Errorf("AppliedCharacterStyle roundtrip mismatch")
	}

	if len(csr2.Children) != len(csr.Children) {
		t.Errorf("Children count mismatch: %d != %d", len(csr2.Children), len(csr.Children))
	}
}

// TestCharacterStyleRange_MarshalWithOtherAttrs tests marshaling with unknown attributes.
func TestCharacterStyleRange_MarshalWithOtherAttrs(t *testing.T) {
	csr := CharacterStyleRange{
		XMLName:               xml.Name{Local: "CharacterStyleRange"},
		AppliedCharacterStyle: "test",
		OtherAttrs: []xml.Attr{
			{Name: xml.Name{Local: "CustomAttr1"}, Value: "value1"},
			{Name: xml.Name{Local: "CustomAttr2"}, Value: "value2"},
		},
		Children: []CharacterChild{
			{Content: &Content{Text: "Text"}},
		},
	}

	data, err := marshalXML(&csr)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	xmlStr := string(data)

	// Verify custom attributes are present
	if !strings.Contains(xmlStr, `CustomAttr1="value1"`) {
		t.Error("CustomAttr1 attribute missing")
	}

	if !strings.Contains(xmlStr, `CustomAttr2="value2"`) {
		t.Error("CustomAttr2 attribute missing")
	}
}

// Helper function to unmarshal XML into a type.
func unmarshalXML(xmlData string, v interface{}) error {
	return xml.Unmarshal([]byte(xmlData), v)
}

// Helper function to marshal a type to XML.
func marshalXML(v interface{}) ([]byte, error) {
	return xml.Marshal(v)
}
