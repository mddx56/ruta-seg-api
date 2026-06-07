package controller

import (
	"net/http"

	"github.com/Caknoooo/go-gin-clean-starter/modules/vehicle_type/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/vehicle_type/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/samber/do"
)

type VehicleTypeController interface {
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	ChangeStatus(ctx *gin.Context)
	FindAll(ctx *gin.Context)
	FindByID(ctx *gin.Context)
}

type vehicleTypeController struct {
	service service.VehicleTypeService
}

func NewVehicleTypeController(injector *do.Injector) (VehicleTypeController, error) {
	service := do.MustInvoke[service.VehicleTypeService](injector)
	return &vehicleTypeController{
		service: service,
	}, nil
}

// CreateVehicleType godoc
// @Summary      Create a new vehicle type
// @Description  Create a new vehicle type with the input payload
// @Tags         vehicle-types
// @Accept       json
// @Produce      json
// @Param        vehicleType  body      dto.VehicleTypeCreateRequest  true  "VehicleType Create Request"
// @Success      201          {object}  utils.Response
// @Failure      400          {object}  utils.Response
// @Failure      500          {object}  utils.Response
// @Router       /api/vehicle-types [post]
func (c *vehicleTypeController) Create(ctx *gin.Context) {
	var req dto.VehicleTypeCreateRequest
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

// UpdateVehicleType godoc
// @Summary      Update an existing vehicle type
// @Description  Update an existing vehicle type with the input payload
// @Tags         vehicle-types
// @Accept       json
// @Produce      json
// @Param        id           path      string                        true  "VehicleType ID"
// @Param        vehicleType  body      dto.VehicleTypeUpdateRequest  true  "VehicleType Update Request"
// @Success      200          {object}  utils.Response
// @Failure      400          {object}  utils.Response
// @Failure      500          {object}  utils.Response
// @Router       /api/vehicle-types/{id} [put]
func (c *vehicleTypeController) Update(ctx *gin.Context) {
	var req dto.VehicleTypeUpdateRequest
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

// DeleteVehicleType godoc
// @Summary      Change status a vehicle type (soft delete)
// @Description  Soft delete a vehicle type by setting status to false
// @Tags         vehicle-types
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "VehicleType ID"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /api/vehicle-types/{id}/status [patch]
func (c  *vehicleTypeController) ChangeStatus(ctx *gin.Context) {
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

// FindAllVehicleTypes godoc
// @Summary      List all vehicle types
// @Description  Get a list of all active vehicle types
// @Tags         vehicle-types
// @Accept       json
// @Produce      json
// @Success      200  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /api/vehicle-types [get]
func (c *vehicleTypeController) FindAll(ctx *gin.Context) {
	res, err := c.service.FindAll(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.BuildResponseFailed(dto.MESSAGE_INTERNAL_SERVER_ERROR, err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS, res))
}

// FindVehicleTypeByID godoc
// @Summary      Get a vehicle type by ID
// @Description  Get a vehicle type by ID
// @Tags         vehicle-types
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "VehicleType ID"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /api/vehicle-types/{id} [get]
func (c *vehicleTypeController) FindByID(ctx *gin.Context) {
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
