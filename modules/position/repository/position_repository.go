package repository

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	redisProvider "github.com/Caknoooo/go-gin-clean-starter/providers/redis"
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
	db    *gorm.DB
	cache redisProvider.DevicePositionCache
}

func NewPositionRepository(db *gorm.DB, cache redisProvider.DevicePositionCache) PositionRepository {
	return &positionRepository{db: db, cache: cache}
}

// --- escritura ---

func (r *positionRepository) Create(ctx context.Context, position *entities.Position) error {
	if err := r.db.WithContext(ctx).Create(position).Error; err != nil {
		return err
	}
	r.updatePositionCache(ctx, position)
	return nil
}

// updatePositionCache escribe en Redis y, como escritura dual de seguridad, en device_last_positions.
func (r *positionRepository) updatePositionCache(ctx context.Context, p *entities.Position) {
	// Redis (fuente principal de lecturas)
	if r.cache != nil {
		pos := redisProvider.CachedPosition{
			IMEI:       p.Imei,
			Latitude:   p.Latitude,
			Longitude:  p.Longitude,
			Speed:      p.Speed,
			Course:     p.Course,
			DeviceTime: p.DeviceTime,
			ServerTime: p.ServerTime,
			Attributes: p.Attributes,
		}
		if err := r.cache.Set(ctx, pos); err != nil {
			log.Printf("[pos-cache] error al escribir en Redis para %s: %v", p.Imei, err)
		}
	}

	// Postgres dual-write: mantiene device_last_positions como respaldo hasta que Redis sea estable
	sql := `
		INSERT INTO device_last_positions (imei, latitude, longitude, speed, course, device_time, server_time, attributes, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW())
		ON CONFLICT (imei) DO UPDATE SET
			latitude    = EXCLUDED.latitude,
			longitude   = EXCLUDED.longitude,
			speed       = EXCLUDED.speed,
			course      = EXCLUDED.course,
			device_time = EXCLUDED.device_time,
			server_time = EXCLUDED.server_time,
			attributes  = EXCLUDED.attributes,
			updated_at  = NOW()
	`
	if err := r.db.WithContext(ctx).Exec(sql,
		p.Imei, p.Latitude, p.Longitude, p.Speed, p.Course,
		p.DeviceTime, p.ServerTime, p.Attributes,
	).Error; err != nil {
		log.Printf("[pos-cache] error al actualizar device_last_positions para %s: %v", p.Imei, err)
	}
}

// --- lecturas ---

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
	// 1. Redis
	if r.cache != nil {
		if pos, ok, _ := r.cache.Get(ctx, imei); ok {
			return cachedToPosition(pos), nil
		}
	}

	// 2. Fallback: device_last_positions
	var cache entities.DeviceLastPosition
	if err := r.db.WithContext(ctx).Where("imei = ?", imei).First(&cache).Error; err == nil {
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

	// 3. Fallback final: tabla positions
	var position entities.Position
	err := r.db.WithContext(ctx).
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
	err := r.db.WithContext(ctx).
		Select("id, latitude, longitude, speed, device_time, attributes").
		Where("device_id = ? AND device_time >= ? AND device_time < ?", imei, start, end).
		Order("device_time ASC").
		Find(&positions).Error
	return positions, err
}

