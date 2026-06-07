package alarm_incident

import (
	// "github.com/Caknoooo/go-gin-clean-starter/modules/alarm_incident/controller"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	// _ = do.MustInvoke[controller.AlarmIncidentController](injector)

	_ = server.Group("/api/alarm_incident")
	{
		// TODO: add your endpoints here
	}
}
