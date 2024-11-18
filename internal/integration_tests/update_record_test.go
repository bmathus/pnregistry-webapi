package integrationtests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

func MakeUpdateRecordRequest(id string, recordBytes []byte, engine *gin.Engine) *httptest.ResponseRecorder {
	url := fmt.Sprintf("/api/records/%s/", id)
	req, _ := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(recordBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	return w
}

func TestUpdateRecordNoDbService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	pn_registry.AddRoutes(engine)

	w := MakeUpdateRecordRequest("0", nil, engine)

	assert.Equal(t, http.StatusInternalServerError, w.Code, "Expected HTTP status 500 Internal Server Error")

	expectedResponse := `{
	"error":"db not found",
	"message":"db not found",
	"status":"Internal Server Error"}`
	assert.JSONEq(t, expectedResponse, w.Body.String(), "Expected JSON response does not match")
}

func TestUpdateRecordWrongDbService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	engine.Use(func(ctx *gin.Context) {
		ctx.Set("db_service", "incorrectType")
		ctx.Next()
	})
	pn_registry.AddRoutes(engine)

	w := MakeUpdateRecordRequest("0", nil, engine)

	expectedResponse := `{
	"error":"cannot cast db_service context to db_service.DbService",
	"message":"db_service context is not of type db_service.DbService",
	"status":"Internal Server Error"}`

	assert.Equal(t, http.StatusInternalServerError, w.Code, "Expected HTTP status 500 Internal Server Error")
	assert.JSONEq(t, expectedResponse, w.Body.String(), "Expected JSON response does not match")
}

// Update record -> Request Body validator (missing field)
func TestUpdateRecordRequestBodyValidation(t *testing.T) {
	engine := SetupTestEngine()

	initialRecord := models.Record{
		Id:         "1",
		FullName:   "John Doe",
		PatientId:  "12345678",
		Employer:   "Example Corp",
		Reason:     "choroba",
		Issued:     models.DateType(time.Date(2024, 11, 8, 0, 0, 0, 0, time.UTC)),
		ValidFrom:  models.DateType(time.Date(2024, 11, 8, 0, 0, 0, 0, time.UTC)),
		ValidUntil: models.DateType(time.Date(2024, 11, 9, 0, 0, 0, 0, time.UTC)),
		CheckUp:    nil,
	}

	err := testDbService.CreateDocument(context.Background(), initialRecord.Id, &initialRecord)
	assert.NoError(t, err, "Expected record to be created successfully")

	initialRecord.Employer = ""

	updatedRecordBytes, _ := json.Marshal(initialRecord)

	w := MakeUpdateRecordRequest("1", updatedRecordBytes, engine)

	assert.Equal(t, http.StatusBadRequest, w.Code, "Expected HTTP status 400 Bad Request")
	assert.Contains(t, w.Body.String(), "Key: 'Record.Employer' Error:Field validation for 'Employer' failed on the 'required' tag")

	// Clean up the database after the test
	t.Cleanup(func() {
		err := testDbService.ClearCollection(context.Background())
		assert.NoError(t, err, "Expected the collection to be cleared after the test")
	})
}

// UpdateRecord -> Route validation (if route ID is the same as ID in record body)
func TestUpdateRecordRouteValidation(t *testing.T) {
	engine := SetupTestEngine()
	initialRecord := models.Record{
		Id:         "1",
		FullName:   "John Doe",
		PatientId:  "12345678",
		Employer:   "Example Corp",
		Reason:     "choroba",
		Issued:     models.DateType(time.Date(2024, 11, 8, 0, 0, 0, 0, time.UTC)),
		ValidFrom:  models.DateType(time.Date(2024, 11, 8, 0, 0, 0, 0, time.UTC)),
		ValidUntil: models.DateType(time.Date(2024, 11, 9, 0, 0, 0, 0, time.UTC)),
		CheckUp:    nil,
	}

	recordBytes, _ := json.Marshal(initialRecord)
	w := MakeUpdateRecordRequest("2", recordBytes, engine)
	assert.Equal(t, http.StatusBadRequest, w.Code, "Expected HTTP status 400 Bad Request")
	assert.Contains(t, w.Body.String(), "Record ID in route does not match ID in request body")
}

// Update record -> Find Documents (connection fail)
func TestUpdateRecordConnectionFailed(t *testing.T) {
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

	record := models.Record{
		Id:         "1",
		FullName:   "John Doe",
		PatientId:  "12345678",
		Employer:   "Example Corp",
		Reason:     "choroba",
		Issued:     models.DateType(time.Date(2024, 11, 8, 0, 0, 0, 0, time.UTC)),
		ValidFrom:  models.DateType(time.Date(2024, 11, 8, 0, 0, 0, 0, time.UTC)),
		ValidUntil: models.DateType(time.Date(2024, 11, 9, 0, 0, 0, 0, time.UTC)),
		CheckUp:    nil,
	}
	recordBytes, _ := json.Marshal(record)
	w := MakeUpdateRecordRequest("1", recordBytes, engine)
	assert.Equal(t, http.StatusBadGateway, w.Code, "Expected HTTP status 502 Bad Gateway")
	assert.Contains(t, w.Body.String(), "connection() error occurred during connection handshake: auth error")

	testDbService = originalDbService
}

