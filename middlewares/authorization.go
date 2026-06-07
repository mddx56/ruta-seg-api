package middlewares

import (
	"net/http"

	"github.com/Caknoooo/go-gin-clean-starter/modules/auth/service"
	"github.com/Caknoooo/go-gin-clean-starter/modules/user/dto"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/utils"
	"github.com/gin-gonic/gin"
)

// AuthorizeAdmin middleware verifies that the authenticated user has admin role
func AuthorizeAdmin(jwtService service.JWTService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, exists := ctx.Get("token")
		if !exists {
			response := utils.BuildResponseFailed(dto.MESSAGE_FAILED_PROSES_REQUEST, "Token no encontrado en el contexto", nil)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		tokenStr, ok := token.(string)
		if !ok {
			response := utils.BuildResponseFailed(dto.MESSAGE_FAILED_PROSES_REQUEST, "Token inválido", nil)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		role, err := jwtService.GetRoleByToken(tokenStr)
		if err != nil {
			response := utils.BuildResponseFailed(dto.MESSAGE_FAILED_PROSES_REQUEST, "Error al obtener rol del token", nil)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		if role != constants.ENUM_ROLE_ADMIN {
			response := utils.BuildResponseFailed(dto.MESSAGE_FAILED_PROSES_REQUEST, "Acceso denegado: se requiere rol de administrador", nil)
			ctx.AbortWithStatusJSON(http.StatusForbidden, response)
			return
		}

		ctx.Next()
	}
}

// AuthorizeInstallerOrAdmin middleware verifies that the authenticated user has installer or admin role
func AuthorizeInstallerOrAdmin(jwtService service.JWTService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, exists := ctx.Get("token")
		if !exists {
			response := utils.BuildResponseFailed(dto.MESSAGE_FAILED_PROSES_REQUEST, "Token no encontrado en el contexto", nil)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		tokenStr, ok := token.(string)
		if !ok {
			response := utils.BuildResponseFailed(dto.MESSAGE_FAILED_PROSES_REQUEST, "Token inválido", nil)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		role, err := jwtService.GetRoleByToken(tokenStr)
		if err != nil {
			response := utils.BuildResponseFailed(dto.MESSAGE_FAILED_PROSES_REQUEST, "Error al obtener rol del token", nil)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		if role != constants.ENUM_ROLE_ADMIN && role != constants.ENUM_ROLE_INSTALLER {
			response := utils.BuildResponseFailed(dto.MESSAGE_FAILED_PROSES_REQUEST, "Acceso denegado: se requiere rol de instalador o administrador", nil)
			ctx.AbortWithStatusJSON(http.StatusForbidden, response)
			return
		}

		ctx.Next()
	}
}
