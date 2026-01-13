package spread

import (
	"encoding/xml"
	"math"
	"testing"
)

func TestNewOval(t *testing.T) {
	tests := []struct {
		name    string
		self    string
		centerX float64
		centerY float64
		width   float64
		height  float64
		layer   string
	}{
		{
			name:    "perfect circle",
			self:    "u1e6",
			centerX: 100,
			centerY: 100,
			width:   50,
			height:  50,
			layer:   "uba",
		},
		{
			name:    "horizontal ellipse",
			self:    "u1e7",
			centerX: 200,
			centerY: 150,
			width:   100,
			height:  50,
			layer:   "uba",
		},
		{
			name:    "vertical ellipse",
			self:    "u1e8",
			centerX: 150,
			centerY: 200,
			width:   50,
			height:  100,
			layer:   "uba",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oval := NewOval(tt.self, tt.centerX, tt.centerY, tt.width, tt.height, tt.layer)

			// Verify basic properties
			if oval.Self != tt.self {
				t.Errorf("Self = %v, want %v", oval.Self, tt.self)
			}
			if oval.ItemLayer != tt.layer {
				t.Errorf("ItemLayer = %v, want %v", oval.ItemLayer, tt.layer)
			}
			if oval.Visible != "true" {
				t.Errorf("Visible = %v, want %v", oval.Visible, "true")
			}

			// Verify bounds were set correctly
			centerX, centerY, width, height, err := oval.GetBounds()
			if err != nil {
				t.Fatalf("GetBounds() error = %v", err)
			}

			if math.Abs(centerX-tt.centerX) > 0.001 {
				t.Errorf("centerX = %v, want %v", centerX, tt.centerX)
			}
			if math.Abs(centerY-tt.centerY) > 0.001 {
				t.Errorf("centerY = %v, want %v", centerY, tt.centerY)
			}
			if math.Abs(width-tt.width) > 0.001 {
				t.Errorf("width = %v, want %v", width, tt.width)
			}
			if math.Abs(height-tt.height) > 0.001 {
				t.Errorf("height = %v, want %v", height, tt.height)
			}
		})
	}
}

