package idml

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/dimelords/idmllib/v3/pkg/common"
	"github.com/dimelords/idmllib/v3/pkg/resources"
)

// fontElements are elements whose text is a font family name, as in
//
//	<AppliedFont type="string">Minion Pro</AppliedFont>
//
// fontAttributes are attributes whose value is one. Both lists were
// derived by scanning the test corpus for every place a name declared in
// Resources/Fonts.xml reappears: these four sites are all of them.
//
// <stFnt:*> is deliberately absent. It occurs only in META-INF/
// metadata.xml, where XMP records the fonts a document happened to use;
// that is a description of the file, not a reference that Fonts.xml has
// to satisfy.
var (
	fontElements = map[string]bool{
		"AppliedFont": true,
		"BulletsFont": true,
	}
	fontAttributes = map[string]bool{
		"WatermarkFontFamily": true,
	}
)

// extractFonts records every font family the document refers to, so that
// findMissingFonts can report families Resources/Fonts.xml never declares
// and findOrphanedFonts does not mistake a font in use for an unused one.
//
// Families are referenced from four files - Styles.xml (paragraph and
// character styles), the stories (fonts applied straight to a run rather
// than through a style), Preferences.xml and designmap.xml (bullet and
// watermark fonts) - and all four are collected here. A gap on this side
// is not harmless: an uncollected reference makes the font look unused,
// and FindOrphans would then invite the caller to delete a font the
// document still renders with.
//
// Every style is scanned, not only the styles the stories apply.
// InDesign declares every family any style names, whether or not text
// currently uses it, so that is what missing-font detection has to
// compare against; for orphan detection it is the conservative
// direction, keeping a family reachable only from an unused style.
func (rm *ResourceManager) extractFonts(deps *dependencySet) error {
	if deps == nil {
		return common.Errorf("idml", "extract fonts", "", "dependency set is nil")
	}

	for _, collect := range []func(*dependencySet) error{
		rm.extractFontsFromStyles,
		rm.extractFontsFromStories,
		rm.extractFontsFromPreferences,
		rm.extractFontsFromDocument,
	} {
		if err := collect(deps); err != nil {
			return err
		}
	}
	return nil
}

// extractFontsFromStyles collects the fonts named by paragraph and
// character style definitions, including styles nested in groups and the
// built-in $ID/ styles - the root style is where InDesign puts a
// document's default font, so skipping it would miss the family most
// text inherits.
func (rm *ResourceManager) extractFontsFromStyles(deps *dependencySet) error {
	styles, err := rm.pkg.Styles()
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			return nil // No styles file, nothing to extract.
		}
		return fmt.Errorf("failed to get styles: %w", err)
	}

	var walkParagraph func(group *resources.ParagraphStyleGroup)
	walkParagraph = func(group *resources.ParagraphStyleGroup) {
		if group == nil {
			return
		}
		for i := range group.ParagraphStyles {
			collectFontsFromProperties(deps, group.ParagraphStyles[i].Properties)
		}
		for i := range group.NestedGroups {
			walkParagraph(&group.NestedGroups[i])
		}
	}
	walkParagraph(styles.RootParagraphStyleGroup)

	var walkCharacter func(group *resources.CharacterStyleGroup)
	walkCharacter = func(group *resources.CharacterStyleGroup) {
		if group == nil {
			return
		}
		for i := range group.CharacterStyles {
			collectFontsFromProperties(deps, group.CharacterStyles[i].Properties)
		}
		for i := range group.NestedGroups {
			walkCharacter(&group.NestedGroups[i])
		}
	}
	walkCharacter(styles.RootCharacterStyleGroup)

	return nil
}

// extractFontsFromStories collects fonts applied directly to text rather
// than through a named style. These arrive as a <Properties> child of a
// style range; neither range type models Properties as a typed field, so
// they land in the raw-element catch-all and are decoded from there.
func (rm *ResourceManager) extractFontsFromStories(deps *dependencySet) error {
	stories, err := rm.pkg.Stories()
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("failed to get stories: %w", err)
	}

	for _, st := range stories {
		if st == nil {
			continue
		}
		for _, psr := range st.Paragraphs() {
			collectFontsFromRaw(deps, psr.OtherElements)
			for _, csr := range psr.Ranges() {
				if csr == nil {
					continue
				}
				for i := range csr.Children {
					if other := csr.Children[i].Other; other != nil {
						collectFontsFromRaw(deps, []common.RawXMLElement{*other})
					}
				}
			}
		}
	}

	return nil
}

