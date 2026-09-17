package idml

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// FuzzReadBytes feeds arbitrary bytes to the package reader, which combines
// ZIP handling, size limits and the XML parsers. It may reject input with an
// error but must never panic, and a package it accepts must be writable.
func FuzzReadBytes(f *testing.F) {
	if seed, err := os.ReadFile(filepath.Join("..", "..", "testdata", "plain.idml")); err == nil {
		f.Add(seed)
	}
	// A minimal hand-built archive with a designmap and one story.
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	mw, _ := zw.CreateHeader(&zip.FileHeader{Name: "mimetype", Method: zip.Store})
	_, _ = mw.Write([]byte("application/vnd.adobe.indesign-idml-package"))
	dw, _ := zw.Create("designmap.xml")
	_, _ = dw.Write([]byte(`<?xml version="1.0"?><Document xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging" DOMVersion="20.4" Self="d"><idPkg:Story src="Stories/Story_u1.xml"/></Document>`))
	sw, _ := zw.Create("Stories/Story_u1.xml")
	_, _ = sw.Write([]byte(`<?xml version="1.0"?><idPkg:Story xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging" DOMVersion="20.4"><Story Self="u1"><ParagraphStyleRange><CharacterStyleRange><Content>x</Content></CharacterStyleRange></ParagraphStyleRange></Story></idPkg:Story>`))
	_ = zw.Close()
	f.Add(buf.Bytes())

	f.Fuzz(func(t *testing.T, data []byte) {
		pkg, err := ReadBytes(data)
		if err != nil {
			return
		}
		_, _ = pkg.Document()
		_, _ = pkg.Stories()
		_, _ = pkg.Spreads()
		_, _ = pkg.MasterSpreads()
		_, _ = pkg.Styles()
		_, _ = pkg.Fonts()
		_, _ = pkg.Graphics()
		out := filepath.Join(t.TempDir(), "out.idml")
		if err := Write(pkg, out); err != nil {
			t.Fatalf("accepted package failed to write: %v", err)
		}
	})
}
