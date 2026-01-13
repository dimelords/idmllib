package spread

import (
	"strconv"
	"strings"

	"github.com/dimelords/idmllib/pkg/common"
)

// Bounds represents the width and height of a frame in points.
type Bounds struct {
	Width  float64
	Height float64
}

// Position represents the x and y coordinates of a frame in points.
type Position struct {
	X float64
	Y float64
}

// Transform represents a full 6-value transformation matrix.
// Format: [a b c d x y]
// Where:
//   - a, d: scale factors
//   - b, c: rotation/skew factors
//   - x, y: translation (position)
type Transform struct {
	A float64 // Scale X
	B float64 // Rotation/Skew
	C float64 // Rotation/Skew
	D float64 // Scale Y
	X float64 // Translation X
	Y float64 // Translation Y
}

// Bounds returns the width and height of the text frame from its GeometricBounds.
// GeometricBounds format: "y1 x1 y2 x2" (top-left and bottom-right coordinates)
//
// If GeometricBounds is empty or invalid, falls back to calculating bounds from
// PathGeometry in Properties (commonly used in InDesign for frames with custom shapes).
//
// Returns an error if both GeometricBounds and PathGeometry are unavailable or malformed.
func (tf *SpreadTextFrame) Bounds() (Bounds, error) {
	// Try GeometricBounds first (most common case)
	bounds, err := parseGeometricBounds(tf.GeometricBounds)
	if err == nil && (bounds.Width > 0 && bounds.Height > 0) {
		return bounds, nil
	}

	// Fallback to PathGeometry if GeometricBounds is empty or invalid
	return tf.BoundsFromPathGeometry()
}

// Position returns the x and y position of the text frame from its ItemTransform.
// ItemTransform format: "a b c d x y" (6-value transformation matrix)
//
// Returns an error if ItemTransform is empty or malformed.
func (tf *SpreadTextFrame) Position() (Position, error) {
	return parseItemTransformPosition(tf.ItemTransform)
}

// Transform returns the full transformation matrix of the text frame.
// ItemTransform format: "a b c d x y" (6-value transformation matrix)
//
// Returns an error if ItemTransform is empty or malformed.
func (tf *SpreadTextFrame) Transform() (Transform, error) {
	return parseItemTransform(tf.ItemTransform)
}

// Bounds returns the width and height of the rectangle from its GeometricBounds.
func (r *Rectangle) Bounds() (Bounds, error) {
	return parseGeometricBounds(r.GeometricBounds)
}

// Position returns the x and y position of the rectangle from its ItemTransform.
func (r *Rectangle) Position() (Position, error) {
	return parseItemTransformPosition(r.ItemTransform)
}

// Transform returns the full transformation matrix of the rectangle.
func (r *Rectangle) Transform() (Transform, error) {
	return parseItemTransform(r.ItemTransform)
}

// Bounds returns the width and height of the oval from its GeometricBounds.
func (o *Oval) Bounds() (Bounds, error) {
	return parseGeometricBounds(o.GeometricBounds)
}

// Position returns the x and y position of the oval from its ItemTransform.
func (o *Oval) Position() (Position, error) {
	return parseItemTransformPosition(o.ItemTransform)
}

// Transform returns the full transformation matrix of the oval.
func (o *Oval) Transform() (Transform, error) {
	return parseItemTransform(o.ItemTransform)
}

// Bounds returns the width and height of the polygon from its GeometricBounds.
func (p *Polygon) Bounds() (Bounds, error) {
	return parseGeometricBounds(p.GeometricBounds)
}

// Position returns the x and y position of the polygon from its ItemTransform.
func (p *Polygon) Position() (Position, error) {
	return parseItemTransformPosition(p.ItemTransform)
}

// Transform returns the full transformation matrix of the polygon.
func (p *Polygon) Transform() (Transform, error) {
	return parseItemTransform(p.ItemTransform)
}

// parseGeometricBounds parses the GeometricBounds string format.
// Format: "y1 x1 y2 x2" (top-left and bottom-right coordinates in points)
func parseGeometricBounds(bounds string) (Bounds, error) {
	if bounds == "" {
		return Bounds{}, common.Errorf("spread", "parse geometric bounds", "", "GeometricBounds is empty")
	}

	parts := strings.Fields(bounds)
	if len(parts) != 4 {
		return Bounds{}, common.Errorf("spread", "parse geometric bounds", "", "invalid GeometricBounds format: expected 4 values, got %d", len(parts))
	}

	// Add recovery for potential panics during parsing
	defer func() {
		if r := recover(); r != nil {
			// This shouldn't happen with proper validation above, but provides safety
		}
	}()

	y1, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return Bounds{}, common.WrapError("spread", "parse geometric bounds", err)
	}

	x1, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return Bounds{}, common.WrapError("spread", "parse geometric bounds", err)
	}

	y2, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return Bounds{}, common.WrapError("spread", "parse geometric bounds", err)
	}

	x2, err := strconv.ParseFloat(parts[3], 64)
	if err != nil {
		return Bounds{}, common.WrapError("spread", "parse geometric bounds", err)
	}

	return Bounds{
		Width:  x2 - x1,
		Height: y2 - y1,
	}, nil
}

