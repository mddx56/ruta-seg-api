package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	fineservice "github.com/Caknoooo/go-gin-clean-starter/modules/fine/service"
	"github.com/Caknoooo/go-gin-clean-starter/modules/lap/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/lap/repository"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/geo"
	providerWS "github.com/Caknoooo/go-gin-clean-starter/providers/websocket"
	"github.com/google/uuid"
	"github.com/samber/do"
)

const (
	// minLapDurationSeconds evita que el ruido del GPS cerca del checkpoint genere
	// vueltas duplicadas de duración casi nula.
	minLapDurationSeconds = 180

	// movingSpeedThresholdKmh: por debajo de esto se considera que el micro está detenido.
	movingSpeedThresholdKmh = 5

	// defaultMaxStopDurationSeconds se usa si la ruta no define su propio umbral.
	defaultMaxStopDurationSeconds = 300 // 5 minutos

	// speedFineCooldown evita generar una multa de velocidad por cada posición mientras
	// el exceso de velocidad se mantiene sostenido.
	speedFineCooldown = 5 * time.Minute

	// overtakingProximityToleranceMeters tolera el ruido normal de GPS al comparar el
	// avance de dos micros sobre la misma polilínea.
	overtakingProximityToleranceMeters = 30.0

	// maxDistanceToRouteMeters: más allá de esto, la posición no se considera "sobre la ruta"
	// y no se usa para comparar el orden entre micros (evita falsos positivos por desvíos/GPS).
	maxDistanceToRouteMeters = 100.0
)

type LapService interface {
	EvaluatePosition(ctx context.Context, vehicleID uuid.UUID, latitude, longitude float64, speedKmh int, eventTime time.Time) error
	FindAll(ctx context.Context) ([]dto.LapResponse, error)
	FindByID(ctx context.Context, id uuid.UUID) (dto.LapResponse, error)
}

type lapService struct {
	repo           repository.LapRepository
	fineService    fineservice.FineService
	wsService      providerWS.WebsocketService
	routePublisher providerWS.RouteEventPublisher
}

func NewLapService(injector *do.Injector) (LapService, error) {
	repo := do.MustInvoke[repository.LapRepository](injector)
	fineSvc := do.MustInvoke[fineservice.FineService](injector)
	wsSvc, _ := do.Invoke[providerWS.WebsocketService](injector)
	routePublisher, _ := do.Invoke[providerWS.RouteEventPublisher](injector)
	return &lapService{
		repo:           repo,
		fineService:    fineSvc,
		wsService:      wsSvc,
		routePublisher: routePublisher,
	}, nil
}

// EvaluatePosition es el motor de reglas: se llama por cada posición nueva que llega
// de un dispositivo. Si el vehículo está asignado a una ruta activa, evalúa las
// infracciones (parada prolongada, exceso de velocidad, adelantamiento) sobre la vuelta
// en curso y, si la posición cae dentro del checkpoint de la ruta, cierra la vuelta y abre una nueva.
func (s *lapService) EvaluatePosition(ctx context.Context, vehicleID uuid.UUID, latitude, longitude float64, speedKmh int, eventTime time.Time) error {
	vehicleRoute, err := s.repo.FindActiveVehicleRoute(ctx, vehicleID)
	if err != nil {
		return err
	}
	if vehicleRoute == nil {
		return nil // el micro no está asignado a ninguna ruta activa
	}

	route, err := s.repo.FindRouteByID(ctx, vehicleRoute.RouteID)
	if err != nil {
		return err
	}
	if route == nil {
		return nil
	}

	openLap, err := s.repo.FindOpenLap(ctx, vehicleID)
	if err != nil {
		return err
	}

	s.publishMicroLive(ctx, route.ID, vehicleID, vehicleRoute.PinNumber, latitude, longitude, speedKmh, openLap, eventTime)

	if openLap != nil {
		if err := s.evaluateInfractions(ctx, openLap, route, speedKmh, latitude, longitude, eventTime); err != nil {
			log.Printf("[lap-engine] error evaluando infracciones para lap %s: %v", openLap.ID, err)
		}
	}

	if route.CheckpointGeofenceID == nil {
		return nil // la ruta no tiene checkpoint de vuelta configurado todavía
	}

	checkpoint, err := s.repo.FindGeofenceByID(ctx, *route.CheckpointGeofenceID)
	if err != nil {
		return err
	}
	if checkpoint == nil || !isInsideGeofence(checkpoint, latitude, longitude) {
		return nil
	}

	if openLap != nil {
		if eventTime.Sub(openLap.StartedAt) < minLapDurationSeconds*time.Second {
			return nil // demasiado cerca del inicio de la vuelta actual, probablemente ruido de GPS
		}

		if err := s.closeLap(ctx, openLap, route, vehicleRoute, latitude, longitude, eventTime); err != nil {
			return err
		}
	}

	nextLapNumber := 1
	if openLap != nil {
		nextLapNumber = openLap.LapNumber + 1
	}

	newLap := entities.Lap{
		VehicleID:              vehicleID,
		RouteID:                vehicleRoute.RouteID,
		LapNumber:              nextLapNumber,
		StartedAt:              eventTime,
		LapStatus:              "IN_PROGRESS",
		AllowedDurationSeconds: route.AllowedLapDurationSeconds,
		LastMovementAt:         &eventTime,
	}

	return s.repo.Create(ctx, &newLap)
}

