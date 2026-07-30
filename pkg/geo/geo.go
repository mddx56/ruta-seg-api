package geo

import "math"

const earthRadiusMeters = 6371000.0

// HaversineMeters calcula la distancia en metros entre dos coordenadas (lat/lon en grados).
func HaversineMeters(lat1, lon1, lat2, lon2 float64) float64 {
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusMeters * c
}

// Point es una coordenada (latitude, longitude) en grados.
type Point struct {
	Latitude  float64
	Longitude float64
}

// PointInPolygon determina si point está dentro del polígono usando ray casting.
// El polígono se recibe como lista ordenada de vértices (no hace falta cerrarlo explícitamente).
func PointInPolygon(point Point, polygon []Point) bool {
	if len(polygon) < 3 {
		return false
	}

	inside := false
	j := len(polygon) - 1
	for i := 0; i < len(polygon); i++ {
		pi, pj := polygon[i], polygon[j]

		intersects := (pi.Latitude > point.Latitude) != (pj.Latitude > point.Latitude) &&
			point.Longitude < (pj.Longitude-pi.Longitude)*(point.Latitude-pi.Latitude)/(pj.Latitude-pi.Latitude)+pi.Longitude

		if intersects {
			inside = !inside
		}
		j = i
	}

	return inside
}
