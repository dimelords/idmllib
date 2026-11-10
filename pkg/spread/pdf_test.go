package spread

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/dimelords/idmllib/v3/pkg/common"
)

// fakePDF is enough of a PDF for the header check; the bytes are only carried, never read.
var fakePDF = append([]byte("%PDF-1.6\n%\xe2\xe3\xcf\xd3\n"), bytes.Repeat([]byte("0123456789 obj endobj\n"), 20)...)

const pdfSpreadXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<idPkg:Spread xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging" DOMVersion="20.4">
	<Spread Self="ud3" PageCount="1">
		<Page Self="ud8" Name="1"/>
		<Rectangle Self="box" ItemTransform="1 0 0 1 0 0"/>
	</Spread>
</idPkg:Spread>`

func TestNewPDFRectangleRoundtrips(t *testing.T) {
	r, err := NewPDFRectangle("ad", "adp", "adl", 10, 20, 300, 120, 1, 1.02, 3, -4,
		EmbeddedPDF{Data: fakePDF, Name: "ad.pdf", Width: 302.5, Height: 118})
	if err != nil {
		t.Fatal(err)
	}
	f, err := ParseSpread([]byte(pdfSpreadXML))
	if err != nil {
		t.Fatal(err)
	}
	f.Spread.Rectangles = append(f.Spread.Rectangles, *r)
	data, err := MarshalSpread(f)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`<PDF Self="adp"`, `ItemTransform="1 0 0 1.02 3 -4"`,
		`ImageTypeName="$ID/Adobe Portable Document Format (PDF)"`,
		`<GraphicBounds Left="0" Top="0" Right="302.5" Bottom="118"`,
		`<PDFAttribute PageNumber="1" PDFCrop="CropPDF" TransparentBackground="true"`,
		`StoredState="Embedded"`, `LinkResourceURI="file:ad.pdf"`, `<Contents><![CDATA[`,
	} {
		if !strings.Contains(string(data), want) {
			t.Errorf("written spread lacks %s", want)
		}
	}

	back, err := ParseSpread(data)
	if err != nil {
		t.Fatal(err)
	}
	var got *Rectangle
	for i := range back.Spread.Rectangles {
		if back.Spread.Rectangles[i].Self == "ad" {
			got = &back.Spread.Rectangles[i]
		}
	}
	if got == nil || got.PDF == nil {
		t.Fatal("the rectangle with its PDF didn't read back")
	}
	if got.ContentType != "GraphicType" || got.Image != nil {
		t.Errorf("ContentType %q, Image %v; want a graphic frame with only the PDF", got.ContentType, got.Image)
	}
	var anchors []string
	for _, p := range got.Properties.PathGeometry.GeometryPathType.PathPointArray.PathPoints {
		anchors = append(anchors, p.Anchor)
	}
	if got, want := strings.Join(anchors, ", "), "10 20, 10 140, 310 140, 310 20"; got != want {
		t.Errorf("frame corners %s, want %s", got, want)
	}
	contents, embedded, err := got.PDF.EmbeddedContents()
	if err != nil || !embedded || !bytes.Equal(contents, fakePDF) {
		t.Errorf("EmbeddedContents() = %d bytes, %v, %v; want the PDF back", len(contents), embedded, err)
	}
	if l, tp, rt, bt, ok := got.PDF.GraphicBounds(); !ok || l != 0 || tp != 0 || rt != 302.5 || bt != 118 {
		t.Errorf("GraphicBounds() = %v %v %v %v %v", l, tp, rt, bt, ok)
	}

	// Writing what was read changes nothing.
	again, err := MarshalSpread(back)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(again, data) {
		t.Error("a second roundtrip changed the spread")
	}
}

func TestNewPDFRectangleOptions(t *testing.T) {
	r, err := NewPDFRectangle("ad", "adp", "adl", 0, 0, 100, 100, 0.5, 0.5, 0, 0,
		EmbeddedPDF{Data: fakePDF, Name: "two.pdf", Width: 200, Height: 200, PageNumber: 2, Crop: "CropTrim"})
	if err != nil {
		t.Fatal(err)
	}
	if a := r.PDF.PDFAttribute; a.PageNumber != "2" || a.PDFCrop != "CropTrim" {
		t.Errorf("PDFAttribute %+v, want page 2, CropTrim", a)
	}
	if r.ItemTransform != "" {
		t.Errorf("the rectangle's ItemTransform is %q; it is left to the caller", r.ItemTransform)
	}
}

// InDesign writes the base64 text in several CDATA sections with line breaks; EmbeddedContents reads that too.
func TestEmbeddedContentsAsInDesignWritesThem(t *testing.T) {
	const x = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<idPkg:Spread xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging" DOMVersion="20.4">
	<Spread Self="ud3" PageCount="1">
		<Rectangle Self="u1" ContentType="GraphicType" ItemTransform="1 0 0 1 0 0">
			<PDF Self="u2" ItemTransform="1 0 0 1 0 0">
				<Properties>
					<Contents><![CDATA[JVBERi0x
LjYK]]><![CDATA[aGVsbG8=]]></Contents>
					<GraphicBounds Left="0" Top="0" Right="702.9920043945312" Bottom="121.88999938964844" />
				</Properties>
				<Link Self="u3" StoredState="Embedded" />
			</PDF>
		</Rectangle>
		<Rectangle Self="u4" ContentType="GraphicType" ItemTransform="1 0 0 1 0 0">
			<PDF Self="u5" ItemTransform="1 0 0 1 0 0">
				<Link Self="u6" LinkResourceURI="file:/ads/linked.pdf" StoredState="Normal" />
			</PDF>
		</Rectangle>
	</Spread>
</idPkg:Spread>`
	f, err := ParseSpread([]byte(x))
	if err != nil {
		t.Fatal(err)
	}
	data, embedded, err := f.Spread.Rectangles[0].PDF.EmbeddedContents()
	if err != nil || !embedded || string(data) != "%PDF-1.6\nhello" {
		t.Errorf("EmbeddedContents() = %q, %v, %v", data, embedded, err)
	}
	if _, _, r, b, ok := f.Spread.Rectangles[0].PDF.GraphicBounds(); !ok || r != 702.9920043945312 || b != 121.88999938964844 {
		t.Errorf("GraphicBounds() right %v bottom %v ok %v", r, b, ok)
	}
	if data, embedded, err := f.Spread.Rectangles[1].PDF.EmbeddedContents(); data != nil || embedded || err != nil {
		t.Errorf("a linked PDF: EmbeddedContents() = %d bytes, %v, %v; want nothing", len(data), embedded, err)
	}
	if _, _, _, _, ok := f.Spread.Rectangles[1].PDF.GraphicBounds(); ok {
		t.Error("a PDF without GraphicBounds reported some")
	}
}

func TestNewPDFRectangleRefuses(t *testing.T) {
	ok := EmbeddedPDF{Data: fakePDF, Name: "ad.pdf", Width: 100, Height: 50}
	for name, c := range map[string]struct {
		pdf    EmbeddedPDF
		sx, sy float64
	}{
		"not a PDF":     {EmbeddedPDF{Data: []byte("GIF89a"), Width: 100, Height: 50}, 1, 1},
		"no width":      {EmbeddedPDF{Data: fakePDF, Height: 50}, 1, 1},
		"zero scale":    {ok, 0, 1},
		"negative page": {EmbeddedPDF{Data: fakePDF, Width: 100, Height: 50, PageNumber: -1}, 1, 1},
	} {
		r, err := NewPDFRectangle("ad", "adp", "adl", 0, 0, 100, 50, c.sx, c.sy, 0, 0, c.pdf)
		var ce *common.Error
		if r != nil || !errors.As(err, &ce) {
			t.Errorf("%s: got %v, %v; want a common.Error", name, r, err)
		}
	}
}
