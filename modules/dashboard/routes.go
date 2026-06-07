package dashboard

import (
	"github.com/Caknoooo/go-gin-clean-starter/middlewares"
	authService "github.com/Caknoooo/go-gin-clean-starter/modules/auth/service"
	"github.com/Caknoooo/go-gin-clean-starter/modules/dashboard/controller"
	"github.com/Caknoooo/go-gin-clean-starter/modules/dashboard/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"gorm.io/gorm"
)

func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	// Register service with proper provider function
	do.Provide(injector, func(i *do.Injector) (service.DashboardService, error) {
		db := do.MustInvokeNamed[*gorm.DB](i, constants.DB)
		return service.NewDashboardService(db), nil
	})

	do.Provide(injector, controller.NewDashboardController)

	dashboardController := do.MustInvoke[controller.DashboardController](injector)
	jwtService := do.MustInvokeNamed[authService.JWTService](injector, constants.JWTService)

	dashboardRoutes := server.Group("/api/dashboard")
	{
		// Protected routes - Admin only
		dashboardRoutes.Use(middlewares.Authenticate(jwtService))
		dashboardRoutes.Use(middlewares.AuthorizeAdmin(jwtService))

		dashboardRoutes.GET("/stats", dashboardController.GetStats)
	}
}