// evaluateInfractions revisa parada prolongada, exceso de velocidad y adelantamiento
// sobre la vuelta en curso, y persiste el estado de seguimiento si algo cambió.
func (s *lapService) evaluateInfractions(ctx context.Context, lap *entities.Lap, route *entities.Route, speedKmh int, latitude, longitude float64, eventTime time.Time) error {
	changed := s.evaluateProlongedStop(ctx, lap, route, speedKmh, latitude, longitude, eventTime)

	if s.evaluateSpeeding(ctx, lap, route, speedKmh, latitude, longitude, eventTime) {
		changed = true
	}

	if route.Geometry != nil {
		overtakingChanged, err := s.evaluateOvertaking(ctx, lap, route, latitude, longitude, eventTime)
		if err != nil {
			log.Printf("[lap-engine] error evaluando adelantamiento para lap %s: %v", lap.ID, err)
		}
		if overtakingChanged {
			changed = true
		}
	}

	if !changed {
		return nil
	}
	return s.repo.UpdateTracking(ctx, lap)
}

// evaluateProlongedStop implementa RF-19. Devuelve true si algún campo de seguimiento cambió.
func (s *lapService) evaluateProlongedStop(ctx context.Context, lap *entities.Lap, route *entities.Route, speedKmh int, latitude, longitude float64, eventTime time.Time) bool {
	if speedKmh > movingSpeedThresholdKmh {
		changed := false
		if lap.LastStopFineAt != nil {
			lap.LastStopFineAt = nil
			changed = true
		}
		if lap.LastMovementAt == nil || eventTime.After(*lap.LastMovementAt) {
			lap.LastMovementAt = &eventTime
			changed = true
		}
		return changed
	}

	if lap.LastMovementAt == nil {
		return false
	}

	threshold := defaultMaxStopDurationSeconds
	if route.MaxStopDurationSeconds != nil && *route.MaxStopDurationSeconds > 0 {
		threshold = *route.MaxStopDurationSeconds
	}

	if eventTime.Sub(*lap.LastMovementAt) <= time.Duration(threshold)*time.Second {
		return false
	}
	if lap.LastStopFineAt != nil {
		return false // ya se multó esta parada, se espera a que retome movimiento
	}

	lapID := lap.ID
	if _, err := s.fineService.GenerateFine(ctx, fineservice.GenerateFineInput{
		VehicleID:    lap.VehicleID,
		FineTypeCode: "PROLONGED_STOP",
		LapID:        &lapID,
		Latitude:     latitude,
		Longitude:    longitude,
		OccurredAt:   eventTime,
	}); err != nil {
		log.Printf("[lap-engine] error generando multa PROLONGED_STOP para lap %s: %v", lap.ID, err)
		return false
	}

	lap.LastStopFineAt = &eventTime
	return true
}

// evaluateSpeeding implementa RF-20. Devuelve true si algún campo de seguimiento cambió.
func (s *lapService) evaluateSpeeding(ctx context.Context, lap *entities.Lap, route *entities.Route, speedKmh int, latitude, longitude float64, eventTime time.Time) bool {
	if route.MaxSpeedKmh == nil || speedKmh <= *route.MaxSpeedKmh {
		return false
	}

	if lap.LastSpeedFineAt != nil && eventTime.Sub(*lap.LastSpeedFineAt) < speedFineCooldown {
		return false
	}

	lapID := lap.ID
	notes := fmt.Sprintf("%d km/h (límite %d km/h)", speedKmh, *route.MaxSpeedKmh)
	if _, err := s.fineService.GenerateFine(ctx, fineservice.GenerateFineInput{
		VehicleID:    lap.VehicleID,
		FineTypeCode: "SPEEDING",
		LapID:        &lapID,
		Latitude:     latitude,
		Longitude:    longitude,
		OccurredAt:   eventTime,
		Notes:        &notes,
	}); err != nil {
		log.Printf("[lap-engine] error generando multa SPEEDING para lap %s: %v", lap.ID, err)
		return false
	}

	lap.LastSpeedFineAt = &eventTime
	return true
}

