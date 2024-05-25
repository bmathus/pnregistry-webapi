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

// date format of our date fields
const dateFormat = "2006-01-02"

// custom marshal of dates to string format
func (d DateType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", time.Time(d).Format(dateFormat))), nil
}

// custom unmarshal and validation of dates
func (d *DateType) UnmarshalJSON(b []byte) error {

	// check if the input is a valid JSON string
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return errors.New("Invalid date format, must be a string in YYYY-MM-DD format")
	}

	// trim leading and trailing whitespaces
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

	*d = DateType(t)
	return nil
}

// custom bson marshaling and unmarshaling
func (d DateType) MarshalBSONValue() (bsontype.Type, []byte, error) {
	return bson.MarshalValue(time.Time(d))
}

func (d *DateType) UnmarshalBSONValue(t bsontype.Type, data []byte) error {
	var tm time.Time
	if err := bson.UnmarshalValue(t, data, &tm); err != nil {
		return err
	}
	*d = DateType(tm)
	return nil
}

// helper functions with dates
func (d DateType) String() string {
	return time.Time(d).Format(dateFormat)
}

func (d DateType) After(u DateType) bool {
	return time.Time(d).After(time.Time(u))
}
