package dto

const (
	// Success Messages
	MESSAGE_SUCCESS_TRIGGER_BACKUP        = "Proceso de backup iniciado correctamente"
	MESSAGE_SUCCESS_TRIGGER_SCHEMA_BACKUP = "Proceso de backup de esquema iniciado correctamente"

	// Failed Messages
	MESSAGE_FAILED_TRIGGER_BACKUP        = "Fallo al iniciar el proceso de backup"
	MESSAGE_FAILED_TRIGGER_SCHEMA_BACKUP = "Fallo al iniciar el proceso de backup de esquema"
	MESSAGE_FAILED_DOWNLOAD_BACKUP       = "Fallo al descargar el archivo de backup"
)

type BackupResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
