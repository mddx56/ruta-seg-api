package controller

import (
	"net/http"

	"github.com/Caknoooo/go-gin-clean-starter/modules/fine/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/fine/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/samber/do"
)

type FineController interface {
	FindAll(ctx *gin.Context)
	FindAllMine(ctx *gin.Context)
	FindByID(ctx *gin.Context)
	Void(ctx *gin.Context)
	FindAllTypes(ctx *gin.Context)
}

type fineController struct {
	service service.FineService
}

func NewFineController(injector *do.Injector) (FineController, error) {
	service := do.MustInvoke[service.FineService](injector)
	return &fineController{
		service: service,
	}, nil
}

// FindAllFines godoc
// @Summary      List all fines (admin)
// @Description  Get a list of all fines generated for micros (automatic or manual)
// @Tags         fines
// @Accept       json
// @Produce      json
// @Success      200  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /api/fines [get]
func (c *fineController) FindAll(ctx *gin.Context) {
	res, err := c.service.FindAll(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.BuildResponseFailed(dto.MESSAGE_INTERNAL_SERVER_ERROR, err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS, res))
}

// FindAllMineFines godoc
// @Summary      List my fines
// @Description  Get the fines generated for the authenticated user's own vehicles (rol Dueño)
// @Tags         fines
// @Accept       json
// @Produce      json
// @Success      200  {object}  utils.Response
// @Failure      401  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /api/fines/mine [get]
func (c *fineController) FindAllMine(ctx *gin.Context) {
	userIDStr := ctx.MustGet("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
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

// FindFineByID godoc
// @Summary      Get a fine by ID
// @Description  Get a fine by ID, including its fine type
// @Tags         fines
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Fine ID"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /api/fines/{id} [get]
func (c *fineController) FindByID(ctx *gin.Context) {
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

// VoidFine godoc
// @Summary      Void a fine
// @Description  Mark a fine as voided after human review (admin only)
// @Tags         fines
// @Accept       json
// @Produce      json
// @Param        id    path      string               true  "Fine ID"
// @Param        fine  body      dto.FineVoidRequest  false "Void reason"
// @Success      200   {object}  utils.Response
// @Failure      400   {object}  utils.Response
// @Failure      500   {object}  utils.Response
// @Router       /api/fines/{id}/void [patch]
func (c *fineController) Void(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, utils.BuildResponseFailed(dto.MESSAGE_FAILED_INVALID_ID, err.Error(), nil))
		return
	}

	var req dto.FineVoidRequest
	_ = ctx.ShouldBindJSON(&req)

	if err := c.service.Void(ctx.Request.Context(), id, req.Notes); err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.BuildResponseFailed(dto.MESSAGE_INTERNAL_SERVER_ERROR, err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.BuildResponseSuccess(dto.MESSAGE_UPDATED, nil))
}

// FindAllFineTypes godoc
// @Summary      List all fine types
// @Description  Get the catalog of fine types (overtaking, lap time, prolonged stop, speeding)
// @Tags         fines
// @Accept       json
// @Produce      json
// @Success      200  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /api/fine-types [get]
func (c *fineController) FindAllTypes(ctx *gin.Context) {
	res, err := c.service.FindAllTypes(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, utils.BuildResponseFailed(dto.MESSAGE_INTERNAL_SERVER_ERROR, err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS, res))
}
