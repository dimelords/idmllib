package common

import (
	"encoding/xml"
	"reflect"
	"strings"
)

// attrField describes one struct field that maps to an XML attribute.
type attrField struct {
	index     []int  // field index path (supports embedded structs)
	name      string // attribute local name
	omitEmpty bool
	any       bool // true for the `xml:",any,attr"` catch-all field
}

// attrFields walks t (a struct type) and returns its attribute-mapped fields,
// including those of embedded structs, in declaration order.
func attrFields(t reflect.Type, parent []int) []attrField {
	var out []attrField
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		idx := append(append([]int{}, parent...), i)
		if f.Anonymous && f.Type.Kind() == reflect.Struct {
			out = append(out, attrFields(f.Type, idx)...)
			continue
		}
		tag, ok := f.Tag.Lookup("xml")
		if !ok {
			continue
		}
		parts := strings.Split(tag, ",")
		if len(parts) < 2 {
			continue
		}
		flags := parts[1:]
		isAttr, isAny, omit := false, false, false
		for _, fl := range flags {
			switch fl {
			case "attr":
				isAttr = true
			case "any":
				isAny = true
			case "omitempty":
				omit = true
			}
		}
		if !isAttr {
			continue
		}
		if isAny {
			if f.Type == reflect.TypeFor[[]xml.Attr]() {
				out = append(out, attrField{index: idx, any: true})
			}
			continue
		}
		if f.Type.Kind() != reflect.String {
			continue
		}
		name := parts[0]
		if name == "" {
			name = f.Name
		}
		out = append(out, attrField{index: idx, name: name, omitEmpty: omit})
	}
	return out
}

// MarshalAttrs builds the XML attribute list for a struct value v (or pointer
// to struct) from its `xml:"Name,attr"` string fields, honoring omitempty, and
// appends the contents of its `xml:",any,attr"` catch-all field if present.
//
// It is intended for types that implement a custom MarshalXML to control child
// order but still want attribute handling to follow the struct tags.
func MarshalAttrs(v any) []xml.Attr {
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil
	}
	var attrs []xml.Attr
	var other []xml.Attr
	for _, af := range attrFields(rv.Type(), nil) {
		fv := rv.FieldByIndex(af.index)
		if af.any {
			if extra, ok := reflect.TypeAssert[[]xml.Attr](fv); ok {
				other = append(other, extra...)
			}
			continue
		}
		s := fv.String()
		if s == "" && af.omitEmpty {
			continue
		}
		attrs = append(attrs, xml.Attr{Name: xml.Name{Local: af.name}, Value: s})
	}
	return append(attrs, other...)
}

// UnmarshalAttrs assigns attrs to the `xml:"Name,attr"` string fields of v
// (a pointer to struct). Attributes with no matching field are stored in the
// `xml:",any,attr"` catch-all field when one exists, so nothing is lost.
//
// Namespace declarations (xmlns / xmlns:prefix) are skipped, since the encoder
// manages those itself.
func UnmarshalAttrs(v any, attrs []xml.Attr) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return
	}
	fields := attrFields(rv.Type(), nil)
	byName := make(map[string]attrField, len(fields))
	var anyField *attrField
	for i := range fields {
		if fields[i].any {
			anyField = &fields[i]
			continue
		}
		byName[fields[i].name] = fields[i]
	}
	var other []xml.Attr
	for _, a := range attrs {
		isNamespaceDecl := a.Name.Space == "xmlns" || (a.Name.Space == "" && a.Name.Local == "xmlns")
		if isNamespaceDecl {
			continue
		}
		if af, ok := byName[a.Name.Local]; ok && a.Name.Space == "" {
			rv.FieldByIndex(af.index).SetString(a.Value)
			continue
		}
		other = append(other, a)
	}
	if anyField != nil && len(other) > 0 {
		rv.FieldByIndex(anyField.index).Set(reflect.ValueOf(other))
	}
}
