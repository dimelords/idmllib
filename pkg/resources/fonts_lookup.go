package resources

// Lookups over the font definitions in Resources/Fonts.xml. These answer
// questions about the file's own contents, so they live here rather than on
// the package coordinator.

// Family returns the font family with the given name.
func (f *FontsFile) Family(name string) (*FontFamily, bool) {
	for i := range f.FontFamilies {
		if f.FontFamilies[i].Name == name {
			return &f.FontFamilies[i], true
		}
	}
	return nil, false
}

// Families returns the name of every font family in the file, in document
// order.
func (f *FontsFile) Families() []string {
	names := make([]string, 0, len(f.FontFamilies))
	for i := range f.FontFamilies {
		names = append(names, f.FontFamilies[i].Name)
	}
	return names
}

// Font returns the font of the given family and style, for example
// ("Minion Pro", "Bold"). The style is matched against FontStyleName.
func (f *FontsFile) Font(family, style string) (*Font, bool) {
	fam, ok := f.Family(family)
	if !ok {
		return nil, false
	}
	for i := range fam.Fonts {
		if fam.Fonts[i].FontStyleName == style {
			return &fam.Fonts[i], true
		}
	}
	return nil, false
}

// Styles returns the available styles of a font family, in document order.
// It returns nil when the family is not present.
func (f *FontsFile) Styles(family string) []string {
	fam, ok := f.Family(family)
	if !ok {
		return nil
	}
	styles := make([]string, 0, len(fam.Fonts))
	for i := range fam.Fonts {
		styles = append(styles, fam.Fonts[i].FontStyleName)
	}
	return styles
}

// PostScriptName returns the PostScript name of the font of the given family
// and style, which is what page items and styles reference.
func (f *FontsFile) PostScriptName(family, style string) (string, bool) {
	font, ok := f.Font(family, style)
	if !ok {
		return "", false
	}
	return font.PostScriptName, true
}
