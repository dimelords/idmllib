package idml

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/dimelords/idmllib/v3/pkg/spread"
	"github.com/dimelords/idmllib/v3/pkg/story"
)

// features.idml was exported from InDesign 2026 and contains a tagged story
// with a tracked change, hyperlink, note, footnote, anchored rectangle and a
// 3x3 table, plus a group, button, master page item and interleaved stacking
// order on the spread.

func readFeatures(t *testing.T) *Package {
	t.Helper()
	pkg, err := Read(filepath.Join("..", "..", "testdata", "features.idml"))
	if err != nil {
		t.Fatal(err)
	}
	return pkg
}

func mainStory(t *testing.T, pkg *Package) *story.Story {
	t.Helper()
	stories, err := pkg.Stories()
	if err != nil {
		t.Fatal(err)
	}
	for _, st := range stories {
		if len(st.Tables()) > 0 {
			return st
		}
	}
	t.Fatal("no story with a table found")
	return nil
}

func onlySpread(t *testing.T, pkg *Package) *spread.Spread {
	t.Helper()
	spreads, err := pkg.Spreads()
	if err != nil {
		t.Fatal(err)
	}
	if len(spreads) != 1 {
		t.Fatalf("got %d spreads, want 1", len(spreads))
	}
	for _, s := range spreads {
		return s
	}
	return nil
}

func TestFeaturesTaggedStoryText(t *testing.T) {
	st := mainStory(t, readFeatures(t))
	text := st.ExtractText()

	for _, want := range []string{"Inserted Heading paragraph", "Body text with a hyperlink word", "Third paragraph before the table."} {
		if !strings.Contains(text, want) {
			t.Errorf("ExtractText missing %q in:\n%s", want, text)
		}
	}
	for _, unwanted := range []string{"Footnote text.", "Editor note", "Cell 0"} {
		if strings.Contains(text, unwanted) {
			t.Errorf("ExtractText should not include %q (not in main flow):\n%s", unwanted, text)
		}
	}
	if len(st.ParagraphStyleRanges) != 0 {
		t.Errorf("tagged story should hold its paragraphs inside XMLElement, got %d top-level ranges", len(st.ParagraphStyleRanges))
	}
	if got := len(st.Paragraphs()); got < 2 {
		t.Errorf("Paragraphs() = %d, want at least 2 (InDesign merges same-style paragraphs into one range)", got)
	}
}

func TestFeaturesTypedInlineContent(t *testing.T) {
	st := mainStory(t, readFeatures(t))
	var changes, notes, links, footnotes, tables int
	var footnoteText, noteText string
	for _, psr := range st.Paragraphs() {
		for _, csr := range psr.Ranges() {
			for _, ch := range csr.Children {
				switch {
				case ch.Change != nil:
					changes++
					if ch.Change.ChangeType != "InsertedText" {
						t.Errorf("ChangeType = %q", ch.Change.ChangeType)
					}
				case ch.Note != nil:
					notes++
					noteText = ch.Note.ExtractText()
				case ch.HyperlinkTextSource != nil:
					links++
				case ch.Footnote != nil:
					footnotes++
					footnoteText = ch.Footnote.ExtractText()
				case ch.Table != nil:
					tables++
				}
			}
		}
	}
	if changes != 1 || notes != 1 || links != 1 || footnotes != 1 || tables != 1 {
		t.Errorf("changes=%d notes=%d links=%d footnotes=%d tables=%d, want 1 each", changes, notes, links, footnotes, tables)
	}
	if strings.TrimSpace(footnoteText) != "Footnote text." {
		t.Errorf("footnote text = %q", footnoteText)
	}
	if noteText != "Editor note" {
		t.Errorf("note text = %q", noteText)
	}
}

func TestFeaturesTable(t *testing.T) {
	st := mainStory(t, readFeatures(t))
	tables := st.Tables()
	if len(tables) != 1 {
		t.Fatalf("got %d tables, want 1", len(tables))
	}
	tb := tables[0]
	if len(tb.Rows) != 3 || len(tb.Columns) != 3 || len(tb.Cells) != 9 {
		t.Errorf("rows=%d cols=%d cells=%d, want 3/3/9", len(tb.Rows), len(tb.Columns), len(tb.Cells))
	}
	if tb.HeaderRowCount != "1" || tb.ColumnCount != "3" {
		t.Errorf("HeaderRowCount=%q ColumnCount=%q", tb.HeaderRowCount, tb.ColumnCount)
	}
	if c := tb.Cell(0, 0); c == nil || c.ExtractText() != "Cell 0" {
		t.Errorf("Cell(0,0) = %v", c)
	}
	if c := tb.Cell(2, 2); c == nil || c.ExtractText() != "Cell 8" {
		t.Errorf("Cell(2,2) = %v", c)
	}
	if tb.AppliedTableStyle == "" {
		t.Error("AppliedTableStyle empty")
	}
}

func TestFeaturesGroupsAndButtons(t *testing.T) {
	sp := onlySpread(t, readFeatures(t))
	if len(sp.Groups) != 1 {
		t.Fatalf("got %d groups, want 1", len(sp.Groups))
	}
	g := sp.Groups[0]
	if len(g.Rectangles) != 2 || len(g.Ovals) != 1 {
		t.Errorf("group has %d rectangles and %d ovals, want 2 and 1", len(g.Rectangles), len(g.Ovals))
	}
	if len(sp.Buttons) != 1 {
		t.Fatalf("got %d buttons, want 1", len(sp.Buttons))
	}
	b := sp.Buttons[0]
	if b.Name != "Btn" {
		t.Errorf("button name = %q", b.Name)
	}
	if len(b.States) != 1 || len(b.States[0].Groups) != 1 || len(b.States[0].Groups[0].Rectangles) != 1 {
		t.Errorf("button state structure unexpected: %d states", len(b.States))
	}
	if n := len(sp.OtherElements); n != 0 {
		for _, o := range sp.OtherElements {
			t.Logf("unmodeled top-level element: %s", o.XMLName.Local)
		}
		t.Errorf("%d unmodeled top-level page items remain", n)
	}
}
