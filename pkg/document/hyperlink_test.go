package document

import (
	"slices"
	"strings"
	"testing"
)

// hyperlinkDesignmap is trimmed from a newspaper page whose crossword frame
// (u78df) links to the online puzzle. The second destination is not used by
// any hyperlink, as in the original.
const hyperlinkDesignmap = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Document xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging" DOMVersion="21.0" Self="d">
	<idPkg:Story src="Stories/Story_u7a10.xml" />
	<HyperlinkURLDestination Self="HyperlinkURLDestination/https%3a//www.puzzleplayz.com/c/ntm_storakrysset" Name="https://www.puzzleplayz.com/c/ntm_storakrysset" DestinationURL="https://www.puzzleplayz.com/c/ntm_storakrysset" Hidden="false" DestinationUniqueKey="1" />
	<HyperlinkURLDestination Self="HyperlinkURLDestination/https%3a//www.puzzleplayz.com/x/ntmrikskryss" Name="https://www.puzzleplayz.com/x/ntmrikskryss" DestinationURL="https://www.puzzleplayz.com/x/ntmrikskryss" Hidden="false" DestinationUniqueKey="2" />
	<HyperlinkPageItemSource Self="u78e1" Name="Hyperlink 1" SourcePageItem="u78df" Hidden="false" />
	<Hyperlink Self="u78e3" Name="https://www.puzzleplayz.com/c/ntm_storakrysset_36708" Source="u78e1" Visible="false" Highlight="None" Width="Thin" BorderStyle="Solid" Hidden="false" DestinationUniqueKey="1" EpubAriaRole="" HypherlinkAltText="">
		<Properties>
			<BorderColor type="enumeration">Black</BorderColor>
			<Destination type="object">HyperlinkURLDestination/https%3a//www.puzzleplayz.com/c/ntm_storakrysset</Destination>
		</Properties>
	</Hyperlink>
	<ColorGroup Self="ColorGroup/[Root Color Group]" Name="[Root Color Group]" IsRootColorGroup="true" />
