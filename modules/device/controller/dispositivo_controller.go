package controller

import (
	"fmt"
	"net/http"

	"github.com/Caknoooo/go-gin-clean-starter/modules/device/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/device/query"
	"github.com/Caknoooo/go-gin-clean-starter/modules/device/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/utils"
	"github.com/Caknoooo/go-pagination"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"gorm.io/gorm"
)

type DeviceController interface {
	CreateDevice(ctx *gin.Context)
	UpdateDevice(ctx *gin.Context)
	ChangeStatus(ctx *gin.Context)
	GetAllDevices(ctx *gin.Context)
	GetSimpleDevices(ctx *gin.Context)
	GetDeviceByIMEI(ctx *gin.Context)
	GetDeviceFullByIMEI(ctx *gin.Context)
	GetDevicesPaginated(ctx *gin.Context)
	BulkValidateDevices(ctx *gin.Context)
	BulkImportDevices(ctx *gin.Context)
	ExportDevices(ctx *gin.Context)
	GetCategorizedDevices(ctx *gin.Context)
}

type deviceController struct {
	service service.DeviceService
	db      *gorm.DB
}

func NewDeviceController(injector *do.Injector) (DeviceController, error) {
	service := do.MustInvoke[service.DeviceService](injector)
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	return &deviceController{
		service: service,
		db:      db,
	}, nil
}

// CreateDevice godoc
// @Summary      Create a new device
// @Description  Create a new device with the input payload
// @Tags         devices
// @Accept       json
// @Produce      json
// @Param        device  body      dto.DeviceCreateRequest  true  "Device Create Request"
// @Success      201     {object}  utils.Response
// @Failure      400     {object}  utils.Response
// @Router       /api/devices [post]
func (c *deviceController) CreateDevice(ctx *gin.Context) {
	var req dto.DeviceCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	if userID, exists := ctx.Get("user_id"); exists {
		if uidStr, ok := userID.(string); ok {
			req.UserAuditID = &uidStr
		}
	}

	result, err := c.service.Create(ctx.Request.Context(), req)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_CREATE_DEVICE, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_CREATE_DEVICE, result)
	ctx.JSON(http.StatusCreated, res)
}

// UpdateDevice godoc
// @Summary      Update device
// @Description  Update device details
// @Tags         devices
// @Accept       json
// @Produce      json
// @Param        device  body      dto.DeviceUpdateRequest  true  "Device Update Request"
// @Success      200     {object}  utils.Response
// @Failure      400     {object}  utils.Response
// @Router       /api/devices [put]
func (c *deviceController) UpdateDevice(ctx *gin.Context) {
	var req dto.DeviceUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	if userID, exists := ctx.Get("user_id"); exists {
		if uidStr, ok := userID.(string); ok {
			req.UserAuditID = &uidStr
		}
	}

	result, err := c.service.Update(ctx.Request.Context(), req)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_UPDATE_DEVICE, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_UPDATE_DEVICE, result)
	ctx.JSON(http.StatusOK, res)
}

// DeleteDevice godoc
// @Summary      Change status device
// @Description  Change status a device by IMEI
// @Tags         devices
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Device IMEI"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Router       /api/devices/{id}/status [patch]
func (c *deviceController) ChangeStatus(ctx *gin.Context) {
	// Ahora el ID es el IMEI (string)
	imei := ctx.Param("id")
	if imei == "" {
		res := utils.BuildResponseFailed("Invalid ID", "IMEI cannot be empty", nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	if err := c.service.ChangeStatus(ctx.Request.Context(), imei); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_DELETE_DEVICE, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_DELETE_DEVICE, nil)
	ctx.JSON(http.StatusOK, res)
}

// GetAllDevices godoc
// @Summary      Get all devices
// @Description  Get a list of all devices
// @Tags         devices
// @Accept       json
// @Produce      json
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Router       /api/devices [get]
func (c *deviceController) GetAllDevices(ctx *gin.Context) {
	result, err := c.service.FindAll(ctx.Request.Context())
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_LIST_DEVICE, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_LIST_DEVICE, result)
	ctx.JSON(http.StatusOK, res)
}

// GetSimpleDevices godoc
// @Summary      Get simple list of devices
// @Description  Get a simple list of devices (IMEI and Name)
// @Tags         devices
// @Accept       json
// @Produce      json
// @Param        available query     boolean false "Filter active devices (no active installations)"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Router       /api/devices/simple [get]
func (c *deviceController) GetSimpleDevices(ctx *gin.Context) {
	available := ctx.Query("available") == "true"
	result, err := c.service.GetSimple(ctx.Request.Context(), available)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_LIST_DEVICE, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_LIST_DEVICE, result)
	ctx.JSON(http.StatusOK, res)
}

