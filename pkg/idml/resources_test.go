package idml

import (
	"strings"
	"testing"
)

// TestParseResourceFile_ParsesBasicFile tests basic resource file parsing.
func TestParseResourceFile_ParsesBasicFile(t *testing.T) {
	resourceXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<idPkg:Graphic xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging" DOMVersion="20.4">
	<Color Self="Color/Black" Model="Process" Space="CMYK" ColorValue="0 0 0 100" Name="Black" />
	<Color Self="Color/Paper" Model="Process" Space="CMYK" ColorValue="0 0 0 0" Name="Paper" />
	<Ink Self="Ink/$ID/Process Cyan" Name="$ID/Process Cyan" Angle="75" />
</idPkg:Graphic>`

	resource, err := ParseResourceFile([]byte(resourceXML))
	if err != nil {
		t.Fatalf("ParseResourceFile() failed: %v", err)
	}

	// Verify ResourceType
	if resource.ResourceType != "Graphic" {
		t.Errorf("ResourceType = %q, want %q", resource.ResourceType, "Graphic")
	}

	// Verify DOMVersion
	if resource.DOMVersion != "20.4" {
		t.Errorf("DOMVersion = %q, want %q", resource.DOMVersion, "20.4")
	}

	// Verify content was captured
	if len(resource.RawContent) == 0 {
		t.Error("RawContent is empty")
	}

	content := string(resource.RawContent)
	if !strings.Contains(content, "Color/Black") {
		t.Error("RawContent missing Color/Black")
	}
}

// TestMarshalResourceFile_MarshalsFile tests resource file marshaling.
func TestMarshalResourceFile_MarshalsFile(t *testing.T) {
	resource := &ResourceFile{
		ResourceType: "Graphic",
		DOMVersion:   "20.4",
		RawContent: []byte(`	<Color Self="Color/Black" Model="Process" Space="CMYK" ColorValue="0 0 0 100" Name="Black" />
	<Ink Self="Ink/$ID/Process Cyan" Name="$ID/Process Cyan" Angle="75" />`),
	}

	xmlData, err := MarshalResourceFile(resource)
	if err != nil {
		t.Fatalf("MarshalResourceFile() failed: %v", err)
	}

	xmlStr := string(xmlData)

	// Verify XML declaration
	if !strings.Contains(xmlStr, `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`) {
		t.Error("Missing XML declaration")
	}

	// Verify idPkg:Graphic wrapper
	if !strings.Contains(xmlStr, `<idPkg:Graphic`) {
		t.Error("Missing idPkg:Graphic wrapper")
	}

	// Verify namespace declaration
	if !strings.Contains(xmlStr, `xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging"`) {
		t.Error("Missing namespace declaration")
	}

	// Verify content is present
	if !strings.Contains(xmlStr, "Color/Black") {
		t.Error("Missing Color/Black in output")
	}
}

// TestResourceRoundtrip_PreservesData tests parse → marshal → parse round trip.
func TestResourceRoundtrip_PreservesData(t *testing.T) {
	originalXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<idPkg:Graphic xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging" DOMVersion="20.4">
	<Color Self="Color/Black" Model="Process" Space="CMYK" ColorValue="0 0 0 100" Name="Black" />
</idPkg:Graphic>`

	// Parse original
	resource1, err := ParseResourceFile([]byte(originalXML))
	if err != nil {
		t.Fatalf("First ParseResourceFile() failed: %v", err)
	}

	// Marshal to XML
	xmlData, err := MarshalResourceFile(resource1)
	if err != nil {
		t.Fatalf("MarshalResourceFile() failed: %v", err)
	}

	// Parse again
	resource2, err := ParseResourceFile(xmlData)
	if err != nil {
		t.Fatalf("Second ParseResourceFile() failed: %v", err)
	}

	// Verify key fields match
	if resource1.ResourceType != resource2.ResourceType {
		t.Errorf("ResourceType mismatch: %q vs %q", resource1.ResourceType, resource2.ResourceType)
	}

	if resource1.DOMVersion != resource2.DOMVersion {
		t.Errorf("DOMVersion mismatch: %q vs %q", resource1.DOMVersion, resource2.DOMVersion)
	}
}
