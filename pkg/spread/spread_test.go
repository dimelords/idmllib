package spread

import (
	"strings"
	"testing"
)

// TestParseSpread tests basic spread parsing.
func TestParseSpread(t *testing.T) {
	spreadXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<idPkg:Spread xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging" DOMVersion="20.4">
	<Spread Self="ud3" PageTransitionType="None" PageCount="1" ShowMasterItems="true">
		<FlattenerPreference LineArtAndTextResolution="300">
			<Properties>
				<RasterVectorBalance type="double">50</RasterVectorBalance>
			</Properties>
		</FlattenerPreference>
		<Page Self="ud8" Name="1" GeometricBounds="0 0 841.88976377 595.27559055">
			<Properties>
				<PageColor type="enumeration">UseMasterColor</PageColor>
			</Properties>
			<MarginPreference ColumnCount="1" Top="36" Bottom="36" Left="36" Right="36" />
			<GridDataInformation FontStyle="Regular" PointSize="12">
				<Properties>
					<AppliedFont type="string">Minion Pro</AppliedFont>
				</Properties>
			</GridDataInformation>
		</Page>
	</Spread>
</idPkg:Spread>`

	sp, err := ParseSpread([]byte(spreadXML))
	if err != nil {
		t.Fatalf("ParseSpread() failed: %v", err)
	}

	// Verify DOMVersion
	if sp.DOMVersion != "20.4" {
		t.Errorf("DOMVersion = %q, want %q", sp.DOMVersion, "20.4")
	}

	// Verify inner spread attributes
	if sp.Spread.Self != "ud3" {
		t.Errorf("Self = %q, want %q", sp.Spread.Self, "ud3")
	}

	if sp.Spread.PageTransitionType != "None" {
		t.Errorf("PageTransitionType = %q, want %q", sp.Spread.PageTransitionType, "None")
	}

	// Verify Page
	if len(sp.Spread.Pages) != 1 {
		t.Fatalf("Pages count = %d, want 1", len(sp.Spread.Pages))
	}

	page := sp.Spread.Pages[0]
	if page.Self != "ud8" {
		t.Errorf("Page.Self = %q, want %q", page.Self, "ud8")
	}
}

// TestMarshalSpread tests spread marshaling.
func TestMarshalSpread(t *testing.T) {
	sp := &File{
		DOMVersion: "20.4",
		Spread: Spread{
			Self:               "ud3",
			PageTransitionType: "None",
			PageCount:          "1",
			Pages: []Page{
				{
					Self:            "ud8",
					Name:            "1",
					GeometricBounds: "0 0 841.88976377 595.27559055",
				},
			},
		},
	}

	xmlData, err := MarshalSpread(sp)
	if err != nil {
		t.Fatalf("MarshalSpread() failed: %v", err)
	}

	xmlStr := string(xmlData)

	// Verify XML declaration
	if !strings.Contains(xmlStr, `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`) {
		t.Error("Missing XML declaration")
	}

	// Verify idPkg:Spread wrapper
	if !strings.Contains(xmlStr, `<idPkg:Spread`) {
		t.Error("Missing idPkg:Spread wrapper")
	}

	// Verify namespace declaration
	if !strings.Contains(xmlStr, `xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging"`) {
		t.Error("Missing namespace declaration")
	}
}

// TestSpreadRoundtrip tests parse → marshal → parse round trip.
func TestSpreadRoundtrip(t *testing.T) {
	originalXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<idPkg:Spread xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging" DOMVersion="20.4">
	<Spread Self="ud3" PageTransitionType="None" PageCount="1">
		<Page Self="ud8" Name="1"></Page>
	</Spread>
</idPkg:Spread>`

	// Parse original
	sp1, err := ParseSpread([]byte(originalXML))
	if err != nil {
		t.Fatalf("First ParseSpread() failed: %v", err)
	}

	// Marshal to XML
	xmlData, err := MarshalSpread(sp1)
	if err != nil {
		t.Fatalf("MarshalSpread() failed: %v", err)
	}

	// Parse again
	sp2, err := ParseSpread(xmlData)
	if err != nil {
		t.Fatalf("Second ParseSpread() failed: %v", err)
	}

	// Verify key fields match
	if sp1.DOMVersion != sp2.DOMVersion {
		t.Errorf("DOMVersion mismatch: %q vs %q", sp1.DOMVersion, sp2.DOMVersion)
	}

	if sp1.Spread.Self != sp2.Spread.Self {
		t.Errorf("Self mismatch: %q vs %q", sp1.Spread.Self, sp2.Spread.Self)
	}
}
