//revive:disable:var-naming
package types

// TextFramePredicate is a function that determines if a TextFrame should be included in the export
type TextFramePredicate func(tf *TextFrame) bool

// SelectedFrame holds a TextFrame together with its parent Spread
type SelectedFrame struct {
	TextFrame TextFrame
	Spread    Spread
}
