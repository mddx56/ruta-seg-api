package service

import (
	"context"
	"testing"
	"time"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"github.com/google/uuid"
)

// straightLineGeoJSON es una polilínea recta a lo largo del ecuador (lat=0), usada
// solo para poder razonar fácilmente sobre las distancias/progreso en los tests.
const straightLineGeoJSON = `{"type":"FeatureCollection","features":[{"type":"Feature","geometry":{"type":"LineString","coordinates":[[0,0],[1,0],[2,0]]},"properties":{}}]}`

// fakeLapRepository implementa repository.LapRepository solo lo necesario para probar
// evaluateOvertaking de forma aislada, sin base de datos.
type fakeLapRepository struct {
	openLapsByRoute     []entities.Lap
	updateTrackingCalls []entities.Lap
}

func (f *fakeLapRepository) Create(_ context.Context, _ *entities.Lap) error   { return nil }
func (f *fakeLapRepository) CloseLap(_ context.Context, _ *entities.Lap) error { return nil }
func (f *fakeLapRepository) FindOpenLap(_ context.Context, _ uuid.UUID) (*entities.Lap, error) {
	return nil, nil
}
func (f *fakeLapRepository) FindActiveVehicleRoute(_ context.Context, _ uuid.UUID) (*entities.VehicleRoute, error) {
	return nil, nil
}
func (f *fakeLapRepository) FindRouteByID(_ context.Context, _ uuid.UUID) (*entities.Route, error) {
	return nil, nil
}
func (f *fakeLapRepository) FindGeofenceByID(_ context.Context, _ uuid.UUID) (*entities.Geofence, error) {
	return nil, nil
}
func (f *fakeLapRepository) FindActiveFare(_ context.Context, _ uuid.UUID, _ time.Time) (*entities.RouteFare, error) {
	return nil, nil
}
func (f *fakeLapRepository) CreateCharge(_ context.Context, _ *entities.LapCharge) error { return nil }
func (f *fakeLapRepository) FindAll(_ context.Context) ([]entities.Lap, error)           { return nil, nil }
func (f *fakeLapRepository) FindByID(_ context.Context, _ uuid.UUID) (entities.Lap, error) {
	return entities.Lap{}, nil
}
func (f *fakeLapRepository) FindVehicleOwnerID(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
	return uuid.Nil, nil
}
func (f *fakeLapRepository) UpdateTracking(_ context.Context, lap *entities.Lap) error {
	f.updateTrackingCalls = append(f.updateTrackingCalls, *lap)
	return nil
}
func (f *fakeLapRepository) FindOpenLapsByRoute(_ context.Context, _ uuid.UUID, _ uuid.UUID) ([]entities.Lap, error) {
	return f.openLapsByRoute, nil
}