// evaluateOvertaking implementa RF-18: ubica al micro sobre el trazado de la ruta
// (map-matching) y compara su avance contra el de otros micros con vuelta en curso en
// la misma ruta y sentido. Si un micro que arrancó su vuelta después ya tiene más avance
// que uno que arrancó antes (más allá de la tolerancia), se considera que lo adelantó.
func (s *lapService) evaluateOvertaking(ctx context.Context, lap *entities.Lap, route *entities.Route, latitude, longitude float64, eventTime time.Time) (bool, error) {
	polylines, err := geo.ParsePolylines(*route.Geometry)
	if err != nil || len(polylines) == 0 {
		return false, err
	}

	idx, progress, distToLine := geo.BestProjection(geo.Point{Latitude: latitude, Longitude: longitude}, polylines)
	if idx == -1 || distToLine > maxDistanceToRouteMeters {
		return false, nil // el micro no está sobre un trazado conocido en este momento
	}

	previousIdx := lap.LastPolylineIndex
	lap.LastPolylineIndex = &idx
	lap.LastProgressMeters = &progress
	lap.LastProgressAt = &eventTime

	if previousIdx == nil || *previousIdx != idx {
		// primera lectura o cambio de sentido: todavía no hay avance previo comparable
		return true, nil
	}

	otherLaps, err := s.repo.FindOpenLapsByRoute(ctx, route.ID, lap.VehicleID)
	if err != nil {
		return true, err
	}

	for i := range otherLaps {
		other := &otherLaps[i]
		if other.LastPolylineIndex == nil || *other.LastPolylineIndex != idx || other.LastProgressMeters == nil {
			continue // no está en el mismo sentido, o todavía no tiene una posición comparable
		}
		if other.OvertakingFined {
			continue // ya se multó a este micro por adelantamiento en su vuelta actual
		}

		// Si "lap" arrancó su vuelta antes que "other", "other" no debería tener más avance.
		if lap.StartedAt.Before(other.StartedAt) && *other.LastProgressMeters > progress+overtakingProximityToleranceMeters {
			otherLapID := other.ID
			if _, err := s.fineService.GenerateFine(ctx, fineservice.GenerateFineInput{
				VehicleID:    other.VehicleID,
				FineTypeCode: "OVERTAKING",
				LapID:        &otherLapID,
				Latitude:     latitude,
				Longitude:    longitude,
				OccurredAt:   eventTime,
			}); err != nil {
				log.Printf("[lap-engine] error generando multa OVERTAKING para lap %s: %v", other.ID, err)
				continue
			}

			other.OvertakingFined = true
			if err := s.repo.UpdateTracking(ctx, other); err != nil {
				log.Printf("[lap-engine] error marcando OvertakingFined para lap %s: %v", other.ID, err)
			}
		}
	}

	return true, nil
}

func (s *lapService) closeLap(ctx context.Context, lap *entities.Lap, route *entities.Route, vehicleRoute *entities.VehicleRoute, latitude, longitude float64, eventTime time.Time) error {
	duration := int(eventTime.Sub(lap.StartedAt).Seconds())
	status := computeLapStatus(duration, lap.AllowedDurationSeconds)

	lap.EndedAt = &eventTime
	lap.DurationSeconds = &duration
	lap.LapStatus = status

	if err := s.repo.CloseLap(ctx, lap); err != nil {
		return err
	}

	if status == "LATE" {
		lapID := lap.ID
		if _, err := s.fineService.GenerateFine(ctx, fineservice.GenerateFineInput{
			VehicleID:    lap.VehicleID,
			FineTypeCode: "LAP_TIME",
			LapID:        &lapID,
			Latitude:     latitude,
			Longitude:    longitude,
			OccurredAt:   eventTime,
		}); err != nil {
			log.Printf("[lap-engine] error generando multa LAP_TIME para lap %s: %v", lap.ID, err)
		}
	}

	if fare, err := s.repo.FindActiveFare(ctx, route.ID, eventTime); err != nil {
		log.Printf("[lap-engine] error buscando tarifa activa para route %s: %v", route.ID, err)
	} else if fare != nil {
		charge := entities.LapCharge{
			LapID:        lap.ID,
			Amount:       fare.AmountPerLap,
			ChargeStatus: "PENDING",
		}
		if err := s.repo.CreateCharge(ctx, &charge); err != nil {
			log.Printf("[lap-engine] error generando cobro para lap %s: %v", lap.ID, err)
		} else {
			lap.Charge = &charge
		}
	}

	s.broadcastLapCompleted(ctx, lap, vehicleRoute)

	return nil
}

