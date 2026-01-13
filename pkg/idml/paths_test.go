package idml

import "testing"

func TestStoryPath_GeneratesCorrectPaths(t *testing.T) {
	tests := []struct {
		id   string
		want string
	}{
		{"u1d8", "Stories/Story_u1d8.xml"},
		{"abc", "Stories/Story_abc.xml"},
		{"123", "Stories/Story_123.xml"},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			if got := StoryPath(tt.id); got != tt.want {
				t.Errorf("StoryPath(%q) = %q, want %q", tt.id, got, tt.want)
			}
		})
	}
}

func TestSpreadPath_GeneratesCorrectPaths(t *testing.T) {
	tests := []struct {
		id   string
		want string
	}{
		{"u210", "Spreads/Spread_u210.xml"},
		{"ub6", "Spreads/Spread_ub6.xml"},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			if got := SpreadPath(tt.id); got != tt.want {
				t.Errorf("SpreadPath(%q) = %q, want %q", tt.id, got, tt.want)
			}
		})
	}
}

func TestMasterSpreadPath_GeneratesCorrectPath(t *testing.T) {
	if got := MasterSpreadPath("ub4"); got != "MasterSpreads/MasterSpread_ub4.xml" {
		t.Errorf("MasterSpreadPath(\"ub4\") = %q, want %q", got, "MasterSpreads/MasterSpread_ub4.xml")
	}
}

func TestIsStoryPath_IdentifiesStoryPaths(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"Stories/Story_u1d8.xml", true},
		{"Stories/Story_abc.xml", true},
		{"Stories/other.xml", true},
		{"Spreads/Spread_u210.xml", false},
		{"Resources/Fonts.xml", false},
		{"designmap.xml", false},
		{"Stories/", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := IsStoryPath(tt.path); got != tt.want {
				t.Errorf("IsStoryPath(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestIsSpreadPath_IdentifiesSpreadPaths(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"Spreads/Spread_u210.xml", true},
		{"Spreads/Spread_ub6.xml", true},
		{"Stories/Story_u1d8.xml", false},
		{"Resources/Fonts.xml", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := IsSpreadPath(tt.path); got != tt.want {
				t.Errorf("IsSpreadPath(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestIsResourcePath_IdentifiesResourcePaths(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"Resources/Fonts.xml", true},
		{"Resources/Styles.xml", true},
		{"Resources/Graphic.xml", true},
		{"Stories/Story_u1d8.xml", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := IsResourcePath(tt.path); got != tt.want {
				t.Errorf("IsResourcePath(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestIsMetaInfPath_IdentifiesMetaInfPaths(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"META-INF/container.xml", true},
		{"META-INF/metadata.xml", true},
		{"XML/Tags.xml", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := IsMetaInfPath(tt.path); got != tt.want {
				t.Errorf("IsMetaInfPath(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestIsXMLPath_IdentifiesXMLPaths(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"XML/Tags.xml", true},
		{"XML/BackingStory.xml", true},
		{"META-INF/container.xml", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := IsXMLPath(tt.path); got != tt.want {
				t.Errorf("IsXMLPath(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestPathConstants_HaveExpectedValues(t *testing.T) {
	// Verify constants have expected values
	tests := []struct {
		name     string
		constant string
		want     string
	}{
		{"PathMimetype", PathMimetype, "mimetype"},
		{"PathDesignmap", PathDesignmap, "designmap.xml"},
		{"PathFonts", PathFonts, "Resources/Fonts.xml"},
		{"PathStyles", PathStyles, "Resources/Styles.xml"},
		{"PathGraphic", PathGraphic, "Resources/Graphic.xml"},
		{"PathPreferences", PathPreferences, "Resources/Preferences.xml"},
		{"PathContainer", PathContainer, "META-INF/container.xml"},
		{"PathTags", PathTags, "XML/Tags.xml"},
		{"PathMasterSpread", PathMasterSpread, "MasterSpreads/MasterSpread_ub4.xml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.constant, tt.want)
			}
		})
	}
}
