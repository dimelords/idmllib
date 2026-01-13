// Package document provides types for InDesign document structure (designmap.xml).
//
// This package contains the Document type and all related types for representing
// the main document manifest, including layers, sections, languages, grids,
// color groups, text variables, and document-level settings.
//
// For IDMS (snippet) files, the Document type also supports inline resources
// and content that would normally be in separate files in IDML format.
package document

import (
	"encoding/xml"

	"github.com/dimelords/idmllib/pkg/common"
	"github.com/dimelords/idmllib/pkg/resources"
	"github.com/dimelords/idmllib/pkg/spread"
	"github.com/dimelords/idmllib/pkg/story"
)

// Document represents the root element of designmap.xml.
// This is the main manifest file for an IDML package.
//
// Phase 2 Implementation: Full parsing of all Document attributes and
// major child elements while maintaining forward compatibility.
type Document struct {
	// XMLName captures the element name. Note: Document has no namespace itself,
	// but declares the idPkg namespace for child elements.
	XMLName xml.Name `xml:"Document"`

	// Xmlns defines the idPkg namespace prefix used by child elements.
	// Example: xmlns:idPkg="http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging"
	Xmlns string `xml:"xmlns:idPkg,attr"`

	// Core Identity Attributes
	DOMVersion string `xml:"DOMVersion,attr"`     // InDesign DOM version (e.g., "20.4")
	Self       string `xml:"Self,attr"`           // Unique identifier (usually "d")
	Name       string `xml:"Name,attr,omitempty"` // Document name

	// Story Management
	StoryList string `xml:"StoryList,attr,omitempty"` // Space-separated list of story IDs

	// Layout and Positioning
	ZeroPoint   string `xml:"ZeroPoint,attr,omitempty"`   // Document zero point (e.g., "0 0")
	ActiveLayer string `xml:"ActiveLayer,attr,omitempty"` // Currently active layer ID

	// Color Management
	CMYKProfile         string `xml:"CMYKProfile,attr,omitempty"`         // CMYK color profile
	RGBProfile          string `xml:"RGBProfile,attr,omitempty"`          // RGB color profile
	SolidColorIntent    string `xml:"SolidColorIntent,attr,omitempty"`    // Solid color rendering intent
	AfterBlendingIntent string `xml:"AfterBlendingIntent,attr,omitempty"` // Post-blending rendering intent
	DefaultImageIntent  string `xml:"DefaultImageIntent,attr,omitempty"`  // Default image rendering intent
	RGBPolicy           string `xml:"RGBPolicy,attr,omitempty"`           // RGB color policy
	CMYKPolicy          string `xml:"CMYKPolicy,attr,omitempty"`          // CMYK color policy
	AccurateLABSpots    string `xml:"AccurateLABSpots,attr,omitempty"`    // LAB spot color accuracy ("true"/"false")

	// MathML Settings (for mathematical typesetting)
	AppliedMathMLFontSize    string `xml:"AppliedMathMLFontSize,attr,omitempty"`    // MathML font size
	AppliedMathMLRgbColor    string `xml:"AppliedMathMLRgbColor,attr,omitempty"`    // MathML RGB color
	PreferMathMLInEpubExport string `xml:"PreferMathMLInEpubExport,attr,omitempty"` // MathML in EPUB export
	TintValue                string `xml:"TintValue,attr,omitempty"`                // Tint value for MathML

	// Child Elements (Phase 2 - Step 2: Properties)
	Properties *common.Properties `xml:"Properties,omitempty"`

	// Step 3: Languages
	Languages []Language `xml:"Language,omitempty"`

	// Step 4: Resource References (idPkg namespace)
	GraphicResource     *ResourceRef `xml:"http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging Graphic,omitempty"`
	FontsResource       *ResourceRef `xml:"http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging Fonts,omitempty"`
	StylesResource      *ResourceRef `xml:"http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging Styles,omitempty"`
	PreferencesResource *ResourceRef `xml:"http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging Preferences,omitempty"`
	TagsResource        *ResourceRef `xml:"http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging Tags,omitempty"`

	// Step 4b: Content Resource References (idPkg namespace)
	// These point to the actual document content (spreads, stories, etc.)
	MasterSpreads []ResourceRef `xml:"http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging MasterSpread,omitempty"`
	Spreads       []ResourceRef `xml:"http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging Spread,omitempty"`
	Stories       []ResourceRef `xml:"http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging Story,omitempty"`
	BackingStory  *ResourceRef  `xml:"http://ns.adobe.com/AdobeInDesign/idml/1.0/packaging BackingStory,omitempty"`

	// Step 5: Major Layout Elements
	Layers []Layer `xml:"Layer,omitempty"`

	// Step 6: Numbering and Grids
	NumberingLists []NumberingList `xml:"NumberingList,omitempty"`
	NamedGrids     []NamedGrid     `xml:"NamedGrid,omitempty"`

	// Step 7: Document Structure
	Sections      []Section      `xml:"Section,omitempty"`
	DocumentUsers []DocumentUser `xml:"DocumentUser,omitempty"`

	// Step 8: Colors, Bullets, and Assignments
	ColorGroups []ColorGroup `xml:"ColorGroup,omitempty"`
	ABullets    []ABullet    `xml:"ABullet,omitempty"`
	Assignments []Assignment `xml:"Assignment,omitempty"`

	// Step 9: Text Variables
	TextVariables []TextVariable `xml:"TextVariable,omitempty"`

	// ========================================================================
	// Phase 4.3: IDMS Inline Resources
	// ========================================================================
	// For IDMS (snippet) files, resources are embedded inline in the Document
	// instead of being in separate files. These fields allow the Document
	// struct to hold both IDML references (above) and IDMS inline content.
	//
	// In IDML files, these fields are typically empty (use ResourceRef instead).
	// In IDMS files, these fields contain the actual resource data.

	// Inline Colors and Swatches (instead of GraphicResource)
	// Note: Colors, Swatches, StrokeStyles use resources types (Phase 5a complete)
	Colors       []resources.Color       `xml:"Color,omitempty"`
	Swatches     []resources.Swatch      `xml:"Swatch,omitempty"`
	StrokeStyles []resources.StrokeStyle `xml:"StrokeStyle,omitempty"`

	// Inline Style Groups (instead of StylesResource)
	// Note: Style groups use resources types (Phase 5c complete)
	RootCharacterStyleGroup *resources.CharacterStyleGroup `xml:"RootCharacterStyleGroup,omitempty"`
	RootParagraphStyleGroup *resources.ParagraphStyleGroup `xml:"RootParagraphStyleGroup,omitempty"`
	RootObjectStyleGroup    *resources.ObjectStyleGroup    `xml:"RootObjectStyleGroup,omitempty"`

	// Inline Content (instead of Spreads/Stories ResourceRefs)
	// Note: InlineSpreads uses spread.SpreadElement (Phase 3 complete)
	// Note: InlineStories uses story.StoryElement (Phase 4 complete)
	InlineSpreads []spread.SpreadElement `xml:"Spread,omitempty"`
	InlineStories []story.StoryElement   `xml:"Story,omitempty"`

	// Required InDesign compatibility elements
	TinDocumentDataObject              *TinDocumentDataObject              `xml:"TinDocumentDataObject,omitempty"`
	TransparencyDefaultContainerObject *TransparencyDefaultContainerObject `xml:"TransparencyDefaultContainerObject,omitempty"`

	// Catch-all for all other child elements not yet explicitly modeled.
	// This includes: KinsokuTable, MojikumiTable, CrossReferenceFormat,
	// ConditionalTextPreference, EndnoteOption, WatermarkPreference, IndexingSortOption,
	// LinkedStoryOption, LinkedPageItemOption, and many more.
	// As we add explicit support for more elements, they move from OtherElements
	// to dedicated fields above.
	OtherElements []common.RawXMLElement `xml:",any"`
}

