package common

import (
	"errors"
	"fmt"
)

// Common sentinel errors shared across all packages.
var (
	// ErrNotFound is returned when a requested item doesn't exist.
	ErrNotFound = errors.New("not found")

	// ErrInvalidFormat is returned when XML or data has invalid structure.
	ErrInvalidFormat = errors.New("invalid format")

	// ErrAlreadyExists is returned when attempting to add an item that already exists.
	ErrAlreadyExists = errors.New("already exists")

	// ErrMissingDependency is returned when a required resource is missing.
	ErrMissingDependency = errors.New("missing required dependency")

	// ErrMissingMetadata is returned when required metadata is missing.
	ErrMissingMetadata = errors.New("missing required metadata")
)

// Error represents an operation error with context.
// This is the unified error type used across all IDML packages.
//
// DESIGN DECISION: Structured Error Context
// Instead of using plain errors or multiple error types per package, we use a single
// Error struct that captures context about where and how the error occurred.
// This provides consistent error handling across all packages while maintaining
// compatibility with Go's error wrapping patterns (errors.Is, errors.As, errors.Unwrap).
// The structured approach enables better debugging and error reporting in complex
// operations that span multiple packages and files.
type Error struct {
	// Package identifies the package where the error originated.
	// Examples: "idml", "idms", "spread", "story", "resources"
	Package string

	// Op describes the operation being performed when the error occurred.
	// Examples: "read", "write", "parse", "marshal"
	Op string

	// Path is the file or resource path involved, if applicable.
	// May be empty if no specific path is involved.
	Path string

	// Err is the underlying error that caused this error.
	Err error
}

// Error implements the error interface with a consistent format.
func (e *Error) Error() string {
	// Format: "package: op [path]: underlying error"
	var msg string
	if e.Package != "" {
		msg = e.Package + ": "
	}
	if e.Op != "" {
		msg += e.Op
	}
	if e.Path != "" {
		msg += " " + e.Path
	}
	if e.Err != nil {
		if msg != "" {
			msg += ": "
		}
		msg += e.Err.Error()
	}
	return msg
}

// Unwrap returns the underlying error for use with errors.Is/As.
func (e *Error) Unwrap() error {
	return e.Err
}

// NewError creates a new Error with the given parameters.
// This is a convenience constructor for creating errors.
func NewError(pkg, op, path string, err error) *Error {
	return &Error{
		Package: pkg,
		Op:      op,
		Path:    path,
		Err:     err,
	}
}

// WrapError wraps an existing error with package and operation context.
// If err is nil, returns nil.
func WrapError(pkg, op string, err error) error {
	if err == nil {
		return nil
	}
	return &Error{
		Package: pkg,
		Op:      op,
		Err:     err,
	}
}

// WrapErrorWithPath wraps an existing error with package, operation, and path context.
// If err is nil, returns nil.
func WrapErrorWithPath(pkg, op, path string, err error) error {
	if err == nil {
		return nil
	}
	return &Error{
		Package: pkg,
		Op:      op,
		Path:    path,
		Err:     err,
	}
}

// Errorf creates a new Error with a formatted message as the underlying error.
func Errorf(pkg, op, path, format string, args ...interface{}) *Error {
	return &Error{
		Package: pkg,
		Op:      op,
		Path:    path,
		Err:     fmt.Errorf(format, args...),
	}
}

// IsNotFound checks if an error is or wraps ErrNotFound.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsInvalidFormat checks if an error is or wraps ErrInvalidFormat.
func IsInvalidFormat(err error) bool {
	return errors.Is(err, ErrInvalidFormat)
}

// IsAlreadyExists checks if an error is or wraps ErrAlreadyExists.
func IsAlreadyExists(err error) bool {
	return errors.Is(err, ErrAlreadyExists)
}
