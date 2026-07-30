package geofence

import (
	"github.com/Caknoooo/go-gin-clean-starter/middlewares"
	auth_service "github.com/Caknoooo/go-gin-clean-starter/modules/auth/service"
	"github.com/Caknoooo/go-gin-clean-starter/modules/geofence/controller"
	"github.com/Caknoooo/go-gin-clean-starter/modules/geofence/repository"
	"github.com/Caknoooo/go-gin-clean-starter/modules/geofence/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(router *gin.Engine, injector *do.Injector) {
	// Register dependencies
	do.Provide(injector, repository.NewGeofenceRepository)
	do.Provide(injector, service.NewGeofenceService)
	do.Provide(injector, controller.NewGeofenceController)

	// Invoke controller
	geofenceController := do.MustInvoke[controller.GeofenceController](injector)
	jwtService := do.MustInvokeNamed[auth_service.JWTService](injector, constants.JWTService)

	// Define routes — cualquier usuario autenticado puede crear/listar geocercas
	// (tanto admins como dueños de rutas las necesitan)
	geofenceGroup := router.Group("/api/geofences")
	{
		geofenceGroup.POST("", middlewares.Authenticate(jwtService), geofenceController.Create)
		geofenceGroup.GET("", middlewares.Authenticate(jwtService), geofenceController.FindAll)
		geofenceGroup.GET("/:id", middlewares.Authenticate(jwtService), geofenceController.FindByID)
		geofenceGroup.PUT("/:id", middlewares.Authenticate(jwtService), geofenceController.Update)
		geofenceGroup.PATCH("/:id/status", middlewares.Authenticate(jwtService), geofenceController.ChangeStatus)
	}
}
