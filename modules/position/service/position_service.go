package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"github.com/Caknoooo/go-gin-clean-starter/modules/position/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/position/repository"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/utils"
	"github.com/google/uuid"
)

type PositionService interface {
	Create(ctx context.Context, req dto.PositionCreateRequest) (dto.PositionResponse, error)
	GetByID(ctx context.Context, id uint64) (dto.PositionResponse, error)
	GetByIMEI(ctx context.Context, imei string) ([]dto.PositionResponse, error)
	GetLastByIMEI(ctx context.Context, imei string) (dto.PositionResponse, error)
	GetLastPositionsOfAllDevices(ctx context.Context) ([]dto.PositionResponse, error)
	GetCoordinatesByIMEIAndDate(ctx context.Context, imei string, date time.Time) ([]dto.PositionCoordinateResponse, error)
	Delete(ctx context.Context, id uint64) error

	// History (Legacy by IMEI)
	GetDeviceHistory(ctx context.Context, imei string, dateStr string) (*dto.HistoryResponse, error)
	GetDeviceRoute(ctx context.Context, imei string, start, end time.Time) (*dto.HistoryResponse, error)

	// Vehicle Centric
	GetVehicleHistory(ctx context.Context, vehicleID string, dateStr string) (*dto.VehicleHistoryResponse, error)
	GetVehicleRoute(ctx context.Context, vehicleID string, start, end time.Time) (*dto.VehicleHistoryResponse, error)
	GetLastPositionOfVehicle(ctx context.Context, vehicleID string) (dto.VehiclePositionResponse, error)
	GetLastPositionsOfAllVehicles(ctx context.Context) ([]dto.VehiclePositionResponse, error)
}

const (
	MinMovingSpeed  = 5 // km/h
	MinStopDuration = 3 * time.Minute
	MaxGapDuration  = 10 * time.Minute
)

// localTime convierte cualquier time.Time al huso horario de Bolivia (UTC-4).
// Úsalo siempre que construyas un DTO de respuesta para que el cliente
// reciba tiempos locales y no UTC-0.
func localTime(t time.Time) time.Time {
	return t.In(utils.BoliviaLocation)
}

type positionService struct {
	repo repository.PositionRepository
}

func NewPositionService(repo repository.PositionRepository) PositionService {
	return &positionService{repo: repo}
}

func (s *positionService) Create(ctx context.Context, req dto.PositionCreateRequest) (dto.PositionResponse, error) {
	position := entities.Position{
		Imei:       req.Imei,
		DeviceTime: req.DeviceTime,
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
		Speed:      req.Speed,
		Course:     req.Course,
		Attributes: req.Attributes,
	}

	if err := s.repo.Create(ctx, &position); err != nil {
		return dto.PositionResponse{}, err
	}

	return dto.PositionResponse{
		ID:         position.ID,
		Imei:       position.Imei,
		ServerTime: localTime(position.ServerTime),
		DeviceTime: localTime(position.DeviceTime),
		Latitude:   position.Latitude,
		Longitude:  position.Longitude,
		Speed:      position.Speed,
		Course:     position.Course,
		Attributes: position.Attributes,
	}, nil
}

func (s *positionService) GetByID(ctx context.Context, id uint64) (dto.PositionResponse, error) {
	position, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return dto.PositionResponse{}, err
	}

	return dto.PositionResponse{
		ID:         position.ID,
		Imei:       position.Imei,
		ServerTime: localTime(position.ServerTime),
		DeviceTime: localTime(position.DeviceTime),
		Latitude:   position.Latitude,
		Longitude:  position.Longitude,
		Speed:      position.Speed,
		Course:     position.Course,
		Attributes: position.Attributes,
	}, nil
}

func (s *positionService) GetByIMEI(ctx context.Context, imei string) ([]dto.PositionResponse, error) {
	positions, err := s.repo.FindByIMEI(ctx, imei)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.PositionResponse, 0)
	for _, pos := range positions {
		responses = append(responses, dto.PositionResponse{
			ID:         pos.ID,
			Imei:       pos.Imei,
			ServerTime: localTime(pos.ServerTime),
			DeviceTime: localTime(pos.DeviceTime),
			Latitude:   pos.Latitude,
			Longitude:  pos.Longitude,
			Speed:      pos.Speed,
			Course:     pos.Course,
			Attributes: pos.Attributes,
		})
	}

	return responses, nil
}

