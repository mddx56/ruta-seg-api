package controller

import (
	"net/http"

	deviceDto "github.com/Caknoooo/go-gin-clean-starter/modules/device/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/device_installation/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/device_installation/query"
	"github.com/Caknoooo/go-gin-clean-starter/modules/device_installation/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/utils"
	"github.com/Caknoooo/go-pagination"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/samber/do"
	"gorm.io/gorm"
)

type DeviceInstallationController interface {
	CreateInstallation(ctx *gin.Context)
	QuickCreateInstallation(ctx *gin.Context)
	GetMyInstallations(ctx *gin.Context)
	GetAllInstallations(ctx *gin.Context)
	GetInstallationsByIMEI(ctx *gin.Context)
	GetInstallationsByVehicleID(ctx *gin.Context)
	UninstallInstallation(ctx *gin.Context)
}

type deviceInstallationController struct {
	service service.DeviceInstallationService
	db      *gorm.DB
}

func NewDeviceInstallationController(injector *do.Injector) (DeviceInstallationController, error) {
	service := do.MustInvoke[service.DeviceInstallationService](injector)
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	return &deviceInstallationController{
		service: service,
		db:      db,
	}, nil
}

// CreateInstallation godoc
// @Summary      Create a new device installation
// @Description  Create a new device-vehicle installation
// @Tags         device-installations
// @Accept       json
// @Produce      json
// @Param        installation  body      dto.DeviceInstallationCreateRequest  true  "Installation Request"
// @Success      201           {object}  utils.Response
// @Failure      400           {object}  utils.Response
// @Router       /api/device-installations [post]
func (c *deviceInstallationController) CreateInstallation(ctx *gin.Context) {
	var req dto.DeviceInstallationCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed("Failed to bind request", err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	userIDStr := ctx.MustGet("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		res := utils.BuildResponseFailed("Invalid user ID in token", err.Error(), nil)
		ctx.JSON(http.StatusUnauthorized, res)
		return
	}

	result, err := c.service.Create(ctx.Request.Context(), req, userID)
	if err != nil {
		res := utils.BuildResponseFailed("Fallo al crear instalacion", err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess("Instalacion creada exitosamente", result)
	ctx.JSON(http.StatusCreated, res)
}

// QuickCreateInstallation godoc
// @Summary      Create a new device installation with IMEI and vehicle chassis
// @Description  Create a new device-vehicle installation using IMEI and vehicle chassis (for mobile quick registration)
// @Tags         device-installations
// @Accept       json
// @Produce      json
// @Param        installation  body      dto.DeviceInstallationQuickCreateRequest  true  "Quick Installation Request"
// @Success      201           {object}  utils.Response
// @Failure      400           {object}  utils.Response
// @Router       /api/device-installations/quick [post]
func (c *deviceInstallationController) QuickCreateInstallation(ctx *gin.Context) {
	var req dto.DeviceInstallationQuickCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed("Failed to bind request", err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	userIDStr := ctx.MustGet("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		res := utils.BuildResponseFailed("Invalid user ID in token", err.Error(), nil)
		ctx.JSON(http.StatusUnauthorized, res)
		return
	}

	result, err := c.service.QuickCreate(ctx.Request.Context(), req, userID)
	if err != nil {
		res := utils.BuildResponseFailed("Fallo al crear instalacion rapida", err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess("Instalacion rapida creada exitosamente", result)
	ctx.JSON(http.StatusCreated, res)
}

// GetMyInstallations godoc
// @Summary      Get my device installations (mobile)
// @Description  Get active installations for vehicles owned by the authenticated user (mobile-friendly detail response)
// @Tags         device-installations
// @Accept       json
// @Produce      json
// @Success      200  {object}  utils.Response
// @Failure      401  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /api/device-installations/mine [get]
func (c *deviceInstallationController) GetMyInstallations(ctx *gin.Context) {
	userIDStr := ctx.MustGet("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		res := utils.BuildResponseFailed("Invalid user ID in token", err.Error(), nil)
		ctx.JSON(http.StatusUnauthorized, res)
		return
	}

	result, err := c.service.GetMine(ctx.Request.Context(), userID)
	if err != nil {
		res := utils.BuildResponseFailed("Fallo al obtener mis instalaciones", err.Error(), nil)
		ctx.JSON(http.StatusInternalServerError, res)
		return
	}

	res := utils.BuildResponseSuccess("Mis instalaciones obtenidas exitosamente", result)
	ctx.JSON(http.StatusOK, res)
}

// GetAllInstallations godoc
// @Summary      Get all device installations
// @Description  Get a paginated list of all device-vehicle installations
// @Tags         device-installations
// @Accept       json
// @Produce      json
// @Param        page      query     int     false  "Page number"
// @Param        limit     query     int     false  "Items per page"
// @Param        sort      query     string  false  "Sort field"
// @Param        order     query     string  false  "Sort order (asc, desc)"
// @Param        search    query     string  false  "Search term"
// @Param        includes  query     string  false  "Relations to include (comma separated)"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Router       /api/device-installations [get]
func (c *deviceInstallationController) GetAllInstallations(ctx *gin.Context) {
	var filter = &query.DeviceInstallationFilter{}
	filter.BindPagination(ctx)

	if err := ctx.ShouldBindQuery(filter); err != nil {
		res := utils.BuildResponseFailed("Fallo al bindear parametros de paginacion", err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	// Forzing includes for output mapping
	filter.Includes = append(filter.Includes, "Vehicle", "UserCreation")

	installations, total, err := pagination.PaginatedQueryWithIncludable[query.DeviceInstallation](c.db, filter)
	if err != nil {
		res := utils.BuildResponseFailed("Fallo al obtener instalaciones", err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	for i := range installations {
		if installations[i].Vehicle != nil {
			installations[i].Chassis = installations[i].Vehicle.Chassis
		}

		if installations[i].UserCreation != nil {
			installations[i].CreatedBy = &query.CreatedBy{
				ID:   installations[i].UserCreation.ID,
				Name: installations[i].UserCreation.Name,
			}
		}
	}

	paginationResponse := pagination.CalculatePagination(filter.Pagination, total)
	res := pagination.NewPaginatedResponse(http.StatusOK, "Instalaciones obtenidas exitosamente", installations, paginationResponse)
	ctx.JSON(http.StatusOK, res)
}

// GetInstallationsByIMEI godoc
// @Summary      Get installations by IMEI
// @Description  Get all vehicle installations for a specific device (IMEI)
// @Tags         device-installations
// @Accept       json
// @Produce      json
// @Param        imei  query     string  true  "IMEI"
// @Success      200        {object}  utils.Response
// @Failure      400        {object}  utils.Response
// @Router       /api/device-installations/device [get]
func (c *deviceInstallationController) GetInstallationsByIMEI(ctx *gin.Context) {
	imei := ctx.Query("imei")
	if imei == "" {
		res := utils.BuildResponseFailed(deviceDto.MESSAGE_FAILED_INVALID_DEVICE_ID, deviceDto.MESSAGE_FAILED_IMEI_REQUIRED, nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.service.GetByIMEI(ctx.Request.Context(), imei)
	if err != nil {
		res := utils.BuildResponseFailed("Fallo al obtener instalaciones", err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess("Instalaciones obtenidas exitosamente", result)
	ctx.JSON(http.StatusOK, res)
}

// GetInstallationsByVehicleID godoc
// @Summary      Get installations by vehicle ID
// @Description  Get all device installations for a specific vehicle
// @Tags         device-installations
// @Accept       json
// @Produce      json
// @Param        vehicle_id  query     string  true  "Vehicle ID"
// @Success      200         {object}  utils.Response
// @Failure      400         {object}  utils.Response
// @Router       /api/device-installations/vehicle [get]
func (c *deviceInstallationController) GetInstallationsByVehicleID(ctx *gin.Context) {
	vehicleID, err := uuid.Parse(ctx.Query("vehicle_id"))
	if err != nil {
		res := utils.BuildResponseFailed("ID de vehiculo invalido", err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.service.GetByVehicleID(ctx.Request.Context(), vehicleID)
	if err != nil {
		res := utils.BuildResponseFailed("Fallo al obtener instalaciones", err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess("Instalaciones obtenidas exitosamente", result)
	ctx.JSON(http.StatusOK, res)
}

// UninstallInstallation godoc
// @Summary      Uninstall a device
// @Description  End a device installation by setting the removed_at timestamp
// @Tags         device-installations
// @Accept       json
// @Produce      json
// @Param        id            path      string                                   true  "Installation ID"
// @Param        uninstallation body     dto.DeviceInstallationUninstallRequest  true  "Uninstall Request"
// @Success      200           {object}  utils.Response
// @Failure      400           {object}  utils.Response
// @Router       /api/device-installations/{id}/uninstall [put]
func (c *deviceInstallationController) UninstallInstallation(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		res := utils.BuildResponseFailed("ID de instalacion invalido", err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	var req dto.DeviceInstallationUninstallRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed("Fallo al bindear request", err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.service.Uninstall(ctx.Request.Context(), id, req)
	if err != nil {
		res := utils.BuildResponseFailed("Fallo al desinstalar dispositivo", err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess("Dispositivo desinstalado exitosamente", result)
	ctx.JSON(http.StatusOK, res)
}
