package document_test

import (
	"bytes"
	"encoding/xml"
	"testing"

	"github.com/dimelords/idmllib/v2/internal/testutil"
	"github.com/dimelords/idmllib/v2/pkg/document"
	"github.com/dimelords/idmllib/v2/pkg/idml"
	"github.com/google/go-cmp/cmp"
)

// TestParseDocumentMinimal tests parsing a minimal designmap.xml.
func TestParseDocumentMinimal(t *testing.T) {
	data := []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Document xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging" DOMVersion="20.4" Self="d" Version="20.4">
	<Properties>
		<Label>
			<KeyValuePair Key="test" Value="hello" />
		</Label>
	</Properties>
</Document>`)

	doc, err := document.ParseDocument(data)
	if err != nil {
		t.Fatalf("document.ParseDocument() error = %v", err)
	}

	// Check core attributes
	if doc.DOMVersion != "20.4" {
		t.Errorf("DOMVersion = %q, want %q", doc.DOMVersion, "20.4")
	}
	if doc.Self != "d" {
		t.Errorf("Self = %q, want %q", doc.Self, "d")
	}

	// Check namespace
	expectedNS := "http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging"
	if doc.Xmlns != expectedNS {
		t.Errorf("Xmlns = %q, want %q", doc.Xmlns, expectedNS)
	}

	// Check Properties
	if doc.Properties == nil {
		t.Fatal("Properties is nil")
	}
	if doc.Properties.Label == nil {
		t.Fatal("Properties.Label is nil")
	}
	if len(doc.Properties.Label.KeyValuePairs) != 1 {
		t.Fatalf("len(KeyValuePairs) = %d, want 1", len(doc.Properties.Label.KeyValuePairs))
	}

	kv := doc.Properties.Label.KeyValuePairs[0]
	if kv.Key != "test" {
		t.Errorf("KeyValuePair.Key = %q, want %q", kv.Key, "test")
	}
	if kv.Value != "hello" {
		t.Errorf("KeyValuePair.Value = %q, want %q", kv.Value, "hello")
	}
}

// TestParseDocumentFull tests parsing the full example designmap.xml.
func TestParseDocumentFull(t *testing.T) {
	data := testutil.ReadTestData(t, "designmap.xml")

	doc, err := document.ParseDocument(data)
	if err != nil {
		t.Fatalf("document.ParseDocument() error = %v", err)
	}

	// Check core attributes
	if doc.DOMVersion != "20.4" {
		t.Errorf("DOMVersion = %q, want %q", doc.DOMVersion, "20.4")
	}
	if doc.Self != "d" {
		t.Errorf("Self = %q, want %q", doc.Self, "d")
	}
	if doc.Name != "example" {
		t.Errorf("Name = %q, want %q", doc.Name, "example")
	}

	// Check namespace
	expectedNS := "http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging"
	if doc.Xmlns != expectedNS {
		t.Errorf("Xmlns = %q, want %q", doc.Xmlns, expectedNS)
	}

	// Check color management attributes
	if doc.CMYKProfile != "U.S. Web Coated (SWOP) v2" {
		t.Errorf("CMYKProfile = %q, want %q", doc.CMYKProfile, "U.S. Web Coated (SWOP) v2")
	}
	if doc.RGBProfile != "sRGB IEC61966-2.1" {
		t.Errorf("RGBProfile = %q, want %q", doc.RGBProfile, "sRGB IEC61966-2.1")
	}

	// Check layout attributes
	if doc.ZeroPoint != "0 0" {
		t.Errorf("ZeroPoint = %q, want %q", doc.ZeroPoint, "0 0")
	}
	if doc.ActiveLayer != "uba" {
		t.Errorf("ActiveLayer = %q, want %q", doc.ActiveLayer, "uba")
	}

	// Check Properties exists
	if doc.Properties == nil {
		t.Error("Properties is nil")
	}

	// Check that OtherElements preserved other elements (like KinsokuTable, MojikumiTable, etc.)
	if len(doc.OtherElements) == 0 {
		t.Error("OtherElements is empty, should contain KinsokuTable, MojikumiTable, etc.")
	}
}

// TestDocumentRoundtrip tests that we can parse and marshal back.
// This test verifies that all elements (both explicitly parsed and catch-all)
// survive a roundtrip without duplication or data loss.
func TestDocumentRoundtrip(t *testing.T) {

	tests := []struct {
		name string
		file string
	}{
		{"minimal", "designmap_minimal.xml"},
		{"full", "designmap.xml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Read original
			originalData := testutil.ReadTestData(t, tt.file)

			// Parse
			doc, err := document.ParseDocument(originalData)
			if err != nil {
				t.Fatalf("document.ParseDocument() error = %v", err)
			}

			// Marshal back
			outputData, err := document.MarshalDocument(doc)
			if err != nil {
				t.Fatalf("document.MarshalDocument() error = %v", err)
			}

			// Parse again
			doc2, err := document.ParseDocument(outputData)
			if err != nil {
				t.Fatalf("document.ParseDocument(output) error = %v", err)
			}

			// Compare structs with custom options
			opts := cmp.Options{
				// Treat nil and empty byte slices as equal
				cmp.Comparer(func(a, b []byte) bool {
					if len(a) == 0 && len(b) == 0 {
						return true
					}
					return bytes.Equal(a, b)
				}),
				// Compare xml.Attr slices, ignoring xmlns declarations that marshaler adds
				cmp.Comparer(func(a, b []xml.Attr) bool {
					// Filter out xmlns attributes from both slices
					filterXmlns := func(attrs []xml.Attr) []xml.Attr {
						var filtered []xml.Attr
						for _, attr := range attrs {
							if attr.Name.Local != "xmlns" {
								filtered = append(filtered, attr)
							}
						}
						return filtered
					}
					aFiltered := filterXmlns(a)
					bFiltered := filterXmlns(b)

					if len(aFiltered) != len(bFiltered) {
						return false
					}
					for i := range aFiltered {
						if aFiltered[i] != bFiltered[i] {
							return false
						}
					}
					return true
				}),
			}
			if diff := cmp.Diff(doc, doc2, opts); diff != "" {
				t.Errorf("Document differs after roundtrip (-want +got):\n%s", diff)
			}
		})
	}
}

// TestPackageDocument tests accessing Document through Package.
func TestPackageDocument(t *testing.T) {
	// Read test IDML
	pkg, err := idml.Read(testutil.TestDataPath(t, "plain.idml"))
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	// Get Document
	doc, err := pkg.Document()
	if err != nil {
		t.Fatalf("Document() error = %v", err)
	}

	// Check basic attributes
	if doc.DOMVersion == "" {
		t.Error("Document.DOMVersion is empty")
	}
	if doc.Self == "" {
		t.Error("Document.Self is empty")
	}
	if doc.Xmlns == "" {
		t.Error("Document.Xmlns is empty")
	}

	// Second call should return cached document
	doc2, err := pkg.Document()
	if err != nil {
		t.Fatalf("Document() second call error = %v", err)
	}
	if doc != doc2 {
		t.Error("Document() returned different instance on second call (should be cached)")
	}
}

// TestDocumentProperties tests Properties and Label parsing.
func TestDocumentProperties(t *testing.T) {
	data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Document xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging" DOMVersion="20.4" Self="d">
	<Properties>
		<Label>
			<KeyValuePair Key="key1" Value="value1" />
			<KeyValuePair Key="key2" Value="value2" />
			<KeyValuePair Key="key3" Value="value3" />
		</Label>
	</Properties>
</Document>`)

	doc, err := document.ParseDocument(data)
	if err != nil {
		t.Fatalf("document.ParseDocument() error = %v", err)
	}

	if doc.Properties == nil {
		t.Fatal("Properties is nil")
	}
	if doc.Properties.Label == nil {
		t.Fatal("Label is nil")
	}

	kvs := doc.Properties.Label.KeyValuePairs
	if len(kvs) != 3 {
		t.Fatalf("len(KeyValuePairs) = %d, want 3", len(kvs))
	}

	// Check each key-value pair
	expected := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	for i, kv := range kvs {
		expectedValue, ok := expected[kv.Key]
		if !ok {
			t.Errorf("KeyValuePair[%d]: unexpected key %q", i, kv.Key)
			continue
		}
		if kv.Value != expectedValue {
			t.Errorf("KeyValuePair[%d]: Key=%q, Value=%q, want %q",
				i, kv.Key, kv.Value, expectedValue)
		}
	}
}

