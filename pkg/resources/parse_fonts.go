package resources

import (
	"encoding/xml"

	"github.com/dimelords/idmllib/v2/internal/xmlutil"
	"github.com/dimelords/idmllib/v2/pkg/common"
)

// ParseFontsFile parses a Fonts.xml file into a FontsFile struct.
func ParseFontsFile(data []byte) (*FontsFile, error) {
	// Add nil check for input data
	if data == nil {
		return nil, common.Errorf("resources", "parse fonts", "", "input data is nil")
	}

	// Add empty data check
	if len(data) == 0 {
		return nil, common.Errorf("resources", "parse fonts", "", "input data is empty")
	}

	var fonts FontsFile
	if err := xml.Unmarshal(data, &fonts); err != nil {
		return nil, common.WrapError("resources", "parse fonts", err)
	}
	return &fonts, nil
}

// MarshalFontsFile marshals a FontsFile struct back to XML with proper formatting.
func MarshalFontsFile(fonts *FontsFile) ([]byte, error) {
	return xmlutil.MarshalIndentWithHeader(fonts, "", "\t")
}
