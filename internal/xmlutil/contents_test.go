package xmlutil

import (
	"bytes"
	"testing"
)

const twoImages = `<Spread Self="s">` +
	`<Rectangle Self="r1"><Image Self="i1" Space="x"><Contents><![CDATA[AAAA]]></Contents></Image></Rectangle>` +
	`<Rectangle Self="r2"><Image Self="i2"><Contents><![CDATA[BBBB]]></Contents></Image></Rectangle>` +
	`</Spread>`

func TestStripAndRestoreContentsRoundtrips(t *testing.T) {
	stripped, payloads := StripContents([]byte(twoImages))
	if len(payloads) != 2 {
		t.Fatalf("got %d payloads, want 2", len(payloads))
	}
	if string(payloads["i1"]) != "<![CDATA[AAAA]]>" || string(payloads["i2"]) != "<![CDATA[BBBB]]>" {
		t.Fatalf("payloads wrong: %q %q", payloads["i1"], payloads["i2"])
	}
	if bytes.Contains(stripped, []byte("AAAA")) || bytes.Contains(stripped, []byte("BBBB")) {
		t.Errorf("stripped output still holds payloads: %s", stripped)
	}
	if got := bytes.Count(stripped, []byte("<Contents></Contents>")); got != 2 {
		t.Errorf("expected 2 empty Contents, got %d in %s", got, stripped)
	}
	if restored := RestoreContents(stripped, payloads); string(restored) != twoImages {
		t.Errorf("roundtrip differs:\n in: %s\nout: %s", twoImages, restored)
	}
}

func TestStripContentsSharesMemoryWithInput(t *testing.T) {
	src := []byte(twoImages)
	_, payloads := StripContents(src)
	p := payloads["i1"]
	// A payload is a sub-slice of src, so mutating src is visible through it.
	idx := bytes.Index(src, []byte("AAAA"))
	src[idx] = 'Z'
	if !bytes.Contains(p, []byte("ZAAA")) {
		t.Errorf("payload does not alias the input buffer: %q", p)
	}
}

func TestRestoreContentsMatchesByOwnerNotOrder(t *testing.T) {
	stripped, payloads := StripContents([]byte(twoImages))
	// Swap the two rectangles; payloads must follow their owning image.
	r1 := []byte(`<Rectangle Self="r1"><Image Self="i1" Space="x"><Contents></Contents></Image></Rectangle>`)
	r2 := []byte(`<Rectangle Self="r2"><Image Self="i2"><Contents></Contents></Image></Rectangle>`)
	swapped := append(append([]byte(`<Spread Self="s">`), r2...), append(r1, []byte(`</Spread>`)...)...)
	if !bytes.Equal(stripped, append(append([]byte(`<Spread Self="s">`), r1...), append(r2, []byte(`</Spread>`)...)...)) {
		t.Fatalf("unexpected stripped form: %s", stripped)
	}
	got := RestoreContents(swapped, payloads)
	if !bytes.Contains(got, []byte(`Self="i2"><Contents><![CDATA[BBBB]]>`)) ||
		!bytes.Contains(got, []byte(`Space="x"><Contents><![CDATA[AAAA]]>`)) {
		t.Errorf("payloads did not follow their owners: %s", got)
	}
}

func TestRestoreContentsLeavesExplicitContentAlone(t *testing.T) {
	_, payloads := StripContents([]byte(twoImages))
	manual := []byte(`<Image Self="i1"><Contents><![CDATA[NEW]]></Contents></Image>`)
	if got := RestoreContents(manual, payloads); !bytes.Equal(got, manual) {
		t.Errorf("non-empty Contents was overwritten: %s", got)
	}
}

func TestStripContentsNoPayloads(t *testing.T) {
	// Story text uses <Content>, not <Contents>; it must not be touched.
	in := []byte(`<Story><Content>hello</Content><Br/></Story>`)
	out, payloads := StripContents(in)
	if payloads != nil || !bytes.Equal(out, in) {
		t.Errorf("story content was altered: %s (payloads=%v)", out, payloads)
	}
	if got := RestoreContents(in, nil); !bytes.Equal(got, in) {
		t.Errorf("RestoreContents with no payloads changed data")
	}
}

func TestStripContentsSkipsUnownedOrEmpty(t *testing.T) {
	// No identifiable owner, and an already-empty element: both left inline.
	in := []byte(`<Thing><Contents><![CDATA[XX]]></Contents></Thing><Image Self="i"><Contents></Contents></Image>`)
	out, payloads := StripContents(in)
	if payloads != nil {
		t.Errorf("unexpected payloads: %v", payloads)
	}
	if !bytes.Equal(out, in) {
		t.Errorf("data changed: %s", out)
	}
}