// TestDocumentNumberingListsAndGrids tests that NumberingLists and NamedGrids parse correctly.
func TestDocumentNumberingListsAndGrids(t *testing.T) {
	// Read test file
	data := testutil.ReadTestData(t, "designmap.xml")

	// Parse
	doc, err := document.ParseDocument(data)
	if err != nil {
		t.Fatalf("document.ParseDocument() error = %v", err)
	}

	// Check NumberingLists
	if len(doc.NumberingLists) == 0 {
		t.Error("NumberingLists is empty, expected at least one")
	}

	if len(doc.NumberingLists) > 0 {
		nl := doc.NumberingLists[0]
		if nl.Self != "NumberingList/$ID/[Default]" {
			t.Errorf("NumberingList.Self = %q, want %q", nl.Self, "NumberingList/$ID/[Default]")
		}
		if nl.Name != "$ID/[Default]" {
			t.Errorf("NumberingList.Name = %q, want %q", nl.Name, "$ID/[Default]")
		}
		if nl.ContinueNumbersAcrossStories != "false" {
			t.Errorf("NumberingList.ContinueNumbersAcrossStories = %q, want %q", nl.ContinueNumbersAcrossStories, "false")
		}
		if nl.ContinueNumbersAcrossDocuments != "false" {
			t.Errorf("NumberingList.ContinueNumbersAcrossDocuments = %q, want %q", nl.ContinueNumbersAcrossDocuments, "false")
		}
	}

	// Check NamedGrids
	if len(doc.NamedGrids) == 0 {
		t.Error("NamedGrids is empty, expected at least one")
	}

	if len(doc.NamedGrids) > 0 {
		ng := doc.NamedGrids[0]
		if ng.Self != "NamedGrid/$ID/[Page Grid]" {
			t.Errorf("NamedGrid.Self = %q, want %q", ng.Self, "NamedGrid/$ID/[Page Grid]")
		}
		if ng.Name != "$ID/[Page Grid]" {
			t.Errorf("NamedGrid.Name = %q, want %q", ng.Name, "$ID/[Page Grid]")
		}

		// Check GridDataInformation
		if ng.GridDataInformation == nil {
			t.Fatal("GridDataInformation is nil")
		}

		gdi := ng.GridDataInformation
		if gdi.FontStyle != "Roman" {
			t.Errorf("GridDataInformation.FontStyle = %q, want %q", gdi.FontStyle, "Roman")
		}
		if gdi.PointSize != "12" {
			t.Errorf("GridDataInformation.PointSize = %q, want %q", gdi.PointSize, "12")
		}
		if gdi.CharacterAki != "0" {
			t.Errorf("GridDataInformation.CharacterAki = %q, want %q", gdi.CharacterAki, "0")
		}
		if gdi.LineAki != "9" {
			t.Errorf("GridDataInformation.LineAki = %q, want %q", gdi.LineAki, "9")
		}
		if gdi.HorizontalScale != "100" {
			t.Errorf("GridDataInformation.HorizontalScale = %q, want %q", gdi.HorizontalScale, "100")
		}
		if gdi.VerticalScale != "100" {
			t.Errorf("GridDataInformation.VerticalScale = %q, want %q", gdi.VerticalScale, "100")
		}

		// Check that Properties exists in GridDataInformation
		if gdi.Properties == nil {
			t.Error("GridDataInformation.Properties is nil")
		}
	}
}

