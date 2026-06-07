package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"github.com/Caknoooo/go-gin-clean-starter/modules/device/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/device/repository"
	vehicleDto "github.com/Caknoooo/go-gin-clean-starter/modules/vehicle/dto"
	"github.com/google/uuid"
	"github.com/samber/do"
)

type DeviceService interface {
	Create(ctx context.Context, req dto.DeviceCreateRequest) (dto.DeviceResponse, error)
	Update(ctx context.Context, req dto.DeviceUpdateRequest) (dto.DeviceResponse, error)
	ChangeStatus(ctx context.Context, imei string) error
	FindAll(ctx context.Context) ([]dto.DeviceResponse, error)
	GetSimple(ctx context.Context, available bool) ([]dto.DeviceSimpleResponse, error)
	FindByIMEI(ctx context.Context, imei string) (dto.DeviceResponse, error)
	FindByIMEIFull(ctx context.Context, imei string) (dto.DeviceFullResponse, error)
	BulkValidate(ctx context.Context, req dto.BulkImportRequest) dto.BulkImportResponse
	BulkImport(ctx context.Context, req dto.BulkImportRequest) dto.BulkImportResponse
	Export(ctx context.Context, includeDisabled bool) ([]dto.DeviceExportItem, error)
	GetCategorizedDevices(ctx context.Context, userID string, isAdmin bool) (dto.CategorizedDevicesResponse, error)
}

type deviceService struct {
	repo        repository.DeviceRepository
	peerChecker PeerChecker
}

func NewDeviceService(injector *do.Injector) (DeviceService, error) {
	repo := do.MustInvoke[repository.DeviceRepository](injector)
	return &deviceService{
		repo:        repo,
		peerChecker: NewPeerChecker(),
	}, nil
}

func (s *deviceService) Create(ctx context.Context, req dto.DeviceCreateRequest) (dto.DeviceResponse, error) {
	// Validar campos requeridos
	if req.IMEI == "" {
		return dto.DeviceResponse{}, fmt.Errorf("el IMEI es requerido")
	}
	if req.Model == "" {
		return dto.DeviceResponse{}, fmt.Errorf("el modelo es requerido")
	}

	// Validar duplicado de IMEI en BD local
	if _, err := s.repo.FindByIMEI(ctx, req.IMEI); err == nil {
		return dto.DeviceResponse{}, fmt.Errorf("el IMEI '%s' ya se encuentra registrado", req.IMEI)
	}

	// Validar que el IMEI no esté registrado en el VPS par
	if exists, err := s.peerChecker.IMEIExistsOnPeer(ctx, req.IMEI); err != nil {
		log.Printf("[PeerChecker] Error consultando VPS par para IMEI %s: %v — se permite el registro", req.IMEI, err)
	} else if exists {
		return dto.DeviceResponse{}, fmt.Errorf("el IMEI '%s' ya se encuentra registrado en el otro servidor", req.IMEI)
	}

	// Validar duplicado de SimPhoneNumber (solo si se proporcionó)
	if req.SimPhoneNumber != nil && *req.SimPhoneNumber != "" {
		if _, err := s.repo.FindBySimPhoneNumber(ctx, *req.SimPhoneNumber); err == nil {
			return dto.DeviceResponse{}, fmt.Errorf("el número de SIM '%s' ya está en uso por otro dispositivo", *req.SimPhoneNumber)
		}
	}

	// Validar duplicado de cod_sim (SimICCID)
	if req.SimICCID != nil && *req.SimICCID != "" {
		if _, err := s.repo.FindByCodSim(ctx, *req.SimICCID); err == nil {
			return dto.DeviceResponse{}, fmt.Errorf("el código SIM '%s' ya está en uso por otro dispositivo", *req.SimICCID)
		}
	}

	// Sanitizar campos opcionales: convertir cadenas vacías a nil
	sanitizeStringPtr := func(s *string) *string {
		if s != nil && *s == "" {
			return nil
		}
		return s
	}

	device := entities.Device{
		IMEI:            req.IMEI,
		Model:           req.Model,
		Protocol:        sanitizeStringPtr(req.Protocol),
		SimPhoneNumber:  sanitizeStringPtr(req.SimPhoneNumber),
		SimICCID:        sanitizeStringPtr(req.SimICCID),
		SimProvider:     sanitizeStringPtr(req.SimProvider),
		APNConf:         sanitizeStringPtr(req.APNConf),
		FirmwareVersion: sanitizeStringPtr(req.FirmwareVersion),
		RemoteIP:        sanitizeStringPtr(req.RemoteIP),
		UserCreator:     req.UserAuditID,
	}

	// UserID and GroupID logic removed as part of normalization
	// Devices are now assigned via installations and group_devices tables

	if err := s.repo.Create(ctx, &device); err != nil {
		return dto.DeviceResponse{}, fmt.Errorf("fallo al crear el dispositivo en base de datos: %w", err)
	}
	return s.mapEntityToDto(device), nil
}

