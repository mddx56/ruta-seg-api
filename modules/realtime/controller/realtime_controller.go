package controller

import (
	"log"
	"net/http"

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
		Hub:    c.wsService.GetHub(),
		Conn:   conn,
		UserID: userID,
		Role:   role,
		Send:   make(chan []byte, 256),
	}

	client.Hub.Register <- client

	// Allow collection of memory referenced by the caller by doing all work in new goroutines.
	// The hub/pump goroutines own the connection lifetime from here on.
	go client.WritePump()
	go client.ReadPump()
}
