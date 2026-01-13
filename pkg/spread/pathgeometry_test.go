package spread

import (
	"testing"

	"github.com/dimelords/idmllib/pkg/common"
)

func TestBoundsFromPathGeometry(t *testing.T) {
	tf := &SpreadTextFrame{
		Properties: &common.Properties{
			PathGeometry: &common.PathGeometry{
				GeometryPathType: &common.GeometryPathType{
					PathPointArray: &common.PathPointArray{
						PathPoints: []common.PathPointType{
							{Anchor: "43.189 74.573"},
							{Anchor: "43.189 142.973"},
							{Anchor: "1534.219 142.973"},
							{Anchor: "1534.219 74.573"},
						},
					},
				},
			},
		},
	}

	bounds, err := tf.BoundsFromPathGeometry()
	if err != nil {
		t.Fatalf("BoundsFromPathGeometry() error = %v", err)
	}

	expectedWidth := 1491.03 // 1534.219 - 43.189
	expectedHeight := 68.4   // 142.973 - 74.573

	// Use tolerance for floating point comparison
	const tolerance = 0.01
	if abs(bounds.Width-expectedWidth) > tolerance {
		t.Errorf("Width = %v, want %v (within %v)", bounds.Width, expectedWidth, tolerance)
	}

	if abs(bounds.Height-expectedHeight) > tolerance {
		t.Errorf("Height = %v, want %v (within %v)", bounds.Height, expectedHeight, tolerance)
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func TestBoundsFallbackToPathGeometry(t *testing.T) {
	// Frame with empty GeometricBounds but valid PathGeometry
	tf := &SpreadTextFrame{
		PageItemBase: PageItemBase{
			GeometricBounds: "", // Empty, should fallback
		},
		Properties: &common.Properties{
			PathGeometry: &common.PathGeometry{
				GeometryPathType: &common.GeometryPathType{
					PathPointArray: &common.PathPointArray{
						PathPoints: []common.PathPointType{
							{Anchor: "100 200"},
							{Anchor: "100 400"},
							{Anchor: "500 400"},
							{Anchor: "500 200"},
						},
					},
				},
			},
		},
	}

	bounds, err := tf.Bounds()
	if err != nil {
		t.Fatalf("Bounds() error = %v", err)
	}

	expectedWidth := 400.0  // 500 - 100
	expectedHeight := 200.0 // 400 - 200

	if bounds.Width != expectedWidth {
		t.Errorf("Width = %v, want %v", bounds.Width, expectedWidth)
	}

	if bounds.Height != expectedHeight {
		t.Errorf("Height = %v, want %v", bounds.Height, expectedHeight)
	}
}
