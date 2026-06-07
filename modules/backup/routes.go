package backup

import (
	"github.com/Caknoooo/go-gin-clean-starter/middlewares"
	authService "github.com/Caknoooo/go-gin-clean-starter/modules/auth/service"
	"github.com/Caknoooo/go-gin-clean-starter/modules/backup/controller"
	"github.com/Caknoooo/go-gin-clean-starter/modules/backup/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	do.Provide(injector, service.NewBackupService)
	do.Provide(injector, controller.NewBackupController)

	backupController := do.MustInvoke[controller.BackupController](injector)
	jwtService := do.MustInvokeNamed[authService.JWTService](injector, constants.JWTService)

	backupRoutes := server.Group("/api/backup")
	{
		// All backup routes protected by Authenticate AND AuthorizeAdmin
		backupRoutes.Use(middlewares.Authenticate(jwtService))
		backupRoutes.Use(middlewares.AuthorizeAdmin(jwtService))

		// Unified Backup Endpoint
		// Use query param ?type=schema for schema backup, defaults to full backup
		backupRoutes.Any("", backupController.HandleUnifiedBackup)
	}
}
