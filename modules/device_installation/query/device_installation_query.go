package query

import (
	"time"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"github.com/Caknoooo/go-pagination"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CreatedBy struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"nombre"`
}

type DeviceInstallation struct {
	InstallationID uuid.UUID  `json:"installation_id"`
	Imei           string     `json:"imei"`
	VehicleID      uuid.UUID  `json:"-"`
	Chassis        *string    `json:"chasis" gorm:"-"`
	InstalledAt    time.Time  `json:"installed_at"`
	RemovedAt      *time.Time `json:"removed_at,omitempty"`
	InstallReason  *string    `json:"install_reason,omitempty"`
	RemovalReason  *string    `json:"removal_reason,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	Status         bool       `json:"status"`

	CreatedBy      *CreatedBy `json:"created_by" gorm:"-"`
	UserCreationID *uuid.UUID `json:"-"`

	Device       *entities.Device  `json:"device,omitempty" gorm:"foreignKey:Imei;references:IMEI"`
	Vehicle      *entities.Vehicle `json:"vehicle,omitempty" gorm:"foreignKey:VehicleID"`
	UserCreation *entities.User    `json:"-" gorm:"foreignKey:UserCreationID"`
}

type DeviceInstallationFilter struct {
	pagination.BaseFilter
}

func (f *DeviceInstallationFilter) ApplyFilters(query *gorm.DB) *gorm.DB {
	return query
}

func (f *DeviceInstallationFilter) GetTableName() string {
	return "device_installations"
}

func (f *DeviceInstallationFilter) GetSearchFields() []string {
	return []string{"imei", "work_order_id"}
}

func (f *DeviceInstallationFilter) GetDefaultSort() string {
	return "installed_at desc"
}

func (f *DeviceInstallationFilter) GetIncludes() []string {
	return f.Includes
}

func (f *DeviceInstallationFilter) GetPagination() pagination.PaginationRequest {
	return f.Pagination
}

func (f *DeviceInstallationFilter) Validate() {
	var validIncludes []string
	allowedIncludes := f.GetAllowedIncludes()
	for _, include := range f.Includes {
		if allowedIncludes[include] {
			validIncludes = append(validIncludes, include)
		}
	}
	f.Includes = validIncludes
}

func (f *DeviceInstallationFilter) GetAllowedIncludes() map[string]bool {
	return map[string]bool{
		"Device":       true,
		"Vehicle":      true,
		"UserCreation": true,
	}
}
