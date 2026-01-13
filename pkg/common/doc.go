// Package common provides shared types used across all IDML domain packages.
//
// This package contains types that are referenced by multiple domain packages
// (document, spread, story, resources) to avoid circular dependencies and
// provide a stable foundation for the IDML type system.
//
// # Shared Types
//
// RawXMLElement: A forward-compatible catch-all for unknown XML elements.
// Used throughout the codebase for preserving elements that aren't yet modeled.
//
// Properties: A common container for metadata and configuration stored as
// key-value pairs in Label elements.
//
// GridDataInformation: Grid layout configuration shared between Document
// (NamedGrid) and Spread (Page) types.
//
// # Usage
//
// Domain packages import common/ to access these types:
//
//	import "github.com/dimelords/idmllib/pkg/common"
//
//	type MyType struct {
//	    Properties *common.Properties
//	    OtherElements []common.RawXMLElement
//	}
//
// # Architecture
//
// The common package is part of Epic 5's architecture refactoring that splits
// pkg/idml into domain-specific packages. See docs/EPIC-5-REFACTORING-ANALYSIS.md
// for detailed design decisions.
package common