func (s *positionService) GetLastByIMEI(ctx context.Context, imei string) (dto.PositionResponse, error) {
	position, err := s.repo.FindLastByIMEI(ctx, imei)
	if err != nil {
		return dto.PositionResponse{}, err
	}

	return dto.PositionResponse{
		ID:         position.ID,
		Imei:       position.Imei,
		ServerTime: localTime(position.ServerTime),
		DeviceTime: localTime(position.DeviceTime),
		Latitude:   position.Latitude,
		Longitude:  position.Longitude,
		Speed:      position.Speed,
		Course:     position.Course,
		Attributes: position.Attributes,
	}, nil
}

func (s *positionService) GetLastPositionsOfAllDevices(ctx context.Context) ([]dto.PositionResponse, error) {
	positions, err := s.repo.FindLastPositions(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.PositionResponse, 0)
	for _, pos := range positions {
		responses = append(responses, dto.PositionResponse{
			ID:         pos.ID,
			Imei:       pos.Imei,
			ServerTime: localTime(pos.ServerTime),
			DeviceTime: localTime(pos.DeviceTime),
			Latitude:   pos.Latitude,
			Longitude:  pos.Longitude,
			Speed:      pos.Speed,
			Course:     pos.Course,
			Attributes: pos.Attributes,
		})
	}

	return responses, nil
}

func (s *positionService) GetCoordinatesByIMEIAndDate(ctx context.Context, imei string, date time.Time) ([]dto.PositionCoordinateResponse, error) {
	positions, err := s.repo.FindByIMEIAndDate(ctx, imei, date)
	if err != nil {
		return []dto.PositionCoordinateResponse{}, err
	}

	responses := make([]dto.PositionCoordinateResponse, 0)

	for _, pos := range positions {
		lat := pos.Latitude
		lng := pos.Longitude
		if lat == 0 && lng == 0 && pos.Geom != "" {
			parsedLat, parsedLng, err := parseWKTPoint(pos.Geom)
			if err == nil {
				lat = parsedLat
				lng = parsedLng
			}
		}

		if lat == 0 && lng == 0 {
			continue
		}

		parsedAttrs := extractAttributes(pos.Attributes)

		responses = append(responses, dto.PositionCoordinateResponse{
			Latitude:   lat,
			Longitude:  lng,
			ServerTime: localTime(pos.ServerTime),
			DeviceTime: localTime(pos.DeviceTime),
			Speed:      pos.Speed,
			Course:     pos.Course,
			Battery:    parsedAttrs.Battery,
			Ignition:   parsedAttrs.Ignition,
			Satellites: parsedAttrs.Satellites,
		})
	}

	return responses, nil
}

// extractAttributes parsea el JSON de atributos del dispositivo y retorna campos individuales
type attrFields struct {
	Battery    *float64
	Ignition   *bool
	Satellites *int
}

func extractAttributes(raw *string) attrFields {
	if raw == nil || *raw == "" {
		return attrFields{}
	}
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(*raw), &data); err != nil {
		return attrFields{}
	}

	var a attrFields

	if v, ok := data["battery"]; ok {
		if f, ok := v.(float64); ok {
			a.Battery = &f
		}
	}
	if v, ok := data["ignition"]; ok {
		if b, ok := v.(bool); ok {
			a.Ignition = &b
		}
	}
	if v, ok := data["satellites"]; ok {
		if f, ok := v.(float64); ok {
			s := int(f)
			a.Satellites = &s
		}
	}

	return a
}

func parseWKTPoint(point string) (lat, lng float64, err error) {
	if point == "" {
		return 0, 0, fmt.Errorf("point is empty")
	}
	str := strings.ToUpper(strings.TrimSpace(point))
	if !strings.HasPrefix(str, "POINT") {
		return 0, 0, fmt.Errorf("invalid WKT format: missing POINT prefix")
	}
	str = strings.TrimPrefix(str, "POINT")
	str = strings.TrimSpace(str)
	str = strings.TrimPrefix(str, "(")
	str = strings.TrimSuffix(str, ")")
	str = strings.TrimSpace(str)
	coords := strings.Fields(str)
	if len(coords) != 2 {
		return 0, 0, fmt.Errorf("invalid point format, expected 2 coordinates")
	}
	lng, err = strconv.ParseFloat(coords[0], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("error parsing longitude: %v", err)
	}
	lat, err = strconv.ParseFloat(coords[1], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("error parsing latitude: %v", err)
	}
	return lat, lng, nil
}

