package app_version

import (
	"github.com/Caknoooo/go-gin-clean-starter/middlewares"
	"github.com/Caknoooo/go-gin-clean-starter/modules/app_version/controller"
	"github.com/Caknoooo/go-gin-clean-starter/modules/app_version/repository"
	"github.com/Caknoooo/go-gin-clean-starter/modules/app_version/service"
	auth_service "github.com/Caknoooo/go-gin-clean-starter/modules/auth/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	do.Provide(injector, repository.NewAppVersionRepository)
	do.Provide(injector, service.NewAppVersionService)
	do.Provide(injector, controller.NewAppVersionController)

	appVersionController := do.MustInvoke[controller.AppVersionController](injector)
	jwtService := do.MustInvokeNamed[auth_service.JWTService](injector, constants.JWTService)

	appVersionRoutes := server.Group("/api/app-version")
	{
		appVersionRoutes.POST("", middlewares.Authenticate(jwtService), middlewares.AuthorizeAdmin(jwtService), appVersionController.Create)
		appVersionRoutes.GET("/latest", appVersionController.GetLatestVersion)
		appVersionRoutes.GET("", appVersionController.GetAll)
		appVersionRoutes.GET("/:id", appVersionController.GetById)
		appVersionRoutes.PUT("/:id", middlewares.Authenticate(jwtService), middlewares.AuthorizeAdmin(jwtService), appVersionController.Update)
		appVersionRoutes.PATCH("/:id/status", middlewares.Authenticate(jwtService), middlewares.AuthorizeAdmin(jwtService), appVersionController.ChangeStatus)
	}
}
