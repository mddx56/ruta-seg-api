package route_fare

import (
	"github.com/Caknoooo/go-gin-clean-starter/middlewares"
	auth_service "github.com/Caknoooo/go-gin-clean-starter/modules/auth/service"
	"github.com/Caknoooo/go-gin-clean-starter/modules/route_fare/controller"
	"github.com/Caknoooo/go-gin-clean-starter/modules/route_fare/repository"
	"github.com/Caknoooo/go-gin-clean-starter/modules/route_fare/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(router *gin.Engine, injector *do.Injector) {
	// Register dependencies
	do.Provide(injector, repository.NewRouteFareRepository)
	do.Provide(injector, service.NewRouteFareService)
	do.Provide(injector, controller.NewRouteFareController)

	// Invoke controller
	routeFareController := do.MustInvoke[controller.RouteFareController](injector)
	jwtService := do.MustInvokeNamed[auth_service.JWTService](injector, constants.JWTService)

	// Define routes — configurar tarifas es solo de admin; consultarlas requiere estar autenticado
	routeFareGroup := router.Group("/api/route-fares")
	{
		routeFareGroup.POST("", middlewares.Authenticate(jwtService), middlewares.AuthorizeAdmin(jwtService), routeFareController.Create)
		routeFareGroup.GET("", middlewares.Authenticate(jwtService), routeFareController.FindAll)
		routeFareGroup.GET("/:id", middlewares.Authenticate(jwtService), routeFareController.FindByID)
		routeFareGroup.PUT("/:id", middlewares.Authenticate(jwtService), middlewares.AuthorizeAdmin(jwtService), routeFareController.Update)
		routeFareGroup.PATCH("/:id/status", middlewares.Authenticate(jwtService), middlewares.AuthorizeAdmin(jwtService), routeFareController.ChangeStatus)
	}
}