// TestDocumentNumberingListsRoundtrip tests that NumberingLists and NamedGrids roundtrip correctly.
func TestDocumentNumberingListsRoundtrip(t *testing.T) {
	// Read original
	originalData := testutil.ReadTestData(t, "designmap.xml")

	// Parse
	doc, err := document.ParseDocument(originalData)
	if err != nil {
		t.Fatalf("document.ParseDocument() error = %v", err)
	}

	// Marshal
	outputData, err := document.MarshalDocument(doc)
	if err != nil {
		t.Fatalf("document.MarshalDocument() error = %v", err)
	}

	// Parse again
	doc2, err := document.ParseDocument(outputData)
	if err != nil {
		t.Fatalf("document.ParseDocument(output) error = %v", err)
	}

	// Compare NumberingLists
	if len(doc2.NumberingLists) != len(doc.NumberingLists) {
		t.Errorf("NumberingLists count changed: %d -> %d", len(doc.NumberingLists), len(doc2.NumberingLists))
	}

	if len(doc2.NumberingLists) > 0 && len(doc.NumberingLists) > 0 {
		if doc2.NumberingLists[0].Self != doc.NumberingLists[0].Self {
			t.Errorf("NumberingList.Self changed: %q -> %q", doc.NumberingLists[0].Self, doc2.NumberingLists[0].Self)
		}
	}

	// Compare NamedGrids
	if len(doc2.NamedGrids) != len(doc.NamedGrids) {
		t.Errorf("NamedGrids count changed: %d -> %d", len(doc.NamedGrids), len(doc2.NamedGrids))
	}

	if len(doc2.NamedGrids) > 0 && len(doc.NamedGrids) > 0 {
		if doc2.NamedGrids[0].Self != doc.NamedGrids[0].Self {
			t.Errorf("NamedGrid.Self changed: %q -> %q", doc.NamedGrids[0].Self, doc2.NamedGrids[0].Self)
		}
	}
}

