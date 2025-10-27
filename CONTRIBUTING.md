# Contributing to IDML Library

Thank you for your interest in contributing to idmllib! We appreciate your help in making this library better.

## Code of Conduct

By participating in this project, you are expected to uphold our standards of respectful and professional communication.

## Development Setup

### Prerequisites

- Go 1.23 or later
- Git
- golangci-lint (for linting)

### Getting Started

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/idmllib.git
   cd idmllib
   ```

3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/dimelords/idmllib.git
   ```

4. Install dependencies:
   ```bash
   go mod download
   ```

5. Make sure tests pass:
   ```bash
   make test
   ```

## Development Workflow

### Creating a Feature Branch

```bash
git checkout -b feature/your-feature-name
```

Use descriptive branch names:
- `feature/` for new features
- `fix/` for bug fixes
- `docs/` for documentation updates
- `test/` for test improvements
- `refactor/` for code refactoring

### Making Changes

1. **Write Tests**: Add tests for new functionality or bug fixes
   ```bash
   # Run tests as you develop
   make test
   
   # Check coverage
   make test-coverage
   ```

2. **Follow Go Conventions**:
   - Run `gofmt` on your code
   - Use meaningful variable and function names
   - Add comments for exported functions
   - Keep functions focused and small

3. **Lint Your Code**:
   ```bash
   make lint
   ```

4. **Format Your Code**:
   ```bash
   make fmt
   ```

### Writing Good Commit Messages

- Use the present tense ("Add feature" not "Added feature")
- Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
- Limit the first line to 72 characters or less
- Reference issues and pull requests liberally after the first line

Examples:
```
Add ColorGroup filtering support

Fix panic when reading corrupted IDML files
Closes #42

Improve test coverage for parser_filter.go
```

## Pull Request Process

1. **Update Documentation**: Update README.md if you're adding new features
2. **Add Tests**: Ensure your code is well-tested
3. **Pass All Checks**: Make sure all CI checks pass
4. **Update CHANGELOG**: Add a note about your changes (if applicable)
5. **Request Review**: Tag maintainers for review

### Pull Request Checklist

- [ ] Tests pass locally (`make test`)
- [ ] Linter passes (`make lint`)
- [ ] Code is formatted (`make fmt`)
- [ ] Documentation is updated
- [ ] Commit messages are clear and descriptive
- [ ] Changes are focused and logical

## Testing Guidelines

### Writing Tests

- Place tests in `*_test.go` files next to the code they test
- Use table-driven tests for multiple similar test cases
- Test both success and error cases
- Use descriptive test names: `TestFunctionName_Scenario`

Example:
```go
func TestExportStoryAsIDMS_NonExistentStory(t *testing.T) {
    pkg, err := Open("testdata/example.idml")
    if err != nil {
        t.Fatal(err)
    }
    defer pkg.Close()
    
    err = pkg.ExportStoryAsIDMS("nonexistent", "output.idms")
    if err == nil {
        t.Error("Expected error for non-existent story")
    }
}
```

### Test Data

If you need to add test data:
- Use small, minimal IDML files
- Place them in `idml/testdata/`
- Document what the test file contains
- Avoid large files (keep under 100KB if possible)

## Code Style Guidelines

### Go Code

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Add godoc comments for all exported functions
- Keep cyclomatic complexity low
- Avoid unnecessary dependencies

### Documentation Comments

All exported functions should have comments:

```go
// ExportStoryAsIDMS exports a story as an InDesign Snippet (IDMS) file.
// The story is identified by its ID (e.g., "u222") and exported to the
// specified output path.
//
// The exported IDMS file includes:
//   - The story content with all formatting
//   - Only the styles actually used
//   - Only the colors and swatches referenced
//   - Only the TextFrames linked to the story
//
// Returns an error if the story doesn't exist or the export fails.
func (p *Package) ExportStoryAsIDMS(storyID, outputPath string) error {
    // ...
}
```

## Reporting Bugs

### Before Submitting

- Check if the issue has already been reported
- Try to reproduce with the latest version
- Gather relevant information (IDML file structure, error messages, etc.)

### Creating a Bug Report

A good bug report should include:

- **Description**: Clear description of the issue
- **Steps to Reproduce**: 
  1. First step
  2. Second step
  3. ...
- **Expected Behavior**: What you expected to happen
- **Actual Behavior**: What actually happened
- **Environment**:
  - OS (macOS, Windows, Linux)
  - Go version
  - idmllib version
- **Sample Files**: If possible, attach a minimal IDML file that reproduces the issue

## Suggesting Features

We love feature suggestions! Before submitting:

1. Check if it's already been suggested
2. Make sure it fits the project scope
3. Consider if it could be implemented as a separate package

Include in your suggestion:
- **Use case**: Why this feature would be useful
- **Proposed solution**: How you envision it working
- **Alternatives**: Other ways to achieve the same goal

## Questions?

- Open a discussion on GitHub
- Check existing issues and pull requests
- Review the documentation

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

Thank you for contributing to idmllib! 🎉
