package xmlutil

import (
	"testing"
)

// TestCompactEmptyElements tests converting empty XML elements to self-closing tags.
func TestCompactEmptyElements(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple empty element",
			input:    `<KeyValuePair Key="test" Value="1"></KeyValuePair>`,
			expected: `<KeyValuePair Key="test" Value="1" />`,
		},
		{
			name:     "empty element with no attributes",
			input:    `<Element></Element>`,
			expected: `<Element/>`,
		},
		{
			name:     "empty element with trailing space in attributes",
			input:    `<Tag attr="value" ></Tag>`,
			expected: `<Tag attr="value" />`,
		},
		{
			name:     "non-empty element should not be converted",
			input:    `<Tag>Content</Tag>`,
			expected: `<Tag>Content</Tag>`,
		},
		{
			name:     "mismatched tags should not be converted",
			input:    `<OpenTag></CloseTag>`,
			expected: `<OpenTag></CloseTag>`,
		},
		{
			name:     "multiple empty elements",
			input:    `<First></First><Second></Second>`,
			expected: `<First/><Second/>`,
		},
		{
			name:     "nested structure with empty elements",
			input:    `<Parent><Child></Child></Parent>`,
			expected: `<Parent><Child/></Parent>`,
		},
		{
			name:     "already self-closing tag",
			input:    `<Tag />`,
			expected: `<Tag />`,
		},
		{
			name:     "complex attributes",
			input:    `<Element id="123" class="test" data-value="abc"></Element>`,
			expected: `<Element id="123" class="test" data-value="abc" />`,
		},
		{
			name:     "empty string",
			input:    ``,
			expected: ``,
		},
		{
			name:     "no empty elements",
			input:    `<Root><Child>Text</Child><Other>More</Other></Root>`,
			expected: `<Root><Child>Text</Child><Other>More</Other></Root>`,
		},
		{
			name:     "mixed empty and non-empty",
			input:    `<Root><Empty></Empty><Full>Content</Full><AlsoEmpty></AlsoEmpty></Root>`,
			expected: `<Root><Empty/><Full>Content</Full><AlsoEmpty/></Root>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompactEmptyElements([]byte(tt.input))
			if string(result) != tt.expected {
				t.Errorf("CompactEmptyElements() =\n%q\nwant\n%q", string(result), tt.expected)
			}
		})
	}
}

// TestCompactEmptyElements_RealWorldIDML tests with actual IDML-like structures.
func TestCompactEmptyElements_RealWorldIDML(t *testing.T) {
	input := `<Story><ParagraphStyleRange AppliedParagraphStyle="ParagraphStyle/Normal"></ParagraphStyleRange><CharacterStyleRange AppliedCharacterStyle="CharacterStyle/$ID/[No character style]"><Content>Hello</Content></CharacterStyleRange></Story>`
	expected := `<Story><ParagraphStyleRange AppliedParagraphStyle="ParagraphStyle/Normal" /><CharacterStyleRange AppliedCharacterStyle="CharacterStyle/$ID/[No character style]"><Content>Hello</Content></CharacterStyleRange></Story>`

	result := CompactEmptyElements([]byte(input))
	if string(result) != expected {
		t.Errorf("CompactEmptyElements() with IDML structure failed\ngot:\n%s\nwant:\n%s", string(result), expected)
	}
}

// TestCompactEmptyElements_PreservesNonMatching tests that non-matching patterns are preserved.
func TestCompactEmptyElements_PreservesNonMatching(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "self-closing tags preserved",
			input: `<Element attr="value" />`,
		},
		{
			name:  "tags with whitespace content preserved",
			input: `<Element> </Element>`,
		},
		{
			name:  "tags with newline preserved",
			input: "<Element>\n</Element>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompactEmptyElements([]byte(tt.input))
			if string(result) != tt.input {
				t.Errorf("CompactEmptyElements() modified input that should be preserved\ngot:  %q\nwant: %q", string(result), tt.input)
			}
		})
	}
}
