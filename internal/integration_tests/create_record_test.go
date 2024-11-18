// internal/integration_tests/handlers_tests/create_record_test.go
package integrationtests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bmathus/pnregistry-webapi/internal/db_service"
	"github.com/bmathus/pnregistry-webapi/internal/pn_registry"
	"github.com/bmathus/pnregistry-webapi/internal/pn_registry/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func MakeCreateRecordRequest(recordString string, engine *gin.Engine) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(http.MethodPost, "/api/records/", bytes.NewBuffer([]byte(recordString)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	return w
}

// Create Record -> DbService integration
func TestCreateRecordNoDbService(t *testing.T) {
	//Setup engine without db_service in context
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	pn_registry.AddRoutes(engine)

	// Make test request
	req, _ := http.NewRequest(http.MethodPost, "/api/records/", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	// Assert returned status code and body
	assert.Equal(t, http.StatusInternalServerError, w.Code, "Expected HTTP status 500 Internal Server Error")
	expectedResponse := `{
	"error":"db not found",
	"message":"db not found",
	"status":"Internal Server Error"}`
	assert.JSONEq(t, expectedResponse, w.Body.String(), "Expected JSON response does not match")
}

func TestCreateRecordWrongDbService(t *testing.T) {
	//Setup engine with wrong db_service in context
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(func(ctx *gin.Context) {
		ctx.Set("db_service", "incorrectType")
		ctx.Next()
	})
	pn_registry.AddRoutes(engine)

	// Make test request
	req, _ := http.NewRequest(http.MethodPost, "/api/records/", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	// Assert returned status code and body
	expectedResponse := `{
	"error":"cannot cast db_service context to db_service.DbService",
	"message":"db_service context is not of type db_service.DbService",
	"status":"Internal Server Error"}`
	assert.Equal(t, http.StatusInternalServerError, w.Code, "Expected HTTP status 500 Internal Server Error")
	assert.JSONEq(t, expectedResponse, w.Body.String(), "Expected JSON response does not match")
}

// Create record -> Find Document -> Create Document (if stored in database)
func TestCreateRecord(t *testing.T) {
	//Setup test engine and test record
	engine := SetupTestEngine()
	record := `{
        "id": "@new",
        "fullName": "John Doe",
        "patientId": "12345678",
        "employer": "Example Corp",
        "reason": "choroba",
        "issued": "2024-11-08",
        "validFrom": "2024-11-08",
        "validUntil": "2024-11-09",
		"checkUp": "2024-11-08"
    }`
	var expectedRecord models.Record
	json.Unmarshal([]byte(record), &expectedRecord)

	//Make request and save record from responce
	w := MakeCreateRecordRequest(record, engine)
	var returnedRecord models.Record
	err := json.Unmarshal(w.Body.Bytes(), &returnedRecord)
	assert.NoError(t, err)

	// Assert new ID and returned record
	assert.NotEqual(t, "@new", returnedRecord.Id, "Expected generated ID instead of placeholder '@new'")
	expectedRecord.Id = returnedRecord.Id
	assert.Equal(t, expectedRecord, returnedRecord, "Returned record expected to match the request record")

	// Check if new record is stored in the database
	storedRecord, err := testDbService.FindDocument(context.Background(), returnedRecord.Id)
	assert.NoError(t, err)
	assert.Equal(t, expectedRecord, *storedRecord, "Stored record expected to match the request record")

	// Register cleanup to run after test completion
	t.Cleanup(func() {
		err := testDbService.ClearCollection(context.Background())
		assert.NoError(t, err, "Expected the collection to be cleared after the test")
	})
}

// Create record -> Find Document -> Create Document (conflict record exist)
func TestCreateRecordConflict(t *testing.T) {
	engine := SetupTestEngine()
	record := `{
        "id": "@new",
        "fullName": "John Doe",
        "patientId": "12345678",
        "employer": "Example Corp",
        "reason": "choroba",
        "issued": "2024-11-08",
        "validFrom": "2024-11-08",
        "validUntil": "2024-11-09",
		"checkUp": "2024-11-08" }`

	w := MakeCreateRecordRequest(record, engine)
	assert.Equal(t, http.StatusCreated, w.Code, "Expected HTTP status 201 Created")

	// Unmarshal received record to struct
	var returnedRecord models.Record
	json.Unmarshal(w.Body.Bytes(), &returnedRecord)

	//Change its dates
	futureDate := models.DateType(time.Date(2024, 11, 11, 0, 0, 0, 0, time.UTC))
	returnedRecord.CheckUp = &futureDate
	returnedRecord.Issued = futureDate
	returnedRecord.ValidFrom = futureDate
	returnedRecord.ValidUntil = futureDate

	// Marshal the updated record back to JSON
	updatedRecordJson, err := json.Marshal(returnedRecord)
	assert.NoError(t, err, "Failed to marshal updated record")

	// Make request with updated record
	w = MakeCreateRecordRequest(string(updatedRecordJson), engine)
	assert.Equal(t, http.StatusConflict, w.Code, "Expected HTTP status 409 Conflict")
	assert.Contains(t, w.Body.String(), "Record already exists")

	t.Cleanup(func() {
		err := testDbService.ClearCollection(context.Background())
		assert.NoError(t, err, "Expected the collection to be cleared after the test")
	})
}

// Create record -> Find Document (connection fail)
func TestCreateRecordConnectionFailed(t *testing.T) {
	originalDbService := testDbService
	testDbService = db_service.NewMongoService[models.Record](db_service.MongoServiceConfig{
		ServerHost: "localhost",
		ServerPort: 27018,
		UserName:   "wronguser",
		Password:   "wrongpass",
		DbName:     "pn-registry-test",
		Collection: "record",
	})

	engine := SetupTestEngine()

	// Valid record to create
	record := `{
		"id": "@new",
		"fullName": "John Doe",
		"patientId": "12345678",
		"employer": "Example Corp",
		"reason": "choroba",
		"issued": "2024-11-08",
		"validFrom": "2024-11-08",
		"validUntil": "2024-11-09",
		"checkUp": "2024-11-08"
	}`
	w := MakeCreateRecordRequest(record, engine)
	assert.Equal(t, http.StatusBadGateway, w.Code, "Expected HTTP status 502 Bad Gateway")
	assert.Contains(t, w.Body.String(), "connection() error occurred during connection handshake: auth error")

	testDbService = originalDbService
}

// Create record -> Request Body Validator (missing field)
func TestCreateRecordMissingFieldValidation(t *testing.T) {
	engine := SetupTestEngine()
	record := `{
        "id": "@new",
        "fullName": "John Doe",
        "patientId": "1234567890",
        "employer": "Example Corp",
        "reason": "choroba",
        "validFrom": "2024-11-08",
        "validUntil": "2024-11-09",
		"checkUp": "2024-11-07"
    }`
	w := MakeCreateRecordRequest(record, engine)
	assert.Equal(t, http.StatusBadRequest, w.Code, "Expected HTTP status 400 Bad Request")
	assert.Contains(t, w.Body.String(), "Key: 'Record.Issued' Error:Field validation for 'Issued' failed on the 'required' tag")
}

// Create record -> Request Body Validator (check up date validation)
func TestCreateRecordCheckUpDateValidation(t *testing.T) {
	engine := SetupTestEngine()
	record := `{
        "id": "@new",
        "fullName": "John Doe",
        "patientId": "1234567890",
        "employer": "Example Corp",
        "reason": "choroba",
        "issued": "2024-11-08",
        "validFrom": "2024-11-08",
        "validUntil": "2024-11-09",
		"checkUp": "2024-11-07"
    }`
	w := MakeCreateRecordRequest(record, engine)
	assert.Equal(t, http.StatusBadRequest, w.Code, "Expected HTTP status 400 Bad Request")
	assert.Contains(t, w.Body.String(), "'Check Up' date can only be on or after 'Valid from' date")
}

// Create record -> Request Body Validator (valid until date validation)
func TestCreateRecordValidUntilDateValidation(t *testing.T) {
	engine := SetupTestEngine()
	record := `{
        "id": "@new",
        "fullName": "John Doe",
        "patientId": "1234567890",
        "employer": "Example Corp",
        "reason": "choroba",
        "issued": "2024-11-08",
        "validFrom": "2024-11-08",
        "validUntil": "2024-11-07",
		"checkUp": "2024-11-08"
    }`
	w := MakeCreateRecordRequest(record, engine)
	assert.Equal(t, http.StatusBadRequest, w.Code, "Expected HTTP status 400 Bad Request")
	assert.Contains(t, w.Body.String(), "'Valid until' date can only be on or after 'Valid from' date")
}

// Create record -> Request Body validator -> custom date marshaling
func TestCreateRecordDateValidation(t *testing.T) {
	engine := SetupTestEngine()
	record := `{
        "id": "@new",
        "fullName": "John Doe",
        "patientId": "123",
        "employer": "Example Corp",
        "reason": "choroba",
        "issued": "0001-01-01",
        "validFrom": "2024-11-08",
        "validUntil": "2024-11-08",
		"checkUp": "2024-11-08"
	}`
	w := MakeCreateRecordRequest(record, engine)
	assert.Equal(t, http.StatusBadRequest, w.Code, "Expected HTTP status 400 Bad Request")
	assert.Contains(t, w.Body.String(), "Date is out of range, must be between 0001-01-02 and 9999-12-31")
}

// Create record -> Request Body Validator -> Patient ID validator
func TestCreateRecordPatientIdValidation(t *testing.T) {
	engine := SetupTestEngine()
	record := `{
        "id": "@new",
        "fullName": "John Doe",
        "patientId": "abc123",
        "employer": "Example Corp",
        "reason": "choroba",
        "issued": "2024-11-08",
        "validFrom": "2024-11-08",
        "validUntil": "2024-11-08",
		"checkUp": "2024-11-08"
    }`
	w := MakeCreateRecordRequest(record, engine)
	assert.Equal(t, http.StatusBadRequest, w.Code, "Expected HTTP status 400 Bad Request")
	assert.Contains(t, w.Body.String(), "Key: 'Record.PatientId' Error:Field validation for 'PatientId' failed on the 'only-digits-max-length-10' tag")
}

// Create record ->  Request Body validator -> Reason validator
func TestCreateRecordReasonValidation(t *testing.T) {
	engine := SetupTestEngine()
	record := `{
        "id": "@new",
        "fullName": "John Doe",
        "patientId": "123",
        "employer": "Example Corp",
        "reason": "invalid reason",
        "issued": "2024-11-08",
        "validFrom": "2024-11-08",
        "validUntil": "2024-11-08",
		"checkUp": "2024-11-08"
    }`
	w := MakeCreateRecordRequest(record, engine)
	assert.Equal(t, http.StatusBadRequest, w.Code, "Expected HTTP status 400 Bad Request")
	assert.Contains(t, w.Body.String(), "Key: 'Record.Reason' Error:Field validation for 'Reason' failed on the 'not-valid-reason-value' tag")
}

// Create record -> Request Body validator -> Max Length Validator
func TestCreateRecordMaxLengthValidation(t *testing.T) {
	engine := SetupTestEngine()
	record := `{
        "id": "@new",
        "fullName": "123456789012345678901234567890123456789012345678901",
        "patientId": "123",
        "employer": "Example Corp",
        "reason": "choroba",
        "issued": "2024-11-08",
        "validFrom": "2024-11-08",
        "validUntil": "2024-11-08",
		"checkUp": "2024-11-08"
	}`
	w := MakeCreateRecordRequest(record, engine)
	assert.Equal(t, http.StatusBadRequest, w.Code, "Expected HTTP status 400 Bad Request")
	assert.Contains(t, w.Body.String(), "Key: 'Record.FullName' Error:Field validation for 'FullName' failed on the 'max-length-50' tag")
}

// Create record -> Full name setter
func TestCreateRecordFullNameSetting(t *testing.T) {
	engine := SetupTestEngine()
	record := `{
        "id": "@new",
        "fullName": "",
        "patientId": "999",
        "employer": "Example Corp",
        "reason": "choroba",
        "issued": "2024-11-08",
        "validFrom": "2024-11-08",
        "validUntil": "2024-11-08",
		"checkUp": "2024-11-08"
	}`
	w := MakeCreateRecordRequest(record, engine)
	assert.Equal(t, http.StatusNotFound, w.Code, "Expected HTTP status 404 Not found")
	assert.Contains(t, w.Body.String(), "Patient's PN records not found to inherit Full Name")
}

// Create record -> Full name validator
func TestCreateRecordFullNameValidation(t *testing.T) {
	engine := SetupTestEngine()
	record := `{
        "id": "@new",
        "fullName": "Matus Bojko",
        "patientId": "999",
        "employer": "Example Corp",
        "reason": "choroba",
        "issued": "2024-11-08",
        "validFrom": "2024-11-08",
        "validUntil": "2024-11-08",
		"checkUp": "2024-11-08"
	}`
	w := MakeCreateRecordRequest(record, engine)
	assert.Equal(t, http.StatusCreated, w.Code, "Expected HTTP status 201 Created")

	recordSamePatientDifName := `{
	    "id": "@new",
	    "fullName": "Jozef Bojko",
	    "patientId": "999",
	    "employer": "Example Corp",
	    "reason": "choroba",
	    "issued": "2024-11-11",
	    "validFrom": "2024-11-11",
	    "validUntil": "2024-11-11",
		"checkUp": "2024-11-11"
	}`

	w = MakeCreateRecordRequest(recordSamePatientDifName, engine)
	assert.Equal(t, http.StatusConflict, w.Code, "Expected HTTP status 409 Conflict")
	assert.Contains(t, w.Body.String(), "Full Name does not correspond to patient (conflict with existing records)")
}

// Create record -> Overlap validator
func TestCreateRecordDateOverlapValidation(t *testing.T) {
	engine := SetupTestEngine()
	record1 := `{
		"id": "@new",
		"fullName": "John Doe",
		"patientId": "564",
		"employer": "Example Corp",
		"reason": "choroba",
		"issued": "2024-11-08",
		"validFrom": "2024-11-08",
		"validUntil": "2024-11-08",
		"checkUp": "2024-11-08"
	}`
	w := MakeCreateRecordRequest(record1, engine)
	assert.Equal(t, http.StatusCreated, w.Code, "Expected HTTP status 201 Created")

	w = MakeCreateRecordRequest(record1, engine)
	assert.Equal(t, http.StatusConflict, w.Code, "Expected HTTP status 409 Conflict")
	assert.Contains(t, w.Body.String(), "Patient already has more up-to-date record or their validity overlap")

	t.Cleanup(func() {
		err := testDbService.ClearCollection(context.Background())
		assert.NoError(t, err, "Expected the collection to be cleared after the test")
	})
}
