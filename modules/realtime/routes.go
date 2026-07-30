package realtime

import (
	"github.com/Caknoooo/go-gin-clean-starter/modules/realtime/controller"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	// Register Controller
	do.Provide(injector, func(i *do.Injector) (controller.RealtimeController, error) {
		return controller.NewRealtimeController(i), nil
	})

	realtimeController := do.MustInvoke[controller.RealtimeController](injector)

	routes := server.Group("/api/realtime")
	{
		routes.GET("/ws", realtimeController.ServeWS)
		// Público, sin autenticación: usado por la pantalla TV y la app pública para
		// ver los micros en vivo de una ruta (?route_id=<uuid>) o de todas.
		routes.GET("/public/ws", realtimeController.ServePublicWS)
	}
}
