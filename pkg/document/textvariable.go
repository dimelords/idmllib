package document

import (
	"encoding/xml"
	"slices"

	"github.com/dimelords/idmllib/v3/pkg/common"
)

// VariableTypeCustomText is the VariableType of a custom text variable, whose text is set in the document rather
// than taken from it (a publication's name, an edition's date).
const VariableTypeCustomText = "CustomTextType"

// textVariablePrefix is how InDesign names a text variable's Self: "dTextVariablen" followed by its name.
const textVariablePrefix = "dTextVariablen"

// CustomTextVariablePreference holds a custom text variable's text, as InDesign writes it:
//
//	<CustomTextVariablePreference><Properties><Contents type="string">text</Contents></Properties></CustomTextVariablePreference>
type CustomTextVariablePreference struct {
	XMLName    xml.Name                      `xml:"CustomTextVariablePreference"`
	Properties *CustomTextVariableProperties `xml:"Properties,omitempty"`

	// Catch-all for other children
	OtherElements []common.RawXMLElement `xml:",any"`

	// OtherAttrs preserves attributes not modeled by a typed field, so
	// nothing is lost when the element is written back.
	OtherAttrs []xml.Attr `xml:",any,attr"`

	// childOrder records the document order of children so MarshalXML can
	// replay them; see common.ChildOrder.
	childOrder common.ChildOrder
}

// CustomTextVariableProperties is the Properties element of a CustomTextVariablePreference.
type CustomTextVariableProperties struct {
	XMLName  xml.Name    `xml:"Properties"`
	Contents *TypedValue `xml:"Contents,omitempty"`

	// Catch-all for other Properties children
	OtherElements []common.RawXMLElement `xml:",any"`

	// OtherAttrs preserves attributes not modeled by a typed field, so
	// nothing is lost when the element is written back.
	OtherAttrs []xml.Attr `xml:",any,attr"`

	// childOrder records the document order of children so MarshalXML can
	// replay them; see common.ChildOrder.
	childOrder common.ChildOrder
}

// TypedValue is a Properties child holding text with its type, such as <Contents type="string">text</Contents>.
type TypedValue struct {
	Type  string `xml:"type,attr,omitempty"`
	Value string `xml:",chardata"`

	// OtherAttrs preserves attributes not modeled by a typed field, so
	// nothing is lost when the element is written back.
	OtherAttrs []xml.Attr `xml:",any,attr"`
}

// NewCustomTextVariable returns a custom text variable named name with the given text, with the Self InDesign
// gives it ("dTextVariablen" + name). Add it to a document with Document.AddTextVariable.
func NewCustomTextVariable(name, contents string) TextVariable {
	return TextVariable{
		Self:         textVariablePrefix + name,
		Name:         name,
		VariableType: VariableTypeCustomText,
		CustomTextPreference: &CustomTextVariablePreference{
			Properties: &CustomTextVariableProperties{Contents: &TypedValue{Type: "string", Value: contents}},
		},
	}
}

// CustomText returns a custom text variable's text, and false when the variable isn't a custom text variable.
func (v *TextVariable) CustomText() (string, bool) {
	if v == nil || v.VariableType != VariableTypeCustomText {
		return "", false
	}
	if p := v.CustomTextPreference; p != nil && p.Properties != nil && p.Properties.Contents != nil {
		return p.Properties.Contents.Value, true
	}
	return "", true
}

// SetCustomText sets a custom text variable's text, creating its preference when it has none. It refuses a
// variable of another type. Instances of the variable in stories keep the ResultText they were written with until
// they are updated too; see story.TextVariableInstance.
func (v *TextVariable) SetCustomText(contents string) error {
	if v == nil || v.VariableType != VariableTypeCustomText {
		typ := ""
		if v != nil {
			typ = v.VariableType
		}
		return common.Errorf("document", "set custom text", "", "variable type %q is not %s", typ, VariableTypeCustomText)
	}
	if v.CustomTextPreference == nil {
		v.CustomTextPreference = &CustomTextVariablePreference{}
	}
	p := v.CustomTextPreference
	if p.Properties == nil {
		p.Properties = &CustomTextVariableProperties{}
	}
	if p.Properties.Contents == nil {
		p.Properties.Contents = &TypedValue{Type: "string"}
	}
	p.Properties.Contents.Value = contents
	return nil
}

// TextVariable returns the document's text variable with the given name, or nil.
func (d *Document) TextVariable(name string) *TextVariable {
	for i := range d.TextVariables {
		if d.TextVariables[i].Name == name {
			return &d.TextVariables[i]
		}
	}
	return nil
}

// AddTextVariable adds a text variable to the document and returns it. In a designmap that was read, the new
// variable is written right after the existing ones, where InDesign keeps them, rather than at the end, which
// InDesign doesn't accept for every element. It refuses a variable without a name, or whose name or Self the
// document already has.
func (d *Document) AddTextVariable(v TextVariable) (*TextVariable, error) {
	if v.Name == "" || v.Self == "" {
		return nil, common.Errorf("document", "add text variable", v.Self, "a text variable needs a name and a Self")
	}
	for _, o := range d.TextVariables {
		if o.Name == v.Name || o.Self == v.Self {
			return nil, common.Errorf("document", "add text variable", v.Self, "the document already has text variable %q (%s)", o.Name, o.Self)
		}
	}
	d.TextVariables = append(d.TextVariables, v)
	if last := lastIndex(d.childOrder, "TextVariable"); last >= 0 {
		d.childOrder = slices.Insert(d.childOrder, last+1, "TextVariable")
	}
	return &d.TextVariables[len(d.TextVariables)-1], nil
}

func lastIndex(o common.ChildOrder, kind string) int {
	for i, k := range slices.Backward(o) {
		if k == kind {
			return i
		}
	}
	return -1
}
