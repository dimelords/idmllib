# Code Quality and Linting Setup

This document describes the linting and code quality setup for the idmlbuild project.

## Overview

The project uses `golangci-lint` for comprehensive code quality checks. The configuration balances code quality enforcement with compatibility for the existing codebase.

## Configuration

### Linting Configuration (`.golangci.yml`)

The linting configuration is designed to:
- Catch critical errors and bugs
- Enforce essential Go best practices
- Be permissive for existing code patterns
- Maintain compatibility with the current codebase

#### Enabled Linters

**Critical Error Checking:**
- `errcheck` - Check for unchecked errors (with exclusions)
- `gosec` - Security-focused linter (with exclusions)

**Code Correctness:**
- `gosimple` - Simplify code
- `ineffassign` - Detect ineffectual assignments
- `unused` - Check for unused constants, variables, functions and types
- `govet` - Go vet (built-in static analyzer)
- `staticcheck` - Comprehensive static analysis

**Code Formatting:**
- `gofmt` - Check formatting
- `goimports` - Check imports formatting and organization
- `misspell` - Check for misspelled words in comments and strings

**Type Checking:**
- `typecheck` - Type checking (like go build)

#### Disabled Linters

The following linters are disabled for existing codebase compatibility:
- `gocritic` - Too many style suggestions for existing code
- `revive` - Too strict for existing code patterns
- `bodyclose` - Not applicable to this codebase
- `noctx` - Too strict for existing code
- `copyloopvar` - May be too strict for existing code

### Exclusions and Exceptions

The configuration includes several exclusions to handle existing code patterns:

1. **Test Files**: More permissive rules for `*_test.go` files
2. **CLI/TUI Code**: Relaxed error checking for user interface code
3. **Utility Functions**: Exceptions for cache management and I/O utilities
4. **XML Processing**: Special handling for XML-related functions
5. **Generated Code**: Exclusions for generated files

## Usage

### Running Linting Locally

```bash
# Run all linters
golangci-lint run

# Run with specific config
golangci-lint run --config=.golangci.yml

# Run on specific files/directories
golangci-lint run ./pkg/...

# Show only new issues
golangci-lint run --new-from-rev=HEAD~1
```

### IDE Integration

Most Go IDEs support golangci-lint integration:

**VS Code:**
- Install the Go extension
- Configure `go.lintTool` to `golangci-lint`

**GoLand/IntelliJ:**
- Enable golangci-lint in Settings → Go → Linter

### CI/CD Integration

Linting runs automatically in GitHub Actions on:
- Every push to `main` and `develop` branches
- Every pull request
- Manual workflow dispatch

The CI pipeline will fail if linting violations are found.

## Installation

### Local Installation

```bash
# Install golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2

# Or using go install
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2
```

### Verification

```bash
# Verify installation
golangci-lint --version

# Run linting
golangci-lint run
```

## Fixing Common Issues

### Unused Variables/Functions

For intentionally unused code that should be kept:
```go
// nolint:unused
func helperFunction() {
    // ...
}
```

### Error Checking

For cases where error checking is intentionally skipped:
```go
xml.EscapeText(&buf, data) // nolint:errcheck
```

### Complex Functions

For functions that exceed complexity limits but cannot be simplified:
```go
// nolint:gocyclo
func complexFunction() {
    // ...
}
```

## Maintenance

### Updating Configuration

When updating the linting configuration:

1. Test changes locally first
2. Consider impact on existing code
3. Update this documentation
4. Ensure CI pipeline still passes

### Adding New Linters

When adding new linters:

1. Start with warnings only
2. Fix existing violations gradually
3. Add appropriate exclusions for legacy code
4. Document the rationale

## Implementation Notes

### Phase 8 Implementation

The current configuration was implemented as part of Phase 8 of the code refactoring project. Key decisions made:

1. **Balanced Approach**: Prioritized catching real bugs over style enforcement
2. **Existing Code Compatibility**: Made configuration permissive for current codebase
3. **Gradual Improvement**: Set foundation for future code quality improvements
4. **CI Integration**: Ensured linting runs on all code changes

### Fixed Issues

During implementation, the following issues were resolved:
- Removed unused struct fields and variables
- Fixed append operations with no values
- Added nolint comments for intentionally unused code
- Corrected error handling patterns
- Formatted code to meet gofmt standards

## Future Improvements

Consider these improvements for future phases:

1. **Stricter Rules**: Gradually enable more linters as code quality improves
2. **Custom Rules**: Add project-specific linting rules
3. **Performance Linting**: Enable performance-focused linters
4. **Documentation Linting**: Enforce documentation standards
5. **Security Hardening**: Enable additional security linters

## Troubleshooting

### Common Issues

**Linting Fails in CI but Passes Locally:**
- Ensure you're using the same golangci-lint version
- Check for platform-specific issues
- Verify configuration file is committed

**Too Many False Positives:**
- Add specific exclusions to `.golangci.yml`
- Use nolint comments for individual cases
- Consider adjusting linter sensitivity

**Performance Issues:**
- Increase timeout in configuration
- Exclude large directories if needed
- Use `--fast` flag for quicker runs

For more information, see the [golangci-lint documentation](https://golangci-lint.run/).