// TestDocumentContentReferences tests that content resource references parse correctly.
func TestDocumentContentReferences(t *testing.T) {
	// Read test file
	data := testutil.ReadTestData(t, "designmap.xml")

	// Parse
	doc, err := document.ParseDocument(data)
	if err != nil {
		t.Fatalf("document.ParseDocument() error = %v", err)
	}

	// Check MasterSpreads
	if len(doc.MasterSpreads) == 0 {
		t.Error("MasterSpreads is empty, expected at least one")
	}
	if len(doc.MasterSpreads) > 0 {
		if doc.MasterSpreads[0].Src == "" {
			t.Error("MasterSpread[0].Src is empty")
		}
		t.Logf("First MasterSpread: %s", doc.MasterSpreads[0].Src)
	}

	// Check Spreads
	if len(doc.Spreads) == 0 {
		t.Error("Spreads is empty, expected at least one")
	}
	if len(doc.Spreads) > 0 {
		if doc.Spreads[0].Src == "" {
			t.Error("Spread[0].Src is empty")
		}
		t.Logf("First Spread: %s", doc.Spreads[0].Src)
	}

	// Check Stories
	if len(doc.Stories) == 0 {
		t.Error("Stories is empty, expected at least one")
	}
	if len(doc.Stories) > 0 {
		if doc.Stories[0].Src == "" {
			t.Error("Story[0].Src is empty")
		}
		t.Logf("First Story: %s", doc.Stories[0].Src)
	}

	// Check BackingStory
	if doc.BackingStory == nil {
		t.Error("BackingStory is nil")
	} else {
		if doc.BackingStory.Src == "" {
			t.Error("BackingStory.Src is empty")
		}
		t.Logf("BackingStory: %s", doc.BackingStory.Src)
	}
}

