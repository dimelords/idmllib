package common

import (
	"encoding/xml"
	"reflect"
	"slices"
	"strings"
)

// Generic order-preserving XML marshaling.
//
// encoding/xml stores children in typed struct fields and writes them back in
// field declaration order, grouped by field. IDML relies on document order in
// many places, so types with more than one kind of child implement
// UnmarshalXML/MarshalXML by delegating to UnmarshalOrdered/MarshalOrdered.
// These walk the struct's `xml` tags reflectively, decode children into the
// matching fields while recording a ChildOrder, and replay that order on
// output. Attributes go through UnmarshalAttrs/MarshalAttrs, so unknown ones
// survive in the `xml:",any,attr"` catch-all.
//
// Supported child field shapes: *Struct, []Struct, Struct, string, and the
// `xml:",any"` []RawXMLElement catch-all. A `xml:",chardata"` string field
// receives text content. Fields of embedded structs are included.

type elemField struct {
	index     []int
	name      string // element local name
	omitEmpty bool
	kind      reflect.Kind // Ptr, Slice, Struct or String
	elemType  reflect.Type // for Ptr/Slice: the element type
}

type elemLayout struct {
	fields    []elemField
	byName    map[string]*elemField
	anyIndex  []int // []RawXMLElement field, nil if none
	charIndex []int // ",chardata" string field, nil if none
	nameIndex []int // XMLName xml.Name field, nil if none
}

var rawSliceType = reflect.TypeFor[[]RawXMLElement]()

func buildLayout(t reflect.Type, parent []int, l *elemLayout) {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		idx := append(append([]int{}, parent...), i)
		if f.Anonymous && f.Type.Kind() == reflect.Struct {
			buildLayout(f.Type, idx, l)
			continue
		}
		if f.Name == "XMLName" && f.Type == reflect.TypeFor[xml.Name]() {
			l.nameIndex = idx
			continue
		}
		tag, ok := f.Tag.Lookup("xml")
		untagged := !ok || tag == "-"
		unexported := f.PkgPath != ""
		if untagged || unexported {
			continue
		}
		parts := strings.Split(tag, ",")
		flags := parts[1:]
		has := func(fl string) bool {
			return slices.Contains(flags, fl)
		}
		switch {
		case has("attr"):
			continue
		case has("innerxml"):
			continue
		case has("chardata"):
			if f.Type.Kind() == reflect.String {
				l.charIndex = idx
			}
			continue
		case has("any"):
			if f.Type == rawSliceType {
				l.anyIndex = idx
			}
			continue
		}
		name := parts[0]
		if sp := strings.LastIndex(name, " "); sp >= 0 {
			name = name[sp+1:] // "namespace name" form
		}
		if name == "" {
			name = f.Name
		}
		ef := elemField{index: idx, name: name, omitEmpty: has("omitempty"), kind: f.Type.Kind()}
		switch f.Type.Kind() {
		case reflect.Pointer:
			ef.elemType = f.Type.Elem()
		case reflect.Slice:
			ef.elemType = f.Type.Elem()
		case reflect.Struct, reflect.String:
		default:
			continue // unsupported shape; left to the caller's own handling
		}
		l.fields = append(l.fields, ef)
	}
}

func layoutOf(t reflect.Type) *elemLayout {
	l := &elemLayout{byName: map[string]*elemField{}}
	buildLayout(t, nil, l)
	for i := range l.fields {
		l.byName[l.fields[i].name] = &l.fields[i]
	}
	return l
}

