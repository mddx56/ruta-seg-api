package providers

import (
	"os"

	"github.com/Caknoooo/go-gin-clean-starter/config"
	appVersionRepoPkg "github.com/Caknoooo/go-gin-clean-starter/modules/app_version/repository"
	authController "github.com/Caknoooo/go-gin-clean-starter/modules/auth/controller"
	authRepo "github.com/Caknoooo/go-gin-clean-starter/modules/auth/repository"
	authService "github.com/Caknoooo/go-gin-clean-starter/modules/auth/service"
	diRepo "github.com/Caknoooo/go-gin-clean-starter/modules/device_installation/repository"
	groupRepo "github.com/Caknoooo/go-gin-clean-starter/modules/group/repository"
	healthController "github.com/Caknoooo/go-gin-clean-starter/modules/health/controller"
	userController "github.com/Caknoooo/go-gin-clean-starter/modules/user/controller"
	"github.com/Caknoooo/go-gin-clean-starter/modules/user/repository"
	userService "github.com/Caknoooo/go-gin-clean-starter/modules/user/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	redisProvider "github.com/Caknoooo/go-gin-clean-starter/providers/redis"
	"github.com/Caknoooo/go-gin-clean-starter/providers/websocket"
	"github.com/samber/do"
	"gorm.io/gorm"
)

func InitDatabase(injector *do.Injector) {
	do.ProvideNamed(injector, constants.DB, func(i *do.Injector) (*gorm.DB, error) {
		return config.SetUpDatabaseConnection(), nil
	})
}

func InitRedis(injector *do.Injector) {
	do.ProvideNamed(injector, "Redis", func(i *do.Injector) (redisProvider.RedisService, error) {
		host := os.Getenv("REDIS_HOST")
		port := os.Getenv("REDIS_PORT")
		password := os.Getenv("REDIS_PASSWORD")
		return redisProvider.NewRedisService(host, port, password)
	})
}

func RegisterDependencies(injector *do.Injector) {
	InitDatabase(injector)
	InitRedis(injector)

	do.ProvideNamed(injector, constants.JWTService, func(i *do.Injector) (authService.JWTService, error) {
		return authService.NewJWTService(), nil
	})

	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	jwtService := do.MustInvokeNamed[authService.JWTService](injector, constants.JWTService)

	userRepository := repository.NewUserRepository(db)
	refreshTokenRepository := authRepo.NewRefreshTokenRepository(db)
	deviceInstallationRepo, _ := diRepo.NewDeviceInstallationRepository(injector)
	groupRepository, _ := groupRepo.NewGroupRepository(injector)
	appVersionRepo, _ := appVersionRepoPkg.NewAppVersionRepository(injector)

	userService := userService.NewUserService(userRepository, deviceInstallationRepo, groupRepository, db)
	redisService, _ := do.InvokeNamed[redisProvider.RedisService](injector, "Redis")
	authService := authService.NewAuthService(userRepository, refreshTokenRepository, appVersionRepo, jwtService, redisService, db)

	do.Provide(
		injector, func(i *do.Injector) (userController.UserController, error) {
			return userController.NewUserController(i, userService), nil
		},
	)

	do.Provide(
		injector, func(i *do.Injector) (authController.AuthController, error) {
			return authController.NewAuthController(i, authService), nil
		},
	)

	do.Provide(
		injector, func(i *do.Injector) (healthController.HealthController, error) {
			return healthController.NewHealthController(), nil
		},
	)

	// WebSocket Service
	do.Provide(injector, func(i *do.Injector) (websocket.WebsocketService, error) {
		hub := websocket.NewHub()
		wsService := websocket.NewWebsocketService(hub)
		// Run hub in background
		go wsService.RunHub()
		return wsService, nil
	})
}
