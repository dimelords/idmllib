package common

import (
	"encoding/xml"
	"testing"
)

// TestProperties_GetAppliedFont tests extracting AppliedFont from Properties.
func TestProperties_GetAppliedFont(t *testing.T) {
	tests := []struct {
		name     string
		xml      string
		expected string
	}{
		{
			name: "AppliedFont with simple text",
			xml: `<Properties>
				<AppliedFont type="string">Polaris Condensed</AppliedFont>
			</Properties>`,
			expected: "Polaris Condensed",
		},
		{
			name: "AppliedFont with different font",
			xml: `<Properties>
				<BasedOn type="string">$ID/[No character style]</BasedOn>
				<AppliedFont type="string">Kepler Std</AppliedFont>
			</Properties>`,
			expected: "Kepler Std",
		},
		{
			name: "No AppliedFont",
			xml: `<Properties>
				<BasedOn type="string">$ID/[No character style]</BasedOn>
			</Properties>`,
			expected: "",
		},
		{
			name: "Empty AppliedFont",
			xml: `<Properties>
				<AppliedFont type="string"></AppliedFont>
			</Properties>`,
			expected: "",
		},
		{
			name:     "Nil properties",
			xml:      "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var props *Properties

			if tt.xml != "" {
				props = &Properties{}
				if err := xml.Unmarshal([]byte(tt.xml), props); err != nil {
					t.Fatalf("Failed to unmarshal XML: %v", err)
				}
			}

			got := props.GetAppliedFont()
			if got != tt.expected {
				t.Errorf("GetAppliedFont() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestProperties_GetBasedOn tests extracting BasedOn from Properties.
func TestProperties_GetBasedOn(t *testing.T) {
	tests := []struct {
		name     string
		xml      string
		expected string
	}{
		{
			name: "BasedOn with built-in style",
			xml: `<Properties>
				<BasedOn type="string">$ID/[No character style]</BasedOn>
			</Properties>`,
			expected: "$ID/[No character style]",
		},
		{
			name: "BasedOn with custom style",
			xml: `<Properties>
				<BasedOn type="string">CharacterStyle/MyBaseStyle</BasedOn>
				<AppliedFont type="string">Kepler Std</AppliedFont>
			</Properties>`,
			expected: "CharacterStyle/MyBaseStyle",
		},
		{
			name: "No BasedOn",
			xml: `<Properties>
				<AppliedFont type="string">Polaris Condensed</AppliedFont>
			</Properties>`,
			expected: "",
		},
		{
			name: "Empty BasedOn",
			xml: `<Properties>
				<BasedOn type="string"></BasedOn>
			</Properties>`,
			expected: "",
		},
		{
			name:     "Nil properties",
			xml:      "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var props *Properties

			if tt.xml != "" {
				props = &Properties{}
				if err := xml.Unmarshal([]byte(tt.xml), props); err != nil {
					t.Fatalf("Failed to unmarshal XML: %v", err)
				}
			}

			got := props.GetBasedOn()
			if got != tt.expected {
				t.Errorf("GetBasedOn() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestProperties_GetAppliedFont_RealWorld tests with actual IDML XML structure.
func TestProperties_GetAppliedFont_RealWorld(t *testing.T) {
	// Real XML from IDML file (typical CharacterStyle Properties)
	realXML := `<Properties>
		<BasedOn type="string">$ID/[No character style]</BasedOn>
		<PreviewColor type="enumeration">Nothing</PreviewColor>
		<Leading type="unit">11</Leading>
		<AppliedFont type="string">Polaris Condensed</AppliedFont>
	</Properties>`

	props := &Properties{}
	if err := xml.Unmarshal([]byte(realXML), props); err != nil {
		t.Fatalf("Failed to unmarshal real XML: %v", err)
	}

	// Test GetAppliedFont
	font := props.GetAppliedFont()
	if font != "Polaris Condensed" {
		t.Errorf("GetAppliedFont() = %q, want %q", font, "Polaris Condensed")
	}

	// Test GetBasedOn
	basedOn := props.GetBasedOn()
	if basedOn != "$ID/[No character style]" {
		t.Errorf("GetBasedOn() = %q, want %q", basedOn, "$ID/[No character style]")
	}
}
