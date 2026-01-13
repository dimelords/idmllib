package spread

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/dimelords/idmllib/v2/pkg/common"
)

// NewPolygon creates a new Polygon with required defaults.
//
// Parameters:
//   - self: Unique identifier for the polygon (e.g., "u1e6")
//   - vertices: Array of [x, y] coordinate pairs defining the polygon vertices
//   - layer: Layer ID (e.g., "uba")
//
// Returns a Polygon with sensible defaults.
func NewPolygon(self string, vertices [][2]float64, layer string) *Polygon {
	polygon := &Polygon{
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

	// Set the vertices using PathGeometry
	polygon.SetVertices(vertices)

	return polygon
}

// NewRegularPolygon creates a regular polygon (all sides equal) centered at a point.
//
// Parameters:
//   - self: Unique identifier
//   - centerX, centerY: Center point coordinates
//   - radius: Distance from center to each vertex
//   - sides: Number of sides (minimum 3)
//   - rotation: Rotation angle in degrees (0 = first vertex points right)
//   - layer: Layer ID
//
// Returns a Polygon with vertices arranged in a regular pattern.
func NewRegularPolygon(self string, centerX, centerY, radius float64, sides int, rotation float64, layer string) *Polygon {
	if sides < 3 {
		sides = 3
	}

	vertices := make([][2]float64, sides)
	angleStep := 360.0 / float64(sides)

	for i := 0; i < sides; i++ {
		// Calculate angle for this vertex
		angle := (float64(i)*angleStep + rotation) * math.Pi / 180

		// Calculate vertex position
		x := centerX + radius*math.Cos(angle)
		y := centerY + radius*math.Sin(angle)

		vertices[i] = [2]float64{x, y}
	}

	return NewPolygon(self, vertices, layer)
}

// GetVertices extracts the polygon vertices from PathGeometry.
//
// Returns:
//   - Array of [x, y] coordinate pairs
//   - error: If PathGeometry is missing or malformed
func (p *Polygon) GetVertices() ([][2]float64, error) {
	if p.Properties == nil || p.Properties.PathGeometry == nil {
		return nil, common.Errorf("spread", "get polygon vertices", "", "PathGeometry is nil")
	}

	pathGeom := p.Properties.PathGeometry
	if pathGeom.GeometryPathType == nil {
		return nil, common.Errorf("spread", "get polygon vertices", "", "GeometryPathType is nil")
	}

	geomPathType := pathGeom.GeometryPathType
	if geomPathType.PathPointArray == nil || len(geomPathType.PathPointArray.PathPoints) < 3 {
		return nil, common.Errorf("spread", "get polygon vertices", "", "insufficient path points (need at least 3 for a polygon)")
	}

	points := geomPathType.PathPointArray.PathPoints
	vertices := make([][2]float64, len(points))

	for i, point := range points {
		// Parse anchor (y x format)
		anchor := strings.Fields(point.Anchor)
		if len(anchor) != 2 {
			return nil, common.Errorf("spread", "get polygon vertices", "", "invalid anchor format for point %d: %s", i, point.Anchor)
		}

		y, err := strconv.ParseFloat(anchor[0], 64)
		if err != nil {
			return nil, common.WrapError("spread", "get polygon vertices", err)
		}

		x, err := strconv.ParseFloat(anchor[1], 64)
		if err != nil {
			return nil, common.WrapError("spread", "get polygon vertices", err)
		}

		vertices[i] = [2]float64{x, y}
	}

	return vertices, nil
}

// SetVertices updates the polygon vertices in PathGeometry.
//
// Parameters:
//   - vertices: Array of [x, y] coordinate pairs
func (p *Polygon) SetVertices(vertices [][2]float64) {
	if len(vertices) < 3 {
		return // Need at least 3 vertices for a polygon
	}

	// Initialize Properties if needed
	if p.Properties == nil {
		p.Properties = &common.Properties{}
	}

	// Initialize PathGeometry if needed
	if p.Properties.PathGeometry == nil {
		p.Properties.PathGeometry = &common.PathGeometry{
			GeometryPathType: &common.GeometryPathType{
				PathOpen: "false", // Polygons are closed paths
				PathPointArray: &common.PathPointArray{
					PathPoints: make([]common.PathPointType, len(vertices)),
				},
			},
		}
	}

	// Ensure we have enough points
	if len(p.Properties.PathGeometry.GeometryPathType.PathPointArray.PathPoints) != len(vertices) {
		p.Properties.PathGeometry.GeometryPathType.PathPointArray.PathPoints = make([]common.PathPointType, len(vertices))
	}

	// Set each vertex
	for i, vertex := range vertices {
		anchor := fmt.Sprintf("%g %g", vertex[1], vertex[0]) // y x format

		p.Properties.PathGeometry.GeometryPathType.PathPointArray.PathPoints[i] = common.PathPointType{
			Anchor:         anchor,
			LeftDirection:  anchor, // For straight edges
			RightDirection: anchor, // For straight edges
		}
	}
}

// VertexCount returns the number of vertices in the polygon.
func (p *Polygon) VertexCount() int {
	vertices, err := p.GetVertices()
	if err != nil {
		return 0
	}
	return len(vertices)
}

// Perimeter calculates the total length of all sides.
func (p *Polygon) Perimeter() float64 {
	vertices, err := p.GetVertices()
	if err != nil || len(vertices) < 2 {
		return 0
	}

	perimeter := 0.0

	// Sum distances between consecutive vertices
	for i := 0; i < len(vertices); i++ {
		next := (i + 1) % len(vertices) // Wrap around to first vertex
		dx := vertices[next][0] - vertices[i][0]
		dy := vertices[next][1] - vertices[i][1]
		perimeter += math.Sqrt(dx*dx + dy*dy)
	}

	return perimeter
}

// Area calculates the area using the shoelace formula.
func (p *Polygon) Area() float64 {
	vertices, err := p.GetVertices()
	if err != nil || len(vertices) < 3 {
		return 0
	}

	// Shoelace formula
	area := 0.0
	for i := 0; i < len(vertices); i++ {
		next := (i + 1) % len(vertices)
		area += vertices[i][0] * vertices[next][1]
		area -= vertices[next][0] * vertices[i][1]
	}

	return math.Abs(area) / 2.0
}

// Centroid calculates the geometric center of the polygon.
//
// Returns:
//   - centerX, centerY: Coordinates of the centroid
//   - error: If vertices cannot be retrieved
func (p *Polygon) Centroid() (centerX, centerY float64, err error) {
	vertices, err := p.GetVertices()
	if err != nil || len(vertices) < 3 {
		return 0, 0, common.Errorf("spread", "calculate centroid", "", "insufficient vertices")
	}

	// Calculate centroid
	for _, vertex := range vertices {
		centerX += vertex[0]
		centerY += vertex[1]
	}

	centerX /= float64(len(vertices))
	centerY /= float64(len(vertices))

	return centerX, centerY, nil
}

// BoundingBox returns the axis-aligned bounding box of the polygon.
//
// Returns:
//   - minX, minY, maxX, maxY: Bounding box coordinates
//   - error: If vertices cannot be retrieved
func (p *Polygon) BoundingBox() (minX, minY, maxX, maxY float64, err error) {
	vertices, err := p.GetVertices()
	if err != nil || len(vertices) == 0 {
		return 0, 0, 0, 0, common.Errorf("spread", "calculate bounding box", "", "no vertices")
	}

	// Initialize with first vertex
	minX, maxX = vertices[0][0], vertices[0][0]
	minY, maxY = vertices[0][1], vertices[0][1]

	// Find min/max
	for i := 1; i < len(vertices); i++ {
		if vertices[i][0] < minX {
			minX = vertices[i][0]
		}
		if vertices[i][0] > maxX {
			maxX = vertices[i][0]
		}
		if vertices[i][1] < minY {
			minY = vertices[i][1]
		}
		if vertices[i][1] > maxY {
			maxY = vertices[i][1]
		}
	}

	return minX, minY, maxX, maxY, nil
}

// IsRegular checks if the polygon is approximately regular (all sides equal).
// Uses a tolerance of 1% for side length variation.
func (p *Polygon) IsRegular() bool {
	vertices, err := p.GetVertices()
	if err != nil || len(vertices) < 3 {
		return false
	}

	// Calculate all side lengths
	sides := make([]float64, len(vertices))
	for i := 0; i < len(vertices); i++ {
		next := (i + 1) % len(vertices)
		dx := vertices[next][0] - vertices[i][0]
		dy := vertices[next][1] - vertices[i][1]
		sides[i] = math.Sqrt(dx*dx + dy*dy)
	}

	// Check if all sides are approximately equal
	avgLength := 0.0
	for _, length := range sides {
		avgLength += length
	}
	avgLength /= float64(len(sides))

	tolerance := avgLength * 0.01 // 1% tolerance

	for _, length := range sides {
		if math.Abs(length-avgLength) > tolerance {
			return false
		}
	}

	return true
}

// SetStroke sets the stroke properties for the polygon.
func (p *Polygon) SetStroke(color string, weight float64, tint float64) {
	p.StrokeColor = color
	p.StrokeWeight = fmt.Sprintf("%g", weight)
	p.StrokeTint = fmt.Sprintf("%g", tint)
}

// SetFill sets the fill properties for the polygon.
func (p *Polygon) SetFill(color string, tint float64) {
	p.FillColor = color
	p.FillTint = fmt.Sprintf("%g", tint)
}
