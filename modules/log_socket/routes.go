package log_socket

import (
	"github.com/Caknoooo/go-gin-clean-starter/middlewares"
	"github.com/Caknoooo/go-gin-clean-starter/modules/log_socket/controller"
	"github.com/Caknoooo/go-gin-clean-starter/modules/log_socket/repository"
	"github.com/Caknoooo/go-gin-clean-starter/modules/log_socket/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"gorm.io/gorm"
)

func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	// Register Repository
	do.Provide(injector, func(i *do.Injector) (repository.LogSocketRepository, error) {
		db := do.MustInvokeNamed[*gorm.DB](i, constants.DB)
		return repository.NewLogSocketRepository(db), nil
	})

	// Register Service
	do.Provide(injector, func(i *do.Injector) (service.LogSocketService, error) {
		repo := do.MustInvoke[repository.LogSocketRepository](i)
		db := do.MustInvokeNamed[*gorm.DB](i, constants.DB)
		return service.NewLogSocketService(repo, db), nil
	})

	// Register Controller
	do.Provide(injector, func(i *do.Injector) (controller.LogSocketController, error) {
		svc := do.MustInvoke[service.LogSocketService](i)
		return controller.NewLogSocketController(i, svc), nil
	})

	logSocketController := do.MustInvoke[controller.LogSocketController](injector)

	routes := server.Group("/api/log-socket")
	{
		routes.GET("", logSocketController.GetAll)
		routes.POST("", middlewares.AuthenticateAPIKey(), logSocketController.Create)
	}
}
