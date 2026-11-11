package spread

import (
	"encoding/xml"
	"strconv"

	"github.com/dimelords/idmllib/v3/pkg/common"
)

// Text wrap modes, as IDML names them in TextWrapPreference's TextWrapMode.
const (
	WrapNone        = "None"
	WrapBoundingBox = "BoundingBoxTextWrap"
	WrapContour     = "Contour"
	WrapJumpObject  = "JumpObjectTextWrap"
	WrapNextColumn  = "NextColumnTextWrap"
)

// NewBoundingBoxWrap returns the text wrap InDesign writes for "Wrap around bounding box" on both sides, with the
// offsets given: how far text keeps from the item's top, left, bottom and right edges, in points. It suits a text
// frame or any other page item.
func NewBoundingBoxWrap(top, left, bottom, right float64) *TextWrapPreference {
	w := &TextWrapPreference{
		Inverse: "false", ApplyToMasterPageOnly: "false", TextWrapSide: "BothSides", TextWrapMode: WrapBoundingBox,
	}
	w.SetOffsets(top, left, bottom, right)
	return w
}

// Offsets returns the wrap's offsets from the item's top, left, bottom and right edges, in points, from its
// TextWrapOffset, and false when it has none.
func (w *TextWrapPreference) Offsets() (top, left, bottom, right float64, ok bool) {
	if w == nil || w.Properties == nil {
		return 0, 0, 0, 0, false
	}
	for _, el := range w.Properties.OtherElements {
		if el.XMLName.Local != "TextWrapOffset" {
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
		return v["Top"], v["Left"], v["Bottom"], v["Right"], true
	}
	return 0, 0, 0, 0, false
}

// SetOffsets sets the wrap's offsets from the item's top, left, bottom and right edges, in points, replacing any
// it had.
func (w *TextWrapPreference) SetOffsets(top, left, bottom, right float64) {
	if w.Properties == nil {
		w.Properties = &common.Properties{}
	}
	el := common.RawXMLElement{XMLName: xml.Name{Local: "TextWrapOffset"}, Attrs: []xml.Attr{
		{Name: xml.Name{Local: "Top"}, Value: num(top)},
		{Name: xml.Name{Local: "Left"}, Value: num(left)},
		{Name: xml.Name{Local: "Bottom"}, Value: num(bottom)},
		{Name: xml.Name{Local: "Right"}, Value: num(right)},
	}}
	for i := range w.Properties.OtherElements {
		if w.Properties.OtherElements[i].XMLName.Local == "TextWrapOffset" {
			w.Properties.OtherElements[i] = el
			return
		}
	}
	w.Properties.OtherElements = append(w.Properties.OtherElements, el)
}
