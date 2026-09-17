package spread

import (
	"fmt"

	"github.com/dimelords/idmllib/v3/pkg/common"
)

// NewImageRectangle builds a Rectangle containing a placed Image with a Link
// to a local file, which is the real IDML shape for a placed photo: an image
// is never a top-level page item on its own, it is nested inside a
// graphic-content Rectangle. Confirmed against real production files:
//
//	<Rectangle ContentType="GraphicType"><Image>...<Link .../></Image></Rectangle>
//
// frameX, frameY, frameWidth and frameHeight are the containing frame's
// absolute page-space bounds in points. The geometry is built as a real
// four-corner PathGeometry rather than the plain GeometricBounds attribute,
// because that attribute only works for <Page> elements; using it for a
// content item silently falls back to a tiny default size instead of
// erroring, confirmed for TextFrame and Rectangle against a real InDesign
// open. Anchors are in X Y order, matching this package's own geometry.go.
//
// imageScaleX, imageScaleY, imageOffsetX and imageOffsetY become the placed
// Image's ItemTransform ("sx 0 0 sy tx ty", the format found in production
// files). The caller is expected to know the right fit or crop scale
// already, for example from crop metadata; this does not compute a
// fit-to-frame scale.
//
// ItemTransform is deliberately left unset. It is not a free choice: an
// absent ItemTransform and an explicit identity matrix mean different things
// to InDesign. Setting "1 0 0 1 0 0" on a frame whose PathGeometry is in
// absolute page coordinates moves it off the page, measured against a real
// InDesign open. Note that Scribus's IDML importer silently skips any page
// item that has no ItemTransform, so a document meant to be read by Scribus
// has to supply the transform its own coordinate convention requires.
//
// resourcePath is a local filesystem path such as "/path/to/photo.jpg"; the
// "file:" URI that real IDML Links use is built here, so callers do not
// construct it themselves. resourceFormat is IDML's own format name, for
// example "$ID/JPEG", used for both the Image's ImageTypeName and the Link's
// LinkResourceFormat.
func NewImageRectangle(
	rectSelf, imageSelf, linkSelf string,
	frameX, frameY, frameWidth, frameHeight float64,
	imageScaleX, imageScaleY, imageOffsetX, imageOffsetY float64,
	resourcePath, resourceFormat string,
) *Rectangle {
	corner := func(x, y float64) common.PathPointType {
		anchor := fmt.Sprintf("%g %g", x, y)
		return common.PathPointType{Anchor: anchor, LeftDirection: anchor, RightDirection: anchor}
	}

	frameGeometry := &common.PathGeometry{
		GeometryPathType: &common.GeometryPathType{
			PathOpen: "false",
			PathPointArray: &common.PathPointArray{
				PathPoints: []common.PathPointType{
					corner(frameX, frameY),
					corner(frameX, frameY+frameHeight),
					corner(frameX+frameWidth, frameY+frameHeight),
					corner(frameX+frameWidth, frameY),
				},
			},
		},
	}

	img := &Image{
		FrameContentBase: FrameContentBase{
			Self:          imageSelf,
			ImageTypeName: resourceFormat,
			ItemTransform: fmt.Sprintf("%g 0 0 %g %g %g", imageScaleX, imageScaleY, imageOffsetX, imageOffsetY),
		},
		Link: &Link{
			Self:               linkSelf,
			LinkResourceURI:    "file:" + resourcePath,
			LinkResourceFormat: resourceFormat,
			StoredState:        "Normal",
		},
	}

	return &Rectangle{
		PageItemBase: PageItemBase{Self: rectSelf},
		ContentType:  "GraphicType",
		Properties:   &common.Properties{PathGeometry: frameGeometry},
		Image:        img,
	}
}
