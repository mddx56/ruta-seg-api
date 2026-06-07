package vehicle

import (
	"github.com/Caknoooo/go-gin-clean-starter/middlewares"
	auth_service "github.com/Caknoooo/go-gin-clean-starter/modules/auth/service"
	"github.com/Caknoooo/go-gin-clean-starter/modules/vehicle/controller"
	"github.com/Caknoooo/go-gin-clean-starter/modules/vehicle/repository"
	vehicleService "github.com/Caknoooo/go-gin-clean-starter/modules/vehicle/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(router *gin.Engine, injector *do.Injector) {
	// Register dependencies
	do.Provide(injector, repository.NewVehicleRepository)
	do.Provide(injector, vehicleService.NewVehicleService)
	do.Provide(injector, controller.NewVehicleController)

	// Invoke controller
	vehicleController := do.MustInvoke[controller.VehicleController](injector)
	jwtService := do.MustInvokeNamed[auth_service.JWTService](injector, constants.JWTService)

	// Define routes
	vehicleGroup := router.Group("/api/vehicles")
	{
		vehicleGroup.POST("", middlewares.Authenticate(jwtService), vehicleController.Create)
		vehicleGroup.GET("", middlewares.Authenticate(jwtService), vehicleController.FindAll)
		vehicleGroup.GET("/simple", middlewares.Authenticate(jwtService), vehicleController.GetSimple)
		vehicleGroup.GET("/by-chassis/:chassis", middlewares.Authenticate(jwtService), vehicleController.FindByChassisFull)
		vehicleGroup.GET("/:id", middlewares.Authenticate(jwtService), vehicleController.FindByID)
		vehicleGroup.PUT("/:id", middlewares.Authenticate(jwtService), vehicleController.Update)
		vehicleGroup.PATCH("/:id/status", middlewares.Authenticate(jwtService), vehicleController.ChangeStatus)
	}
}