// parseItemTransformPosition parses only the position (x, y) from ItemTransform.
// Format: "a b c d x y" (6-value transformation matrix)
func parseItemTransformPosition(transform string) (Position, error) {
	if transform == "" {
		return Position{}, common.Errorf("spread", "parse item transform position", "", "ItemTransform is empty")
	}

	parts := strings.Fields(transform)
	if len(parts) != 6 {
		return Position{}, common.Errorf("spread", "parse item transform position", "", "invalid ItemTransform format: expected 6 values, got %d", len(parts))
	}

	x, err := strconv.ParseFloat(parts[4], 64)
	if err != nil {
		return Position{}, common.WrapError("spread", "parse item transform position", err)
	}

	y, err := strconv.ParseFloat(parts[5], 64)
	if err != nil {
		return Position{}, common.WrapError("spread", "parse item transform position", err)
	}

	return Position{X: x, Y: y}, nil
}

// parseItemTransform parses the full ItemTransform matrix.
// Format: "a b c d x y" (6-value transformation matrix)
func parseItemTransform(transform string) (Transform, error) {
	if transform == "" {
		return Transform{}, common.Errorf("spread", "parse item transform", "", "ItemTransform is empty")
	}

	parts := strings.Fields(transform)
	if len(parts) != 6 {
		return Transform{}, common.Errorf("spread", "parse item transform", "", "invalid ItemTransform format: expected 6 values, got %d", len(parts))
	}

	var t Transform
	var err error

	t.A, err = strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return Transform{}, common.WrapError("spread", "parse item transform", err)
	}

	t.B, err = strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return Transform{}, common.WrapError("spread", "parse item transform", err)
	}

	t.C, err = strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return Transform{}, common.WrapError("spread", "parse item transform", err)
	}

	t.D, err = strconv.ParseFloat(parts[3], 64)
	if err != nil {
		return Transform{}, common.WrapError("spread", "parse item transform", err)
	}

	t.X, err = strconv.ParseFloat(parts[4], 64)
	if err != nil {
		return Transform{}, common.WrapError("spread", "parse item transform", err)
	}

	t.Y, err = strconv.ParseFloat(parts[5], 64)
	if err != nil {
		return Transform{}, common.WrapError("spread", "parse item transform", err)
	}

	return t, nil
}

// BoundsFromPathGeometry calculates bounds from PathGeometry in Properties.
// This is used as a fallback when GeometricBounds attribute is not present.
// PathGeometry contains PathPointArray with anchor points that define the frame shape.
//
// Returns an error if Properties, PathGeometry, or PathPointArray is nil/empty.
func (tf *SpreadTextFrame) BoundsFromPathGeometry() (Bounds, error) {
	if tf.Properties == nil {
		return Bounds{}, common.Errorf("spread", "get bounds from path geometry", "", "Properties is nil")
	}
	if tf.Properties.PathGeometry == nil {
		return Bounds{}, common.Errorf("spread", "get bounds from path geometry", "", "PathGeometry is nil")
	}
	if tf.Properties.PathGeometry.GeometryPathType == nil {
		return Bounds{}, common.Errorf("spread", "get bounds from path geometry", "", "GeometryPathType is nil")
	}
	if tf.Properties.PathGeometry.GeometryPathType.PathPointArray == nil {
		return Bounds{}, common.Errorf("spread", "get bounds from path geometry", "", "PathPointArray is nil")
	}

	points := tf.Properties.PathGeometry.GeometryPathType.PathPointArray.PathPoints
	if len(points) == 0 {
		return Bounds{}, common.Errorf("spread", "get bounds from path geometry", "", "PathPointArray is empty")
	}

	// Parse all anchor points to find min/max X and Y
	minX, minY := parseAnchorPoint(points[0].Anchor)
	maxX, maxY := minX, minY

	for _, point := range points[1:] {
		x, y := parseAnchorPoint(point.Anchor)
		if x < minX {
			minX = x
		}
		if x > maxX {
			maxX = x
		}
		if y < minY {
			minY = y
		}
		if y > maxY {
			maxY = y
		}
	}

	return Bounds{
		Width:  maxX - minX,
		Height: maxY - minY,
	}, nil
}

// parseAnchorPoint parses an anchor point string "x y" and returns x, y as floats.
// Returns 0, 0 if parsing fails (handles errors silently for convenience).
func parseAnchorPoint(anchor string) (float64, float64) {
	parts := strings.Fields(anchor)
	if len(parts) != 2 {
		return 0, 0
	}

	x, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, 0
	}

	y, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return 0, 0
	}

	return x, y
}