// Language represents a language definition in the document.
// Languages define localization settings for text.
type Language struct {
	XMLName xml.Name `xml:"Language"`

	// Identification
	Self string `xml:"Self,attr"`         // Unique identifier (e.g., "Language/$ID/English%3a UK")
	Name string `xml:"Name,attr"`         // Display name (e.g., "$ID/English: UK")
	Id   string `xml:"Id,attr,omitempty"` // Numeric language ID

	// Language Components
	PrimaryLanguageName string `xml:"PrimaryLanguageName,attr,omitempty"` // Primary language (e.g., "$ID/English")
	SublanguageName     string `xml:"SublanguageName,attr,omitempty"`     // Sublanguage/region (e.g., "$ID/UK")

	// Typographic Settings
	SingleQuotes string `xml:"SingleQuotes,attr,omitempty"` // Single quote characters (e.g., "''")
	DoubleQuotes string `xml:"DoubleQuotes,attr,omitempty"` // Double quote characters (e.g., """")

	// Processing Vendors
	HyphenationVendor string `xml:"HyphenationVendor,attr,omitempty"` // Hyphenation provider (e.g., "Proximity", "Hunspell")
	SpellingVendor    string `xml:"SpellingVendor,attr,omitempty"`    // Spell checker provider
}

// ResourceRef represents an idPkg:* resource reference element.
// These point to external XML files in the IDML package.
//
// Examples:
//
//	<idPkg:Graphic src="Resources/Graphic.xml" />
//	<idPkg:Fonts src="Resources/Fonts.xml" />
//	<idPkg:Styles src="Resources/Styles.xml" />
type ResourceRef struct {
	XMLName xml.Name // Will be set to the namespaced element name
	Src     string   `xml:"src,attr"`
}

