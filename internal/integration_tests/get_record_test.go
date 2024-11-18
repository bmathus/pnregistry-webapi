package integrationtests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bmathus/pnregistry-webapi/internal/pn_registry"
	"github.com/bmathus/pnregistry-webapi/internal/pn_registry/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func MakeGetRecordRequest(id string, engine *gin.Engine) *httptest.ResponseRecorder {
	url := fmt.Sprintf("/api/records/%s/", id)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	return w
}

func TestGetRecordNoDbService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	pn_registry.AddRoutes(engine)

	w := MakeGetRecordRequest("0", engine)

	assert.Equal(t, http.StatusInternalServerError, w.Code, "Expected HTTP status 500 Internal Server Error")

	expectedResponse := `{
	"error":"db not found",
	"message":"db not found",
	"status":"Internal Server Error"}`
	assert.JSONEq(t, expectedResponse, w.Body.String(), "Expected JSON response does not match")
}

func TestGetRecordWrongDbService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	engine.Use(func(ctx *gin.Context) {
		ctx.Set("db_service", "incorrectType")
		ctx.Next()
	})

	pn_registry.AddRoutes(engine)

	w := MakeGetRecordRequest("0", engine)
	expectedResponse := `{
	"error":"cannot cast db_service context to db_service.DbService",
	"message":"db_service context is not of type db_service.DbService",
	"status":"Internal Server Error"}`

	assert.Equal(t, http.StatusInternalServerError, w.Code, "Expected HTTP status 500 Internal Server Error")
	assert.JSONEq(t, expectedResponse, w.Body.String(), "Expected JSON response does not match")
}

func TestGetRecord(t *testing.T) {
	engine := SetupTestEngine()

	record := models.Record{
		Id:         "1919",
		FullName:   "John Doe",
		PatientId:  "8877",
		Employer:   "Example Corp",
		Reason:     "choroba",
		Issued:     models.DateType(time.Date(2024, 11, 8, 0, 0, 0, 0, time.UTC)),
		ValidFrom:  models.DateType(time.Date(2024, 11, 8, 0, 0, 0, 0, time.UTC)),
		ValidUntil: models.DateType(time.Date(2024, 11, 9, 0, 0, 0, 0, time.UTC)),
		CheckUp:    nil,
	}

	err := testDbService.CreateDocument(context.Background(), record.Id, &record)
	assert.NoError(t, err, "Expected record to be created successfully")

	w := MakeGetRecordRequest(record.Id, engine)
	assert.Equal(t, http.StatusOK, w.Code, "Expected HTTP status 200 OK")

	var returnedRecord models.Record
	err = json.Unmarshal(w.Body.Bytes(), &returnedRecord)
	assert.NoError(t, err)
	assert.Equal(t, record, returnedRecord)

	// Clean up the database after the test
	t.Cleanup(func() {
		err := testDbService.ClearCollection(context.Background())
		assert.NoError(t, err, "Expected the collection to be cleared after the test")
	})
}
func TestGetRecordNotFound(t *testing.T) {
	engine := SetupTestEngine()
	w := MakeGetRecordRequest("3333", engine)
	assert.Equal(t, http.StatusNotFound, w.Code, "Expected HTTP status 404 Not found")
	assert.Contains(t, w.Body.String(), "Record with specified ID not found")
}
