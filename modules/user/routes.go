package user

import (
	"github.com/Caknoooo/go-gin-clean-starter/middlewares"
	"github.com/Caknoooo/go-gin-clean-starter/modules/auth/service"
	"github.com/Caknoooo/go-gin-clean-starter/modules/user/controller"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	userController := do.MustInvoke[controller.UserController](injector)
	jwtService := do.MustInvokeNamed[service.JWTService](injector, constants.JWTService)

	userRoutes := server.Group("/api/user")
	{
		userRoutes.POST("", userController.Create)
		userRoutes.GET("", middlewares.Authenticate(jwtService), middlewares.AuthorizeAdmin(jwtService), userController.GetAllUser)
		userRoutes.GET("/simple", middlewares.Authenticate(jwtService), middlewares.AuthorizeInstallerOrAdmin(jwtService), userController.GetAllSimple)
		userRoutes.GET("/me", middlewares.Authenticate(jwtService), userController.Me)
		userRoutes.PUT("/me", middlewares.Authenticate(jwtService), userController.UpdateMe)
		userRoutes.PUT("/me/password", middlewares.Authenticate(jwtService), userController.ChangeMyPassword)
		userRoutes.GET("/:id", middlewares.Authenticate(jwtService), middlewares.AuthorizeInstallerOrAdmin(jwtService), userController.GetById)
		userRoutes.GET("/:id/devices", middlewares.Authenticate(jwtService), userController.GetInstalledDevices)
		userRoutes.PUT("/:id/reset-password", middlewares.Authenticate(jwtService), middlewares.AuthorizeAdmin(jwtService), userController.AdminResetPassword)
		userRoutes.PUT("/:id", middlewares.Authenticate(jwtService), userController.Update)
		userRoutes.PATCH("/:id/block", middlewares.Authenticate(jwtService), middlewares.AuthorizeAdmin(jwtService), userController.UpdateBlockStatus)
		userRoutes.PATCH("/:id/status", middlewares.Authenticate(jwtService), middlewares.AuthorizeAdmin(jwtService), userController.ChangeStatus)
	}
}
