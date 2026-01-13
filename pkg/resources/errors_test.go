package resources

import (
	"errors"
	"testing"

	"github.com/dimelords/idmllib/pkg/common"
)

// TestError_Error tests the Error() method of the common.Error type.
func TestError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "error with operation",
			err:      common.WrapError("resources", "parse fonts", errors.New("unexpected EOF")),
			expected: "resources: parse fonts: unexpected EOF",
		},
		{
			name:     "error without operation",
			err:      common.WrapError("resources", "", errors.New("file not found")),
			expected: "resources: : file not found",
		},
		{
			name:     "error with ErrInvalidFormat",
			err:      common.WrapError("resources", "parse styles", common.ErrInvalidFormat),
			expected: "resources: parse styles: invalid format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			if result != tt.expected {
				t.Errorf("Error() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestError_Unwrap tests the Unwrap() method of the common.Error type.
func TestError_Unwrap(t *testing.T) {
	baseErr := errors.New("base error")
	err := common.WrapError("resources", "test operation", baseErr)

	if !errors.Is(err, baseErr) {
		t.Error("errors.Is() should find the base error after unwrapping")
	}
}

// TestError_WithErrInvalidFormat tests that ErrInvalidFormat can be detected.
func TestError_WithErrInvalidFormat(t *testing.T) {
	err := common.WrapError("resources", "parse document", ErrInvalidFormat)

	if !errors.Is(err, ErrInvalidFormat) {
		t.Error("errors.Is() should detect ErrInvalidFormat")
	}
}

// TestErrInvalidFormat tests that the exported error variable exists.
func TestErrInvalidFormat(t *testing.T) {
	if ErrInvalidFormat == nil {
		t.Error("ErrInvalidFormat should not be nil")
	}

	expected := "invalid format"
	if ErrInvalidFormat.Error() != expected {
		t.Errorf("ErrInvalidFormat.Error() = %q, want %q", ErrInvalidFormat.Error(), expected)
	}
}
