package spread

import (
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"image"
	_ "image/jpeg" // DecodeConfig for ImageInfo
	_ "image/png"  // DecodeConfig for ImageInfo

	"github.com/dimelords/idmllib/v3/pkg/common"
)

// IDML's format names for placed images, used for an Image's ImageTypeName and its Link's LinkResourceFormat.
const (
	FormatJPEG = "$ID/JPEG"
	FormatPNG  = "$ID/Portable Network Graphics (PNG)"
)

// EmbeddedImage is an image to place in a frame with the file itself inside the IDML, as InDesign writes an image
// whose link has been embedded: the bytes go base64-encoded into the Image element's Contents, its GraphicBounds
// give its size, and the Link's StoredState is "Embedded". The document then opens with the image in place, without
// the file next to it. ImageInfo reads Format, the size and the resolution from a JPEG or PNG file.
type EmbeddedImage struct {
	// Data is the image file.
	Data []byte
	// Name is the file name the Link records, such as "photo.jpg". No file is needed behind it.
	Name string
	// Format is IDML's format name: FormatJPEG or FormatPNG.
	Format string
	// Width and Height are the image's size in points at its resolution, which is what the frame shows at scale
	// 1: its pixels / its ppi × 72.
	Width, Height float64
	// PPIX and PPIY are its resolution in pixels per inch, written as the Image's ActualPpi. Zero leaves it out,
	// and InDesign reads it from the file.
	PPIX, PPIY float64
}

// NewEmbeddedImageRectangle builds a Rectangle with an embedded image in it, the shape InDesign writes for a placed
// image whose link has been embedded:
//
//	<Rectangle ContentType="GraphicType"><Image><Properties><Contents/><GraphicBounds/></Properties>
//	  <Link StoredState="Embedded"/></Image></Rectangle>
//
// The frame and the image's placement work as in NewImageRectangle: frameX, frameY, frameWidth and frameHeight are
// the frame's bounds, written as a four-corner PathGeometry, and imageScaleX, imageScaleY, imageX and imageY
// become the image's ItemTransform ("sx 0 0 sy x y"), in the frame's coordinates. An image at scale 1 placed at the
// frame's corner shows at its own size; this does not compute a fit.
//
// It refuses data that isn't an image of the format given, a format other than FormatJPEG and FormatPNG, and a
// size or scale that isn't positive.
func NewEmbeddedImageRectangle(
	rectSelf, imageSelf, linkSelf string,
	frameX, frameY, frameWidth, frameHeight float64,
	imageScaleX, imageScaleY, imageX, imageY float64,
	img EmbeddedImage,
) (*Rectangle, error) {
	fail := func(format string, args ...any) (*Rectangle, error) {
		return nil, common.Errorf("spread", "new embedded image rectangle", rectSelf, format, args...)
	}
	switch {
	case img.Format != FormatJPEG && img.Format != FormatPNG:
		return fail("format %q; want %s or %s", img.Format, FormatJPEG, FormatPNG)
	case !hasSignature(img.Data, img.Format):
		return fail("the data is not a %s file", img.Format)
	case img.Width <= 0 || img.Height <= 0:
		return fail("the image is %g × %g pt; both must be positive", img.Width, img.Height)
	case imageScaleX <= 0 || imageScaleY <= 0:
		return fail("the scale is %g × %g; both must be positive", imageScaleX, imageScaleY)
	}
	r := NewImageRectangle(rectSelf, imageSelf, linkSelf, frameX, frameY, frameWidth, frameHeight,
		imageScaleX, imageScaleY, imageX, imageY, img.Name, img.Format)
	r.Image.ItemTransform = fmt.Sprintf("%s 0 0 %s %s %s", num(imageScaleX), num(imageScaleY), num(imageX), num(imageY))
	r.Image.Properties = &common.Properties{OtherElements: []common.RawXMLElement{
		{XMLName: xml.Name{Local: "Contents"}, Content: cdataBase64(img.Data)},
		graphicBoundsElement(0, 0, img.Width, img.Height),
	}}
	if img.PPIX > 0 && img.PPIY > 0 {
		r.Image.ActualPpi = num(img.PPIX) + " " + num(img.PPIY)
	}
	r.Image.Link.StoredState = "Embedded"
	return r, nil
}

// SetGraphicBounds records the image's bounds in its own coordinates: (0, 0) to its size in points at its
// resolution. InDesign needs them to apply the image's ItemTransform. Without them it keeps the transform's
// offset but places the image at 100%, so an image built with NewImageRectangle at a scale other than 1 needs this.
// NewEmbeddedImageRectangle writes them.
func (i *Image) SetGraphicBounds(left, top, right, bottom float64) {
	if i.Properties == nil {
		i.Properties = &common.Properties{}
	}
	el := graphicBoundsElement(left, top, right, bottom)
	for k := range i.Properties.OtherElements {
		if i.Properties.OtherElements[k].XMLName.Local == "GraphicBounds" {
			i.Properties.OtherElements[k] = el
			return
		}
	}
	i.Properties.OtherElements = append(i.Properties.OtherElements, el)
}

