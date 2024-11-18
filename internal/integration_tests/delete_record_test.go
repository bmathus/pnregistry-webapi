package integrationtests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bmathus/pnregistry-webapi/internal/db_service"
	"github.com/bmathus/pnregistry-webapi/internal/pn_registry"
	"github.com/bmathus/pnregistry-webapi/internal/pn_registry/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func MakeDeleteRecordRequest(id string, engine *gin.Engine) *httptest.ResponseRecorder {
	url := fmt.Sprintf("/api/records/%s/", id)
	req, _ := http.NewRequest(http.MethodDelete, url, nil)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	return w
}

func TestDeleteRecordNoDbService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	pn_registry.AddRoutes(engine)

	w := MakeDeleteRecordRequest("0", engine)

	assert.Equal(t, http.StatusInternalServerError, w.Code, "Expected HTTP status 500 Internal Server Error")

	expectedResponse := `{
	"error":"db not found",
	"message":"db not found",
	"status":"Internal Server Error"}`
	assert.JSONEq(t, expectedResponse, w.Body.String(), "Expected JSON response does not match")
}

func TestDeleteRecordWrongDbService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	engine.Use(func(ctx *gin.Context) {
		ctx.Set("db_service", "incorrectType")
		ctx.Next()
	})

	pn_registry.AddRoutes(engine)

	w := MakeDeleteRecordRequest("0", engine)
	expectedResponse := `{
	"error":"cannot cast db_service context to db_service.DbService",
	"message":"db_service context is not of type db_service.DbService",
	"status":"Internal Server Error"}`

	assert.Equal(t, http.StatusInternalServerError, w.Code, "Expected HTTP status 500 Internal Server Error")
	assert.JSONEq(t, expectedResponse, w.Body.String(), "Expected JSON response does not match")
}

func TestDeleteRecord(t *testing.T) {
	engine := SetupTestEngine()
	record := `{
        "id": "@new",
        "fullName": "John Doe",
        "patientId": "777777",
        "employer": "Example Corp",
        "reason": "choroba",
        "issued": "2024-11-08",
        "validFrom": "2024-11-08",
        "validUntil": "2024-11-09",
		"checkUp": "2024-11-08"
    }`
	w := MakeCreateRecordRequest(record, engine)
	var returnedRecord models.Record
	err := json.Unmarshal(w.Body.Bytes(), &returnedRecord)
	assert.NoError(t, err)

	w = MakeDeleteRecordRequest(returnedRecord.Id, engine)
	assert.Equal(t, http.StatusNoContent, w.Code, "Expected HTTP status 204 No content")

	storedRecord, err := testDbService.FindDocument(context.Background(), returnedRecord.Id)
	assert.Nil(t, storedRecord, "Expected no record found (nil)")
	assert.Equal(t, db_service.ErrNotFound, err, "Expected error to be ErrNotFound")
}

func TestDeleteRecordNotFound(t *testing.T) {
	engine := SetupTestEngine()
	w := MakeDeleteRecordRequest("3333", engine)
	assert.Equal(t, http.StatusNotFound, w.Code, "Expected HTTP status 404 Not found")
	assert.Contains(t, w.Body.String(), "Record with specified ID not found")
}
