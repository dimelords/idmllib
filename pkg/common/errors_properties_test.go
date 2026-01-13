package common

import (
	"errors"
	"fmt"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestProperty1_ErrorConstructionConsistency tests that all errors
// are properly wrapped with context.
// **Feature: code-refactoring-improvements, Property 1: Error construction consistency**
func TestProperty1_ErrorConstructionConsistency(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Property: WrapError creates errors with package and operation context
	properties.Property("WrapError includes package and operation", prop.ForAll(
		func(pkg, op string, err error) bool {
			if pkg == "" || op == "" || err == nil {
				return true // Skip invalid inputs
			}

			wrapped := WrapError(pkg, op, err)
			if wrapped == nil {
				return false
			}

			// Check that the wrapped error contains the original error
			if !errors.Is(wrapped, err) {
				return false
			}

			// Check that we can extract a common.Error
			var commonErr *Error
			if !errors.As(wrapped, &commonErr) {
				return false
			}

			// Verify the fields are set correctly
			return commonErr.Package == pkg &&
				commonErr.Op == op &&
				commonErr.Path == "" &&
				errors.Is(commonErr.Err, err)
		},
		genPackageName(),
		genOperationName(),
		genError(),
	))

	// Property: WrapErrorWithPath creates errors with package, operation, and path context
	properties.Property("WrapErrorWithPath includes package, operation, and path", prop.ForAll(
		func(pkg, op, path string, err error) bool {
			if pkg == "" || op == "" || err == nil {
				return true // Skip invalid inputs
			}

			wrapped := WrapErrorWithPath(pkg, op, path, err)
			if wrapped == nil {
				return false
			}

			// Check that the wrapped error contains the original error
			if !errors.Is(wrapped, err) {
				return false
			}

			// Check that we can extract a common.Error
			var commonErr *Error
			if !errors.As(wrapped, &commonErr) {
				return false
			}

			// Verify the fields are set correctly
			return commonErr.Package == pkg &&
				commonErr.Op == op &&
				commonErr.Path == path &&
				errors.Is(commonErr.Err, err)
		},
		genPackageName(),
		genOperationName(),
		gen.AlphaString(),
		genError(),
	))

	// Property: Errorf creates errors with formatted messages
	properties.Property("Errorf creates formatted errors", prop.ForAll(
		func(pkg, op, path, format string, arg string) bool {
			if pkg == "" || op == "" || format == "" {
				return true // Skip invalid inputs
			}

			err := Errorf(pkg, op, path, format, arg)
			if err == nil {
				return false
			}

			// Check that we can extract a common.Error
			var commonErr *Error
			if !errors.As(err, &commonErr) {
				return false
			}

			// Verify the fields are set correctly
			return commonErr.Package == pkg &&
				commonErr.Op == op &&
				commonErr.Path == path &&
				commonErr.Err != nil
		},
		genPackageName(),
		genOperationName(),
		gen.AlphaString(),
		gen.Const("test format: %s"),
		gen.AlphaString(),
	))

	// Property: Error wrapping preserves error chains
	properties.Property("Error wrapping preserves error chains", prop.ForAll(
		func(pkg1, op1, pkg2, op2 string, baseErr error) bool {
			if pkg1 == "" || op1 == "" || pkg2 == "" || op2 == "" || baseErr == nil {
				return true // Skip invalid inputs
			}

			// Create a chain: baseErr -> level1 -> level2
			level1 := WrapError(pkg1, op1, baseErr)
			level2 := WrapError(pkg2, op2, level1)

			// Should be able to find the base error through the chain
			if !errors.Is(level2, baseErr) {
				return false
			}

			// Should be able to find intermediate errors
			if !errors.Is(level2, level1) {
				return false
			}

			// Should be able to extract common.Error from any level
			var commonErr *Error
			return errors.As(level2, &commonErr) && errors.As(level1, &commonErr)
		},
		genPackageName(),
		genOperationName(),
		genPackageName(),
		genOperationName(),
		genError(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty7_ErrorWrappingFollowsGoConventions tests that wrapped errors
// support Go's standard error wrapping conventions.
// **Feature: code-refactoring-improvements, Property 7: Error wrapping follows Go conventions**
func TestProperty7_ErrorWrappingFollowsGoConventions(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Property: Wrapped errors support errors.Unwrap
	properties.Property("wrapped errors support errors.Unwrap", prop.ForAll(
		func(pkg, op, path string, err error) bool {
			if pkg == "" || op == "" || err == nil {
				return true // Skip invalid inputs
			}

			// Test WrapError
			wrapped := WrapError(pkg, op, err)
			if wrapped == nil {
				return false
			}

			// Should be able to unwrap to get the original error
			unwrapped := errors.Unwrap(wrapped)
			if unwrapped != err {
				return false
			}

			// Test WrapErrorWithPath
			wrappedWithPath := WrapErrorWithPath(pkg, op, path, err)
			if wrappedWithPath == nil {
				return false
			}

			// Should be able to unwrap to get the original error
			unwrappedWithPath := errors.Unwrap(wrappedWithPath)
			return unwrappedWithPath == err
		},
		genPackageName(),
		genOperationName(),
		gen.AlphaString(),
		genError(),
	))

	// Property: errors.Is works with sentinel errors
	properties.Property("errors.Is works with sentinel errors", prop.ForAll(
		func(pkg, op, path string, sentinelErr error) bool {
			if pkg == "" || op == "" {
				return true // Skip invalid inputs
			}

			// Wrap the sentinel error
			wrapped := WrapErrorWithPath(pkg, op, path, sentinelErr)
			if wrapped == nil {
				return false
			}

			// errors.Is should find the sentinel error through the wrapper
			if !errors.Is(wrapped, sentinelErr) {
				return false
			}

			// Test with multiple levels of wrapping
			doubleWrapped := WrapError("outer", "operation", wrapped)
			if doubleWrapped == nil {
				return false
			}

			// Should still find the original sentinel error
			return errors.Is(doubleWrapped, sentinelErr)
		},
		genPackageName(),
		genOperationName(),
		gen.AlphaString(),
		genSentinelError(),
	))

	// Property: errors.As works with common.Error type
	properties.Property("errors.As works with common.Error type", prop.ForAll(
		func(pkg, op, path string, err error) bool {
			if pkg == "" || op == "" || err == nil {
				return true // Skip invalid inputs
			}

			// Create wrapped error
			wrapped := WrapErrorWithPath(pkg, op, path, err)
			if wrapped == nil {
				return false
			}

			// Should be able to extract common.Error using errors.As
			var commonErr *Error
			if !errors.As(wrapped, &commonErr) {
				return false
			}

			// Verify the extracted error has correct fields
			if commonErr.Package != pkg || commonErr.Op != op || commonErr.Path != path {
				return false
			}

			// Test with multiple levels of wrapping
			doubleWrapped := WrapError("outer", "operation", wrapped)
			if doubleWrapped == nil {
				return false
			}

			// Should be able to extract both Error instances
			var outerErr *Error
			var innerErr *Error
			return errors.As(doubleWrapped, &outerErr) && errors.As(wrapped, &innerErr)
		},
		genPackageName(),
		genOperationName(),
		gen.AlphaString(),
		genError(),
	))

	// Property: Error chains preserve unwrapping behavior
	properties.Property("error chains preserve unwrapping behavior", prop.ForAll(
		func(pkg1, op1, pkg2, op2 string, baseErr error) bool {
			if pkg1 == "" || op1 == "" || pkg2 == "" || op2 == "" || baseErr == nil {
				return true // Skip invalid inputs
			}

			// Create a chain: baseErr -> level1 -> level2
			level1 := WrapError(pkg1, op1, baseErr)
			level2 := WrapError(pkg2, op2, level1)

			// Direct unwrap should give us level1
			if errors.Unwrap(level2) != level1 {
				return false
			}

			// Unwrapping level1 should give us baseErr
			if errors.Unwrap(level1) != baseErr {
				return false
			}

			// errors.Is should work through the entire chain
			if !errors.Is(level2, baseErr) {
				return false
			}

			// errors.As should work at each level
			var err1, err2 *Error
			return errors.As(level1, &err1) && errors.As(level2, &err2)
		},
		genPackageName(),
		genOperationName(),
		genPackageName(),
		genOperationName(),
		genError(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// genSentinelError generates sentinel errors for testing
func genSentinelError() gopter.Gen {
	return gen.OneConstOf(
		ErrNotFound,
		ErrInvalidFormat,
		ErrAlreadyExists,
		ErrMissingDependency,
		ErrMissingMetadata,
	)
}

// genPackageName generates valid package names for IDML packages
func genPackageName() gopter.Gen {
	return gen.OneConstOf(
		"idml",
		"idms",
		"document",
		"spread",
		"story",
		"resources",
		"analysis",
		"common",
	)
}

// genOperationName generates valid operation names
func genOperationName() gopter.Gen {
	return gen.OneConstOf(
		"read",
		"write",
		"parse",
		"marshal",
		"validate",
		"export",
		"import",
		"select",
		"add",
		"remove",
		"update",
	)
}

// genError generates various types of errors for testing
func genError() gopter.Gen {
	return gen.OneConstOf(
		errors.New("test error"),
		fmt.Errorf("formatted error: %s", "test"),
		ErrNotFound,
		ErrInvalidFormat,
		ErrAlreadyExists,
		ErrMissingDependency,
		ErrMissingMetadata,
	)
}
