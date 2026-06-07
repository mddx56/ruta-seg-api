package dto

import (
	"errors"
	"time"

	vehicleDto "github.com/Caknoooo/go-gin-clean-starter/modules/vehicle/dto"
)

const (
	// Failed
	MESSAGE_FAILED_GET_DATA_FROM_BODY = "fallo al obtener data from body"
	MESSAGE_FAILED_CREATE_DEVICE      = "fallo al crear dispositivo"
	MESSAGE_FAILED_GET_LIST_DEVICE    = "fallo al obtener list dispositivo"
	MESSAGE_FAILED_GET_DEVICE         = "fallo al obtener dispositivo"
	MESSAGE_FAILED_UPDATE_DEVICE      = "fallo al actualizar dispositivo"
	MESSAGE_FAILED_DELETE_DEVICE      = "fallo al eliminar dispositivo"
	MESSAGE_FAILED_INVALID_DEVICE_ID  = "imei invalido"
	MESSAGE_FAILED_IMEI_REQUIRED      = "imei es requerido"
	MESSAGE_FAILED_BULK_IMPORT_DEVICE = "fallo en la importación masiva de dispositivos"
	MESSAGE_FAILED_BULK_VALIDATE      = "fallo en la validación masiva de dispositivos"

	// Success
	MESSAGE_SUCCESS_CREATE_DEVICE      = "éxito al crear dispositivo"
	MESSAGE_SUCCESS_GET_LIST_DEVICE    = "éxito al obtener list dispositivo"
	MESSAGE_SUCCESS_GET_DEVICE         = "éxito al obtener dispositivo"
	MESSAGE_SUCCESS_UPDATE_DEVICE      = "éxito al actualizar dispositivo"
	MESSAGE_SUCCESS_DELETE_DEVICE      = "éxito al eliminar dispositivo"
	MESSAGE_SUCCESS_BULK_IMPORT_DEVICE = "importación masiva completada"
	MESSAGE_SUCCESS_BULK_VALIDATE      = "validación masiva completada"
	MESSAGE_SUCCESS_EXPORT_DEVICE      = "exportación de dispositivos completada"
)

var (
	ErrCreateDevice      = errors.New("fallo al crear dispositivo")
	ErrGetDeviceById     = errors.New("fallo al obtener dispositivo por imei")
	ErrUpdateDevice      = errors.New("fallo al actualizar dispositivo")
	ErrDeviceNotFound    = errors.New("dispositivo no encontrado")
	ErrDeleteDevice      = errors.New("fallo al eliminar dispositivo")
	ErrIMEIAlreadyExists = errors.New("imei ya existe")
	ErrGetListDevice     = errors.New("fallo al obtener lista de dispositivos")
)

