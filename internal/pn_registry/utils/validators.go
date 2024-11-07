package utils

import (
	"regexp"

	"github.com/bmathus/pnregistry-webapi/internal/pn_registry/models"
	"github.com/go-playground/validator/v10"
)

// custom validator for PatientId fields - only digits and max length 10
func PatientIDValidator(fl validator.FieldLevel) bool {
	patientID := fl.Field().String()
	matched, _ := regexp.MatchString(`^\d{1,10}$`, patientID)
	return matched
}

// custom validator for fullName and employer fields-  max length of 50 characters
func MaxLengthValidator(fl validator.FieldLevel) bool {
	return len(fl.Field().String()) <= 50
}

// custom validator for reason field - validates only possible values
func ReasonValidator(fl validator.FieldLevel) bool {
	// validation vith struct for O(1) speed
	var ValidReasons = map[string]struct{}{
		models.Choroba:                     {},
		models.Uraz:                        {},
		models.ChorobaZPovolania:           {},
		models.KarantenneOpatrenieIzolacia: {},
		models.PracovnyUraz:                {},
		models.Ine:                         {},
	}

	reason := fl.Field().String()
	_, valid := ValidReasons[reason]
	return valid
}
