package idml

import (
	"errors"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/dimelords/idmllib/v3/pkg/common"
	"github.com/dimelords/idmllib/v3/pkg/story"
)

// textvariables.idml was written by InDesign 2026: custom text variables "Publication" and "Weekday", placed on a
// page ("Head: <Publication> / <Weekday>") and on the parent page ("<Publication>").
const textVariablesFixture = "../../testdata/textvariables.idml"

func TestReadCustomTextVariables(t *testing.T) {
	pkg, err := Read(textVariablesFixture)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := pkg.Document()
	if err != nil {
		t.Fatal(err)
	}
	v := doc.TextVariable("Publication")
	if v == nil || v.Self != "dTextVariablenPublication" {
		t.Fatalf("Publication: %+v", v)
	}
	if text, ok := v.CustomText(); !ok || text != "The Daily Example" {
		t.Errorf("CustomText() = %q, %v", text, ok)
	}
	if _, ok := doc.TextVariable("Chapter Number").CustomText(); ok {
		t.Error("a built-in variable reported custom text")
	}

	stories, err := pkg.Stories()
	if err != nil {
		t.Fatal(err)
	}
	var texts []string
	instances := 0
	for _, st := range stories {
		instances += len(st.TextVariableInstances())
		texts = append(texts, st.ExtractText())
	}
	if instances != 3 {
		t.Errorf("%d instances, want 3", instances)
	}
	if !strings.Contains(strings.Join(texts, "|"), "Head: The Daily Example / Tuesday") {
		t.Errorf("story texts %q", texts)
	}
}

func TestSetCustomTextVariable(t *testing.T) {
	pkg, err := Read(textVariablesFixture)
	if err != nil {
		t.Fatal(err)
	}
	n, err := pkg.SetCustomTextVariable("Publication", "The Evening Example")
	if err != nil || n != 2 {
		t.Fatalf("SetCustomTextVariable = %d, %v; want 2 instances", n, err)
	}
	out := filepath.Join(t.TempDir(), "out.idml")
	if err := Write(pkg, out); err != nil {
		t.Fatal(err)
	}
	back, err := Read(out)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := back.Document()
	if err != nil {
		t.Fatal(err)
	}
	if text, _ := doc.TextVariable("Publication").CustomText(); text != "The Evening Example" {
		t.Errorf("after writing, Publication is %q", text)
	}
	if text, _ := doc.TextVariable("Weekday").CustomText(); text != "Tuesday" {
		t.Errorf("Weekday changed to %q", text)
	}
	stories, err := back.Stories()
	if err != nil {
		t.Fatal(err)
	}
	for path, st := range stories {
		for _, inst := range st.TextVariableInstances() {
			want := map[string]string{"Publication": "The Evening Example", "Weekday": "Tuesday"}[inst.Name]
			if inst.ResultText != want {
				t.Errorf("%s: %s shows %q, want %q", path, inst.Name, inst.ResultText, want)
			}
		}
	}

	if _, err := pkg.SetCustomTextVariable("Nothing", "x"); err == nil {
		t.Error("an unknown variable: no error")
	}
	var ce *common.Error
	if _, err := pkg.SetCustomTextVariable("Chapter Number", "x"); !errors.As(err, &ce) {
		t.Errorf("a built-in variable: %v, want a common.Error", err)
	}
}

// A variable added to a document that was read goes right after the existing ones, and a new story can show it.
func TestAddCustomTextVariable(t *testing.T) {
	pkg, err := Read(textVariablesFixture)
	if err != nil {
		t.Fatal(err)
	}
	v, err := pkg.AddCustomTextVariable("Edition Date", "10 November 2026")
	if err != nil {
		t.Fatal(err)
	}
	if v.Self != "dTextVariablenEdition Date" {
		t.Errorf("Self %q", v.Self)
	}
	if _, err := pkg.AddCustomTextVariable("Edition Date", "again"); err == nil {
		t.Error("a second variable of the same name: no error")
	}
	st := story.Story{Self: "uedition", ParagraphStyleRanges: []story.ParagraphStyleRange{{
		AppliedParagraphStyle: "ParagraphStyle/$ID/NormalParagraphStyle",
		CharacterStyleRanges:  []story.CharacterStyleRange{story.NewTextVariableRange("", "uinst", v.Name, v.Self, "10 November 2026")},
	}}}
	if err := pkg.AddStory("Stories/Story_uedition.xml", &st, ValidationOptions{}); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "out.idml")
	if err := Write(pkg, out); err != nil {
		t.Fatal(err)
	}

	back, err := Read(out)
	if err != nil {
		t.Fatal(err)
	}
	data, err := back.FileData(PathDesignmap)
	if err != nil {
		t.Fatal(err)
	}
	// The designmap's top-level children in order: the new variable follows the last existing one.
	var names []string
	for _, m := range regexp.MustCompile(`(?m)^\t<([A-Za-z][\w:]*)(?: Self="([^"]*)")?`).FindAllStringSubmatch(string(data), -1) {
		names = append(names, m[1]+" "+m[2])
	}
	i := indexOf(names, "TextVariable dTextVariablenEdition Date")
	if i < 1 || !strings.HasPrefix(names[i-1], "TextVariable ") || (i+1 < len(names) && strings.HasPrefix(names[i+1], "TextVariable ")) {
		t.Errorf("the new variable is not right after the others: %v", names)
	}
	nst, err := back.Story("Stories/Story_uedition.xml")
	if err != nil {
		t.Fatal(err)
	}
	insts := nst.TextVariableInstances()
	if len(insts) != 1 || insts[0].AssociatedTextVariable != "dTextVariablenEdition Date" || nst.ExtractText() != "10 November 2026" {
		t.Errorf("instances %+v, text %q", insts, nst.ExtractText())
	}
	raw, err := back.FileData("Stories/Story_uedition.xml")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `PageNumberType="TextVariable"><TextVariableInstance Self="uinst" Name="Edition Date" ResultText="10 November 2026" AssociatedTextVariable="dTextVariablenEdition Date"`) &&
		!regexp.MustCompile(`PageNumberType="TextVariable">\s*<TextVariableInstance Self="uinst" Name="Edition Date" ResultText="10 November 2026" AssociatedTextVariable="dTextVariablenEdition Date"`).Match(raw) {
		t.Errorf("story written as\n%s", raw)
	}
}

func indexOf(l []string, s string) int {
	for i, v := range l {
		if v == s {
			return i
		}
	}
	return -1
}
