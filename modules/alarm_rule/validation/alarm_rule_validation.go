package validation

import (
	"github.com/go-playground/validator/v10"
)

type AlarmRuleValidation struct {
	validate *validator.Validate
}

func NewAlarmRuleValidation() *AlarmRuleValidation {
	validate := validator.New()
	return &AlarmRuleValidation{
		validate: validate,
	}
}