func TestEvaluateOvertakingDetectsVehicleAhead(t *testing.T) {
	fine := &fakeFineService{}
	polylineIdx := 0

	startedEarlier := time.Now().Add(-10 * time.Minute)
	startedLater := time.Now().Add(-5 * time.Minute)

	otherVehicleID := uuid.New()
	repo := &fakeLapRepository{
		openLapsByRoute: []entities.Lap{
			{
				ID:                 uuid.New(),
				VehicleID:          otherVehicleID,
				StartedAt:          startedLater, // arrancó después...
				LastPolylineIndex:  &polylineIdx,
				LastProgressMeters: floatPtr(50000), // ...pero ya tiene mucho más avance
			},
		},
	}

	s := &lapService{repo: repo, fineService: fine}

	route := &entities.Route{ID: uuid.New()}
	geometry := straightLineGeoJSON
	route.Geometry = &geometry

	lap := &entities.Lap{
		ID:                 uuid.New(),
		VehicleID:          uuid.New(),
		StartedAt:          startedEarlier, // arrancó antes, debería ir adelante
		LastPolylineIndex:  &polylineIdx,   // ya tenía una lectura previa comparable
		LastProgressMeters: floatPtr(1000),
	}

	// Punto cerca del inicio de la polilínea -> poco progreso para "lap".
	changed, err := s.evaluateOvertaking(context.Background(), lap, route, 0.0001, 0.01, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !changed {
		t.Fatal("expected tracking fields to be updated")
	}

	if len(fine.generated) != 1 {
		t.Fatalf("expected exactly one OVERTAKING fine, got %d: %+v", len(fine.generated), fine.generated)
	}
	if fine.generated[0].FineTypeCode != "OVERTAKING" {
		t.Fatalf("expected fine type OVERTAKING, got %s", fine.generated[0].FineTypeCode)
	}
	if fine.generated[0].VehicleID != otherVehicleID {
		t.Fatalf("expected the fine to target the vehicle that overtook (other), got %s", fine.generated[0].VehicleID)
	}

	if len(repo.updateTrackingCalls) != 1 || !repo.updateTrackingCalls[0].OvertakingFined {
		t.Fatal("expected the offending vehicle's lap to be marked OvertakingFined")
	}
}

func TestEvaluateOvertakingSkipsAlreadyFinedVehicle(t *testing.T) {
	fine := &fakeFineService{}
	polylineIdx := 0

	repo := &fakeLapRepository{
		openLapsByRoute: []entities.Lap{
			{
				ID:                 uuid.New(),
				VehicleID:          uuid.New(),
				StartedAt:          time.Now().Add(-5 * time.Minute),
				LastPolylineIndex:  &polylineIdx,
				LastProgressMeters: floatPtr(50000),
				OvertakingFined:    true, // ya se le multó en esta vuelta
			},
		},
	}

	s := &lapService{repo: repo, fineService: fine}

	route := &entities.Route{ID: uuid.New()}
	geometry := straightLineGeoJSON
	route.Geometry = &geometry

	lap := &entities.Lap{
		ID:                 uuid.New(),
		VehicleID:          uuid.New(),
		StartedAt:          time.Now().Add(-10 * time.Minute),
		LastPolylineIndex:  &polylineIdx,
		LastProgressMeters: floatPtr(1000),
	}

	_, err := s.evaluateOvertaking(context.Background(), lap, route, 0.0001, 0.01, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fine.generated) != 0 {
		t.Fatalf("expected no new fine for a vehicle already fined this lap, got %+v", fine.generated)
	}
}

func TestEvaluateOvertakingIgnoresDifferentDirection(t *testing.T) {
	fine := &fakeFineService{}
	polylineIdxSelf := 0
	polylineIdxOther := 1

	repo := &fakeLapRepository{
		openLapsByRoute: []entities.Lap{
			{
				ID:                 uuid.New(),
				VehicleID:          uuid.New(),
				StartedAt:          time.Now().Add(-5 * time.Minute),
				LastPolylineIndex:  &polylineIdxOther, // va en el otro sentido
				LastProgressMeters: floatPtr(50000),
			},
		},
	}

	s := &lapService{repo: repo, fineService: fine}

	route := &entities.Route{ID: uuid.New()}
	geometry := straightLineGeoJSON
	route.Geometry = &geometry

	lap := &entities.Lap{
		ID:                 uuid.New(),
		VehicleID:          uuid.New(),
		StartedAt:          time.Now().Add(-10 * time.Minute),
		LastPolylineIndex:  &polylineIdxSelf,
		LastProgressMeters: floatPtr(1000),
	}

	_, err := s.evaluateOvertaking(context.Background(), lap, route, 0.0001, 0.01, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fine.generated) != 0 {
		t.Fatalf("expected no fine when the other vehicle is on a different polyline/direction, got %+v", fine.generated)
	}
}

func TestEvaluateOvertakingFirstReadingDoesNotCompare(t *testing.T) {
	fine := &fakeFineService{}

	repo := &fakeLapRepository{}
	s := &lapService{repo: repo, fineService: fine}

	route := &entities.Route{ID: uuid.New()}
	geometry := straightLineGeoJSON
	route.Geometry = &geometry

	// Sin LastPolylineIndex previo: es la primera lectura de este vehículo en la vuelta.
	lap := &entities.Lap{ID: uuid.New(), VehicleID: uuid.New(), StartedAt: time.Now()}

	changed, err := s.evaluateOvertaking(context.Background(), lap, route, 0.0001, 0.01, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !changed {
		t.Fatal("expected the first reading to still record tracking fields")
	}
	if lap.LastPolylineIndex == nil {
		t.Fatal("expected LastPolylineIndex to be set after the first reading")
	}
	if len(fine.generated) != 0 {
		t.Fatal("expected no fine on the very first reading (nothing to compare against yet)")
	}
}
