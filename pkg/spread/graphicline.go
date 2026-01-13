package spread

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/dimelords/idmllib/v2/pkg/common"
)

// NewGraphicLine creates a new GraphicLine with required defaults.
//
// Parameters:
//   - self: Unique identifier for the line (e.g., "u1e6")
//   - x1, y1: Starting point coordinates in points
//   - x2, y2: Ending point coordinates in points
//   - layer: Layer ID (e.g., "uba")
//
// Returns a GraphicLine with sensible defaults for a simple straight line.
func NewGraphicLine(self string, x1, y1, x2, y2 float64, layer string) *GraphicLine {
	line := &GraphicLine{
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

	// Set the endpoints using PathGeometry
	line.SetEndpoints(x1, y1, x2, y2)

	return line
}

// GetEndpoints extracts the line endpoints from PathGeometry.
//
// Returns:
//   - x1, y1: Starting point coordinates
//   - x2, y2: Ending point coordinates
//   - error: If PathGeometry is missing or malformed
//
// For simple straight lines, PathGeometry contains two PathPointType entries.
func (gl *GraphicLine) GetEndpoints() (x1, y1, x2, y2 float64, err error) {
	if gl.Properties == nil || gl.Properties.PathGeometry == nil {
		return 0, 0, 0, 0, common.Errorf("spread", "get endpoints", "", "PathGeometry is nil")
	}

	pathGeom := gl.Properties.PathGeometry
	if pathGeom.GeometryPathType == nil {
		return 0, 0, 0, 0, common.Errorf("spread", "get endpoints", "", "GeometryPathType is nil")
	}

	geomPathType := pathGeom.GeometryPathType
	if geomPathType.PathPointArray == nil || len(geomPathType.PathPointArray.PathPoints) < 2 {
		return 0, 0, 0, 0, common.Errorf("spread", "get endpoints", "", "insufficient path points (need at least 2)")
	}

	points := geomPathType.PathPointArray.PathPoints

	// Parse first point (y1 x1 format)
	anchor1 := strings.Fields(points[0].Anchor)
	if len(anchor1) != 2 {
		return 0, 0, 0, 0, common.Errorf("spread", "get endpoints", "", "invalid anchor format for first point: %s", points[0].Anchor)
	}

	y1, err = strconv.ParseFloat(anchor1[0], 64)
	if err != nil {
		return 0, 0, 0, 0, common.WrapError("spread", "get endpoints", err)
	}

	x1, err = strconv.ParseFloat(anchor1[1], 64)
	if err != nil {
		return 0, 0, 0, 0, common.WrapError("spread", "get endpoints", err)
	}

	// Parse second point (y2 x2 format)
	anchor2 := strings.Fields(points[1].Anchor)
	if len(anchor2) != 2 {
		return 0, 0, 0, 0, common.Errorf("spread", "get endpoints", "", "invalid anchor format for second point: %s", points[1].Anchor)
	}

	y2, err = strconv.ParseFloat(anchor2[0], 64)
	if err != nil {
		return 0, 0, 0, 0, common.WrapError("spread", "get endpoints", err)
	}

	x2, err = strconv.ParseFloat(anchor2[1], 64)
	if err != nil {
		return 0, 0, 0, 0, common.WrapError("spread", "get endpoints", err)
	}

	return x1, y1, x2, y2, nil
}

// SetEndpoints updates the line endpoints in PathGeometry.
//
// Parameters:
//   - x1, y1: Starting point coordinates
//   - x2, y2: Ending point coordinates
//
// Creates PathGeometry if it doesn't exist.
func (gl *GraphicLine) SetEndpoints(x1, y1, x2, y2 float64) {
	// Format: "y x" for each point
	anchor1 := fmt.Sprintf("%g %g", y1, x1)
	anchor2 := fmt.Sprintf("%g %g", y2, x2)

	// Initialize Properties if needed
	if gl.Properties == nil {
		gl.Properties = &common.Properties{}
	}

	// Initialize PathGeometry if needed
	if gl.Properties.PathGeometry == nil {
		gl.Properties.PathGeometry = &common.PathGeometry{
			GeometryPathType: &common.GeometryPathType{
				PathOpen: "true", // Lines are open paths
				PathPointArray: &common.PathPointArray{
					PathPoints: make([]common.PathPointType, 2),
				},
			},
		}
	}

	// Ensure we have at least 2 points
	if len(gl.Properties.PathGeometry.GeometryPathType.PathPointArray.PathPoints) < 2 {
		gl.Properties.PathGeometry.GeometryPathType.PathPointArray.PathPoints = make([]common.PathPointType, 2)
	}

	// Set the anchor points (for straight lines, left/right directions match anchor)
	gl.Properties.PathGeometry.GeometryPathType.PathPointArray.PathPoints[0] = common.PathPointType{
		Anchor:         anchor1,
		LeftDirection:  anchor1,
		RightDirection: anchor1,
	}

	gl.Properties.PathGeometry.GeometryPathType.PathPointArray.PathPoints[1] = common.PathPointType{
		Anchor:         anchor2,
		LeftDirection:  anchor2,
		RightDirection: anchor2,
	}
}

// Length calculates the Euclidean distance between the line's endpoints.
//
// Returns:
//   - Line length in points
//   - 0 if PathGeometry is invalid
func (gl *GraphicLine) Length() float64 {
	x1, y1, x2, y2, err := gl.GetEndpoints()
	if err != nil {
		return 0
	}

	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx*dx + dy*dy)
}

// Angle returns the line's angle in degrees (0-360).
//
// The angle is measured from the horizontal axis (right is 0°),
// going counter-clockwise.
//
// Returns:
//   - Angle in degrees [0, 360)
//   - 0 if PathGeometry is invalid
func (gl *GraphicLine) Angle() float64 {
	x1, y1, x2, y2, err := gl.GetEndpoints()
	if err != nil {
		return 0
	}

	// Calculate angle in radians
	radians := math.Atan2(y2-y1, x2-x1)

	// Convert to degrees
	degrees := radians * 180 / math.Pi

	// Normalize to [0, 360)
	if degrees < 0 {
		degrees += 360
	}

	return degrees
}

// IsHorizontal returns true if the line is perfectly horizontal (within tolerance).
func (gl *GraphicLine) IsHorizontal() bool {
	_, y1, _, y2, err := gl.GetEndpoints()
	if err != nil {
		return false
	}
	return math.Abs(y2-y1) < 0.01 // Tolerance of 0.01 points
}

// IsVertical returns true if the line is perfectly vertical (within tolerance).
func (gl *GraphicLine) IsVertical() bool {
	x1, _, x2, _, err := gl.GetEndpoints()
	if err != nil {
		return false
	}
	return math.Abs(x2-x1) < 0.01 // Tolerance of 0.01 points
}
