package pn_registry

import (
	"time"
)

// Custom date type to work with formats dd-mm-yyyy without time part
type DateType time.Time

// Enum constant of possible values of field "reason"
const (
	Choroba                     = "choroba"
	Uraz                        = "uraz"
	ChorobaZPovolania           = "choroba z povolania"
	KarantenneOpatrenieIzolacia = "karantenne opatrenie/izolacia"
	PracovnyUraz                = "pracovny uraz"
	Ine                         = "ine"
)

type Record struct {
	Id          string    `json:"id" bson:"id" binding:"required"`
	FullName    string    `json:"fullName,omitempty" bson:"fullName,omitempty" binding:"max-length-50"`
	PatientId   string    `json:"patientId" bson:"patientId" binding:"required,only-digits-max-length-10"`
	Employer    string    `json:"employer" bson:"employer" binding:"required,max-length-50"`
	Reason      string    `json:"reason" bson:"reason" binding:"required,not-valid-reason-value"`
	Issued      DateType  `json:"issued" bson:"issued" binding:"required"`
	ValidFrom   DateType  `json:"validFrom" bson:"validFrom" binding:"required"`
	ValidUntil  DateType  `json:"validUntil" bson:"validUntil" binding:"required"`
	CheckUp     *DateType `json:"checkUp,omitempty" bson:"checkUp,omitempty"`
	CheckUpDone bool      `json:"checkUpDone" bson:"checkUpDone"`
}
