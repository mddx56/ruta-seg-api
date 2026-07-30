package vehicle_route

import (
	"github.com/Caknoooo/go-gin-clean-starter/middlewares"
	auth_service "github.com/Caknoooo/go-gin-clean-starter/modules/auth/service"
	"github.com/Caknoooo/go-gin-clean-starter/modules/vehicle_route/controller"
	"github.com/Caknoooo/go-gin-clean-starter/modules/vehicle_route/repository"
	"github.com/Caknoooo/go-gin-clean-starter/modules/vehicle_route/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(router *gin.Engine, injector *do.Injector) {
	// Register dependencies
	do.Provide(injector, repository.NewVehicleRouteRepository)
	do.Provide(injector, service.NewVehicleRouteService)
	do.Provide(injector, controller.NewVehicleRouteController)

	// Invoke controller
	vehicleRouteController := do.MustInvoke[controller.VehicleRouteController](injector)
	jwtService := do.MustInvokeNamed[auth_service.JWTService](injector, constants.JWTService)

	// Define routes — asignar/editar/deshabilitar una asignación es solo de admin;
	// listar/ver requiere estar autenticado (dueños y admin consultan sus asignaciones)
	vehicleRouteGroup := router.Group("/api/vehicle-routes")
	{
		vehicleRouteGroup.POST("", middlewares.Authenticate(jwtService), middlewares.AuthorizeAdmin(jwtService), vehicleRouteController.Create)
		vehicleRouteGroup.POST("/register-micro", middlewares.Authenticate(jwtService), middlewares.AuthorizeAdmin(jwtService), vehicleRouteController.RegisterMicro)
		vehicleRouteGroup.GET("", middlewares.Authenticate(jwtService), vehicleRouteController.FindAll)
		vehicleRouteGroup.GET("/:id", middlewares.Authenticate(jwtService), vehicleRouteController.FindByID)
		vehicleRouteGroup.PUT("/:id", middlewares.Authenticate(jwtService), middlewares.AuthorizeAdmin(jwtService), vehicleRouteController.Update)
		vehicleRouteGroup.PATCH("/:id/status", middlewares.Authenticate(jwtService), middlewares.AuthorizeAdmin(jwtService), vehicleRouteController.ChangeStatus)
	}
}
