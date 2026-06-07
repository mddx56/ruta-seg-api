package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// InstallationSlot representa el tramo de tiempo en que un IMEI
// estuvo instalado en un vehículo, ya con removed_at resuelto (COALESCE → now).
type InstallationSlot struct {
	Imei        string
	InstalledAt time.Time
	RemovedAt   time.Time
}

type PositionRepository interface {
	Create(ctx context.Context, position *entities.Position) error
	FindByID(ctx context.Context, id uint64) (entities.Position, error)
	FindByIMEI(ctx context.Context, imei string) ([]entities.Position, error)
	FindLastByIMEI(ctx context.Context, imei string) (entities.Position, error)
	FindByIMEIAndDate(ctx context.Context, imei string, date time.Time) ([]entities.Position, error)
	Delete(ctx context.Context, id uint64) error

	// History por IMEI (legacy)
	FindForHistory(ctx context.Context, imei string, start, end time.Time) ([]entities.Position, error)
	FindLastPositions(ctx context.Context) ([]entities.Position, error)

	// History por Vehículo (correcto)
	FindSlotsByVehicleAndRange(ctx context.Context, vehicleID uuid.UUID, start, end time.Time) ([]InstallationSlot, error)
	FindForHistoryBySlots(ctx context.Context, slots []InstallationSlot, globalStart, globalEnd time.Time) ([]entities.Position, error)
	FindLastPositionsByVehicles(ctx context.Context) ([]VehicleLastPosition, error)
	FindLastPositionByVehicle(ctx context.Context, vehicleID uuid.UUID) (VehicleLastPosition, error)
}

// VehicleLastPosition es el resultado aplanado del LATERAL JOIN.
type VehicleLastPosition struct {
	VehicleID  string    `gorm:"column:vehicle_id"`
	Placa      string    `gorm:"column:placa"`
	Imei       string    `gorm:"column:imei"`
	Latitude   float64   `gorm:"column:latitude"`
	Longitude  float64   `gorm:"column:longitude"`
	Speed      int       `gorm:"column:speed"`
	Course     int       `gorm:"column:course"`
	DeviceTime time.Time `gorm:"column:device_time"`
	ServerTime time.Time `gorm:"column:server_time"`
	Attributes *string   `gorm:"column:attributes"`
}

type positionRepository struct {
	db *gorm.DB
}

func NewPositionRepository(db *gorm.DB) PositionRepository {
	return &positionRepository{db: db}
}

func (r *positionRepository) Create(ctx context.Context, position *entities.Position) error {
	return r.db.WithContext(ctx).Create(position).Error
}

func (r *positionRepository) FindByID(ctx context.Context, id uint64) (entities.Position, error) {
	var position entities.Position
	err := r.db.WithContext(ctx).
		Preload("Device").
		First(&position, id).Error
	return position, err
}

func (r *positionRepository) FindByIMEI(ctx context.Context, imei string) ([]entities.Position, error) {
	var positions []entities.Position
	err := r.db.WithContext(ctx).
		Where("device_id = ?", imei).
		Order("server_time DESC").
		Find(&positions).Error
	return positions, err
}

func (r *positionRepository) FindLastByIMEI(ctx context.Context, imei string) (entities.Position, error) {
	var cache entities.DeviceLastPosition
	err := r.db.WithContext(ctx).
		Where("imei = ?", imei).
		First(&cache).Error

	if err == nil {
		return entities.Position{
			Imei:       cache.IMEI,
			Latitude:   cache.Latitude,
			Longitude:  cache.Longitude,
			Speed:      cache.Speed,
			Course:     cache.Course,
			DeviceTime: cache.DeviceTime,
			ServerTime: cache.ServerTime,
			Attributes: cache.Attributes,
		}, nil
	}

	// Fallback to positions table en caso de no encontrar en caché
	var position entities.Position
	err = r.db.WithContext(ctx).
		Where("device_id = ?", imei).
		Order("server_time DESC").
		First(&position).Error
	return position, err
}

func (r *positionRepository) FindByIMEIAndDate(ctx context.Context, imei string, date time.Time) ([]entities.Position, error) {
	var positions []entities.Position

	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	err := r.db.WithContext(ctx).
		Where("device_id = ?", imei).
		Where("server_time >= ?", startOfDay).
		Where("server_time < ?", endOfDay).
		Order("server_time DESC").
		Find(&positions).Error
	return positions, err
}

func (r *positionRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entities.Position{}, id).Error
}

func (r *positionRepository) FindForHistory(ctx context.Context, imei string, start, end time.Time) ([]entities.Position, error) {
	var positions []entities.Position
	// Select specific fields including ID for the raw query later.
	// Note: We don't fetch Geom here to save memory, it's used in CalculateRoute.
	err := r.db.WithContext(ctx).
		Select("id, latitude, longitude, speed, device_time, attributes").
		Where("device_id = ? AND device_time >= ? AND device_time < ?", imei, start, end).
		Order("device_time ASC").
		Find(&positions).Error
	return positions, err
}



