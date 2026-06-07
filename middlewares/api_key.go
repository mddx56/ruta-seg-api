package middlewares

import (
	"net/http"
	"os"

	"github.com/Caknoooo/go-gin-clean-starter/modules/user/dto"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/utils"
	"github.com/gin-gonic/gin"
)

func AuthenticateAPIKey() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		apiKey := ctx.GetHeader("X-API-Key")

		if apiKey == "" {
			response := utils.BuildResponseFailed(dto.MESSAGE_FAILED_PROSES_REQUEST, "API key not found", nil)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		expectedAPIKey := os.Getenv("API_KEY")
		if expectedAPIKey == "" {
			response := utils.BuildResponseFailed(dto.MESSAGE_FAILED_PROSES_REQUEST, "API key not configured", nil)
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, response)
			return
		}

		if apiKey != expectedAPIKey {
			response := utils.BuildResponseFailed(dto.MESSAGE_FAILED_PROSES_REQUEST, "Invalid API key", nil)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		ctx.Next()
	}
}
