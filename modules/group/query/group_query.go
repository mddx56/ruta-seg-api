package query

import (
	"time"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"github.com/Caknoooo/go-pagination"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Group struct {
	ID          uuid.UUID         `json:"id"`
	Name        string            `json:"name"`
	Description *string           `json:"description,omitempty"`
	UserID      uuid.UUID         `json:"user_id"`
	User        *entities.User    `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Devices     []entities.Device `json:"devices,omitempty" gorm:"many2many:group_devices;foreignKey:ID;joinForeignKey:GroupID;References:IMEI;joinReferences:DeviceIMEI"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type GroupFilter struct {
	pagination.BaseFilter
	UserID string `form:"user_id" json:"user_id"`
}

func (f *GroupFilter) ApplyFilters(query *gorm.DB) *gorm.DB {
	if f.UserID != "" {
		query = query.Where("user_id = ?", f.UserID)
	}
	return query
}

func (f *GroupFilter) GetTableName() string {
	return "groups"
}

func (f *GroupFilter) GetSearchFields() []string {
	return []string{"groups.name", "groups.description"}
}

func (f *GroupFilter) GetDefaultSort() string {
	return "created_at desc"
}

func (f *GroupFilter) GetIncludes() []string {
	return f.Includes
}

func (f *GroupFilter) GetPagination() pagination.PaginationRequest {
	return f.Pagination
}

func (f *GroupFilter) Validate() {
	var validIncludes []string
	allowedIncludes := f.GetAllowedIncludes()
	for _, include := range f.Includes {
		if allowedIncludes[include] {
			validIncludes = append(validIncludes, include)
		}
	}
	f.Includes = validIncludes
}

func (f *GroupFilter) GetAllowedIncludes() map[string]bool {
	return map[string]bool{
		"User":    true,
		"Devices": true,
	}
}
