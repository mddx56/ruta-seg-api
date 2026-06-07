package controller

import (
	"net/http"

	"github.com/Caknoooo/go-gin-clean-starter/modules/alarm_rule/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/alarm_rule/service"
	"github.com/Caknoooo/go-gin-clean-starter/modules/alarm_rule/validation"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"gorm.io/gorm"
)

type (
	AlarmRuleController interface {
		Create(ctx *gin.Context)
		FindAll(ctx *gin.Context)
	}

	alarmRuleController struct {
		alarmRuleService    service.AlarmRuleService
		alarmRuleValidation *validation.AlarmRuleValidation
		db                  *gorm.DB
	}
)

func NewAlarmRuleController(injector *do.Injector, s service.AlarmRuleService) AlarmRuleController {
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	alarmRuleValidation := validation.NewAlarmRuleValidation()
	return &alarmRuleController{
		alarmRuleService:    s,
		alarmRuleValidation: alarmRuleValidation,
		db:                  db,
	}
}

func (c *alarmRuleController) Create(ctx *gin.Context) {
	var req dto.AlarmRuleCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.BuildResponseFailed("Solicitud Inválida", err.Error(), nil))
		return
	}

	res, err := c.alarmRuleService.Create(ctx.Request.Context(), req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.BuildResponseFailed("Error Interno", err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusCreated, utils.BuildResponseSuccess("Regla de Alarma Creada", res))
}

func (c *alarmRuleController) FindAll(ctx *gin.Context) {
	res, err := c.alarmRuleService.FindAll(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.BuildResponseFailed("Error Interno", err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.BuildResponseSuccess("Exito", res))
}