// TestDocumentSectionsAndUsers tests that Sections and DocumentUsers parse correctly.
func TestDocumentSectionsAndUsers(t *testing.T) {
	// Read test file
	data := testutil.ReadTestData(t, "designmap.xml")

	// Parse
	doc, err := document.ParseDocument(data)
	if err != nil {
		t.Fatalf("document.ParseDocument() error = %v", err)
	}

	// Check Sections
	if len(doc.Sections) == 0 {
		t.Error("Sections is empty, expected at least one")
	}

	if len(doc.Sections) > 0 {
		sec := doc.Sections[0]
		if sec.Self != "ub4" {
			t.Errorf("Section.Self = %q, want %q", sec.Self, "ub4")
		}
		if sec.Name != "A" {
			t.Errorf("Section.Name = %q, want %q", sec.Name, "A")
		}
		if sec.Length != "2" {
			t.Errorf("Section.Length = %q, want %q", sec.Length, "2")
		}
		if sec.ContinueNumbering != "false" {
			t.Errorf("Section.ContinueNumbering = %q, want %q", sec.ContinueNumbering, "false")
		}
		if sec.PageNumberStart != "22" {
			t.Errorf("Section.PageNumberStart = %q, want %q", sec.PageNumberStart, "22")
		}
		if sec.SectionPrefix != "A" {
			t.Errorf("Section.SectionPrefix = %q, want %q", sec.SectionPrefix, "A")
		}

		// Check Properties exists
		if sec.Properties == nil {
			t.Error("Section.Properties is nil")
		}
	}

	// Check DocumentUsers
	if len(doc.DocumentUsers) == 0 {
		t.Error("DocumentUsers is empty, expected at least one")
	}

	if len(doc.DocumentUsers) > 0 {
		user := doc.DocumentUsers[0]
		if user.Self != "dDocumentUser0" {
			t.Errorf("DocumentUser.Self = %q, want %q", user.Self, "dDocumentUser0")
		}
		if user.UserName != "$ID/Unknown User Name" {
			t.Errorf("DocumentUser.UserName = %q, want %q", user.UserName, "$ID/Unknown User Name")
		}

		// Check Properties exists
		if user.Properties == nil {
			t.Error("DocumentUser.Properties is nil")
		}
	}
}

// TestDocumentContentReferencesRoundtrip tests that content references roundtrip correctly.
func TestDocumentContentReferencesRoundtrip(t *testing.T) {
	// Read original
	originalData := testutil.ReadTestData(t, "designmap.xml")

	// Parse
	doc, err := document.ParseDocument(originalData)
	if err != nil {
		t.Fatalf("document.ParseDocument() error = %v", err)
	}

	// Marshal
	outputData, err := document.MarshalDocument(doc)
	if err != nil {
		t.Fatalf("document.MarshalDocument() error = %v", err)
	}

	// Parse again
	doc2, err := document.ParseDocument(outputData)
	if err != nil {
		t.Fatalf("document.ParseDocument(output) error = %v", err)
	}

	// Compare counts
	if len(doc2.MasterSpreads) != len(doc.MasterSpreads) {
		t.Errorf("MasterSpreads count changed: %d -> %d", len(doc.MasterSpreads), len(doc2.MasterSpreads))
	}
	if len(doc2.Spreads) != len(doc.Spreads) {
		t.Errorf("Spreads count changed: %d -> %d", len(doc.Spreads), len(doc2.Spreads))
	}
	if len(doc2.Stories) != len(doc.Stories) {
		t.Errorf("Stories count changed: %d -> %d", len(doc.Stories), len(doc2.Stories))
	}
	if len(doc2.Sections) != len(doc.Sections) {
		t.Errorf("Sections count changed: %d -> %d", len(doc.Sections), len(doc2.Sections))
	}
	if len(doc2.DocumentUsers) != len(doc.DocumentUsers) {
		t.Errorf("DocumentUsers count changed: %d -> %d", len(doc.DocumentUsers), len(doc2.DocumentUsers))
	}

	// Compare first elements if they exist
	if len(doc2.Stories) > 0 && len(doc.Stories) > 0 {
		if doc2.Stories[0].Src != doc.Stories[0].Src {
			t.Errorf("Story[0].Src changed: %q -> %q", doc.Stories[0].Src, doc2.Stories[0].Src)
		}
	}
}

