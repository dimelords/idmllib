package document

import (
	"encoding/xml"

	"github.com/dimelords/idmllib/v3/pkg/common"
)

// idPkgNamespace is the IDML packaging namespace used by resource references.
const idPkgNamespace = "http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging"

// kindOther is the ChildOrder kind used for children kept as RawXMLElement.
const kindOther = "other"

// documentChildKinds lists the child element names Document models with a
// typed field. Anything else is preserved as RawXMLElement.
var documentChildKinds = map[string]bool{
	"Properties": true, "Language": true, "Layer": true, "NumberingList": true,
	"NamedGrid": true, "Section": true, "DocumentUser": true, "ColorGroup": true,
	"ABullet": true, "Assignment": true, "TextVariable": true, "Color": true,
	"Swatch": true, "StrokeStyle": true, "RootCharacterStyleGroup": true,
	"RootParagraphStyleGroup": true, "RootObjectStyleGroup": true,
	"TinDocumentDataObject": true, "TransparencyDefaultContainerObject": true,
	"Spread": true, "Story": true,
}

// recordChild appends the kind of the child element that starts at start to
// the document's child order. Resource references are keyed by "idPkg:Name".
func (d *Document) recordChild(start xml.StartElement) {
	switch {
	case start.Name.Space == idPkgNamespace:
		d.childOrder.Record("idPkg:" + start.Name.Local)
	case documentChildKinds[start.Name.Local]:
		d.childOrder.Record(start.Name.Local)
	default:
		d.childOrder.Record(kindOther)
	}
}

func wrapMarshal(err error) error {
	if err != nil {
		return common.WrapError("document", "marshal document", err)
	}
	return nil
}

// ptrKind returns a ChildKind for a single optional child.
func ptrKind(name string, present bool, v any) common.ChildKind {
	return common.OptionalKind(name, present, func(e *xml.Encoder) error { return wrapMarshal(e.Encode(v)) })
}

// sliceKind returns a ChildKind for a repeated child; get returns the i-th item.
func sliceKind(name string, n int, get func(i int) any) common.ChildKind {
	return common.ChildKind{Name: name, Len: n, Encode: func(e *xml.Encoder, i int) error {
		return wrapMarshal(e.Encode(get(i)))
	}}
}

// marshalChildren writes the Document's children, replaying the order recorded
// while parsing. Children of a Document built programmatically (no recorded
// order) are written grouped by kind in the order listed below, which matches
// the order InDesign itself uses for a freshly exported designmap.xml.
func (d Document) marshalChildren(e *xml.Encoder) error {
	kinds := []common.ChildKind{
		ptrKind("Properties", d.Properties != nil, d.Properties),
		sliceKind("Language", len(d.Languages), func(i int) any { return d.Languages[i] }),

		ptrKind("idPkg:Graphic", d.GraphicResource != nil, d.GraphicResource),
		ptrKind("idPkg:Fonts", d.FontsResource != nil, d.FontsResource),
		ptrKind("idPkg:Styles", d.StylesResource != nil, d.StylesResource),
		ptrKind("idPkg:Preferences", d.PreferencesResource != nil, d.PreferencesResource),
		ptrKind("idPkg:Tags", d.TagsResource != nil, d.TagsResource),
		sliceKind("idPkg:MasterSpread", len(d.MasterSpreads), func(i int) any { return d.MasterSpreads[i] }),
		sliceKind("idPkg:Spread", len(d.Spreads), func(i int) any { return d.Spreads[i] }),
		sliceKind("idPkg:Story", len(d.Stories), func(i int) any { return d.Stories[i] }),
		ptrKind("idPkg:BackingStory", d.BackingStory != nil, d.BackingStory),

		sliceKind("Layer", len(d.Layers), func(i int) any { return d.Layers[i] }),
		sliceKind("NumberingList", len(d.NumberingLists), func(i int) any { return d.NumberingLists[i] }),
		sliceKind("NamedGrid", len(d.NamedGrids), func(i int) any { return d.NamedGrids[i] }),
		sliceKind("Section", len(d.Sections), func(i int) any { return d.Sections[i] }),
		sliceKind("DocumentUser", len(d.DocumentUsers), func(i int) any { return d.DocumentUsers[i] }),
		sliceKind("ColorGroup", len(d.ColorGroups), func(i int) any { return d.ColorGroups[i] }),
		sliceKind("ABullet", len(d.ABullets), func(i int) any { return d.ABullets[i] }),
		sliceKind("Assignment", len(d.Assignments), func(i int) any { return d.Assignments[i] }),
		sliceKind("TextVariable", len(d.TextVariables), func(i int) any { return d.TextVariables[i] }),

		// IDMS inline content (used by snippets instead of resource references)
		sliceKind("Color", len(d.Colors), func(i int) any { return d.Colors[i] }),
		sliceKind("Swatch", len(d.Swatches), func(i int) any { return d.Swatches[i] }),
		sliceKind("StrokeStyle", len(d.StrokeStyles), func(i int) any { return d.StrokeStyles[i] }),
		namedKind("RootCharacterStyleGroup", d.RootCharacterStyleGroup != nil, d.RootCharacterStyleGroup),
		namedKind("RootParagraphStyleGroup", d.RootParagraphStyleGroup != nil, d.RootParagraphStyleGroup),
		namedKind("RootObjectStyleGroup", d.RootObjectStyleGroup != nil, d.RootObjectStyleGroup),
		ptrKind("TinDocumentDataObject", d.TinDocumentDataObject != nil, d.TinDocumentDataObject),
		ptrKind("TransparencyDefaultContainerObject",
			d.TransparencyDefaultContainerObject != nil,
			d.TransparencyDefaultContainerObject),
		sliceKind("Spread", len(d.InlineSpreads), func(i int) any { return d.InlineSpreads[i] }),
		sliceKind("Story", len(d.InlineStories), func(i int) any { return d.InlineStories[i] }),

		sliceKind(kindOther, len(d.OtherElements), func(i int) any { return d.OtherElements[i] }),
	}
	return common.EncodeChildren(e, d.childOrder, kinds)
}

// namedKind is ptrKind for children whose Go type carries no XMLName tag and
// whose element name therefore has to be given explicitly (the Root*StyleGroup
// elements reuse the group types used for nested groups).
func namedKind(name string, present bool, v any) common.ChildKind {
	return common.OptionalKind(name, present, func(e *xml.Encoder) error {
		return wrapMarshal(e.EncodeElement(v, xml.StartElement{Name: xml.Name{Local: name}}))
	})
}
