package controller

import (
	"net/http"
	"strconv"

	"github.com/Caknoooo/go-gin-clean-starter/modules/app_version/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/app_version/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

type (
	AppVersionController interface {
		Create(ctx *gin.Context)
		GetLatestVersion(ctx *gin.Context)
		GetAll(ctx *gin.Context)
		GetById(ctx *gin.Context)
		Update(ctx *gin.Context)
		ChangeStatus(ctx *gin.Context)
	}

	appVersionController struct {
		appVersionService service.AppVersionService
	}
)

func NewAppVersionController(injector *do.Injector) (AppVersionController, error) {
	appVersionService := do.MustInvoke[service.AppVersionService](injector)
	return &appVersionController{
		appVersionService: appVersionService,
	}, nil
}

func (c *appVersionController) Create(ctx *gin.Context) {
	var req dto.AppVersionRequest
	if err := ctx.ShouldBind(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_CREATE_APP_VERSION, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.appVersionService.Create(ctx.Request.Context(), req)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_CREATE_APP_VERSION, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_CREATE_APP_VERSION, result)
	ctx.JSON(http.StatusOK, res)
}

func (c *appVersionController) GetLatestVersion(ctx *gin.Context) {
	result, err := c.appVersionService.GetLatestVersion(ctx.Request.Context())
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_APP_VERSION, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_APP_VERSION, result)
	ctx.JSON(http.StatusOK, res)
}

func (c *appVersionController) GetAll(ctx *gin.Context) {
	result, err := c.appVersionService.GetAll(ctx.Request.Context())
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_APP_VERSION, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_APP_VERSION, result)
	ctx.JSON(http.StatusOK, res)
}

func (c *appVersionController) GetById(ctx *gin.Context) {
	idParam := ctx.Param("id")
	appId, err := strconv.Atoi(idParam)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_APP_VERSION, "Invalid App ID", nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.appVersionService.GetById(ctx.Request.Context(), appId)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_APP_VERSION, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_APP_VERSION, result)
	ctx.JSON(http.StatusOK, res)
}

func (c *appVersionController) Update(ctx *gin.Context) {
	idParam := ctx.Param("id")
	appId, err := strconv.Atoi(idParam)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_UPDATE_APP_VERSION, "Invalid App ID", nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	var req dto.AppVersionRequest
	if err := ctx.ShouldBind(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_UPDATE_APP_VERSION, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.appVersionService.Update(ctx.Request.Context(), req, appId)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_UPDATE_APP_VERSION, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_UPDATE_APP_VERSION, result)
	ctx.JSON(http.StatusOK, res)
}

func (c  *appVersionController) ChangeStatus(ctx *gin.Context) {
	appId, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.BuildResponseFailed("ID Invalido", "Invalid App ID", nil))
		return
	}

	if err := c.appVersionService.ChangeStatus(ctx.Request.Context(), appId); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.BuildResponseFailed("Fallo procesar solicitud", err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.BuildResponseSuccess("estado actualizado correctamente", nil))
}
