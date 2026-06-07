package validation

import (
	"github.com/go-playground/validator/v10"
)

type AlarmIncidentValidation struct {
	validate *validator.Validate
}

func NewAlarmIncidentValidation() *AlarmIncidentValidation {
	validate := validator.New()
	return &AlarmIncidentValidation{
		validate: validate,
	}
}