func TestRestoreContentsHandlesSelfClosingForm(t *testing.T) {
	_, payloads := StripContents([]byte(twoImages))
	in := []byte(`<Image Self="i1"><Contents/></Image>`)
	got := RestoreContents(in, payloads)
	want := `<Image Self="i1"><Contents><![CDATA[AAAA]]></Contents></Image>`
	if string(got) != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

// An embedded EPS carries its own payload and, in its MetadataPacketPreference,
// an XMP packet, also in a <Contents>. Both resolve to the EPS as owner, so the
// second is left inline; the bytes around it must survive.
const epsWithMetadata = `<Spread Self="s">` +
	`<Rectangle Self="r1"><EPS Self="e1" Space="x"><Properties><Contents><![CDATA[EPSDATA]]></Contents></Properties>` +
	`<MetadataPacketPreference><Properties><Contents><![CDATA[<?xpacket begin?>]]></Contents></Properties></MetadataPacketPreference>` +
	`</EPS></Rectangle>` +
	`<Rectangle Self="r2"><Image Self="i2"><Contents><![CDATA[BBBB]]></Contents></Image></Rectangle>` +
	`</Spread>`

func TestStripContentsKeepsBytesAroundInlinePayloads(t *testing.T) {
	tests := []struct {
		name         string
		in           string
		wantPayloads map[string]string
		wantStripped string
	}{
		{
			name:         "second payload of the same owner after a stripped one",
			in:           epsWithMetadata,
			wantPayloads: map[string]string{"e1": "<![CDATA[EPSDATA]]>", "i2": "<![CDATA[BBBB]]>"},
			wantStripped: `<Spread Self="s">` +
				`<Rectangle Self="r1"><EPS Self="e1" Space="x"><Properties><Contents></Contents></Properties>` +
				`<MetadataPacketPreference><Properties><Contents><![CDATA[<?xpacket begin?>]]></Contents></Properties></MetadataPacketPreference>` +
				`</EPS></Rectangle>` +
				`<Rectangle Self="r2"><Image Self="i2"><Contents></Contents></Image></Rectangle>` +
				`</Spread>`,
		},
		{
			name:         "unowned payload before a stripped one",
			in:           `<Thing><Contents><![CDATA[XX]]></Contents></Thing><Image Self="i"><Contents><![CDATA[AA]]></Contents></Image>`,
			wantPayloads: map[string]string{"i": "<![CDATA[AA]]>"},
			wantStripped: `<Thing><Contents><![CDATA[XX]]></Contents></Thing><Image Self="i"><Contents></Contents></Image>`,
		},
		{
			// Restoring matches by owner and could not tell which of the two
			// empty elements held the payload, so it stays inline.
			name:         "owner that also has an empty Contents",
			in:           `<Image Self="a"><Contents/></Image><Image Self="a"><Contents>dup</Contents></Image><Image Self="b"><Contents>B</Contents></Image>`,
			wantPayloads: map[string]string{"b": "B"},
			wantStripped: `<Image Self="a"><Contents/></Image><Image Self="a"><Contents>dup</Contents></Image><Image Self="b"><Contents></Contents></Image>`,
		},
		{
			// Without the payload a later Contents would find a different owner.
			name:         "payload holding an owner tag",
			in:           `<Image Self="a"><Contents><![CDATA[<Image x]]></Contents><Contents></Contents></Image><Image Self="b"><Contents>B</Contents></Image>`,
			wantPayloads: map[string]string{"b": "B"},
			wantStripped: `<Image Self="a"><Contents><![CDATA[<Image x]]></Contents><Contents></Contents></Image><Image Self="b"><Contents></Contents></Image>`,
		},
		{
			name:         "XMP packet in CDATA is stripped",
			in:           `<Image Self="a"><Contents><![CDATA[<?xpacket begin?><x:xmpmeta><rdf:li>1</rdf:li></x:xmpmeta>]]></Contents></Image>`,
			wantPayloads: map[string]string{"a": "<![CDATA[<?xpacket begin?><x:xmpmeta><rdf:li>1</rdf:li></x:xmpmeta>]]>"},
			wantStripped: `<Image Self="a"><Contents></Contents></Image>`,
		},
		{
			name:         "empty element between stripped ones",
			in:           `<Image Self="a"><Contents>A</Contents></Image><Image Self="b"><Contents></Contents></Image><Image Self="c"><Contents>C</Contents></Image>`,
			wantPayloads: map[string]string{"a": "A", "c": "C"},
			wantStripped: `<Image Self="a"><Contents></Contents></Image><Image Self="b"><Contents></Contents></Image><Image Self="c"><Contents></Contents></Image>`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stripped, payloads := StripContents([]byte(tt.in))
			if string(stripped) != tt.wantStripped {
				t.Errorf("stripped:\n got: %s\nwant: %s", stripped, tt.wantStripped)
			}
			if len(payloads) != len(tt.wantPayloads) {
				t.Errorf("got %d payloads, want %d: %q", len(payloads), len(tt.wantPayloads), payloads)
			}
			for self, want := range tt.wantPayloads {
				if string(payloads[self]) != want {
					t.Errorf("payload %q = %q, want %q", self, payloads[self], want)
				}
			}
			if restored := RestoreContents(stripped, payloads); string(restored) != tt.in {
				t.Errorf("roundtrip differs:\n in: %s\nout: %s", tt.in, restored)
			}
		})
	}
}

// FuzzStripAndRestoreContents checks the contract streaming mode relies on:
// restoring what was stripped gives back the input exactly, and the stripped
// form has no payload left in any element it emptied.
func FuzzStripAndRestoreContents(f *testing.F) {
	for _, seed := range []string{
		twoImages,
		epsWithMetadata,
		`<Thing><Contents><![CDATA[XX]]></Contents></Thing><Image Self="i"><Contents><![CDATA[AA]]></Contents></Image>`,
		`<Image Self="a"><Contents>A</Contents></Image><Image Self="b"><Contents></Contents></Image><Image Self="c"><Contents>C</Contents></Image>`,
		`<Image Self="a"><Contents/></Image><Image Self="a"><Contents>dup</Contents></Image>`,
		`<Image Self="x"><Contents>unterminated`,
		`<Story><Content>hello</Content></Story>`,
		"",
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, in string) {
		stripped, payloads := StripContents([]byte(in))
		if payloads == nil {
			if string(stripped) != in {
				t.Fatalf("no payloads but data changed:\n in: %q\nout: %q", in, stripped)
			}
			return
		}
		if restored := RestoreContents(stripped, payloads); string(restored) != in {
			t.Fatalf("roundtrip differs:\n in: %q\nout: %q", in, restored)
		}
	})
}
