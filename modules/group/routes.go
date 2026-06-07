package group

import (
	"github.com/Caknoooo/go-gin-clean-starter/middlewares"
	auth_service "github.com/Caknoooo/go-gin-clean-starter/modules/auth/service"
	"github.com/Caknoooo/go-gin-clean-starter/modules/group/controller"
	"github.com/Caknoooo/go-gin-clean-starter/modules/group/repository"
	groupService "github.com/Caknoooo/go-gin-clean-starter/modules/group/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	do.Provide(injector, repository.NewGroupRepository)
	do.Provide(injector, groupService.NewGroupService)
	do.Provide(injector, controller.NewGroupController)

	groupController := do.MustInvoke[controller.GroupController](injector)
	jwtService := do.MustInvokeNamed[auth_service.JWTService](injector, constants.JWTService)

	routes := server.Group("/api/group")
	{
		routes.POST("", middlewares.Authenticate(jwtService), groupController.Create)
		routes.PUT("", middlewares.Authenticate(jwtService), groupController.Update)
		routes.GET("", middlewares.Authenticate(jwtService), groupController.FindAll)
		routes.GET("/user/:user_id", middlewares.Authenticate(jwtService), groupController.FindAllByUserID)
		routes.PATCH("/:id/status", middlewares.Authenticate(jwtService), middlewares.AuthorizeInstallerOrAdmin(jwtService), groupController.ChangeStatus)
		routes.POST("/assign", middlewares.Authenticate(jwtService), groupController.AssignDevice)
		routes.POST("/remove", middlewares.Authenticate(jwtService), groupController.RemoveDevice)
	}
}
