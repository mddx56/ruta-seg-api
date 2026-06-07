package controller

import (
	"net/http"

	"github.com/Caknoooo/go-gin-clean-starter/modules/model/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/model/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/samber/do"
)

type ModelController interface {
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	ChangeStatus(ctx *gin.Context)
	FindAll(ctx *gin.Context)
	FindByID(ctx *gin.Context)
}

type modelController struct {
	service service.ModelService
}

func NewModelController(injector *do.Injector) (ModelController, error) {
	service := do.MustInvoke[service.ModelService](injector)
	return &modelController{
		service: service,
	}, nil
}

// CreateModel godoc
// @Summary      Create a new model
// @Description  Create a new vehicle model with the input payload
// @Tags         models
// @Accept       json
// @Produce      json
// @Param        model  body      dto.ModelCreateRequest  true  "Model Create Request"
// @Success      201    {object}  utils.Response
// @Failure      400    {object}  utils.Response
// @Failure      500    {object}  utils.Response
// @Router       /api/models [post]
func (c *modelController) Create(ctx *gin.Context) {
	var req dto.ModelCreateRequest
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

// UpdateModel godoc
// @Summary      Update an existing model
// @Description  Update an existing vehicle model with the input payload
// @Tags         models
// @Accept       json
// @Produce      json
// @Param        id     path      string                  true  "Model ID"
// @Param        model  body      dto.ModelUpdateRequest  true  "Model Update Request"
// @Success      200    {object}  utils.Response
// @Failure      400    {object}  utils.Response
// @Failure      500    {object}  utils.Response
// @Router       /api/models/{id} [put]
func (c *modelController) Update(ctx *gin.Context) {
	var req dto.ModelUpdateRequest
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

// DeleteModel godoc
// @Summary      Change status a model (soft delete)
// @Description  Soft delete a vehicle model by setting status to false
// @Tags         models
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Model ID"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /api/models/{id}/status [patch]
func (c  *modelController) ChangeStatus(ctx *gin.Context) {
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

// FindAllModels godoc
// @Summary      List all models
// @Description  Get a list of all active vehicle models with their make and type
// @Tags         models
// @Accept       json
// @Produce      json
// @Success      200  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /api/models [get]
func (c *modelController) FindAll(ctx *gin.Context) {
	res, err := c.service.FindAll(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.BuildResponseFailed(dto.MESSAGE_INTERNAL_SERVER_ERROR, err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS, res))
}

// FindModelByID godoc
// @Summary      Get a model by ID
// @Description  Get a vehicle model by ID with its make and type
// @Tags         models
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Model ID"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /api/models/{id} [get]
func (c *modelController) FindByID(ctx *gin.Context) {
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
