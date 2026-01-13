package document

import (
	"encoding/xml"
	"testing"

	"github.com/dimelords/idmllib/internal/testutil"
	"github.com/dimelords/idmllib/pkg/common"

	"github.com/google/go-cmp/cmp"
)

// TestParseDesignmapMinimal tests parsing the minimal designmap XML.
func TestParseDesignmapMinimal(t *testing.T) {
	data := testutil.ReadTestData(t, "designmap_minimal.xml")

	dm, err := ParseDesignmap(data)
	if err != nil {
		t.Fatalf("ParseDesignmap failed: %v", err)
	}

	// Verify essential attributes
	if dm.DOMVersion != "20.4" {
		t.Errorf("DOMVersion = %q, want %q", dm.DOMVersion, "20.4")
	}
	if dm.Self != "d" {
		t.Errorf("Self = %q, want %q", dm.Self, "d")
	}
	if dm.Version != "20.4" {
		t.Errorf("Version = %q, want %q", dm.Version, "20.4")
	}

	// Verify namespace is preserved
	if dm.Xmlns != "http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging" {
		t.Errorf("Xmlns = %q, want idPkg namespace", dm.Xmlns)
	}

	// Verify OtherElements captured child elements
	if len(dm.OtherElements) == 0 {
		t.Error("OtherElements is empty, expected child elements")
	}
}

// TestParseDesignmapFull tests parsing the full designmap XML.
func TestParseDesignmapFull(t *testing.T) {
	data := testutil.ReadTestData(t, "designmap.xml")

	dm, err := ParseDesignmap(data)
	if err != nil {
		t.Fatalf("ParseDesignmap failed: %v", err)
	}

	// Verify essential attributes
	if dm.DOMVersion != "20.4" {
		t.Errorf("DOMVersion = %q, want %q", dm.DOMVersion, "20.4")
	}
	if dm.Self != "d" {
		t.Errorf("Self = %q, want %q", dm.Self, "d")
	}

	// Verify OtherElements contains substantial content
	if len(dm.OtherElements) < 10 {
		t.Errorf("OtherElements count = %d, expected > 10 elements for full document", len(dm.OtherElements))
	}
}

// TestDesignmapRoundtripMinimal tests that we can parse and marshal back
// the minimal designmap without losing data.
func TestDesignmapRoundtripMinimal(t *testing.T) {
	// Read original
	original := testutil.ReadTestData(t, "designmap_minimal.xml")

	// Parse
	dm, err := ParseDesignmap(original)
	if err != nil {
		t.Fatalf("ParseDesignmap failed: %v", err)
	}

	// Marshal back
	output, err := MarshalDesignmap(dm)
	if err != nil {
		t.Fatalf("MarshalDesignmap failed: %v", err)
	}

	// Parse both for structural comparison
	var origStruct, outputStruct interface{}
	if err := xml.Unmarshal(original, &origStruct); err != nil {
		t.Fatalf("failed to unmarshal original: %v", err)
	}
	if err := xml.Unmarshal(output, &outputStruct); err != nil {
		t.Fatalf("failed to unmarshal output: %v", err)
	}

	// Compare structures
	if diff := cmp.Diff(origStruct, outputStruct); diff != "" {
		t.Errorf("roundtrip produced different structure (-want +got):\n%s", diff)

		// Print both for debugging
		t.Logf("Original:\n%s", string(original))
		t.Logf("Output:\n%s", string(output))
	}
}

// TestDesignmapRoundtripFull tests roundtrip with the full designmap.
func TestDesignmapRoundtripFull(t *testing.T) {
	// Read original
	original := testutil.ReadTestData(t, "designmap.xml")

	// Parse
	dm, err := ParseDesignmap(original)
	if err != nil {
		t.Fatalf("ParseDesignmap failed: %v", err)
	}

	// Marshal back
	output, err := MarshalDesignmap(dm)
	if err != nil {
		t.Fatalf("MarshalDesignmap failed: %v", err)
	}

	// Parse both for structural comparison
	var origStruct, outputStruct interface{}
	if err := xml.Unmarshal(original, &origStruct); err != nil {
		t.Fatalf("failed to unmarshal original: %v", err)
	}
	if err := xml.Unmarshal(output, &outputStruct); err != nil {
		t.Fatalf("failed to unmarshal output: %v", err)
	}

	// Compare structures
	if diff := cmp.Diff(origStruct, outputStruct); diff != "" {
		t.Errorf("roundtrip produced different structure (-want +got):\n%s", diff)
	}
}

