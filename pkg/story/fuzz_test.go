package story

import (
	"os"
	"path/filepath"
	"testing"
)

// FuzzParseStory feeds arbitrary bytes to the story parser. The parser may
// reject input with an error but must never panic, and anything it accepts
// must survive a marshal and a second parse.
func FuzzParseStory(f *testing.F) {
	if seed, err := os.ReadFile(filepath.Join("..", "..", "testdata", "story_u1d8.xml")); err == nil {
		f.Add(seed)
	}
	f.Add([]byte(`<?xml version="1.0"?><idPkg:Story xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging" DOMVersion="20.4"><Story Self="u1"><ParagraphStyleRange><CharacterStyleRange><Content>a<?ACE 4?>b</Content><Br/></CharacterStyleRange></ParagraphStyleRange></Story></idPkg:Story>`))
	f.Add([]byte(`<idPkg:Story><Story><XMLElement MarkupTag="t"><ParagraphStyleRange/></XMLElement></Story></idPkg:Story>`))

	f.Fuzz(func(t *testing.T, data []byte) {
		st, err := ParseStory(data)
		if err != nil {
			return
		}
		out, err := MarshalStory(st)
		if err != nil {
			t.Fatalf("accepted input failed to marshal: %v", err)
		}
		if _, err := ParseStory(out); err != nil {
			t.Fatalf("marshaled output failed to parse: %v\n%s", err, out)
		}
		_ = st.ExtractText()
	})
}