// Layer represents a document layer for organizing content.
// Layers control visibility, locking, and printing of content.
type Layer struct {
	XMLName xml.Name `xml:"Layer"`

	// Identification
	Self string `xml:"Self,attr"` // Unique identifier (e.g., "uba")
	Name string `xml:"Name,attr"` // Display name (e.g., "Editorial")

	// Visibility and Interaction
	Visible    string `xml:"Visible,attr,omitempty"`    // Layer visibility ("true"/"false")
	Locked     string `xml:"Locked,attr,omitempty"`     // Lock layer ("true"/"false")
	IgnoreWrap string `xml:"IgnoreWrap,attr,omitempty"` // Ignore text wrap ("true"/"false")

	// Guides
	ShowGuides string `xml:"ShowGuides,attr,omitempty"` // Show guides ("true"/"false")
	LockGuides string `xml:"LockGuides,attr,omitempty"` // Lock guides ("true"/"false")

	// Layer Behavior
	UI         string `xml:"UI,attr,omitempty"`         // Show in UI ("true"/"false")
	Expendable string `xml:"Expendable,attr,omitempty"` // Can be deleted ("true"/"false")
	Printable  string `xml:"Printable,attr,omitempty"`  // Print layer ("true"/"false")

	// Properties may contain LayerColor and other settings
	Properties *common.Properties `xml:"Properties,omitempty"`

	// Catch-all for other Layer children
	OtherElements []common.RawXMLElement `xml:",any"`
}

// NumberingList represents a numbering list definition.
// NumberingLists control automatic numbering of paragraphs across stories and documents.
type NumberingList struct {
	XMLName xml.Name `xml:"NumberingList"`

	// Identification
	Self string `xml:"Self,attr"` // Unique identifier (e.g., "NumberingList/$ID/[Default]")
	Name string `xml:"Name,attr"` // Display name (e.g., "$ID/[Default]")

	// Numbering Behavior
	ContinueNumbersAcrossStories   string `xml:"ContinueNumbersAcrossStories,attr,omitempty"`   // Continue across stories ("true"/"false")
	ContinueNumbersAcrossDocuments string `xml:"ContinueNumbersAcrossDocuments,attr,omitempty"` // Continue across documents ("true"/"false")

	// Catch-all for future attributes or child elements
	OtherElements []common.RawXMLElement `xml:",any"`
}

