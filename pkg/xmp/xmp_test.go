package xmp

import (
	"strings"
	"testing"
)

func TestAddThumbnail_EmptyMetadata(t *testing.T) {
	xmp := Parse("")
	err := xmp.AddThumbnail("base64data", 512, 512)
	if err == nil {
		t.Error("Expected error for empty metadata, got nil")
	}
	if !strings.Contains(err.Error(), "no XMP metadata") {
		t.Errorf("Expected error message to contain 'no XMP metadata', got: %s", err.Error())
	}
}

func TestAddThumbnail_Success(t *testing.T) {
	xmpData := `<?xpacket begin="" id="W5M0MpCehiHzreSzNTczkc9d"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
  <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
    <rdf:Description xmlns:xmp="http://ns.adobe.com/xap/1.0/">
      <xmp:CreateDate>2026-01-13T11:50:31+01:00</xmp:CreateDate>
      <xmp:MetadataDate>2026-01-13T11:50:31+01:00</xmp:MetadataDate>
      <xmp:ModifyDate>2026-01-13T11:50:31+01:00</xmp:ModifyDate>
      <xmp:CreatorTool>Adobe InDesign 20.5</xmp:CreatorTool>
    </rdf:Description>
  </rdf:RDF>
</x:xmpmeta>
<?xpacket end="r"?>`

	xmp := Parse(xmpData)
	err := xmp.AddThumbnail("base64encodeddata", 512, 512)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	result := xmp.String()
	if !strings.Contains(result, "<xmp:Thumbnails>") {
		t.Error("Expected result to contain <xmp:Thumbnails>")
	}
	if !strings.Contains(result, "<xmpGImg:format>JPEG</xmpGImg:format>") {
		t.Error("Expected result to contain format")
	}
	if !strings.Contains(result, "<xmpGImg:width>512</xmpGImg:width>") {
		t.Error("Expected result to contain width")
	}
	if !strings.Contains(result, "<xmpGImg:height>512</xmpGImg:height>") {
		t.Error("Expected result to contain height")
	}
	if !strings.Contains(result, "base64encodeddata") {
		t.Error("Expected result to contain thumbnail data")
	}
}

func TestAddThumbnail_ReplacesExisting(t *testing.T) {
	xmpData := `<?xpacket begin="" id="W5M0MpCehiHzreSzNTczkc9d"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
  <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
    <rdf:Description xmlns:xmp="http://ns.adobe.com/xap/1.0/">
      <xmp:CreateDate>2026-01-13T11:50:31+01:00</xmp:CreateDate>
      <xmp:MetadataDate>2026-01-13T11:50:31+01:00</xmp:MetadataDate>
      <xmp:ModifyDate>2026-01-13T11:50:31+01:00</xmp:ModifyDate>
      <xmp:Thumbnails>
         <rdf:Alt>
            <rdf:li rdf:parseType="Resource">
               <xmpGImg:format>JPEG</xmpGImg:format>
               <xmpGImg:width>256</xmpGImg:width>
               <xmpGImg:height>256</xmpGImg:height>
               <xmpGImg:image>olddata</xmpGImg:image>
            </rdf:li>
         </rdf:Alt>
      </xmp:Thumbnails>
      <xmp:CreatorTool>Adobe InDesign 20.5</xmp:CreatorTool>
    </rdf:Description>
  </rdf:RDF>
</x:xmpmeta>
<?xpacket end="r"?>`

	xmp := Parse(xmpData)
	err := xmp.AddThumbnail("newdata", 1024, 1024)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	result := xmp.String()
	if strings.Contains(result, "olddata") {
		t.Error("Expected old thumbnail data to be removed")
	}
	if !strings.Contains(result, "newdata") {
		t.Error("Expected new thumbnail data to be present")
	}
	if !strings.Contains(result, "<xmpGImg:width>1024</xmpGImg:width>") {
		t.Error("Expected new width to be present")
	}
	if !strings.Contains(result, "<xmpGImg:height>1024</xmpGImg:height>") {
		t.Error("Expected new height to be present")
	}
}
