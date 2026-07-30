package geo

import (
	"encoding/json"
	"math"
)

const metersPerDegreeLat = 111320.0

func metersPerDegreeLon(latDegrees float64) float64 {
	return 111320.0 * math.Cos(latDegrees*math.Pi/180)
}

type geoJSONFeatureCollection struct {
	Features []geoJSONFeature `json:"features"`
}

type geoJSONFeature struct {
	Geometry geoJSONGeometry `json:"geometry"`
}

type geoJSONGeometry struct {
	Type        string          `json:"type"`
	Coordinates json.RawMessage `json:"coordinates"`
}

// ParsePolylines extrae todas las polilíneas (LineString / MultiLineString) de un GeoJSON
// FeatureCollection, devolviendo cada una como una lista ordenada de puntos (lat/lon).
func ParsePolylines(geoJSON string) ([][]Point, error) {
	var fc geoJSONFeatureCollection
	if err := json.Unmarshal([]byte(geoJSON), &fc); err != nil {
		return nil, err
	}

	var polylines [][]Point
	for _, feature := range fc.Features {
		switch feature.Geometry.Type {
		case "LineString":
			var coords [][2]float64
			if err := json.Unmarshal(feature.Geometry.Coordinates, &coords); err != nil {
				continue
			}
			polylines = append(polylines, coordsToPoints(coords))
		case "MultiLineString":
			var lines [][][2]float64
			if err := json.Unmarshal(feature.Geometry.Coordinates, &lines); err != nil {
				continue
			}
			for _, coords := range lines {
				polylines = append(polylines, coordsToPoints(coords))
			}
		}
	}

	return polylines, nil
}

func coordsToPoints(coords [][2]float64) []Point {
	points := make([]Point, 0, len(coords))
	for _, c := range coords {
		// GeoJSON usa [longitude, latitude]
		points = append(points, Point{Latitude: c[1], Longitude: c[0]})
	}
	return points
}

// ProjectOntoPolyline proyecta point sobre la polilínea y devuelve:
//   - progressMeters: distancia acumulada desde el inicio de la polilínea hasta la proyección
//   - distanceToLineMeters: distancia aproximada del punto original a la polilínea
//
// Usa una proyección local plana (equirectangular) por segmento, válida para las distancias
// cortas entre puntos consecutivos de una ruta urbana.
func ProjectOntoPolyline(point Point, polyline []Point) (progressMeters float64, distanceToLineMeters float64) {
	if len(polyline) < 2 {
		return 0, math.Inf(1)
	}

	var cumulative float64
	bestDistance := math.Inf(1)
	var bestProgress float64

	for i := 0; i < len(polyline)-1; i++ {
		a, b := polyline[i], polyline[i+1]
		segmentLen := HaversineMeters(a.Latitude, a.Longitude, b.Latitude, b.Longitude)

		t, distToSegment := projectOntoSegment(point, a, b)
		progressAtProjection := cumulative + t*segmentLen

		if distToSegment < bestDistance {
			bestDistance = distToSegment
			bestProgress = progressAtProjection
		}

		cumulative += segmentLen
	}

	return bestProgress, bestDistance
}

// BestProjection prueba point contra varias polilíneas candidatas (p.ej. los distintos
// "sentido" de una misma línea) y devuelve el índice de la más cercana, junto con su progreso
// y distancia. polylineIndex es -1 si no hay polilíneas válidas.
func BestProjection(point Point, polylines [][]Point) (polylineIndex int, progressMeters float64, distanceToLineMeters float64) {
	polylineIndex = -1
	distanceToLineMeters = math.Inf(1)

	for i, polyline := range polylines {
		progress, dist := ProjectOntoPolyline(point, polyline)
		if dist < distanceToLineMeters {
			distanceToLineMeters = dist
			progressMeters = progress
			polylineIndex = i
		}
	}

	return
}

// PolylineLength suma la distancia Haversine entre puntos consecutivos de la polilínea.
func PolylineLength(polyline []Point) float64 {
	var total float64
	for i := 0; i < len(polyline)-1; i++ {
		total += HaversineMeters(polyline[i].Latitude, polyline[i].Longitude, polyline[i+1].Latitude, polyline[i+1].Longitude)
	}
	return total
}

func projectOntoSegment(p, a, b Point) (t float64, distanceMeters float64) {
	toXY := func(pt Point) (x, y float64) {
		x = (pt.Longitude - a.Longitude) * metersPerDegreeLon(a.Latitude)
		y = (pt.Latitude - a.Latitude) * metersPerDegreeLat
		return
	}

	bx, by := toXY(b)
	px, py := toXY(p)

	lengthSq := bx*bx + by*by
	if lengthSq == 0 {
		return 0, HaversineMeters(p.Latitude, p.Longitude, a.Latitude, a.Longitude)
	}

	t = (px*bx + py*by) / lengthSq
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}

	projX, projY := t*bx, t*by
	distX, distY := px-projX, py-projY
	distanceMeters = math.Sqrt(distX*distX + distY*distY)

	return t, distanceMeters
}