// Update record -> Full name setter
func TestUpdateRecordFullNameSetting(t *testing.T) {
	engine := SetupTestEngine()
	record := models.Record{
		Id:         "1",
		FullName:   "",
		PatientId:  "12345678",
		Employer:   "Example Corp",
		Reason:     "choroba",
		Issued:     models.DateType(time.Date(2024, 11, 8, 0, 0, 0, 0, time.UTC)),
		ValidFrom:  models.DateType(time.Date(2024, 11, 8, 0, 0, 0, 0, time.UTC)),
		ValidUntil: models.DateType(time.Date(2024, 11, 9, 0, 0, 0, 0, time.UTC)),
		CheckUp:    nil,
	}
	recordBytes, _ := json.Marshal(record)
	w := MakeUpdateRecordRequest("1", recordBytes, engine)
	assert.Equal(t, http.StatusNotFound, w.Code, "Expected HTTP status 404 Not found")
	assert.Contains(t, w.Body.String(), "Patient's PN records not found to inherit Full Name")
}

// Update record -> Full name validation
func TestUpdateRecordFullNameValidation(t *testing.T) {
	engine := SetupTestEngine()
	record := models.Record{
		Id:         "1",
		FullName:   "Matus Bojko",
		PatientId:  "12345678",
		Employer:   "Example Corp",
		Reason:     "choroba",
		Issued:     models.DateType(time.Date(2024, 11, 8, 0, 0, 0, 0, time.UTC)),
		ValidFrom:  models.DateType(time.Date(2024, 11, 8, 0, 0, 0, 0, time.UTC)),
		ValidUntil: models.DateType(time.Date(2024, 11, 9, 0, 0, 0, 0, time.UTC)),
		CheckUp:    nil,
	}
	err := testDbService.CreateDocument(context.Background(), record.Id, &record)
	assert.NoError(t, err, "Expected record to be created successfully")

	record.Id = "2"
	record.Issued = models.DateType(time.Date(2024, 11, 11, 0, 0, 0, 0, time.UTC))
	record.ValidFrom = models.DateType(time.Date(2024, 11, 11, 0, 0, 0, 0, time.UTC))
	record.ValidUntil = models.DateType(time.Date(2024, 11, 11, 0, 0, 0, 0, time.UTC))

	err = testDbService.CreateDocument(context.Background(), record.Id, &record)
	assert.NoError(t, err, "Expected record to be created successfully")

	record.FullName = "Jozef Bojko"
	recordBytes, _ := json.Marshal(record)
	w := MakeUpdateRecordRequest("2", recordBytes, engine)

	assert.Equal(t, http.StatusConflict, w.Code, "Expected HTTP status 409 Conflict")
	assert.Contains(t, w.Body.String(), "Full Name does not correspond to patient (conflict with existing records)")

	// Clean up the database after the test
	t.Cleanup(func() {
		err := testDbService.ClearCollection(context.Background())
		assert.NoError(t, err, "Expected the collection to be cleared after the test")
	})
}

// Update record -> Overlap validator
func TestUpdateRecordDateOverlapValidation(t *testing.T) {
	engine := SetupTestEngine()
	record := models.Record{
		Id:         "1",
		FullName:   "Matus Bojko",
		PatientId:  "12345678",
		Employer:   "Example Corp",
		Reason:     "choroba",
		Issued:     models.DateType(time.Date(2024, 11, 8, 0, 0, 0, 0, time.UTC)),
		ValidFrom:  models.DateType(time.Date(2024, 11, 8, 0, 0, 0, 0, time.UTC)),
		ValidUntil: models.DateType(time.Date(2024, 11, 9, 0, 0, 0, 0, time.UTC)),
		CheckUp:    nil,
	}
	err := testDbService.CreateDocument(context.Background(), record.Id, &record)
	assert.NoError(t, err, "Expected record to be created successfully")

	record.Id = "2"

	recordBytes, _ := json.Marshal(record)
	w := MakeUpdateRecordRequest("2", recordBytes, engine)
	assert.Equal(t, http.StatusConflict, w.Code, "Expected HTTP status 409 Conflict")
	assert.Contains(t, w.Body.String(), "Patient already has more up-to-date record or their validity overlap")

	// Clean up the database after the test
	t.Cleanup(func() {
		err := testDbService.ClearCollection(context.Background())
		assert.NoError(t, err, "Expected the collection to be cleared after the test")
	})
}

