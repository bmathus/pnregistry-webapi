package pn_registry

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

type DateType time.Time

const dateFormat = "2006-01-02"

// MarshalJSON implements the json.Marshaler interface.
func (d DateType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", time.Time(d).Format(dateFormat))), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (d *DateType) UnmarshalJSON(b []byte) error {
	// Check if the input is a valid JSON string
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return errors.New("Invalid date format, must be a string in YYYY-MM-DD format")
	}

	// Trim leading and trailing whitespaces
	s = strings.TrimSpace(s)
	if s == "" {
		return errors.New("Date string is empty")
	}

	t, err := time.Parse(dateFormat, s)
	if err != nil {
		return fmt.Errorf("Invalid date format, must be YYYY-MM-DD: %w", err)
	}

	if t.Before(time.Date(0001, 1, 2, 0, 0, 0, 0, time.UTC)) || t.After(time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)) {
		return errors.New("Date is out of range, must be between 0001-01-02 and 9999-12-31")
	}

	// if err != nil {
	// 	// Check if the error is due to the date being out of range
	// 	if _, ok := err.(*time.ParseError); ok {
	// 		return errors.New("Date is out of range, must be between 0000-01-01 and 9999-12-31")
	// 	}
	// 	return fmt.Errorf("Date format is invalid, must be YYYY-MM-DD: %w", err)
	// }

	*d = DateType(t)
	return nil
}

// MarshalBSONValue implements the bson.ValueMarshaler interface.
func (d DateType) MarshalBSONValue() (bsontype.Type, []byte, error) {
	return bson.MarshalValue(time.Time(d))
}

// UnmarshalBSONValue implements the bson.ValueUnmarshaler interface.
func (d *DateType) UnmarshalBSONValue(t bsontype.Type, data []byte) error {
	var tm time.Time
	if err := bson.UnmarshalValue(t, data, &tm); err != nil {
		return err
	}
	*d = DateType(tm)
	return nil
}

// String implements the Stringer interface.
func (d DateType) String() string {
	return time.Time(d).Format(dateFormat)
}

func (d DateType) After(u DateType) bool {
	return time.Time(d).After(time.Time(u))
}

type Record struct {
	Id          string    `json:"id" bson:"id" binding:"required"`
	FullName    string    `json:"fullName,omitempty" bson:"fullName,omitempty"`
	PatientId   string    `json:"patientId" bson:"patientId" binding:"required,only-digits-max-length-10"`
	Employer    string    `json:"employer" bson:"employer" binding:"required"`
	Reason      string    `json:"reason" bson:"reason" binding:"required"`
	Issued      DateType  `json:"issued" bson:"issued" binding:"required"`
	ValidFrom   DateType  `json:"validFrom" bson:"validFrom" binding:"required"`
	ValidUntil  DateType  `json:"validUntil" bson:"validUntil" binding:"required"`
	CheckUp     *DateType `json:"checkUp,omitempty" bson:"checkUp,omitempty"`
	CheckUpDone bool      `json:"checkUpDone" bson:"checkUpDone"`
}
