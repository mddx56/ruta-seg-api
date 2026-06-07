package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Caknoooo/go-gin-clean-starter/modules/backup/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/backup/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

type BackupController interface {
	TriggerBackup(ctx *gin.Context)
	TriggerSchemaBackup(ctx *gin.Context)
	DownloadBackup(ctx *gin.Context)
	DownloadSchemaBackup(ctx *gin.Context)
	HandleUnifiedBackup(ctx *gin.Context)
}

type backupController struct {
	backupService service.BackupService
}

func NewBackupController(injector *do.Injector) (BackupController, error) {
	backupService := do.MustInvoke[service.BackupService](injector)
	return &backupController{
		backupService: backupService,
	}, nil
}

// TriggerBackup godoc
// @Summary      Iniciar Backup Completo
// @Description  Inicia el proceso de respaldo completo de la base de datos en el servidor externo.
// @Tags         backup
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dto.BackupResponse "Proceso iniciado correctamente"
// @Failure      401  {object}  utils.Response     "No autorizado"
// @Failure      403  {object}  utils.Response     "Acceso denegado (Solo Admin)"
// @Failure      500  {object}  utils.Response     "Error interno del servidor"
// @Router       /api/backup/trigger [post]
func (c *backupController) TriggerBackup(ctx *gin.Context) {
	result, err := c.backupService.TriggerBackup(ctx.Request.Context())
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_TRIGGER_BACKUP, c.parseExternalError(err), nil)
		ctx.JSON(http.StatusInternalServerError, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_TRIGGER_BACKUP, result)
	ctx.JSON(http.StatusOK, res)
}

// TriggerSchemaBackup godoc
// @Summary      Iniciar Backup del Esquema
// @Description  Inicia el proceso de respaldo *solo del esquema* (sin datos) en el servidor externo.
// @Tags         backup
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dto.BackupResponse "Proceso iniciado correctamente"
// @Failure      401  {object}  utils.Response     "No autorizado"
// @Failure      403  {object}  utils.Response     "Acceso denegado (Solo Admin)"
// @Failure      500  {object}  utils.Response     "Error interno del servidor"
// @Router       /api/backup/schema/trigger [post]
func (c *backupController) TriggerSchemaBackup(ctx *gin.Context) {
	result, err := c.backupService.TriggerSchemaBackup(ctx.Request.Context())
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_TRIGGER_SCHEMA_BACKUP, c.parseExternalError(err), nil)
		ctx.JSON(http.StatusInternalServerError, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_TRIGGER_SCHEMA_BACKUP, result)
	ctx.JSON(http.StatusOK, res)
}

// DownloadBackup godoc
// @Summary      Descargar Backup Completo
// @Description  Descarga el archivo de respaldo completo (.sql) más reciente desde el servidor externo.
// @Tags         backup
// @Produce      application/octet-stream
// @Security     BearerAuth
// @Success      200  {file}    file               "Archivo SQL de respaldo"
// @Failure      401  {object}  utils.Response     "No autorizado"
// @Failure      403  {object}  utils.Response     "Acceso denegado (Solo Admin)"
// @Failure      500  {object}  utils.Response     "Error interno o al descargar"
// @Router       /api/backup/download [get]
func (c *backupController) DownloadBackup(ctx *gin.Context) {
	reader, filename, contentLength, err := c.backupService.DownloadBackup(ctx.Request.Context())
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_DOWNLOAD_BACKUP, c.parseExternalError(err), nil)
		ctx.JSON(http.StatusInternalServerError, res)
		return
	}
	defer reader.Close()

	c.streamDownload(ctx, reader, filename, contentLength)
}

// DownloadSchemaBackup godoc
// @Summary      Descargar Backup del Esquema
// @Description  Descarga el archivo de respaldo del esquema (.sql) más reciente desde el servidor externo.
// @Tags         backup
// @Produce      application/octet-stream
// @Security     BearerAuth
// @Success      200  {file}    file               "Archivo SQL de esquema"
// @Failure      401  {object}  utils.Response     "No autorizado"
// @Failure      403  {object}  utils.Response     "Acceso denegado (Solo Admin)"
// @Failure      500  {object}  utils.Response     "Error interno o al descargar"
// @Router       /api/backup/schema/download [get]
func (c *backupController) DownloadSchemaBackup(ctx *gin.Context) {
	reader, filename, contentLength, err := c.backupService.DownloadSchemaBackup(ctx.Request.Context())
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_DOWNLOAD_BACKUP, c.parseExternalError(err), nil) // Reusing message or define specific one
		ctx.JSON(http.StatusInternalServerError, res)
		return
	}
	defer reader.Close()

	c.streamDownload(ctx, reader, filename, contentLength)
}

func (c *backupController) streamDownload(ctx *gin.Context, reader io.Reader, filename string, contentLength int64) {
	// Set headers for file download
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	ctx.Header("Content-Type", "application/octet-stream")
	if contentLength > 0 {
		ctx.Header("Content-Length", fmt.Sprintf("%d", contentLength))
	}

	// Stream the file directly to the response body
	// Using io.Copy avoids loading the whole file into memory
	_, err := io.Copy(ctx.Writer, reader)
	if err != nil {
		// If streaming fails mid-way, we can't really change the status code anymore
		// effectively, but logging it is good practice.
		fmt.Printf("Error streaming file: %v\n", err)
	}
}

// HandleUnifiedBackup handles both Full and Schema backups based on query param "type"
// POST /api/backup?type=full   -> Trigger Full Backup
// POST /api/backup?type=schema -> Trigger Schema Backup
// GET  /api/backup?type=full   -> Download Full Backup
// GET  /api/backup?type=schema -> Download Schema Backup
// HandleUnifiedBackup triggers the backup creation and then downloads it immediately
// Any Method /api/backup?type=full   -> Trigger Full + Download
// Any Method /api/backup?type=schema -> Trigger Schema + Download
func (c *backupController) HandleUnifiedBackup(ctx *gin.Context) {
	backupType := ctx.DefaultQuery("type", "full")
	var err error

	// 1. Trigger the backup first
	if backupType == "schema" {
		_, err = c.backupService.TriggerSchemaBackup(ctx.Request.Context())
	} else {
		_, err = c.backupService.TriggerBackup(ctx.Request.Context())
	}

	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_TRIGGER_BACKUP, c.parseExternalError(err), nil)
		ctx.JSON(http.StatusInternalServerError, res)
		return
	}

	// 2. If Trigger successful, proceed to Download
	if backupType == "schema" {
		c.DownloadSchemaBackup(ctx)
	} else {
		c.DownloadBackup(ctx)
	}
}

func (c *backupController) parseExternalError(err error) string {
	var jsonObj map[string]interface{}
	if json.Unmarshal([]byte(err.Error()), &jsonObj) == nil {
		if msg, ok := jsonObj["message"].(string); ok {
			return msg
		}
	}
	return err.Error()
}
