package service

import (
	"context"
	"sort"
	"time"

	"github.com/Caknoooo/go-gin-clean-starter/modules/route/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/route/repository"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/geo"
	redisProvider "github.com/Caknoooo/go-gin-clean-starter/providers/redis"
	"github.com/google/uuid"
	"github.com/samber/do"
)

const (
	// recentPositionThreshold: más allá de esto, un micro no se considera "en ruta"
	// aunque tenga una asignación activa (RF-02/RF-13: solo los que están trabajando).
	recentPositionThreshold = 2 * time.Minute

	// defaultAverageSpeedKmh se usa para estimar el ETA cuando el micro está detenido/lento
	// y la ruta no tiene suficiente información (allowed_lap_duration_seconds) para inferir
	// una velocidad promedio propia.
	defaultAverageSpeedKmh = 20.0

	// minEtaSpeedKmh evita ETAs absurdamente largos (o división por ~0) cuando el micro
	// está detenido justo en el momento de la consulta.
	minEtaSpeedKmh = 8.0

	// maxDistanceToRouteMeters: más allá de esto, la posición del micro no se considera
	// sobre el trazado conocido y no se usa para estimar su ETA.
	maxDistanceToRouteMeters = 100.0
)

type RouteLiveService interface {
	FindLiveVehicles(ctx context.Context, routeID uuid.UUID) ([]dto.LiveVehicleResponse, error)
	FindETA(ctx context.Context, routeID uuid.UUID, latitude, longitude float64) ([]dto.EtaResponse, error)
}

type routeLiveService struct {
	routeRepo repository.RouteRepository
	liveRepo  repository.RouteLiveRepository
	posCache  redisProvider.DevicePositionCache
}

func NewRouteLiveService(injector *do.Injector) (RouteLiveService, error) {
	routeRepo := do.MustInvoke[repository.RouteRepository](injector)
	liveRepo := do.MustInvoke[repository.RouteLiveRepository](injector)

	var posCache redisProvider.DevicePositionCache
	if redisSvc, err := do.InvokeNamed[redisProvider.RedisService](injector, "Redis"); err == nil {
		posCache = redisProvider.NewDevicePositionCache(redisSvc)
	}

	return &routeLiveService{
		routeRepo: routeRepo,
		liveRepo:  liveRepo,
		posCache:  posCache,
	}, nil
}

func (s *routeLiveService) FindLiveVehicles(ctx context.Context, routeID uuid.UUID) ([]dto.LiveVehicleResponse, error) {
	if s.posCache == nil {
		return []dto.LiveVehicleResponse{}, nil
	}

	vehicles, err := s.liveRepo.FindActiveVehiclesForRoute(ctx, routeID)
	if err != nil {
		return nil, err
	}
	if len(vehicles) == 0 {
		return []dto.LiveVehicleResponse{}, nil
	}

	imeis := make([]string, 0, len(vehicles))
	for _, v := range vehicles {
		imeis = append(imeis, v.IMEI)
	}

	positions, err := s.posCache.MGet(ctx, imeis)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	responses := make([]dto.LiveVehicleResponse, 0, len(vehicles))
	for _, v := range vehicles {
		pos, ok := positions[v.IMEI]
		if !ok {
			continue // sin posición conocida todavía
		}
		if now.Sub(pos.ServerTime) > recentPositionThreshold {
			continue // posición muy vieja: no se considera "en ruta" en este momento
		}

		resp := dto.LiveVehicleResponse{
			VehicleID:          v.VehicleID,
			PinNumber:          v.PinNumber,
			Latitude:           pos.Latitude,
			Longitude:          pos.Longitude,
			SpeedKmh:           pos.Speed,
			LastUpdateAt:       pos.ServerTime,
			SecondsSinceUpdate: int(now.Sub(pos.ServerTime).Seconds()),
		}

		if lap, err := s.liveRepo.FindOpenLap(ctx, v.VehicleID); err == nil && lap != nil {
			resp.LapNumber = &lap.LapNumber
			resp.LapStatus = &lap.LapStatus
		}

		responses = append(responses, resp)
	}

	return responses, nil
}

func (s *routeLiveService) FindETA(ctx context.Context, routeID uuid.UUID, latitude, longitude float64) ([]dto.EtaResponse, error) {
	route, err := s.routeRepo.FindByID(ctx, routeID)
	if err != nil {
		return nil, err
	}
	if route.Geometry == nil {
		return []dto.EtaResponse{}, nil
	}

	polylines, err := geo.ParsePolylines(*route.Geometry)
	if err != nil || len(polylines) == 0 {
		return []dto.EtaResponse{}, nil
	}

	queryPoint := geo.Point{Latitude: latitude, Longitude: longitude}
	queryIdx, queryProgress, _ := geo.BestProjection(queryPoint, polylines)
	if queryIdx == -1 {
		return []dto.EtaResponse{}, nil
	}
	polylineLength := geo.PolylineLength(polylines[queryIdx])

	averageSpeedKmh := defaultAverageSpeedKmh
	if route.AllowedLapDurationSeconds != nil && *route.AllowedLapDurationSeconds > 0 && polylineLength > 0 {
		averageSpeedKmh = (polylineLength / 1000) / (float64(*route.AllowedLapDurationSeconds) / 3600)
	}

	liveVehicles, err := s.FindLiveVehicles(ctx, routeID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.EtaResponse, 0, len(liveVehicles))
	for _, v := range liveVehicles {
		vIdx, vProgress, vDist := geo.BestProjection(geo.Point{Latitude: v.Latitude, Longitude: v.Longitude}, polylines)
		if vIdx != queryIdx || vDist > maxDistanceToRouteMeters {
			continue // no está sobre el mismo sentido/trazado conocido
		}

		remaining := queryProgress - vProgress
		if remaining < 0 {
			// ya pasó por este punto en la vuelta actual: falta que complete el resto
			// del trazado y vuelva a llegar (la ruta es un lazo cerrado por vuelta)
			remaining = (polylineLength - vProgress) + queryProgress
		}

		speedKmh := float64(v.SpeedKmh)
		if speedKmh < minEtaSpeedKmh {
			speedKmh = averageSpeedKmh
		}
		speedMetersPerSecond := (speedKmh * 1000) / 3600
		if speedMetersPerSecond <= 0 {
			continue
		}

		responses = append(responses, dto.EtaResponse{
			VehicleID:      v.VehicleID,
			PinNumber:      v.PinNumber,
			DistanceMeters: remaining,
			EtaSeconds:     int(remaining / speedMetersPerSecond),
		})
	}

	sort.Slice(responses, func(i, j int) bool { return responses[i].EtaSeconds < responses[j].EtaSeconds })

	return responses, nil
}