// EmbeddedContents returns the file of an image that is embedded in the IDML, decoded from its Contents. It returns
// false when the image has no Contents, which is the case for an image that is only linked.
func (i *Image) EmbeddedContents() ([]byte, bool, error) {
	if i == nil {
		return nil, false, nil
	}
	return embeddedContents(i.Properties, i.Self, "image")
}

// GraphicBounds returns the image's bounds in its own coordinates, as its GraphicBounds record them, and false when
// it has none.
func (i *Image) GraphicBounds() (left, top, right, bottom float64, ok bool) {
	if i == nil {
		return 0, 0, 0, 0, false
	}
	return graphicBounds(i.Properties)
}

func graphicBoundsElement(left, top, right, bottom float64) common.RawXMLElement {
	return common.RawXMLElement{XMLName: xml.Name{Local: "GraphicBounds"}, Attrs: []xml.Attr{
		{Name: xml.Name{Local: "Left"}, Value: num(left)},
		{Name: xml.Name{Local: "Top"}, Value: num(top)},
		{Name: xml.Name{Local: "Right"}, Value: num(right)},
		{Name: xml.Name{Local: "Bottom"}, Value: num(bottom)},
	}}
}

func hasSignature(data []byte, format string) bool {
	switch format {
	case FormatJPEG:
		return bytes.HasPrefix(data, []byte{0xFF, 0xD8, 0xFF})
	case FormatPNG:
		return bytes.HasPrefix(data, []byte("\x89PNG\r\n\x1a\n"))
	}
	return false
}

// ImageFileInfo is what ImageInfo reads from an image file.
type ImageFileInfo struct {
	// Format is IDML's format name: FormatJPEG or FormatPNG.
	Format string
	// Width and Height are the size in pixels.
	Width, Height int
	// PPIX and PPIY are the resolution in pixels per inch: a JPEG's JFIF density, a PNG's pHYs. 72 when the file
	// gives none, as InDesign assumes.
	PPIX, PPIY float64
}

// Size returns the image's size in points at its resolution, as EmbeddedImage and SetGraphicBounds take it.
func (f ImageFileInfo) Size() (width, height float64) {
	return float64(f.Width) / f.PPIX * 72, float64(f.Height) / f.PPIY * 72
}

// ImageInfo reads a JPEG's or PNG's format, size in pixels and resolution, for EmbeddedImage.
func ImageInfo(data []byte) (ImageFileInfo, error) {
	cfg, kind, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return ImageFileInfo{}, common.Errorf("spread", "image info", "", "%v", err)
	}
	info := ImageFileInfo{Width: cfg.Width, Height: cfg.Height, PPIX: 72, PPIY: 72}
	switch kind {
	case "jpeg":
		info.Format = FormatJPEG
		if x, y, ok := jfifDensity(data); ok {
			info.PPIX, info.PPIY = x, y
		}
	case "png":
		info.Format = FormatPNG
		if x, y, ok := pngDensity(data); ok {
			info.PPIX, info.PPIY = x, y
		}
	default:
		return ImageFileInfo{}, common.Errorf("spread", "image info", "", "a %s image; JPEG and PNG are supported", kind)
	}
	return info, nil
}

// jfifDensity reads the resolution in a JPEG's JFIF segment (APP0), which follows the start of image.
func jfifDensity(data []byte) (x, y float64, ok bool) {
	if len(data) < 18 || data[2] != 0xFF || data[3] != 0xE0 || string(data[6:11]) != "JFIF\x00" {
		return 0, 0, false
	}
	dx, dy := float64(binary.BigEndian.Uint16(data[14:16])), float64(binary.BigEndian.Uint16(data[16:18]))
	if dx <= 0 || dy <= 0 {
		return 0, 0, false
	}
	switch data[13] {
	case 1: // dots per inch
		return dx, dy, true
	case 2: // dots per centimetre
		return dx * 2.54, dy * 2.54, true
	}
	return 0, 0, false // an aspect ratio only
}

// pngDensity reads a PNG's pHYs chunk, when its unit is the metre.
func pngDensity(data []byte) (x, y float64, ok bool) {
	for i := 8; i+12 <= len(data); {
		n := int(binary.BigEndian.Uint32(data[i : i+4]))
		kind := string(data[i+4 : i+8])
		body := i + 8
		if body+n > len(data) {
			return 0, 0, false
		}
		if kind == "pHYs" && n == 9 && data[body+8] == 1 {
			ppmX, ppmY := binary.BigEndian.Uint32(data[body:body+4]), binary.BigEndian.Uint32(data[body+4:body+8])
			return float64(ppmX) * 0.0254, float64(ppmY) * 0.0254, ppmX > 0 && ppmY > 0
		}
		if kind == "IDAT" || kind == "IEND" {
			return 0, 0, false
		}
		i = body + n + 4 // the data and its CRC
	}
	return 0, 0, false
}