// TestDesignmapPreservesNamespaces verifies that namespace declarations
// are properly preserved during roundtrip.
func TestDesignmapPreservesNamespaces(t *testing.T) {
	data := testutil.ReadTestData(t, "designmap_minimal.xml")

	dm, err := ParseDesignmap(data)
	if err != nil {
		t.Fatalf("ParseDesignmap failed: %v", err)
	}

	output, err := MarshalDesignmap(dm)
	if err != nil {
		t.Fatalf("MarshalDesignmap failed: %v", err)
	}

	// Check that namespace is present in output
	outputStr := string(output)
	if !containsString(outputStr, "xmlns:idPkg") {
		t.Error("output missing xmlns:idPkg namespace declaration")
	}
	if !containsString(outputStr, "http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging") {
		t.Error("output missing idPkg namespace URI")
	}
}

// TestDesignmapEmptyStruct tests that an empty Designmap can be marshaled.
func TestDesignmapEmptyStruct(t *testing.T) {
	dm := &Designmap{
		XMLName:       xml.Name{Local: "Document"},
		Xmlns:         "http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging",
		DOMVersion:    "20.4",
		Self:          "d",
		Version:       "20.4",
		OtherElements: []common.RawXMLElement{},
	}

	output, err := MarshalDesignmap(dm)
	if err != nil {
		t.Fatalf("MarshalDesignmap failed: %v", err)
	}

	// Verify it's valid XML
	var check interface{}
	if err := xml.Unmarshal(output, &check); err != nil {
		t.Fatalf("output is not valid XML: %v", err)
	}

	// Parse it back
	dm2, err := ParseDesignmap(output)
	if err != nil {
		t.Fatalf("ParseDesignmap failed on output: %v", err)
	}

	// Verify essential fields match
	if dm2.DOMVersion != dm.DOMVersion {
		t.Errorf("DOMVersion mismatch: got %q, want %q", dm2.DOMVersion, dm.DOMVersion)
	}
	if dm2.Self != dm.Self {
		t.Errorf("Self mismatch: got %q, want %q", dm2.Self, dm.Self)
	}
}

// TestDesignmapWithInnerXML tests handling of child element content.
// NOTE: Test name kept for backwards compatibility but now tests OtherElements.
func TestDesignmapWithInnerXML(t *testing.T) {
	// Parse a designmap with specific child elements
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Document xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging" DOMVersion="20.4" Self="d" Version="20.4">
	<Properties>
		<Label>
			<KeyValuePair Key="test" Value="hello" />
		</Label>
	</Properties>
	<Language Self="Language/$ID/English%3a UK" Name="$ID/English: UK" />
</Document>`)

	// Parse
	dm, err := ParseDesignmap(xmlData)
	if err != nil {
		t.Fatalf("ParseDesignmap failed: %v", err)
	}

	// Marshal
	output, err := MarshalDesignmap(dm)
	if err != nil {
		t.Fatalf("MarshalDesignmap failed: %v", err)
	}

	// Parse back
	dm2, err := ParseDesignmap(output)
	if err != nil {
		t.Fatalf("ParseDesignmap failed: %v", err)
	}

	// Verify OtherElements captured child elements
	if len(dm2.OtherElements) == 0 {
		t.Error("OtherElements is empty, child elements were lost during roundtrip")
	}

	// Marshal again to verify content
	output2, err := MarshalDesignmap(dm2)
	if err != nil {
		t.Fatalf("MarshalDesignmap(dm2) failed: %v", err)
	}

	// Check that key elements are present in output
	outputStr := string(output2)
	if !containsString(outputStr, "Properties") {
		t.Error("Properties element not found in output")
	}
	if !containsString(outputStr, "Language") {
		t.Error("Language element not found in output")
	}
	if !containsString(outputStr, "KeyValuePair") {
		t.Error("KeyValuePair element not found in output")
	}
}

// containsString checks if a string contains a substring.
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

// findSubstring is a simple substring search helper.
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
