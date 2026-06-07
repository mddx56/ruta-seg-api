package controller

import (
	"net/http"

	"github.com/Caknoooo/go-gin-clean-starter/modules/alarm_type/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/alarm_type/service"
	"github.com/Caknoooo/go-gin-clean-starter/modules/alarm_type/validation"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"gorm.io/gorm"
)

type (
	AlarmTypeController interface {
		Create(ctx *gin.Context)
		FindAll(ctx *gin.Context)
	}

	alarmTypeController struct {
		alarmTypeService    service.AlarmTypeService
		alarmTypeValidation *validation.AlarmTypeValidation
		db                  *gorm.DB
	}
)

func NewAlarmTypeController(injector *do.Injector, s service.AlarmTypeService) AlarmTypeController {
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	alarmTypeValidation := validation.NewAlarmTypeValidation()
	return &alarmTypeController{
		alarmTypeService:    s,
		alarmTypeValidation: alarmTypeValidation,
		db:                  db,
	}
}

func (c *alarmTypeController) Create(ctx *gin.Context) {
	var req dto.AlarmTypeCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.BuildResponseFailed("Solicitud Inválida", err.Error(), nil))
		return
	}

	res, err := c.alarmTypeService.Create(ctx.Request.Context(), req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.BuildResponseFailed("Error Interno", err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusCreated, utils.BuildResponseSuccess("Tipo de Alarma Creado", res))
}

func (c *alarmTypeController) FindAll(ctx *gin.Context) {
	res, err := c.alarmTypeService.FindAll(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.BuildResponseFailed("Error Interno", err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.BuildResponseSuccess("Exito", res))
}
