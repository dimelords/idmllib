package common

import (
	"bytes"
	"encoding/xml"
	"strings"
	"testing"
)

type attrBase struct {
	Self string `xml:"Self,attr"`
	Name string `xml:"Name,attr,omitempty"`
}

type attrThing struct {
	attrBase
	Color      string     `xml:"Color,attr,omitempty"`
	Count      int        `xml:"Count,attr"` // non-string attrs are ignored by the helpers
	Ignored    string     `xml:"Ignored"`    // element, not attribute
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

func TestUnmarshalAttrsAssignsKnownAndKeepsUnknown(t *testing.T) {
	var v attrThing
	UnmarshalAttrs(&v, []xml.Attr{
		{Name: xml.Name{Local: "Self"}, Value: "u1"},
		{Name: xml.Name{Local: "Color"}, Value: "Red"},
		{Name: xml.Name{Local: "Unknown"}, Value: "x"},
		{Name: xml.Name{Space: "xmlns", Local: "idPkg"}, Value: "ns"},
	})
	if v.Self != "u1" || v.Color != "Red" {
		t.Fatalf("typed fields not set: %+v", v)
	}
	if len(v.OtherAttrs) != 1 || v.OtherAttrs[0].Name.Local != "Unknown" {
		t.Fatalf("unknown attribute not preserved: %+v", v.OtherAttrs)
	}
}

func TestMarshalAttrsHonorsOmitEmptyAndAppendsOthers(t *testing.T) {
	v := attrThing{attrBase: attrBase{Self: "u1"}, OtherAttrs: []xml.Attr{{Name: xml.Name{Local: "Extra"}, Value: "1"}}}
	got := MarshalAttrs(v)
	want := []string{"Self=u1", "Extra=1"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i, a := range got {
		if a.Name.Local+"="+a.Value != want[i] {
			t.Errorf("attr %d: got %s=%s, want %s", i, a.Name.Local, a.Value, want[i])
		}
	}
	// pointer input works the same
	if len(MarshalAttrs(&v)) != 2 {
		t.Errorf("pointer input should give the same result")
	}
	if MarshalAttrs(nil) != nil || MarshalAttrs((*attrThing)(nil)) != nil {
		t.Errorf("nil input should give nil")
	}
}

func TestMarshalAttrsRoundtripsThroughEncodingXML(t *testing.T) {
	in := `<Thing Self="u1" Name="n" Color="Red" Count="0" Zeta="z" Alpha="a"></Thing>`
	var v attrThing
	if err := xml.Unmarshal([]byte(in), &v); err != nil {
		t.Fatal(err)
	}
	// encoding/xml's own ",any,attr" handling collected the unknowns
	if len(v.OtherAttrs) != 2 {
		t.Fatalf("expected 2 unknown attrs, got %v", v.OtherAttrs)
	}
	attrs := MarshalAttrs(v)
	var names []string
	for _, a := range attrs {
		names = append(names, a.Name.Local)
	}
	if strings.Join(names, ",") != "Self,Name,Color,Zeta,Alpha" {
		t.Errorf("unexpected attribute order/content: %v", names)
	}
}

type orderedThing struct {
	As    []string
	Bs    []string
	other []string
	order ChildOrder
}

func (o *orderedThing) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			var s string
			if err := d.DecodeElement(&s, &t); err != nil {
				return err
			}
			switch t.Name.Local {
			case "A":
				o.As = append(o.As, s)
				o.order.Record("A")
			case "B":
				o.Bs = append(o.Bs, s)
				o.order.Record("B")
			default:
				o.other = append(o.other, t.Name.Local+":"+s)
				o.order.Record(kindOther)
			}
		case xml.EndElement:
			return nil
		}
	}
}

func (o orderedThing) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name = xml.Name{Local: "T"}
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	enc := func(name, val string) error {
		return e.EncodeElement(val, xml.StartElement{Name: xml.Name{Local: name}})
	}
	kinds := []ChildKind{
		{Name: "A", Len: len(o.As), Encode: func(e *xml.Encoder, i int) error { return enc("A", o.As[i]) }},
		{Name: "B", Len: len(o.Bs), Encode: func(e *xml.Encoder, i int) error { return enc("B", o.Bs[i]) }},
		{Name: kindOther, Len: len(o.other), Encode: func(e *xml.Encoder, i int) error {
			parts := strings.SplitN(o.other[i], ":", 2)
			return enc(parts[0], parts[1])
		}},
	}
	if err := EncodeChildren(e, o.order, kinds); err != nil {
		return err
	}
	return e.EncodeToken(start.End())
}

func TestEncodeChildrenPreservesInterleavedOrder(t *testing.T) {
	in := `<T><B>1</B><A>2</A><X>3</X><A>4</A><B>5</B></T>`
	var o orderedThing
	if err := xml.Unmarshal([]byte(in), &o); err != nil {
		t.Fatal(err)
	}
	out, err := xml.Marshal(o)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, []byte(in)) {
		t.Errorf("order not preserved:\n in: %s\nout: %s", in, out)
	}
}

func TestEncodeChildrenAppendsNewAndSkipsRemoved(t *testing.T) {
	in := `<T><B>1</B><A>2</A><A>3</A></T>`
	var o orderedThing
	if err := xml.Unmarshal([]byte(in), &o); err != nil {
		t.Fatal(err)
	}
	o.As = o.As[:1]                  // removed the last A
	o.Bs = append(o.Bs, "new")       // appended a B
	o.other = append(o.other, "Z:z") // appended an unknown
	out, err := xml.Marshal(o)
	if err != nil {
		t.Fatal(err)
	}
	want := `<T><B>1</B><A>2</A><B>new</B><Z>z</Z></T>`
	if string(out) != want {
		t.Errorf("\n got: %s\nwant: %s", out, want)
	}
}

func TestEncodeChildrenWithoutRecordedOrderGroupsByKind(t *testing.T) {
	o := orderedThing{As: []string{"a"}, Bs: []string{"b"}, other: []string{"Z:z"}}
	out, err := xml.Marshal(o)
	if err != nil {
		t.Fatal(err)
	}
	want := `<T><A>a</A><B>b</B><Z>z</Z></T>`
	if string(out) != want {
		t.Errorf("\n got: %s\nwant: %s", out, want)
	}
}

func TestPropertiesPreservesChildOrder(t *testing.T) {
	in := `<Properties><PageColor type="enumeration">UseMasterColor</PageColor><Descriptor type="list"><ListItem type="string">x</ListItem></Descriptor><Label><KeyValuePair Key="k" Value="v"></KeyValuePair></Label></Properties>`
	var p Properties
	if err := xml.Unmarshal([]byte(in), &p); err != nil {
		t.Fatal(err)
	}
	if p.Label == nil || len(p.OtherElements) != 2 {
		t.Fatalf("unexpected parse result: label=%v others=%d", p.Label, len(p.OtherElements))
	}
	out, err := xml.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != in {
		t.Errorf("order not preserved:\n in: %s\nout: %s", in, out)
	}
}
