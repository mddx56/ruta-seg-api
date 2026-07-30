package lap

import (
	"github.com/Caknoooo/go-gin-clean-starter/middlewares"
	auth_service "github.com/Caknoooo/go-gin-clean-starter/modules/auth/service"
	"github.com/Caknoooo/go-gin-clean-starter/modules/lap/controller"
	"github.com/Caknoooo/go-gin-clean-starter/modules/lap/repository"
	"github.com/Caknoooo/go-gin-clean-starter/modules/lap/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(router *gin.Engine, injector *do.Injector) {
	// Register dependencies
	do.Provide(injector, repository.NewLapRepository)
	do.Provide(injector, service.NewLapService)
	do.Provide(injector, controller.NewLapController)

	// Invoke controller
	lapController := do.MustInvoke[controller.LapController](injector)
	jwtService := do.MustInvokeNamed[auth_service.JWTService](injector, constants.JWTService)

	// Las vueltas las genera el motor de reglas al procesar posiciones; acá solo se consultan.
	lapGroup := router.Group("/api/laps")
	{
		lapGroup.GET("", middlewares.Authenticate(jwtService), lapController.FindAll)
		lapGroup.GET("/:id", middlewares.Authenticate(jwtService), lapController.FindByID)
	}
}
