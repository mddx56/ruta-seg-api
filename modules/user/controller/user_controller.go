package controller

import (
	"net/http"

	"github.com/Caknoooo/go-gin-clean-starter/modules/user/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/user/query"
	"github.com/Caknoooo/go-gin-clean-starter/modules/user/service"
	"github.com/Caknoooo/go-gin-clean-starter/modules/user/validation"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/utils"
	"github.com/Caknoooo/go-pagination"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"gorm.io/gorm"
)

type (
	UserController interface {
		Me(ctx *gin.Context)
		Create(ctx *gin.Context)
		GetAllUser(ctx *gin.Context)
		GetAllSimple(ctx *gin.Context)
		UpdateMe(ctx *gin.Context)
		ChangeMyPassword(ctx *gin.Context)
		AdminResetPassword(ctx *gin.Context)
		Update(ctx *gin.Context)
		UpdateBlockStatus(ctx *gin.Context)
		ChangeStatus(ctx *gin.Context)
		GetInstalledDevices(ctx *gin.Context)
		GetById(ctx *gin.Context)
	}

	userController struct {
		userService    service.UserService
		userValidation *validation.UserValidation
		db             *gorm.DB
	}
)

func NewUserController(injector *do.Injector, us service.UserService) UserController {
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	userValidation := validation.NewUserValidation()
	return &userController{
		userService:    us,
		userValidation: userValidation,
		db:             db,
	}
}

// GetAllUser godoc
// @Summary      List all users
// @Description  Get a list of all users with pagination and filtering
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        page      query     int     false  "Page number"
// @Param        limit     query     int     false  "Items per page"
// @Param        sort      query     string  false  "Sort field"
// @Param        order     query     string  false  "Sort order (asc, desc)"
// @Param        search    query     string  false  "Search term"
// @Param        includes  query     string  false  "Relations to include (comma separated)"
// @Success      200       {object}  utils.Response
// @Failure      400       {object}  utils.Response
// @Router       /api/user [get]
func (c *userController) GetAllUser(ctx *gin.Context) {
	var filter = &query.UserFilter{}
	filter.BindPagination(ctx)

	ctx.ShouldBindQuery(filter)

	users, total, err := pagination.PaginatedQueryWithIncludable[query.User](c.db, filter)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_USER, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	// Enriquecer roles con literales para UI
	items := make([]dto.UserListItemResponse, 0, len(users))
	for _, u := range users {
		roleLiteral := u.Role
		switch u.Role {
		case constants.ENUM_ROLE_ADMIN:
			roleLiteral = constants.ENUM_ROLE_ADMIN_LITERAL
		case constants.ENUM_ROLE_INSTALLER:
			roleLiteral = constants.ENUM_ROLE_INSTALLER_LITERAL
		case constants.ENUM_ROLE_USER:
			roleLiteral = constants.ENUM_ROLE_USER_LITERAL
		}

		items = append(items, dto.UserListItemResponse{
			ID:          u.ID,
			Name:        u.Name,
			Username:    u.Username,
			Email:       u.Email,
			TelpNumber:  u.TelpNumber,
			Role:        u.Role,
			RoleLiteral: roleLiteral,
			ImageUrl:    u.ImageUrl,
			IsVerified:  u.IsVerified,
			IsBlocked:   u.IsBlocked,
			Status:      u.Status,
			CreatedAt:   u.CreatedAt,
			UpdatedAt:   u.UpdatedAt,
		})
	}

	paginationResponse := pagination.CalculatePagination(filter.Pagination, total)
	response := pagination.NewPaginatedResponse(http.StatusOK, dto.MESSAGE_SUCCESS_GET_LIST_USER, items, paginationResponse)
	ctx.JSON(http.StatusOK, response)
}

// GetAllSimple godoc
// @Summary      List simple users
// @Description  Get a list of users without pagination (only role user)
// @Tags         user
// @Accept       json
// @Produce      json
// @Success      200  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /api/user/simple [get]
func (c *userController) GetAllSimple(ctx *gin.Context) {
	var users []dto.UserSimpleItemResponse
	if err := c.db.
		Table("users").
		Select("id", "name", "username").
		Where("role = ?", constants.ENUM_ROLE_USER).
		Order("name asc").
		Find(&users).Error; err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_LIST_USER, err.Error(), nil)
		ctx.JSON(http.StatusInternalServerError, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_LIST_USER, users)
	ctx.JSON(http.StatusOK, res)
}