// GetDeviceByIMEI godoc
// @Summary      Get device by IMEI
// @Description  Get a device by its IMEI
// @Tags         devices
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Device IMEI"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Router       /api/devices/{id} [get]
func (c *deviceController) GetDeviceByIMEI(ctx *gin.Context) {
	imei := ctx.Param("id")
	if imei == "" {
		res := utils.BuildResponseFailed("Invalid ID", "IMEI cannot be empty", nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.service.FindByIMEI(ctx.Request.Context(), imei)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DEVICE, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_DEVICE, result)
	ctx.JSON(http.StatusOK, res)
}

// GetDeviceFullByIMEI godoc
// @Summary      Get device by IMEI (full)
// @Description  Get a device by its IMEI with vehicle and installations info
// @Tags         devices
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Device IMEI"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Router       /api/devices/{id}/full [get]
func (c *deviceController) GetDeviceFullByIMEI(ctx *gin.Context) {
	imei := ctx.Param("id")
	if imei == "" {
		res := utils.BuildResponseFailed("Invalid ID", "IMEI cannot be empty", nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.service.FindByIMEIFull(ctx.Request.Context(), imei)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DEVICE, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_DEVICE, result)
	ctx.JSON(http.StatusOK, res)
}

// GetDevicesPaginated godoc
// @Summary      Get list of devices with pagination
// @Description  Get a list of devices with pagination and filters
// @Tags         devices
// @Accept       json
// @Produce      json
// @Param        page      query     int     false  "Page number"
// @Param        limit     query     int     false  "Items per page"
// @Param        sort      query     string  false  "Sort field"
// @Param        order     query     string  false  "Sort order (asc, desc)"
// @Param        search    query     string  false  "Search term"
// @Success      200       {object}  utils.Response
// @Failure      400       {object}  utils.Response
// @Router       /api/devices/list [get]
func (c *deviceController) GetDevicesPaginated(ctx *gin.Context) {
	var filter = &query.DeviceFilter{}
	filter.BindPagination(ctx)

	// Bind query params to filter struct
	if err := ctx.ShouldBindQuery(filter); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_LIST_DEVICE, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	devices, total, err := pagination.PaginatedQueryWithIncludable[query.DeviceQuery](c.db, filter)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_LIST_DEVICE, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	paginationResponse := pagination.CalculatePagination(filter.Pagination, total)
	response := pagination.NewPaginatedResponse(http.StatusOK, dto.MESSAGE_SUCCESS_GET_LIST_DEVICE, devices, paginationResponse)
	ctx.JSON(http.StatusOK, response)
}

// BulkValidateDevices godoc
// @Summary      Pre-validate bulk device import
// @Description  Validates a list of devices without persisting. Returns per-item validation results.
// @Tags         devices
// @Accept       json
// @Produce      json
// @Param        devices  body      dto.BulkImportRequest  true  "Bulk Import Items"
// @Success      200      {object}  utils.Response
// @Failure      400      {object}  utils.Response
// @Router       /api/devices/bulk/validate [post]
func (c *deviceController) BulkValidateDevices(ctx *gin.Context) {
	var req dto.BulkImportRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	result := c.service.BulkValidate(ctx.Request.Context(), req)
	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_BULK_VALIDATE, result)
	ctx.JSON(http.StatusOK, res)
}

// BulkImportDevices godoc
// @Summary      Bulk import devices
// @Description  Import multiple devices at once. Returns detailed per-item results with errors.
// @Tags         devices
// @Accept       json
// @Produce      json
// @Param        devices  body      dto.BulkImportRequest  true  "Bulk Import Items"
// @Success      200      {object}  utils.Response
// @Failure      400      {object}  utils.Response
// @Router       /api/devices/bulk/import [post]
func (c *deviceController) BulkImportDevices(ctx *gin.Context) {
	var req dto.BulkImportRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	if userID, exists := ctx.Get("user_id"); exists {
		if uidStr, ok := userID.(string); ok {
			req.UserAuditID = &uidStr
		}
	}

	result := c.service.BulkImport(ctx.Request.Context(), req)

	httpStatus := http.StatusOK
	if result.TotalFailed > 0 && result.TotalSuccess == 0 {
		httpStatus = http.StatusBadRequest
	}

	var res utils.Response
	if result.TotalFailed > 0 {
		res = utils.BuildResponseFailed(dto.MESSAGE_FAILED_BULK_IMPORT_DEVICE, 
			fmt.Sprintf("Algunos dispositivos no pudieron ser importados, revise el detalle (%d de %d con errores)", result.TotalFailed, result.TotalReceived), result)
	} else {
		res = utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_BULK_IMPORT_DEVICE, result)
	}
	ctx.JSON(httpStatus, res)
}

// ExportDevices godoc
// @Summary      Export devices for Excel
// @Description  Returns a list of devices with basic columns for Excel export. By default only active devices. Use ?all=true to include disabled.
// @Tags         devices
// @Produce      json
// @Param        all  query     boolean  false  "Include disabled devices"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Router       /api/devices/export [get]
func (c *deviceController) ExportDevices(ctx *gin.Context) {
	includeDisabled := ctx.Query("all") == "true"

	result, err := c.service.Export(ctx.Request.Context(), includeDisabled)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_LIST_DEVICE, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_EXPORT_DEVICE, result)
	ctx.JSON(http.StatusOK, res)
}

// GetCategorizedDevices godoc
// @Summary      Get devices categorized statistics
// @Description  Get a list of devices categorized by their connection status (TODOS, EN VIVO, DETENIDOS, SIN CONEXION) for the current user and admin.
// @Tags         devices
// @Accept       json
// @Produce      json
// @Success      200       {object}  utils.Response
// @Failure      400       {object}  utils.Response
// @Router       /api/devices/categories [get]
func (c *deviceController) GetCategorizedDevices(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		res := utils.BuildResponseFailed("No autorizado", "Usuario no encontrado en el token", nil)
		ctx.JSON(http.StatusUnauthorized, res)
		return
	}

	role, exists := ctx.Get("role")
	isAdmin := false
	if exists {
		if roleStr, ok := role.(string); ok && roleStr == "admin" {
			isAdmin = true
		}
	}

	uidStr := userID.(string)

	result, err := c.service.GetCategorizedDevices(ctx.Request.Context(), uidStr, isAdmin)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_LIST_DEVICE, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_LIST_DEVICE, result)
	ctx.JSON(http.StatusOK, res)
}

