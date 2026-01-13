package idml

import (
	"fmt"

	"github.com/dimelords/idmllib/v2/pkg/common"
	"github.com/dimelords/idmllib/v2/pkg/resources"
)

// GetFontPostScriptName looks up the PostScript name for a font from Fonts.xml.
// This is useful for precise font matching in typography systems like HarfBuzz.
//
// Parameters:
//   - fontFamily: The font family name (e.g., "Helvetica", "Arial")
//   - fontStyle: The font style name (e.g., "Regular", "Bold", "Italic")
//
// Returns the PostScript name (e.g., "Helvetica-Bold") or empty string if not found.
func (p *Package) GetFontPostScriptName(fontFamily, fontStyle string) (string, error) {
	fonts, err := p.Fonts()
	if err != nil {
		return "", common.WrapError("idml", "get font postscript name", fmt.Errorf("failed to read Fonts.xml: %w", err))
	}

	// Search through font families
	for _, family := range fonts.FontFamilies {
		if family.Name == fontFamily {
			// Search through fonts in this family
			for _, font := range family.Fonts {
				// Match on FontStyleName
				if font.FontStyleName == fontStyle {
					return font.PostScriptName, nil
				}
			}
			// Family found but style not matched
			return "", common.WrapError("idml", "get font postscript name", fmt.Errorf("font style %q not found in family %q", fontStyle, fontFamily))
		}
	}

	// Family not found
	return "", common.WrapError("idml", "get font postscript name", fmt.Errorf("font family %q not found", fontFamily))
}

// GetFontByStyle retrieves a specific font from Fonts.xml by family and style.
// This provides access to all font metadata including Status, FontType, etc.
//
// Parameters:
//   - fontFamily: The font family name (e.g., "Helvetica", "Arial")
//   - fontStyle: The font style name (e.g., "Regular", "Bold", "Italic")
//
// Returns the Font struct or error if not found.
func (p *Package) GetFontByStyle(fontFamily, fontStyle string) (*resources.Font, error) {
	fonts, err := p.Fonts()
	if err != nil {
		return nil, common.WrapError("idml", "get font by style", fmt.Errorf("failed to read Fonts.xml: %w", err))
	}

	// Search through font families
	for _, family := range fonts.FontFamilies {
		if family.Name == fontFamily {
			// Search through fonts in this family
			for i := range family.Fonts {
				if family.Fonts[i].FontStyleName == fontStyle {
					return &family.Fonts[i], nil
				}
			}
			// Family found but style not matched
			return nil, common.WrapError("idml", "get font by style", fmt.Errorf("font style %q not found in family %q", fontStyle, fontFamily))
		}
	}

	// Family not found
	return nil, common.WrapError("idml", "get font by style", fmt.Errorf("font family %q not found", fontFamily))
}

// ListFontFamilies returns all font family names available in the document.
// Useful for font discovery and validation.
func (p *Package) ListFontFamilies() ([]string, error) {
	fonts, err := p.Fonts()
	if err != nil {
		return nil, common.WrapError("idml", "list font families", fmt.Errorf("failed to read Fonts.xml: %w", err))
	}

	families := make([]string, 0, len(fonts.FontFamilies))
	for _, family := range fonts.FontFamilies {
		families = append(families, family.Name)
	}

	return families, nil
}

// ListFontStyles returns all available font styles for a given font family.
// Useful for discovering available weights and styles.
func (p *Package) ListFontStyles(fontFamily string) ([]string, error) {
	fonts, err := p.Fonts()
	if err != nil {
		return nil, common.WrapError("idml", "list font styles", fmt.Errorf("failed to read Fonts.xml: %w", err))
	}

	for _, family := range fonts.FontFamilies {
		if family.Name == fontFamily {
			styles := make([]string, 0, len(family.Fonts))
			for _, font := range family.Fonts {
				styles = append(styles, font.FontStyleName)
			}
			return styles, nil
		}
	}

	return nil, common.WrapError("idml", "list font styles", fmt.Errorf("font family %q not found", fontFamily))
}
