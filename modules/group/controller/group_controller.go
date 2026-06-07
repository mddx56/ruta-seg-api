package controller

import (
	"net/http"

	"github.com/Caknoooo/go-gin-clean-starter/modules/group/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/group/query"
	"github.com/Caknoooo/go-gin-clean-starter/modules/group/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/utils"
	"github.com/Caknoooo/go-pagination"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"gorm.io/gorm"
)

type GroupController interface {
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	ChangeStatus(ctx *gin.Context)
	FindAll(ctx *gin.Context)
	FindAllByUserID(ctx *gin.Context)
	AssignDevice(ctx *gin.Context)
	RemoveDevice(ctx *gin.Context)
}

type groupController struct {
	service service.GroupService
	db      *gorm.DB
}

func NewGroupController(injector *do.Injector) (GroupController, error) {
	service := do.MustInvoke[service.GroupService](injector)
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	return &groupController{
		service: service,
		db:      db,
	}, nil
}

func (c *groupController) Create(ctx *gin.Context) {
	var req dto.GroupCreateRequest
	if err := ctx.ShouldBind(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_CREATE_GROUP, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	// Assuming UserID comes from body as per DTO, but usually it comes from token.
	// However, the DTO has UserID field and it is required.
	// If the user wants to enforce current user, we might overwrite it here.
	// For now, I'll respect the DTO.

	result, err := c.service.Create(ctx.Request.Context(), req)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_CREATE_GROUP, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_CREATE_GROUP, result)
	ctx.JSON(http.StatusOK, res)
}

func (c *groupController) Update(ctx *gin.Context) {
	var req dto.GroupUpdateRequest
	if err := ctx.ShouldBind(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_UPDATE_GROUP, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	// The DTO requires ID in the body.

	result, err := c.service.Update(ctx.Request.Context(), req)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_UPDATE_GROUP, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_UPDATE_GROUP, result)
	ctx.JSON(http.StatusOK, res)
}

func (c  *groupController) ChangeStatus(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, utils.BuildResponseFailed("ID Invalido", "empty", nil))
		return
	}

	if err := c.service.ChangeStatus(ctx.Request.Context(), id); err != nil {
		ctx.JSON(http.StatusBadRequest, utils.BuildResponseFailed("Fallo procesar solicitud", err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.BuildResponseSuccess("estado actualizado correctamente", nil))
}

// FindAll godoc
// @Summary      Get all groups
// @Description  Get a paginated list of all groups
// @Tags         groups
// @Accept       json
// @Produce      json
// @Param        page      query     int     false  "Page number"
// @Param        limit     query     int     false  "Items per page"
// @Param        user_id   query     string  false  "User ID Filter"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Router       /api/group [get]
func (c *groupController) FindAll(ctx *gin.Context) {
	var filter = &query.GroupFilter{}
	filter.BindPagination(ctx)

	if err := ctx.ShouldBindQuery(filter); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_GROUPS, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	// Optional: Force filter by user_id from token if not admin?
	// For now, let's keep it flexible as per User implementation

	groups, total, err := pagination.PaginatedQueryWithIncludable[query.Group](c.db, filter)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_GROUPS, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	paginationResponse := pagination.CalculatePagination(filter.Pagination, total)
	res := pagination.NewPaginatedResponse(http.StatusOK, dto.MESSAGE_SUCCESS_GET_GROUPS, groups, paginationResponse)
	ctx.JSON(http.StatusOK, res)
}

func (c *groupController) FindAllByUserID(ctx *gin.Context) {
	userID := ctx.Param("user_id")
	if userID == "" {
		// Fallback to token user if not provided in param?
		// The service requires explicitly passing userID.
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_GROUPS, "User ID is required", nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.service.FindAllByUserID(ctx.Request.Context(), userID)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_GROUPS, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_GROUPS, result)
	ctx.JSON(http.StatusOK, res)
}

// AssignDevice godoc
// @Summary      Assign a device to a group
// @Description  Assign a device to a specific group
// @Tags         groups
// @Accept       json
// @Produce      json
// @Param        assignment body      dto.GroupAssignDeviceRequest  true  "Assignment Request"
// @Success      200        {object}  utils.Response
// @Failure      400        {object}  utils.Response
// @Router       /api/group/assign [post]
func (c *groupController) AssignDevice(ctx *gin.Context) {
	var req dto.GroupAssignDeviceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_UPDATE_GROUP, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	userID := ctx.MustGet("user_id").(string)
	if err := c.service.AssignDevice(ctx.Request.Context(), req, userID); err != nil {
		res := utils.BuildResponseFailed("Fallo al asignar dispositivo al grupo", err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess("Dispositivo asignado exitosamente", nil)
	ctx.JSON(http.StatusOK, res)
}

// RemoveDevice godoc
// @Summary      Remove a device from a group
// @Description  Remove a device from a specific group
// @Tags         groups
// @Accept       json
// @Produce      json
// @Param        removel body      dto.GroupRemoveDeviceRequest  true  "Remove Request"
// @Success      200     {object}  utils.Response
// @Failure      400     {object}  utils.Response
// @Router       /api/group/remove [post]
func (c *groupController) RemoveDevice(ctx *gin.Context) {
	var req dto.GroupRemoveDeviceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_UPDATE_GROUP, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	if err := c.service.RemoveDevice(ctx.Request.Context(), req); err != nil {
		res := utils.BuildResponseFailed("Fallo al remover dispositivo del grupo", err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess("Dispositivo removido exitosamente", nil)
	ctx.JSON(http.StatusOK, res)
}
