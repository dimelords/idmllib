package spread

import (
	"encoding/xml"
	"math"
	"testing"

	"github.com/dimelords/idmllib/pkg/common"
)

func TestNewGraphicLine(t *testing.T) {
	tests := []struct {
		name  string
		self  string
		x1    float64
		y1    float64
		x2    float64
		y2    float64
		layer string
	}{
		{
			name:  "simple horizontal line",
			self:  "u1e6",
			x1:    100,
			y1:    50,
			x2:    200,
			y2:    50,
			layer: "uba",
		},
		{
			name:  "simple vertical line",
			self:  "u1e7",
			x1:    100,
			y1:    50,
			x2:    100,
			y2:    150,
			layer: "uba",
		},
		{
			name:  "diagonal line",
			self:  "u1e8",
			x1:    0,
			y1:    0,
			x2:    100,
			y2:    100,
			layer: "uba",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line := NewGraphicLine(tt.self, tt.x1, tt.y1, tt.x2, tt.y2, tt.layer)

			// Verify basic properties
			if line.Self != tt.self {
				t.Errorf("Self = %v, want %v", line.Self, tt.self)
			}
			if line.ItemLayer != tt.layer {
				t.Errorf("ItemLayer = %v, want %v", line.ItemLayer, tt.layer)
			}
			if line.Visible != "true" {
				t.Errorf("Visible = %v, want %v", line.Visible, "true")
			}

			// Verify endpoints were set
			x1, y1, x2, y2, err := line.GetEndpoints()
			if err != nil {
				t.Fatalf("GetEndpoints() error = %v", err)
			}

			if math.Abs(x1-tt.x1) > 0.001 {
				t.Errorf("x1 = %v, want %v", x1, tt.x1)
			}
			if math.Abs(y1-tt.y1) > 0.001 {
				t.Errorf("y1 = %v, want %v", y1, tt.y1)
			}
			if math.Abs(x2-tt.x2) > 0.001 {
				t.Errorf("x2 = %v, want %v", x2, tt.x2)
			}
			if math.Abs(y2-tt.y2) > 0.001 {
				t.Errorf("y2 = %v, want %v", y2, tt.y2)
			}
		})
	}
}

func TestGraphicLine_SetEndpoints(t *testing.T) {
	line := &GraphicLine{
		PageItemBase: PageItemBase{
			Self: "u1e6",
		},
	}

	// Set endpoints
	line.SetEndpoints(10, 20, 30, 40)

	// Verify PathGeometry was created
	if line.Properties == nil {
		t.Fatal("Properties is nil")
	}
	if line.Properties.PathGeometry == nil {
		t.Fatal("PathGeometry is nil")
	}
	if line.Properties.PathGeometry.GeometryPathType == nil {
		t.Fatal("GeometryPathType is nil")
	}
	if line.Properties.PathGeometry.GeometryPathType.PathPointArray == nil {
		t.Fatal("PathPointArray is nil")
	}

	points := line.Properties.PathGeometry.GeometryPathType.PathPointArray.PathPoints
	if len(points) != 2 {
		t.Fatalf("PathPoints count = %d, want 2", len(points))
	}

	// Verify endpoints
	x1, y1, x2, y2, err := line.GetEndpoints()
	if err != nil {
		t.Fatalf("GetEndpoints() error = %v", err)
	}

	if math.Abs(x1-10) > 0.001 || math.Abs(y1-20) > 0.001 ||
		math.Abs(x2-30) > 0.001 || math.Abs(y2-40) > 0.001 {
		t.Errorf("endpoints = (%v, %v) -> (%v, %v), want (10, 20) -> (30, 40)",
			x1, y1, x2, y2)
	}
}

func TestGraphicLine_GetEndpoints(t *testing.T) {
	tests := []struct {
		name    string
		line    *GraphicLine
		wantX1  float64
		wantY1  float64
		wantX2  float64
		wantY2  float64
		wantErr bool
	}{
		{
			name: "valid endpoints",
			line: &GraphicLine{
				Properties: &common.Properties{
					PathGeometry: &common.PathGeometry{
						GeometryPathType: &common.GeometryPathType{
							PathPointArray: &common.PathPointArray{
								PathPoints: []common.PathPointType{
									{Anchor: "20 10", LeftDirection: "20 10", RightDirection: "20 10"},
									{Anchor: "40 30", LeftDirection: "40 30", RightDirection: "40 30"},
								},
							},
						},
					},
				},
			},
			wantX1: 10,
			wantY1: 20,
			wantX2: 30,
			wantY2: 40,
		},
		{
			name: "nil properties",
			line: &GraphicLine{
				Properties: nil,
			},
			wantErr: true,
		},
		{
			name: "nil path geometry",
			line: &GraphicLine{
				Properties: &common.Properties{
					PathGeometry: nil,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x1, y1, x2, y2, err := tt.line.GetEndpoints()

			if (err != nil) != tt.wantErr {
				t.Errorf("GetEndpoints() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if math.Abs(x1-tt.wantX1) > 0.001 {
					t.Errorf("x1 = %v, want %v", x1, tt.wantX1)
				}
				if math.Abs(y1-tt.wantY1) > 0.001 {
					t.Errorf("y1 = %v, want %v", y1, tt.wantY1)
				}
				if math.Abs(x2-tt.wantX2) > 0.001 {
					t.Errorf("x2 = %v, want %v", x2, tt.wantX2)
				}
				if math.Abs(y2-tt.wantY2) > 0.001 {
					t.Errorf("y2 = %v, want %v", y2, tt.wantY2)
				}
			}
		})
	}
}

func TestGraphicLine_Length(t *testing.T) {
	tests := []struct {
		name       string
		x1, y1     float64
		x2, y2     float64
		wantLength float64
	}{
		{
			name:       "horizontal line 100 units",
			x1:         0,
			y1:         0,
			x2:         100,
			y2:         0,
			wantLength: 100,
		},
		{
			name:       "vertical line 50 units",
			x1:         0,
			y1:         0,
			x2:         0,
			y2:         50,
			wantLength: 50,
		},
		{
			name:       "3-4-5 triangle diagonal",
			x1:         0,
			y1:         0,
			x2:         3,
			y2:         4,
			wantLength: 5,
		},
		{
			name:       "diagonal line sqrt(2)",
			x1:         0,
			y1:         0,
			x2:         1,
			y2:         1,
			wantLength: math.Sqrt(2),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line := NewGraphicLine("test", tt.x1, tt.y1, tt.x2, tt.y2, "uba")
			length := line.Length()

			if math.Abs(length-tt.wantLength) > 0.001 {
				t.Errorf("Length() = %v, want %v", length, tt.wantLength)
			}
		})
	}
}