// Me godoc
// @Summary      Get current user
// @Description  Get details of the currently logged-in user
// @Tags         user
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Router       /api/user/me [get]
func (c *userController) Me(ctx *gin.Context) {
	userId := ctx.MustGet("user_id").(string)

	result, err := c.userService.GetUserById(ctx.Request.Context(), userId)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_USER, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_USER, result)
	ctx.JSON(http.StatusOK, res)
}

// UpdateMe godoc
// @Summary      Update my profile
// @Description  Update authenticated user data (without password). Validates unique email and username.
// @Tags         user
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        user  body      dto.UserMeUpdateRequest  true  "User Me Update Request"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Router       /api/user/me [put]
func (c *userController) UpdateMe(ctx *gin.Context) {
	userId := ctx.MustGet("user_id").(string)

	var req dto.UserMeUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.userService.UpdateMe(ctx.Request.Context(), userId, req)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_UPDATE_USER, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_UPDATE_USER, result)
	ctx.JSON(http.StatusOK, res)
}

// ChangeMyPassword godoc
// @Summary      Change my password
// @Description  Change authenticated user password (requires current_password)
// @Tags         user
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.UserChangePasswordRequest  true  "Change Password Request"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Router       /api/user/me/password [put]
func (c *userController) ChangeMyPassword(ctx *gin.Context) {
	userId := ctx.MustGet("user_id").(string)

	var req dto.UserChangePasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	if err := c.userService.ChangeMyPassword(ctx.Request.Context(), userId, req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_UPDATE_USER, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess("contraseña actualizada correctamente", nil)
	ctx.JSON(http.StatusOK, res)
}

// AdminResetPassword godoc
// @Summary      Reset user password (admin)
// @Description  Admin resets a user's password by user ID
// @Tags         user
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string                       true  "User ID"
// @Param        body  body      dto.AdminResetPasswordRequest true  "Reset Password Request"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Router       /api/user/{id}/reset-password [put]
func (c *userController) AdminResetPassword(ctx *gin.Context) {
	targetUserId := ctx.Param("id")
	if targetUserId == "" {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_PROSES_REQUEST, "ID de usuario requerido", nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	var req dto.AdminResetPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	if err := c.userService.AdminResetPassword(ctx.Request.Context(), targetUserId, req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_UPDATE_USER, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess("contraseña reseteada correctamente", nil)
	ctx.JSON(http.StatusOK, res)
}

// Create godoc
// @Summary      Create user
// @Description  Create a new user
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        user    body      dto.UserCreateRequest  true  "User Create Request"
// @Success      200     {object}  utils.Response
// @Failure      400     {object}  utils.Response
// @Router       /api/user [post]
func (c *userController) Create(ctx *gin.Context) {
	var req dto.UserCreateRequest
	if err := ctx.ShouldBind(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	if err := c.userValidation.ValidateUserCreateRequest(req); err != nil {
		res := utils.BuildResponseFailed("Fallo en la validación", err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.userService.Create(ctx.Request.Context(), req)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_REGISTER_USER, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_REGISTER_USER, result)
	ctx.JSON(http.StatusOK, res)
}

// Update godoc
// @Summary      Update user
// @Description  Update user details
// @Tags         user
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path      string                 true  "User ID"
// @Param        user    body      dto.UserUpdateRequest  true  "User Update Request"
// @Success      200     {object}  utils.Response
// @Failure      400     {object}  utils.Response
// @Router       /api/user/{id} [put]
func (c *userController) Update(ctx *gin.Context) {
	var req dto.UserUpdateRequest
	if err := ctx.ShouldBind(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	if err := c.userValidation.ValidateUserUpdateRequest(req); err != nil {
		res := utils.BuildResponseFailed("Fallo en la validación", err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	pathId := ctx.Param("id")
	tokenUserId := ctx.MustGet("user_id").(string)

	// Fetch requesting user to check role
	requestingUser, err := c.userService.GetUserById(ctx.Request.Context(), tokenUserId)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_USER, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, res)
		return
	}

	// Permission check: allow if admin OR if updating self
	if requestingUser.Role != constants.ENUM_ROLE_ADMIN && pathId != tokenUserId {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_DENIED_ACCESS, "Solo puedes editar tu propio perfil", nil)
		ctx.AbortWithStatusJSON(http.StatusForbidden, res)
		return
	}

	result, err := c.userService.Update(ctx.Request.Context(), req, pathId)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_UPDATE_USER, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_UPDATE_USER, result)
	ctx.JSON(http.StatusOK, res)
}

// UpdateBlockStatus godoc
// @Summary      Block or unblock user
// @Description  Update user block status
// @Tags         user
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path      string                true  "User ID"
// @Param        status  body      dto.UserBlockRequest  true  "Block Request"
// @Success      200     {object}  utils.Response
// @Failure      400     {object}  utils.Response
// @Router       /api/user/{id}/block [patch]
func (c *userController) UpdateBlockStatus(ctx *gin.Context) {
	var req dto.UserBlockRequest
	if err := ctx.ShouldBind(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}
	if req.IsBlocked == nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, "is_blocked es requerido", nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	userId := ctx.Param("id")
	// TODO: verify if sender is admin. Middleware or service logic?
	// The requirement: "solo rele tipo admin puede realizar esa accion"
	// We can check the role from the context user (set by auth middleware).

	// Assuming AUTH middleware puts user role or we fetch it.
	// ME endpoint fetches user. But context usually has user_id.
	// We might need to check role here or rely on specific middleware.

	// Let's implement basics first. Role check typically done via middleware logic or here.
	// Fetch current user role:
	currentUserID := ctx.MustGet("user_id").(string)
	currentUser, err := c.userService.GetUserById(ctx.Request.Context(), currentUserID)
	if err != nil {
		res := utils.BuildResponseFailed("Failed to verify permissions", err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, res)
		return
	}

	if currentUser.Role != constants.ENUM_ROLE_ADMIN {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_DENIED_ACCESS, "Solo el administrador puede realizar esta acción", nil)
		ctx.AbortWithStatusJSON(http.StatusForbidden, res)
		return
	}

	if err := c.userService.UpdateBlockStatus(ctx.Request.Context(), userId, *req.IsBlocked); err != nil {
		res := utils.BuildResponseFailed("Fallo al cambiar estado del usuario", err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess("Usuario bloqueado correctamente", nil)
	ctx.JSON(http.StatusOK, res)
}

// ChangeStatus godoc
// @Summary      Change user status (soft delete / restore)
// @Description  Toggle user status (true to false, false to true). Admin only.
// @Tags         user
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path      string  true  "User ID"
// @Success      200     {object}  utils.Response
// @Failure      400     {object}  utils.Response
// @Router       /api/user/{id}/status [patch]
func (c *userController) ChangeStatus(ctx *gin.Context) {
	userId := ctx.Param("id")

	// Verify permissions (admin)
	currentUserID := ctx.MustGet("user_id").(string)
	currentUser, err := c.userService.GetUserById(ctx.Request.Context(), currentUserID)
	if err != nil {
		res := utils.BuildResponseFailed("Fallo al verificar permisos", err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, res)
		return
	}

	if currentUser.Role != constants.ENUM_ROLE_ADMIN {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_DENIED_ACCESS, "Solo el administrador puede realizar esta acción", nil)
		ctx.AbortWithStatusJSON(http.StatusForbidden, res)
		return
	}

	if err := c.userService.ChangeStatus(ctx.Request.Context(), userId); err != nil {
		res := utils.BuildResponseFailed("Fallo al cambiar estado del usuario", err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess("Usuario deshabilitado correctamente", nil)
	ctx.JSON(http.StatusOK, res)
}

// GetInstalledDevices godoc
// @Summary      Get installed devices
// @Description  Get list of installed devices for a specific user
// @Tags         user
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Router       /api/user/{id}/devices [get]
func (c *userController) GetInstalledDevices(ctx *gin.Context) {
	userId := ctx.Param("id")

	if userId == "" {
		res := utils.BuildResponseFailed("fallo al obtener dispositivos", "ID de usuario requerido", nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	tokenUserId := ctx.MustGet("user_id").(string)
	requestingUser, err := c.userService.GetUserById(ctx.Request.Context(), tokenUserId)
	if err != nil {
		res := utils.BuildResponseFailed("fallo al obtener permisos", err.Error(), nil)
		ctx.JSON(http.StatusInternalServerError, res)
		return
	}
	isAdmin := requestingUser.Role == constants.ENUM_ROLE_ADMIN

	devices, err := c.userService.GetInstalledDevices(ctx.Request.Context(), userId, isAdmin)
	if err != nil {
		res := utils.BuildResponseFailed("fallo al obtener dispositivos", err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess("lista de dispositivos obtenida correctamente", devices)
	ctx.JSON(http.StatusOK, res)
}

// GetById godoc
// @Summary      Get user by ID
// @Description  Get user details by user ID
// @Tags         user
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Router       /api/user/{id} [get]
func (c *userController) GetById(ctx *gin.Context) {
	userId := ctx.Param("id")

	if userId == "" {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_USER, "ID de usuario requerido", nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.userService.GetUserById(ctx.Request.Context(), userId)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_USER, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_USER, result)
	ctx.JSON(http.StatusOK, res)
}
