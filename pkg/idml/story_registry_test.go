package idml

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/dimelords/idmllib/v3/pkg/story"
)

func newTestStory(id string) *story.Story {
	st := &story.Story{Self: id}
	psr := story.ParagraphStyleRange{AppliedParagraphStyle: "ParagraphStyle/$ID/NormalParagraphStyle"}
	psr.CharacterStyleRanges = append(psr.CharacterStyleRanges, story.NewCharacterStyleRange("", []story.Content{{Text: "Hello"}}))
	st.ParagraphStyleRanges = append(st.ParagraphStyleRanges, psr)
	return st
}

// TestAddStoryRegistersInDesignmap checks that a story added through the API
// is referenced from designmap.xml, both as an idPkg:Story element and in the
// StoryList attribute, and that removing it undoes both. Without the
// reference InDesign ignores the story file entirely.
func TestAddStoryRegistersInDesignmap(t *testing.T) {
	pkg, err := Read(filepath.Join("..", "..", "testdata", "plain.idml"))
	if err != nil {
		t.Fatal(err)
	}
	const id = "utest1"
	filename := StoryPath(id)
	if err := pkg.AddStory(filename, newTestStory(id), ValidationOptions{}); err != nil {
		t.Fatalf("AddStory: %v", err)
	}

	out := filepath.Join(t.TempDir(), "out.idml")
	if err := Write(pkg, out); err != nil {
		t.Fatal(err)
	}
	pkg2, err := Read(out)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := pkg2.Document()
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, ref := range doc.Stories {
		if ref.Src == filename {
			found = true
		}
	}
	if !found {
		t.Errorf("designmap has no idPkg:Story reference to %s", filename)
	}
	if !strings.Contains(" "+doc.StoryList+" ", " "+id+" ") {
		t.Errorf("StoryList %q does not contain %s", doc.StoryList, id)
	}
	if !strings.Contains(string(mustFileData(t, pkg2, PathDesignmap)), `<idPkg:Story src="`+filename+`"`) {
		t.Errorf("designmap.xml does not contain a prefixed idPkg:Story element for %s", filename)
	}

	if _, err := pkg2.RemoveStory(filename, false); err != nil {
		t.Fatalf("RemoveStory: %v", err)
	}
	doc, _ = pkg2.Document()
	for _, ref := range doc.Stories {
		if ref.Src == filename {
			t.Errorf("reference to %s still present after RemoveStory", filename)
		}
	}
	if strings.Contains(" "+doc.StoryList+" ", " "+id+" ") {
		t.Errorf("StoryList %q still contains %s after RemoveStory", doc.StoryList, id)
	}
}

func mustFileData(t *testing.T, pkg *Package, path string) []byte {
	t.Helper()
	data, err := pkg.FileData(path)
	if err != nil {
		t.Fatalf("FileData(%s): %v", path, err)
	}
	return data
}

func TestStoryIDFromPath(t *testing.T) {
	cases := map[string]string{
		"Stories/Story_u123.xml": "u123",
		"Story_abc.xml":          "abc",
		"Stories/other.xml":      "other",
	}
	for in, want := range cases {
		if got := storyIDFromPath(in); got != want {
			t.Errorf("storyIDFromPath(%q) = %q, want %q", in, got, want)
		}
	}
}
