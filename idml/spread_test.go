package idml

import (
	"os"
	"testing"
)

func TestSpread(t *testing.T) {
	p := Package{}
	data, _ := os.ReadFile("../testdata/Spread_u210.xml")
	spread, _ := p.readSpread(data)

	if len(spread.Pages) != 2 {
		t.Errorf("Expected 2 pages, got %d", len(spread.Pages))
		t.Fail()
	}

	if len(spread.TextFrames) != 12 {
		t.Errorf("Expected 12 text frames, got %d", len(spread.TextFrames))
		t.Fail()
	}
}
