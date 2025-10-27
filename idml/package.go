package idml

import (
	"archive/zip"

	"github.com/dimelords/idmllib/types"
)

// Package represents an opened IDML file
type Package struct {
	path    string
	reader  *zip.ReadCloser
	Stories []types.Story
	Spreads []types.Spread
}
