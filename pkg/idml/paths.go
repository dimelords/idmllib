package idml

// IDML package file paths and prefixes.
// These constants define the standard paths used within IDML packages.

const (
	// Root files
	PathMimetype  = "mimetype"
	PathDesignmap = "designmap.xml"

	// Resource files
	PathFonts       = "Resources/Fonts.xml"
	PathStyles      = "Resources/Styles.xml"
	PathGraphic     = "Resources/Graphic.xml"
	PathPreferences = "Resources/Preferences.xml"

	// Metadata files
	PathContainer    = "META-INF/container.xml"
	PathTags         = "XML/Tags.xml"
	PathBackingStory = "XML/BackingStory.xml"

	// Master spread template
	PathMasterSpread = "MasterSpreads/MasterSpread_ub4.xml"

	// Directory prefixes (with trailing slash)
	PrefixStories       = "Stories/"
	PrefixSpreads       = "Spreads/"
	PrefixMasterSpreads = "MasterSpreads/"
	PrefixResources     = "Resources/"
	PrefixMetaInf       = "META-INF/"
	PrefixXML           = "XML/"

	// File extension
	ExtXML = ".xml"
)

// StoryPath returns the standard path for a story file.
// Example: StoryPath("u1d8") returns "Stories/Story_u1d8.xml"
func StoryPath(id string) string {
	return PrefixStories + "Story_" + id + ExtXML
}

// SpreadPath returns the standard path for a spread file.
// Example: SpreadPath("u210") returns "Spreads/Spread_u210.xml"
func SpreadPath(id string) string {
	return PrefixSpreads + "Spread_" + id + ExtXML
}

// MasterSpreadPath returns the standard path for a master spread file.
// Example: MasterSpreadPath("ub4") returns "MasterSpreads/MasterSpread_ub4.xml"
func MasterSpreadPath(id string) string {
	return PrefixMasterSpreads + "MasterSpread_" + id + ExtXML
}

// IsStoryPath checks if a path is in the Stories directory.
func IsStoryPath(path string) bool {
	return len(path) > len(PrefixStories) &&
		path[:len(PrefixStories)] == PrefixStories &&
		len(path) > 4 && path[len(path)-4:] == ExtXML
}

// IsSpreadPath checks if a path is in the Spreads directory.
func IsSpreadPath(path string) bool {
	return len(path) > len(PrefixSpreads) &&
		path[:len(PrefixSpreads)] == PrefixSpreads &&
		len(path) > 4 && path[len(path)-4:] == ExtXML
}

// IsResourcePath checks if a path is in the Resources directory.
func IsResourcePath(path string) bool {
	return len(path) > len(PrefixResources) &&
		path[:len(PrefixResources)] == PrefixResources &&
		len(path) > 4 && path[len(path)-4:] == ExtXML
}

// IsMetaInfPath checks if a path is in the META-INF directory.
func IsMetaInfPath(path string) bool {
	return len(path) > len(PrefixMetaInf) &&
		path[:len(PrefixMetaInf)] == PrefixMetaInf
}

// IsXMLPath checks if a path is in the XML directory.
func IsXMLPath(path string) bool {
	return len(path) > len(PrefixXML) &&
		path[:len(PrefixXML)] == PrefixXML
}
