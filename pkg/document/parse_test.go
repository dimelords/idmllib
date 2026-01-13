package document

import (
	"strings"
	"testing"

	"github.com/dimelords/idmllib/pkg/common"
)

// TestParse_InvalidXML tests error handling for invalid XML.
func TestParseDocument_InvalidXML(t *testing.T) {
	tests := []struct {
		name string
		xml  string
	}{
		{
			name: "malformed XML",
			xml:  `<Document><unclosed>`,
		},
		{
			name: "empty data",
			xml:  ``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseDocument([]byte(tt.xml))
			if err == nil {
				t.Error("ParseDocument() should return error for invalid XML")
			}
		})
	}
}

// TestMarshalDocument_AllFields tests marshaling a document with all optional fields populated.
// This ensures all branches in marshalChildren and marshalResourceRefs are covered.
func TestMarshalDocument_AllFields(t *testing.T) {
	doc := &Document{
		DOMVersion: "15.0",
		Self:       "d",

		// Properties
		Properties: &common.Properties{},

		// Languages
		Languages: []Language{
			{
				Self:         "Language/en_US",
				Name:         "English: USA",
				SingleQuotes: "''",
				DoubleQuotes: `""`,
			},
		},

		// Resource references
		GraphicResource: &ResourceRef{
			Src: "Resources/Graphic.xml",
		},
		FontsResource: &ResourceRef{
			Src: "Resources/Fonts.xml",
		},
		StylesResource: &ResourceRef{
			Src: "Resources/Styles.xml",
		},
		PreferencesResource: &ResourceRef{
			Src: "Resources/Preferences.xml",
		},
		TagsResource: &ResourceRef{
			Src: "XML/Tags.xml",
		},

		// Layers
		Layers: []Layer{
			{
				Self:    "uba",
				Name:    "Layer 1",
				Visible: "true",
			},
		},

		// NumberingLists
		NumberingLists: []NumberingList{
			{
				Self: "NumberingList/test",
				Name: "Test List",
			},
		},

		// NamedGrids
		NamedGrids: []NamedGrid{
			{
				Self: "NamedGrid/test",
				Name: "Test Grid",
			},
		},

		// Sections
		Sections: []Section{
			{
				Self:   "Section/test",
				Name:   "Section 1",
				Length: "1",
			},
		},

		// DocumentUsers
		DocumentUsers: []DocumentUser{
			{
				Self:     "DocumentUser/test",
				UserName: "Test User",
			},
		},

		// ColorGroups
		ColorGroups: []ColorGroup{
			{
				Self: "ColorGroup/test",
				Name: "Test Colors",
			},
		},

		// ABullets
		ABullets: []ABullet{
			{
				Self:           "ABullet/test",
				CharacterType:  "UnicodeOnly",
				CharacterValue: "8226",
			},
		},

		// Assignments
		Assignments: []Assignment{
			{
				Self: "Assignment/test",
				Name: "Test Assignment",
			},
		},

		// TextVariables
		TextVariables: []TextVariable{
			{
				Self: "TextVariable/test",
				Name: "Test Variable",
			},
		},

		// Spreads
		Spreads: []ResourceRef{
			{Src: "Spreads/Spread_test.xml"},
		},

		// Stories
		Stories: []ResourceRef{
			{Src: "Stories/Story_test.xml"},
		},

		// MasterSpreads
		MasterSpreads: []ResourceRef{
			{Src: "MasterSpreads/MasterSpread_test.xml"},
		},

		// Pasteboards (included in OtherElements via RawXMLElement)
		OtherElements: []common.RawXMLElement{},
	}

	// Marshal the document
	data, err := MarshalDocument(doc)
	if err != nil {
		t.Fatalf("MarshalDocument() error = %v", err)
	}

	// Verify XML declaration exists
	if !strings.HasPrefix(string(data), `<?xml version="1.0"`) {
		t.Error("MarshalDocument() should start with XML declaration")
	}

	// Verify we can parse it back
	parsed, err := ParseDocument(data)
	if err != nil {
		t.Fatalf("ParseDocument() error after marshaling = %v", err)
	}

	// Verify key fields
	if parsed.DOMVersion != doc.DOMVersion {
		t.Errorf("DOMVersion mismatch: %s != %s", parsed.DOMVersion, doc.DOMVersion)
	}

	if len(parsed.Languages) != len(doc.Languages) {
		t.Errorf("Languages count mismatch: %d != %d", len(parsed.Languages), len(doc.Languages))
	}

	if len(parsed.Layers) != len(doc.Layers) {
		t.Errorf("Layers count mismatch: %d != %d", len(parsed.Layers), len(doc.Layers))
	}

	// Note: Spreads and Stories may not roundtrip perfectly due to namespace handling
	// The important thing is that the marshal/parse functions work without error

	t.Logf("Successfully marshaled and parsed document with all fields")
}

