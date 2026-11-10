package spread

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"

	"github.com/dimelords/idmllib/v3/pkg/common"
)

// pdfFormat is IDML's format name for a placed PDF, used for the PDF's ImageTypeName and its Link's
// LinkResourceFormat.
const pdfFormat = "$ID/Adobe Portable Document Format (PDF)"

// EmbeddedPDF is a PDF to place in a frame with the file itself inside the IDML, as InDesign writes a PDF whose link
// has been embedded: the bytes go base64-encoded into the PDF element's Contents, and the Link's StoredState is
// "Embedded". The document then opens with the graphic in place, without the file next to it.
type EmbeddedPDF struct {
	// Data is the PDF file.
	Data []byte
	// Name is the file name the Link records, such as "ad.pdf". No file is needed behind it.
	Name string
	// Width and Height are the size in points of the box that is placed, which is what the frame shows at scale
	// 1: with Crop "CropPDF", the page's crop box (its media box when it has none). Reading them from the PDF is
	// left to the caller.
	Width, Height float64
	// PageNumber is the page of the PDF that is placed, from 1. Zero means 1.
	PageNumber int
	// Crop is IDML's PDFCrop value, which names the box that is placed: "CropPDF" (the default when empty),
	// "CropArt", "CropTrim", "CropBleed", "CropMedia", "CropContentAllLayers" or "CropContentVisibleLayers".
	Crop string
}

// NewPDFRectangle builds a Rectangle with an embedded PDF in it, the shape InDesign writes for a placed PDF:
//
//	<Rectangle ContentType="GraphicType"><PDF><Properties><Contents/><GraphicBounds/></Properties>
//	  <PDFAttribute/><Link StoredState="Embedded"/></PDF></Rectangle>
//
// frameX, frameY, frameWidth and frameHeight are the frame's bounds, written as a four-corner PathGeometry as
// NewImageRectangle does, and the Rectangle's ItemTransform is left unset in the same way. pdfScaleX, pdfScaleY,
// pdfX and pdfY become the PDF's ItemTransform ("sx 0 0 sy x y"), in the frame's coordinates: (pdfX, pdfY) is
// where the placed box's top-left corner goes. A PDF at scale 1 placed at the frame's corner with the frame's
// size fills it exactly; a box larger than the frame, or moved off its corner, is cropped by the frame. This
// does not compute a fit.
//
// It refuses data that doesn't start with a PDF header, a size or scale that isn't positive, and a negative page
// number.
func NewPDFRectangle(
	rectSelf, pdfSelf, linkSelf string,
	frameX, frameY, frameWidth, frameHeight float64,
	pdfScaleX, pdfScaleY, pdfX, pdfY float64,
	pdf EmbeddedPDF,
) (*Rectangle, error) {
	fail := func(format string, args ...any) (*Rectangle, error) {
		return nil, common.Errorf("spread", "new PDF rectangle", rectSelf, format, args...)
	}
	switch {
	case !bytes.HasPrefix(pdf.Data, []byte("%PDF-")):
		return fail("the data is not a PDF: it doesn't start with %%PDF-")
	case pdf.Width <= 0 || pdf.Height <= 0:
		return fail("the placed box is %g × %g pt; both must be positive", pdf.Width, pdf.Height)
	case pdfScaleX <= 0 || pdfScaleY <= 0:
		return fail("the scale is %g × %g; both must be positive", pdfScaleX, pdfScaleY)
	case pdf.PageNumber < 0:
		return fail("page number %d; pages count from 1", pdf.PageNumber)
	}
	page, crop := pdf.PageNumber, pdf.Crop
	if page == 0 {
		page = 1
	}
	if crop == "" {
		crop = "CropPDF"
	}

	r := NewImageRectangle(rectSelf, "", "", frameX, frameY, frameWidth, frameHeight, 1, 1, 0, 0, "", "")
	r.Image = nil
	r.PDF = &PDF{
		FrameContentBase: FrameContentBase{
			Self:          pdfSelf,
			ImageTypeName: pdfFormat,
			ItemTransform: fmt.Sprintf("%s 0 0 %s %s %s", num(pdfScaleX), num(pdfScaleY), num(pdfX), num(pdfY)),
		},
		Properties: &common.Properties{OtherElements: []common.RawXMLElement{
			{XMLName: xml.Name{Local: "Contents"}, Content: cdataBase64(pdf.Data)},
			{XMLName: xml.Name{Local: "GraphicBounds"}, Attrs: []xml.Attr{
				{Name: xml.Name{Local: "Left"}, Value: "0"},
				{Name: xml.Name{Local: "Top"}, Value: "0"},
				{Name: xml.Name{Local: "Right"}, Value: num(pdf.Width)},
				{Name: xml.Name{Local: "Bottom"}, Value: num(pdf.Height)},
			}},
		}},
		PDFAttribute: &PDFAttribute{PageNumber: strconv.Itoa(page), PDFCrop: crop, TransparentBackground: "true"},
		Link: &Link{
			Self:               linkSelf,
			LinkResourceURI:    "file:" + pdf.Name,
			LinkResourceFormat: pdfFormat,
			StoredState:        "Embedded",
		},
	}
	return r, nil
}

// EmbeddedContents returns the file of a PDF that is embedded in the IDML, decoded from its Contents. It returns
// false when the PDF has no Contents, which is the case for a PDF that is only linked.
func (p *PDF) EmbeddedContents() ([]byte, bool, error) {
	if p == nil || p.Properties == nil {
		return nil, false, nil
	}
	for _, el := range p.Properties.OtherElements {
		if el.XMLName.Local != "Contents" {
			continue
		}
		// InDesign writes the base64 text in several CDATA sections, with line breaks.
		text := strings.NewReplacer("<![CDATA[", "", "]]>", "").Replace(string(el.Content))
		text = strings.Join(strings.Fields(text), "")
		data, err := base64.StdEncoding.DecodeString(text)
		if err != nil {
			return nil, true, common.Errorf("spread", "decode embedded PDF", p.Self, "%v", err)
		}
		return data, true, nil
	}
	return nil, false, nil
}

// GraphicBounds returns the bounds of the placed box in the PDF's own coordinates, as the PDF's GraphicBounds
// records them, and false when it has none.
func (p *PDF) GraphicBounds() (left, top, right, bottom float64, ok bool) {
	if p == nil || p.Properties == nil {
		return 0, 0, 0, 0, false
	}
	for _, el := range p.Properties.OtherElements {
		if el.XMLName.Local != "GraphicBounds" {
			continue
		}
		v := map[string]float64{}
		for _, a := range el.Attrs {
			f, err := strconv.ParseFloat(a.Value, 64)
			if err != nil {
				return 0, 0, 0, 0, false
			}
			v[a.Name.Local] = f
		}
		return v["Left"], v["Top"], v["Right"], v["Bottom"], true
	}
	return 0, 0, 0, 0, false
}

// cdataBase64 returns data base64-encoded in a CDATA section, in lines of 76 characters as InDesign writes it.
// Base64 text never contains "]]>", so one section holds all of it.
func cdataBase64(data []byte) []byte {
	enc := base64.StdEncoding.EncodeToString(data)
	var b bytes.Buffer
	b.Grow(len(enc) + len(enc)/76 + 16)
	b.WriteString("<![CDATA[")
	for i := 0; i < len(enc); i += 76 {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(enc[i:min(i+76, len(enc))])
	}
	b.WriteString("]]>")
	return b.Bytes()
}

// num formats a number the shortest way that reads back exactly.
func num(f float64) string { return strconv.FormatFloat(f, 'f', -1, 64) }