type (
	DeviceCreateRequest struct {
		IMEI            string  `json:"imei" binding:"required,min=15,max=20"`
		Model           string  `json:"model" binding:"required"`
		Protocol        *string `json:"protocol,omitempty"`
		SimPhoneNumber  *string `json:"sim_phone_number,omitempty"`
		SimICCID        *string `json:"cod_sim,omitempty"`
		SimProvider     *string `json:"sim_provider,omitempty"`
		APNConf         *string `json:"apn_conf,omitempty"`
		FirmwareVersion *string `json:"firmware_version,omitempty"`
		RemoteIP        *string `json:"remote_ip,omitempty"`
		UserID          *string `json:"user_id,omitempty"`
		GroupID         *string `json:"group_id,omitempty"`

		// Auditoría
		UserAuditID     *string `json:"-"`
	}

	GroupInfo struct {
		ID    string               `json:"id"`
		Name  string               `json:"name"`
		Owner *vehicleDto.UserInfo `json:"owner,omitempty"`
	}

	DeviceResponse struct {
		IMEI           string  `json:"imei"`
		Model          string  `json:"model"`
		SimPhoneNumber *string `json:"sim_phone_number,omitempty"`
		SimProvider    *string `json:"sim_provider,omitempty"`
		SimICCID       *string `json:"cod_sim,omitempty"`
		Status         bool    `json:"status"`

		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`

		ActiveVehicle *vehicleDto.VehicleResponse `json:"active_vehicle,omitempty"`
		Groups        []GroupInfo                 `json:"groups,omitempty"`
	}

	DeviceUpdateRequest struct {
		IMEI            string  `json:"imei" binding:"required"` // PK reference
		Model           *string `json:"model,omitempty"`
		Protocol        *string `json:"protocol,omitempty"`
		SimPhoneNumber  *string `json:"sim_phone_number,omitempty"`
		SimICCID        *string `json:"cod_sim,omitempty"`
		SimProvider     *string `json:"sim_provider,omitempty"`
		APNConf         *string `json:"apn_conf,omitempty"`
		FirmwareVersion *string `json:"firmware_version,omitempty"`
		RemoteIP        *string `json:"remote_ip,omitempty"`
		UserID          *string `json:"user_id,omitempty"`
		GroupID         *string `json:"group_id,omitempty"`

		// Auditoría
		UserAuditID     *string `json:"-"`
	}

	DeviceSimpleResponse struct {
		IMEI string `json:"imei"`
	}

	DeviceInstallationItem struct {
		InstallationID string                      `json:"installation_id"`
		InstalledAt    time.Time                   `json:"installed_at"`
		RemovedAt      *time.Time                  `json:"removed_at,omitempty"`
		InstallReason  *string                     `json:"install_reason,omitempty"`
		RemovalReason  *string                     `json:"removal_reason,omitempty"`
		Vehicle        *vehicleDto.VehicleResponse `json:"vehicle,omitempty"`
	}

	DeviceFullResponse struct {
		Device                   DeviceResponse           `json:"device"`
		Installations            []DeviceInstallationItem `json:"installations"`
		AvailableForInstallation bool                     `json:"available_for_installation"`
	}

	// ── Bulk Import DTOs ─────────────────────────────────────────────

	BulkImportItem struct {
		IMEI           string  `json:"imei" binding:"required"`
		CodSim         string  `json:"cod_sim" binding:"required"`
		SimPhoneNumber *string `json:"sim_phone_number,omitempty"`
		SimProvider    *string `json:"sim_provider,omitempty"`
	}

	BulkImportRequest struct {
		Items []BulkImportItem `json:"items" binding:"required,min=1,dive"`

		// Auditoría
		UserAuditID *string `json:"-"`
	}

	BulkImportItemResult struct {
		Row            int      `json:"row"`
		IMEI           string   `json:"imei"`
		CodSim         string   `json:"cod_sim"`
		SimPhoneNumber *string  `json:"sim_phone_number,omitempty"`
		SimProvider    *string  `json:"sim_provider,omitempty"`
		Success        bool     `json:"success"`
		Errors         []string `json:"errors,omitempty"`
	}

	BulkImportResponse struct {
		TotalReceived int                    `json:"total_received"`
		TotalSuccess  int                    `json:"total_success"`
		TotalFailed   int                    `json:"total_failed"`
		Results       []BulkImportItemResult `json:"results"`
	}

	// ── Export DTOs ──────────────────────────────────────────────────

	DeviceExportItem struct {
		IMEI           string  `json:"imei"`
		CodSim         *string `json:"cod_sim"`
		SimPhoneNumber *string `json:"sim_phone_number"`
		SimProvider    *string `json:"sim_provider"`
		Status         string  `json:"status"`
	}

	// ── Categorized Dashboard DTOs ───────────────────────────────────

	CategorizedDevice struct {
		IMEI       string    `json:"imei"`
		Placa      string    `json:"placa"`
		Make       string    `json:"make"`
		Model      string    `json:"model"`
		Color      string    `json:"color"`
		Latitude   float64   `json:"latitude"`
		Longitude  float64   `json:"longitude"`
		Speed      int       `json:"speed"`
		Course     int       `json:"course"`
		DeviceTime time.Time `json:"device_time"`
		Category   string    `json:"category"`
		Battery    float64   `json:"battery"`
		Ignition   *bool     `json:"ignition,omitempty"`
	}

	CategorizedDevicesResponse = []CategorizedDevice
)
