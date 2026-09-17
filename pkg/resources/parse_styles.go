package resources

import (
	"encoding/xml"

	"github.com/dimelords/idmllib/v2/internal/xmlutil"
	"github.com/dimelords/idmllib/v2/pkg/common"
)

// ParseStylesFile parses a Styles.xml file into a StylesFile struct.
func ParseStylesFile(data []byte) (*StylesFile, error) {
	// Add nil check for input data
	if data == nil {
		return nil, common.Errorf("resources", "parse styles", "", "input data is nil")
	}

	// Add empty data check
	if len(data) == 0 {
		return nil, common.Errorf("resources", "parse styles", "", "input data is empty")
	}

	var styles StylesFile
	if err := xml.Unmarshal(data, &styles); err != nil {
		return nil, common.WrapError("resources", "parse styles", err)
	}
	return &styles, nil
}

// MarshalStylesFile marshals a StylesFile struct back to XML with proper formatting.
func MarshalStylesFile(styles *StylesFile) ([]byte, error) {
	return xmlutil.MarshalIndentWithHeader(styles, "", "\t")
}
