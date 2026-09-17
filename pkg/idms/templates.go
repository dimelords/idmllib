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

// getSnippetTemplate parses the embedded snippet template once, on first use.
// It is currently unused but kept for planned snippet generation.
var getSnippetTemplate = sync.OnceValues(func() (*template.Template, error) { //nolint:unused
	return template.New("snippet").Parse(string(snippet))
})

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
