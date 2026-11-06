package idml

import (
	"strings"
	"testing"
)

// taggedBackingStory is a backing story with a tagged structure under the root, as documents from an XML
// workflow have: elements with attributes, pointing at stories that may no longer exist.
const taggedBackingStory = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<idPkg:BackingStory xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging" DOMVersion="21.2">
	<XmlStory Self="u97" UserText="true" IsEndnoteStory="false" AppliedTOCStyle="n" TrackChanges="false" StoryTitle="$ID/" AppliedNamedGrid="n">
		<ParagraphStyleRange AppliedParagraphStyle="ParagraphStyle/$ID/NormalParagraphStyle">
			<CharacterStyleRange AppliedCharacterStyle="CharacterStyle/$ID/[No character style]">
				<XMLElement Self="di3" MarkupTag="XMLTag/Root">
					<XMLElement Self="di3i4" MarkupTag="XMLTag/Story" XMLContent="u2ed">
						<XMLAttribute Self="di3i4a" Name="ElementID" Value="144006385" />
						<XMLElement Self="di3i4i1" MarkupTag="XMLTag/Text" XMLContent="u2f2" />
					</XMLElement>
					<XMLElement Self="di3i5" MarkupTag="XMLTag/Story" XMLContent="u6f6" />
				</XMLElement>
			</CharacterStyleRange>
		</ParagraphStyleRange>
	</XmlStory>
</idPkg:BackingStory>
`

func TestResetXMLStructureKeepsOnlyTheRoot(t *testing.T) {
	pkg := loadPlainIDML(t)
	pkg.setFileData(PathBackingStory, []byte(taggedBackingStory))

	removed, err := pkg.ResetXMLStructure()
	if err != nil {
		t.Fatal(err)
	}
	if removed != 3 { // di3i4, di3i4i1 and di3i5; the root stays
		t.Errorf("removed %d elements, want 3", removed)
	}

	pkg2 := reread(t, pkg, "xmlreset.idml")
	got := string(mustFileData(t, pkg2, PathBackingStory))
	for _, want := range []string{`<XmlStory Self="u97"`, `<XMLElement Self="di3" MarkupTag="XMLTag/Root" />`, `DOMVersion="21.2"`} {
		if !strings.Contains(got, want) {
			t.Errorf("backing story lacks %s:\n%s", want, got)
		}
	}
	for _, gone := range []string{"XMLContent", "XMLAttribute", "XMLTag/Story", "di3i4"} {
		if strings.Contains(got, gone) {
			t.Errorf("backing story still has %s:\n%s", gone, got)
		}
	}
	if n, err := pkg2.ResetXMLStructure(); err != nil || n != 0 {
		t.Errorf("a second reset removed %d (err %v), want 0", n, err)
	}
}

func TestResetXMLStructureLeavesAnEmptyStructureAlone(t *testing.T) {
	pkg := loadPlainIDML(t)
	before := string(mustFileData(t, pkg, PathBackingStory))
	removed, err := pkg.ResetXMLStructure()
	if err != nil || removed != 0 {
		t.Fatalf("removed %d, err %v; want 0, nil", removed, err)
	}
	if pkg.IsModified(PathBackingStory) {
		t.Error("an empty structure should not be marked modified")
	}
	if after := string(mustFileData(t, pkg, PathBackingStory)); after != before {
		t.Error("an empty structure was rewritten")
	}
}

func TestResetXMLStructureRejectsBrokenBackingStory(t *testing.T) {
	pkg := loadPlainIDML(t)
	pkg.setFileData(PathBackingStory, []byte(`<idPkg:BackingStory xmlns:idPkg="x"><XmlStory Self="u1"/></idPkg:BackingStory>`))
	if _, err := pkg.ResetXMLStructure(); err == nil {
		t.Error("a backing story without a root element should fail")
	}
}
