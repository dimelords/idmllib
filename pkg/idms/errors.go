package idms

import (
	"github.com/dimelords/idmllib/pkg/common"
)

// Legacy aliases for backward compatibility
// Deprecated: Use common.ErrNotFound instead
var ErrNotFound = common.ErrNotFound

// Deprecated: Use common.ErrInvalidFormat instead
var ErrInvalidFormat = common.ErrInvalidFormat

// Deprecated: Use common.ErrMissingMetadata instead
var ErrMissingMetadata = common.ErrMissingMetadata