// UnmarshalOrdered decodes the element starting at start into v (a pointer to
// struct), recording the document order of children in order.
func UnmarshalOrdered(v any, order *ChildOrder, d *xml.Decoder, start xml.StartElement) error {
	if d == nil {
		return Errorf("common", "unmarshal ordered", "", "decoder is nil")
	}
	rv := reflect.ValueOf(v)
	isStructPointer := rv.Kind() == reflect.Pointer && !rv.IsNil() && rv.Elem().Kind() == reflect.Struct
	if !isStructPointer {
		return Errorf("common", "unmarshal ordered", "", "target must be a non-nil pointer to struct")
	}
	rv = rv.Elem()
	l := layoutOf(rv.Type())
	if l.nameIndex != nil {
		rv.FieldByIndex(l.nameIndex).Set(reflect.ValueOf(start.Name))
	}
	UnmarshalAttrs(v, start.Attr)
	*order = nil

	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			ef, ok := l.byName[t.Name.Local]
			if !ok {
				if l.anyIndex == nil {
					if err := d.Skip(); err != nil {
						return err
					}
					continue
				}
				raw, err := DecodeRaw(d, t)
				if err != nil {
					return err
				}
				af := rv.FieldByIndex(l.anyIndex)
				af.Set(reflect.Append(af, reflect.ValueOf(raw)))
				order.Record(kindOther)
				continue
			}
			fv := rv.FieldByIndex(ef.index)
			switch ef.kind {
			case reflect.Pointer:
				nv := reflect.New(ef.elemType)
				if err := d.DecodeElement(nv.Interface(), &t); err != nil {
					return err
				}
				fv.Set(nv)
			case reflect.Slice:
				nv := reflect.New(ef.elemType)
				if err := d.DecodeElement(nv.Interface(), &t); err != nil {
					return err
				}
				fv.Set(reflect.Append(fv, nv.Elem()))
			case reflect.Struct:
				if err := d.DecodeElement(fv.Addr().Interface(), &t); err != nil {
					return err
				}
			case reflect.String:
				var s string
				if err := d.DecodeElement(&s, &t); err != nil {
					return err
				}
				fv.SetString(s)
			}
			order.Record(ef.name)
		case xml.CharData:
			if l.charIndex != nil {
				cf := rv.FieldByIndex(l.charIndex)
				cf.SetString(cf.String() + string(t))
			}
		case xml.EndElement:
			return nil
		}
	}
}

// MarshalOrdered encodes v (a struct or pointer to struct) as an element,
// replaying order for its children. The element name is start.Name, which the
// encoder always supplies; the struct's XMLName value is a fallback for
// callers that build the start element by hand.
func MarshalOrdered(v any, order ChildOrder, e *xml.Encoder, start xml.StartElement) error {
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return Errorf("common", "marshal ordered", "", "value must be a struct")
	}
	l := layoutOf(rv.Type())

	if start.Name.Local == "" {
		if l.nameIndex != nil {
			if n, ok := reflect.TypeAssert[xml.Name](rv.FieldByIndex(l.nameIndex)); ok && n.Local != "" {
				start.Name = xml.Name{Local: n.Local}
			}
		}
	}
	start.Name.Space = "" // namespaces are handled by prefixed local names in IDML
	start.Attr = append(start.Attr, MarshalAttrs(v)...)
	if err := e.EncodeToken(start); err != nil {
		return err
	}

	if l.charIndex != nil {
		if s := rv.FieldByIndex(l.charIndex).String(); s != "" {
			if err := e.EncodeToken(xml.CharData(s)); err != nil {
				return err
			}
		}
	}

	kinds := make([]ChildKind, 0, len(l.fields)+1)
	for i := range l.fields {
		ef := l.fields[i]
		fv := rv.FieldByIndex(ef.index)
		name := xml.StartElement{Name: xml.Name{Local: ef.name}}
		switch ef.kind {
		case reflect.Pointer:
			kinds = append(kinds, OptionalKind(ef.name, !fv.IsNil(), func(e *xml.Encoder) error {
				return e.EncodeElement(fv.Interface(), name)
			}))
		case reflect.Slice:
			kinds = append(kinds, ChildKind{Name: ef.name, Len: fv.Len(), Encode: func(e *xml.Encoder, i int) error {
				return e.EncodeElement(fv.Index(i).Interface(), name)
			}})
		case reflect.Struct:
			kinds = append(kinds, OptionalKind(ef.name, true, func(e *xml.Encoder) error {
				return e.EncodeElement(fv.Interface(), name)
			}))
		case reflect.String:
			s := fv.String()
			kinds = append(kinds, OptionalKind(ef.name, s != "" || !ef.omitEmpty, func(e *xml.Encoder) error {
				return e.EncodeElement(s, name)
			}))
		}
	}
	if l.anyIndex != nil {
		af := rv.FieldByIndex(l.anyIndex)
		kinds = append(kinds, ChildKind{Name: kindOther, Len: af.Len(), Encode: func(e *xml.Encoder, i int) error {
			return e.Encode(af.Index(i).Interface())
		}})
	}
	if err := EncodeChildren(e, order, kinds); err != nil {
		return err
	}
	return e.EncodeToken(start.End())
}
