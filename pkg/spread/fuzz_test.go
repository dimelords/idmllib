package spread

import (
	"os"
	"path/filepath"
	"testing"
)

// FuzzParseSpread feeds arbitrary bytes to the spread parser. The parser may
// reject input with an error but must never panic, and anything it accepts
// must survive a marshal and a second parse.
func FuzzParseSpread(f *testing.F) {
	if seed, err := os.ReadFile(filepath.Join("..", "..", "testdata", "Spread_u210.xml")); err == nil {
		f.Add(seed)
	}
	f.Add([]byte(`<idPkg:Spread xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging" DOMVersion="20.4"><Spread Self="s"><Page Self="p"/><Rectangle Self="r"/><TextFrame Self="t"/><Group Self="g"><Oval Self="o"/></Group></Spread></idPkg:Spread>`))
	f.Add([]byte(`<idPkg:MasterSpread DOMVersion="20.4"><MasterSpread Self="m" Name="A-Master"><Page Self="p"/></MasterSpread></idPkg:MasterSpread>`))

	f.Fuzz(func(t *testing.T, data []byte) {
		sp, err := ParseSpread(data)
		if err != nil {
			return
		}
		out, err := MarshalSpread(sp)
		if err != nil {
			t.Fatalf("accepted input failed to marshal: %v", err)
		}
		if _, err := ParseSpread(out); err != nil {
			t.Fatalf("marshaled output failed to parse: %v\n%s", err, out)
		}
	})
}