func (s *deviceService) Update(ctx context.Context, req dto.DeviceUpdateRequest) (dto.DeviceResponse, error) {
	device, err := s.repo.FindByIMEI(ctx, req.IMEI)
	if err != nil {
		return dto.DeviceResponse{}, dto.ErrDeviceNotFound
	}

	// No permitimos actualizar el IMEI directamente aquí porque es PK
	// si se necesita, sería un proceso de migración o borrado/creado

	if req.Model != nil {
		device.Model = *req.Model
	}
	if req.Protocol != nil {
		device.Protocol = req.Protocol
	}
	if req.SimPhoneNumber != nil {
		device.SimPhoneNumber = req.SimPhoneNumber
	}
	if req.SimICCID != nil {
		// Validar duplicado de cod_sim solo si cambió
		if *req.SimICCID != "" && (device.SimICCID == nil || *device.SimICCID != *req.SimICCID) {
			if existing, err := s.repo.FindByCodSim(ctx, *req.SimICCID); err == nil && existing.IMEI != device.IMEI {
				return dto.DeviceResponse{}, fmt.Errorf("el código SIM '%s' ya está en uso por el dispositivo '%s'", *req.SimICCID, existing.IMEI)
			}
		}
		device.SimICCID = req.SimICCID
	}
	if req.SimProvider != nil {
		device.SimProvider = req.SimProvider
	}
	if req.APNConf != nil {
		device.APNConf = req.APNConf
	}
	if req.FirmwareVersion != nil {
		device.FirmwareVersion = req.FirmwareVersion
	}

	if req.RemoteIP != nil {
		device.RemoteIP = req.RemoteIP
	}
	device.UserUpdater = req.UserAuditID

	// UserID and GroupID update logic removed as part of normalization

	if err := s.repo.Update(ctx, &device); err != nil {
		return dto.DeviceResponse{}, dto.ErrUpdateDevice
	}

	return s.mapEntityToDto(device), nil
}

func (s *deviceService) ChangeStatus(ctx context.Context, imei string) error {
	if err := s.repo.ChangeStatus(ctx, imei); err != nil {
		return dto.ErrDeleteDevice
	}
	return nil
}

func (s *deviceService) FindAll(ctx context.Context) ([]dto.DeviceResponse, error) {
	devices, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, dto.ErrGetListDevice
	}

	responses := make([]dto.DeviceResponse, 0)
	for _, device := range devices {
		responses = append(responses, s.mapEntityToDto(device))
	}

	return responses, nil
}

func (s *deviceService) GetSimple(ctx context.Context, available bool) ([]dto.DeviceSimpleResponse, error) {
	devices, err := s.repo.FindAllSimple(ctx, available)
	if err != nil {
		return nil, dto.ErrGetListDevice
	}

	responses := make([]dto.DeviceSimpleResponse, 0)
	for _, device := range devices {
		responses = append(responses, dto.DeviceSimpleResponse{
			IMEI: device.IMEI,
		})
	}

	return responses, nil
}

func (s *deviceService) FindByIMEI(ctx context.Context, imei string) (dto.DeviceResponse, error) {
	device, err := s.repo.FindByIMEI(ctx, imei)
	if err != nil {
		return dto.DeviceResponse{}, dto.ErrDeviceNotFound
	}

	return s.mapEntityToDto(device), nil
}