func TestOval_GetBounds(t *testing.T) {
	tests := []struct {
		name        string
		oval        *Oval
		wantCenterX float64
		wantCenterY float64
		wantWidth   float64
		wantHeight  float64
		wantErr     bool
	}{
		{
			name: "valid bounds",
			oval: &Oval{
				PageItemBase: PageItemBase{
					GeometricBounds: "50 75 150 175",
				},
			},
			wantCenterX: 125,
			wantCenterY: 100,
			wantWidth:   100,
			wantHeight:  100,
		},
		{
			name: "empty bounds",
			oval: &Oval{
				PageItemBase: PageItemBase{
					GeometricBounds: "",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid format",
			oval: &Oval{
				PageItemBase: PageItemBase{
					GeometricBounds: "50 75 150", // missing one value
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			centerX, centerY, width, height, err := tt.oval.GetBounds()

			if (err != nil) != tt.wantErr {
				t.Errorf("GetBounds() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if math.Abs(centerX-tt.wantCenterX) > 0.001 {
					t.Errorf("centerX = %v, want %v", centerX, tt.wantCenterX)
				}
				if math.Abs(centerY-tt.wantCenterY) > 0.001 {
					t.Errorf("centerY = %v, want %v", centerY, tt.wantCenterY)
				}
				if math.Abs(width-tt.wantWidth) > 0.001 {
					t.Errorf("width = %v, want %v", width, tt.wantWidth)
				}
				if math.Abs(height-tt.wantHeight) > 0.001 {
					t.Errorf("height = %v, want %v", height, tt.wantHeight)
				}
			}
		})
	}
}

func TestOval_SetBounds(t *testing.T) {
	oval := &Oval{
		PageItemBase: PageItemBase{
			Self: "u1e6",
		},
	}

	// Set bounds
	oval.SetBounds(100, 200, 50, 80)

	// Verify GeometricBounds was set correctly
	if oval.GeometricBounds == "" {
		t.Fatal("GeometricBounds is empty")
	}

	// Verify by reading back
	centerX, centerY, width, height, err := oval.GetBounds()
	if err != nil {
		t.Fatalf("GetBounds() error = %v", err)
	}

	if math.Abs(centerX-100) > 0.001 || math.Abs(centerY-200) > 0.001 ||
		math.Abs(width-50) > 0.001 || math.Abs(height-80) > 0.001 {
		t.Errorf("bounds = (%v, %v, %v, %v), want (100, 200, 50, 80)",
			centerX, centerY, width, height)
	}
}

func TestOval_Area(t *testing.T) {
	tests := []struct {
		name     string
		centerX  float64
		centerY  float64
		width    float64
		height   float64
		wantArea float64
	}{
		{
			name:     "circle radius 10",
			centerX:  0,
			centerY:  0,
			width:    20,
			height:   20,
			wantArea: math.Pi * 10 * 10, // π * r²
		},
		{
			name:     "ellipse 100x50",
			centerX:  0,
			centerY:  0,
			width:    100,
			height:   50,
			wantArea: math.Pi * 50 * 25, // π * a * b
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oval := NewOval("test", tt.centerX, tt.centerY, tt.width, tt.height, "uba")
			area := oval.Area()

			if math.Abs(area-tt.wantArea) > 0.1 {
				t.Errorf("Area() = %v, want %v", area, tt.wantArea)
			}
		})
	}
}

func TestOval_Circumference(t *testing.T) {
	tests := []struct {
		name       string
		width      float64
		height     float64
		wantApprox float64 // Approximate expected value
	}{
		{
			name:       "circle diameter 20",
			width:      20,
			height:     20,
			wantApprox: 2 * math.Pi * 10, // 2πr ≈ 62.83
		},
		{
			name:       "ellipse 100x50",
			width:      100,
			height:     50,
			wantApprox: 240, // Approximate circumference for 100x50 ellipse
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oval := NewOval("test", 0, 0, tt.width, tt.height, "uba")
			circ := oval.Circumference()

			// Allow 5% tolerance for approximation
			tolerance := tt.wantApprox * 0.05
			if math.Abs(circ-tt.wantApprox) > tolerance {
				t.Errorf("Circumference() = %v, want approximately %v", circ, tt.wantApprox)
			}
		})
	}
}

func TestOval_IsCircle(t *testing.T) {
	tests := []struct {
		name   string
		width  float64
		height float64
		want   bool
	}{
		{"perfect circle", 50, 50, true},
		{"horizontal ellipse", 100, 50, false},
		{"vertical ellipse", 50, 100, false},
		{"nearly circle", 50, 50.005, true}, // Within tolerance
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oval := NewOval("test", 0, 0, tt.width, tt.height, "uba")
			got := oval.IsCircle()
			if got != tt.want {
				t.Errorf("IsCircle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOval_DimensionMethods(t *testing.T) {
	oval := NewOval("test", 100, 200, 80, 60, "uba")

	if w := oval.Width(); math.Abs(w-80) > 0.001 {
		t.Errorf("Width() = %v, want 80", w)
	}

	if h := oval.Height(); math.Abs(h-60) > 0.001 {
		t.Errorf("Height() = %v, want 60", h)
	}

	if cx := oval.CenterX(); math.Abs(cx-100) > 0.001 {
		t.Errorf("CenterX() = %v, want 100", cx)
	}

	if cy := oval.CenterY(); math.Abs(cy-200) > 0.001 {
		t.Errorf("CenterY() = %v, want 200", cy)
	}
}

func TestOval_SetStroke(t *testing.T) {
	oval := NewOval("test", 0, 0, 50, 50, "uba")

	oval.SetStroke("Color/Black", 2.5, 100)

	if oval.StrokeColor != "Color/Black" {
		t.Errorf("StrokeColor = %v, want Color/Black", oval.StrokeColor)
	}
	if oval.StrokeWeight != "2.5" {
		t.Errorf("StrokeWeight = %v, want 2.5", oval.StrokeWeight)
	}
	if oval.StrokeTint != "100" {
		t.Errorf("StrokeTint = %v, want 100", oval.StrokeTint)
	}
}

func TestOval_SetFill(t *testing.T) {
	oval := NewOval("test", 0, 0, 50, 50, "uba")

	oval.SetFill("Color/Red", 75)

	if oval.FillColor != "Color/Red" {
		t.Errorf("FillColor = %v, want Color/Red", oval.FillColor)
	}
	if oval.FillTint != "75" {
		t.Errorf("FillTint = %v, want 75", oval.FillTint)
	}
}

func TestOval_RoundTrip(t *testing.T) {
	original := NewOval("u1e6", 100, 150, 80, 60, "uba")
	original.AppliedObjectStyle = "ObjectStyle/$ID/[Normal Graphics Frame]"
	original.SetStroke("Color/Black", 1.5, 100)
	original.SetFill("Color/Red", 50)

	// Marshal to XML
	xmlData, err := xml.MarshalIndent(original, "", "\t")
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	// Unmarshal back
	var restored Oval
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
	if restored.StrokeColor != original.StrokeColor {
		t.Errorf("StrokeColor = %v, want %v", restored.StrokeColor, original.StrokeColor)
	}
	if restored.FillColor != original.FillColor {
		t.Errorf("FillColor = %v, want %v", restored.FillColor, original.FillColor)
	}

	// Verify bounds
	oCenterX, oCenterY, oWidth, oHeight, err := original.GetBounds()
	if err != nil {
		t.Fatalf("original GetBounds() error: %v", err)
	}

	rCenterX, rCenterY, rWidth, rHeight, err := restored.GetBounds()
	if err != nil {
		t.Fatalf("restored GetBounds() error: %v", err)
	}

	if math.Abs(rCenterX-oCenterX) > 0.001 || math.Abs(rCenterY-oCenterY) > 0.001 ||
		math.Abs(rWidth-oWidth) > 0.001 || math.Abs(rHeight-oHeight) > 0.001 {
		t.Errorf("bounds mismatch: got (%v,%v,%v,%v), want (%v,%v,%v,%v)",
			rCenterX, rCenterY, rWidth, rHeight, oCenterX, oCenterY, oWidth, oHeight)
	}
}
