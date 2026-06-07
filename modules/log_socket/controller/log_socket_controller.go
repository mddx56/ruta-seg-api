package controller

import (
	"net/http"

	"github.com/Caknoooo/go-gin-clean-starter/modules/log_socket/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/log_socket/query"
	"github.com/Caknoooo/go-gin-clean-starter/modules/log_socket/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/utils"
	"github.com/Caknoooo/go-pagination"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"gorm.io/gorm"
)

type (
	LogSocketController interface {
		GetAll(ctx *gin.Context)
		Create(ctx *gin.Context)
	}

	logSocketController struct {
		logSocketService service.LogSocketService
		db               *gorm.DB
	}
)

func NewLogSocketController(injector *do.Injector, ls service.LogSocketService) LogSocketController {
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	return &logSocketController{
		logSocketService: ls,
		db:               db,
	}
}

// GetAll godoc
// @Summary      List all log sockets
// @Description  Get a list of all log sockets with pagination and filtering by date
// @Tags         log-socket
// @Accept       json
// @Produce      json
// @Param        page      query     int     false  "Page number"
// @Param        limit     query     int     false  "Items per page"
// @Param        date      query     string  false  "Date (YYYY-MM-DD)"
// @Success      200       {object}  utils.Response
// @Failure      400       {object}  utils.Response
// @Router       /api/log-socket [get]
func (c *logSocketController) GetAll(ctx *gin.Context) {
	var filter = &query.LogSocketFilter{}
	filter.BindPagination(ctx)

	ctx.ShouldBindQuery(filter)

	logSockets, total, err := pagination.PaginatedQueryWithIncludable[query.LogSocket](c.db, filter)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_LOG_SOCKET, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	paginationResponse := pagination.CalculatePagination(filter.Pagination, total)
	response := pagination.NewPaginatedResponse(http.StatusOK, dto.MESSAGE_SUCCESS_GET_LOG_SOCKET, logSockets, paginationResponse)
	ctx.JSON(http.StatusOK, response)
}

// Create godoc
// @Summary      Create a new log socket
// @Description  Create a new log socket entry
// @Tags         log-socket
// @Accept       json
// @Produce      json
// @Param        request   body      dto.LogSocketCreateRequest  true  "Log socket data"
// @Success      201       {object}  utils.Response
// @Failure      400       {object}  utils.Response
// @Failure      500       {object}  utils.Response
// @Router       /api/log-socket [post]
func (c *logSocketController) Create(ctx *gin.Context) {
	var req dto.LogSocketCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_CREATE_LOG_SOCKET, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	logSocket, err := c.logSocketService.Create(ctx.Request.Context(), req)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_CREATE_LOG_SOCKET, err.Error(), nil)
		ctx.JSON(http.StatusInternalServerError, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_CREATE_LOG_SOCKET, logSocket)
	ctx.JSON(http.StatusCreated, res)
}
