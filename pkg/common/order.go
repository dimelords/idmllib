package common

import "encoding/xml"

// ChildOrder records the sequence of child kinds seen while unmarshaling an
// element whose children are stored in several typed slices. MarshalXML uses
// it to replay the children in their original document order, which matters
// in IDML: page item order on a spread is the stacking (z) order, and the
// order of ranges in a story is the text order.
//
// Each entry is a kind name chosen by the owning type, typically the child
// element's local name. The zero value means "no recorded order", in which
// case children are emitted grouped by kind in the order the kinds are listed.
type ChildOrder []string

// Record appends a kind to the order.
func (o *ChildOrder) Record(kind string) {
	*o = append(*o, kind)
}

// ChildKind describes one typed group of children for EncodeChildren.
type ChildKind struct {
	// Name is the kind identifier that matches entries in the ChildOrder.
	Name string
	// Len is the number of children currently held for this kind.
	Len int
	// Encode writes the i-th child of this kind.
	Encode func(e *xml.Encoder, i int) error
}

// EncodeChildren writes children following the recorded order. For each entry
// in order, the next not-yet-written child of that kind is emitted. Children
// that remain after the order is exhausted (for example items appended by the
// modification API) are written afterwards, grouped by kind in the order the
// kinds are listed. Entries in order whose kind has run out of children are
// skipped, so removing items never causes an error.
func EncodeChildren(e *xml.Encoder, order ChildOrder, kinds []ChildKind) error {
	cursor := make(map[string]int, len(kinds))
	byName := make(map[string]*ChildKind, len(kinds))
	for i := range kinds {
		byName[kinds[i].Name] = &kinds[i]
	}
	for _, k := range order {
		kd, ok := byName[k]
		if !ok {
			continue
		}
		i := cursor[k]
		if i >= kd.Len {
			continue
		}
		if err := kd.Encode(e, i); err != nil {
			return err
		}
		cursor[k] = i + 1
	}
	for _, kd := range kinds {
		for i := cursor[kd.Name]; i < kd.Len; i++ {
			if err := kd.Encode(e, i); err != nil {
				return err
			}
		}
	}
	return nil
}

// OptionalKind returns a ChildKind for a single optional child held behind a
// pointer. present reports whether the child exists; encode writes it.
func OptionalKind(name string, present bool, encode func(e *xml.Encoder) error) ChildKind {
	n := 0
	if present {
		n = 1
	}
	return ChildKind{Name: name, Len: n, Encode: func(e *xml.Encoder, _ int) error { return encode(e) }}
}

// DecodeRaw decodes the element that begins at start into a RawXMLElement,
// preserving its name, attributes and inner XML verbatim.
func DecodeRaw(d *xml.Decoder, start xml.StartElement) (RawXMLElement, error) {
	var raw RawXMLElement
	if err := d.DecodeElement(&raw, &start); err != nil {
		return raw, err
	}
	raw.XMLName = start.Name
	raw.Attrs = append([]xml.Attr(nil), start.Attr...)
	return raw, nil
}

// ForEachChild replays order over the child kinds given by names and their
// current counts, calling visit(kindIndex, i) in document order and then for
// any children left over, grouped by kind. It is the read-only counterpart of
// EncodeChildren, for code that needs to traverse children in document order.
func ForEachChild(order ChildOrder, names []string, lens []int, visit func(kind, i int)) {
	if len(lens) < len(names) {
		names = names[:len(lens)] // a kind without a count has no children
	}
	cursor := make([]int, len(names))
	index := make(map[string]int, len(names))
	for i, n := range names {
		index[n] = i
	}
	for _, k := range order {
		ki, ok := index[k]
		if !ok || cursor[ki] >= lens[ki] {
			continue
		}
		visit(ki, cursor[ki])
		cursor[ki]++
	}
	for ki := range names {
		for i := cursor[ki]; i < lens[ki]; i++ {
			visit(ki, i)
		}
	}
}