func (r *positionRepository) FindLastPositions(ctx context.Context) ([]entities.Position, error) {
	var caches []entities.DeviceLastPosition
	err := r.db.WithContext(ctx).Find(&caches).Error
	if err != nil {
		return nil, err
	}
	
	var positions []entities.Position
	for _, cache := range caches {
		positions = append(positions, entities.Position{
			Imei:       cache.IMEI,
			Latitude:   cache.Latitude,
			Longitude:  cache.Longitude,
			Speed:      cache.Speed,
			Course:     cache.Course,
			DeviceTime: cache.DeviceTime,
			ServerTime: cache.ServerTime,
			Attributes: cache.Attributes,
		})
	}
	return positions, nil
}

// ---------------------------------------------------------------------------
// Métodos orientados a Vehículo (correctos)
// ---------------------------------------------------------------------------

// FindSlotsByVehicleAndRange devuelve los tramos (IMEI + ventana de tiempo)
// en que el vehículo tuvo un dispositivo instalado, que se solapan con [start, end).
func (r *positionRepository) FindSlotsByVehicleAndRange(
	ctx context.Context, vehicleID uuid.UUID, start, end time.Time,
) ([]InstallationSlot, error) {
	type row struct {
		Imei        string
		InstalledAt time.Time
		RemovedAt   *time.Time
	}
	var rows []row
	err := r.db.WithContext(ctx).
		Table("device_installations").
		Select("imei, installed_at, removed_at").
		Where("vehicle_id = ?", vehicleID).
		Where("installed_at < ?", end).
		Where("removed_at IS NULL OR removed_at > ?", start).
		Order("installed_at ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	now := time.Now()
	slots := make([]InstallationSlot, len(rows))
	for i, row := range rows {
		removedAt := now
		if row.RemovedAt != nil {
			removedAt = *row.RemovedAt
		}
		slots[i] = InstallationSlot{
			Imei:        row.Imei,
			InstalledAt: row.InstalledAt,
			RemovedAt:   removedAt,
		}
	}
	return slots, nil
}

// FindForHistoryBySlots busca posiciones de múltiples slots en una sola query
// usando OR dinámico — PostgreSQL usa Bitmap Index Scan por cada condición.
func (r *positionRepository) FindForHistoryBySlots(
	ctx context.Context, slots []InstallationSlot, globalStart, globalEnd time.Time,
) ([]entities.Position, error) {
	if len(slots) == 0 {
		return nil, nil
	}
	conditions := make([]string, 0, len(slots))
	args := make([]interface{}, 0, len(slots)*3)
	for _, slot := range slots {
		slotStart := globalStart
		if slot.InstalledAt.After(globalStart) {
			slotStart = slot.InstalledAt
		}
		slotEnd := globalEnd
		if slot.RemovedAt.Before(globalEnd) {
			slotEnd = slot.RemovedAt
		}
		conditions = append(conditions, "(device_id = ? AND device_time >= ? AND device_time < ?)")
		args = append(args, slot.Imei, slotStart, slotEnd)
	}
	var positions []entities.Position
	err := r.db.WithContext(ctx).
		Select("id, device_id, latitude, longitude, speed, course, device_time, attributes").
		Where(strings.Join(conditions, " OR "), args...).
		Order("device_time ASC").
		Find(&positions).Error
	return positions, err
}

// FindLastPositionsByVehicles usa un simple JOIN a device_last_positions caché
func (r *positionRepository) FindLastPositionsByVehicles(ctx context.Context) ([]VehicleLastPosition, error) {
	var results []VehicleLastPosition
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			di.vehicle_id::text,
			v.placa,
			di.imei,
			dlp.latitude,
			dlp.longitude,
			dlp.speed,
			dlp.course,
			dlp.device_time,
			dlp.server_time,
			dlp.attributes
		FROM device_installations di
		JOIN vehicles v ON v.id = di.vehicle_id
		JOIN device_last_positions dlp ON dlp.imei = di.imei
		WHERE di.removed_at IS NULL
		ORDER BY v.placa ASC
	`).Scan(&results).Error
	return results, err
}

// FindLastPositionByVehicle obtiene la última posición de UN vehículo específico.
func (r *positionRepository) FindLastPositionByVehicle(ctx context.Context, vehicleID uuid.UUID) (VehicleLastPosition, error) {
	var result VehicleLastPosition
	err := r.db.WithContext(ctx).Raw(fmt.Sprintf(`
		SELECT
			di.vehicle_id::text,
			v.placa,
			di.imei,
			dlp.latitude,
			dlp.longitude,
			dlp.speed,
			dlp.course,
			dlp.device_time,
			dlp.server_time,
			dlp.attributes
		FROM device_installations di
		JOIN vehicles v ON v.id = di.vehicle_id
		JOIN device_last_positions dlp ON dlp.imei = di.imei
		WHERE di.vehicle_id = '%s'
		  AND di.removed_at IS NULL
		LIMIT 1
	`, vehicleID.String())).Scan(&result).Error
	return result, err
}
