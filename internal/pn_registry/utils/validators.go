package utils

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/bmathus/pnregistry-webapi/internal/pn_registry/models"
	"github.com/gin-gonic/gin"
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

func RequestBodyValidator(ctx *gin.Context) (*models.Record, error) {
	newRecord := models.Record{}

	if err := ctx.ShouldBindJSON(&newRecord); err != nil {
		return nil, err
	}

	if newRecord.CheckUp != nil && newRecord.ValidFrom.After(*newRecord.CheckUp) {
		return nil, fmt.Errorf("'Check Up' date can only be on or after 'Valid from' date")
	}

	if newRecord.ValidFrom.After(newRecord.ValidUntil) {
		return nil, fmt.Errorf("'Valid until' date can only be on or after 'Valid from' date")
	}
	return &newRecord, nil

}

func FullNameSetter(record *models.Record, patientRecords []models.Record) (int, error) {
	if record.FullName == "" { // is fullName is not provided
		if len(patientRecords) != 0 { // inherit fullname from existing records
			record.FullName = patientRecords[0].FullName
		} else {
			return http.StatusNotFound, fmt.Errorf("Patient's PN records not found to inherit Full Name")
		}
	}

	return http.StatusOK, nil
}

func FullNameValidator(record *models.Record, patientRecords []models.Record) (int, error) {
	if (len(patientRecords) != 0) && record.FullName != patientRecords[0].FullName {
		return http.StatusConflict, fmt.Errorf("Full Name does not correspond to patient (conflict with existing records)")
	}
	return http.StatusOK, nil
}

func OverlapValidator(newRecord *models.Record, patientRecords []models.Record) error {
	for _, record := range patientRecords {
		if !newRecord.ValidFrom.After(record.ValidUntil) {
			return fmt.Errorf("Patient already has more up-to-date record or their validity overlap")
		}
	}
	return nil

}
