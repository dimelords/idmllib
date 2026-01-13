package spread

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/dimelords/idmllib/pkg/common"
)

// NewOval creates a new Oval with required defaults.
//
// Parameters:
//   - self: Unique identifier for the oval (e.g., "u1e6")
//   - centerX, centerY: Center point coordinates in points
//   - width, height: Oval dimensions in points
//   - layer: Layer ID (e.g., "uba")
//
// Returns an Oval with sensible defaults for a simple ellipse.
func NewOval(self string, centerX, centerY, width, height float64, layer string) *Oval {
	oval := &Oval{
		PageItemBase: PageItemBase{
			Self:      self,
			ItemLayer: layer,
			Visible:   "true",
		},
		ContentType:         "Unassigned",
		LockState:           "None",
		Locked:              "false",
		LocalDisplaySetting: "Default",
	}

	// Set the bounds
	oval.SetBounds(centerX, centerY, width, height)

	return oval
}

// GetBounds extracts the oval's bounding box from GeometricBounds.
//
// Returns:
//   - centerX, centerY: Center point of the oval
//   - width, height: Oval dimensions
//   - error: If GeometricBounds is missing or malformed
//
// GeometricBounds format: "y1 x1 y2 x2" (top-left to bottom-right)
func (o *Oval) GetBounds() (centerX, centerY, width, height float64, err error) {
	if o.GeometricBounds == "" {
		return 0, 0, 0, 0, common.Errorf("spread", "get oval bounds", "", "GeometricBounds is empty")
	}

	parts := strings.Fields(o.GeometricBounds)
	if len(parts) != 4 {
		return 0, 0, 0, 0, common.Errorf("spread", "get oval bounds", "", "invalid GeometricBounds format: %s (expected 4 values)", o.GeometricBounds)
	}

	// Parse bounds (y1 x1 y2 x2)
	y1, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, 0, 0, 0, common.WrapError("spread", "get oval bounds", err)
	}
	x1, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return 0, 0, 0, 0, common.WrapError("spread", "get oval bounds", err)
	}
	y2, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return 0, 0, 0, 0, common.WrapError("spread", "get oval bounds", err)
	}
	x2, err := strconv.ParseFloat(parts[3], 64)
	if err != nil {
		return 0, 0, 0, 0, common.WrapError("spread", "get oval bounds", err)
	}

	// Calculate dimensions
	width = x2 - x1
	height = y2 - y1
	centerX = x1 + width/2
	centerY = y1 + height/2

	return centerX, centerY, width, height, nil
}

// SetBounds updates the oval's bounding box in GeometricBounds.
//
// Parameters:
//   - centerX, centerY: Center point of the oval
//   - width, height: Oval dimensions
func (o *Oval) SetBounds(centerX, centerY, width, height float64) {
	// Calculate bounds (top-left to bottom-right)
	x1 := centerX - width/2
	y1 := centerY - height/2
	x2 := centerX + width/2
	y2 := centerY + height/2

	// Format: "y1 x1 y2 x2"
	o.GeometricBounds = fmt.Sprintf("%g %g %g %g", y1, x1, y2, x2)
}

// Area calculates the area of the oval.
//
// Returns:
//   - Area in square points
//   - 0 if GeometricBounds is invalid
func (o *Oval) Area() float64 {
	_, _, width, height, err := o.GetBounds()
	if err != nil {
		return 0
	}

	// Area of ellipse: π * a * b (where a and b are semi-major and semi-minor axes)
	a := width / 2
	b := height / 2
	return math.Pi * a * b
}

// Circumference calculates the approximate circumference of the oval.
//
// Returns:
//   - Approximate circumference in points
//   - 0 if GeometricBounds is invalid
//
// Uses Ramanujan's approximation for ellipse circumference.
func (o *Oval) Circumference() float64 {
	_, _, width, height, err := o.GetBounds()
	if err != nil {
		return 0
	}

	a := width / 2  // semi-major axis
	b := height / 2 // semi-minor axis

	// Ramanujan's first approximation: π * (3(a+b) - √((3a+b)(a+3b)))
	sum := a + b
	term := (3*a + b) * (a + 3*b)
	return math.Pi * (3*sum - math.Sqrt(term))
}

// IsCircle returns true if the oval is a perfect circle (width equals height within tolerance).
func (o *Oval) IsCircle() bool {
	_, _, width, height, err := o.GetBounds()
	if err != nil {
		return false
	}
	return math.Abs(width-height) < 0.01 // Tolerance of 0.01 points
}

// Width returns the oval's width.
func (o *Oval) Width() float64 {
	_, _, width, _, err := o.GetBounds()
	if err != nil {
		return 0
	}
	return width
}

// Height returns the oval's height.
func (o *Oval) Height() float64 {
	_, _, _, height, err := o.GetBounds()
	if err != nil {
		return 0
	}
	return height
}

// CenterX returns the oval's center X coordinate.
func (o *Oval) CenterX() float64 {
	centerX, _, _, _, err := o.GetBounds()
	if err != nil {
		return 0
	}
	return centerX
}

// CenterY returns the oval's center Y coordinate.
func (o *Oval) CenterY() float64 {
	_, centerY, _, _, err := o.GetBounds()
	if err != nil {
		return 0
	}
	return centerY
}

// SetStroke sets the stroke properties for the oval.
//
// Parameters:
//   - color: Color reference (e.g., "Color/Black")
//   - weight: Stroke weight in points
//   - tint: Tint percentage (0-100), use 100 for solid color
func (o *Oval) SetStroke(color string, weight float64, tint float64) {
	o.StrokeColor = color
	o.StrokeWeight = fmt.Sprintf("%g", weight)
	o.StrokeTint = fmt.Sprintf("%g", tint)
}

// SetFill sets the fill properties for the oval.
//
// Parameters:
//   - color: Color reference (e.g., "Color/Red")
//   - tint: Tint percentage (0-100), use 100 for solid color
func (o *Oval) SetFill(color string, tint float64) {
	o.FillColor = color
	o.FillTint = fmt.Sprintf("%g", tint)
}
