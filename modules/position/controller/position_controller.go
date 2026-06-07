package controller

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	deviceDto "github.com/Caknoooo/go-gin-clean-starter/modules/device/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/position/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/position/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/utils"
	providerWS "github.com/Caknoooo/go-gin-clean-starter/providers/websocket"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"gorm.io/gorm"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
)

type PositionController interface {
	CreatePosition(ctx *gin.Context)
	GetPositionByID(ctx *gin.Context)
	GetPositionsByIMEI(ctx *gin.Context)
	GetLastPosition(ctx *gin.Context)
	GetLastPositionsOfAllDevices(ctx *gin.Context)
	GetCoordinatesByIMEIAndDate(ctx *gin.Context)
	GetDeviceHistory(ctx *gin.Context)
	GetDeviceRoute(ctx *gin.Context)
	GetPositionsWithVehicleInfoByIMEIAndDate(ctx *gin.Context)
	// DeletePosition(ctx *gin.Context)
}

type positionController struct {
	positionService service.PositionService
	wsService       providerWS.WebsocketService
	db              *gorm.DB
}

func NewPositionController(injector *do.Injector, ps service.PositionService) PositionController {
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	// WS is optional: if not registered yet, don't crash
	wsSvc, _ := do.Invoke[providerWS.WebsocketService](injector)
	return &positionController{
		positionService: ps,
		wsService:       wsSvc,
		db:              db,
	}
}

// CreatePosition godoc
// @Summary      Create a new position
// @Description  Create a new position record for a device
// @Tags         positions
// @Accept       json
// @Produce      json
// @Param        position  body      dto.PositionCreateRequest  true  "Position Create Request"
// @Success      201       {object}  utils.Response
// @Failure      400       {object}  utils.Response
// @Router       /api/positions [post]
func (c *positionController) CreatePosition(ctx *gin.Context) {
	var req dto.PositionCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.positionService.Create(ctx.Request.Context(), req)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_CREATE_POSITION, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	// Broadcast to relevant WebSocket clients and admins (non-blocking)
	if c.wsService != nil {
		var userIDs []string
		
		// Obtener directamente solo a los dueños del vehículo donde está el GPS instalado
		c.db.WithContext(ctx.Request.Context()).
			Table("device_installations").
			Select("vehicles.user_id").
			Joins("JOIN vehicles ON vehicles.id = device_installations.vehicle_id").
			Where("device_installations.imei = ? AND device_installations.removed_at IS NULL AND device_installations.status = ?", req.Imei, true).
			Pluck("vehicles.user_id", &userIDs)

		parsedAttrs := extractBroadcastAttributes(req.Attributes)
		go c.wsService.BroadcastPosition(userIDs, providerWS.DevicePositionData{
			IMEI:       result.Imei,
			Latitude:   result.Latitude,
			Longitude:  result.Longitude,
			Speed:      result.Speed,
			Course:     result.Course,
			DeviceTime: result.DeviceTime,
			ServerTime: result.ServerTime,
			Battery:    parsedAttrs.battery,
			Ignition:   parsedAttrs.ignition,
			Satellites: parsedAttrs.satellites,
		})
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_CREATE_POSITION, result)
	ctx.JSON(http.StatusCreated, res)
}

type broadcastAttrs struct {
	battery    *float64
	ignition   *bool
	satellites *int
}

func extractBroadcastAttributes(raw *string) broadcastAttrs {
	if raw == nil {
		return broadcastAttrs{}
	}
	attrs := struct {
		Battery    *float64 `json:"battery"`
		Ignition   *bool    `json:"ignition"`
		Satellites *int     `json:"satellites"`
	}{}
	_ = json.Unmarshal([]byte(*raw), &attrs)
	return broadcastAttrs{
		battery:    attrs.Battery,
		ignition:   attrs.Ignition,
		satellites: attrs.Satellites,
	}
}

// GetPositionByID godoc
// @Summary      Get position by ID
// @Description  Get a position by its ID
// @Tags         positions
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Position ID"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Router       /api/positions/{id} [get]
func (c *positionController) GetPositionByID(ctx *gin.Context) {
	// Position ID sigue siendo uint64 en la entidad, no cambió a string.
	// La relación DeviceID es la que cambió.
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		res := utils.BuildResponseFailed("Invalid ID", err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.positionService.GetByID(ctx.Request.Context(), id)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_POSITION, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_POSITION, result)
	ctx.JSON(http.StatusOK, res)
}

