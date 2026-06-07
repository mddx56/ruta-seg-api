package position

import (
	"github.com/Caknoooo/go-gin-clean-starter/middlewares"
	"github.com/Caknoooo/go-gin-clean-starter/modules/auth/service"
	"github.com/Caknoooo/go-gin-clean-starter/modules/position/controller"
	"github.com/Caknoooo/go-gin-clean-starter/modules/position/repository"
	positionService "github.com/Caknoooo/go-gin-clean-starter/modules/position/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"gorm.io/gorm"
)

func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	do.Provide(injector, func(i *do.Injector) (repository.PositionRepository, error) {
		db := do.MustInvokeNamed[*gorm.DB](i, constants.DB)
		return repository.NewPositionRepository(db), nil
	})

	do.Provide(injector, func(i *do.Injector) (positionService.PositionService, error) {
		repo := do.MustInvoke[repository.PositionRepository](i)
		return positionService.NewPositionService(repo), nil
	})

	do.Provide(injector, func(i *do.Injector) (controller.PositionController, error) {
		svc := do.MustInvoke[positionService.PositionService](i)
		return controller.NewPositionController(i, svc), nil
	})

	positionController := do.MustInvoke[controller.PositionController](injector)
	jwtService := do.MustInvokeNamed[service.JWTService](injector, constants.JWTService)

	routes := server.Group("/api/positions")
	{
		routes.GET("/history", middlewares.Authenticate(jwtService), positionController.GetDeviceHistory)
		routes.POST("", positionController.CreatePosition) // Sin autenticación

		routes.GET("/last", middlewares.Authenticate(jwtService), positionController.GetLastPosition)
		routes.GET("/latest", middlewares.Authenticate(jwtService), positionController.GetLastPositionsOfAllDevices)
		routes.GET("/device", middlewares.Authenticate(jwtService), positionController.GetPositionsByIMEI)
		routes.GET("/device-details", middlewares.Authenticate(jwtService), middlewares.AuthorizeAdmin(jwtService), positionController.GetPositionsWithVehicleInfoByIMEIAndDate)
		routes.GET("/coordinates", middlewares.Authenticate(jwtService), positionController.GetCoordinatesByIMEIAndDate)
		routes.GET("/route", middlewares.Authenticate(jwtService), positionController.GetDeviceRoute)
		routes.GET("/:id", middlewares.Authenticate(jwtService), positionController.GetPositionByID)
		// routes.DELETE("/:id", middlewares.Authenticate(jwtService), positionController.DeletePosition)
	}
}
