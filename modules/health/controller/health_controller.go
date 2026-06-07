package controller

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type HealthController interface {
	Check(ctx *gin.Context)
}

type healthController struct {
	startTime time.Time
}

func NewHealthController() HealthController {
	return &healthController{
		startTime: time.Now(),
	}
}

// Check godoc
// @Summary      Health Check
// @Description  Verifica si el servidor está corriendo
// @Tags         health
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /health [get]
func (c *healthController) Check(ctx *gin.Context) {
	uptime := time.Since(c.startTime)

	ctx.JSON(http.StatusOK, gin.H{
		"status":    true,
		"message":   "El servidor está corriendo correctamente",
		"uptime":    uptime.String(),
		"timestamp": time.Now().Format(time.RFC3339),
	})
}