// NamedGrid represents a named grid layout definition.
// NamedGrids are used primarily for CJK typography and grid-based layout systems.
type NamedGrid struct {
	XMLName xml.Name `xml:"NamedGrid"`

	// Identification
	Self string `xml:"Self,attr"` // Unique identifier (e.g., "NamedGrid/$ID/[Page Grid]")
	Name string `xml:"Name,attr"` // Display name (e.g., "$ID/[Page Grid]")

	// Grid Configuration
	GridDataInformation *common.GridDataInformation `xml:"GridDataInformation,omitempty"`

	// Catch-all for other NamedGrid children
	OtherElements []common.RawXMLElement `xml:",any"`
}

// Section represents a document section with its own page numbering.
// Sections allow different numbering schemes, prefixes, and styles within a document.
type Section struct {
	XMLName xml.Name `xml:"Section"`

	// Identification
	Self string `xml:"Self,attr"` // Unique identifier (e.g., "ub4")
	Name string `xml:"Name,attr"` // Section name (e.g., "A")

	// Page Range
	Length    string `xml:"Length,attr,omitempty"`    // Number of pages in section
	PageStart string `xml:"PageStart,attr,omitempty"` // Starting page ID

	// Numbering Configuration
	ContinueNumbering    string `xml:"ContinueNumbering,attr,omitempty"`    // Continue from previous section ("true"/"false")
	IncludeSectionPrefix string `xml:"IncludeSectionPrefix,attr,omitempty"` // Include prefix in page numbers ("true"/"false")
	PageNumberStart      string `xml:"PageNumberStart,attr,omitempty"`      // Starting page number
	SectionPrefix        string `xml:"SectionPrefix,attr,omitempty"`        // Prefix for page numbers (e.g., "A")
	Marker               string `xml:"Marker,attr,omitempty"`               // Section marker

	// Alternate Layout (for digital publishing)
	AlternateLayout       string `xml:"AlternateLayout,attr,omitempty"`       // Alternate layout name
	AlternateLayoutLength string `xml:"AlternateLayoutLength,attr,omitempty"` // Length in alternate layout

	// Properties may contain PageNumberStyle and Label
	Properties *common.Properties `xml:"Properties,omitempty"`

	// Catch-all for other Section children
	OtherElements []common.RawXMLElement `xml:",any"`
}

// DocumentUser represents a user who has worked on the document.
// This tracks collaboration and tracks changes ownership.
type DocumentUser struct {
	XMLName xml.Name `xml:"DocumentUser"`

	// Identification
	Self     string `xml:"Self,attr"`     // Unique identifier (e.g., "dDocumentUser0")
	UserName string `xml:"UserName,attr"` // User's name

	// Properties may contain UserColor and other settings
	Properties *common.Properties `xml:"Properties,omitempty"`

	// Catch-all for other DocumentUser children
	OtherElements []common.RawXMLElement `xml:",any"`
}

// ColorGroup represents a group of color swatches for organization.
// ColorGroups help organize and manage color swatches in the document.
type ColorGroup struct {
	XMLName xml.Name `xml:"ColorGroup"`

	// Identification
	Self string `xml:"Self,attr"` // Unique identifier (e.g., "ColorGroup/[Root Color Group]")
	Name string `xml:"Name,attr"` // Group name (e.g., "[Root Color Group]")

	// Root Group Indicator
	IsRootColorGroup string `xml:"IsRootColorGroup,attr,omitempty"` // Is this the root group ("true"/"false")

	// Color Swatches in this group
	ColorGroupSwatches []ColorGroupSwatch `xml:"ColorGroupSwatch,omitempty"`

	// Catch-all for other ColorGroup children
	OtherElements []common.RawXMLElement `xml:",any"`
}

// ColorGroupSwatch represents a reference to a color swatch within a color group.
type ColorGroupSwatch struct {
	XMLName xml.Name `xml:"ColorGroupSwatch"`

	// Identification
	Self          string `xml:"Self,attr"`          // Unique identifier
	SwatchItemRef string `xml:"SwatchItemRef,attr"` // Reference to the swatch (e.g., "Color/Black")

	// Catch-all for other ColorGroupSwatch children
	OtherElements []common.RawXMLElement `xml:",any"`
}

