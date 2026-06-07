package controller

import (
	"net/http"

	"github.com/Caknoooo/go-gin-clean-starter/modules/make/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/make/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/samber/do"
)

type MakeController interface {
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	ChangeStatus(ctx *gin.Context)
	FindAll(ctx *gin.Context)
	FindByID(ctx *gin.Context)
}

type makeController struct {
	service service.MakeService
}

func NewMakeController(injector *do.Injector) (MakeController, error) {
	service := do.MustInvoke[service.MakeService](injector)
	return &makeController{
		service: service,
	}, nil
}

// CreateMake godoc
// @Summary      Create a new make
// @Description  Create a new vehicle make with the input payload
// @Tags         makes
// @Accept       json
// @Produce      json
// @Param        make  body      dto.MakeCreateRequest  true  "Make Create Request"
// @Success      201   {object}  utils.Response
// @Failure      400   {object}  utils.Response
// @Failure      500   {object}  utils.Response
// @Router       /api/makes [post]
func (c *makeController) Create(ctx *gin.Context) {
	var req dto.MakeCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.BuildResponseFailed(dto.MESSAGE_FAILED_BAD_REQUEST, err.Error(), nil))
		return
	}

	res, err := c.service.Create(ctx.Request.Context(), req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.BuildResponseFailed(dto.MESSAGE_INTERNAL_SERVER_ERROR, err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusCreated, utils.BuildResponseSuccess(dto.MESSAGE_CREATED, res))
}

// UpdateMake godoc
// @Summary      Update an existing make
// @Description  Update an existing vehicle make with the input payload
// @Tags         makes
// @Accept       json
// @Produce      json
// @Param        id    path      string                 true  "Make ID"
// @Param        make  body      dto.MakeUpdateRequest  true  "Make Update Request"
// @Success      200   {object}  utils.Response
// @Failure      400   {object}  utils.Response
// @Failure      500   {object}  utils.Response
// @Router       /api/makes/{id} [put]
func (c *makeController) Update(ctx *gin.Context) {
	var req dto.MakeUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.BuildResponseFailed(dto.MESSAGE_FAILED_BAD_REQUEST, err.Error(), nil))
		return
	}

	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.BuildResponseFailed(dto.MESSAGE_FAILED_INVALID_ID, err.Error(), nil))
		return
	}
	req.ID = id

	res, err := c.service.Update(ctx.Request.Context(), req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.BuildResponseFailed(dto.MESSAGE_INTERNAL_SERVER_ERROR, err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.BuildResponseSuccess(dto.MESSAGE_UPDATED, res))
}

// DeleteMake godoc
// @Summary      Change status a make (soft delete)
// @Description  Soft delete a vehicle make by setting status to false
// @Tags         makes
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Make ID"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /api/makes/{id}/status [patch]
func (c  *makeController) ChangeStatus(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.BuildResponseFailed("ID Invalido", err.Error(), nil))
		return
	}

	if err := c.service.ChangeStatus(ctx.Request.Context(), id); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.BuildResponseFailed("Fallo procesar solicitud", err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.BuildResponseSuccess("estado actualizado correctamente", nil))
}

// FindAllMakes godoc
// @Summary      List all makes
// @Description  Get a list of all active vehicle makes
// @Tags         makes
// @Accept       json
// @Produce      json
// @Success      200  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /api/makes [get]
func (c *makeController) FindAll(ctx *gin.Context) {
	res, err := c.service.FindAll(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.BuildResponseFailed(dto.MESSAGE_INTERNAL_SERVER_ERROR, err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS, res))
}

// FindMakeByID godoc
// @Summary      Get a make by ID
// @Description  Get a vehicle make by ID
// @Tags         makes
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Make ID"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /api/makes/{id} [get]
func (c *makeController) FindByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
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
