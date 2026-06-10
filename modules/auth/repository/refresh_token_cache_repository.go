package repository

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	redisProvider "github.com/Caknoooo/go-gin-clean-starter/providers/redis"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	rtKeyPrefix     = "auth:rt:"
	rtUserKeyPrefix = "auth:rt:user:"
)

type cachedToken struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Token     string     `json:"token"`
	ExpiresAt time.Time  `json:"expires_at"`
	User      cachedUser `json:"user"`
}

type cachedUser struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Username   *string `json:"username"`
	Email      string  `json:"email"`
	TelpNumber string  `json:"telp_number"`
	Role       string  `json:"role"`
	ImageUrl   string  `json:"image_url"`
	IsVerified bool    `json:"is_verified"`
	IsBlocked  bool    `json:"is_blocked"`
	Status     bool    `json:"status"`
}

type refreshTokenCacheRepository struct {
	repo  RefreshTokenRepository
	redis redisProvider.RedisService
}

// NewRefreshTokenCacheRepository decora un RefreshTokenRepository con caché Redis.
// Si redis es nil devuelve el repo original sin cambios.
func NewRefreshTokenCacheRepository(repo RefreshTokenRepository, redis redisProvider.RedisService) RefreshTokenRepository {
	if redis == nil {
		return repo
	}
	return &refreshTokenCacheRepository{repo: repo, redis: redis}
}

func rtTokenKey(token string) string { return rtKeyPrefix + token }
func rtUserKey(userID string) string  { return rtUserKeyPrefix + userID }

func (r *refreshTokenCacheRepository) Create(ctx context.Context, tx *gorm.DB, token entities.RefreshToken) (entities.RefreshToken, error) {
	created, err := r.repo.Create(ctx, tx, token)
	if err != nil {
		return entities.RefreshToken{}, err
	}
	r.warmCache(ctx, created)
	return created, nil
}

// FindByToken intenta resolver desde Redis; si falla cae a Postgres y calienta el caché.
func (r *refreshTokenCacheRepository) FindByToken(ctx context.Context, tx *gorm.DB, token string) (entities.RefreshToken, error) {
	if rt, ok := r.getFromCache(ctx, token); ok {
		return rt, nil
	}
	rt, err := r.repo.FindByToken(ctx, tx, token)
	if err != nil {
		return entities.RefreshToken{}, err
	}
	r.warmCache(ctx, rt)
	return rt, nil
}

func (r *refreshTokenCacheRepository) DeleteByToken(ctx context.Context, tx *gorm.DB, token string) error {
	userID := r.getUserIDFromCache(ctx, token)
	if err := r.repo.DeleteByToken(ctx, tx, token); err != nil {
		return err
	}
	r.evictToken(ctx, token, userID)
	return nil
}

func (r *refreshTokenCacheRepository) DeleteByUserID(ctx context.Context, tx *gorm.DB, userID string) error {
	tokens := r.getUserTokensFromCache(ctx, userID)
	if err := r.repo.DeleteByUserID(ctx, tx, userID); err != nil {
		return err
	}
	client := r.redis.Client()
	for _, t := range tokens {
		client.Del(ctx, rtTokenKey(t))
	}
	client.Del(ctx, rtUserKey(userID))
	return nil
}

func (r *refreshTokenCacheRepository) DeleteExpired(ctx context.Context, tx *gorm.DB) error {
	// Redis expira los tokens automáticamente por TTL; solo limpiamos Postgres.
	return r.repo.DeleteExpired(ctx, tx)
}

// warmCache guarda el token en Redis con TTL y lo registra en el índice del usuario.
func (r *refreshTokenCacheRepository) warmCache(ctx context.Context, rt entities.RefreshToken) {
	ttl := time.Until(rt.ExpiresAt)
	if ttl <= 0 {
		return
	}
	ct := cachedToken{
		ID:        rt.ID.String(),
		UserID:    rt.UserID.String(),
		Token:     rt.Token,
		ExpiresAt: rt.ExpiresAt,
		User: cachedUser{
			ID:         rt.User.ID.String(),
			Name:       rt.User.Name,
			Username:   rt.User.Username,
			Email:      rt.User.Email,
			TelpNumber: rt.User.TelpNumber,
			Role:       rt.User.Role,
			ImageUrl:   rt.User.ImageUrl,
			IsVerified: rt.User.IsVerified,
			IsBlocked:  rt.User.IsBlocked,
			Status:     rt.User.Status,
		},
	}
	data, err := json.Marshal(ct)
	if err != nil {
		log.Printf("[rt-cache] marshal error: %v", err)
		return
	}
	if err := r.redis.Set(ctx, rtTokenKey(rt.Token), string(data), ttl); err != nil {
		log.Printf("[rt-cache] set error: %v", err)
		return
	}
	// Índice por usuario para logout masivo
	client := r.redis.Client()
	userKey := rtUserKey(rt.UserID.String())
	client.SAdd(ctx, userKey, rt.Token)
	client.ExpireNX(ctx, userKey, ttl)
}

func (r *refreshTokenCacheRepository) getFromCache(ctx context.Context, token string) (entities.RefreshToken, bool) {
	val, err := r.redis.Get(ctx, rtTokenKey(token))
	if err != nil {
		return entities.RefreshToken{}, false
	}
	var ct cachedToken
	if err := json.Unmarshal([]byte(val), &ct); err != nil {
		log.Printf("[rt-cache] unmarshal error: %v", err)
		return entities.RefreshToken{}, false
	}
	if time.Now().After(ct.ExpiresAt) {
		return entities.RefreshToken{}, false
	}
	tokenID, err := uuid.Parse(ct.ID)
	if err != nil {
		return entities.RefreshToken{}, false
	}
	userID, err := uuid.Parse(ct.UserID)
	if err != nil {
		return entities.RefreshToken{}, false
	}
	userEntityID, err := uuid.Parse(ct.User.ID)
	if err != nil {
		return entities.RefreshToken{}, false
	}
	rt := entities.RefreshToken{
		ID:        tokenID,
		UserID:    userID,
		Token:     ct.Token,
		ExpiresAt: ct.ExpiresAt,
		User: entities.User{
			ID:         userEntityID,
			Name:       ct.User.Name,
			Username:   ct.User.Username,
			Email:      ct.User.Email,
			TelpNumber: ct.User.TelpNumber,
			Role:       ct.User.Role,
			ImageUrl:   ct.User.ImageUrl,
			IsVerified: ct.User.IsVerified,
			IsBlocked:  ct.User.IsBlocked,
		},
	}
	rt.User.Status = ct.User.Status
	return rt, true
}

func (r *refreshTokenCacheRepository) getUserIDFromCache(ctx context.Context, token string) string {
	val, err := r.redis.Get(ctx, rtTokenKey(token))
	if err != nil {
		return ""
	}
	var ct cachedToken
	if err := json.Unmarshal([]byte(val), &ct); err != nil {
		return ""
	}
	return ct.UserID
}

func (r *refreshTokenCacheRepository) evictToken(ctx context.Context, token string, userID string) {
	r.redis.Delete(ctx, rtTokenKey(token))
	if userID != "" {
		r.redis.Client().SRem(ctx, rtUserKey(userID), token)
	}
}

func (r *refreshTokenCacheRepository) getUserTokensFromCache(ctx context.Context, userID string) []string {
	members, err := r.redis.Client().SMembers(ctx, rtUserKey(userID)).Result()
	if err != nil {
		return nil
	}
	return members
}
