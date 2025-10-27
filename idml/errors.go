// Package idml provides functionality for parsing and processing InDesign IDML files.
package idml

import "fmt"

// StoryNotFoundError is returned when a story with the given ID cannot be found
type StoryNotFoundError struct {
	StoryID string
}

func (e *StoryNotFoundError) Error() string {
	return fmt.Sprintf("story not found: %s", e.StoryID)
}

// SpreadNotFoundError is returned when no spreads contain the specified story
type SpreadNotFoundError struct {
	StoryID string
}

func (e *SpreadNotFoundError) Error() string {
	return fmt.Sprintf("no spreads found for story %s", e.StoryID)
}

// FileNotFoundError is returned when a required file is not found in the IDML archive
type FileNotFoundError struct {
	FileName string
}

func (e *FileNotFoundError) Error() string {
	return fmt.Sprintf("file not found in IDML: %s", e.FileName)
}

// ParseError is returned when XML parsing fails
type ParseError struct {
	FileName string
	Err      error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("failed to parse %s: %v", e.FileName, e.Err)
}

func (e *ParseError) Unwrap() error {
	return e.Err
}
