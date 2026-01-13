package idms

import (
	"bytes"
	_ "embed"
	"fmt"
	"sync"
	"text/template"
)

// Template variables and functions are currently unused but kept for potential future use
// nolint:unused
//
//go:embed templates/snippet.xml
var snippet []byte

// Template initialization with sync.Once for thread-safe lazy loading
// nolint:unused
var (
	snippetTmpl     *template.Template
	snippetTmplErr  error
	snippetTmplOnce sync.Once
)

// getSnippetTemplate returns the parsed snippet template.
// The template is parsed once and cached for subsequent calls.
// This function is currently unused but kept for potential future use
// nolint:unused
func getSnippetTemplate() (*template.Template, error) {
	snippetTmplOnce.Do(func() {
		snippetTmpl, snippetTmplErr = template.New("snippet").Parse(string(snippet))
	})
	return snippetTmpl, snippetTmplErr
}

// generateSnippet creates a customized snippet based on options.
// This function is currently unused but kept for potential future use
// nolint:unused
func generateSnippet() ([]byte, error) {
	tmpl, err := getSnippetTemplate()
	if err != nil {
		return nil, fmt.Errorf("failed to parse snippet template: %w", err)
	}

	data := struct {
		DOMVersion string
	}{
		DOMVersion: "20.4",
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute snippet template: %w", err)
	}

	return buf.Bytes(), nil
}
