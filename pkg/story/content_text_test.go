package story

import (
	"encoding/xml"
	"testing"
)

func TestContentPreservesProcessingInstructions(t *testing.T) {
	in := `<Content><?ACE 4?>	Footnote text.<?ACE 7?>more</Content>`
	var c Content
	if err := xml.Unmarshal([]byte(in), &c); err != nil {
		t.Fatal(err)
	}
	if c.Text != "\tFootnote text.more" {
		t.Errorf("Text = %q", c.Text)
	}
	if !c.HasInstructions() || len(c.Segments) != 4 {
		t.Fatalf("expected 4 segments with instructions, got %+v", c.Segments)
	}
	out, err := xml.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != `<Content><?ACE 4?>&#x9;Footnote text.<?ACE 7?>more</Content>` { // the encoder escapes the tab; semantically identical
		t.Errorf("roundtrip changed content:\n in: %s\nout: %s", in, out)
	}

	c.SetText("plain")
	out, _ = xml.Marshal(c)
	if string(out) != `<Content>plain</Content>` {
		t.Errorf("SetText output = %s", out)
	}
}

func TestContentWithoutInstructionsHasNoSegments(t *testing.T) {
	var c Content
	if err := xml.Unmarshal([]byte(`<Content>Hello &amp; bye</Content>`), &c); err != nil {
		t.Fatal(err)
	}
	if c.Text != "Hello & bye" || c.Segments != nil {
		t.Errorf("Text=%q Segments=%v", c.Text, c.Segments)
	}
	out, _ := xml.Marshal(c)
	if string(out) != `<Content>Hello &amp; bye</Content>` {
		t.Errorf("out = %s", out)
	}
}
