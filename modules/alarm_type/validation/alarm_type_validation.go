package validation

import (
	"github.com/go-playground/validator/v10"
)

type AlarmTypeValidation struct {
	validate *validator.Validate
}

func NewAlarmTypeValidation() *AlarmTypeValidation {
	validate := validator.New()
	return &AlarmTypeValidation{
		validate: validate,
	}
}
