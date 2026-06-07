package vehicletype

import (
	"github.com/Caknoooo/go-gin-clean-starter/middlewares"
	auth_service "github.com/Caknoooo/go-gin-clean-starter/modules/auth/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/Caknoooo/go-gin-clean-starter/modules/vehicle_type/controller"
	"github.com/Caknoooo/go-gin-clean-starter/modules/vehicle_type/repository"
	"github.com/Caknoooo/go-gin-clean-starter/modules/vehicle_type/service"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(router *gin.Engine, injector *do.Injector) {
	jwtService := do.MustInvokeNamed[auth_service.JWTService](injector, constants.JWTService)

	// Register dependencies
	do.Provide(injector, repository.NewVehicleTypeRepository)
	do.Provide(injector, service.NewVehicleTypeService)
	do.Provide(injector, controller.NewVehicleTypeController)

	// Invoke controller
	vehicleTypeController := do.MustInvoke[controller.VehicleTypeController](injector)

	// Define routes
	vehicleTypeGroup := router.Group("/api/vehicle-types")
	{
		vehicleTypeGroup.POST("", vehicleTypeController.Create)
		vehicleTypeGroup.GET("", vehicleTypeController.FindAll)
		vehicleTypeGroup.GET("/:id", vehicleTypeController.FindByID)
		vehicleTypeGroup.PUT("/:id", vehicleTypeController.Update)
		vehicleTypeGroup.PATCH("/:id/status", middlewares.Authenticate(jwtService), middlewares.AuthorizeInstallerOrAdmin(jwtService), vehicleTypeController.ChangeStatus)
	}
}
