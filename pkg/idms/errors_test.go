package idms

import (
	"errors"
	"testing"

	"github.com/dimelords/idmllib/pkg/common"
)

// TestError_Error tests the Error() method of common.Error
func TestError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "simple error",
			err:      common.WrapError("idms", "export selection", errors.New("no items selected")),
			expected: "idms: export selection: no items selected",
		},
		{
			name:     "nested error",
			err:      common.WrapError("idms", "build package", common.WrapError("idms", "create spread", errors.New("invalid ID"))),
			expected: "idms: build package: idms: create spread: invalid ID",
		},
		{
			name:     "with path",
			err:      common.WrapErrorWithPath("idms", "read", "snippet.idms", errors.New("file not found")),
			expected: "idms: read snippet.idms: file not found",
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

// TestError_Unwrap tests the Unwrap() method of idms.Error
func TestError_Unwrap(t *testing.T) {
	innerErr := errors.New("inner error")
	err := &common.Error{
		Package: "idms",
		Op:      "test operation",
		Err:     innerErr,
	}

	unwrapped := err.Unwrap()
	if unwrapped != innerErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, innerErr)
	}
}

// TestError_IsError tests using errors.Is with common.Error
func TestError_IsError(t *testing.T) {
	innerErr := errors.New("specific error")
	err := common.WrapError("idms", "test operation", innerErr)

	if !errors.Is(err, innerErr) {
		t.Error("errors.Is() should find the inner error")
	}

	differentErr := errors.New("different error")
	if errors.Is(err, differentErr) {
		t.Error("errors.Is() should not match a different error")
	}
}

// TestError_Chaining tests error chaining with multiple levels
func TestError_Chaining(t *testing.T) {
	baseErr := errors.New("base error")
	err1 := common.WrapError("idms", "level 1", baseErr)
	err2 := common.WrapError("idms", "level 2", err1)
	err3 := common.WrapError("idms", "level 3", err2)

	// Should be able to unwrap all the way down
	if !errors.Is(err3, baseErr) {
		t.Error("errors.Is() should find the base error through multiple levels")
	}

	// Check the error message contains all operations
	errMsg := err3.Error()
	if !contains(errMsg, "level 3") || !contains(errMsg, "level 2") || !contains(errMsg, "level 1") {
		t.Errorf("Error message should contain all operation names, got: %q", errMsg)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
