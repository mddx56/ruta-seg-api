package redis

import (
	"context"
	"encoding/json"
	"log"
	"time"
)

const (
	devicePosKeyPrefix = "device:pos:"
	devicePosIndex     = "device:pos:index"
)

// CachedPosition es la estructura almacenada en Redis para la última posición de un dispositivo.
type CachedPosition struct {
	IMEI       string    `json:"imei"`
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	Speed      int       `json:"speed"`
	Course     int       `json:"course"`
	DeviceTime time.Time `json:"device_time"`
	ServerTime time.Time `json:"server_time"`
	Attributes *string   `json:"attributes,omitempty"`
}

// DevicePositionCache gestiona el caché de última posición por dispositivo en Redis.
type DevicePositionCache interface {
	Set(ctx context.Context, pos CachedPosition) error
	// SetNX escribe solo si la clave no existe — usar en warm-up para no pisar datos recientes.
	SetNX(ctx context.Context, pos CachedPosition) error
	Get(ctx context.Context, imei string) (CachedPosition, bool, error)
	MGet(ctx context.Context, imeis []string) (map[string]CachedPosition, error)
	GetAll(ctx context.Context) ([]CachedPosition, error)
}

type devicePositionCache struct {
	redis RedisService
}

func NewDevicePositionCache(redis RedisService) DevicePositionCache {
	return &devicePositionCache{redis: redis}
}

func devicePosKey(imei string) string { return devicePosKeyPrefix + imei }

func (c *devicePositionCache) Set(ctx context.Context, pos CachedPosition) error {
	data, err := json.Marshal(pos)
	if err != nil {
		return err
	}
	if err := c.redis.Set(ctx, devicePosKey(pos.IMEI), string(data), 0); err != nil {
		return err
	}
	// Mantener índice de IMEIs para GetAll
	if err := c.redis.Client().SAdd(ctx, devicePosIndex, pos.IMEI).Err(); err != nil {
		log.Printf("[pos-cache] error al agregar IMEI al índice: %v", err)
	}
	return nil
}

// SetNX escribe la posición solo si la clave no existe todavía (usado en warm-up
// para no sobreescribir posiciones recientes con datos históricos).
func (c *devicePositionCache) SetNX(ctx context.Context, pos CachedPosition) error {
	data, err := json.Marshal(pos)
	if err != nil {
		return err
	}
	ok, err := c.redis.Client().SetNX(ctx, devicePosKey(pos.IMEI), string(data), 0).Result()
	if err != nil {
		return err
	}
	if ok {
		if err := c.redis.Client().SAdd(ctx, devicePosIndex, pos.IMEI).Err(); err != nil {
			log.Printf("[pos-cache] error al agregar IMEI al índice (SetNX): %v", err)
		}
	}
	return nil
}

func (c *devicePositionCache) Get(ctx context.Context, imei string) (CachedPosition, bool, error) {
	val, err := c.redis.Get(ctx, devicePosKey(imei))
	if err != nil {
		return CachedPosition{}, false, nil
	}
	var pos CachedPosition
	if err := json.Unmarshal([]byte(val), &pos); err != nil {
		log.Printf("[pos-cache] unmarshal error para %s: %v", imei, err)
		return CachedPosition{}, false, nil
	}
	return pos, true, nil
}

func (c *devicePositionCache) MGet(ctx context.Context, imeis []string) (map[string]CachedPosition, error) {
	if len(imeis) == 0 {
		return map[string]CachedPosition{}, nil
	}
	keys := make([]string, len(imeis))
	for i, imei := range imeis {
		keys[i] = devicePosKey(imei)
	}
	vals, err := c.redis.Client().MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}
	result := make(map[string]CachedPosition, len(imeis))
	for i, val := range vals {
		if val == nil {
			continue
		}
		str, ok := val.(string)
		if !ok {
			continue
		}
		var pos CachedPosition
		if err := json.Unmarshal([]byte(str), &pos); err != nil {
			log.Printf("[pos-cache] unmarshal error para %s: %v", imeis[i], err)
			continue
		}
		result[imeis[i]] = pos
	}
	return result, nil
}

// GetAll devuelve todas las posiciones en caché usando el índice de IMEIs.
func (c *devicePositionCache) GetAll(ctx context.Context) ([]CachedPosition, error) {
	imeis, err := c.redis.Client().SMembers(ctx, devicePosIndex).Result()
	if err != nil || len(imeis) == 0 {
		return nil, err
	}
	positions, err := c.MGet(ctx, imeis)
	if err != nil {
		return nil, err
	}
	result := make([]CachedPosition, 0, len(positions))
	for _, pos := range positions {
		result = append(result, pos)
	}
	return result, nil
}
