package spread

import (
	"encoding/xml"
	"math"
	"testing"
)

func TestNewPolygon(t *testing.T) {
	vertices := [][2]float64{
		{0, 0},
		{100, 0},
		{100, 100},
		{0, 100},
	}

	polygon := NewPolygon("u1e6", vertices, "uba")

	// Verify basic properties
	if polygon.Self != "u1e6" {
		t.Errorf("Self = %v, want u1e6", polygon.Self)
	}
	if polygon.ItemLayer != "uba" {
		t.Errorf("ItemLayer = %v, want uba", polygon.ItemLayer)
	}
	if polygon.Visible != "true" {
		t.Errorf("Visible = %v, want true", polygon.Visible)
	}

	// Verify vertices
	gotVertices, err := polygon.GetVertices()
	if err != nil {
		t.Fatalf("GetVertices() error = %v", err)
	}

	if len(gotVertices) != len(vertices) {
		t.Fatalf("vertex count = %d, want %d", len(gotVertices), len(vertices))
	}

	for i, want := range vertices {
		got := gotVertices[i]
		if math.Abs(got[0]-want[0]) > 0.001 || math.Abs(got[1]-want[1]) > 0.001 {
			t.Errorf("vertex %d = %v, want %v", i, got, want)
		}
	}
}

func TestNewRegularPolygon(t *testing.T) {
	tests := []struct {
		name     string
		centerX  float64
		centerY  float64
		radius   float64
		sides    int
		rotation float64
	}{
		{
			name:     "triangle",
			centerX:  100,
			centerY:  100,
			radius:   50,
			sides:    3,
			rotation: 0,
		},
		{
			name:     "square",
			centerX:  100,
			centerY:  100,
			radius:   50,
			sides:    4,
			rotation: 45, // Rotated 45 degrees
		},
		{
			name:     "hexagon",
			centerX:  100,
			centerY:  100,
			radius:   50,
			sides:    6,
			rotation: 0,
		},
		{
			name:     "octagon",
			centerX:  100,
			centerY:  100,
			radius:   50,
			sides:    8,
			rotation: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			polygon := NewRegularPolygon("test", tt.centerX, tt.centerY, tt.radius, tt.sides, tt.rotation, "uba")

			// Verify vertex count
			if count := polygon.VertexCount(); count != tt.sides {
				t.Errorf("VertexCount() = %d, want %d", count, tt.sides)
			}

			// Verify it's regular
			if !polygon.IsRegular() {
				t.Error("IsRegular() = false, want true for regular polygon")
			}

			// Verify centroid is approximately at center
			cx, cy, err := polygon.Centroid()
			if err != nil {
				t.Fatalf("Centroid() error = %v", err)
			}

			if math.Abs(cx-tt.centerX) > 0.1 || math.Abs(cy-tt.centerY) > 0.1 {
				t.Errorf("Centroid() = (%v, %v), want approximately (%v, %v)",
					cx, cy, tt.centerX, tt.centerY)
			}
		})
	}
}

func TestPolygon_GetSetVertices(t *testing.T) {
	polygon := &Polygon{
		PageItemBase: PageItemBase{
			Self: "test",
		},
	}

	// Set vertices
	vertices := [][2]float64{
		{0, 0},
		{50, 0},
		{50, 50},
		{0, 50},
	}
	polygon.SetVertices(vertices)

	// Get vertices back
	got, err := polygon.GetVertices()
	if err != nil {
		t.Fatalf("GetVertices() error = %v", err)
	}

	if len(got) != len(vertices) {
		t.Fatalf("vertex count = %d, want %d", len(got), len(vertices))
	}

	for i, want := range vertices {
		if math.Abs(got[i][0]-want[0]) > 0.001 || math.Abs(got[i][1]-want[1]) > 0.001 {
			t.Errorf("vertex %d = %v, want %v", i, got[i], want)
		}
	}
}

func TestPolygon_Perimeter(t *testing.T) {
	tests := []struct {
		name          string
		vertices      [][2]float64
		wantPerimeter float64
	}{
		{
			name: "unit square",
			vertices: [][2]float64{
				{0, 0},
				{1, 0},
				{1, 1},
				{0, 1},
			},
			wantPerimeter: 4.0,
		},
		{
			name: "rectangle 3x4",
			vertices: [][2]float64{
				{0, 0},
				{3, 0},
				{3, 4},
				{0, 4},
			},
			wantPerimeter: 14.0,
		},
		{
			name: "triangle 3-4-5",
			vertices: [][2]float64{
				{0, 0},
				{3, 0},
				{0, 4},
			},
			wantPerimeter: 12.0, // 3 + 4 + 5
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			polygon := NewPolygon("test", tt.vertices, "uba")
			perimeter := polygon.Perimeter()

			if math.Abs(perimeter-tt.wantPerimeter) > 0.001 {
				t.Errorf("Perimeter() = %v, want %v", perimeter, tt.wantPerimeter)
			}
		})
	}
}

func TestPolygon_Area(t *testing.T) {
	tests := []struct {
		name     string
		vertices [][2]float64
		wantArea float64
	}{
		{
			name: "unit square",
			vertices: [][2]float64{
				{0, 0},
				{1, 0},
				{1, 1},
				{0, 1},
			},
			wantArea: 1.0,
		},
		{
			name: "rectangle 3x4",
			vertices: [][2]float64{
				{0, 0},
				{3, 0},
				{3, 4},
				{0, 4},
			},
			wantArea: 12.0,
		},
		{
			name: "triangle base 4 height 3",
			vertices: [][2]float64{
				{0, 0},
				{4, 0},
				{0, 3},
			},
			wantArea: 6.0, // (4 * 3) / 2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			polygon := NewPolygon("test", tt.vertices, "uba")
			area := polygon.Area()

			if math.Abs(area-tt.wantArea) > 0.001 {
				t.Errorf("Area() = %v, want %v", area, tt.wantArea)
			}
		})
	}
}

