package dto

import (
	"time"
)

type (
	AppVersionRequest struct {
		VersionName          string `json:"version_name" binding:"required"`
		VersionCode          string `json:"version_code" binding:"required,max=10"`
		UrlPlaystore         string `json:"url_playstore" binding:"required,max=500"`
		UrlApplestore        string `json:"url_applestore" binding:"required,max=500"`
		FechaRelease         string `json:"fecha_release" binding:"required"` // Receive as string and parse in service
		MiniSupportedVersion string `json:"mini_supported_version" binding:"required"`
		IsForceUpdate        bool   `json:"is_force_update"`
		Plataform            string `json:"plataform" binding:"required"`
	}

	AppVersionResponse struct {
		AppId                int       `json:"app_id"`
		VersionName          string    `json:"version_name"`
		VersionCode          string    `json:"version_code"`
		UrlPlaystore         string    `json:"url_playstore"`
		UrlApplestore        string    `json:"url_applestore"`
		FechaRelease         time.Time `json:"fecha_release"`
		MiniSupportedVersion string    `json:"mini_supported_version"`
		IsForceUpdate        bool      `json:"is_force_update"`
		Plataform            string    `json:"plataform"`
		CreatedAt            time.Time `json:"created_at"`
		UpdatedAt            time.Time `json:"updated_at"`
	}
)

const (
	MESSAGE_SUCCESS_CREATE_APP_VERSION = "App version creada correctamente"
	MESSAGE_FAILED_CREATE_APP_VERSION  = "Error al crear app version"
	MESSAGE_SUCCESS_GET_APP_VERSION    = "App version obtenida correctamente"
	MESSAGE_FAILED_GET_APP_VERSION     = "Error al obtener app version"
	MESSAGE_SUCCESS_UPDATE_APP_VERSION = "App version actualizada correctamente"
	MESSAGE_FAILED_UPDATE_APP_VERSION  = "Error al actualizar app version"
	MESSAGE_SUCCESS_DELETE_APP_VERSION = "App version eliminada correctamente"
	MESSAGE_FAILED_DELETE_APP_VERSION  = "Error al eliminar app version"
)
