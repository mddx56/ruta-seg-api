package middlewares

import (
	"fmt"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/Caknoooo/go-gin-clean-starter/pkg/utils"
	"github.com/gin-gonic/gin"
)

type rateLimitEntry struct {
	mu          sync.Mutex
	count       int
	windowStart time.Time
}

type RateLimiter struct {
	requests int
	window   time.Duration
	clients  sync.Map
}

func NewRateLimiter(requests int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: requests,
		window:   window,
	}
	go rl.cleanup()
	return rl
}

// cleanup elimina entradas vencidas cada cierto intervalo para no acumular memoria.
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window * 2)
	defer ticker.Stop()
	for range ticker.C {
		rl.clients.Range(func(key, value any) bool {
			entry := value.(*rateLimitEntry)
			entry.mu.Lock()
			expired := time.Since(entry.windowStart) > rl.window
			entry.mu.Unlock()
			if expired {
				rl.clients.Delete(key)
			}
			return true
		})
	}
}

func (rl *RateLimiter) allow(ip string) bool {
	now := time.Now()

	val, _ := rl.clients.LoadOrStore(ip, &rateLimitEntry{windowStart: now})
	entry := val.(*rateLimitEntry)

	entry.mu.Lock()
	defer entry.mu.Unlock()

	// Reinicia la ventana si ya expiró
	if time.Since(entry.windowStart) > rl.window {
		entry.count = 0
		entry.windowStart = now
	}

	entry.count++
	return entry.count <= rl.requests
}

func formatWindow(d time.Duration) string {
	if d < time.Minute {
		seconds := int(math.Round(d.Seconds()))
		if seconds == 1 {
			return "1 segundo"
		}
		return fmt.Sprintf("%d segundos", seconds)
	}
	minutes := int(math.Round(d.Minutes()))
	if minutes == 1 {
		return "1 minuto"
	}
	return fmt.Sprintf("%d minutos", minutes)
}

// RateLimit devuelve un middleware que limita a `requests` peticiones por `window` por IP.
// Ejemplo: RateLimit(10, time.Minute) → máx 10 req/min por IP.
func RateLimit(requests int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(requests, window)
	errMsg := fmt.Sprintf("límite de intentos alcanzado, espere %s e intente de nuevo", formatWindow(window))

	return func(ctx *gin.Context) {
		ip := ctx.ClientIP()
		if !limiter.allow(ip) {
			res := utils.BuildResponseFailed("demasiadas solicitudes", errMsg, nil)
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, res)
			return
		}
		ctx.Next()
	}
}
