package xmlutil

import "bytes"

// Embedded image payloads.
//
// In IDML a graphic page item stores its embedded image as a <Contents>
// element holding a CDATA-wrapped base64 payload:
//
//	<Image Self="u26a" ...><Contents><![CDATA[/9j/2wBD...]]></Contents></Image>
//
// These payloads dominate the size of a spread file: in the test corpus one
// 121 MB spread is 99.9% payload. Parsing such a file materializes the payload
// a second time, doubling memory. StripContents removes the payloads from a
// copy of the XML and hands back the originals as sub-slices of the input, so
// the parsed document can be small while the bytes stay available for writing.
// RestoreContents puts them back.

var (
	contentsOpen  = []byte("<Contents>")
	contentsClose = []byte("</Contents>")
	// contentsEmpty forms that RestoreContents will fill in.
	contentsEmptyPair = []byte("<Contents></Contents>")
	contentsSelfClose = []byte("<Contents/>")
	selfAttr          = []byte(` Self="`)
	cdataOpen         = []byte("<![CDATA[")
	cdataClose        = []byte("]]>")
)

// contentsOwners are the page item elements that can carry an embedded
// <Contents> payload, longest first so prefix matching is unambiguous.
var contentsOwners = [][]byte{
	[]byte("<ImportedPage "),
	[]byte("<Image "),
	[]byte("<PICT "),
	[]byte("<WMF "),
	[]byte("<EPS "),
	[]byte("<PDF "),
}

// ownerSelfBefore returns the Self attribute of the nearest page item start
// tag that precedes pos, which is the element the <Contents> at pos belongs
// to. It returns "" when no owner can be identified.
func ownerSelfBefore(data []byte, pos int) string {
	start := -1
	for _, owner := range contentsOwners {
		if i := bytes.LastIndex(data[:pos], owner); i > start {
			start = i
		}
	}
	if start < 0 {
		return ""
	}
	// The Self attribute must be inside this start tag.
	end := bytes.IndexByte(data[start:pos], '>')
	if end < 0 {
		return ""
	}
	tag := data[start : start+end]
	_, rest, ok := bytes.Cut(tag, selfAttr)
	if !ok {
		return ""
	}
	value, _, ok := bytes.Cut(rest, []byte(`"`))
	if !ok {
		return ""
	}
	return string(value)
}

// StripContents returns a copy of data with every embedded <Contents> payload
// removed, plus the payloads keyed by the Self attribute of the page item that
// owns them. Payload values are sub-slices of data and share its memory, so
// they cost nothing as long as data is retained.
//
// A payload whose owner cannot be identified, whose owner's Self is already
// present, or whose owner also has an empty <Contents> is left inline so
// nothing is lost: RestoreContents matches by owner, so it could not tell the
// emptied element from one that was empty to begin with. When no payload is
// found the returned slice is data itself and the map is nil.
func StripContents(data []byte) ([]byte, map[string][]byte) {
	if !bytes.Contains(data, contentsOpen) {
		return data, nil
	}
	var (
		out      []byte
		payloads map[string][]byte
		search   int // where to look for the next <Contents>
		copied   int // data before this is already in out
		empty    = ownersWithEmptyContents(data)
	)
	for {
		i := bytes.Index(data[search:], contentsOpen)
		if i < 0 {
			break
		}
		openAt := search + i
		bodyAt := openAt + len(contentsOpen)
		j := bytes.Index(data[bodyAt:], contentsClose)
		if j < 0 {
			break // unterminated; leave the rest untouched
		}
		bodyEnd := bodyAt + j
		search = bodyEnd + len(contentsClose)

		// A payload that stays inline is copied along with the bytes around
		// it, so only the search moves past it here.
		self := ownerSelfBefore(data, openAt)
		if self == "" || payloads[self] != nil || bodyEnd == bodyAt || empty[self] ||
			hasOwnerTag(data[bodyAt:bodyEnd]) {
			continue
		}
		if payloads == nil {
			payloads = make(map[string][]byte)
			out = make([]byte, 0, len(data)/64+1024)
		}
		payloads[self] = data[bodyAt:bodyEnd]
		out = append(out, data[copied:bodyAt]...) // up to and including <Contents>
		copied = bodyEnd                          // skip the payload
	}
	if payloads == nil {
		return data, nil
	}
	out = append(out, data[copied:]...)
	return out, payloads
}

