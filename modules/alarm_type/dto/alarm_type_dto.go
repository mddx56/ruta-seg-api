package dto

import "github.com/google/uuid"

type (
	AlarmTypeCreateRequest struct {
		Code        string `json:"code" binding:"required"`
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Severity    string `json:"severity"`
	}

	AlarmTypeResponse struct {
		ID          uuid.UUID `json:"id"`
		Code        string    `json:"code"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Severity    string    `json:"severity"`
	}
)
