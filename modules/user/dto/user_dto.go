package dto

import (
	"errors"
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

const (
	// Failed
	MESSAGE_FAILED_GET_DATA_FROM_BODY = "fallo obtener datos del body"
	MESSAGE_FAILED_REGISTER_USER      = "fallo crear usuario"
	MESSAGE_FAILED_GET_LIST_USER      = "fallo obtener lista de usuarios"
	MESSAGE_FAILED_TOKEN_NOT_VALID    = "token no valido"
	MESSAGE_FAILED_TOKEN_NOT_FOUND    = "token no encontrado"
	MESSAGE_FAILED_GET_USER           = "fallo obtener usuario"
	MESSAGE_FAILED_LOGIN              = "fallo iniciar sesion"
	MESSAGE_FAILED_UPDATE_USER        = "fallo actualizar usuario"
	MESSAGE_FAILED_DELETE_USER        = "fallo eliminar usuario"
	MESSAGE_FAILED_PROSES_REQUEST     = "fallo procesar solicitud"
	MESSAGE_FAILED_DENIED_ACCESS      = "acceso denegado"
	MESSAGE_FAILED_VERIFY_EMAIL       = "fallo verificar email"

	// Success
	MESSAGE_SUCCESS_REGISTER_USER           = "usuario creado correctamente"
	MESSAGE_SUCCESS_GET_LIST_USER           = "lista de usuarios obtenida correctamente"
	MESSAGE_SUCCESS_GET_USER                = "usuario obtenido correctamente"
	MESSAGE_SUCCESS_LOGIN                   = "sesion iniciada correctamente"
	MESSAGE_SUCCESS_UPDATE_USER             = "usuario actualizado correctamente"
	MESSAGE_SUCCESS_DELETE_USER             = "usuario eliminado correctamente"
	MESSAGE_SEND_VERIFICATION_EMAIL_SUCCESS = "verificacion de email enviada correctamente"
	MESSAGE_SUCCESS_VERIFY_EMAIL            = "email verificado correctamente"
)

var (
	ErrCreateUser             = errors.New("fallo crear usuario")
	ErrGetUserById            = errors.New("fallo obtener usuario por id")
	ErrGetUserByEmail         = errors.New("fallo obtener usuario por email")
	ErrEmailAlreadyExists     = errors.New("email ya existe")
	ErrUpdateUser             = errors.New("fallo actualizar usuario")
	ErrUserNotFound           = errors.New("usuario no encontrado")
	ErrEmailNotFound          = errors.New("email no encontrado")
	ErrDeleteUser             = errors.New("fallo eliminar usuario")
	ErrTokenInvalid           = errors.New("token invalid")
	ErrTokenExpired           = errors.New("token expirado")
	ErrAccountAlreadyVerified = errors.New("cuenta ya verificada")
	ErrUsernameAlreadyExists  = errors.New("username ya existe")
	ErrCurrentPasswordInvalid = errors.New("contraseña actual incorrecta")
)

type (
	UserCreateRequest struct {
		Name       string                `json:"name" form:"name" binding:"required,min=2,max=100"`
		Username   *string               `json:"username" form:"username" binding:"omitempty,min=4,max=20"`
		TelpNumber string                `json:"telp_number" form:"telp_number" binding:"omitempty,min=7,max=20"`
		Email      string                `json:"email" form:"email" binding:"required,email"`
		Password   string                `json:"password" form:"password" binding:"required,min=4"`
		Image      *multipart.FileHeader `json:"image" form:"image" swaggerignore:"true"`
	}

	UserResponse struct {
		ID         string  `json:"id"`
		Name       string  `json:"name"`
		Username   *string `json:"username"`
		Email      string  `json:"email"`
		TelpNumber string  `json:"telp_number"`
		Role       string  `json:"role"`
		RoleLiteral string `json:"role_literal"`
		ImageUrl   string  `json:"image_url"`
		IsVerified bool    `json:"is_verified"`
		IsBlocked  bool    `json:"is_blocked"`
		Status     bool    `json:"status"`
	}
	UserUpdateRequest struct {
		Name       string `json:"name" form:"name" binding:"omitempty,min=2,max=100"`
		TelpNumber string `json:"telp_number" form:"telp_number" binding:"omitempty,min=7,max=20"`
		Email      string `json:"email" form:"email" binding:"omitempty,email"`
	}

	// UserMeUpdateRequest permite actualizar perfil sin contraseña.
	UserMeUpdateRequest struct {
		Name       string  `json:"name" binding:"omitempty,min=2,max=100"`
		Username   *string `json:"username" binding:"omitempty,min=4,max=20"`
		TelpNumber string  `json:"telp_number" binding:"omitempty,min=7,max=20"`
		Email      string  `json:"email" binding:"omitempty,email"`
		ImageUrl   *string `json:"image_url" binding:"omitempty"`
	}

	UserChangePasswordRequest struct {
		CurrentPassword string `json:"current_password" binding:"required,min=4"`
		NewPassword     string `json:"new_password" binding:"required,min=4"`
	}

	AdminResetPasswordRequest struct {
		NewPassword string `json:"new_password" binding:"required,min=4"`
	}

	UserUpdateResponse struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		TelpNumber string `json:"telp_number"`
		Role       string `json:"role"`
		RoleLiteral string `json:"role_literal"`
		Email      string `json:"email"`
		IsVerified bool   `json:"is_verified"`
	}

	SendVerificationEmailRequest struct {
		Email string `json:"email" form:"email" binding:"required"`
	}

	VerifyEmailRequest struct {
		Token string `json:"token" form:"token" binding:"required"`
	}

	VerifyEmailResponse struct {
		Email      string `json:"email"`
		IsVerified bool   `json:"is_verified"`
	}

	UserLoginRequest struct {
		Email    *string `json:"email" form:"email" binding:"omitempty"`
		Username *string `json:"username" form:"username" binding:"omitempty"`
		Password string  `json:"password" form:"password" binding:"required"`
	}

	UserBlockRequest struct {
		// pointer para permitir false (desbloquear) y detectar ausencia del campo
		IsBlocked *bool `json:"is_blocked" binding:"required"`
	}

	UserInstalledDeviceResponse struct {
		InstallationID uuid.UUID   `json:"installation_id"`
		Device         DeviceInfo  `json:"device"`
		Vehicle        VehicleInfo `json:"vehicle"`
		InstalledAt    time.Time   `json:"installed_at"`
		Status         bool        `json:"status"`
	}

	DeviceInfo struct {
		IMEI       string `json:"imei"`
		DeviceType string `json:"device_type,omitempty"`
	}

	VehicleInfo struct {
		ID          uuid.UUID `json:"id"`
		Placa       string    `json:"placa"`
		Description *string   `json:"description"`
		Year        *int      `json:"year"`
		Model       string    `json:"model"`
		Make        string    `json:"make"`
		Color       *string   `json:"color,omitempty"`
	}

	UserDevicesAndGroupsResponse struct {
		InstalledDevices []UserInstalledDeviceResponse `json:"installed_devices"`
		DeviceGroups     []GroupDeviceInfo             `json:"device_groups"`
	}

	GroupDeviceInfo struct {
		ID          uuid.UUID                     `json:"id"`
		Name        string                        `json:"name"`
		Description *string                       `json:"description,omitempty"`
		Devices     []UserInstalledDeviceResponse `json:"devices"`
		CreatedAt   time.Time                     `json:"created_at"`
	}

	UserListItemResponse struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Username    string `json:"username"`
		Email       string `json:"email"`
		TelpNumber  string `json:"telp_number"`
		Role        string `json:"role"`
		RoleLiteral string `json:"role_literal"`
		ImageUrl    string `json:"image_url"`
		IsVerified  bool   `json:"is_verified"`
		IsBlocked   bool   `json:"is_blocked"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		Status      bool      `json:"status"`
	}

	UserSimpleItemResponse struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Username string `json:"username"`
	}
)
