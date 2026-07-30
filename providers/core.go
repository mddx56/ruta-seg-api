package providers

import (
	"context"
	"errors"
	"os"

	"github.com/Caknoooo/go-gin-clean-starter/config"
	appVersionRepoPkg "github.com/Caknoooo/go-gin-clean-starter/modules/app_version/repository"
	authController "github.com/Caknoooo/go-gin-clean-starter/modules/auth/controller"
	authRepo "github.com/Caknoooo/go-gin-clean-starter/modules/auth/repository"
	authService "github.com/Caknoooo/go-gin-clean-starter/modules/auth/service"
	diRepo "github.com/Caknoooo/go-gin-clean-starter/modules/device_installation/repository"
	groupRepo "github.com/Caknoooo/go-gin-clean-starter/modules/group/repository"
	healthController "github.com/Caknoooo/go-gin-clean-starter/modules/health/controller"
	notificationService "github.com/Caknoooo/go-gin-clean-starter/modules/notification/service"
	userController "github.com/Caknoooo/go-gin-clean-starter/modules/user/controller"
	"github.com/Caknoooo/go-gin-clean-starter/modules/user/repository"
	userService "github.com/Caknoooo/go-gin-clean-starter/modules/user/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/fcm"
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
	redisService, _ := do.InvokeNamed[redisProvider.RedisService](injector, "Redis")
	refreshTokenRepository := authRepo.NewRefreshTokenCacheRepository(
		authRepo.NewRefreshTokenRepository(db),
		redisService,
	)
	deviceInstallationRepo, _ := diRepo.NewDeviceInstallationRepository(injector)
	groupRepository, _ := groupRepo.NewGroupRepository(injector)
	appVersionRepo, _ := appVersionRepoPkg.NewAppVersionRepository(injector)

	userService := userService.NewUserService(userRepository, deviceInstallationRepo, groupRepository, db)
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

		// Si Redis está disponible, escucha el canal de eventos de rutas y los reenvía
		// al hub local (así el canal público de WS funciona igual con varias instancias).
		if redisSvc, err := do.InvokeNamed[redisProvider.RedisService](i, "Redis"); err == nil {
			go websocket.StartRouteEventSubscriber(context.Background(), redisSvc, hub)
		}

		return wsService, nil
	})

	// RouteEventPublisher: usado por el motor de reglas (lap/fine) para publicar
	// posiciones en vivo y eventos de vuelta al canal público de micros en ruta.
	do.Provide(injector, func(i *do.Injector) (websocket.RouteEventPublisher, error) {
		redisSvc, err := do.InvokeNamed[redisProvider.RedisService](i, "Redis")
		if err != nil {
			return nil, err
		}
		return websocket.NewRouteEventPublisher(redisSvc), nil
	})

	// FCM PushSender: opcional. Si no hay credenciales configuradas (FCM_CREDENTIALS_FILE
	// o FCM_CREDENTIALS_JSON), los módulos que lo usan (notification) simplemente no envían
	// push y siguen funcionando (persistencia + WebSocket siguen intactos).
	do.Provide(injector, func(i *do.Injector) (notificationService.PushSender, error) {
		ctx := context.Background()

		if raw := os.Getenv("FCM_CREDENTIALS_JSON"); raw != "" {
			return fcm.NewClientFromJSON(ctx, []byte(raw))
		}
		if path := os.Getenv("FCM_CREDENTIALS_FILE"); path != "" {
			return fcm.NewClientFromFile(ctx, path)
		}

		return nil, errors.New("FCM no configurado: falta FCM_CREDENTIALS_FILE o FCM_CREDENTIALS_JSON")
	})
}