// broadcastLapCompleted notifica que una vuelta se cerró por dos canales: al dueño del
// micro (+ admins) por el WebSocket privado existente, y al canal público de la ruta
// (TV / app pública) vía Redis, para que puedan mostrar el conteo de vueltas sin login.
func (s *lapService) broadcastLapCompleted(ctx context.Context, lap *entities.Lap, vehicleRoute *entities.VehicleRoute) {
	if s.wsService != nil {
		if ownerID, err := s.repo.FindVehicleOwnerID(ctx, lap.VehicleID); err == nil {
			event := struct {
				Event string          `json:"event"`
				Data  dto.LapResponse `json:"data"`
			}{
				Event: "LAP_COMPLETED",
				Data:  toResponse(*lap),
			}

			if payload, err := json.Marshal(event); err == nil {
				s.wsService.BroadcastToUsers([]string{ownerID.String()}, payload)
			}
		}
	}

	if s.routePublisher != nil {
		topic := "route:" + lap.RouteID.String()
		data := dto.MicroLapCompletedEvent{
			VehicleID: lap.VehicleID,
			PinNumber: vehicleRoute.PinNumber,
			Lap:       toResponse(*lap),
		}
		if err := s.routePublisher.Publish(ctx, topic, "LAP_COMPLETED", data); err != nil {
			log.Printf("[lap-engine] error publicando LAP_COMPLETED al canal público: %v", err)
		}
	}
}

// publishMicroLive emite la posición de un micro con ruta activa al canal público
// (best-effort: nunca bloquea ni falla el resto del motor de reglas).
func (s *lapService) publishMicroLive(ctx context.Context, routeID, vehicleID uuid.UUID, pinNumber string, latitude, longitude float64, speedKmh int, openLap *entities.Lap, eventTime time.Time) {
	if s.routePublisher == nil {
		return
	}

	event := dto.MicroLiveEvent{
		VehicleID: vehicleID,
		RouteID:   routeID,
		PinNumber: pinNumber,
		Latitude:  latitude,
		Longitude: longitude,
		SpeedKmh:  speedKmh,
		EventTime: eventTime,
	}
	if openLap != nil {
		event.LapNumber = &openLap.LapNumber
		event.LapStatus = &openLap.LapStatus
	}

	topic := "route:" + routeID.String()
	if err := s.routePublisher.Publish(ctx, topic, "MICRO_POSITION", event); err != nil {
		log.Printf("[lap-engine] error publicando MICRO_POSITION al canal público: %v", err)
	}
}

func computeLapStatus(durationSeconds int, allowedSeconds *int) string {
	if allowedSeconds == nil || *allowedSeconds <= 0 {
		return "ON_TIME"
	}
	if durationSeconds > *allowedSeconds {
		return "LATE"
	}
	if durationSeconds < *allowedSeconds/2 {
		return "TOO_FAST"
	}
	return "ON_TIME"
}

func isInsideGeofence(gf *entities.Geofence, lat, lon float64) bool {
	if gf == nil || len(gf.Points) == 0 {
		return false
	}

	switch gf.Type {
	case "CIRCLE":
		if gf.Radius == nil || *gf.Radius <= 0 {
			return false
		}
		center := gf.Points[0]
		return geo.HaversineMeters(lat, lon, center.Latitude, center.Longitude) <= *gf.Radius
	case "POLYGON":
		polygon := make([]geo.Point, 0, len(gf.Points))
		for _, p := range gf.Points {
			polygon = append(polygon, geo.Point{Latitude: p.Latitude, Longitude: p.Longitude})
		}
		return geo.PointInPolygon(geo.Point{Latitude: lat, Longitude: lon}, polygon)
	default:
		return false
	}
}

func (s *lapService) FindAll(ctx context.Context) ([]dto.LapResponse, error) {
	laps, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.LapResponse, 0, len(laps))
	for _, l := range laps {
		responses = append(responses, toResponse(l))
	}

	return responses, nil
}

func (s *lapService) FindByID(ctx context.Context, id uuid.UUID) (dto.LapResponse, error) {
	lap, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return dto.LapResponse{}, err
	}

	return toResponse(lap), nil
}

func toResponse(l entities.Lap) dto.LapResponse {
	res := dto.LapResponse{
		ID:                     l.ID,
		VehicleID:              l.VehicleID,
		RouteID:                l.RouteID,
		LapNumber:              l.LapNumber,
		StartedAt:              l.StartedAt,
		EndedAt:                l.EndedAt,
		DurationSeconds:        l.DurationSeconds,
		AllowedDurationSeconds: l.AllowedDurationSeconds,
		LapStatus:              l.LapStatus,
		CreatedAt:              l.CreatedAt,
	}

	if l.Charge != nil {
		res.Charge = &dto.LapChargeResponse{
			ID:           l.Charge.ID,
			Amount:       l.Charge.Amount,
			ChargeStatus: l.Charge.ChargeStatus,
			PaidAt:       l.Charge.PaidAt,
		}
	}

	return res
}
