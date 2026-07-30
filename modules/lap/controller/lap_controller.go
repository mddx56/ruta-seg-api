package controller

import (
	"net/http"

	"github.com/Caknoooo/go-gin-clean-starter/modules/lap/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/lap/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/samber/do"
)

type LapController interface {
	FindAll(ctx *gin.Context)
	FindByID(ctx *gin.Context)
}

type lapController struct {
	service service.LapService
}

func NewLapController(injector *do.Injector) (LapController, error) {
	service := do.MustInvoke[service.LapService](injector)
	return &lapController{
		service: service,
	}, nil
}

// FindAllLaps godoc
// @Summary      List all laps
// @Description  Get a list of all laps recorded by the lap-counting engine
// @Tags         laps
// @Accept       json
// @Produce      json
// @Success      200  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /api/laps [get]
func (c *lapController) FindAll(ctx *gin.Context) {
	res, err := c.service.FindAll(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.BuildResponseFailed(dto.MESSAGE_INTERNAL_SERVER_ERROR, err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS, res))
}

// FindLapByID godoc
// @Summary      Get a lap by ID
// @Description  Get a lap by ID
// @Tags         laps
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Lap ID"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /api/laps/{id} [get]
func (c *lapController) FindByID(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.BuildResponseFailed(dto.MESSAGE_FAILED_INVALID_ID, err.Error(), nil))
		return
	}

	res, err := c.service.FindByID(ctx.Request.Context(), id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.BuildResponseFailed(dto.MESSAGE_INTERNAL_SERVER_ERROR, err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS, res))
}
