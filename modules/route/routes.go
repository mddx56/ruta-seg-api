package route

import (
	"github.com/Caknoooo/go-gin-clean-starter/middlewares"
	auth_service "github.com/Caknoooo/go-gin-clean-starter/modules/auth/service"
	"github.com/Caknoooo/go-gin-clean-starter/modules/route/controller"
	"github.com/Caknoooo/go-gin-clean-starter/modules/route/repository"
	"github.com/Caknoooo/go-gin-clean-starter/modules/route/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(router *gin.Engine, injector *do.Injector) {
	// Register dependencies
	do.Provide(injector, repository.NewRouteRepository)
	do.Provide(injector, repository.NewRouteLiveRepository)
	do.Provide(injector, service.NewRouteService)
	do.Provide(injector, service.NewRouteLiveService)
	do.Provide(injector, controller.NewRouteController)

	// Invoke controller
	routeController := do.MustInvoke[controller.RouteController](injector)
	jwtService := do.MustInvokeNamed[auth_service.JWTService](injector, constants.JWTService)

	// Define routes — GET es público (sin auth) porque el rol Público y la pantalla TV
	// necesitan ver las rutas sin login; Create/Update/ChangeStatus quedan solo para admin.
	routeGroup := router.Group("/api/routes")
	{
		routeGroup.POST("", middlewares.Authenticate(jwtService), middlewares.AuthorizeAdmin(jwtService), routeController.Create)
		routeGroup.GET("", routeController.FindAll)
		routeGroup.GET("/:id", routeController.FindByID)
		routeGroup.GET("/:id/live", routeController.FindLiveVehicles)
		routeGroup.GET("/:id/eta", routeController.FindETA)
		routeGroup.PUT("/:id", middlewares.Authenticate(jwtService), middlewares.AuthorizeAdmin(jwtService), routeController.Update)
		routeGroup.PATCH("/:id/status", middlewares.Authenticate(jwtService), middlewares.AuthorizeAdmin(jwtService), routeController.ChangeStatus)
	}
}
