package spread_test

import (
	"testing"

	"github.com/dimelords/idmllib/pkg/spread"
)

func TestSpreadTextFrame_Bounds(t *testing.T) {
	tests := []struct {
		name            string
		geometricBounds string
		wantWidth       float64
		wantHeight      float64
		wantErr         bool
	}{
		{
			name:            "valid bounds",
			geometricBounds: "0 0 100 200",
			wantWidth:       200,
			wantHeight:      100,
		},
		{
			name:            "negative coordinates",
			geometricBounds: "-50 -30 50 70",
			wantWidth:       100,
			wantHeight:      100,
		},
		{
			name:            "decimal values",
			geometricBounds: "10.5 20.3 110.5 220.3",
			wantWidth:       200.0,
			wantHeight:      100.0,
		},
		{
			name:            "empty bounds",
			geometricBounds: "",
			wantErr:         true,
		},
		{
			name:            "invalid format - too few values",
			geometricBounds: "0 0 100",
			wantErr:         true,
		},
		{
			name:            "invalid format - non-numeric",
			geometricBounds: "0 0 abc 200",
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tf := &spread.SpreadTextFrame{
				PageItemBase: spread.PageItemBase{
					GeometricBounds: tt.geometricBounds,
				},
			}

			bounds, err := tf.Bounds()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if bounds.Width != tt.wantWidth {
				t.Errorf("Width = %v, want %v", bounds.Width, tt.wantWidth)
			}

			if bounds.Height != tt.wantHeight {
				t.Errorf("Height = %v, want %v", bounds.Height, tt.wantHeight)
			}
		})
	}
}

func TestSpreadTextFrame_Position(t *testing.T) {
	tests := []struct {
		name          string
		itemTransform string
		wantX         float64
		wantY         float64
		wantErr       bool
	}{
		{
			name:          "valid transform",
			itemTransform: "1 0 0 1 100 200",
			wantX:         100,
			wantY:         200,
		},
		{
			name:          "transform with rotation",
			itemTransform: "0.707 0.707 -0.707 0.707 150 250",
			wantX:         150,
			wantY:         250,
		},
		{
			name:          "decimal values",
			itemTransform: "1.5 0 0 1.5 123.456 789.012",
			wantX:         123.456,
			wantY:         789.012,
		},
		{
			name:          "empty transform",
			itemTransform: "",
			wantErr:       true,
		},
		{
			name:          "invalid format - too few values",
			itemTransform: "1 0 0 1 100",
			wantErr:       true,
		},
		{
			name:          "invalid format - non-numeric",
			itemTransform: "1 0 0 1 abc 200",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tf := &spread.SpreadTextFrame{
				PageItemBase: spread.PageItemBase{
					ItemTransform: tt.itemTransform,
				},
			}

			pos, err := tf.Position()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if pos.X != tt.wantX {
				t.Errorf("X = %v, want %v", pos.X, tt.wantX)
			}

			if pos.Y != tt.wantY {
				t.Errorf("Y = %v, want %v", pos.Y, tt.wantY)
			}
		})
	}
}

func TestSpreadTextFrame_Transform(t *testing.T) {
	tests := []struct {
		name          string
		itemTransform string
		want          spread.Transform
		wantErr       bool
	}{
		{
			name:          "identity transform",
			itemTransform: "1 0 0 1 100 200",
			want: spread.Transform{
				A: 1, B: 0, C: 0, D: 1, X: 100, Y: 200,
			},
		},
		{
			name:          "scaled transform",
			itemTransform: "2 0 0 2 50 75",
			want: spread.Transform{
				A: 2, B: 0, C: 0, D: 2, X: 50, Y: 75,
			},
		},
		{
			name:          "rotated transform",
			itemTransform: "0.707 0.707 -0.707 0.707 150 250",
			want: spread.Transform{
				A: 0.707, B: 0.707, C: -0.707, D: 0.707, X: 150, Y: 250,
			},
		},
		{
			name:          "empty transform",
			itemTransform: "",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tf := &spread.SpreadTextFrame{
				PageItemBase: spread.PageItemBase{
					ItemTransform: tt.itemTransform,
				},
			}

			transform, err := tf.Transform()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if transform != tt.want {
				t.Errorf("Transform = %+v, want %+v", transform, tt.want)
			}
		})
	}
}

func TestRectangle_Bounds(t *testing.T) {
	rect := &spread.Rectangle{
		PageItemBase: spread.PageItemBase{
			GeometricBounds: "0 0 100 200",
		},
	}

	bounds, err := rect.Bounds()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if bounds.Width != 200 {
		t.Errorf("Width = %v, want 200", bounds.Width)
	}

	if bounds.Height != 100 {
		t.Errorf("Height = %v, want 100", bounds.Height)
	}
}

func TestOval_Bounds(t *testing.T) {
	oval := &spread.Oval{
		PageItemBase: spread.PageItemBase{
			GeometricBounds: "10 20 110 220",
		},
	}

	bounds, err := oval.Bounds()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if bounds.Width != 200 {
		t.Errorf("Width = %v, want 200", bounds.Width)
	}

	if bounds.Height != 100 {
		t.Errorf("Height = %v, want 100", bounds.Height)
	}
}

func TestPolygon_Bounds(t *testing.T) {
	polygon := &spread.Polygon{
		PageItemBase: spread.PageItemBase{
			GeometricBounds: "5 10 105 210",
		},
	}

	bounds, err := polygon.Bounds()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if bounds.Width != 200 {
		t.Errorf("Width = %v, want 200", bounds.Width)
	}

	if bounds.Height != 100 {
		t.Errorf("Height = %v, want 100", bounds.Height)
	}
}
