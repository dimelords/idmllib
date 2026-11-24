package common

import (
	"encoding/xml"
	"strings"
	"testing"
)

type rawHolder struct {
	XMLName xml.Name        `xml:"Holder"`
	Other   []RawXMLElement `xml:",any"`
}

// An unknown element is kept as it is, prefixed or not, unless its local name isn't an XML name: encoding/xml
// accepts <a:0 xmlns:a="…">, which written back would be <0 xmlns="…">, unreadable. Found by FuzzParseStory.
func TestRawXMLElementName(t *testing.T) {
	for _, c := range []struct {
		in string
		ok bool
	}{
		{`<Holder><Unknown x="1"><Inner/></Unknown></Holder>`, true},
		{`<Holder><a:b xmlns:a="urn:x"/></Holder>`, true},
		{`<Holder><a:b_1.c-d xmlns:a="urn:x"/></Holder>`, true},
		{`<Holder><a:Å xmlns:a="urn:x"/></Holder>`, true},
		{`<Holder><Unknown><a:0/></Unknown></Holder>`, true}, // inside an element kept raw, its bytes are kept as they are
		{`<Holder><a:0 xmlns:a="A00000000000"/></Holder>`, false},
		{`<Holder><a:0/></Holder>`, false},
		{`<Holder><a:-b xmlns:a="urn:x"/></Holder>`, false},
		{`<Holder><a:.b xmlns:a="urn:x"/></Holder>`, false},
	} {
		var h rawHolder
		err := xml.Unmarshal([]byte(c.in), &h)
		if (err == nil) != c.ok {
			t.Errorf("%s: error %v, want ok %v", c.in, err, c.ok)
			continue
		}
		if err != nil {
			if !strings.Contains(err.Error(), "is not an XML name") {
				t.Errorf("%s: error %v", c.in, err)
			}
			continue
		}
		out, err := xml.Marshal(h)
		if err != nil {
			t.Fatalf("%s: marshal: %v", c.in, err)
		}
		var again rawHolder
		if err := xml.Unmarshal(out, &again); err != nil {
			t.Errorf("%s: written as %s, which doesn't read again: %v", c.in, out, err)
		}
	}
}

func TestIsNCName(t *testing.T) {
	for s, want := range map[string]bool{"a": true, "_a": true, "Story": true, "a1": true, "a-b.c": true, "Å": true,
		"": false, "0": false, "-a": false, ".a": false, "a:b": false, "1a": false} {
		if got := isNCName(s); got != want {
			t.Errorf("isNCName(%q) = %v, want %v", s, got, want)
		}
	}
}