func TestGraphicLine_Angle(t *testing.T) {
	tests := []struct {
		name      string
		x1, y1    float64
		x2, y2    float64
		wantAngle float64
	}{
		{
			name:      "horizontal right (0°)",
			x1:        0,
			y1:        0,
			x2:        10,
			y2:        0,
			wantAngle: 0,
		},
		{
			name:      "vertical down (90°)",
			x1:        0,
			y1:        0,
			x2:        0,
			y2:        10,
			wantAngle: 90,
		},
		{
			name:      "horizontal left (180°)",
			x1:        10,
			y1:        0,
			x2:        0,
			y2:        0,
			wantAngle: 180,
		},
		{
			name:      "vertical up (270°)",
			x1:        0,
			y1:        10,
			x2:        0,
			y2:        0,
			wantAngle: 270,
		},
		{
			name:      "diagonal 45°",
			x1:        0,
			y1:        0,
			x2:        10,
			y2:        10,
			wantAngle: 45,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line := NewGraphicLine("test", tt.x1, tt.y1, tt.x2, tt.y2, "uba")
			angle := line.Angle()

			if math.Abs(angle-tt.wantAngle) > 0.1 {
				t.Errorf("Angle() = %v, want %v", angle, tt.wantAngle)
			}
		})
	}
}

func TestGraphicLine_IsHorizontal(t *testing.T) {
	tests := []struct {
		name string
		x1   float64
		y1   float64
		x2   float64
		y2   float64
		want bool
	}{
		{"horizontal", 0, 10, 100, 10, true},
		{"vertical", 10, 0, 10, 100, false},
		{"diagonal", 0, 0, 100, 100, false},
		{"nearly horizontal", 0, 10, 100, 10.005, true}, // Within tolerance
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line := NewGraphicLine("test", tt.x1, tt.y1, tt.x2, tt.y2, "uba")
			got := line.IsHorizontal()
			if got != tt.want {
				t.Errorf("IsHorizontal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGraphicLine_IsVertical(t *testing.T) {
	tests := []struct {
		name string
		x1   float64
		y1   float64
		x2   float64
		y2   float64
		want bool
	}{
		{"vertical", 10, 0, 10, 100, true},
		{"horizontal", 0, 10, 100, 10, false},
		{"diagonal", 0, 0, 100, 100, false},
		{"nearly vertical", 10, 0, 10.005, 100, true}, // Within tolerance
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line := NewGraphicLine("test", tt.x1, tt.y1, tt.x2, tt.y2, "uba")
			got := line.IsVertical()
			if got != tt.want {
				t.Errorf("IsVertical() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGraphicLine_RoundTrip tests that a GraphicLine can be marshaled and unmarshaled
func TestGraphicLine_RoundTrip(t *testing.T) {
	original := NewGraphicLine("u1e6", 100, 50, 200, 150, "uba")
	original.AppliedObjectStyle = "ObjectStyle/$ID/[Normal Graphics Frame]"
	original.ContentType = "Unassigned"
	original.LockState = "None"

	// Marshal to XML
	xmlData, err := xml.MarshalIndent(original, "", "\t")
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	// Unmarshal back
	var restored GraphicLine
	err = xml.Unmarshal(xmlData, &restored)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	// Verify key fields
	if restored.Self != original.Self {
		t.Errorf("Self = %v, want %v", restored.Self, original.Self)
	}
	if restored.ItemLayer != original.ItemLayer {
		t.Errorf("ItemLayer = %v, want %v", restored.ItemLayer, original.ItemLayer)
	}

	// Verify endpoints
	x1, y1, x2, y2, err := restored.GetEndpoints()
	if err != nil {
		t.Fatalf("GetEndpoints() error: %v", err)
	}

	ox1, oy1, ox2, oy2, err := original.GetEndpoints()
	if err != nil {
		t.Fatalf("original GetEndpoints() error: %v", err)
	}

	if math.Abs(x1-ox1) > 0.001 || math.Abs(y1-oy1) > 0.001 ||
		math.Abs(x2-ox2) > 0.001 || math.Abs(y2-oy2) > 0.001 {
		t.Errorf("endpoints mismatch: got (%v,%v)->(%v,%v), want (%v,%v)->(%v,%v)",
			x1, y1, x2, y2, ox1, oy1, ox2, oy2)
	}
}
