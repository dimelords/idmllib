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
// A payload whose owner cannot be identified, or whose owner's Self is already
// present, is left inline so nothing is lost. When no payload is found the
// returned slice is data itself and the map is nil.
func StripContents(data []byte) ([]byte, map[string][]byte) {
	if !bytes.Contains(data, contentsOpen) {
		return data, nil
	}
	var (
		out      []byte
		payloads map[string][]byte
		pos      int
	)
	for {
		i := bytes.Index(data[pos:], contentsOpen)
		if i < 0 {
			break
		}
		openAt := pos + i
		bodyAt := openAt + len(contentsOpen)
		j := bytes.Index(data[bodyAt:], contentsClose)
		if j < 0 {
			break // unterminated; leave the rest untouched
		}
		bodyEnd := bodyAt + j

		self := ownerSelfBefore(data, openAt)
		if self == "" || payloads[self] != nil || bodyEnd == bodyAt {
			pos = bodyEnd + len(contentsClose)
			continue
		}
		if payloads == nil {
			payloads = make(map[string][]byte)
			out = make([]byte, 0, len(data)/64+1024)
		}
		payloads[self] = data[bodyAt:bodyEnd]
		out = append(out, data[pos:bodyAt]...) // up to and including <Contents>
		pos = bodyEnd                          // skip the payload
	}
	if payloads == nil {
		return data, nil
	}
	out = append(out, data[pos:]...)
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
	var out []byte
	pos := 0
	for {
		next, form := nextEmptyContents(data, pos)
		if next < 0 {
			break
		}
		self := ownerSelfBefore(data, next)
		payload, ok := payloads[self]
		if !ok {
			pos = next + len(form)
			continue
		}
		if out == nil {
			out = make([]byte, 0, len(data)+payloadBytes(payloads))
		}
		out = append(out, data[pos:next]...)
		out = append(out, contentsOpen...)
		out = append(out, payload...)
		out = append(out, contentsClose...)
		pos = next + len(form)
	}
	if out == nil {
		return data
	}
	return append(out, data[pos:]...)
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
