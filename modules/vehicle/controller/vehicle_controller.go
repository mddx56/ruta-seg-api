package controller

import (
	"fmt"
	"net/http"

	"github.com/Caknoooo/go-gin-clean-starter/modules/vehicle/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/vehicle/query"
	"github.com/Caknoooo/go-gin-clean-starter/modules/vehicle/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/utils"
	"github.com/Caknoooo/go-pagination"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/samber/do"
	"gorm.io/gorm"
)

type VehicleController interface {
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	ChangeStatus(ctx *gin.Context)
	FindAll(ctx *gin.Context)
	FindByID(ctx *gin.Context)
	GetMyVehicles(ctx *gin.Context)
	GetSimple(ctx *gin.Context)
	FindByChassisFull(ctx *gin.Context)
}

type vehicleController struct {
	service service.VehicleService
	db      *gorm.DB
}

func NewVehicleController(injector *do.Injector) (VehicleController, error) {
	service := do.MustInvoke[service.VehicleService](injector)
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	return &vehicleController{
		service: service,
		db:      db,
	}, nil
}

// CreateVehicle godoc
// @Summary      Create a new vehicle
// @Description  Create a new vehicle with the input payload
// @Tags         vehicles
// @Accept       json
// @Produce      json
// @Param        vehicle  body      dto.VehicleCreateRequest  true  "Vehicle Create Request"
// @Success      201      {object}  utils.Response
// @Failure      400      {object}  utils.Response
// @Failure      500      {object}  utils.Response
// @Router       /api/vehicles [post]
func (c *vehicleController) Create(ctx *gin.Context) {
	var req dto.VehicleCreateRequest
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

// UpdateVehicle godoc
// @Summary      Update an existing vehicle
// @Description  Update an existing vehicle with the input payload
// @Tags         vehicles
// @Accept       json
// @Produce      json
// @Param        id       path      string                    true  "Vehicle ID"
// @Param        vehicle  body      dto.VehicleUpdateRequest  true  "Vehicle Update Request"
// @Success      200      {object}  utils.Response
// @Failure      400      {object}  utils.Response
// @Failure      500      {object}  utils.Response
// @Router       /api/vehicles/{id} [put]
func (c *vehicleController) Update(ctx *gin.Context) {
	var req dto.VehicleUpdateRequest
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

// DeleteVehicle godoc
// @Summary      Change status a vehicle (soft delete)
// @Description  Soft delete a vehicle by setting status to false
// @Tags         vehicles
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Vehicle ID"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /api/vehicles/{id}/status [patch]
func (c  *vehicleController) ChangeStatus(ctx *gin.Context) {
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

// FindAllVehicles godoc
// @Summary      List all vehicles
// @Description  Get a list of all active vehicles with their user and model information
// @Tags         vehicles
// @Accept       json
// @Produce      json
// @Param        page      query     int     false  "Page number"
// @Param        limit     query     int     false  "Items per page"
// @Param        sort      query     string  false  "Sort field"
// @Param        order     query     string  false  "Sort order (asc, desc)"
// @Param        search    query     string  false  "Search term"
// @Param        includes  query     string  false  "Relations to include (comma separated)"
// @Success      200  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /api/vehicles [get]
func (c *vehicleController) FindAll(ctx *gin.Context) {
	var filter = &query.VehicleFilter{}
	filter.BindPagination(ctx)

	if err := ctx.ShouldBindQuery(filter); err != nil {
		res := utils.BuildResponseFailed("Fallo al bindear parametros de paginacion", err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	vehicles, total, err := pagination.PaginatedQueryWithIncludable[query.Vehicle](c.db, filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.BuildResponseFailed(dto.MESSAGE_INTERNAL_SERVER_ERROR, err.Error(), nil))
		return
	}

	var responses []dto.VehicleResponse
	for _, v := range vehicles {
		responses = append(responses, v.ToResponse())
	}

	paginationResponse := pagination.CalculatePagination(filter.Pagination, total)
	res := pagination.NewPaginatedResponse(http.StatusOK, dto.MESSAGE_SUCCESS, responses, paginationResponse)
	ctx.JSON(http.StatusOK, res)
}

// FindVehicleByID godoc
// @Summary      Get a vehicle by ID
// @Description  Get a vehicle by ID with its user and model information
// @Tags         vehicles
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Vehicle ID"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /api/vehicles/{id} [get]
func (c *vehicleController) GetMyVehicles(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, utils.BuildResponseFailed("No autorizado", "user_id no encontrado en token", nil))
		return
	}

	id, err := uuid.Parse(fmt.Sprintf("%v", userID))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.BuildResponseFailed(dto.MESSAGE_FAILED_INVALID_ID, err.Error(), nil))
		return
	}

	res, err := c.service.FindByUserID(ctx.Request.Context(), id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.BuildResponseFailed(dto.MESSAGE_INTERNAL_SERVER_ERROR, err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS, res))
}

func (c *vehicleController) FindByID(ctx *gin.Context) {
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

// GetSimpleVehicles godoc
// @Summary      Get a simple list of vehicles
// @Description  Get a list of vehicles with only ID and Placa
// @Tags         vehicles
// @Accept       json
// @Produce      json
// @Param        available query     boolean false "Filter active vehicles (no active installations)"
// @Success      200  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /api/vehicles/simple [get]
func (c *vehicleController) GetSimple(ctx *gin.Context) {
	available := ctx.Query("available") == "true"
	res, err := c.service.GetSimple(ctx.Request.Context(), available)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.BuildResponseFailed(dto.MESSAGE_INTERNAL_SERVER_ERROR, err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS, res))
}

// FindVehicleByChassisFull godoc
// @Summary      Get a vehicle by chassis (full)
// @Description  Get a vehicle by chassis with full information (make, type, model, owner, installation and device)
// @Tags         vehicles
// @Accept       json
// @Produce      json
// @Param        chassis  path      string  true  "Vehicle chassis"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /api/vehicles/by-chassis/{chassis} [get]
func (c *vehicleController) FindByChassisFull(ctx *gin.Context) {
	chassis := ctx.Param("chassis")
	if chassis == "" {
		ctx.JSON(http.StatusBadRequest, utils.BuildResponseFailed(dto.MESSAGE_FAILED_BAD_REQUEST, "chasis requerido", nil))
		return
	}

	res, err := c.service.FindByChassisFull(ctx.Request.Context(), chassis)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.BuildResponseFailed(dto.MESSAGE_INTERNAL_SERVER_ERROR, err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS, res))
}
