package fine

import (
	"github.com/Caknoooo/go-gin-clean-starter/middlewares"
	auth_service "github.com/Caknoooo/go-gin-clean-starter/modules/auth/service"
	"github.com/Caknoooo/go-gin-clean-starter/modules/fine/controller"
	"github.com/Caknoooo/go-gin-clean-starter/modules/fine/repository"
	"github.com/Caknoooo/go-gin-clean-starter/modules/fine/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(router *gin.Engine, injector *do.Injector) {
	// Register dependencies
	do.Provide(injector, repository.NewFineRepository)
	do.Provide(injector, repository.NewFineTypeRepository)
	do.Provide(injector, service.NewFineService)
	do.Provide(injector, controller.NewFineController)

	// Invoke controller
	fineController := do.MustInvoke[controller.FineController](injector)
	jwtService := do.MustInvokeNamed[auth_service.JWTService](injector, constants.JWTService)

	// Las multas las genera el motor de reglas. Ver todas es solo de admin (para no
	// exponer infracciones de otros dueños); cada dueño consulta las suyas en /mine.
	fineGroup := router.Group("/api/fines")
	{
		fineGroup.GET("", middlewares.Authenticate(jwtService), middlewares.AuthorizeAdmin(jwtService), fineController.FindAll)
		fineGroup.GET("/mine", middlewares.Authenticate(jwtService), fineController.FindAllMine)
		fineGroup.GET("/:id", middlewares.Authenticate(jwtService), middlewares.AuthorizeAdmin(jwtService), fineController.FindByID)
		fineGroup.PATCH("/:id/void", middlewares.Authenticate(jwtService), middlewares.AuthorizeAdmin(jwtService), fineController.Void)
	}

	router.GET("/api/fine-types", middlewares.Authenticate(jwtService), fineController.FindAllTypes)
}
