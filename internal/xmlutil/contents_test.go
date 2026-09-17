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
