package common

import (
	"errors"
	"testing"
)

func TestError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		expected string
	}{
		{
			name:     "full error",
			err:      &Error{Package: "idml", Op: "read", Path: "test.xml", Err: errors.New("file not found")},
			expected: "idml: read test.xml: file not found",
		},
		{
			name:     "no path",
			err:      &Error{Package: "idml", Op: "parse", Err: errors.New("invalid XML")},
			expected: "idml: parse: invalid XML",
		},
		{
			name:     "no package",
			err:      &Error{Op: "read", Path: "test.xml", Err: errors.New("permission denied")},
			expected: "read test.xml: permission denied",
		},
		{
			name:     "only error",
			err:      &Error{Err: errors.New("something went wrong")},
			expected: "something went wrong",
		},
		{
			name:     "package and op only",
			err:      &Error{Package: "spread", Op: "marshal"},
			expected: "spread: marshal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestError_Unwrap(t *testing.T) {
	underlying := errors.New("underlying error")
	err := &Error{Package: "idml", Op: "read", Err: underlying}

	if err.Unwrap() != underlying {
		t.Error("Unwrap() did not return the underlying error")
	}

	// Test with errors.Is
	if !errors.Is(err, underlying) {
		t.Error("errors.Is should return true for underlying error")
	}
}

func TestNewError_CreatesError(t *testing.T) {
	underlying := errors.New("test error")
	err := NewError("idml", "parse", "doc.xml", underlying)

	if err.Package != "idml" {
		t.Errorf("Package = %q, want %q", err.Package, "idml")
	}
	if err.Op != "parse" {
		t.Errorf("Op = %q, want %q", err.Op, "parse")
	}
	if err.Path != "doc.xml" {
		t.Errorf("Path = %q, want %q", err.Path, "doc.xml")
	}
	if err.Err != underlying {
		t.Error("Err not set correctly")
	}
}

func TestWrapError_WrapsError(t *testing.T) {
	// Test with non-nil error
	underlying := errors.New("test error")
	err := WrapError("idms", "export", underlying)
	if err == nil {
		t.Fatal("WrapError returned nil for non-nil error")
	}

	commonErr, ok := err.(*Error)
	if !ok {
		t.Fatal("WrapError did not return *Error")
	}
	if commonErr.Package != "idms" || commonErr.Op != "export" {
		t.Errorf("WrapError set incorrect fields: %+v", commonErr)
	}

	// Test with nil error
	if WrapError("idms", "export", nil) != nil {
		t.Error("WrapError should return nil for nil error")
	}
}

func TestWrapErrorWithPath_WrapsErrorWithPath(t *testing.T) {
	underlying := errors.New("test error")
	err := WrapErrorWithPath("spread", "parse", "Spreads/Spread_u1.xml", underlying)

	commonErr, ok := err.(*Error)
	if !ok {
		t.Fatal("WrapErrorWithPath did not return *Error")
	}
	if commonErr.Path != "Spreads/Spread_u1.xml" {
		t.Errorf("Path = %q, want %q", commonErr.Path, "Spreads/Spread_u1.xml")
	}

	// Test with nil error
	if WrapErrorWithPath("spread", "parse", "test.xml", nil) != nil {
		t.Error("WrapErrorWithPath should return nil for nil error")
	}
}

func TestErrorf_FormatsError(t *testing.T) {
	err := Errorf("resources", "parse", "Fonts.xml", "unexpected element: %s", "UnknownFont")

	expected := "resources: parse Fonts.xml: unexpected element: UnknownFont"
	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}
}

func TestSentinelErrors_WorkWithErrorsIs(t *testing.T) {
	// Test that sentinel errors work with errors.Is
	tests := []struct {
		name      string
		wrapped   error
		checkFunc func(error) bool
		expected  bool
	}{
		{
			name:      "IsNotFound with ErrNotFound",
			wrapped:   &Error{Package: "idml", Op: "get", Err: ErrNotFound},
			checkFunc: IsNotFound,
			expected:  true,
		},
		{
			name:      "IsNotFound with other error",
			wrapped:   &Error{Package: "idml", Op: "get", Err: errors.New("other")},
			checkFunc: IsNotFound,
			expected:  false,
		},
		{
			name:      "IsInvalidFormat with ErrInvalidFormat",
			wrapped:   &Error{Package: "spread", Op: "parse", Err: ErrInvalidFormat},
			checkFunc: IsInvalidFormat,
			expected:  true,
		},
		{
			name:      "IsAlreadyExists with ErrAlreadyExists",
			wrapped:   &Error{Package: "idml", Op: "add", Err: ErrAlreadyExists},
			checkFunc: IsAlreadyExists,
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.checkFunc(tt.wrapped); got != tt.expected {
				t.Errorf("check function returned %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestErrorChain_WorksCorrectly(t *testing.T) {
	// Test that error wrapping chain works correctly
	base := ErrNotFound
	level1 := &Error{Package: "story", Op: "get", Err: base}
	level2 := &Error{Package: "idml", Op: "story", Path: "Stories/Story_u1.xml", Err: level1}

	// Should be able to find ErrNotFound through the chain
	if !errors.Is(level2, ErrNotFound) {
		t.Error("errors.Is should find ErrNotFound through chain")
	}

	// Should be able to unwrap to level1
	var storyErr *Error
	if !errors.As(level2, &storyErr) {
		t.Error("errors.As should find *Error in chain")
	}
}