</Document>`

func TestParseHyperlinks(t *testing.T) {
	doc, err := ParseDocument([]byte(hyperlinkDesignmap))
	if err != nil {
		t.Fatalf("ParseDocument() error = %v", err)
	}

	if len(doc.HyperlinkURLDestinations) != 2 || len(doc.HyperlinkPageItemSources) != 1 || len(doc.Hyperlinks) != 1 {
		t.Fatalf("got %d URL destinations, %d page item sources, %d hyperlinks, want 2, 1, 1",
			len(doc.HyperlinkURLDestinations), len(doc.HyperlinkPageItemSources), len(doc.Hyperlinks))
	}

	for _, el := range doc.OtherElements {
		if strings.HasPrefix(el.XMLName.Local, "Hyperlink") {
			t.Errorf("%s left in OtherElements", el.XMLName.Local)
		}
	}

	h := doc.Hyperlinks[0]
	if h.Self != "u78e3" || h.Source != "u78e1" || h.DestinationUniqueKey != "1" || h.Visible != "false" {
		t.Errorf("Hyperlink = %+v", h)
	}

	if got, want := h.Destination(), "HyperlinkURLDestination/https%3a//www.puzzleplayz.com/c/ntm_storakrysset"; got != want {
		t.Errorf("Destination() = %q, want %q", got, want)
	}

	if src := doc.HyperlinkPageItemSources[0]; src.SourcePageItem != "u78df" {
		t.Errorf("SourcePageItem = %q, want u78df", src.SourcePageItem)
	}

	if dest := doc.HyperlinkURLDestinations[1]; dest.DestinationURL != "https://www.puzzleplayz.com/x/ntmrikskryss" ||
		dest.DestinationUniqueKey != "2" {
		t.Errorf("second destination = %+v", dest)
	}
}

func TestPageItemURLs(t *testing.T) {
	doc, err := ParseDocument([]byte(hyperlinkDesignmap))
	if err != nil {
		t.Fatalf("ParseDocument() error = %v", err)
	}

	if got, want := doc.PageItemURLs("u78df"), []string{"https://www.puzzleplayz.com/c/ntm_storakrysset"}; !slices.Equal(got, want) {
		t.Errorf("PageItemURLs(u78df) = %q, want %q", got, want)
	}

	if got := doc.PageItemURLs("u78f9"); got != nil {
		t.Errorf("PageItemURLs(u78f9) = %q, want none for a frame without a hyperlink", got)
	}

	if got := doc.PageItemHyperlinks("u78df"); len(got) != 1 || got[0].Self != "u78e3" {
		t.Errorf("PageItemHyperlinks(u78df) = %+v, want u78e3", got)
	}
}

func TestHyperlinkURLResolution(t *testing.T) {
	doc := &Document{
		HyperlinkURLDestinations: []HyperlinkURLDestination{
			{Self: "HyperlinkURLDestination/URL 1", DestinationURL: "https://example.com/1", DestinationUniqueKey: "1"},
			{Self: "HyperlinkURLDestination/URL 2", DestinationURL: "https://example.com/2", DestinationUniqueKey: "2"},
		},
	}

	tests := []struct {
		name   string
		link   Hyperlink
		want   string
		wantOK bool
	}{
		{
			name:   "by the destination in Properties",
			link:   withDestination(Hyperlink{DestinationUniqueKey: "1"}, "HyperlinkURLDestination/URL 2"),
			want:   "https://example.com/2",
			wantOK: true,
		},
		{
			name:   "by DestinationUniqueKey without Properties",
			link:   Hyperlink{DestinationUniqueKey: "1"},
			want:   "https://example.com/1",
			wantOK: true,
		},
		{
			name: "page destination is not a URL",
			link: withDestination(Hyperlink{DestinationUniqueKey: "1"}, "HyperlinkPageDestination/Sida 1"),
		},
		{
			name: "unknown key",
			link: Hyperlink{DestinationUniqueKey: "9"},
		},
		{
			name: "no destination at all",
			link: Hyperlink{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := doc.HyperlinkURL(tt.link)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("HyperlinkURL() = %q, %v, want %q, %v", got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

func withDestination(h Hyperlink, dest string) Hyperlink {
	doc, err := ParseDocument([]byte(`<Document Self="d"><Hyperlink Self="h" Name="h"><Properties>` +
		`<Destination type="object">` + dest + `</Destination></Properties></Hyperlink></Document>`))
	if err != nil {
		panic(err)
	}

	h.Properties = doc.Hyperlinks[0].Properties

	return h
}

func TestHyperlinksRoundtrip(t *testing.T) {
	doc, err := ParseDocument([]byte(hyperlinkDesignmap))
	if err != nil {
		t.Fatalf("ParseDocument() error = %v", err)
	}

	out, err := MarshalDocument(doc)
	if err != nil {
		t.Fatalf("MarshalDocument() error = %v", err)
	}

	xml := string(out)

	// Empty attributes InDesign writes must survive, not be dropped.
	for _, want := range []string{`EpubAriaRole=""`, `HypherlinkAltText=""`, `<BorderColor type="enumeration">Black</BorderColor>`} {
		if !strings.Contains(xml, want) {
			t.Errorf("output lacks %s:\n%s", want, xml)
		}
	}

	assertOrder(t, xml, "Stories/Story_u7a10.xml", "<HyperlinkURLDestination", "ntmrikskryss", "<HyperlinkPageItemSource", "<Hyperlink ", "<ColorGroup")

	again, err := ParseDocument(out)
	if err != nil {
		t.Fatalf("ParseDocument(output) error = %v", err)
	}

	if got := again.PageItemURLs("u78df"); len(got) != 1 || got[0] != "https://www.puzzleplayz.com/c/ntm_storakrysset" {
		t.Errorf("PageItemURLs after roundtrip = %q", got)
	}
}

// TestHyperlinksOrderInBuiltDocument checks where hyperlinks go in a document
// built in code: after the sections, document users and stories and before
// the color groups, as InDesign writes them.
func TestHyperlinksOrderInBuiltDocument(t *testing.T) {
	doc := &Document{
		Self:                     "d",
		Sections:                 []Section{{Self: "ub4", Name: "A"}},
		ColorGroups:              []ColorGroup{{Self: "ColorGroup/[Root Color Group]", Name: "[Root Color Group]"}},
		Hyperlinks:               []Hyperlink{{Self: "h1", Name: "Link", Source: "s1"}},
		HyperlinkPageItemSources: []HyperlinkPageItemSource{{Self: "s1", Name: "Hyperlink 1", SourcePageItem: "u1"}},
		HyperlinkURLDestinations: []HyperlinkURLDestination{{Self: "HyperlinkURLDestination/u", Name: "u", DestinationURL: "https://example.com"}},
	}

	out, err := MarshalDocument(doc)
	if err != nil {
		t.Fatalf("MarshalDocument() error = %v", err)
	}

	assertOrder(t, string(out), "<Section", "<HyperlinkURLDestination", "<HyperlinkPageItemSource", "<Hyperlink ", "<ColorGroup")
}

func assertOrder(t *testing.T, s string, parts ...string) {
	t.Helper()

	pos := 0

	for _, p := range parts {
		i := strings.Index(s[pos:], p)
		if i < 0 {
			t.Fatalf("%q missing or out of order in:\n%s", p, s)
		}

		pos += i + len(p)
	}
}
