package device

import (
	"github.com/Caknoooo/go-gin-clean-starter/middlewares"
	auth_service "github.com/Caknoooo/go-gin-clean-starter/modules/auth/service"
	"github.com/Caknoooo/go-gin-clean-starter/modules/device/controller"
	"github.com/Caknoooo/go-gin-clean-starter/modules/device/repository"
	deviceService "github.com/Caknoooo/go-gin-clean-starter/modules/device/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	do.Provide(injector, repository.NewDeviceRepository)
	do.Provide(injector, deviceService.NewDeviceService)
	do.Provide(injector, controller.NewDeviceController)

	deviceController := do.MustInvoke[controller.DeviceController](injector)
	jwtService := do.MustInvokeNamed[auth_service.JWTService](injector, constants.JWTService)

	routes := server.Group("/api/devices")
	{
		routes.POST("", middlewares.Authenticate(jwtService), deviceController.CreateDevice)
		routes.PUT("", middlewares.Authenticate(jwtService), deviceController.UpdateDevice)
		routes.GET("", deviceController.GetAllDevices)            // Sin autenticación
		routes.GET("/categories", middlewares.Authenticate(jwtService), deviceController.GetCategorizedDevices)
		routes.GET("/list", deviceController.GetDevicesPaginated) // Nuevo endpoint paginado
		routes.GET("/simple", middlewares.Authenticate(jwtService), deviceController.GetSimpleDevices)
		routes.GET("/:id/full", middlewares.Authenticate(jwtService), deviceController.GetDeviceFullByIMEI)
		routes.GET("/:id", middlewares.Authenticate(jwtService), deviceController.GetDeviceByIMEI)
		routes.PATCH("/:id/status", middlewares.Authenticate(jwtService), middlewares.AuthorizeInstallerOrAdmin(jwtService), deviceController.ChangeStatus)

		// Export
		routes.GET("/export", middlewares.Authenticate(jwtService), middlewares.AuthorizeInstallerOrAdmin(jwtService), deviceController.ExportDevices)

		// Bulk import
		routes.POST("/bulk/validate", middlewares.Authenticate(jwtService), middlewares.AuthorizeInstallerOrAdmin(jwtService), deviceController.BulkValidateDevices)
		routes.POST("/bulk/import", middlewares.Authenticate(jwtService), middlewares.AuthorizeInstallerOrAdmin(jwtService), deviceController.BulkImportDevices)
	}
}