func TestPolygon_Centroid(t *testing.T) {
	// Unit square centered at origin
	vertices := [][2]float64{
		{-0.5, -0.5},
		{0.5, -0.5},
		{0.5, 0.5},
		{-0.5, 0.5},
	}

	polygon := NewPolygon("test", vertices, "uba")
	cx, cy, err := polygon.Centroid()

	if err != nil {
		t.Fatalf("Centroid() error = %v", err)
	}

	if math.Abs(cx) > 0.001 || math.Abs(cy) > 0.001 {
		t.Errorf("Centroid() = (%v, %v), want (0, 0)", cx, cy)
	}
}

func TestPolygon_BoundingBox(t *testing.T) {
	vertices := [][2]float64{
		{10, 20},
		{50, 10},
		{60, 40},
		{30, 50},
	}

	polygon := NewPolygon("test", vertices, "uba")
	minX, minY, maxX, maxY, err := polygon.BoundingBox()

	if err != nil {
		t.Fatalf("BoundingBox() error = %v", err)
	}

	if math.Abs(minX-10) > 0.001 {
		t.Errorf("minX = %v, want 10", minX)
	}
	if math.Abs(minY-10) > 0.001 {
		t.Errorf("minY = %v, want 10", minY)
	}
	if math.Abs(maxX-60) > 0.001 {
		t.Errorf("maxX = %v, want 60", maxX)
	}
	if math.Abs(maxY-50) > 0.001 {
		t.Errorf("maxY = %v, want 50", maxY)
	}
}

func TestPolygon_IsRegular(t *testing.T) {
	tests := []struct {
		name string
		poly *Polygon
		want bool
	}{
		{
			name: "regular hexagon",
			poly: NewRegularPolygon("test", 0, 0, 50, 6, 0, "uba"),
			want: true,
		},
		{
			name: "regular pentagon",
			poly: NewRegularPolygon("test", 0, 0, 50, 5, 0, "uba"),
			want: true,
		},
		{
			name: "irregular quadrilateral",
			poly: NewPolygon("test", [][2]float64{
				{0, 0},
				{100, 0},
				{100, 50},
				{0, 50},
			}, "uba"),
			want: false, // Rectangle, not regular
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.poly.IsRegular()
			if got != tt.want {
				t.Errorf("IsRegular() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPolygon_SetStroke(t *testing.T) {
	polygon := NewPolygon("test", [][2]float64{{0, 0}, {10, 0}, {10, 10}}, "uba")
	polygon.SetStroke("Color/Black", 2.5, 100)

	if polygon.StrokeColor != "Color/Black" {
		t.Errorf("StrokeColor = %v, want Color/Black", polygon.StrokeColor)
	}
	if polygon.StrokeWeight != "2.5" {
		t.Errorf("StrokeWeight = %v, want 2.5", polygon.StrokeWeight)
	}
	if polygon.StrokeTint != "100" {
		t.Errorf("StrokeTint = %v, want 100", polygon.StrokeTint)
	}
}

func TestPolygon_SetFill(t *testing.T) {
	polygon := NewPolygon("test", [][2]float64{{0, 0}, {10, 0}, {10, 10}}, "uba")
	polygon.SetFill("Color/Red", 75)

	if polygon.FillColor != "Color/Red" {
		t.Errorf("FillColor = %v, want Color/Red", polygon.FillColor)
	}
	if polygon.FillTint != "75" {
		t.Errorf("FillTint = %v, want 75", polygon.FillTint)
	}
}

func TestPolygon_RoundTrip(t *testing.T) {
	vertices := [][2]float64{
		{0, 0},
		{50, 0},
		{50, 50},
		{25, 75},
		{0, 50},
	}

	original := NewPolygon("u1e6", vertices, "uba")
	original.AppliedObjectStyle = "ObjectStyle/$ID/[Normal Graphics Frame]"
	original.SetStroke("Color/Black", 1.5, 100)
	original.SetFill("Color/Blue", 50)

	// Marshal to XML
	xmlData, err := xml.MarshalIndent(original, "", "\t")
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	// Unmarshal back
	var restored Polygon
	err = xml.Unmarshal(xmlData, &restored)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	// Verify key fields
	if restored.Self != original.Self {
		t.Errorf("Self = %v, want %v", restored.Self, original.Self)
	}
	if restored.StrokeColor != original.StrokeColor {
		t.Errorf("StrokeColor = %v, want %v", restored.StrokeColor, original.StrokeColor)
	}
	if restored.FillColor != original.FillColor {
		t.Errorf("FillColor = %v, want %v", restored.FillColor, original.FillColor)
	}

	// Verify vertices
	oVertices, err := original.GetVertices()
	if err != nil {
		t.Fatalf("original GetVertices() error: %v", err)
	}

	rVertices, err := restored.GetVertices()
	if err != nil {
		t.Fatalf("restored GetVertices() error: %v", err)
	}

	if len(rVertices) != len(oVertices) {
		t.Fatalf("vertex count mismatch: got %d, want %d", len(rVertices), len(oVertices))
	}

	for i := range oVertices {
		if math.Abs(rVertices[i][0]-oVertices[i][0]) > 0.001 ||
			math.Abs(rVertices[i][1]-oVertices[i][1]) > 0.001 {
			t.Errorf("vertex %d mismatch: got %v, want %v", i, rVertices[i], oVertices[i])
		}
	}
}
