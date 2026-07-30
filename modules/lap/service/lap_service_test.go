package service

import (
	"context"
	"testing"
	"time"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	finedto "github.com/Caknoooo/go-gin-clean-starter/modules/fine/dto"
	fineservice "github.com/Caknoooo/go-gin-clean-starter/modules/fine/service"
	"github.com/google/uuid"
)

func intPtr(v int) *int {
	return &v
}

func floatPtr(v float64) *float64 {
	return &v
}

// fakeFineService registra cada llamada a GenerateFine para poder verificarla en los tests,
// sin depender de una base de datos real.
type fakeFineService struct {
	generated []fineservice.GenerateFineInput
	err       error
}

func (f *fakeFineService) GenerateFine(_ context.Context, input fineservice.GenerateFineInput) (entities.Fine, error) {
	if f.err != nil {
		return entities.Fine{}, f.err
	}
	f.generated = append(f.generated, input)
	return entities.Fine{ID: uuid.New()}, nil
}

func (f *fakeFineService) Void(_ context.Context, _ uuid.UUID, _ *string) error { return nil }
func (f *fakeFineService) FindAll(_ context.Context) ([]finedto.FineResponse, error) {
	return nil, nil
}
func (f *fakeFineService) FindAllMine(_ context.Context, _ uuid.UUID) ([]finedto.FineResponse, error) {
	return nil, nil
}
func (f *fakeFineService) FindByID(_ context.Context, _ uuid.UUID) (finedto.FineResponse, error) {
	return finedto.FineResponse{}, nil
}
func (f *fakeFineService) FindAllTypes(_ context.Context) ([]finedto.FineTypeResponse, error) {
	return nil, nil
}

func TestComputeLapStatusNoAllowedDuration(t *testing.T) {
	if status := computeLapStatus(500, nil); status != "ON_TIME" {
		t.Fatalf("expected ON_TIME when no allowed duration is set, got %s", status)
	}
}

func TestComputeLapStatusOnTime(t *testing.T) {
	if status := computeLapStatus(1000, intPtr(1200)); status != "ON_TIME" {
		t.Fatalf("expected ON_TIME for duration within range, got %s", status)
	}
}

func TestComputeLapStatusLate(t *testing.T) {
	if status := computeLapStatus(1300, intPtr(1200)); status != "LATE" {
		t.Fatalf("expected LATE when duration exceeds allowed, got %s", status)
	}
}

func TestComputeLapStatusTooFast(t *testing.T) {
	if status := computeLapStatus(400, intPtr(1200)); status != "TOO_FAST" {
		t.Fatalf("expected TOO_FAST when duration is less than half the allowed, got %s", status)
	}
}

func TestIsInsideGeofenceCircleInside(t *testing.T) {
	gf := &entities.Geofence{
		Type:   "CIRCLE",
		Radius: floatPtr(50),
		Points: []entities.GeofencePoint{{Latitude: -17.790538, Longitude: -63.171961}},
	}
	if !isInsideGeofence(gf, -17.790538, -63.171961) {
		t.Fatal("expected the exact center point to be inside the circle")
	}
}

func TestIsInsideGeofenceCircleOutside(t *testing.T) {
	gf := &entities.Geofence{
		Type:   "CIRCLE",
		Radius: floatPtr(10),
		Points: []entities.GeofencePoint{{Latitude: -17.790538, Longitude: -63.171961}},
	}
	// ~1km de distancia, muy fuera de un radio de 10m
	if isInsideGeofence(gf, -17.800538, -63.171961) {
		t.Fatal("expected a point ~1km away to be outside a 10m radius circle")
	}
}

func TestIsInsideGeofenceNoPoints(t *testing.T) {
	gf := &entities.Geofence{Type: "CIRCLE", Radius: floatPtr(50)}
	if isInsideGeofence(gf, -17.790538, -63.171961) {
		t.Fatal("a geofence without points can never contain a point")
	}
}

func TestEvaluateProlongedStopMovingUpdatesMovementAndClearsFine(t *testing.T) {
	fine := &fakeFineService{}
	s := &lapService{fineService: fine}

	past := time.Now().Add(-10 * time.Minute)
	lap := &entities.Lap{ID: uuid.New(), LastMovementAt: &past, LastStopFineAt: &past}

	now := time.Now()
	changed := s.evaluateProlongedStop(context.Background(), lap, &entities.Route{}, movingSpeedThresholdKmh+1, 0, 0, now)

	if !changed {
		t.Fatal("expected movement to be reported as a change")
	}
	if lap.LastStopFineAt != nil {
		t.Fatal("expected LastStopFineAt to be cleared once the vehicle moves again")
	}
	if lap.LastMovementAt == nil || !lap.LastMovementAt.Equal(now) {
		t.Fatalf("expected LastMovementAt to be updated to now, got %v", lap.LastMovementAt)
	}
	if len(fine.generated) != 0 {
		t.Fatal("no fine should be generated while the vehicle is moving")
	}
}

func TestEvaluateProlongedStopGeneratesFineAfterThreshold(t *testing.T) {
	fine := &fakeFineService{}
	s := &lapService{fineService: fine}

	stoppedSince := time.Now().Add(-(defaultMaxStopDurationSeconds + 60) * time.Second)
	lap := &entities.Lap{ID: uuid.New(), VehicleID: uuid.New(), LastMovementAt: &stoppedSince}

	changed := s.evaluateProlongedStop(context.Background(), lap, &entities.Route{}, 0, -17.79, -63.17, time.Now())

	if !changed {
		t.Fatal("expected a change (fine generated + LastStopFineAt set)")
	}
	if len(fine.generated) != 1 || fine.generated[0].FineTypeCode != "PROLONGED_STOP" {
		t.Fatalf("expected exactly one PROLONGED_STOP fine, got %+v", fine.generated)
	}
	if lap.LastStopFineAt == nil {
		t.Fatal("expected LastStopFineAt to be set after fining")
	}
}

