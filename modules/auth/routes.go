package auth

import (
	"time"

	"github.com/Caknoooo/go-gin-clean-starter/middlewares"
	"github.com/Caknoooo/go-gin-clean-starter/modules/auth/controller"
	"github.com/Caknoooo/go-gin-clean-starter/modules/auth/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	authController := do.MustInvoke[controller.AuthController](injector)
	jwtService := do.MustInvokeNamed[service.JWTService](injector, constants.JWTService)

	const (
		signupMaxRequests = 3
		signupWindow      = time.Minute
	)
	signupLimit := middlewares.RateLimit(signupMaxRequests, signupWindow)

	authRoutes := server.Group("/api/auth")
	{
		authRoutes.POST("/register", authController.Register)
		authRoutes.POST("/signup", signupLimit, authController.Signup)
		authRoutes.POST("/login", authController.Login)
		authRoutes.POST("/refresh", authController.RefreshToken)
		authRoutes.POST("/logout", middlewares.Authenticate(jwtService), authController.Logout)
		authRoutes.POST("/send-verification-email", authController.SendVerificationEmail)
		authRoutes.POST("/verify-email", authController.VerifyEmail)
		authRoutes.POST("/send-password-reset", authController.SendPasswordReset)
		authRoutes.POST("/reset-password", authController.ResetPassword)
		authRoutes.GET("/google", authController.GoogleRedirect)
		authRoutes.GET("/google/callback", authController.GoogleCallback)
	}
}