func (s *deviceService) FindByIMEIFull(ctx context.Context, imei string) (dto.DeviceFullResponse, error) {
	device, err := s.repo.FindByIMEIFull(ctx, imei)
	if err != nil {
		return dto.DeviceFullResponse{}, dto.ErrDeviceNotFound
	}

	base := s.mapEntityToDto(device)
	installations := make([]dto.DeviceInstallationItem, 0)
	available := true
	for _, inst := range device.Installations {
		if inst.RemovedAt == nil && inst.Status {
			available = false
		}
		item := dto.DeviceInstallationItem{
			InstallationID: inst.InstallationID.String(),
			InstalledAt:    inst.InstalledAt,
			RemovedAt:      inst.RemovedAt,
			InstallReason:  inst.InstallReason,
			RemovalReason:  inst.RemovalReason,
		}

		if inst.Vehicle != nil {
			v := inst.Vehicle
			item.Vehicle = &vehicleDto.VehicleResponse{
				ID:          v.ID,
				Placa:       v.Placa,
				Description: v.Description,
				Year:        v.Year,
				KmLiter:     v.KmLiter,
				Chassis:     v.Chassis,
				Color:       v.Color,
				PhotoURL:    v.PhotoURL,
				CreatedAt:   v.CreatedAt,
				UpdatedAt:   v.UpdatedAt,
				Status:      v.Status,
			}

			if v.User != nil {
				item.Vehicle.User = &vehicleDto.UserInfo{
					ID:    v.User.ID,
					Name:  v.User.Name,
					Email: v.User.Email,
				}
			}
			if v.Model != nil {
				item.Vehicle.Model = &vehicleDto.ModelInfo{
					ID:        v.Model.ID,
					ModelName: v.Model.ModelName,
				}
				if v.Model.Make != nil {
					item.Vehicle.Model.Make = &vehicleDto.MakeInfo{
						ID:       v.Model.Make.ID,
						MakeName: v.Model.Make.MakeName,
					}
				}
			}
		}

		installations = append(installations, item)
	}

	return dto.DeviceFullResponse{
		Device:                   base,
		Installations:            installations,
		AvailableForInstallation: available,
	}, nil
}

func (s *deviceService) mapEntityToDto(device entities.Device) dto.DeviceResponse {
	res := dto.DeviceResponse{
		IMEI:           device.IMEI,
		Model:          device.Model,
		SimPhoneNumber: device.SimPhoneNumber,
		SimProvider:    device.SimProvider,
		SimICCID:       device.SimICCID,
		Status:         device.Status,
		CreatedAt:      device.CreatedAt,
		UpdatedAt:      device.UpdatedAt,
	}

	if len(device.GroupDevices) > 0 {
		res.Groups = make([]dto.GroupInfo, 0)
		for _, gd := range device.GroupDevices {
			if gd.Group != nil {
				groupInfo := dto.GroupInfo{
					ID:   gd.Group.ID.String(),
					Name: gd.Group.Name,
				}

				if gd.Group.User != nil {
					groupInfo.Owner = &vehicleDto.UserInfo{
						ID:    gd.Group.User.ID,
						Name:  gd.Group.User.Name,
						Email: gd.Group.User.Email,
					}
				}
				res.Groups = append(res.Groups, groupInfo)
			}
		}
	}

	if len(device.Installations) > 0 {
		inst := device.Installations[0]
		if inst.Vehicle != nil {
			v := inst.Vehicle
			res.ActiveVehicle = &vehicleDto.VehicleResponse{
				ID:          v.ID,
				Placa:       v.Placa,
				Description: v.Description,
				Year:        v.Year,
				KmLiter:     v.KmLiter,
				CreatedAt:   v.CreatedAt,
				UpdatedAt:   v.UpdatedAt,
				Status:      true,
			}

			if v.User != nil {
				res.ActiveVehicle.User = &vehicleDto.UserInfo{
					ID:    v.User.ID,
					Name:  v.User.Name,
					Email: v.User.Email,
				}
			}

			if v.Model != nil {
				res.ActiveVehicle.Model = &vehicleDto.ModelInfo{
					ID:        v.Model.ID,
					ModelName: v.Model.ModelName,
				}
				if v.Model.Make != nil {
					res.ActiveVehicle.Model.Make = &vehicleDto.MakeInfo{
						ID:       v.Model.Make.ID,
						MakeName: v.Model.Make.MakeName,
					}
				}
			}
		}
	}

	return res
}

// ── Bulk helpers ───────────────────────────────────────────────────────

