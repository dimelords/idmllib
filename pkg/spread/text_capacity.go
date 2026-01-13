package spread

import (
	"encoding/xml"
	"strconv"
	"strings"
)

// TextCapacityInfo contains text frame capacity information extracted from TextFramePreference.
// This provides the actual usable text dimensions accounting for columns, gutters, and insets.
type TextCapacityInfo struct {
	// ColumnWidth is the width of each text column in points
	ColumnWidth float64

	// ColumnCount is the number of text columns
	ColumnCount int

	// ColumnGutter is the space between columns in points
	ColumnGutter float64

	// InsetSpacing defines the internal padding [top, left, bottom, right] in points
	InsetSpacing [4]float64

	// EffectiveWidth is the total usable width for text across all columns
	// Formula: (ColumnWidth × ColumnCount) + (ColumnGutter × (ColumnCount - 1))
	EffectiveWidth float64

	// GeometricWidth is the frame's total width from GeometricBounds (for reference)
	GeometricWidth float64
}

// TextCapacity extracts text frame capacity information from TextFramePreference.
// Returns nil if TextFramePreference is not found or parsing fails.
//
// The returned TextCapacityInfo provides precise text dimensions that InDesign uses internally,
// which is critical for accurate text fitting calculations. This accounts for:
// - Multiple columns and gutters
// - Internal frame insets (padding)
// - Fixed column widths vs. flexible layouts
//
// Example:
//
//	capacity := frame.TextCapacity()
//	if capacity != nil {
//	    fmt.Printf("Usable text width: %.2fpt across %d column(s)\n",
//	        capacity.EffectiveWidth, capacity.ColumnCount)
//	}
func (f *SpreadTextFrame) TextCapacity() *TextCapacityInfo {
	info := &TextCapacityInfo{
		ColumnCount: 1, // Default to single column
	}

	// Get geometric bounds for reference
	if bounds, err := f.Bounds(); err == nil {
		info.GeometricWidth = bounds.Width
	}

	// Parse TextFramePreference from OtherElements
	var foundPreference bool
	for _, elem := range f.OtherElements {
		if elem.XMLName.Local == "TextFramePreference" {
			foundPreference = true

			// Extract attributes from TextFramePreference
			for _, attr := range elem.Attrs {
				switch attr.Name.Local {
				case "TextColumnFixedWidth":
					// This is the actual width InDesign uses for text layout
					if width, err := strconv.ParseFloat(attr.Value, 64); err == nil {
						info.ColumnWidth = width
					}

				case "TextColumnCount":
					if count, err := strconv.Atoi(attr.Value); err == nil && count > 0 {
						info.ColumnCount = count
					}

				case "TextColumnGutter":
					if gutter, err := strconv.ParseFloat(attr.Value, 64); err == nil {
						info.ColumnGutter = gutter
					}
				}
			}

			// Parse nested Properties for InsetSpacing from elem.Content
			info.InsetSpacing = parseInsetSpacingFromContent(elem.Content)

			break
		}
	}

	if !foundPreference {
		return nil
	}

	// Calculate effective width
	// If we have TextColumnFixedWidth, use it directly
	if info.ColumnWidth > 0 {
		// Effective width = all columns + gutters between them
		info.EffectiveWidth = (info.ColumnWidth * float64(info.ColumnCount)) +
			(info.ColumnGutter * float64(info.ColumnCount-1))
	} else if info.GeometricWidth > 0 {
		// Fallback: use geometric width minus insets
		usableWidth := info.GeometricWidth - info.InsetSpacing[1] - info.InsetSpacing[3] // left + right

		// Distribute across columns
		if info.ColumnCount > 1 {
			totalGutter := info.ColumnGutter * float64(info.ColumnCount-1)
			info.ColumnWidth = (usableWidth - totalGutter) / float64(info.ColumnCount)
		} else {
			info.ColumnWidth = usableWidth
		}

		info.EffectiveWidth = usableWidth
	}

	return info
}

// Helper structures for parsing InsetSpacing XML
type propertiesContainer struct {
	InsetSpacing *insetSpacingList `xml:"InsetSpacing"`
}

type insetSpacingList struct {
	Items []insetListItem `xml:"ListItem"`
}

type insetListItem struct {
	Value string `xml:",chardata"`
}

// parseInsetSpacingFromContent extracts InsetSpacing from raw XML content.
// InsetSpacing format: <Properties><InsetSpacing type="list"><ListItem>0</ListItem>...</InsetSpacing></Properties>
// Returns [top, left, bottom, right]
func parseInsetSpacingFromContent(content []byte) [4]float64 {
	var insets [4]float64

	if len(content) == 0 {
		return insets
	}

	// Parse the XML content
	var props propertiesContainer
	decoder := xml.NewDecoder(strings.NewReader(string(content)))
	if err := decoder.Decode(&props); err != nil {
		return insets
	}

	// Extract values from ListItems
	if props.InsetSpacing != nil {
		for i, item := range props.InsetSpacing.Items {
			if i >= 4 {
				break
			}
			if val, err := strconv.ParseFloat(strings.TrimSpace(item.Value), 64); err == nil {
				insets[i] = val
			}
		}
	}

	return insets
}