// RestoreContents returns a copy of data with each empty <Contents> element
// filled in from payloads, matched by the Self attribute of the owning page
// item. Elements that already carry content are left alone, so a payload the
// caller set explicitly is never overwritten. data is returned unchanged when
// payloads is empty or nothing matches.
func RestoreContents(data []byte, payloads map[string][]byte) []byte {
	if len(payloads) == 0 {
		return data
	}
	var (
		out    []byte
		search int // where to look for the next empty <Contents>
		copied int // data before this is already in out
	)
	for {
		next, form := nextEmptyContents(data, search)
		if next < 0 {
			break
		}
		search = next + len(form)

		// An empty element with no payload stays as it is and is copied
		// along with the bytes around it.
		self := ownerSelfBefore(data, next)
		payload, ok := payloads[self]
		if !ok {
			continue
		}
		if out == nil {
			out = make([]byte, 0, len(data)+payloadBytes(payloads))
		}
		out = append(out, data[copied:next]...)
		out = append(out, contentsOpen...)
		out = append(out, payload...)
		out = append(out, contentsClose...)
		copied = search
	}
	if out == nil {
		return data
	}
	return append(out, data[copied:]...)
}

// hasOwnerTag reports whether a payload holds what looks like an owner's
// start tag. Removing such a payload would change which owner
// ownerSelfBefore finds for a later <Contents>, so the stripped and the
// original data would disagree about ownership. Base64 never contains "<", so
// the common case is one scan inside the CDATA wrapper.
func hasOwnerTag(body []byte) bool {
	inner := bytes.TrimSuffix(bytes.TrimPrefix(body, cdataOpen), cdataClose)
	if bytes.IndexByte(inner, '<') < 0 {
		return false
	}
	for _, owner := range contentsOwners {
		if bytes.Contains(body, owner) {
			return true
		}
	}
	return false
}

// ownersWithEmptyContents returns the Self of every owner that has an empty
// <Contents></Contents> or <Contents/> element in data. It visits each
// "<Contents" once, so it stays linear in the size of data.
func ownersWithEmptyContents(data []byte) map[string]bool {
	var owners map[string]bool
	prefix := contentsOpen[:len(contentsOpen)-1] // "<Contents"
	for pos := 0; ; {
		i := bytes.Index(data[pos:], prefix)
		if i < 0 {
			return owners
		}
		at := pos + i
		rest := data[at:]
		pos = at + len(prefix)
		if !bytes.HasPrefix(rest, contentsEmptyPair) && !bytes.HasPrefix(rest, contentsSelfClose) {
			continue
		}
		if self := ownerSelfBefore(data, at); self != "" {
			if owners == nil {
				owners = make(map[string]bool)
			}
			owners[self] = true
		}
	}
}

// nextEmptyContents finds the next empty <Contents> element at or after pos
// and reports its offset and the exact bytes that form it.
func nextEmptyContents(data []byte, pos int) (int, []byte) {
	a := bytes.Index(data[pos:], contentsEmptyPair)
	b := bytes.Index(data[pos:], contentsSelfClose)
	switch {
	case a < 0 && b < 0:
		return -1, nil
	case b < 0 || (a >= 0 && a <= b):
		return pos + a, contentsEmptyPair
	default:
		return pos + b, contentsSelfClose
	}
}

func payloadBytes(payloads map[string][]byte) int {
	n := 0
	for _, p := range payloads {
		n += len(p)
	}
	return n
}
