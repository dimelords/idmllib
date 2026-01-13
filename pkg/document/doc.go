// Package document provides types and functions for working with IDML document metadata.
//
// The document package contains all types related to the designmap.xml file, which is
// the main manifest in an IDML package. This includes:
//   - Document: The root element containing document-level metadata and resource references
//   - DocumentWithMetadata: Document wrapper that preserves processing instructions
//   - Parsing functions: ParseDocument, ParseDocumentWithMetadata
//   - Marshaling functions: MarshalDocument, MarshalDocumentWithMetadata
//   - Legacy types: Designmap, DesignmapMinimal (deprecated, for backward compatibility)
//
// # Architecture
//
// This package is part of the Phase 2 refactoring (Epic 5) that splits the monolithic
// pkg/idml package into domain-specific packages:
//   - pkg/common: Shared types used across all packages
//   - pkg/document: Document metadata (this package)
//   - pkg/spread: Spread and page layout types (Phase 3)
//   - pkg/story: Text content and story types (Phase 4)
//   - pkg/resources: Styles, fonts, graphics (Phase 5)
//
// # Usage
//
// Parse a designmap.xml file:
//
//	data, _ := os.ReadFile("designmap.xml")
//	doc, err := document.ParseDocument(data)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(doc.DOMVersion, doc.Name)
//
// Parse with metadata preservation:
//
//	docMeta, err := document.ParseDocumentWithMetadata(data)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	// Access processing instructions
//	for _, pi := range docMeta.ProcessingInstructions {
//	    fmt.Printf("<?%s %s ?>\n", pi.Target, pi.Inst)
//	}
//
// Marshal back to XML:
//
//	xmlData, err := document.MarshalDocumentWithMetadata(docMeta)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	os.WriteFile("output.xml", xmlData, 0644)
//
// # Backward Compatibility
//
// For backward compatibility during the migration, all types are aliased in
// pkg/idml/types_aliases.go. External packages can continue using idml.Document
// which points to document.Document.
package document