// Update record -> not lates pn record date change
func TestUpdateRecordDateChangeNotLatestRecordValidation(t *testing.T) {
	engine := SetupTestEngine()
	record := models.Record{
		Id:         "1",
		FullName:   "John Doe",
		PatientId:  "12345678",
		Employer:   "Example Corp",
		Reason:     "choroba",
		Issued:     models.DateType(time.Date(2024, 11, 8, 0, 0, 0, 0, time.UTC)),
		ValidFrom:  models.DateType(time.Date(2024, 11, 8, 0, 0, 0, 0, time.UTC)),
		ValidUntil: models.DateType(time.Date(2024, 11, 9, 0, 0, 0, 0, time.UTC)),
		CheckUp:    nil,
	}

	recordLater := models.Record{
		Id:         "2",
		FullName:   "John Doe",
		PatientId:  "12345678",
		Employer:   "Example Corp",
		Reason:     "choroba",
		Issued:     models.DateType(time.Date(2024, 12, 8, 0, 0, 0, 0, time.UTC)),
		ValidFrom:  models.DateType(time.Date(2024, 12, 8, 0, 0, 0, 0, time.UTC)),
		ValidUntil: models.DateType(time.Date(2024, 12, 9, 0, 0, 0, 0, time.UTC)),
		CheckUp:    nil,
	}

	err := testDbService.CreateDocument(context.Background(), record.Id, &record)
	assert.NoError(t, err, "Expected record to be created successfully")

	err = testDbService.CreateDocument(context.Background(), recordLater.Id, &recordLater)
	assert.NoError(t, err, "Expected record to be created successfully")

	record.ValidUntil = models.DateType(time.Date(2024, 11, 15, 0, 0, 0, 0, time.UTC))

	recordBytes, _ := json.Marshal(record)
	w := MakeUpdateRecordRequest("1", recordBytes, engine)

	assert.Equal(t, http.StatusBadRequest, w.Code, "Expected HTTP status 400 Bad request")
	assert.Contains(t, w.Body.String(), "Validity dates of not the latest PN record can not be updated")

	// Clean up the database after the test
	t.Cleanup(func() {
		err := testDbService.ClearCollection(context.Background())
		assert.NoError(t, err, "Expected the collection to be cleared after the test")
	})
}

// Update record -> Update document (record not found)
func TestUpdateRecordNotFound(t *testing.T) {
	engine := SetupTestEngine()
	record := models.Record{
		Id:         "1",
		FullName:   "John Doe",
		PatientId:  "12345678",
		Employer:   "Example Corp",
		Reason:     "choroba",
		Issued:     models.DateType(time.Date(2024, 11, 8, 0, 0, 0, 0, time.UTC)),
		ValidFrom:  models.DateType(time.Date(2024, 11, 8, 0, 0, 0, 0, time.UTC)),
		ValidUntil: models.DateType(time.Date(2024, 11, 9, 0, 0, 0, 0, time.UTC)),
		CheckUp:    nil,
	}

	recordBytes, _ := json.Marshal(record)
	w := MakeUpdateRecordRequest("1", recordBytes, engine)
	assert.Equal(t, http.StatusNotFound, w.Code, "Expected HTTP status 404 Not found")
	assert.Contains(t, w.Body.String(), "Record with specified ID not found")
}

// Update record -> FindDocuments -> Update Document)
func TestUpdateRecord(t *testing.T) {
	engine := SetupTestEngine()
	initialRecord := models.Record{
		Id:         "1",
		FullName:   "John Doe",
		PatientId:  "12345678",
		Employer:   "Example Corp",
		Reason:     "choroba",
		Issued:     models.DateType(time.Date(2024, 11, 8, 0, 0, 0, 0, time.UTC)),
		ValidFrom:  models.DateType(time.Date(2024, 11, 8, 0, 0, 0, 0, time.UTC)),
		ValidUntil: models.DateType(time.Date(2024, 11, 9, 0, 0, 0, 0, time.UTC)),
		CheckUp:    nil,
	}
	err := testDbService.CreateDocument(context.Background(), initialRecord.Id, &initialRecord)
	assert.NoError(t, err, "Failed to seed initial record")

	checkUp := models.DateType(time.Date(2024, 11, 12, 0, 0, 0, 0, time.UTC))
	initialRecord.FullName = "Mathew Doe"
	initialRecord.PatientId = "87654321"
	initialRecord.Employer = "Updated Corp"
	initialRecord.Reason = "uraz"
	initialRecord.Issued = models.DateType(time.Date(2024, 11, 10, 0, 0, 0, 0, time.UTC))
	initialRecord.ValidFrom = models.DateType(time.Date(2024, 11, 10, 0, 0, 0, 0, time.UTC))
	initialRecord.ValidUntil = models.DateType(time.Date(2024, 11, 11, 0, 0, 0, 0, time.UTC))
	initialRecord.CheckUp = &checkUp

	updatedRecordBytes, _ := json.Marshal(initialRecord)
	w := MakeUpdateRecordRequest("1", updatedRecordBytes, engine)
	var returnedRecord models.Record
	err = json.Unmarshal(w.Body.Bytes(), &returnedRecord)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code, "Expected HTTP status 200 Ok")
	assert.Equal(t, initialRecord, returnedRecord, "Returned record expected to match the initial record")

	storedRecord, err := testDbService.FindDocument(context.Background(), "1")
	assert.NoError(t, err)
	assert.Equal(t, &initialRecord, storedRecord, "Stored record in DB expected to match the initial record")

	// Clean up the database after the test
	t.Cleanup(func() {
		err := testDbService.ClearCollection(context.Background())
		assert.NoError(t, err, "Expected the collection to be cleared after the test")
	})
}
