package utils

import "time"

// BoliviaLocation es la zona horaria predeterminada para el negocio (Bolivia, UTC-4).
// Es imperativo utilizar esto en lugar de `time.Parse` para evitar desfasar franjas
// horarias hacia UTC y comprometer la precisión de los reportes e historiales GPS.
var BoliviaLocation = time.FixedZone("UTC-4", -4*60*60)

// ParseLocalDate parsea una cadena de texto "YYYY-MM-DD" anclándola forzosamente
// a las 00:00:00 del horario local (Bolivia). Evita las bombas de tiempo de UTC puro.
func ParseLocalDate(dateStr string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", dateStr, BoliviaLocation)
}

// ParseLocalDateTime parsea una cadena "YYYY-MM-DD HH:mm:ss" anclándola
// al horario local. Previene que el historial de ruta (start_time y end_time)
// se cargue con un desfase de 4 horas en la base de datos local.
func ParseLocalDateTime(dateTimeStr string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02 15:04:05", dateTimeStr, BoliviaLocation)
}

// NowLocal devuelve la hora actual forzada y convertida explícitamente a UTC-4.
// Útil para cuando no provean fecha (ej. endpoints predeterminados por el día actual)
func NowLocal() time.Time {
	return time.Now().In(BoliviaLocation)
}

// StartOfDayLocal recibe una fecha general y la lleva estrictamente al comienzo del día local (00:00:00 UTC-4).
func StartOfDayLocal(t time.Time) time.Time {
	localT := t.In(BoliviaLocation)
	return time.Date(localT.Year(), localT.Month(), localT.Day(), 0, 0, 0, 0, BoliviaLocation)
}
