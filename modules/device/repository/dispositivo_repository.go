package repository

import (
	"context"
	"log"
	"time"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	redisProvider "github.com/Caknoooo/go-gin-clean-starter/providers/redis"
	"github.com/samber/do"
	"gorm.io/gorm"
)

type DeviceRepository interface {
	Create(ctx context.Context, device *entities.Device) error
	CreateBatch(ctx context.Context, devices []entities.Device) error
	Update(ctx context.Context, device *entities.Device) error
	ChangeStatus(ctx context.Context, imei string) error
	Delete(ctx context.Context, imei string) error
	FindAll(ctx context.Context) ([]entities.Device, error)
	FindAllSimple(ctx context.Context, available bool) ([]entities.Device, error)
	FindByIMEI(ctx context.Context, imei string) (entities.Device, error)
	FindByIMEIs(ctx context.Context, imeis []string) ([]entities.Device, error)
	FindByIMEIFull(ctx context.Context, imei string) (entities.Device, error)
	FindBySimPhoneNumber(ctx context.Context, phone string) (entities.Device, error)
	FindByCodSim(ctx context.Context, codSim string) (entities.Device, error)
	FindByCodSims(ctx context.Context, codSims []string) ([]entities.Device, error)
	FindAllForExport(ctx context.Context, includeDisabled bool) ([]entities.Device, error)
	GetDevicesWithLastPosition(ctx context.Context, userID string, isAdmin bool) ([]DeviceStatusRow, error)
}

type DeviceStatusRow struct {
	IMEI       string     `gorm:"column:imei"`
	Placa      string     `gorm:"column:placa"`
	Make       string     `gorm:"column:make"`
	Model      string     `gorm:"column:model"`
	Color      string     `gorm:"column:color"`
	Latitude   *float64   `gorm:"column:latitude"`
	Longitude  *float64   `gorm:"column:longitude"`
	Speed      *int       `gorm:"column:speed"`
	Course     *int       `gorm:"column:course"`
	DeviceTime *time.Time `gorm:"column:device_time"`
	ServerTime *time.Time `gorm:"column:server_time"`
	Attributes *string    `gorm:"column:attributes"`
}

type deviceRepository struct {
	db    *gorm.DB
	cache redisProvider.DevicePositionCache
}

func NewDeviceRepository(injector *do.Injector) (DeviceRepository, error) {
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	redisService, _ := do.InvokeNamed[redisProvider.RedisService](injector, "Redis")
	var cache redisProvider.DevicePositionCache
	if redisService != nil {
		cache = redisProvider.NewDevicePositionCache(redisService)
	}
	return &deviceRepository{db: db, cache: cache}, nil
}

func (r *deviceRepository) Create(ctx context.Context, device *entities.Device) error {
	return r.db.WithContext(ctx).Create(device).Error
}

func (r *deviceRepository) CreateBatch(ctx context.Context, devices []entities.Device) error {
	return r.db.WithContext(ctx).Create(&devices).Error
}

func (r *deviceRepository) Update(ctx context.Context, device *entities.Device) error {
	return r.db.WithContext(ctx).Save(device).Error
}

func (r *deviceRepository) Delete(ctx context.Context, imei string) error {
	return r.db.WithContext(ctx).Where("imei = ?", imei).Delete(&entities.Device{}).Error
}

func (r *deviceRepository) ChangeStatus(ctx context.Context, imei string) error {
	return r.db.WithContext(ctx).Exec("UPDATE devices SET status = NOT status WHERE imei = ?", imei).Error
}

func (r *deviceRepository) FindAll(ctx context.Context) ([]entities.Device, error) {
	var devices []entities.Device
	err := r.db.WithContext(ctx).Where("status = ?", true).Find(&devices).Error
	return devices, err
}

func (r *deviceRepository) FindAllSimple(ctx context.Context, available bool) ([]entities.Device, error) {
	var devices []entities.Device
	query := r.db.WithContext(ctx).Select("imei", "model")

	if available {
		query = query.Where("imei NOT IN (?)",
			r.db.Table("device_installations").Select("imei").Where("removed_at IS NULL"))
	}

	err := query.Find(&devices).Error
	return devices, err
}

func (r *deviceRepository) FindByIMEI(ctx context.Context, imei string) (entities.Device, error) {
	var device entities.Device
	err := r.db.WithContext(ctx).
		Preload("Installations", "removed_at IS NULL").
		Preload("Installations.Vehicle").
		Preload("Installations.Vehicle.User").
		Preload("Installations.Vehicle.Model").
		Preload("Installations.Vehicle.Model.Make").
		Preload("Installations.Vehicle.Model.VehicleType").
		Preload("GroupDevices").
		Preload("GroupDevices.Group").
		Preload("GroupDevices.Group.User").
		Where("imei = ?", imei).
		First(&device).Error
	return device, err
}

func (r *deviceRepository) FindByIMEIs(ctx context.Context, imeis []string) ([]entities.Device, error) {
	var devices []entities.Device
	if len(imeis) == 0 {
		return devices, nil
	}
	err := r.db.WithContext(ctx).Where("imei IN ?", imeis).Find(&devices).Error
	return devices, err
}

