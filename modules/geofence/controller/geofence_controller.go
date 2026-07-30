package controller

import (
	"net/http"

	"github.com/Caknoooo/go-gin-clean-starter/modules/geofence/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/geofence/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/samber/do"
)

type GeofenceController interface {
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	ChangeStatus(ctx *gin.Context)
	FindAll(ctx *gin.Context)
	FindByID(ctx *gin.Context)
}

type geofenceController struct {
	service service.GeofenceService
}

func NewGeofenceController(injector *do.Injector) (GeofenceController, error) {
	service := do.MustInvoke[service.GeofenceService](injector)
	return &geofenceController{
		service: service,
	}, nil
}

// CreateGeofence godoc
// @Summary      Create a new geofence
// @Description  Create a new geofence (circle or polygon) with its ordered points
// @Tags         geofences
// @Accept       json
// @Produce      json
// @Param        geofence  body      dto.GeofenceCreateRequest  true  "Geofence Create Request"
// @Success      201       {object}  utils.Response
// @Failure      400       {object}  utils.Response
// @Failure      500       {object}  utils.Response
// @Router       /api/geofences [post]
func (c *geofenceController) Create(ctx *gin.Context) {
	var req dto.GeofenceCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.BuildResponseFailed(dto.MESSAGE_FAILED_BAD_REQUEST, err.Error(), nil))
		return
	}

	userIDStr := ctx.MustGet("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, utils.BuildResponseFailed("No autorizado", "user_id inválido en el token", nil))
		return
	}

	res, err := c.service.Create(ctx.Request.Context(), req, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.BuildResponseFailed(dto.MESSAGE_INTERNAL_SERVER_ERROR, err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusCreated, utils.BuildResponseSuccess(dto.MESSAGE_CREATED, res))
}

// UpdateGeofence godoc
// @Summary      Update an existing geofence
// @Description  Update an existing geofence; if points is present, it fully replaces the existing point list
// @Tags         geofences
// @Accept       json
// @Produce      json
// @Param        id        path      string                     true  "Geofence ID"
// @Param        geofence  body      dto.GeofenceUpdateRequest  true  "Geofence Update Request"
// @Success      200       {object}  utils.Response
// @Failure      400       {object}  utils.Response
// @Failure      500       {object}  utils.Response
// @Router       /api/geofences/{id} [put]
func (c *geofenceController) Update(ctx *gin.Context) {
	var req dto.GeofenceUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.BuildResponseFailed(dto.MESSAGE_FAILED_BAD_REQUEST, err.Error(), nil))
		return
	}

	id, err := uuid.Parse(ctx.Param("id"))
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

// ChangeGeofenceStatus godoc
// @Summary      Change status of a geofence (soft delete)
// @Description  Toggle the status of a geofence
// @Tags         geofences
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Geofence ID"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /api/geofences/{id}/status [patch]
func (c *geofenceController) ChangeStatus(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.BuildResponseFailed(dto.MESSAGE_FAILED_INVALID_ID, err.Error(), nil))
		return
	}

	if err := c.service.ChangeStatus(ctx.Request.Context(), id); err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.BuildResponseFailed(dto.MESSAGE_INTERNAL_SERVER_ERROR, err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.BuildResponseSuccess("estado actualizado correctamente", nil))
}

// FindAllGeofences godoc
// @Summary      List all geofences
// @Description  Get a list of all active geofences with their points
// @Tags         geofences
// @Accept       json
// @Produce      json
// @Success      200  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /api/geofences [get]
func (c *geofenceController) FindAll(ctx *gin.Context) {
	res, err := c.service.FindAll(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.BuildResponseFailed(dto.MESSAGE_INTERNAL_SERVER_ERROR, err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS, res))
}

// FindGeofenceByID godoc
// @Summary      Get a geofence by ID
// @Description  Get a geofence with its points by ID
// @Tags         geofences
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Geofence ID"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /api/geofences/{id} [get]
func (c *geofenceController) FindByID(ctx *gin.Context) {
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
