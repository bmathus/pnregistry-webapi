package integrationtests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bmathus/pnregistry-webapi/internal/pn_registry"
	"github.com/bmathus/pnregistry-webapi/internal/pn_registry/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func MakeGetRecordAllRequest(engine *gin.Engine) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(http.MethodGet, "/api/records/", nil)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	return w
}

func TestGetRecordAllNoDbService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	pn_registry.AddRoutes(engine)

	w := MakeGetRecordAllRequest(engine)

	assert.Equal(t, http.StatusInternalServerError, w.Code, "Expected HTTP status 500 Internal Server Error")

	expectedResponse := `{
	"error":"db not found",
	"message":"db not found",
	"status":"Internal Server Error"}`
	assert.JSONEq(t, expectedResponse, w.Body.String(), "Expected JSON response does not match")
}

func TestGetRecordAllWrongDbService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	engine.Use(func(ctx *gin.Context) {
		ctx.Set("db_service", "incorrectType")
		ctx.Next()
	})

	pn_registry.AddRoutes(engine)

	w := MakeGetRecordAllRequest(engine)
	expectedResponse := `{
	"error":"cannot cast db_service context to db_service.DbService",
	"message":"db_service context is not of type db_service.DbService",
	"status":"Internal Server Error"}`

	assert.Equal(t, http.StatusInternalServerError, w.Code, "Expected HTTP status 500 Internal Server Error")
	assert.JSONEq(t, expectedResponse, w.Body.String(), "Expected JSON response does not match")
}

func TestGetRecordAll(t *testing.T) {
	engine := SetupTestEngine()

	// Create sample records directly in the database
	record1 := models.Record{
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

	record2 := models.Record{
		Id:         "2",
		FullName:   "Jane Smith",
		PatientId:  "87654321",
		Employer:   "Another Corp",
		Reason:     "uraz",
		Issued:     models.DateType(time.Date(2024, 11, 9, 0, 0, 0, 0, time.UTC)),
		ValidFrom:  models.DateType(time.Date(2024, 11, 9, 0, 0, 0, 0, time.UTC)),
		ValidUntil: models.DateType(time.Date(2024, 11, 10, 0, 0, 0, 0, time.UTC)),
		CheckUp:    nil,
	}

	// Insert records directly into the database
	err := testDbService.CreateDocument(context.Background(), record1.Id, &record1)
	assert.NoError(t, err, "Expected record1 to be created successfully")

	err = testDbService.CreateDocument(context.Background(), record2.Id, &record2)
	assert.NoError(t, err, "Expected record2 to be created successfully")

	// Perform a GET request to retrieve all records
	w := MakeGetRecordAllRequest(engine)

	// Assert response status and body
	assert.Equal(t, http.StatusOK, w.Code, "Expected HTTP status 200 OK")

	var records []models.Record
	err = json.Unmarshal(w.Body.Bytes(), &records)
	assert.NoError(t, err, "Expected response body to be unmarshaled without error")

	// Verify that the retrieved records match the inserted ones
	assert.Len(t, records, 2, "Expected to retrieve 2 records")
	assert.Contains(t, records, record1, "Expected record1 in the response")
	assert.Contains(t, records, record2, "Expected record2 in the response")

	// Clean up the database after the test
	t.Cleanup(func() {
		err := testDbService.ClearCollection(context.Background())
		assert.NoError(t, err, "Expected the collection to be cleared after the test")
	})
}
