package query

import (
	"time"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"github.com/Caknoooo/go-gin-clean-starter/modules/vehicle/dto"
	"github.com/Caknoooo/go-pagination"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Vehicle struct {
	ID            uuid.UUID                     `json:"id"`
	Placa         string                        `json:"placa"`
	Description   *string                       `json:"description,omitempty"`
	Year          *int                          `json:"year,omitempty"`
	KmLiter       *float64                      `json:"km_liter,omitempty"`
	Chassis       *string                       `json:"chasis"`
	Color         *string                       `json:"color,omitempty"`
	PhotoURL      *string                       `json:"photo_url,omitempty"`
	Status        bool                          `json:"status"`
	ModelID       uuid.UUID                     `json:"model_id"`
	UserID        uuid.UUID                     `json:"user_id"`
	CreatedAt     time.Time                     `json:"created_at"`
	UpdatedAt     time.Time                     `json:"updated_at"`
	User          *entities.User                `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Model         *entities.Model               `json:"model,omitempty" gorm:"foreignKey:ModelID"`
	Installations []entities.DeviceInstallation `json:"installations,omitempty" gorm:"foreignKey:VehicleID"`
}

func (v *Vehicle) ToResponse() dto.VehicleResponse {
	resp := dto.VehicleResponse{
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
		resp.User = &dto.UserInfo{
			ID:    v.User.ID,
			Name:  v.User.Name,
			Email: v.User.Email,
		}
	}

	if v.Model != nil {
		modelInfo := &dto.ModelInfo{
			ID:        v.Model.ID,
			ModelName: v.Model.ModelName,
		}
		if v.Model.Make != nil {
			modelInfo.Make = &dto.MakeInfo{
				ID:       v.Model.Make.ID,
				MakeName: v.Model.Make.MakeName,
			}
		}
		resp.Model = modelInfo
	}

	// Find active installation
	for _, inst := range v.Installations {
		if inst.RemovedAt == nil {
			resp.ActiveInstallation = &dto.VehicleInstallationInfo{
				InstallationID: inst.InstallationID,
				DeviceIMEI:     inst.Imei,
				InstalledAt:    inst.InstalledAt,
			}

			// Try to find group info if available through device relations
			// Note: This requires preloading Device and its GroupDevices
			if inst.Device != nil && len(inst.Device.GroupDevices) > 0 {
				groupDevice := inst.Device.GroupDevices[0]
				if groupDevice.Group != nil {
					resp.Group = &dto.GroupInfo{
						ID:   groupDevice.Group.ID,
						Name: groupDevice.Group.Name,
					}
				}
			}
			break
		}
	}

	return resp
}

type VehicleFilter struct {
	pagination.BaseFilter
}

func (f *VehicleFilter) ApplyFilters(query *gorm.DB) *gorm.DB {
	// Solo vehículos activos
	return query.Where("status = ?", true)
}

func (f *VehicleFilter) GetTableName() string {
	return "vehicles"
}

func (f *VehicleFilter) GetSearchFields() []string {
	return []string{"placa", "chassis", "description"}
}

func (f *VehicleFilter) GetDefaultSort() string {
	return "created_at desc"
}

func (f *VehicleFilter) GetIncludes() []string {
	return f.Includes
}

func (f *VehicleFilter) GetPagination() pagination.PaginationRequest {
	return f.Pagination
}

func (f *VehicleFilter) Validate() {
	var validIncludes []string
	allowedIncludes := f.GetAllowedIncludes()
	for _, include := range f.Includes {
		if allowedIncludes[include] {
			validIncludes = append(validIncludes, include)
		}
	}
	f.Includes = validIncludes
}

func (f *VehicleFilter) GetAllowedIncludes() map[string]bool {
	// Según repos original, se incluyen:
	// Preload("User").
	// Preload("Model").
	// Preload("Model.Make").
	return map[string]bool{
		"User":                              true,
		"Model":                             true,
		"Model.Make":                        true,
		"Installations":                     true,
		"Installations.Device":              true,
		"Installations.Device.GroupDevices": true,
		"Installations.Device.GroupDevices.Group": true,
	}
}