// extractFontsFromPreferences collects the document-wide bullet font,
// which lives in Resources/Preferences.xml rather than in any style.
func (rm *ResourceManager) extractFontsFromPreferences(deps *dependencySet) error {
	prefs, err := rm.pkg.Preferences()
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("failed to get preferences: %w", err)
	}
	if prefs != nil {
		collectFontsFromRaw(deps, prefs.Elements)
	}
	return nil
}

// extractFontsFromDocument collects fonts named by designmap.xml itself -
// the watermark font family and the document's own bullet font default.
func (rm *ResourceManager) extractFontsFromDocument(deps *dependencySet) error {
	doc, err := rm.pkg.Document()
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("failed to get document: %w", err)
	}
	if doc == nil {
		return nil
	}
	collectFontsFromProperties(deps, doc.Properties)
	collectFontsFromRaw(deps, doc.OtherElements)
	// Bullet definitions are modeled, so their fonts are not reachable
	// through the catch-all above.
	for i := range doc.ABullets {
		collectFontsFromProperties(deps, doc.ABullets[i].Properties)
		collectFontsFromRaw(deps, doc.ABullets[i].OtherElements)
	}
	return nil
}

// collectFontsFromProperties records the fonts named inside one
// <Properties> element.
func collectFontsFromProperties(deps *dependencySet, props *common.Properties) {
	if props == nil {
		return
	}
	collectFontsFromRaw(deps, props.OtherElements)
}

// collectFontsFromRaw records the fonts named by a set of unmodeled
// elements, descending into their contents. An element may either be a
// font reference itself or merely contain one further down.
func collectFontsFromRaw(deps *dependencySet, elems []common.RawXMLElement) {
	for i := range elems {
		elem := &elems[i]
		for _, attr := range elem.Attrs {
			if fontAttributes[attr.Name.Local] {
				addFont(deps, attr.Value)
			}
		}
		// RawXMLElement.Content is inner XML, so a font-bearing element
		// holds its family name there as plain text, while any other
		// element may hold nested references as markup.
		if fontElements[elem.XMLName.Local] {
			addFont(deps, string(elem.Content))
			continue
		}
		collectFontsFromXML(deps, elem.Content)
	}
}

// collectFontsFromXML records the fonts named anywhere in a fragment of
// XML. The input is an inner-XML fragment, which may hold several
// top-level elements or none, so it is walked token by token.
func collectFontsFromXML(deps *dependencySet, data []byte) {
	if len(data) == 0 {
		return
	}

	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		tok, err := dec.Token()
		if err != nil {
			// Including io.EOF: a fragment simply runs out. Malformed
			// markup stops the scan rather than failing the analysis,
			// which only feeds advisory checks - and the bytes are
			// still written back verbatim either way.
			if !errors.Is(err, io.EOF) {
				return
			}
			return
		}

		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		for _, attr := range start.Attr {
			if fontAttributes[attr.Name.Local] {
				addFont(deps, attr.Value)
			}
		}
		if fontElements[start.Name.Local] {
			var text string
			if err := dec.DecodeElement(&text, &start); err == nil {
				addFont(deps, text)
			}
		}
	}
}

// addFont records a family name, ignoring blank ones and InDesign's
// placeholders. "$ID/" on its own is what InDesign writes for "no font
// chosen" - every fixture carries several - and the "$ID/" prefix
// generally marks a built-in rather than a real family, so neither is a
// reference that Fonts.xml is expected to declare.
func addFont(deps *dependencySet, family string) {
	family = strings.TrimSpace(family)
	if family == "" || strings.HasPrefix(family, "$ID/") {
		return
	}
	deps.fonts[family] = true
}
