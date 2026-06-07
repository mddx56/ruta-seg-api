package health

import (
	"github.com/Caknoooo/go-gin-clean-starter/modules/health/controller"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	healthController := do.MustInvoke[controller.HealthController](injector)

	server.GET("/health", healthController.Check)
}
