package controller

import (
	"net/http"

	"github.com/Caknoooo/go-gin-clean-starter/modules/vehicle_route/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/vehicle_route/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/samber/do"
)

type VehicleRouteController interface {
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	ChangeStatus(ctx *gin.Context)
	FindAll(ctx *gin.Context)
	FindByID(ctx *gin.Context)
	RegisterMicro(ctx *gin.Context)
}

type vehicleRouteController struct {
	service service.VehicleRouteService
}

func NewVehicleRouteController(injector *do.Injector) (VehicleRouteController, error) {
	service := do.MustInvoke[service.VehicleRouteService](injector)
	return &vehicleRouteController{
		service: service,
	}, nil
}

// CreateVehicleRoute godoc
// @Summary      Assign a vehicle to a route
// @Description  Assign a micro (vehicle) to a route, with the pin number shown on the map
// @Tags         vehicle-routes
// @Accept       json
// @Produce      json
// @Param        vehicle_route  body      dto.VehicleRouteCreateRequest  true  "Vehicle Route Create Request"
// @Success      201            {object}  utils.Response
// @Failure      400            {object}  utils.Response
// @Failure      500            {object}  utils.Response
// @Router       /api/vehicle-routes [post]
func (c *vehicleRouteController) Create(ctx *gin.Context) {
	var req dto.VehicleRouteCreateRequest
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

// RegisterMicro godoc
// @Summary      Register a microbus in one call
// @Description  Creates (or reuses) the Model under VehicleType code=BUS, creates the Vehicle, and optionally assigns it to a route with a pin number
// @Tags         vehicle-routes
// @Accept       json
// @Produce      json
// @Param        micro  body      dto.RegisterMicroRequest  true  "Register Micro Request"
// @Success      201    {object}  utils.Response
// @Failure      400    {object}  utils.Response
// @Router       /api/vehicle-routes/register-micro [post]
func (c *vehicleRouteController) RegisterMicro(ctx *gin.Context) {
	var req dto.RegisterMicroRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.BuildResponseFailed(dto.MESSAGE_FAILED_BAD_REQUEST, err.Error(), nil))
		return
	}

	res, err := c.service.RegisterMicro(ctx.Request.Context(), req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.BuildResponseFailed(dto.MESSAGE_FAILED_BAD_REQUEST, err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusCreated, utils.BuildResponseSuccess(dto.MESSAGE_CREATED, res))
}

// UpdateVehicleRoute godoc
// @Summary      Update a vehicle-route assignment
// @Description  Update the route, pin number or active flag of an existing assignment
// @Tags         vehicle-routes
// @Accept       json
// @Produce      json
// @Param        id             path      string                         true  "Vehicle Route ID"
// @Param        vehicle_route  body      dto.VehicleRouteUpdateRequest  true  "Vehicle Route Update Request"
// @Success      200            {object}  utils.Response
// @Failure      400            {object}  utils.Response
// @Failure      500            {object}  utils.Response
// @Router       /api/vehicle-routes/{id} [put]
func (c *vehicleRouteController) Update(ctx *gin.Context) {
	var req dto.VehicleRouteUpdateRequest
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

// ChangeVehicleRouteStatus godoc
// @Summary      Change status of a vehicle-route assignment (soft delete)
// @Description  Toggle the status of a vehicle-route assignment
// @Tags         vehicle-routes
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Vehicle Route ID"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /api/vehicle-routes/{id}/status [patch]
func (c *vehicleRouteController) ChangeStatus(ctx *gin.Context) {
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

// FindAllVehicleRoutes godoc
// @Summary      List all vehicle-route assignments
// @Description  Get a list of all active vehicle-route assignments
// @Tags         vehicle-routes
// @Accept       json
// @Produce      json
// @Success      200  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /api/vehicle-routes [get]
func (c *vehicleRouteController) FindAll(ctx *gin.Context) {
	res, err := c.service.FindAll(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.BuildResponseFailed(dto.MESSAGE_INTERNAL_SERVER_ERROR, err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS, res))
}

// FindVehicleRouteByID godoc
// @Summary      Get a vehicle-route assignment by ID
// @Description  Get a vehicle-route assignment by ID
// @Tags         vehicle-routes
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Vehicle Route ID"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /api/vehicle-routes/{id} [get]
func (c *vehicleRouteController) FindByID(ctx *gin.Context) {
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
