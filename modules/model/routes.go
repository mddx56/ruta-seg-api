package model

import (
	"github.com/Caknoooo/go-gin-clean-starter/middlewares"
	auth_service "github.com/Caknoooo/go-gin-clean-starter/modules/auth/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/Caknoooo/go-gin-clean-starter/modules/model/controller"
	"github.com/Caknoooo/go-gin-clean-starter/modules/model/repository"
	"github.com/Caknoooo/go-gin-clean-starter/modules/model/service"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(router *gin.Engine, injector *do.Injector) {
	jwtService := do.MustInvokeNamed[auth_service.JWTService](injector, constants.JWTService)

	// Register dependencies
	do.Provide(injector, repository.NewModelRepository)
	do.Provide(injector, service.NewModelService)
	do.Provide(injector, controller.NewModelController)

	// Invoke controller
	modelController := do.MustInvoke[controller.ModelController](injector)

	// Define routes
	modelGroup := router.Group("/api/models")
	{
		modelGroup.POST("", modelController.Create)
		modelGroup.GET("", modelController.FindAll)
		modelGroup.GET("/:id", modelController.FindByID)
		modelGroup.PUT("/:id", modelController.Update)
		modelGroup.PATCH("/:id/status", middlewares.Authenticate(jwtService), middlewares.AuthorizeInstallerOrAdmin(jwtService), modelController.ChangeStatus)
	}
}
