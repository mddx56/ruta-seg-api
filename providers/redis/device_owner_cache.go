package redis

import (
	"context"
	"encoding/json"
	"time"
)

const (
	deviceOwnerKeyPrefix = "device:owner:"
	deviceOwnerTTL       = 15 * time.Minute
)

// DeviceOwner es el dueño (vehículo + usuario) resuelto para un IMEI.
type DeviceOwner struct {
	VehicleID string `json:"vehicle_id"`
	UserID    string `json:"user_id"`
}

// DeviceOwnerCache cachea a qué vehículo/usuario pertenece un IMEI.
type DeviceOwnerCache interface {
	Get(ctx context.Context, imei string) (DeviceOwner, bool)
	Set(ctx context.Context, imei string, owner DeviceOwner) error
	Invalidate(ctx context.Context, imei string) error
}

type deviceOwnerCache struct {
	redis RedisService
}

func NewDeviceOwnerCache(redis RedisService) DeviceOwnerCache {
	return &deviceOwnerCache{redis: redis}
}

func deviceOwnerKey(imei string) string { return deviceOwnerKeyPrefix + imei }

func (c *deviceOwnerCache) Get(ctx context.Context, imei string) (DeviceOwner, bool) {
	val, err := c.redis.Get(ctx, deviceOwnerKey(imei))
	if err != nil {
		return DeviceOwner{}, false
	}
	var owner DeviceOwner
	if err := json.Unmarshal([]byte(val), &owner); err != nil {
		return DeviceOwner{}, false
	}
	return owner, true
}

func (c *deviceOwnerCache) Set(ctx context.Context, imei string, owner DeviceOwner) error {
	data, err := json.Marshal(owner)
	if err != nil {
		return err
	}
	return c.redis.Set(ctx, deviceOwnerKey(imei), string(data), deviceOwnerTTL)
}

func (c *deviceOwnerCache) Invalidate(ctx context.Context, imei string) error {
	return c.redis.Delete(ctx, deviceOwnerKey(imei))
}
