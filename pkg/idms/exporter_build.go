package idms

import (
	"encoding/xml"

	"github.com/dimelords/idmllib/v2/pkg/common"
	"github.com/dimelords/idmllib/v2/pkg/document"
	"github.com/dimelords/idmllib/v2/pkg/idml"
	"github.com/dimelords/idmllib/v2/pkg/resources"
	"github.com/dimelords/idmllib/v2/pkg/spread"
	"github.com/dimelords/idmllib/v2/pkg/story"
)

// ============================================================================
// Phase 4.3: Build Minimal IDMS Package
// ============================================================================

// buildMinimalPackage constructs a complete IDMS package from the selection and extracted resources.
// IDMS uses an inline structure where all resources are embedded directly in the Document.
func (e *Exporter) buildMinimalPackage(sel *idml.Selection, resources *ExtractedResources) (*Package, error) {
	// Create new IDMS package
	pkg := New()

	// Set up processing instructions with full AID metadata
	pkg.SetDefaultAIDProcessingInstructions("PageItem")

	// Create minimal Document with inline resources
	doc := &document.Document{
		Self:       "d",
		DOMVersion: "20.4",
	}

	// Add required defaults (Black color, None swatch, Solid stroke)
	doc.Colors = e.buildDefaultColors()
	doc.Swatches = e.buildDefaultSwatches()
	doc.StrokeStyles = e.buildDefaultStrokeStyles()

	// Add extracted colors and swatches
	if resources.Graphics != nil {
		doc.Colors = append(doc.Colors, resources.Graphics.Colors...)
		doc.Swatches = append(doc.Swatches, resources.Graphics.Swatches...)
	}

	// Add style groups
	if resources.Styles != nil {
		doc.RootCharacterStyleGroup = resources.Styles.RootCharacterStyleGroup
		doc.RootParagraphStyleGroup = resources.Styles.RootParagraphStyleGroup
		doc.RootObjectStyleGroup = resources.Styles.RootObjectStyleGroup
	} else {
		// Add minimal default styles
		doc.RootCharacterStyleGroup = e.buildDefaultCharacterStyleGroup()
		doc.RootParagraphStyleGroup = e.buildDefaultParagraphStyleGroup()
		doc.RootObjectStyleGroup = e.buildDefaultObjectStyleGroup()
	}

	// Add default NumberingList (required for InDesign compatibility)
	doc.NumberingLists = e.buildDefaultNumberingLists()

	// Add TinDocumentDataObject and TransparencyDefaultContainerObject
	doc.TinDocumentDataObject = e.buildTinDocumentDataObject()
	doc.TransparencyDefaultContainerObject = e.buildTransparencyDefaults()

	// Add default Layer (required for IDMS)
	doc.Layers = e.buildDefaultLayers()

	// Add ColorGroup (required for IDMS)
	doc.ColorGroups = e.buildDefaultColorGroups()

	// Add spreads with selected items (IDMS uses inline spreads, not ResourceRef)
	spreadElem := e.buildSpreadFromSelection(sel)
	doc.InlineSpreads = []spread.SpreadElement{spreadElem}

	// Add stories inline (IDMS uses inline stories, not ResourceRef)
	if len(resources.Stories) > 0 {
		doc.InlineStories = make([]story.StoryElement, 0, len(resources.Stories))
		for _, st := range resources.Stories {
			doc.InlineStories = append(doc.InlineStories, st.StoryElement)
		}
	}

	// Add XMP metadata
	pkg.XMPMetadata = e.generateXMPMetadata()

	pkg.Document = doc

	return pkg, nil
}

// buildDefaultColors returns required default colors (Black).
func (e *Exporter) buildDefaultColors() []resources.Color {
	return []resources.Color{
		{
			Self:                      "Color/Black",
			Model:                     "Process",
			Space:                     "CMYK",
			ColorValue:                "0 0 0 100",
			ColorOverride:             "Specialblack",
			ConvertToHsb:              "false",
			AlternateSpace:            "NoAlternateColor",
			AlternateColorValue:       "",
			Name:                      "Black",
			ColorEditable:             "false",
			ColorRemovable:            "false",
			Visible:                   "true",
			SwatchCreatorID:           "7937",
			SwatchColorGroupReference: "u12ColorGroupSwatch3",
		},
	}
}