// GetPositionsByIMEI godoc
// @Summary      Get positions by IMEI
// @Description  Get all positions for a specific device defined by IMEI
// @Tags         positions
// @Accept       json
// @Produce      json
// @Param        imei  query     string  true  "IMEI"
// @Success      200        {object}  utils.Response
// @Failure      400        {object}  utils.Response
// @Router       /api/positions/device [get]
func (c *positionController) GetPositionsByIMEI(ctx *gin.Context) {
	// Ahora imei es lo que usamos
	imei := ctx.Query("imei")
	if imei == "" {
		res := utils.BuildResponseFailed(deviceDto.MESSAGE_FAILED_INVALID_DEVICE_ID, deviceDto.MESSAGE_FAILED_IMEI_REQUIRED, nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.positionService.GetByIMEI(ctx.Request.Context(), imei)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_LIST_POSITION, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_LIST_POSITION, result)
	ctx.JSON(http.StatusOK, res)
}

// GetLastPosition godoc
// @Summary      Get last position by IMEI
// @Description  Get the last position for a specific device defined by IMEI
// @Tags         positions
// @Accept       json
// @Produce      json
// @Param        imei       query     string  true  "IMEI"
// @Success      200        {object}  utils.Response
// @Failure      400        {object}  utils.Response
// @Router       /api/positions/last [get]
func (c *positionController) GetLastPosition(ctx *gin.Context) {
	vehicleID := ctx.Query("vehicle_id")
	if vehicleID != "" {
		result, err := c.positionService.GetLastPositionOfVehicle(ctx.Request.Context(), vehicleID)
		if err != nil {
			res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_POSITION, err.Error(), nil)
			ctx.JSON(http.StatusBadRequest, res)
			return
		}
		res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_POSITION, result)
		ctx.JSON(http.StatusOK, res)
		return
	}

	imei := ctx.Query("imei")
	if imei == "" {
		res := utils.BuildResponseFailed(deviceDto.MESSAGE_FAILED_INVALID_DEVICE_ID, "vehicle_id or imei is required", nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.positionService.GetLastByIMEI(ctx.Request.Context(), imei)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_POSITION, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_POSITION, result)
	ctx.JSON(http.StatusOK, res)
}

// GetLastPositionsOfAllDevices godoc
// @Summary      Get last position of all devices
// @Description  Get the last position for all devices
// @Tags         positions
// @Accept       json
// @Produce      json
// @Success      200        {object}  utils.Response
// @Failure      400        {object}  utils.Response
// @Router       /api/positions/latest [get]
func (c *positionController) GetLastPositionsOfAllDevices(ctx *gin.Context) {
	result, err := c.positionService.GetLastPositionsOfAllVehicles(ctx.Request.Context())
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_LIST_POSITION, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_LIST_POSITION, result)
	ctx.JSON(http.StatusOK, res)
}

