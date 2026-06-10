package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/common-nighthawk/go-figure"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/Caknoooo/go-gin-clean-starter/database/migrations"
	"github.com/Caknoooo/go-gin-clean-starter/docs"
	"github.com/Caknoooo/go-gin-clean-starter/middlewares"
	"github.com/Caknoooo/go-gin-clean-starter/modules/alarm_incident"
	"github.com/Caknoooo/go-gin-clean-starter/modules/alarm_rule"
	"github.com/Caknoooo/go-gin-clean-starter/modules/alarm_type"
	appversionmodule "github.com/Caknoooo/go-gin-clean-starter/modules/app_version"
	"github.com/Caknoooo/go-gin-clean-starter/modules/auth"
	authRepo "github.com/Caknoooo/go-gin-clean-starter/modules/auth/repository"
	"github.com/Caknoooo/go-gin-clean-starter/modules/backup"
	"github.com/Caknoooo/go-gin-clean-starter/modules/dashboard"
	"github.com/Caknoooo/go-gin-clean-starter/modules/device"
	deviceinstallation "github.com/Caknoooo/go-gin-clean-starter/modules/device_installation"
	"github.com/Caknoooo/go-gin-clean-starter/modules/group"
	"github.com/Caknoooo/go-gin-clean-starter/modules/health"
	logsocket "github.com/Caknoooo/go-gin-clean-starter/modules/log_socket"
	makemodule "github.com/Caknoooo/go-gin-clean-starter/modules/make"
	modelmodule "github.com/Caknoooo/go-gin-clean-starter/modules/model"
	"github.com/Caknoooo/go-gin-clean-starter/modules/position"
	"github.com/Caknoooo/go-gin-clean-starter/modules/realtime"
	"github.com/Caknoooo/go-gin-clean-starter/modules/user"
	vehiclemodule "github.com/Caknoooo/go-gin-clean-starter/modules/vehicle"
	vehicletype "github.com/Caknoooo/go-gin-clean-starter/modules/vehicle_type"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"github.com/Caknoooo/go-gin-clean-starter/providers"
	redisProvider "github.com/Caknoooo/go-gin-clean-starter/providers/redis"
	"github.com/Caknoooo/go-gin-clean-starter/script"
	"gorm.io/gorm"

	devService "github.com/Caknoooo/go-gin-clean-starter/modules/device/service"
	posService "github.com/Caknoooo/go-gin-clean-starter/modules/position/service"
	"github.com/Caknoooo/go-gin-clean-starter/providers/grpc_server"
	providerWS "github.com/Caknoooo/go-gin-clean-starter/providers/websocket"
)

// @title           RutaSegura API
// @version         0.1.2
// @description     API para gestionar motos
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  mddx56@gmail.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8888
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func args(injector *do.Injector) bool {
	if len(os.Args) > 1 {
		flag := script.Commands(injector)
		return flag
	}

	return true
}

func run(server *gin.Engine) {
	server.Static("/assets", "./assets")
	server.StaticFile("/logs", "./logs.html")
	server.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	port := os.Getenv("GOLANG_PORT")
	if port == "" {
		port = "8888"
	}

	var serve string
	if os.Getenv("APP_ENV") == "localhost" {
		serve = "0.0.0.0:" + port
	} else {
		serve = ":" + port
	}

	myFigure := figure.NewColorFigure("Mddx56", "", "green", true)
	myFigure.Print()
	log.Println("\nSwagger UI is available at http://localhost:" + port + "/swagger/index.html")

	if err := server.Run(serve); err != nil {
		log.Fatalf("error running server: %v", err)
	}
}

func main() {
	var (
		injector = do.New()
	)

	if host := os.Getenv("APP_HOST"); host != "" {
		docs.SwaggerInfo.Host = host
	}

	providers.RegisterDependencies(injector)

	if !args(injector) {
		return
	}

	server := gin.Default()

	// Middleware de compresión JSON (GZIP) Optimización
	server.Use(gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedPaths([]string{"/api/realtime/ws"})))

	server.Use(middlewares.CORSMiddleware())

	// Register module routes
	health.RegisterRoutes(server, injector)
	appversionmodule.RegisterRoutes(server, injector)
	user.RegisterRoutes(server, injector)
	auth.RegisterRoutes(server, injector)
	alarm_rule.RegisterRoutes(server, injector)
	alarm_incident.RegisterRoutes(server, injector)
	alarm_type.RegisterRoutes(server, injector)
	device.RegisterRoutes(server, injector)
	makemodule.RegisterRoutes(server, injector)
	vehicletype.RegisterRoutes(server, injector)
	modelmodule.RegisterRoutes(server, injector)
	vehiclemodule.RegisterRoutes(server, injector)
	position.RegisterRoutes(server, injector)
	deviceinstallation.RegisterRoutes(server, injector)
	logsocket.RegisterRoutes(server, injector)
	group.RegisterRoutes(server, injector)
	backup.RegisterRoutes(server, injector)
	dashboard.RegisterRoutes(server, injector)
	realtime.RegisterRoutes(server, injector)

	db, err := do.InvokeNamed[*gorm.DB](injector, constants.DB)
	if err == nil {
		go func() {
			repo := authRepo.NewRefreshTokenRepository(db)
			// Run immediately and then every 24 hours
			_ = repo.DeleteExpired(context.Background(), nil)
			ticker := time.NewTicker(24 * time.Hour)
			defer ticker.Stop()
			for range ticker.C {
				_ = repo.DeleteExpired(context.Background(), nil)
			}
		}()

		// Warm-up: carga device_last_positions desde Postgres a Redis al iniciar
		if redisSvc, redisErr := do.InvokeNamed[redisProvider.RedisService](injector, "Redis"); redisErr == nil {
			go func() {
				cache := redisProvider.NewDevicePositionCache(redisSvc)
				var rows []entities.DeviceLastPosition
				if err := db.Find(&rows).Error; err != nil {
					log.Printf("[pos-cache] warm-up error al leer device_last_positions: %v", err)
					return
				}
				ctx := context.Background()
				for _, row := range rows {
					_ = cache.SetNX(ctx, redisProvider.CachedPosition{
						IMEI:       row.IMEI,
						Latitude:   row.Latitude,
						Longitude:  row.Longitude,
						Speed:      row.Speed,
						Course:     row.Course,
						DeviceTime: row.DeviceTime,
						ServerTime: row.ServerTime,
						Attributes: row.Attributes,
					})
				}
				log.Printf("[pos-cache] warm-up completado: %d dispositivos cargados en Redis", len(rows))
			}()
		}
	}

	// ── INICIAR GRPC SERVER ──
	go func() {
		devSvc := do.MustInvoke[devService.DeviceService](injector)
		posSvc := do.MustInvoke[posService.PositionService](injector)
		wsSvc, _ := do.Invoke[providerWS.WebsocketService](injector)

		portGRPC := os.Getenv("GRPC_PORT")
		if portGRPC == "" {
			portGRPC = ":50051"
		} else {
			portGRPC = ":" + portGRPC
		}

		grpc_server.StartGRPCServer(devSvc, posSvc, wsSvc, db, portGRPC)
	}()

	run(server)
}
