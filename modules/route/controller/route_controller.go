package controller

import (
	"net/http"
	"strconv"

	"github.com/Caknoooo/go-gin-clean-starter/modules/route/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/route/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/samber/do"
)

type RouteController interface {
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	ChangeStatus(ctx *gin.Context)
	FindAll(ctx *gin.Context)
	FindByID(ctx *gin.Context)
	FindLiveVehicles(ctx *gin.Context)
	FindETA(ctx *gin.Context)
}

type routeController struct {
	service     service.RouteService
	liveService service.RouteLiveService
}

func NewRouteController(injector *do.Injector) (RouteController, error) {
	routeService := do.MustInvoke[service.RouteService](injector)
	liveService := do.MustInvoke[service.RouteLiveService](injector)
	return &routeController{
		service:     routeService,
		liveService: liveService,
	}, nil
}

// CreateRoute godoc
// @Summary      Create a new route
// @Description  Create a new bus route (line), optionally with its stops and GeoJSON geometry
// @Tags         routes
// @Accept       json
// @Produce      json
// @Param        route  body      dto.RouteCreateRequest  true  "Route Create Request"
// @Success      201    {object}  utils.Response
// @Failure      400    {object}  utils.Response
// @Failure      500    {object}  utils.Response
// @Router       /api/routes [post]
func (c *routeController) Create(ctx *gin.Context) {
	var req dto.RouteCreateRequest
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

// UpdateRoute godoc
// @Summary      Update an existing route
// @Description  Update an existing route; if stops is present, it fully replaces the existing stop list
// @Tags         routes
// @Accept       json
// @Produce      json
// @Param        id     path      string                  true  "Route ID"
// @Param        route  body      dto.RouteUpdateRequest  true  "Route Update Request"
// @Success      200    {object}  utils.Response
// @Failure      400    {object}  utils.Response
// @Failure      500    {object}  utils.Response
// @Router       /api/routes/{id} [put]
func (c *routeController) Update(ctx *gin.Context) {
	var req dto.RouteUpdateRequest
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

// ChangeRouteStatus godoc
// @Summary      Change status of a route (soft delete)
// @Description  Toggle the status of a route
// @Tags         routes
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Route ID"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /api/routes/{id}/status [patch]
func (c *routeController) ChangeStatus(ctx *gin.Context) {
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

// FindAllRoutes godoc
// @Summary      List all routes
// @Description  Get a list of all active routes with their stops (public, no auth required)
// @Tags         routes
// @Accept       json
// @Produce      json
// @Success      200  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /api/routes [get]
func (c *routeController) FindAll(ctx *gin.Context) {
	res, err := c.service.FindAll(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.BuildResponseFailed(dto.MESSAGE_INTERNAL_SERVER_ERROR, err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS, res))
}

// FindRouteByID godoc
// @Summary      Get a route by ID
// @Description  Get a route with its stops and geometry by ID (public, no auth required)
// @Tags         routes
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Route ID"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /api/routes/{id} [get]
func (c *routeController) FindByID(ctx *gin.Context) {
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

// FindLiveVehiclesByRoute godoc
// @Summary      List live vehicles on a route
// @Description  Get the micros currently "en ruta" (recent position) for a route, with their pin, position and lap info (public, no auth required)
// @Tags         routes
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Route ID"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /api/routes/{id}/live [get]
func (c *routeController) FindLiveVehicles(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.BuildResponseFailed(dto.MESSAGE_FAILED_INVALID_ID, err.Error(), nil))
		return
	}

	res, err := c.liveService.FindLiveVehicles(ctx.Request.Context(), id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.BuildResponseFailed(dto.MESSAGE_INTERNAL_SERVER_ERROR, err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS, res))
}

// FindETAByRoute godoc
// @Summary      Estimate arrival time of live vehicles to a point
// @Description  Get the estimated time of arrival of each live vehicle on a route to the queried lat/lon, sorted soonest-first (public, no auth required)
// @Tags         routes
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Route ID"
// @Param        lat  query     number  true  "Latitude of the query point"
// @Param        lon  query     number  true  "Longitude of the query point"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /api/routes/{id}/eta [get]
func (c *routeController) FindETA(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.BuildResponseFailed(dto.MESSAGE_FAILED_INVALID_ID, err.Error(), nil))
		return
	}

	lat, err := strconv.ParseFloat(ctx.Query("lat"), 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.BuildResponseFailed(dto.MESSAGE_FAILED_BAD_REQUEST, "parámetro 'lat' inválido o faltante", nil))
		return
	}
	lon, err := strconv.ParseFloat(ctx.Query("lon"), 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.BuildResponseFailed(dto.MESSAGE_FAILED_BAD_REQUEST, "parámetro 'lon' inválido o faltante", nil))
		return
	}

	res, err := c.liveService.FindETA(ctx.Request.Context(), id, lat, lon)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.BuildResponseFailed(dto.MESSAGE_INTERNAL_SERVER_ERROR, err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS, res))
}