// GetCoordinatesByIMEIAndDate godoc
// @Summary      Get position coordinates by IMEI and date
// @Description  Get position coordinates (latitude, longitude) for a specific device and date for frontend mapping
// @Tags         positions
// @Accept       json
// @Produce      json
// @Param        imei       query     string  true   "IMEI"
// @Param        date       query     string  false  "Date (YYYY-MM-DD format, default: today)"
// @Success      200        {object}  utils.Response
// @Failure      400        {object}  utils.Response
// @Router       /api/positions/coordinates [get]
func (c *positionController) GetCoordinatesByIMEIAndDate(ctx *gin.Context) {
	imei := ctx.Query("imei")
	if imei == "" {
		res := utils.BuildResponseFailed(deviceDto.MESSAGE_FAILED_INVALID_DEVICE_ID, deviceDto.MESSAGE_FAILED_IMEI_REQUIRED, nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	// Parse date parameter, default to today
	dateStr := ctx.Query("date")
	var err error
	var date time.Time
	if dateStr == "" {
		date = utils.NowLocal()
	} else {
		date, err = utils.ParseLocalDate(dateStr)
		if err != nil {
			res := utils.BuildResponseFailed("Invalid date format, use YYYY-MM-DD", err.Error(), nil)
			ctx.JSON(http.StatusBadRequest, res)
			return
		}
	}

	result, err := c.positionService.GetCoordinatesByIMEIAndDate(ctx.Request.Context(), imei, date)
	if err != nil {
		res := utils.BuildResponseFailed("Failed to get position coordinates", err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess("Position coordinates retrieved successfully", result)
	ctx.JSON(http.StatusOK, res)
}

// GetDeviceHistory godoc
// @Summary      Get device history
// @Description  Get a detailed history timeline for a device and date (trips, stops, gaps)
// @Tags         positions
// @Accept       json
// @Produce      json
// @Param        imei       query     string  true   "Device IMEI"
// @Param        date       query     string  true   "Date (YYYY-MM-DD)"
// @Success      200        {object}  utils.Response
// @Failure      400        {object}  utils.Response
// @Router       /api/positions/history [get]
func (c *positionController) GetDeviceHistory(ctx *gin.Context) {
	dateStr := ctx.Query("date")
	if dateStr == "" {
		res := utils.BuildResponseFailed("fecha invalida", "fecha es requerida (YYYY-MM-DD)", nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	// Validate date format cheaply
	if _, err := utils.ParseLocalDate(dateStr); err != nil {
		res := utils.BuildResponseFailed("Formato invalido de fecha", "Use YYYY-MM-DD", nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	vehicleID := ctx.Query("vehicle_id")
	if vehicleID != "" {
		result, err := c.positionService.GetVehicleHistory(ctx.Request.Context(), vehicleID, dateStr)
		if err != nil {
			res := utils.BuildResponseFailed("Fallo al obtener historial del vehiculo", err.Error(), nil)
			ctx.JSON(http.StatusBadRequest, res)
			return
		}
		res := utils.BuildResponseSuccess("Historial del vehiculo obtenido exitosamente", result)
		ctx.JSON(http.StatusOK, res)
		return
	}

	imei := ctx.Query("imei")
	if imei == "" {
		res := utils.BuildResponseFailed(deviceDto.MESSAGE_FAILED_INVALID_DEVICE_ID, "vehicle_id or imei is required", nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.positionService.GetDeviceHistory(ctx.Request.Context(), imei, dateStr)
	if err != nil {
		res := utils.BuildResponseFailed("Fallo al obtener historial del dispositivo", err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess("Historial del dispositivo obtenido exitosamente", result)
	ctx.JSON(http.StatusOK, res)
}

// GetDeviceRoute godoc
// @Summary      Get device route history (Replay)
// @Description  Get a detailed history timeline for a device within a specific time range
// @Tags         positions
// @Accept       json
// @Produce      json
// @Param        imei       query     string  true   "Device IMEI"
// @Param        start_time query     string  true   "Start Time (YYYY-MM-DD HH:mm:ss)"
// @Param        end_time   query     string  true   "End Time (YYYY-MM-DD HH:mm:ss)"
// @Success      200        {object}  utils.Response
// @Failure      400        {object}  utils.Response
// @Router       /api/positions/route [get]
func (c *positionController) GetDeviceRoute(ctx *gin.Context) {
	startTimeStr := ctx.Query("start_time")
	endTimeStr := ctx.Query("end_time")

	if startTimeStr == "" || endTimeStr == "" {
		res := utils.BuildResponseFailed("parametros invalidos", "start_time y end_time son requeridos", nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	startTime, err := utils.ParseLocalDateTime(startTimeStr)
	if err != nil {
		res := utils.BuildResponseFailed("formato invalido start_time", "Use YYYY-MM-DD HH:mm:ss", nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	endTime, err := utils.ParseLocalDateTime(endTimeStr)
	if err != nil {
		res := utils.BuildResponseFailed("formato invalido end_time", "Use YYYY-MM-DD HH:mm:ss", nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	if endTime.Before(startTime) {
		res := utils.BuildResponseFailed("rango invalido", "end_time debe ser despues de start_time", nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	vehicleID := ctx.Query("vehicle_id")
	if vehicleID != "" {
		result, err := c.positionService.GetVehicleRoute(ctx.Request.Context(), vehicleID, startTime, endTime)
		if err != nil {
			res := utils.BuildResponseFailed("Fallo al obtener ruta del vehiculo", err.Error(), nil)
			ctx.JSON(http.StatusBadRequest, res)
			return
		}
		res := utils.BuildResponseSuccess("Ruta del vehiculo obtenida exitosamente", result)
		ctx.JSON(http.StatusOK, res)
		return
	}

	imei := ctx.Query("imei")
	if imei == "" {
		res := utils.BuildResponseFailed(deviceDto.MESSAGE_FAILED_INVALID_DEVICE_ID, "vehicle_id or imei is required", nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.positionService.GetDeviceRoute(ctx.Request.Context(), imei, startTime, endTime)
	if err != nil {
		res := utils.BuildResponseFailed("Fallo al obtener ruta del dispositivo", err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess("Dispositivo ruta obtenido exitosamente", result)
	ctx.JSON(http.StatusOK, res)
}

// GetPositionsWithVehicleInfoByIMEIAndDate godoc
// @Summary      Get positions with vehicle info by IMEI and date
// @Description  Get positions and vehicle details for admin
// @Tags         positions
// @Accept       json
// @Produce      json
// @Param        imei       query     string  true   "IMEI"
// @Param        date       query     string  true   "Date (YYYY-MM-DD)"
// @Success      200        {object}  utils.Response
// @Failure      400        {object}  utils.Response
// @Router       /api/positions/device-details [get]
func (c *positionController) GetPositionsWithVehicleInfoByIMEIAndDate(ctx *gin.Context) {
	imei := ctx.Query("imei")
	if imei == "" {
		res := utils.BuildResponseFailed(deviceDto.MESSAGE_FAILED_INVALID_DEVICE_ID, deviceDto.MESSAGE_FAILED_IMEI_REQUIRED, nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	dateStr := ctx.Query("date")
	if dateStr == "" {
		res := utils.BuildResponseFailed("Fecha requerida", "date is required (YYYY-MM-DD)", nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	date, err := utils.ParseLocalDate(dateStr)
	if err != nil {
		res := utils.BuildResponseFailed("Formato de fecha invalido", "Use YYYY-MM-DD", nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	positions, err := c.positionService.GetCoordinatesByIMEIAndDate(ctx.Request.Context(), imei, date)
	if err != nil {
		res := utils.BuildResponseFailed("Failed to get positions", err.Error(), nil)
		ctx.JSON(http.StatusInternalServerError, res)
		return
	}

	var installation entities.DeviceInstallation
	err = c.db.WithContext(ctx.Request.Context()).
		Where("imei = ? AND removed_at IS NULL AND status = ?", imei, true).
		Preload("Vehicle").
		Preload("Vehicle.User").
		Preload("Vehicle.Model").
		Preload("Vehicle.Model.Make").
		Preload("Vehicle.Model.VehicleType").
		First(&installation).Error

	var vehicleInfo dto.DeliveryVehicleInfo

	if err == nil && installation.Vehicle != nil {
		v := installation.Vehicle
		vehicleInfo.Placa = v.Placa
		vehicleInfo.InstalledAt = installation.InstalledAt

		if v.User != nil {
			vehicleInfo.OwnerName = v.User.Name
		}
		if v.Model != nil {
			vehicleInfo.Model = v.Model.ModelName
			if v.Model.Make != nil {
				vehicleInfo.Brand = v.Model.Make.MakeName
			}
			if v.Model.VehicleType != nil {
				vehicleInfo.Type = v.Model.VehicleType.TypeName
			}
		}
	}

	result := dto.PositionsWithVehicleInfoResponse{
		Positions: positions,
		Vehicle:   vehicleInfo,
	}

	res := utils.BuildResponseSuccess("Posiciones y vehiculo obtenidos exitosamente", result)
	ctx.JSON(http.StatusOK, res)
}

// DeletePosition godoc
// // @Summary      Delete position
// // @Description  Delete a position by ID
// // @Tags         positions
// // @Accept       json
// // @Produce      json
// // @Param        id   path      int  true  "Position ID"
// // @Success      200  {object}  utils.Response
// // @Failure      400  {object}  utils.Response
// // @Router       /api/positions/{id} [delete]
// func (c *positionController) DeletePosition(ctx *gin.Context) {
// 	idStr := ctx.Param("id")
// 	id, err := strconv.ParseUint(idStr, 10, 64)
// 	if err != nil {
// 		res := utils.BuildResponseFailed("Invalid ID", err.Error(), nil)
// 		ctx.JSON(http.StatusBadRequest, res)
// 		return
// 	}

// 	if err := c.positionService.Delete(ctx.Request.Context(), id); err != nil {
// 		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_DELETE_POSITION, err.Error(), nil)
// 		ctx.JSON(http.StatusBadRequest, res)
// 		return
// 	}

// 	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_DELETE_POSITION, nil)
// 	ctx.JSON(http.StatusOK, res)
// }
