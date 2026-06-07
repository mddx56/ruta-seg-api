package query

import (
	"time"

	"github.com/Caknoooo/go-pagination"
	"gorm.io/gorm"
)

type DeviceQuery struct {
	IMEI           string    `json:"imei"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Model          string    `json:"model"`
	SimPhoneNumber string    `json:"sim_phone_number"`
	SimICCID       string    `json:"cod_sim" gorm:"column:sim_icc_id"`
	SimProvider    string    `json:"sim_provider"`
	Status         bool      `json:"status"`
}

type DeviceFilter struct {
	pagination.BaseFilter
}

func (f *DeviceFilter) ApplyFilters(query *gorm.DB) *gorm.DB {
	// Aquí se pueden agregar filtros específicos si f.Search tiene valor, por ejemplo buscar por IMEI o Modelo
	// El BaseFilter ya maneja sort y pagination genérico, pero el search custom va acá si go-pagination no lo hace automático para todo.
	// Asumimos que go-pagination usa GetSearchFields para los LIKE.
	return query
}

func (f *DeviceFilter) GetTableName() string {
	return "devices"
}

func (f *DeviceFilter) GetSearchFields() []string {
	return []string{"imei", "model", "sim_phone_number", "sim_icc_id"}
}

func (f *DeviceFilter) GetDefaultSort() string {
	return "created_at desc"
}

func (f *DeviceFilter) GetIncludes() []string {
	return f.Includes
}

func (f *DeviceFilter) GetPagination() pagination.PaginationRequest {
	return f.Pagination
}

func (f *DeviceFilter) Validate() {
	var validIncludes []string
	allowedIncludes := f.GetAllowedIncludes()
	for _, include := range f.Includes {
		if allowedIncludes[include] {
			validIncludes = append(validIncludes, include)
		}
	}
	f.Includes = validIncludes
}

func (f *DeviceFilter) GetAllowedIncludes() map[string]bool {
	// Por ahora no permitimos includes complejos para mantenerlo simple como pidió el usuario
	return map[string]bool{}
}
