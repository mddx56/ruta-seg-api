package redis

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisService define la interfaz para interactuar con Redis aplicando el Principio de Segregación de Interfaces (ISP).
type RedisService interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	// GetInt devuelve el entero almacenado en la clave. exists=false si la clave no existe.
	GetInt(ctx context.Context, key string) (count int64, exists bool, err error)
	// IncrWithTTL incrementa el contador y renueva el TTL en un pipeline atómico (ventana deslizante).
	IncrWithTTL(ctx context.Context, key string, expiration time.Duration) (int64, error)
	Client() *redis.Client // Expone el cliente en crudo para operaciones complejas como Pub/Sub
}

type redisService struct {
	client *redis.Client
}

// NewRedisService inicializa y retorna el servicio de Redis.
func NewRedisService(host string, port string, password string) (RedisService, error) {
	addr := fmt.Sprintf("%s:%s", host, port)

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("⚠️ No se pudo conectar a Redis en %s: %v", addr, err)
		return nil, err
	}

	log.Printf("🚀 Conectado a Redis exitosamente en %s", addr)

	return &redisService{
		client: client,
	}, nil
}

func (r *redisService) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *redisService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *redisService) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *redisService) GetInt(ctx context.Context, key string) (int64, bool, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, false, nil
		}
		return 0, false, err
	}
	count, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, false, nil
	}
	return count, true, nil
}

func (r *redisService) IncrWithTTL(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	pipe := r.client.Pipeline()
	incrCmd := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, expiration)
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, err
	}
	return incrCmd.Val(), nil
}

func (r *redisService) Client() *redis.Client {
	return r.client
}
