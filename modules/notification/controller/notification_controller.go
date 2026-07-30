package controller

import (
	"net/http"

	"github.com/Caknoooo/go-gin-clean-starter/modules/notification/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/notification/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/samber/do"
)

type NotificationController interface {
	RegisterDeviceToken(ctx *gin.Context)
	UnregisterDeviceToken(ctx *gin.Context)
	FindAllMine(ctx *gin.Context)
	MarkRead(ctx *gin.Context)
}

type notificationController struct {
	service service.NotificationService
}

func NewNotificationController(injector *do.Injector) (NotificationController, error) {
	svc := do.MustInvoke[service.NotificationService](injector)
	return &notificationController{service: svc}, nil
}

func currentUserID(ctx *gin.Context) (uuid.UUID, error) {
	return uuid.Parse(ctx.MustGet("user_id").(string))
}

// RegisterDeviceToken godoc
// @Summary      Register a push notification device token
// @Description  Register (or reassign) the FCM token of the authenticated user's mobile device
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Param        device_token  body      dto.RegisterDeviceTokenRequest  true  "Device Token"
// @Success      201  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /api/notifications/device-tokens [post]
func (c *notificationController) RegisterDeviceToken(ctx *gin.Context) {
	var req dto.RegisterDeviceTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.BuildResponseFailed(dto.MESSAGE_FAILED_BAD_REQUEST, err.Error(), nil))
		return
	}

	userID, err := currentUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, utils.BuildResponseFailed("No autorizado", "user_id inválido en el token", nil))
		return
	}

	if err := c.service.RegisterDeviceToken(ctx.Request.Context(), userID, req.Token, req.Platform); err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.BuildResponseFailed(dto.MESSAGE_INTERNAL_SERVER_ERROR, err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusCreated, utils.BuildResponseSuccess(dto.MESSAGE_CREATED, nil))
}

// UnregisterDeviceToken godoc
// @Summary      Unregister a push notification device token
// @Description  Remove a device token (e.g. on logout) so it stops receiving push notifications
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Param        device_token  body      dto.UnregisterDeviceTokenRequest  true  "Device Token"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /api/notifications/device-tokens [delete]
func (c *notificationController) UnregisterDeviceToken(ctx *gin.Context) {
	var req dto.UnregisterDeviceTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.BuildResponseFailed(dto.MESSAGE_FAILED_BAD_REQUEST, err.Error(), nil))
		return
	}

	if err := c.service.UnregisterDeviceToken(ctx.Request.Context(), req.Token); err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.BuildResponseFailed(dto.MESSAGE_INTERNAL_SERVER_ERROR, err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS, nil))
}

// FindAllMyNotifications godoc
// @Summary      List my notifications
// @Description  Get the authenticated user's notification history
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Success      200  {object}  utils.Response
// @Failure      401  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /api/notifications [get]
func (c *notificationController) FindAllMine(ctx *gin.Context) {
	userID, err := currentUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, utils.BuildResponseFailed("No autorizado", "user_id inválido en el token", nil))
		return
	}

	res, err := c.service.FindAllMine(ctx.Request.Context(), userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.BuildResponseFailed(dto.MESSAGE_INTERNAL_SERVER_ERROR, err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS, res))
}

// MarkNotificationRead godoc
// @Summary      Mark a notification as read
// @Description  Mark one of the authenticated user's notifications as read
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Notification ID"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /api/notifications/{id}/read [patch]
func (c *notificationController) MarkRead(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.BuildResponseFailed(dto.MESSAGE_FAILED_INVALID_ID, err.Error(), nil))
		return
	}

	userID, err := currentUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, utils.BuildResponseFailed("No autorizado", "user_id inválido en el token", nil))
		return
	}

	if err := c.service.MarkRead(ctx.Request.Context(), id, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.BuildResponseFailed(dto.MESSAGE_INTERNAL_SERVER_ERROR, err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.BuildResponseSuccess(dto.MESSAGE_UPDATED, nil))
}
