package controller

import (
	"net/http"

	"github.com/Caknoooo/go-gin-clean-starter/modules/route_fare/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/route_fare/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/samber/do"
)

type RouteFareController interface {
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	ChangeStatus(ctx *gin.Context)
	FindAll(ctx *gin.Context)
	FindByID(ctx *gin.Context)
}

type routeFareController struct {
	service service.RouteFareService
}

func NewRouteFareController(injector *do.Injector) (RouteFareController, error) {
	service := do.MustInvoke[service.RouteFareService](injector)
	return &routeFareController{
		service: service,
	}, nil
}

// CreateRouteFare godoc
// @Summary      Create a new route fare
// @Description  Set the amount charged per completed lap on a route, effective from a given date
// @Tags         route-fares
// @Accept       json
// @Produce      json
// @Param        route_fare  body      dto.RouteFareCreateRequest  true  "Route Fare Create Request"
// @Success      201         {object}  utils.Response
// @Failure      400         {object}  utils.Response
// @Failure      500         {object}  utils.Response
// @Router       /api/route-fares [post]
func (c *routeFareController) Create(ctx *gin.Context) {
	var req dto.RouteFareCreateRequest
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

// UpdateRouteFare godoc
// @Summary      Update an existing route fare
// @Description  Update the amount or effective date of an existing route fare
// @Tags         route-fares
// @Accept       json
// @Produce      json
// @Param        id          path      string                      true  "Route Fare ID"
// @Param        route_fare  body      dto.RouteFareUpdateRequest  true  "Route Fare Update Request"
// @Success      200         {object}  utils.Response
// @Failure      400         {object}  utils.Response
// @Failure      500         {object}  utils.Response
// @Router       /api/route-fares/{id} [put]
func (c *routeFareController) Update(ctx *gin.Context) {
	var req dto.RouteFareUpdateRequest
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

// ChangeRouteFareStatus godoc
// @Summary      Change status of a route fare (soft delete)
// @Description  Toggle the status of a route fare
// @Tags         route-fares
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Route Fare ID"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /api/route-fares/{id}/status [patch]
func (c *routeFareController) ChangeStatus(ctx *gin.Context) {
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

// FindAllRouteFares godoc
// @Summary      List all route fares
// @Description  Get a list of all active route fares
// @Tags         route-fares
// @Accept       json
// @Produce      json
// @Success      200  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /api/route-fares [get]
func (c *routeFareController) FindAll(ctx *gin.Context) {
	res, err := c.service.FindAll(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.BuildResponseFailed(dto.MESSAGE_INTERNAL_SERVER_ERROR, err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS, res))
}

// FindRouteFareByID godoc
// @Summary      Get a route fare by ID
// @Description  Get a route fare by ID
// @Tags         route-fares
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Route Fare ID"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /api/route-fares/{id} [get]
func (c *routeFareController) FindByID(ctx *gin.Context) {
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
