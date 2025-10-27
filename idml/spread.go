package idml

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/dimelords/idmllib/types"
)

func (p *Package) readSpread(data []byte) (types.Spread, error) {
	var spread types.IDPkgSpread

	if err := xml.Unmarshal(data, &spread); err != nil {
		return types.Spread{}, fmt.Errorf("failed to parse spread: %w", err)
	}

	return spread.Spread, nil
}

func isSpread(name string) bool {
	return strings.HasPrefix(name, "Spreads/Spread") && strings.HasSuffix(name, ".xml")
}