func (r *positionRepository) FindLastPositions(ctx context.Context) ([]entities.Position, error) {
	// 1. Redis
	if r.cache != nil {
		if cached, err := r.cache.GetAll(ctx); err == nil && len(cached) > 0 {
			positions := make([]entities.Position, 0, len(cached))
			for _, c := range cached {
				positions = append(positions, cachedToPosition(c))
			}
			return positions, nil
		}
	}

	// 2. Fallback: device_last_positions
	var caches []entities.DeviceLastPosition
	if err := r.db.WithContext(ctx).Find(&caches).Error; err != nil {
		return nil, err
	}
	positions := make([]entities.Position, 0, len(caches))
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
// Métodos orientados a Vehículo
// ---------------------------------------------------------------------------

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

// vehicleInstallRow es el resultado del JOIN vehículo+instalación sin datos de posición.
type vehicleInstallRow struct {
	VehicleID string `gorm:"column:vehicle_id"`
	Placa     string `gorm:"column:placa"`
	IMEI      string `gorm:"column:imei"`
}

// FindLastPositionsByVehicles obtiene datos de instalación desde Postgres
// y posiciones desde Redis, con fallback a device_last_positions.
func (r *positionRepository) FindLastPositionsByVehicles(ctx context.Context) ([]VehicleLastPosition, error) {
	var installs []vehicleInstallRow
	if err := r.db.WithContext(ctx).Raw(`
		SELECT di.vehicle_id::text, v.placa, di.imei
		FROM device_installations di
		JOIN vehicles v ON v.id = di.vehicle_id
		WHERE di.removed_at IS NULL
		ORDER BY v.placa ASC
	`).Scan(&installs).Error; err != nil {
		return nil, err
	}
	if len(installs) == 0 {
		return nil, nil
	}

	imeis := make([]string, len(installs))
	for i, inst := range installs {
		imeis[i] = inst.IMEI
	}

	posMap := r.resolvePositions(ctx, imeis)

	results := make([]VehicleLastPosition, 0, len(installs))
	for _, inst := range installs {
		pos, ok := posMap[inst.IMEI]
		if !ok {
			continue
		}
		results = append(results, VehicleLastPosition{
			VehicleID:  inst.VehicleID,
			Placa:      inst.Placa,
			Imei:       inst.IMEI,
			Latitude:   pos.Latitude,
			Longitude:  pos.Longitude,
			Speed:      pos.Speed,
			Course:     pos.Course,
			DeviceTime: pos.DeviceTime,
			ServerTime: pos.ServerTime,
			Attributes: pos.Attributes,
		})
	}
	return results, nil
}

// FindLastPositionByVehicle obtiene la última posición de UN vehículo específico.
func (r *positionRepository) FindLastPositionByVehicle(ctx context.Context, vehicleID uuid.UUID) (VehicleLastPosition, error) {
	var inst vehicleInstallRow
	if err := r.db.WithContext(ctx).Raw(fmt.Sprintf(`
		SELECT di.vehicle_id::text, v.placa, di.imei
		FROM device_installations di
		JOIN vehicles v ON v.id = di.vehicle_id
		WHERE di.vehicle_id = '%s'
		  AND di.removed_at IS NULL
		LIMIT 1
	`, vehicleID.String())).Scan(&inst).Error; err != nil {
		return VehicleLastPosition{}, err
	}
	if inst.IMEI == "" {
		return VehicleLastPosition{}, gorm.ErrRecordNotFound
	}

	posMap := r.resolvePositions(ctx, []string{inst.IMEI})
	pos, ok := posMap[inst.IMEI]
	if !ok {
		return VehicleLastPosition{}, gorm.ErrRecordNotFound
	}
	return VehicleLastPosition{
		VehicleID:  inst.VehicleID,
		Placa:      inst.Placa,
		Imei:       inst.IMEI,
		Latitude:   pos.Latitude,
		Longitude:  pos.Longitude,
		Speed:      pos.Speed,
		Course:     pos.Course,
		DeviceTime: pos.DeviceTime,
		ServerTime: pos.ServerTime,
		Attributes: pos.Attributes,
	}, nil
}

// resolvePositions busca posiciones en Redis y completa los misses con device_last_positions.
func (r *positionRepository) resolvePositions(ctx context.Context, imeis []string) map[string]redisProvider.CachedPosition {
	result := make(map[string]redisProvider.CachedPosition, len(imeis))

	// Intento Redis
	if r.cache != nil {
		if cached, err := r.cache.MGet(ctx, imeis); err == nil {
			for k, v := range cached {
				result[k] = v
			}
		}
	}

	// Fallback Postgres para los IMEIs que no estaban en Redis
	var missing []string
	for _, imei := range imeis {
		if _, ok := result[imei]; !ok {
			missing = append(missing, imei)
		}
	}
	if len(missing) > 0 {
		var dbRows []entities.DeviceLastPosition
		if err := r.db.WithContext(ctx).Where("imei IN ?", missing).Find(&dbRows).Error; err == nil {
			for _, row := range dbRows {
				result[row.IMEI] = redisProvider.CachedPosition{
					IMEI:       row.IMEI,
					Latitude:   row.Latitude,
					Longitude:  row.Longitude,
					Speed:      row.Speed,
					Course:     row.Course,
					DeviceTime: row.DeviceTime,
					ServerTime: row.ServerTime,
					Attributes: row.Attributes,
				}
			}
		}
	}
	return result
}

// cachedToPosition convierte un CachedPosition en entities.Position.
func cachedToPosition(c redisProvider.CachedPosition) entities.Position {
	return entities.Position{
		Imei:       c.IMEI,
		Latitude:   c.Latitude,
		Longitude:  c.Longitude,
		Speed:      c.Speed,
		Course:     c.Course,
		DeviceTime: c.DeviceTime,
		ServerTime: c.ServerTime,
		Attributes: c.Attributes,
	}
}
