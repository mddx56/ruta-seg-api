package geo

import "testing"

func TestHaversineMetersZeroDistance(t *testing.T) {
	d := HaversineMeters(-17.790538, -63.171961, -17.790538, -63.171961)
	if d != 0 {
		t.Fatalf("expected 0 meters for identical points, got %f", d)
	}
}

func TestHaversineMetersKnownDistance(t *testing.T) {
	// ~111.32 km por grado de latitud en el ecuador (aprox.)
	d := HaversineMeters(0, 0, 1, 0)
	if d < 110500 || d > 111700 {
		t.Fatalf("expected ~111km for 1 degree latitude at equator, got %f meters", d)
	}
}

func TestPointInPolygonInside(t *testing.T) {
	square := []Point{
		{Latitude: 0, Longitude: 0},
		{Latitude: 0, Longitude: 10},
		{Latitude: 10, Longitude: 10},
		{Latitude: 10, Longitude: 0},
	}
	if !PointInPolygon(Point{Latitude: 5, Longitude: 5}, square) {
		t.Fatal("expected point (5,5) to be inside the square")
	}
}

func TestPointInPolygonOutside(t *testing.T) {
	square := []Point{
		{Latitude: 0, Longitude: 0},
		{Latitude: 0, Longitude: 10},
		{Latitude: 10, Longitude: 10},
		{Latitude: 10, Longitude: 0},
	}
	if PointInPolygon(Point{Latitude: 20, Longitude: 20}, square) {
		t.Fatal("expected point (20,20) to be outside the square")
	}
}

func TestPointInPolygonTooFewVertices(t *testing.T) {
	line := []Point{{Latitude: 0, Longitude: 0}, {Latitude: 1, Longitude: 1}}
	if PointInPolygon(Point{Latitude: 0, Longitude: 0}, line) {
		t.Fatal("a polygon with fewer than 3 vertices can never contain a point")
	}
}