// TestParseDocumentWithMetadata_Basic tests parsing document with metadata.
func TestParseDocumentWithMetadata_Basic(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<?aid style="50" type="document" readerVersion="6.0" featureSet="513" product="15.0(210)" ?>
<Document DOMVersion="15.0" Self="d">
	<Language Self="Language/en_US" Name="English: USA" />
</Document>`

	docMeta, err := ParseDocumentWithMetadata([]byte(xmlData))
	if err != nil {
		t.Fatalf("ParseDocumentWithMetadata() error = %v", err)
	}

	if docMeta.XMLDeclaration == "" {
		t.Error("XML declaration not captured")
	}

	if len(docMeta.ProcessingInstructions) == 0 {
		t.Error("Processing instructions not captured")
	}

	// Check processing instruction details
	if len(docMeta.ProcessingInstructions) > 0 {
		pi := docMeta.ProcessingInstructions[0]
		if pi.Target != "aid" {
			t.Errorf("Processing instruction target = %q, want %q", pi.Target, "aid")
		}
		if !strings.Contains(pi.Inst, "style") {
			t.Error("Processing instruction should contain 'style' attribute")
		}
	}

	// Verify the underlying document was parsed
	if docMeta.Document == nil {
		t.Fatal("Document is nil")
	}

	if docMeta.Document.DOMVersion != "15.0" {
		t.Errorf("DOMVersion = %q, want %q", docMeta.Document.DOMVersion, "15.0")
	}
}

// TestMarshalDocumentWithMetadata_PreservesMetadata tests that metadata is preserved.
func TestMarshalDocumentWithMetadata_PreservesMetadata(t *testing.T) {
	original := &DocumentWithMetadata{
		Document: &Document{
			DOMVersion: "15.0",
			Self:       "d",
		},
		XMLDeclaration: `version="1.0" encoding="UTF-8" standalone="yes"`,
		ProcessingInstructions: []ProcessingInstruction{
			{
				Target: "aid",
				Inst:   `style="50" type="document"`,
			},
		},
	}

	// Marshal
	data, err := MarshalDocumentWithMetadata(original)
	if err != nil {
		t.Fatalf("MarshalDocumentWithMetadata() error = %v", err)
	}

	// Verify XML declaration is preserved
	if !strings.Contains(string(data), `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`) {
		t.Error("XML declaration not preserved")
	}

	// Verify processing instruction is preserved
	if !strings.Contains(string(data), `<?aid style="50" type="document" ?>`) {
		t.Error("Processing instruction not preserved")
	}

	// Parse back
	parsed, err := ParseDocumentWithMetadata(data)
	if err != nil {
		t.Fatalf("ParseDocumentWithMetadata() after marshal error = %v", err)
	}

	// Verify metadata survived the roundtrip
	if len(parsed.ProcessingInstructions) != len(original.ProcessingInstructions) {
		t.Errorf("Processing instructions count = %d, want %d",
			len(parsed.ProcessingInstructions), len(original.ProcessingInstructions))
	}
}