func TestEvaluateProlongedStopDoesNotRefireWhileStillStopped(t *testing.T) {
	fine := &fakeFineService{}
	s := &lapService{fineService: fine}

	stoppedSince := time.Now().Add(-(defaultMaxStopDurationSeconds + 600) * time.Second)
	alreadyFinedAt := time.Now().Add(-(defaultMaxStopDurationSeconds + 60) * time.Second)
	lap := &entities.Lap{ID: uuid.New(), LastMovementAt: &stoppedSince, LastStopFineAt: &alreadyFinedAt}

	changed := s.evaluateProlongedStop(context.Background(), lap, &entities.Route{}, 0, 0, 0, time.Now())

	if changed {
		t.Fatal("expected no change: the stop episode was already fined")
	}
	if len(fine.generated) != 0 {
		t.Fatal("expected no new fine while still stopped since the last fine")
	}
}

func TestEvaluateProlongedStopBeforeThresholdDoesNothing(t *testing.T) {
	fine := &fakeFineService{}
	s := &lapService{fineService: fine}

	stoppedSince := time.Now().Add(-30 * time.Second)
	lap := &entities.Lap{ID: uuid.New(), LastMovementAt: &stoppedSince}

	changed := s.evaluateProlongedStop(context.Background(), lap, &entities.Route{}, 0, 0, 0, time.Now())

	if changed || len(fine.generated) != 0 {
		t.Fatal("expected no fine before the stop threshold elapses")
	}
}

func TestEvaluateProlongedStopUsesRouteSpecificThreshold(t *testing.T) {
	fine := &fakeFineService{}
	s := &lapService{fineService: fine}

	stoppedSince := time.Now().Add(-90 * time.Second)
	lap := &entities.Lap{ID: uuid.New(), LastMovementAt: &stoppedSince}
	route := &entities.Route{MaxStopDurationSeconds: intPtr(60)}

	changed := s.evaluateProlongedStop(context.Background(), lap, route, 0, 0, 0, time.Now())

	if !changed || len(fine.generated) != 1 {
		t.Fatal("expected a fine using the shorter route-specific threshold (60s < 90s stopped)")
	}
}

func TestEvaluateSpeedingNoLimitConfigured(t *testing.T) {
	fine := &fakeFineService{}
	s := &lapService{fineService: fine}

	lap := &entities.Lap{ID: uuid.New()}
	changed := s.evaluateSpeeding(context.Background(), lap, &entities.Route{}, 200, 0, 0, time.Now())

	if changed || len(fine.generated) != 0 {
		t.Fatal("expected no speeding fine when the route has no configured speed limit")
	}
}

func TestEvaluateSpeedingUnderLimit(t *testing.T) {
	fine := &fakeFineService{}
	s := &lapService{fineService: fine}

	lap := &entities.Lap{ID: uuid.New()}
	route := &entities.Route{MaxSpeedKmh: intPtr(60)}
	changed := s.evaluateSpeeding(context.Background(), lap, route, 50, 0, 0, time.Now())

	if changed || len(fine.generated) != 0 {
		t.Fatal("expected no fine when speed is under the limit")
	}
}

func TestEvaluateSpeedingOverLimitGeneratesFine(t *testing.T) {
	fine := &fakeFineService{}
	s := &lapService{fineService: fine}

	lap := &entities.Lap{ID: uuid.New(), VehicleID: uuid.New()}
	route := &entities.Route{MaxSpeedKmh: intPtr(60)}
	changed := s.evaluateSpeeding(context.Background(), lap, route, 90, -17.79, -63.17, time.Now())

	if !changed || len(fine.generated) != 1 || fine.generated[0].FineTypeCode != "SPEEDING" {
		t.Fatalf("expected exactly one SPEEDING fine, got %+v", fine.generated)
	}
	if lap.LastSpeedFineAt == nil {
		t.Fatal("expected LastSpeedFineAt to be set after fining")
	}
}

func TestEvaluateSpeedingRespectsCooldown(t *testing.T) {
	fine := &fakeFineService{}
	s := &lapService{fineService: fine}

	recentFine := time.Now().Add(-1 * time.Minute)
	lap := &entities.Lap{ID: uuid.New(), LastSpeedFineAt: &recentFine}
	route := &entities.Route{MaxSpeedKmh: intPtr(60)}
	changed := s.evaluateSpeeding(context.Background(), lap, route, 90, 0, 0, time.Now())

	if changed || len(fine.generated) != 0 {
		t.Fatal("expected no new fine within the cooldown window")
	}
}

func TestEvaluateSpeedingRefiresAfterCooldown(t *testing.T) {
	fine := &fakeFineService{}
	s := &lapService{fineService: fine}

	oldFine := time.Now().Add(-(speedFineCooldown + time.Minute))
	lap := &entities.Lap{ID: uuid.New(), LastSpeedFineAt: &oldFine}
	route := &entities.Route{MaxSpeedKmh: intPtr(60)}
	changed := s.evaluateSpeeding(context.Background(), lap, route, 90, 0, 0, time.Now())

	if !changed || len(fine.generated) != 1 {
		t.Fatal("expected a new fine once the cooldown window has passed")
	}
}