// bulkValidateItems validates items against payload duplicates and DB duplicates.
// Returns per-item results. If persistOnSuccess is true, valid items are inserted.
func (s *deviceService) bulkValidateItems(ctx context.Context, items []dto.BulkImportItem, userAuditID *string, persistOnSuccess bool) dto.BulkImportResponse {
	results := make([]dto.BulkImportItemResult, 0, len(items))

	var batchID *string
	if persistOnSuccess {
		id := uuid.New().String()
		batchID = &id
	}

	// ── 1. Collect all values for batch DB lookups ──────────────────
	allIMEIs := make([]string, 0, len(items))
	allCodSims := make([]string, 0, len(items))
	for _, item := range items {
		if item.IMEI != "" {
			allIMEIs = append(allIMEIs, item.IMEI)
		}
		if item.CodSim != "" {
			allCodSims = append(allCodSims, item.CodSim)
		}
	}

	// ── 2. Batch lookup existing records ────────────────────────────
	existingIMEISet := make(map[string]bool)
	existingCodSimSet := make(map[string]bool)

	if len(allIMEIs) > 0 {
		if existingDevices, err := s.repo.FindByIMEIs(ctx, allIMEIs); err == nil {
			for _, d := range existingDevices {
				existingIMEISet[d.IMEI] = true
			}
		}
		// Verificar también en el VPS par
		if peerFound, err := s.peerChecker.IMEIsExistOnPeer(ctx, allIMEIs); err != nil {
			log.Printf("[PeerChecker] Error consultando VPS par en bulk: %v — se omite validación cruzada", err)
		} else {
			for imei := range peerFound {
				existingIMEISet[imei] = true
			}
		}
	}
	if len(allCodSims) > 0 {
		if existingByCodSim, err := s.repo.FindByCodSims(ctx, allCodSims); err == nil {
			for _, d := range existingByCodSim {
				if d.SimICCID != nil {
					existingCodSimSet[*d.SimICCID] = true
				}
			}
		}
	}

	// ── 3. Track payload-level duplicates ───────────────────────────
	seenIMEI := make(map[string]int) // value = first row (1-based)
	seenCodSim := make(map[string]int)

	// ── 4. Validate each item ───────────────────────────────────────
	successCount := 0
	devicesToCreate := make([]entities.Device, 0)

	for i, item := range items {
		row := i + 1
		var itemErrors []string

		// Validaciones de campos obligatorios
		if item.IMEI == "" {
			itemErrors = append(itemErrors, "Falta el IMEI, es obligatorio")
		} else if len(item.IMEI) < 15 || len(item.IMEI) > 20 {
			itemErrors = append(itemErrors, "El IMEI no tiene un formato válido, debe tener entre 15 y 20 dígitos")
		}

		if item.CodSim == "" {
			itemErrors = append(itemErrors, "Falta el Código SIM, es obligatorio")
		}

		// Verificar duplicados dentro del archivo
		if item.IMEI != "" {
			if firstRow, exists := seenIMEI[item.IMEI]; exists {
				itemErrors = append(itemErrors, fmt.Sprintf("Este IMEI está repetido en el archivo, ya aparece en la fila %d", firstRow))
			} else {
				seenIMEI[item.IMEI] = row
			}
		}
		if item.CodSim != "" {
			if firstRow, exists := seenCodSim[item.CodSim]; exists {
				itemErrors = append(itemErrors, fmt.Sprintf("Este Código SIM está repetido en el archivo, ya aparece en la fila %d", firstRow))
			} else {
				seenCodSim[item.CodSim] = row
			}
		}

		// Verificar si ya existen en el sistema
		if item.IMEI != "" && existingIMEISet[item.IMEI] {
			itemErrors = append(itemErrors, "Este IMEI ya se encuentra registrado en el sistema (puede ser en este servidor o en el servidor par)")
		}
		if item.CodSim != "" && existingCodSimSet[item.CodSim] {
			itemErrors = append(itemErrors, "Este Código SIM ya se encuentra registrado en el sistema")
		}

		result := dto.BulkImportItemResult{
			Row:            row,
			IMEI:           item.IMEI,
			CodSim:         item.CodSim,
			SimPhoneNumber: item.SimPhoneNumber,
			SimProvider:    item.SimProvider,
			Success:        len(itemErrors) == 0,
			Errors:         itemErrors,
		}

		if len(itemErrors) == 0 && persistOnSuccess {
			codSim := item.CodSim
			device := entities.Device{
				IMEI:           item.IMEI,
				Model:          "GT06",
				SimICCID:       &codSim,
				SimPhoneNumber: item.SimPhoneNumber,
				SimProvider:    item.SimProvider,
				Batch:          batchID,
				UserCreator:    userAuditID,
			}
			devicesToCreate = append(devicesToCreate, device)
		}

		if len(itemErrors) == 0 {
			successCount++
		}
		results = append(results, result)
	}

	// ── 5. Guardar los dispositivos válidos ─────────────────────────
	if persistOnSuccess && len(devicesToCreate) > 0 {
		for _, device := range devicesToCreate {
			if err := s.repo.Create(ctx, &device); err != nil {
				for j := range results {
					if results[j].IMEI == device.IMEI && results[j].Success {
						results[j].Success = false
						results[j].Errors = append(results[j].Errors, "No se pudo guardar este dispositivo, intente nuevamente")
						successCount--
						break
					}
				}
			}
		}
	}

	return dto.BulkImportResponse{
		TotalReceived: len(items),
		TotalSuccess:  successCount,
		TotalFailed:   len(items) - successCount,
		Results:       results,
	}
}

