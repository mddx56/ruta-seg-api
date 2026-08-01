package controller

import (
	"log"
	"net/http"
	"strings"

	"github.com/Caknoooo/go-gin-clean-starter/modules/auth/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/Caknoooo/go-gin-clean-starter/providers/websocket"
	"github.com/gin-gonic/gin"
	gorillaWs "github.com/gorilla/websocket"
	"github.com/samber/do"
)

var upgrader = gorillaWs.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all connections for now
		return true
	},
}

type RealtimeController interface {
	ServeWS(ctx *gin.Context)
	ServePublicWS(ctx *gin.Context)
}

type realtimeController struct {
	wsService  websocket.WebsocketService
	jwtService service.JWTService
}

func NewRealtimeController(injector *do.Injector) RealtimeController {
	wsSvc := do.MustInvoke[websocket.WebsocketService](injector)
	jwtSvc := do.MustInvokeNamed[service.JWTService](injector, constants.JWTService)
	return &realtimeController{
		wsService:  wsSvc,
		jwtService: jwtSvc,
	}
}

func (c *realtimeController) ServeWS(ctx *gin.Context) {
	// Upgrade the connection first so we can signal errors via WS close codes
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Println("ws upgrade error:", err)
		return
	}

	tokenString := ctx.Query("token")
	if tokenString == "" {
		_ = conn.WriteMessage(gorillaWs.CloseMessage, gorillaWs.FormatCloseMessage(4002, "No token provided"))
		conn.Close()
		return
	}

	userID, err := c.jwtService.GetUserIDByToken(tokenString)
	if err != nil {
		// Distinguish expired vs. otherwise invalid tokens so the client can act accordingly
		if err.Error() == "token is expired" {
			log.Println("[WS] Token expirado para conexión entrante, cerrando con 4001")
			_ = conn.WriteMessage(gorillaWs.CloseMessage, gorillaWs.FormatCloseMessage(4001, "TOKEN_EXPIRED"))
		} else {
			log.Println("[WS] Token inválido, cerrando con 4002")
			_ = conn.WriteMessage(gorillaWs.CloseMessage, gorillaWs.FormatCloseMessage(4002, "INVALID_TOKEN"))
		}
		conn.Close()
		return
	}

	role, _ := c.jwtService.GetRoleByToken(tokenString)

	client := &websocket.Client{
		Hub:               c.wsService.GetHub(),
		Conn:              conn,
		UserID:            userID,
		Role:              role,
		AllowedVehicleIDs: parseVehicleIDsFilter(ctx.Query("vehicle_ids")),
		Send:              make(chan []byte, 256),
	}

	client.Hub.Register <- client

	// Allow collection of memory referenced by the caller by doing all work in new goroutines.
	// The hub/pump goroutines own the connection lifetime from here on.
	go client.WritePump()
	go client.ReadPump()
}

// parseVehicleIDsFilter convierte "?vehicle_ids=a,b,c" en un set; vacío devuelve nil (sin filtro).
func parseVehicleIDsFilter(raw string) map[string]bool {
	if raw == "" {
		return nil
	}
	ids := make(map[string]bool)
	for _, id := range strings.Split(raw, ",") {
		if id = strings.TrimSpace(id); id != "" {
			ids[id] = true
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return ids
}

// ServePublicWS expone un canal de solo lectura, sin autenticación, para clientes
// públicos (pantalla TV, app pública) que quieren ver los micros en vivo de una ruta.
// Con ?route_id=<uuid> se suscribe solo a esa ruta; sin ese parámetro, recibe los
// eventos de todas las rutas (topic "route:all").
func (c *realtimeController) ServePublicWS(ctx *gin.Context) {
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Println("ws upgrade error:", err)
		return
	}

	topic := websocket.TopicAll
	if routeID := ctx.Query("route_id"); routeID != "" {
		topic = "route:" + routeID
	}

	client := &websocket.Client{
		Hub:   c.wsService.GetHub(),
		Conn:  conn,
		Role:  "public",
		Topic: topic,
		Send:  make(chan []byte, 256),
	}

	client.Hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}
