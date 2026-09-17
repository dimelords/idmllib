package idml

import (
	"github.com/dimelords/idmllib/v3/pkg/common"
	"github.com/dimelords/idmllib/v3/pkg/resources"
)

// Preferences returns the parsed Resources/Preferences.xml. The result is
// cached; edits made through the returned pointer must be followed by
// MarkModified(PathPreferences) or SetPreferences to be written.
func (p *Package) Preferences() (*resources.PreferencesFile, error) {
	if p.preferences != nil {
		return p.preferences, nil
	}
	entry, err := p.getFileEntry(PathPreferences)
	if err != nil {
		return nil, err
	}
	prefs, err := resources.ParsePreferencesFile(entry.data)
	if err != nil {
		return nil, common.WrapErrorWithPath("idml", "parse preferences", PathPreferences, err)
	}
	p.preferences = prefs
	return prefs, nil
}

// SetPreferences replaces the cached preferences and marks the file modified.
func (p *Package) SetPreferences(prefs *resources.PreferencesFile) {
	p.preferences = prefs
	p.markDirty(PathPreferences)
}

// FileData returns a copy of the raw bytes of a file in the package, for
// example "Resources/Styles.xml" or "META-INF/metadata.xml". It reflects the
// bytes as read, or as last written by a modification method; objects edited
// through a getter pointer are only reflected after Write.
func (p *Package) FileData(path string) ([]byte, error) {
	entry, err := p.getFileEntry(path)
	if err != nil {
		return nil, err
	}
	return append([]byte(nil), entry.data...), nil
}

// Tags returns the parsed XML/Tags.xml, which defines the XML tags that a
// story's XMLElement can refer to. The result is cached; edits made through
// the returned pointer must be followed by MarkModified(PathTags) or SetTags.
func (p *Package) Tags() (*resources.TagsFile, error) {
	if p.tags != nil {
		return p.tags, nil
	}
	entry, err := p.getFileEntry(PathTags)
	if err != nil {
		return nil, err
	}
	tags, err := resources.ParseTagsFile(entry.data)
	if err != nil {
		return nil, common.WrapErrorWithPath("idml", "parse tags", PathTags, err)
	}
	p.tags = tags
	return tags, nil
}

// SetTags replaces the cached tag definitions and marks the file modified.
func (p *Package) SetTags(tags *resources.TagsFile) {
	p.tags = tags
	p.markDirty(PathTags)
}