// buildDefaultSwatches returns required default swatches (None).
func (e *Exporter) buildDefaultSwatches() []resources.Swatch {
	return []resources.Swatch{
		{
			Self:                      "Swatch/None",
			Name:                      "None",
			ColorEditable:             "false",
			ColorRemovable:            "false",
			Visible:                   "true",
			SwatchCreatorID:           "7937",
			SwatchColorGroupReference: "u12ColorGroupSwatch0",
		},
	}
}

// buildDefaultStrokeStyles returns required default stroke style (Solid).
func (e *Exporter) buildDefaultStrokeStyles() []resources.StrokeStyle {
	return []resources.StrokeStyle{
		{
			Self: "StrokeStyle/$ID/Solid",
			Name: "$ID/Solid",
		},
	}
}

// buildDefaultCharacterStyleGroup returns minimal character style group.
func (e *Exporter) buildDefaultCharacterStyleGroup() *resources.CharacterStyleGroup {
	csg := &resources.CharacterStyleGroup{
		Self: "u7a",
		CharacterStyles: []resources.CharacterStyle{
			{
				Self:                     "CharacterStyle/$ID/[No character style]",
				Name:                     "$ID/[No character style]",
				Imported:                 "false",
				SplitDocument:            "false",
				EmitCss:                  "true",
				IncludeClass:             "true",
				ExtendedKeyboardShortcut: "0 0 0",
			},
		},
	}
	// Set XMLName explicitly to use "RootCharacterStyleGroup" tag
	csg.XMLName = xml.Name{Local: "RootCharacterStyleGroup"}
	return csg
}

// buildDefaultParagraphStyleGroup returns minimal paragraph style group.
func (e *Exporter) buildDefaultParagraphStyleGroup() *resources.ParagraphStyleGroup {
	psg := &resources.ParagraphStyleGroup{
		Self: "u79",
		ParagraphStyles: []resources.ParagraphStyle{
			{
				Self:                     "ParagraphStyle/$ID/NormalParagraphStyle",
				Name:                     "$ID/NormalParagraphStyle",
				Imported:                 "false",
				NextStyle:                "ParagraphStyle/$ID/NormalParagraphStyle",
				SplitDocument:            "false",
				EmitCss:                  "true",
				IncludeClass:             "true",
				ExtendedKeyboardShortcut: "0 0 0",
			},
		},
	}
	// Set XMLName explicitly to use "RootParagraphStyleGroup" tag
	psg.XMLName = xml.Name{Local: "RootParagraphStyleGroup"}
	return psg
}

// buildDefaultObjectStyleGroup returns minimal object style group.
func (e *Exporter) buildDefaultObjectStyleGroup() *resources.ObjectStyleGroup {
	osg := &resources.ObjectStyleGroup{
		Self: "u8a",
		ObjectStyles: []resources.ObjectStyle{
			{
				Self: "ObjectStyle/$ID/[Normal Text Frame]",
				Name: "$ID/[Normal Text Frame]",
			},
		},
	}
	// Set XMLName explicitly to use "RootObjectStyleGroup" tag
	osg.XMLName = xml.Name{Local: "RootObjectStyleGroup"}
	return osg
}

// buildDefaultNumberingLists returns the default NumberingList definition.
// This is required for InDesign compatibility with paragraph styles.
func (e *Exporter) buildDefaultNumberingLists() []document.NumberingList {
	return []document.NumberingList{
		{
			Self:                           "NumberingList/$ID/[Default]",
			Name:                           "$ID/[Default]",
			ContinueNumbersAcrossStories:   "false",
			ContinueNumbersAcrossDocuments: "false",
		},
	}
}

// buildTinDocumentDataObject returns the TinDocumentDataObject element.
// This is required for InDesign compatibility.
func (e *Exporter) buildTinDocumentDataObject() *document.TinDocumentDataObject {
	return &document.TinDocumentDataObject{}
}

