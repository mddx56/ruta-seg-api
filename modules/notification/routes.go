package notification

import (
	"github.com/Caknoooo/go-gin-clean-starter/middlewares"
	auth_service "github.com/Caknoooo/go-gin-clean-starter/modules/auth/service"
	"github.com/Caknoooo/go-gin-clean-starter/modules/notification/controller"
	"github.com/Caknoooo/go-gin-clean-starter/modules/notification/repository"
	"github.com/Caknoooo/go-gin-clean-starter/modules/notification/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(router *gin.Engine, injector *do.Injector) {
	// Register dependencies
	do.Provide(injector, repository.NewDeviceTokenRepository)
	do.Provide(injector, repository.NewNotificationRepository)
	do.Provide(injector, service.NewNotificationService)
	do.Provide(injector, controller.NewNotificationController)

	// Invoke controller
	notificationController := do.MustInvoke[controller.NotificationController](injector)
	jwtService := do.MustInvokeNamed[auth_service.JWTService](injector, constants.JWTService)

	// Todo requiere estar autenticado: cada usuario gestiona sus propios tokens y notificaciones.
	notificationGroup := router.Group("/api/notifications")
	{
		notificationGroup.POST("/device-tokens", middlewares.Authenticate(jwtService), notificationController.RegisterDeviceToken)
		notificationGroup.DELETE("/device-tokens", middlewares.Authenticate(jwtService), notificationController.UnregisterDeviceToken)
		notificationGroup.GET("", middlewares.Authenticate(jwtService), notificationController.FindAllMine)
		notificationGroup.PATCH("/:id/read", middlewares.Authenticate(jwtService), notificationController.MarkRead)
	}
}
