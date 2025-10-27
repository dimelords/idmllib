//revive:disable:var-naming
package types

// Label represents a label with key-value pairs in InDesign documents
type Label struct {
	KeyValuePairs []KeyValuePair `xml:"KeyValuePair"`
}

// KeyValuePair represents a key-value pair in a Label
type KeyValuePair struct {
	Key   string `xml:"Key,attr"`
	Value string `xml:"Value,attr"`
}

// InCopyExportOption represents export options for InCopy
type InCopyExportOption struct {
	IncludeGraphicProxies string `xml:"IncludeGraphicProxies,attr"`
	IncludeAllResources   string `xml:"IncludeAllResources,attr"`
}
