package geo

import (
	"math"
	"testing"
)

func TestParsePolylinesLineString(t *testing.T) {
	geoJSON := `{"type":"FeatureCollection","features":[{"type":"Feature","geometry":{"type":"LineString","coordinates":[[-63.0,-17.0],[-63.01,-17.01]]},"properties":{}}]}`
	polylines, err := ParsePolylines(geoJSON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(polylines) != 1 || len(polylines[0]) != 2 {
		t.Fatalf("expected 1 polyline with 2 points, got %+v", polylines)
	}
	if polylines[0][0].Latitude != -17.0 || polylines[0][0].Longitude != -63.0 {
		t.Fatalf("expected lat/lon to be swapped from GeoJSON [lon,lat] order, got %+v", polylines[0][0])
	}
}

func TestParsePolylinesMultiLineString(t *testing.T) {
	geoJSON := `{"type":"FeatureCollection","features":[{"type":"Feature","geometry":{"type":"MultiLineString","coordinates":[[[-63.0,-17.0],[-63.01,-17.01],[-63.02,-17.02]]]},"properties":{}}]}`
	polylines, err := ParsePolylines(geoJSON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(polylines) != 1 || len(polylines[0]) != 3 {
		t.Fatalf("expected 1 polyline with 3 points, got %+v", polylines)
	}
}

func TestProjectOntoPolylineAtStart(t *testing.T) {
	polyline := []Point{{Latitude: 0, Longitude: 0}, {Latitude: 0, Longitude: 1}}
	progress, dist := ProjectOntoPolyline(Point{Latitude: 0, Longitude: 0}, polyline)
	if progress != 0 {
		t.Fatalf("expected 0 progress at the start of the polyline, got %f", progress)
	}
	if dist > 1 {
		t.Fatalf("expected ~0 distance for a point exactly on the polyline, got %f", dist)
	}
}

func TestProjectOntoPolylineMidpointAdvancesProgress(t *testing.T) {
	polyline := []Point{{Latitude: 0, Longitude: 0}, {Latitude: 0, Longitude: 1}, {Latitude: 0, Longitude: 2}}
	pStart, _ := ProjectOntoPolyline(Point{Latitude: 0, Longitude: 0.1}, polyline)
	pMid, _ := ProjectOntoPolyline(Point{Latitude: 0, Longitude: 1.5}, polyline)
	if !(pMid > pStart) {
		t.Fatalf("expected progress to increase along the polyline, got start=%f mid=%f", pStart, pMid)
	}
}

func TestProjectOntoPolylineTooFewPoints(t *testing.T) {
	_, dist := ProjectOntoPolyline(Point{Latitude: 0, Longitude: 0}, []Point{{Latitude: 0, Longitude: 0}})
	if !math.IsInf(dist, 1) {
		t.Fatalf("expected +Inf distance for a polyline with fewer than 2 points, got %f", dist)
	}
}

func TestBestProjectionPicksClosestPolyline(t *testing.T) {
	near := []Point{{Latitude: 0, Longitude: 0}, {Latitude: 0, Longitude: 1}}
	far := []Point{{Latitude: 10, Longitude: 0}, {Latitude: 10, Longitude: 1}}

	idx, _, dist := BestProjection(Point{Latitude: 0.001, Longitude: 0.5}, [][]Point{far, near})
	if idx != 1 {
		t.Fatalf("expected the closer polyline (index 1) to win, got index %d (dist=%f)", idx, dist)
	}
}

func TestBestProjectionNoPolylines(t *testing.T) {
	idx, _, _ := BestProjection(Point{Latitude: 0, Longitude: 0}, nil)
	if idx != -1 {
		t.Fatalf("expected -1 when there are no candidate polylines, got %d", idx)
	}
}

func TestPolylineLengthSumsSegments(t *testing.T) {
	polyline := []Point{{Latitude: 0, Longitude: 0}, {Latitude: 0, Longitude: 1}, {Latitude: 0, Longitude: 2}}
	oneSegment := HaversineMeters(0, 0, 0, 1)
	length := PolylineLength(polyline)
	if length < oneSegment*1.99 || length > oneSegment*2.01 {
		t.Fatalf("expected length ~= 2x one segment (%f), got %f", oneSegment*2, length)
	}
}

func TestPolylineLengthSinglePoint(t *testing.T) {
	if length := PolylineLength([]Point{{Latitude: 0, Longitude: 0}}); length != 0 {
		t.Fatalf("expected 0 length for a single point, got %f", length)
	}
}
