package query

import (
	"time"

	"github.com/Caknoooo/go-pagination"
	"gorm.io/gorm"
)

type LogSocket struct {
	LogID     uint      `json:"log_id"`
	Payload   string    `json:"payload"`
	CreatedAt time.Time `json:"created_at"`
}

type LogSocketFilter struct {
	pagination.BaseFilter
	Date string `form:"date" json:"date"` // Format: YYYY-MM-DD
}

func (f *LogSocketFilter) ApplyFilters(query *gorm.DB) *gorm.DB {
	if f.Date != "" {
		query = query.Where("DATE(created_at) = ?", f.Date)
	} else {
		// Default to current day
		query = query.Where("DATE(created_at) = CURRENT_DATE")
	}
	return query
}

func (f *LogSocketFilter) GetTableName() string {
	return "log_sockets"
}

func (f *LogSocketFilter) GetSearchFields() []string {
	return []string{"payload"}
}

func (f *LogSocketFilter) GetDefaultSort() string {
	return "created_at desc"
}

func (f *LogSocketFilter) GetIncludes() []string {
	return f.Includes
}

func (f *LogSocketFilter) GetPagination() pagination.PaginationRequest {
	return f.Pagination
}

func (f *LogSocketFilter) Validate() {
	var validIncludes []string
	allowedIncludes := f.GetAllowedIncludes()
	for _, include := range f.Includes {
		if allowedIncludes[include] {
			validIncludes = append(validIncludes, include)
		}
	}
	f.Includes = validIncludes
}

func (f *LogSocketFilter) GetAllowedIncludes() map[string]bool {
	return map[string]bool{}
}