// ABullet represents a bullet character definition.
// ABullets define the character and font used for bullets in lists.
type ABullet struct {
	XMLName xml.Name `xml:"ABullet"`

	// Identification
	Self string `xml:"Self,attr"` // Unique identifier (e.g., "dABullet0")

	// Character Definition
	CharacterType  string `xml:"CharacterType,attr,omitempty"`  // Type of character ("UnicodeOnly", "UnicodeWithFont")
	CharacterValue string `xml:"CharacterValue,attr,omitempty"` // Unicode value (e.g., "8226" for bullet point)

	// Properties may contain BulletsFont and BulletsFontStyle
	Properties *common.Properties `xml:"Properties,omitempty"`

	// Catch-all for other ABullet children
	OtherElements []common.RawXMLElement `xml:",any"`
}

// Assignment represents an InCopy assignment for collaborative editing.
// Assignments define which content is assigned to specific users for editing.
type Assignment struct {
	XMLName xml.Name `xml:"Assignment"`

	// Identification
	Self     string `xml:"Self,attr"`     // Unique identifier (e.g., "uc9")
	Name     string `xml:"Name,attr"`     // Assignment name
	UserName string `xml:"UserName,attr"` // User assigned to this content

	// Export and Package Settings
	ExportOptions           string `xml:"ExportOptions,attr,omitempty"`           // What to export (e.g., "AssignedSpreads")
	IncludeLinksWhenPackage string `xml:"IncludeLinksWhenPackage,attr,omitempty"` // Include linked files ("true"/"false")
	FilePath                string `xml:"FilePath,attr,omitempty"`                // Assignment file path

	// Properties may contain FrameColor and other settings
	Properties *common.Properties `xml:"Properties,omitempty"`

	// Catch-all for other Assignment children
	OtherElements []common.RawXMLElement `xml:",any"`
}

// TextVariable represents a dynamic text variable in the document.
// TextVariables automatically update based on document properties or context
// (e.g., page numbers, dates, file names, running headers).
type TextVariable struct {
	XMLName xml.Name `xml:"TextVariable"`

	// Identification
	Self string `xml:"Self,attr"` // Unique identifier (e.g., "dTextVariablenChapter Number")
	Name string `xml:"Name,attr"` // Display name (e.g., "Chapter Number")

	// Variable Type
	VariableType string `xml:"VariableType,attr,omitempty"` // Type of variable (e.g., "ChapterNumberType", "CreationDateType")

	// Variable preferences (different types for different variable kinds)
	// Note: Only one of these will be present based on VariableType
	ChapterNumberPreference       *ChapterNumberVariablePreference   `xml:"ChapterNumberVariablePreference,omitempty"`
	DatePreference                *DateVariablePreference            `xml:"DateVariablePreference,omitempty"`
	FileNamePreference            *FileNameVariablePreference        `xml:"FileNameVariablePreference,omitempty"`
	CaptionMetadataPreference     *CaptionMetadataVariablePreference `xml:"CaptionMetadataVariablePreference,omitempty"`
	PageNumberPreference          *PageNumberVariablePreference      `xml:"PageNumberVariablePreference,omitempty"`
	MatchParagraphStylePreference *MatchParagraphStylePreference     `xml:"MatchParagraphStylePreference,omitempty"`

	// Catch-all for other TextVariable children or unknown preference types
	OtherElements []common.RawXMLElement `xml:",any"`
}

// ChapterNumberVariablePreference contains settings for chapter number variables.
type ChapterNumberVariablePreference struct {
	XMLName    xml.Name `xml:"ChapterNumberVariablePreference"`
	TextBefore string   `xml:"TextBefore,attr,omitempty"` // Text before the number
	Format     string   `xml:"Format,attr,omitempty"`     // Number format (e.g., "Current")
	TextAfter  string   `xml:"TextAfter,attr,omitempty"`  // Text after the number
}

