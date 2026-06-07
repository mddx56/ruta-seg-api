package controller

import (
	"net/http"

	"github.com/Caknoooo/go-gin-clean-starter/modules/dashboard/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/dashboard/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

type DashboardController interface {
	GetStats(ctx *gin.Context)
}

type dashboardController struct {
	dashboardService service.DashboardService
}

func NewDashboardController(injector *do.Injector) (DashboardController, error) {
	dashboardService := do.MustInvoke[service.DashboardService](injector)
	return &dashboardController{
		dashboardService: dashboardService,
	}, nil
}

// GetStats godoc
// @Summary      Obtener Estadísticas del Dashboard
// @Description  Retorna estadísticas básicas del sistema para el dashboard administrativo
// @Tags         dashboard
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  utils.Response{data=dto.DashboardStatsResponse}
// @Failure      401  {object}  utils.Response     "No autorizado"
// @Failure      403  {object}  utils.Response     "Acceso denegado (Solo Admin)"
// @Failure      500  {object}  utils.Response     "Error interno del servidor"
// @Router       /api/dashboard/stats [get]
func (c *dashboardController) GetStats(ctx *gin.Context) {
	stats, err := c.dashboardService.GetStats(ctx.Request.Context())
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_STATS, err.Error(), nil)
		ctx.JSON(http.StatusInternalServerError, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_STATS, stats)
	ctx.JSON(http.StatusOK, res)
}
