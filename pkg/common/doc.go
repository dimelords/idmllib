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
//	import "github.com/dimelords/idmllib/v3/pkg/common"
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

// # XML fidelity helpers
//
// ChildOrder, EncodeChildren, ForEachChild, MarshalAttrs, UnmarshalAttrs,
// MarshalOrdered and UnmarshalOrdered exist so the domain packages can
// preserve attribute sets and child order when writing IDML back (see
// docs/FIDELITY.md). They are exported only because the domain packages live
// outside this one; application code should not need them.