func (r *deviceRepository) FindByIMEIFull(ctx context.Context, imei string) (entities.Device, error) {
	var device entities.Device
	err := r.db.WithContext(ctx).
		Preload("Installations").
		Preload("Installations.Vehicle").
		Preload("Installations.Vehicle.User").
		Preload("Installations.Vehicle.Model").
		Preload("Installations.Vehicle.Model.Make").
		Preload("Installations.Vehicle.Model.VehicleType").
		Preload("GroupDevices").
		Preload("GroupDevices.Group").
		Preload("GroupDevices.Group.User").
		Where("imei = ?", imei).
		First(&device).Error
	return device, err
}

func (r *deviceRepository) FindBySimPhoneNumber(ctx context.Context, phone string) (entities.Device, error) {
	var device entities.Device
	err := r.db.WithContext(ctx).Where("sim_phone_number = ?", phone).First(&device).Error
	return device, err
}

func (r *deviceRepository) FindByCodSim(ctx context.Context, codSim string) (entities.Device, error) {
	var device entities.Device
	err := r.db.WithContext(ctx).Where("sim_icc_id = ?", codSim).First(&device).Error
	return device, err
}

func (r *deviceRepository) FindByCodSims(ctx context.Context, codSims []string) ([]entities.Device, error) {
	var devices []entities.Device
	if len(codSims) == 0 {
		return devices, nil
	}
	err := r.db.WithContext(ctx).Where("sim_icc_id IN ?", codSims).Find(&devices).Error
	return devices, err
}

func (r *deviceRepository) FindAllForExport(ctx context.Context, includeDisabled bool) ([]entities.Device, error) {
	var devices []entities.Device
	query := r.db.WithContext(ctx).Select("imei", "sim_icc_id", "sim_phone_number", "sim_provider", "status")
	if !includeDisabled {
		query = query.Where("status = ?", true)
	}
	err := query.Order("created_at DESC").Find(&devices).Error
	return devices, err
}

func (r *deviceRepository) GetDevicesWithLastPosition(ctx context.Context, userID string, isAdmin bool) ([]DeviceStatusRow, error) {
	// Paso 1: metadatos de instalación/vehículo desde Postgres (sin posición)
	type installMeta struct {
		IMEI  string `gorm:"column:imei"`
		Placa string `gorm:"column:placa"`
		Make  string `gorm:"column:make"`
		Model string `gorm:"column:model"`
		Color string `gorm:"column:color"`
	}
	metaQuery := `
		SELECT
			di.imei,
			COALESCE(v.placa, '')       AS placa,
			COALESCE(vm.make_name, '')  AS make,
			COALESCE(vmo.model_name,'') AS model,
			COALESCE(v.color, '')       AS color
		FROM device_installations di
		JOIN vehicles v ON v.id = di.vehicle_id
		LEFT JOIN models vmo ON vmo.id = v.model_id
		LEFT JOIN makes  vm  ON vm.id  = vmo.make_id
		WHERE di.removed_at IS NULL AND di.status = true
	`
	args := []interface{}{}
	if !isAdmin {
		metaQuery += " AND v.user_id = ?"
		args = append(args, userID)
	}
	metaQuery += " ORDER BY v.placa ASC"

	var metas []installMeta
	if err := r.db.WithContext(ctx).Raw(metaQuery, args...).Scan(&metas).Error; err != nil {
		return nil, err
	}
	if len(metas) == 0 {
		return nil, nil
	}

	// Paso 2: posiciones desde Redis con fallback a device_last_positions
	imeis := make([]string, len(metas))
	for i, m := range metas {
		imeis[i] = m.IMEI
	}
	posMap := r.resolvePositions(ctx, imeis)

	// Paso 3: unir en Go (equivale al LEFT JOIN original)
	rows := make([]DeviceStatusRow, 0, len(metas))
	for _, m := range metas {
		row := DeviceStatusRow{
			IMEI:  m.IMEI,
			Placa: m.Placa,
			Make:  m.Make,
			Model: m.Model,
			Color: m.Color,
		}
		if pos, ok := posMap[m.IMEI]; ok {
			row.Latitude   = &pos.Latitude
			row.Longitude  = &pos.Longitude
			row.Speed      = &pos.Speed
			row.Course     = &pos.Course
			row.DeviceTime = &pos.DeviceTime
			row.ServerTime = &pos.ServerTime
			row.Attributes = pos.Attributes
		}
		rows = append(rows, row)
	}
	return rows, nil
}

// resolvePositions busca posiciones en Redis y completa los misses con device_last_positions.
func (r *deviceRepository) resolvePositions(ctx context.Context, imeis []string) map[string]redisProvider.CachedPosition {
	result := make(map[string]redisProvider.CachedPosition, len(imeis))

	if r.cache != nil {
		if cached, err := r.cache.MGet(ctx, imeis); err == nil {
			for k, v := range cached {
				result[k] = v
			}
		} else {
			log.Printf("[pos-cache] MGet error: %v", err)
		}
	}

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

