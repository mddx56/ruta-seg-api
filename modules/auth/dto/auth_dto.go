package dto

import (
	"errors"

	appDto "github.com/Caknoooo/go-gin-clean-starter/modules/app_version/dto"
	userDto "github.com/Caknoooo/go-gin-clean-starter/modules/user/dto"
)

const (
	MESSAGE_FAILED_REFRESH_TOKEN        = "fallo al renovar token"
	MESSAGE_SUCCESS_REFRESH_TOKEN       = "renovacion exitosa de token"
	MESSAGE_FAILED_LOGOUT               = "fallo al cerrar sesion"
	MESSAGE_SUCCESS_LOGOUT              = "cerrar sesion exitoso"
	MESSAGE_FAILED_SEND_PASSWORD_RESET  = "fallo al enviar solicitud de restablecimiento de contraseña"
	MESSAGE_SUCCESS_SEND_PASSWORD_RESET = "envio exitoso de solicitud de restablecimiento de contraseña"
	MESSAGE_FAILED_RESET_PASSWORD       = "fallo al restablecer contraseña"
	MESSAGE_SUCCESS_RESET_PASSWORD      = "restablecimiento exitoso de contraseña"
	MESSAGE_FAILED_GOOGLE_LOGIN         = "fallo el login con Google"
	MESSAGE_SUCCESS_GOOGLE_LOGIN        = "login con Google exitoso"
)

var (
	ErrRefreshTokenNotFound = errors.New("refresh token no encontrado")
	ErrRefreshTokenExpired  = errors.New("refresh token expirado")
	ErrInvalidCredentials   = errors.New("credenciales invalidas")
	ErrPasswordResetToken   = errors.New("token de restablecimiento de contraseña invalido")
	ErrUserBlocked          = errors.New("usuario sin acceso, comuniquese con el administrador")
	ErrUserDisabled         = errors.New("cuenta desactivada, contacte al administrador")
	ErrLoginRateLimited     = errors.New("demasiados intentos fallidos, intente de nuevo en 15 minutos")
)

type (
	RefreshTokenRequest struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	TokenResponse struct {
		AccessToken  string                     `json:"access_token"`
		RefreshToken string                     `json:"refresh_token"`
		User         userDto.UserResponse       `json:"user"`
		App          *appDto.AppVersionResponse `json:"app"`
	}

	SendPasswordResetRequest struct {
		Email string `json:"email" binding:"required,email"`
	}

	ResetPasswordRequest struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=4"`
	}
)
