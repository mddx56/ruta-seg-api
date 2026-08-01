package service

import (
	"context"

	redisProvider "github.com/Caknoooo/go-gin-clean-starter/providers/redis"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DeviceOwner identifica al vehículo y usuario dueños de un IMEI.
type DeviceOwner struct {
	VehicleID uuid.UUID
	UserID    string
}

// DeviceOwnerResolver resuelve IMEI -> vehículo/usuario, con caché de por medio.
type DeviceOwnerResolver interface {
	Resolve(ctx context.Context, imei string) (DeviceOwner, bool)
}

type deviceOwnerResolver struct {
	db    *gorm.DB
	cache redisProvider.DeviceOwnerCache // puede ser nil si Redis no está disponible
}

func NewDeviceOwnerResolver(db *gorm.DB, cache redisProvider.DeviceOwnerCache) DeviceOwnerResolver {
	return &deviceOwnerResolver{db: db, cache: cache}
}

type deviceOwnerRow struct {
	VehicleID uuid.UUID
	UserID    string
}

func (r *deviceOwnerResolver) Resolve(ctx context.Context, imei string) (DeviceOwner, bool) {
	if r.cache != nil {
		if cached, ok := r.cache.Get(ctx, imei); ok {
			if vehicleID, err := uuid.Parse(cached.VehicleID); err == nil {
				return DeviceOwner{VehicleID: vehicleID, UserID: cached.UserID}, true
			}
		}
	}

	var row deviceOwnerRow
	err := r.db.WithContext(ctx).
		Table("device_installations").
		Select("vehicles.id AS vehicle_id, vehicles.user_id AS user_id").
		Joins("JOIN vehicles ON vehicles.id = device_installations.vehicle_id").
		Where("device_installations.imei = ? AND device_installations.removed_at IS NULL AND device_installations.status = ?", imei, true).
		Limit(1).
		Take(&row).Error
	if err != nil {
		return DeviceOwner{}, false
	}

	owner := DeviceOwner{VehicleID: row.VehicleID, UserID: row.UserID}
	if r.cache != nil {
		// Best-effort: si falla el cacheo, el próximo Resolve vuelve a golpear la DB.
		_ = r.cache.Set(ctx, imei, redisProvider.DeviceOwner{VehicleID: owner.VehicleID.String(), UserID: owner.UserID})
	}
	return owner, true
}