func (s *positionService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

func (s *positionService) GetDeviceHistory(ctx context.Context, imei string, dateStr string) (*dto.HistoryResponse, error) {
	// 0. Validar existencia del dispositivo
	_, err := s.repo.FindLastByIMEI(ctx, imei)
	if err != nil {
		return nil, fmt.Errorf("dispositivo no encontrado: %s", imei)
	}

	// 1. Parsear fechas (Todo el día: 00:00 a 23:59)
	startDay, err := utils.ParseLocalDate(dateStr)
	if err != nil {
		return nil, fmt.Errorf("fecha invalida: %v", err)
	}
	endDay := startDay.Add(24 * time.Hour)

	return s.analyzeDeviceHistory(ctx, imei, dateStr, startDay, endDay)
}

func (s *positionService) GetDeviceRoute(ctx context.Context, imei string, start, end time.Time) (*dto.HistoryResponse, error) {
	// 0. Validar existencia del dispositivo
	_, err := s.repo.FindLastByIMEI(ctx, imei)
	if err != nil {
		return nil, fmt.Errorf("dispositivo no encontrado: %s", imei)
	}

	// Use start date as the "date" label for now, or a custom string
	dateLabel := start.Format("2006-01-02 15:04") + " - " + end.Format("15:04")
	if start.YearDay() != end.YearDay() {
		dateLabel = start.Format("2006-01-02 15:04") + " - " + end.Format("2006-01-02 15:04")
	}
	return s.analyzeDeviceHistory(ctx, imei, dateLabel, start, end)
}

func (s *positionService) analyzeDeviceHistory(ctx context.Context, imei string, label string, start, end time.Time) (*dto.HistoryResponse, error) {
	// 2. Traer puntos crudos
	positions, err := s.repo.FindForHistory(ctx, imei, start, end)
	if err != nil {
		return nil, err
	}

	res, err := s.analyzePositions(ctx, imei, "", label, positions)
	if err != nil {
		return nil, err
	}

	// Convertir VehicleHistoryResponse de vuelta a HistoryResponse para la función deprecada
	return &dto.HistoryResponse{
		IMEI:          res.VehicleID, // Aquí VehicleID es el IMEI
		Date:          res.Date,
		TotalDistance: res.TotalDistance,
		TotalDuration: res.TotalDuration,
		MaxSpeed:      res.MaxSpeed,
		Events:        res.Events,
	}, nil
}

// ---------------------------------------------------------------------------
// Vehículos - Historial (Correcto)
// ---------------------------------------------------------------------------

func (s *positionService) GetVehicleHistory(ctx context.Context, vehicleID string, dateStr string) (*dto.VehicleHistoryResponse, error) {
	parsedUUID, err := uuid.Parse(vehicleID)
	if err != nil {
		return nil, fmt.Errorf("uuid de vehículo inválido: %v", err)
	}

	start, err := utils.ParseLocalDate(dateStr)
	if err != nil {
		return nil, fmt.Errorf("fecha inválida: %v", err)
	}
	end := start.Add(24 * time.Hour)

	slots, err := s.repo.FindSlotsByVehicleAndRange(ctx, parsedUUID, start, end)
	if err != nil {
		return nil, err
	}
	if len(slots) == 0 {
		return &dto.VehicleHistoryResponse{VehicleID: vehicleID, Date: dateStr, Events: []dto.TimelineEvent{}}, nil
	}

	positions, err := s.repo.FindForHistoryBySlots(ctx, slots, start, end)
	if err != nil {
		return nil, err
	}

	// Opcional: Obtener placa para la respuesta
	placa := "" // TODO: se podría buscar si se necesita.
	return s.analyzePositions(ctx, vehicleID, placa, dateStr, positions)
}

func (s *positionService) GetVehicleRoute(ctx context.Context, vehicleID string, start, end time.Time) (*dto.VehicleHistoryResponse, error) {
	parsedUUID, err := uuid.Parse(vehicleID)
	if err != nil {
		return nil, fmt.Errorf("uuid de vehículo inválido: %v", err)
	}

	dateLabel := start.Format("2006-01-02 15:04") + " - " + end.Format("15:04")
	if start.YearDay() != end.YearDay() {
		dateLabel = start.Format("2006-01-02 15:04") + " - " + end.Format("2006-01-02 15:04")
	}

	slots, err := s.repo.FindSlotsByVehicleAndRange(ctx, parsedUUID, start, end)
	if err != nil {
		return nil, err
	}
	if len(slots) == 0 {
		return &dto.VehicleHistoryResponse{VehicleID: vehicleID, Date: dateLabel, Events: []dto.TimelineEvent{}}, nil
	}

	positions, err := s.repo.FindForHistoryBySlots(ctx, slots, start, end)
	if err != nil {
		return nil, err
	}

	return s.analyzePositions(ctx, vehicleID, "", dateLabel, positions)
}

func (s *positionService) GetLastPositionOfVehicle(ctx context.Context, vehicleID string) (dto.VehiclePositionResponse, error) {
	parsedUUID, err := uuid.Parse(vehicleID)
	if err != nil {
		return dto.VehiclePositionResponse{}, fmt.Errorf("uuid de vehículo inválido: %v", err)
	}

	pos, err := s.repo.FindLastPositionByVehicle(ctx, parsedUUID)
	if err != nil {
		return dto.VehiclePositionResponse{}, err
	}

	return dto.VehiclePositionResponse{
		VehicleID:  pos.VehicleID,
		Placa:      pos.Placa,
		Imei:       pos.Imei,
		Latitude:   pos.Latitude,
		Longitude:  pos.Longitude,
		Speed:      pos.Speed,
		Course:     pos.Course,
		DeviceTime: localTime(pos.DeviceTime),
		ServerTime: localTime(pos.ServerTime),
		Attributes: pos.Attributes,
	}, nil
}

func (s *positionService) GetLastPositionsOfAllVehicles(ctx context.Context) ([]dto.VehiclePositionResponse, error) {
	positions, err := s.repo.FindLastPositionsByVehicles(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.VehiclePositionResponse, len(positions))
	for i, pos := range positions {
		responses[i] = dto.VehiclePositionResponse{
			VehicleID:  pos.VehicleID,
			Placa:      pos.Placa,
			Imei:       pos.Imei,
			Latitude:   pos.Latitude,
			Longitude:  pos.Longitude,
			Speed:      pos.Speed,
			Course:     pos.Course,
			DeviceTime: localTime(pos.DeviceTime),
			ServerTime: localTime(pos.ServerTime),
			Attributes: pos.Attributes,
		}
	}

	return responses, nil
}

// analyzePositions is the core state machine, agnostic to how positions were fetched.
func (s *positionService) analyzePositions(ctx context.Context, id string, title string, label string, positions []entities.Position) (*dto.VehicleHistoryResponse, error) {
	if len(positions) == 0 {
		return &dto.VehicleHistoryResponse{VehicleID: id, Placa: title, Date: label, Events: []dto.TimelineEvent{}}, nil
	}

	// Convertir todos los timestamps a hora local (Bolivia UTC-4) antes de
	// procesar la máquina de estados. Así todos los StartTime/EndTime/Timestamp
	// que se emiten en la respuesta ya están en hora local.
	for i := range positions {
		positions[i].DeviceTime = positions[i].DeviceTime.In(utils.BoliviaLocation)
		positions[i].ServerTime = positions[i].ServerTime.In(utils.BoliviaLocation)
	}

	// 3. Variables de Estado
	var events []dto.TimelineEvent
	var tripPositions []entities.Position // Puntos completos para extraer metadata

	// Punteros para lógica de estado
	var potentialStopStart *entities.Position
	// Métricas del viaje actual
	currentTripMaxSpeed := 0
	currentTripSpeedSum := 0
	currentTripCount := 0

	// Inicio del bloque actual (sea viaje o parada)
	blockStartTime := positions[0].DeviceTime

	// 4. Determinar estado inicial basado en el primer punto
	currentState := "stopped" // Por defecto
	if positions[0].Speed > MinMovingSpeed {
		currentState = "moving"
		tripPositions = append(tripPositions, positions[0])
		currentTripMaxSpeed = positions[0].Speed
		currentTripSpeedSum = positions[0].Speed
		currentTripCount = 1
	}

	// 5. Iteración (State Machine) - Empezamos desde el segundo punto
	for i := 1; i < len(positions); i++ {
		pos := positions[i]
		isMoving := pos.Speed > MinMovingSpeed

		// --- A. DETECCIÓN DE GAP (Pérdida de señal) ---
		timeDiff := pos.DeviceTime.Sub(positions[i-1].DeviceTime)
		if timeDiff > MaxGapDuration {
			// Cerrar lo que estuviera abierto
			if currentState == "moving" {
				s.finalizeTrip(ctx, &events, tripPositions, blockStartTime, positions[i-1].DeviceTime, currentTripMaxSpeed, currentTripSpeedSum, currentTripCount)
			} else {
				s.finalizeStop(&events, positions[i-1], blockStartTime, positions[i-1].DeviceTime)
			}

			// Resetear estado tras el gap
			currentState = "stopped"
			if isMoving {
				currentState = "moving"
			}
			blockStartTime = pos.DeviceTime
			tripPositions = []entities.Position{}
			if currentState == "moving" {
				tripPositions = append(tripPositions, pos)
				currentTripMaxSpeed = pos.Speed
				currentTripSpeedSum = pos.Speed
				currentTripCount = 1
			}
			potentialStopStart = nil
			continue
		}

		// --- B. LÓGICA DE TRANSICIÓN ---
		switch currentState {
		case "moving":
			if isMoving {
				// Seguimos moviéndonos
				tripPositions = append(tripPositions, pos)
				if pos.Speed > currentTripMaxSpeed {
					currentTripMaxSpeed = pos.Speed
				}
				currentTripSpeedSum += pos.Speed
				currentTripCount++
				potentialStopStart = nil // Cancelamos posible parada
			} else {
				// El auto se detuvo momentáneamente
				if potentialStopStart == nil {
					potentialStopStart = &pos
				}

				// ¿Cuánto tiempo lleva detenido?
				stopDuration := pos.DeviceTime.Sub(potentialStopStart.DeviceTime)

				if stopDuration > MinStopDuration {
					// CONFIRMADO: ES UNA PARADA REAL
					s.finalizeTrip(ctx, &events, tripPositions, blockStartTime, potentialStopStart.DeviceTime, currentTripMaxSpeed, currentTripSpeedSum, currentTripCount)

					// 2. Cambiamos estado
					currentState = "stopped"
					blockStartTime = potentialStopStart.DeviceTime
					tripPositions = []entities.Position{}
					potentialStopStart = nil
				} else {
					// Todavía es una parada corta (semáforo), lo tratamos como parte del viaje
					tripPositions = append(tripPositions, pos)
				}
			}
		case "stopped":
			if isMoving {
				// EMPEZÓ A MOVERSE
				s.finalizeStop(&events, positions[i-1], blockStartTime, positions[i-1].DeviceTime)

				// 2. Cambiamos estado
				currentState = "moving"
				blockStartTime = pos.DeviceTime
				tripPositions = []entities.Position{pos}
				currentTripMaxSpeed = pos.Speed
				currentTripSpeedSum = pos.Speed
				currentTripCount = 1
				potentialStopStart = nil
			} else {
				// Sigue detenido
			}
		}
	}

	// 5. Cerrar el último evento pendiente
	lastPos := positions[len(positions)-1]
	if currentState == "moving" {
		s.finalizeTrip(ctx, &events, tripPositions, blockStartTime, lastPos.DeviceTime, currentTripMaxSpeed, currentTripSpeedSum, currentTripCount)
	} else {
		s.finalizeStop(&events, lastPos, blockStartTime, lastPos.DeviceTime)
	}

	// 6. Calcular Totales Generales
	var totalDist float64
	var maxSpeedGlobal int

	for _, e := range events {
		if e.Type == "trip" {
			data := e.Data.(dto.TripData)
			totalDist += data.DistanceKm
			if data.MaxSpeed > maxSpeedGlobal {
				maxSpeedGlobal = data.MaxSpeed
			}
		}
	}

	// 7. Retornar Respuesta
	resp := &dto.VehicleHistoryResponse{
		VehicleID:     id,
		Placa:         title,
		Date:          label,
		TotalDistance: math.Round(totalDist*100) / 100, // Redondear 2 decimales
		TotalDuration: fmtDuration(positions[len(positions)-1].DeviceTime.Sub(positions[0].DeviceTime)),
		MaxSpeed:      maxSpeedGlobal,
		Events:        events,
	}

	return resp, nil
}

// --- FUNCIONES AUXILIARES (HELPERS) ---

func (s *positionService) finalizeStop(events *[]dto.TimelineEvent, pos entities.Position, start, end time.Time) {
	duration := end.Sub(start)
	// Solo agregamos si la parada tiene duración relevante (> 1 min)
	if duration > time.Minute {
		fmtDur := fmtDuration(duration)
		battery := s.extractBatteryLevel(pos.Attributes)
		*events = append(*events, dto.TimelineEvent{
			Type:      "stop",
			StartTime: start,
			EndTime:   end,
			Duration:  fmtDur,
			Data: dto.StopData{
				Latitude:     pos.Latitude,
				Longitude:    pos.Longitude,
				BatteryLevel: battery,
				StartTime:    start,
				EndTime:      end,
				Duration:     fmtDur,
				// Aquí podrías llamar a una función de Reverse Geocoding async
			},
		})
	}
}

func (s *positionService) finalizeTrip(_ context.Context, events *[]dto.TimelineEvent, positions []entities.Position, start, end time.Time, maxSpeed, speedSum, count int) {
	if len(positions) < 2 {
		return
	}

	var routePoints []dto.RoutePoint
	var coords [][]float64
	var distKm float64

	for i, p := range positions {
		routePoints = append(routePoints, dto.RoutePoint{
			Latitude:     p.Latitude,
			Longitude:    p.Longitude,
			Speed:        p.Speed,
			Course:       p.Course,
			Timestamp:    p.DeviceTime.Format(time.RFC3339),
			BatteryLevel: s.extractBatteryLevel(p.Attributes),
		})

		coords = append(coords, []float64{p.Longitude, p.Latitude})

		if i > 0 {
			prev := positions[i-1]
			distKm += haversineDistance(prev.Latitude, prev.Longitude, p.Latitude, p.Longitude)
		}
	}

	geoJSON := map[string]interface{}{
		"type":        "LineString",
		"coordinates": coords,
	}

	avgSpeed := 0
	if count > 0 {
		avgSpeed = speedSum / count
	}

	fmtDur := fmtDuration(end.Sub(start))
	*events = append(*events, dto.TimelineEvent{
		Type:      "trip",
		StartTime: start,
		EndTime:   end,
		Duration:  fmtDur,
		Data: dto.TripData{
			DistanceKm:  math.Round(distKm*100) / 100, // Redondear a 2 decimales
			MaxSpeed:    maxSpeed,
			AvgSpeed:    avgSpeed,
			PathGeoJSON: geoJSON, // Formato objeto JSON directo
			Points:      routePoints,
			StartTime:   start,
			EndTime:     end,
			Duration:    fmtDur,
		},
	})
}

func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0 // Earth radius in km
	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLon := (lon2 - lon1) * math.Pi / 180.0
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180.0)*math.Cos(lat2*math.Pi/180.0)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

func fmtDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}

func (s *positionService) extractBatteryLevel(attrs *string) float64 {
	if attrs == nil {
		return 0
	}
	var data map[string]interface{}
	// Intenta parsear como JSON
	if err := json.Unmarshal([]byte(*attrs), &data); err != nil {
		return 0
	}

	// Buscar claves comunes de batería
	if val, ok := data["batteryLevel"]; ok {
		if v, ok := val.(float64); ok {
			return v
		}
	}
	if val, ok := data["battery"]; ok {
		if v, ok := val.(float64); ok {
			return v
		}
	}
	if val, ok := data["power"]; ok {
		if v, ok := val.(float64); ok {
			return v
		}
	}

	return 0
}
