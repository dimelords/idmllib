package document

import (
	"encoding/xml"
	"strings"
	"testing"
)

func TestCustomTextVariable(t *testing.T) {
	v := NewCustomTextVariable("Publication", "The Daily Example")
	data, err := xml.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	want := `<TextVariable Self="dTextVariablenPublication" Name="Publication" VariableType="CustomTextType">` +
		`<CustomTextVariablePreference><Properties><Contents type="string">The Daily Example</Contents></Properties></CustomTextVariablePreference></TextVariable>`
	if string(data) != want {
		t.Errorf("written as\n%s\nwant\n%s", data, want)
	}
	if err := v.SetCustomText("The Evening Example"); err != nil {
		t.Fatal(err)
	}
	if s, ok := v.CustomText(); !ok || s != "The Evening Example" {
		t.Errorf("CustomText() = %q, %v", s, ok)
	}

	// A custom variable without a preference gets one; another type is refused.
	bare := TextVariable{Self: "dTextVariablenX", Name: "X", VariableType: VariableTypeCustomText}
	if err := bare.SetCustomText("x"); err != nil {
		t.Fatal(err)
	}
	if s, _ := bare.CustomText(); s != "x" {
		t.Errorf("bare variable: %q", s)
	}
	date := TextVariable{Self: "dTextVariablenOutput Date", Name: "Output Date", VariableType: "OutputDateType"}
	if err := date.SetCustomText("x"); err == nil {
		t.Error("setting custom text on a date variable: no error")
	}
}

func TestAddTextVariableRefusesDuplicates(t *testing.T) {
	var d Document
	if _, err := d.AddTextVariable(NewCustomTextVariable("A", "a")); err != nil {
		t.Fatal(err)
	}
	if _, err := d.AddTextVariable(NewCustomTextVariable("A", "b")); err == nil || !strings.Contains(err.Error(), `"A"`) {
		t.Errorf("a second A: %v", err)
	}
	if _, err := d.AddTextVariable(TextVariable{Self: "x"}); err == nil {
		t.Error("a variable without a name: no error")
	}
	if d.TextVariable("A") == nil || d.TextVariable("B") != nil {
		t.Error("TextVariable lookup")
	}
}
