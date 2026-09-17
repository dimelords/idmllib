package idml

import (
	"github.com/dimelords/idmllib/v3/internal/xmlutil"
)

// Streaming mode.
//
// A spread that contains embedded images is dominated by their base64
// payloads: in the test corpus one 121 MB spread is 99.9% payload. Parsing
// such a spread normally materializes every payload a second time inside the
// parsed document, so holding a parsed package costs roughly twice the file
// size.
//
// With ReadOptions.StreamingMode the payloads are kept only once, in the raw
// file bytes the package already holds, and the parsed spread carries empty
// <Contents> elements instead. Writing restores them:
//
//   - a spread that was not modified is written back from its raw bytes, so
//     the payloads are preserved exactly and no work is done at all;
//   - a modified spread is re-marshaled and each empty <Contents> is filled in
//     from the retained payload of the page item that owns it.
//
// The trade-off is that Contents is empty in the parsed document. Use
// ImageContents to read a payload, or read without streaming mode.

// contentsStore holds the stripped image payloads of one package, keyed by
// spread path and then by the Self attribute of the owning page item. The
// values are sub-slices of the raw file bytes, so they add no memory.
type contentsStore map[string]map[string][]byte

// prepareSpreadData strips embedded image payloads from a spread file before
// parsing when streaming mode is on, remembering them for write. It returns
// the bytes to parse.
func (p *Package) prepareSpreadData(filename string, data []byte) []byte {
	if !p.streamingMode {
		return data
	}
	stripped, payloads := xmlutil.StripContents(data)
	if len(payloads) == 0 {
		return data
	}
	if p.streamedContents == nil {
		p.streamedContents = make(contentsStore)
	}
	p.streamedContents[filename] = payloads
	return stripped
}

// restoreSpreadData puts stripped image payloads back into freshly marshaled
// spread XML. It is a no-op outside streaming mode and for spreads that had no
// embedded images.
func (p *Package) restoreSpreadData(filename string, data []byte) []byte {
	payloads := p.streamedContents[filename]
	if len(payloads) == 0 {
		return data
	}
	return xmlutil.RestoreContents(data, payloads)
}

// StreamingMode reports whether the package was read with
// ReadOptions.StreamingMode.
func (p *Package) StreamingMode() bool {
	return p.streamingMode
}

// ImageContents returns the embedded payload of a page item in a spread, for
// example the base64 image data of the <Image> with the given Self attribute.
// The returned bytes include the CDATA wrapper as stored in the file and must
// not be modified; they alias the package's own buffers.
//
// In streaming mode the payload was stripped from the parsed spread and is
// served from here. Outside streaming mode the payload is part of the parsed
// spread and this returns false.
func (p *Package) ImageContents(spreadPath, itemSelf string) ([]byte, bool) {
	payload, ok := p.streamedContents[spreadPath][itemSelf]
	return payload, ok
}