func (s *deviceService) BulkValidate(ctx context.Context, req dto.BulkImportRequest) dto.BulkImportResponse {
	return s.bulkValidateItems(ctx, req.Items, req.UserAuditID, false)
}

func (s *deviceService) BulkImport(ctx context.Context, req dto.BulkImportRequest) dto.BulkImportResponse {
	return s.bulkValidateItems(ctx, req.Items, req.UserAuditID, true)
}

func (s *deviceService) Export(ctx context.Context, includeDisabled bool) ([]dto.DeviceExportItem, error) {
	devices, err := s.repo.FindAllForExport(ctx, includeDisabled)
	if err != nil {
		return nil, dto.ErrGetListDevice
	}

	items := make([]dto.DeviceExportItem, 0, len(devices))
	for _, d := range devices {
		status := "Activo"
		if !d.Status {
			status = "Inactivo"
		}
		items = append(items, dto.DeviceExportItem{
			IMEI:           d.IMEI,
			CodSim:         d.SimICCID,
			SimPhoneNumber: d.SimPhoneNumber,
			SimProvider:    d.SimProvider,
			Status:         status,
		})
	}

	return items, nil
}

func (s *deviceService) GetCategorizedDevices(ctx context.Context, userID string, isAdmin bool) (dto.CategorizedDevicesResponse, error) {
	rows, err := s.repo.GetDevicesWithLastPosition(ctx, userID, isAdmin)
	if err != nil {
		return dto.CategorizedDevicesResponse{}, err
	}

	now := time.Now()
	result := make(dto.CategorizedDevicesResponse, 0, len(rows))

	for _, row := range rows {
		battery := 100.0
		var ignition *bool

		if row.Attributes != nil && *row.Attributes != "" {
			var attrs map[string]interface{}
			if err := json.Unmarshal([]byte(*row.Attributes), &attrs); err == nil {
				for _, key := range []string{"batteryLevel", "battery", "power"} {
					if val, ok := attrs[key]; ok {
						if v, ok := val.(float64); ok {
							battery = v
							break
						}
					}
				}
				if val, ok := attrs["ignition"]; ok {
					if v, ok := val.(bool); ok {
						b := v
						ignition = &b
					}
				}
			}
		}

		// Dereference nullable position fields safely
		lat, lng, speed, course := 0.0, 0.0, 0, 0
		var deviceTime time.Time
		if row.Latitude != nil {
			lat = *row.Latitude
		}
		if row.Longitude != nil {
			lng = *row.Longitude
		}
		if row.Speed != nil {
			speed = *row.Speed
		}
		if row.Course != nil {
			course = *row.Course
		}
		if row.DeviceTime != nil {
			deviceTime = *row.DeviceTime
		}

		// Default: offline (si no tiene coordenadas nunca o es demasiado antigua)
		category := "offline"
		if row.ServerTime != nil {
			hoursDiff := now.Sub(*row.ServerTime).Hours()
			isOffline := hoursDiff > 1.5 || (battery < 20 && hoursDiff > 1.0)
			
			if !isOffline {
				if speed > 5 {
					category = "live"
				} else if ignition != nil && *ignition {
					category = "idling"
				} else {
					category = "parked"
				}
			}
		}

		result = append(result, dto.CategorizedDevice{
			IMEI:       row.IMEI,
			Placa:      row.Placa,
			Make:       row.Make,
			Model:      row.Model,
			Color:      row.Color,
			Latitude:   lat,
			Longitude:  lng,
			Speed:      speed,
			Course:     course,
			DeviceTime: deviceTime,
			Battery:    battery,
			Ignition:   ignition,
			Category:   category,
		})
	}

	return result, nil
}
