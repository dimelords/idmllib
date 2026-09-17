package resources

import (
	"encoding/xml"

	"github.com/dimelords/idmllib/v2/internal/xmlutil"
	"github.com/dimelords/idmllib/v2/pkg/common"
)

// ParseGraphicFile parses a Graphic.xml file into a GraphicFile struct.
func ParseGraphicFile(data []byte) (*GraphicFile, error) {
	// Add nil check for input data
	if data == nil {
		return nil, common.Errorf("resources", "parse graphic", "", "input data is nil")
	}

	// Add empty data check
	if len(data) == 0 {
		return nil, common.Errorf("resources", "parse graphic", "", "input data is empty")
	}

	var graphic GraphicFile
	if err := xml.Unmarshal(data, &graphic); err != nil {
		return nil, common.WrapError("resources", "parse graphic", err)
	}
	return &graphic, nil
}

// MarshalGraphicFile marshals a GraphicFile struct back to XML with proper formatting.
func MarshalGraphicFile(graphic *GraphicFile) ([]byte, error) {
	return xmlutil.MarshalIndentWithHeader(graphic, "", "\t")
}