// DateVariablePreference contains settings for date variables.
type DateVariablePreference struct {
	XMLName    xml.Name `xml:"DateVariablePreference"`
	TextBefore string   `xml:"TextBefore,attr,omitempty"` // Text before the date
	Format     string   `xml:"Format,attr,omitempty"`     // Date format (e.g., "dd/MM/yy", "d MMMM yyyy h:mm aa")
	TextAfter  string   `xml:"TextAfter,attr,omitempty"`  // Text after the date
}

// FileNameVariablePreference contains settings for file name variables.
type FileNameVariablePreference struct {
	XMLName          xml.Name `xml:"FileNameVariablePreference"`
	TextBefore       string   `xml:"TextBefore,attr,omitempty"`       // Text before the file name
	IncludePath      string   `xml:"IncludePath,attr,omitempty"`      // Include file path ("true"/"false")
	IncludeExtension string   `xml:"IncludeExtension,attr,omitempty"` // Include file extension ("true"/"false")
	TextAfter        string   `xml:"TextAfter,attr,omitempty"`        // Text after the file name
}

// CaptionMetadataVariablePreference contains settings for caption metadata variables.
type CaptionMetadataVariablePreference struct {
	XMLName              xml.Name `xml:"CaptionMetadataVariablePreference"`
	TextBefore           string   `xml:"TextBefore,attr,omitempty"`           // Text before the metadata
	MetadataProviderName string   `xml:"MetadataProviderName,attr,omitempty"` // Metadata source (e.g., "$ID/#LinkInfoNameStr")
	TextAfter            string   `xml:"TextAfter,attr,omitempty"`            // Text after the metadata
}

// PageNumberVariablePreference contains settings for page number variables.
type PageNumberVariablePreference struct {
	XMLName    xml.Name `xml:"PageNumberVariablePreference"`
	TextBefore string   `xml:"TextBefore,attr,omitempty"` // Text before the page number
	Format     string   `xml:"Format,attr,omitempty"`     // Number format (e.g., "Current")
	TextAfter  string   `xml:"TextAfter,attr,omitempty"`  // Text after the page number
	Scope      string   `xml:"Scope,attr,omitempty"`      // Scope (e.g., "SectionScope")
}

// MatchParagraphStylePreference contains settings for running header variables.
type MatchParagraphStylePreference struct {
	XMLName               xml.Name `xml:"MatchParagraphStylePreference"`
	TextBefore            string   `xml:"TextBefore,attr,omitempty"`            // Text before the matched text
	TextAfter             string   `xml:"TextAfter,attr,omitempty"`             // Text after the matched text
	AppliedParagraphStyle string   `xml:"AppliedParagraphStyle,attr,omitempty"` // Paragraph style to match
	SearchStrategy        string   `xml:"SearchStrategy,attr,omitempty"`        // Search strategy (e.g., "FirstOnPage")
	ChangeCase            string   `xml:"ChangeCase,attr,omitempty"`            // Case transformation (e.g., "None")
	DeleteEndPunctuation  string   `xml:"DeleteEndPunctuation,attr,omitempty"`  // Delete punctuation ("true"/"false")
}

// ============================================================================
// Phase 4.3: IDMS-specific Types
// ============================================================================

// TinDocumentDataObject represents InDesign's internal document data.
// This element is required for InDesign compatibility in IDMS files.
// It typically contains color space and rendering intent settings.
type TinDocumentDataObject struct {
	XMLName xml.Name `xml:"TinDocumentDataObject,omitempty"`
	// Contains internal InDesign data - preserved as-is for compatibility
	OtherElements []common.RawXMLElement `xml:",any"`
}

// TransparencyDefaultContainerObject contains default transparency settings.
// This element is required for InDesign compatibility in IDMS files.
// It defines default opacity and blend mode settings for the document.
type TransparencyDefaultContainerObject struct {
	XMLName xml.Name `xml:"TransparencyDefaultContainerObject,omitempty"`
	// Contains transparency defaults - preserved as-is for compatibility
	OtherElements []common.RawXMLElement `xml:",any"`
}