// TestDocumentColorGroupsAndBullets tests that ColorGroups and ABullets parse correctly.
func TestDocumentColorGroupsAndBullets(t *testing.T) {
	// Read test file
	data := testutil.ReadTestData(t, "designmap.xml")

	// Parse
	doc, err := document.ParseDocument(data)
	if err != nil {
		t.Fatalf("document.ParseDocument() error = %v", err)
	}

	// Check ColorGroups
	if len(doc.ColorGroups) == 0 {
		t.Error("ColorGroups is empty, expected at least one")
	}

	if len(doc.ColorGroups) > 0 {
		cg := doc.ColorGroups[0]
		if cg.Self != "ColorGroup/[Root Color Group]" {
			t.Errorf("ColorGroup.Self = %q, want %q", cg.Self, "ColorGroup/[Root Color Group]")
		}
		if cg.Name != "[Root Color Group]" {
			t.Errorf("ColorGroup.Name = %q, want %q", cg.Name, "[Root Color Group]")
		}
		if cg.IsRootColorGroup != "true" {
			t.Errorf("ColorGroup.IsRootColorGroup = %q, want %q", cg.IsRootColorGroup, "true")
		}

		// Check ColorGroupSwatches
		if len(cg.ColorGroupSwatches) == 0 {
			t.Error("ColorGroupSwatches is empty, expected at least one")
		}

		if len(cg.ColorGroupSwatches) > 0 {
			swatch := cg.ColorGroupSwatches[0]
			if swatch.SwatchItemRef == "" {
				t.Error("ColorGroupSwatch.SwatchItemRef is empty")
			}
			t.Logf("First swatch: %s -> %s", swatch.Self, swatch.SwatchItemRef)
		}
	}

	// Check ABullets
	if len(doc.ABullets) == 0 {
		t.Error("ABullets is empty, expected at least one")
	}

	if len(doc.ABullets) > 0 {
		bullet := doc.ABullets[0]
		if bullet.Self != "dABullet0" {
			t.Errorf("ABullet.Self = %q, want %q", bullet.Self, "dABullet0")
		}
		if bullet.CharacterType != "UnicodeOnly" {
			t.Errorf("ABullet.CharacterType = %q, want %q", bullet.CharacterType, "UnicodeOnly")
		}
		if bullet.CharacterValue != "8226" {
			t.Errorf("ABullet.CharacterValue = %q, want %q (bullet point)", bullet.CharacterValue, "8226")
		}

		// Check Properties exists
		if bullet.Properties == nil {
			t.Error("ABullet.Properties is nil")
		}
	}
}

// TestDocumentAssignments tests that Assignments parse correctly.
func TestDocumentAssignments(t *testing.T) {
	// Read test file
	data := testutil.ReadTestData(t, "designmap.xml")

	// Parse
	doc, err := document.ParseDocument(data)
	if err != nil {
		t.Fatalf("document.ParseDocument() error = %v", err)
	}

	// Check Assignments
	if len(doc.Assignments) == 0 {
		t.Error("Assignments is empty, expected at least one")
	}

	if len(doc.Assignments) > 0 {
		asn := doc.Assignments[0]
		if asn.Self != "uc9" {
			t.Errorf("Assignment.Self = %q, want %q", asn.Self, "uc9")
		}
		if asn.Name != "$ID/UnassignedInCopy" {
			t.Errorf("Assignment.Name = %q, want %q", asn.Name, "$ID/UnassignedInCopy")
		}
		if asn.ExportOptions != "AssignedSpreads" {
			t.Errorf("Assignment.ExportOptions = %q, want %q", asn.ExportOptions, "AssignedSpreads")
		}
		if asn.IncludeLinksWhenPackage != "true" {
			t.Errorf("Assignment.IncludeLinksWhenPackage = %q, want %q", asn.IncludeLinksWhenPackage, "true")
		}

		// Check Properties exists
		if asn.Properties == nil {
			t.Error("Assignment.Properties is nil")
		}
	}
}

