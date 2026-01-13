package idml

import (
	"errors"
	"fmt"
)

// NewMissingResourcesError creates an error from a MissingResources struct.
// This is used when validation detects missing dependencies.
func NewMissingResourcesError(mr *MissingResources) error {
	if !mr.HasMissing() {
		return nil
	}

	msg := "missing resources:"
	if len(mr.Fonts) > 0 {
		msg += fmt.Sprintf(" %d fonts", len(mr.Fonts))
	}
	if len(mr.ParagraphStyles) > 0 {
		msg += fmt.Sprintf(" %d paragraph styles", len(mr.ParagraphStyles))
	}
	if len(mr.CharacterStyles) > 0 {
		msg += fmt.Sprintf(" %d character styles", len(mr.CharacterStyles))
	}
	if len(mr.ObjectStyles) > 0 {
		msg += fmt.Sprintf(" %d object styles", len(mr.ObjectStyles))
	}
	if len(mr.Colors) > 0 {
		msg += fmt.Sprintf(" %d colors", len(mr.Colors))
	}
	if len(mr.Swatches) > 0 {
		msg += fmt.Sprintf(" %d swatches", len(mr.Swatches))
	}
	if len(mr.Layers) > 0 {
		msg += fmt.Sprintf(" %d layers", len(mr.Layers))
	}

	return errors.New(msg)
}
