package idml

import (
	"errors"
	"testing"
)

func TestStoryNotFoundError(t *testing.T) {
	err := &StoryNotFoundError{StoryID: "u123"}
	expected := "story not found: u123"

	if err.Error() != expected {
		t.Errorf("Expected error message %q, got %q", expected, err.Error())
	}

	// Test error type assertion
	var storyErr *StoryNotFoundError
	if !errors.As(err, &storyErr) {
		t.Error("Error should be assertable as StoryNotFoundError")
	}
}

func TestSpreadNotFoundError(t *testing.T) {
	err := &SpreadNotFoundError{StoryID: "u456"}
	expected := "no spreads found for story u456"

	if err.Error() != expected {
		t.Errorf("Expected error message %q, got %q", expected, err.Error())
	}

	var spreadErr *SpreadNotFoundError
	if !errors.As(err, &spreadErr) {
		t.Error("Error should be assertable as SpreadNotFoundError")
	}
}

func TestFileNotFoundError(t *testing.T) {
	err := &FileNotFoundError{FileName: "test.xml"}
	expected := "file not found in IDML: test.xml"

	if err.Error() != expected {
		t.Errorf("Expected error message %q, got %q", expected, err.Error())
	}

	var fileErr *FileNotFoundError
	if !errors.As(err, &fileErr) {
		t.Error("Error should be assertable as FileNotFoundError")
	}
}

func TestParseError(t *testing.T) {
	wrappedErr := errors.New("invalid XML")
	err := &ParseError{
		FileName: "Story.xml",
		Err:      wrappedErr,
	}

	expected := "failed to parse Story.xml: invalid XML"
	if err.Error() != expected {
		t.Errorf("Expected error message %q, got %q", expected, err.Error())
	}

	// Test error unwrapping
	if !errors.Is(err, wrappedErr) {
		t.Error("ParseError should unwrap to the original error")
	}

	var parseErr *ParseError
	if !errors.As(err, &parseErr) {
		t.Error("Error should be assertable as ParseError")
	}
}