// TestDocumentColorGroupsRoundtrip tests that ColorGroups, ABullets, and Assignments roundtrip correctly.
func TestDocumentColorGroupsRoundtrip(t *testing.T) {
	// Read original
	originalData := testutil.ReadTestData(t, "designmap.xml")

	// Parse
	doc, err := document.ParseDocument(originalData)
	if err != nil {
		t.Fatalf("document.ParseDocument() error = %v", err)
	}

	// Marshal
	outputData, err := document.MarshalDocument(doc)
	if err != nil {
		t.Fatalf("document.MarshalDocument() error = %v", err)
	}

	// Parse again
	doc2, err := document.ParseDocument(outputData)
	if err != nil {
		t.Fatalf("document.ParseDocument(output) error = %v", err)
	}

	// Compare counts
	if len(doc2.ColorGroups) != len(doc.ColorGroups) {
		t.Errorf("ColorGroups count changed: %d -> %d", len(doc.ColorGroups), len(doc2.ColorGroups))
	}
	if len(doc2.ABullets) != len(doc.ABullets) {
		t.Errorf("ABullets count changed: %d -> %d", len(doc.ABullets), len(doc2.ABullets))
	}
	if len(doc2.Assignments) != len(doc.Assignments) {
		t.Errorf("Assignments count changed: %d -> %d", len(doc.Assignments), len(doc2.Assignments))
	}

	// Compare first elements if they exist
	if len(doc2.ColorGroups) > 0 && len(doc.ColorGroups) > 0 {
		if doc2.ColorGroups[0].Self != doc.ColorGroups[0].Self {
			t.Errorf("ColorGroup[0].Self changed: %q -> %q", doc.ColorGroups[0].Self, doc2.ColorGroups[0].Self)
		}
		// Check swatches count
		if len(doc2.ColorGroups[0].ColorGroupSwatches) != len(doc.ColorGroups[0].ColorGroupSwatches) {
			t.Errorf("ColorGroup[0] swatches count changed: %d -> %d",
				len(doc.ColorGroups[0].ColorGroupSwatches),
				len(doc2.ColorGroups[0].ColorGroupSwatches))
		}
	}
}