// buildTransparencyDefaults returns transparency default settings.
// This is required for InDesign compatibility.
func (e *Exporter) buildTransparencyDefaults() *document.TransparencyDefaultContainerObject {
	return &document.TransparencyDefaultContainerObject{}
}

// buildSpreadFromSelection creates a spread containing the selected page items.
func (e *Exporter) buildSpreadFromSelection(sel *idml.Selection) spread.SpreadElement {
	spreadElem := spread.SpreadElement{
		Self: "ue6",
	}

	// Copy selected text frames
	if len(sel.TextFrames) > 0 {
		spreadElem.TextFrames = make([]spread.SpreadTextFrame, len(sel.TextFrames))
		for i, tf := range sel.TextFrames {
			spreadElem.TextFrames[i] = *tf
		}
	}

	// Copy selected rectangles
	if len(sel.Rectangles) > 0 {
		spreadElem.Rectangles = make([]spread.Rectangle, len(sel.Rectangles))
		for i, rect := range sel.Rectangles {
			spreadElem.Rectangles[i] = *rect
		}
	}

	// Copy selected ovals
	if len(sel.Ovals) > 0 {
		spreadElem.Ovals = make([]spread.Oval, len(sel.Ovals))
		for i, oval := range sel.Ovals {
			spreadElem.Ovals[i] = *oval
		}
	}

	// Copy selected polygons
	if len(sel.Polygons) > 0 {
		spreadElem.Polygons = make([]spread.Polygon, len(sel.Polygons))
		for i, polygon := range sel.Polygons {
			spreadElem.Polygons[i] = *polygon
		}
	}

	// Copy selected graphic lines
	if len(sel.GraphicLines) > 0 {
		spreadElem.GraphicLines = make([]spread.GraphicLine, len(sel.GraphicLines))
		for i, line := range sel.GraphicLines {
			spreadElem.GraphicLines[i] = *line
		}
	}

	// Copy selected groups
	if len(sel.Groups) > 0 {
		spreadElem.Groups = make([]spread.Group, len(sel.Groups))
		for i, group := range sel.Groups {
			spreadElem.Groups[i] = *group
		}
	}

	return spreadElem
}

// buildDefaultLayers returns the default Layer definition required by InDesign.
func (e *Exporter) buildDefaultLayers() []document.Layer {
	return []document.Layer{
		{
			Self:       "uba",
			Name:       "Layer 1",
			Visible:    "true",
			Locked:     "false",
			IgnoreWrap: "false",
			ShowGuides: "true",
			LockGuides: "false",
			UI:         "true",
			Expendable: "true",
			Printable:  "true",
			Properties: &common.Properties{
				OtherElements: []common.RawXMLElement{
					{
						XMLName: xml.Name{Local: "LayerColor"},
						Attrs: []xml.Attr{
							{Name: xml.Name{Local: "type"}, Value: "enumeration"},
						},
						Content: []byte("LightBlue"),
					},
				},
			},
		},
	}
}

// buildDefaultColorGroups returns the ColorGroup definition required by InDesign.
func (e *Exporter) buildDefaultColorGroups() []document.ColorGroup {
	return []document.ColorGroup{
		{
			Self:             "ColorGroup/[Root Color Group]",
			Name:             "[Root Color Group]",
			IsRootColorGroup: "true",
			ColorGroupSwatches: []document.ColorGroupSwatch{
				{
					Self:          "u12ColorGroupSwatch0",
					SwatchItemRef: "Swatch/None",
				},
				{
					Self:          "u12ColorGroupSwatch3",
					SwatchItemRef: "Color/Black",
				},
			},
		},
	}
}

// generateXMPMetadata generates minimal XMP metadata for the IDMS snippet.
func (e *Exporter) generateXMPMetadata() string {
	return `<?xpacket begin="" id="W5M0MpCehiHzreSzNTczkc9d"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/" x:xmptk="Adobe XMP Core 7.0-c000 1.000000, 0000/00/00-00:00:00">
   <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
   </rdf:RDF>
</x:xmpmeta>
<?xpacket end="r"?>`
}
