package make

import (
	"github.com/Caknoooo/go-gin-clean-starter/middlewares"
	auth_service "github.com/Caknoooo/go-gin-clean-starter/modules/auth/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/Caknoooo/go-gin-clean-starter/modules/make/controller"
	"github.com/Caknoooo/go-gin-clean-starter/modules/make/repository"
	"github.com/Caknoooo/go-gin-clean-starter/modules/make/service"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(router *gin.Engine, injector *do.Injector) {
	jwtService := do.MustInvokeNamed[auth_service.JWTService](injector, constants.JWTService)

	// Register dependencies
	do.Provide(injector, repository.NewMakeRepository)
	do.Provide(injector, service.NewMakeService)
	do.Provide(injector, controller.NewMakeController)

	// Invoke controller
	makeController := do.MustInvoke[controller.MakeController](injector)

	// Define routes
	makeGroup := router.Group("/api/makes")
	{
		makeGroup.POST("", makeController.Create)
		makeGroup.GET("", makeController.FindAll)
		makeGroup.GET("/:id", makeController.FindByID)
		makeGroup.PUT("/:id", makeController.Update)
		makeGroup.PATCH("/:id/status", middlewares.Authenticate(jwtService), middlewares.AuthorizeInstallerOrAdmin(jwtService), makeController.ChangeStatus)
	}
}