// TestDocumentTextVariables tests that TextVariables parse correctly.
func TestDocumentTextVariables(t *testing.T) {
	// Read test file
	data := testutil.ReadTestData(t, "designmap.xml")

	// Parse
	doc, err := document.ParseDocument(data)
	if err != nil {
		t.Fatalf("document.ParseDocument() error = %v", err)
	}

	// Check TextVariables
	if len(doc.TextVariables) == 0 {
		t.Error("TextVariables is empty, expected at least one")
	}

	t.Logf("Found %d TextVariables", len(doc.TextVariables))

	// Check for specific variable types
	var foundChapterNumber, foundCreationDate, foundFileName, foundRunningHeader bool

	for _, tv := range doc.TextVariables {
		switch tv.Name {
		case "Chapter Number":
			foundChapterNumber = true
			if tv.VariableType != "ChapterNumberType" {
				t.Errorf("Chapter Number VariableType = %q, want %q", tv.VariableType, "ChapterNumberType")
			}
			if tv.ChapterNumberPreference == nil {
				t.Error("Chapter Number should have ChapterNumberPreference")
			} else {
				if tv.ChapterNumberPreference.Format != "Current" {
					t.Errorf("ChapterNumberPreference.Format = %q, want %q", tv.ChapterNumberPreference.Format, "Current")
				}
			}

		case "Creation Date":
			foundCreationDate = true
			if tv.VariableType != "CreationDateType" {
				t.Errorf("Creation Date VariableType = %q, want %q", tv.VariableType, "CreationDateType")
			}
			if tv.DatePreference == nil {
				t.Error("Creation Date should have DatePreference")
			} else {
				if tv.DatePreference.Format != "dd/MM/yy" {
					t.Errorf("DatePreference.Format = %q, want %q", tv.DatePreference.Format, "dd/MM/yy")
				}
			}

		case "File Name":
			foundFileName = true
			if tv.VariableType != "FileNameType" {
				t.Errorf("File Name VariableType = %q, want %q", tv.VariableType, "FileNameType")
			}
			if tv.FileNamePreference == nil {
				t.Error("File Name should have FileNamePreference")
			} else {
				if tv.FileNamePreference.IncludePath != "false" {
					t.Errorf("FileNamePreference.IncludePath = %q, want %q", tv.FileNamePreference.IncludePath, "false")
				}
				if tv.FileNamePreference.IncludeExtension != "false" {
					t.Errorf("FileNamePreference.IncludeExtension = %q, want %q", tv.FileNamePreference.IncludeExtension, "false")
				}
			}

		case "Running Header":
			foundRunningHeader = true
			if tv.VariableType != "MatchParagraphStyleType" {
				t.Errorf("Running Header VariableType = %q, want %q", tv.VariableType, "MatchParagraphStyleType")
			}
			if tv.MatchParagraphStylePreference == nil {
				t.Error("Running Header should have MatchParagraphStylePreference")
			} else {
				if tv.MatchParagraphStylePreference.SearchStrategy != "FirstOnPage" {
					t.Errorf("MatchParagraphStylePreference.SearchStrategy = %q, want %q",
						tv.MatchParagraphStylePreference.SearchStrategy, "FirstOnPage")
				}
			}
		}
	}

	// Verify we found the expected variables
	if !foundChapterNumber {
		t.Error("Did not find 'Chapter Number' variable")
	}
	if !foundCreationDate {
		t.Error("Did not find 'Creation Date' variable")
	}
	if !foundFileName {
		t.Error("Did not find 'File Name' variable")
	}
	if !foundRunningHeader {
		t.Error("Did not find 'Running Header' variable")
	}
}

// TestDocumentTextVariablesRoundtrip tests that TextVariables roundtrip correctly.
func TestDocumentTextVariablesRoundtrip(t *testing.T) {
	// Read original
	originalData := testutil.ReadTestData(t, "designmap.xml")

	// Parse
	doc, err := document.ParseDocument(originalData)
	if err != nil {
		t.Fatalf("document.ParseDocument() error = %v", err)
	}

	// Marshal
	outputData, err := document.MarshalDocument(doc)
	if err != nil {
		t.Fatalf("document.MarshalDocument() error = %v", err)
	}

	// Parse again
	doc2, err := document.ParseDocument(outputData)
	if err != nil {
		t.Fatalf("document.ParseDocument(output) error = %v", err)
	}

	// Compare counts
	if len(doc2.TextVariables) != len(doc.TextVariables) {
		t.Errorf("TextVariables count changed: %d -> %d", len(doc.TextVariables), len(doc2.TextVariables))
	}

	// Compare first variable if exists
	if len(doc2.TextVariables) > 0 && len(doc.TextVariables) > 0 {
		if doc2.TextVariables[0].Self != doc.TextVariables[0].Self {
			t.Errorf("TextVariable[0].Self changed: %q -> %q",
				doc.TextVariables[0].Self, doc2.TextVariables[0].Self)
		}
		if doc2.TextVariables[0].Name != doc.TextVariables[0].Name {
			t.Errorf("TextVariable[0].Name changed: %q -> %q",
				doc.TextVariables[0].Name, doc2.TextVariables[0].Name)
		}
		if doc2.TextVariables[0].VariableType != doc.TextVariables[0].VariableType {
			t.Errorf("TextVariable[0].VariableType changed: %q -> %q",
				doc.TextVariables[0].VariableType, doc2.TextVariables[0].VariableType)
		}
	}
}
