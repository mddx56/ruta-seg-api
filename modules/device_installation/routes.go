package device_installation

import (
	"github.com/Caknoooo/go-gin-clean-starter/middlewares"
	"github.com/Caknoooo/go-gin-clean-starter/modules/auth/service"
	"github.com/Caknoooo/go-gin-clean-starter/modules/device_installation/controller"
	"github.com/Caknoooo/go-gin-clean-starter/modules/device_installation/repository"
	diService "github.com/Caknoooo/go-gin-clean-starter/modules/device_installation/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	do.Provide(injector, repository.NewDeviceInstallationRepository)
	do.Provide(injector, diService.NewDeviceInstallationService)
	do.Provide(injector, controller.NewDeviceInstallationController)

	diController := do.MustInvoke[controller.DeviceInstallationController](injector)
	jwtService := do.MustInvokeNamed[service.JWTService](injector, constants.JWTService)

	routes := server.Group("/api/device-installations")
	{
		routes.POST("", middlewares.Authenticate(jwtService), middlewares.AuthorizeInstallerOrAdmin(jwtService), diController.CreateInstallation)
		routes.POST("/quick", middlewares.Authenticate(jwtService), middlewares.AuthorizeInstallerOrAdmin(jwtService), diController.QuickCreateInstallation)
		routes.GET("/mine", middlewares.Authenticate(jwtService), diController.GetMyInstallations)
		routes.GET("", middlewares.Authenticate(jwtService), diController.GetAllInstallations)
		routes.GET("/device", middlewares.Authenticate(jwtService), diController.GetInstallationsByIMEI)
		routes.GET("/vehicle", middlewares.Authenticate(jwtService), diController.GetInstallationsByVehicleID)
		routes.PUT("/:id/uninstall", middlewares.Authenticate(jwtService), diController.UninstallInstallation)
	}
}
