package handlers

import (
	"net/http"

	"github.com/bmathus/pnregistry-webapi/internal/db_service"
	"github.com/bmathus/pnregistry-webapi/internal/pn_registry/models"
	"github.com/bmathus/pnregistry-webapi/internal/pn_registry/utils"
	"github.com/gin-gonic/gin"
)

// UpdateRecord - Updates specific PN record
func UpdateRecord(ctx *gin.Context) {
	db, message, err := db_service.GetDatabaseService(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "Internal Server Error", "message": message, "error": err.Error()})
		return
	}

	// Fields validation + Validate dates
	updatedRecord, err := utils.RequestBodyValidator(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "Bad Request", "message": "Invalid request body", "error": err.Error()})
		return
	}

	// Ensure the ID in the URL matches the ID in the request body
	recordId := ctx.Param("recordId")
	if updatedRecord.Id != recordId {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "Bad Request", "message": "Record ID in route does not match ID in request body"})
		return
	}

	// Fetching patient's records by patient ID to validate conflict with updated record
	patientRecords, err := db.FindDocuments(ctx, "patientId", updatedRecord.PatientId)

	if err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "Bad Gateway", "message": "Failed to fetch existing records", "error": err.Error()})
		return
	}

	// Full name setting
	status, err := utils.FullNameSetter(updatedRecord, patientRecords)
	if err != nil {
		ctx.JSON(status, gin.H{"status": http.StatusText(status), "message": err.Error()})
		return
	}

	// Filter out updated record from all patient's records
	var recordToUpdate *models.Record //record we are updating but from db
	recordIsLatest := false           //if updated record is latest for patient
	patientRecords, recordToUpdate, recordIsLatest = utils.FilterUpdatedAndLatest(patientRecords, updatedRecord.Id)

	// Check if FullName matches the existing records
	status, err = utils.FullNameValidator(updatedRecord, patientRecords)
	if err != nil {
		ctx.JSON(status, gin.H{"status": http.StatusText(status), "message": err.Error()})
		return
	}

	// Date validity overlap validation - if record changed patient or patient is the same and record its latest
	if recordToUpdate == nil || (recordToUpdate != nil && recordIsLatest) {
		err = utils.OverlapValidator(updatedRecord, patientRecords)
		if err != nil {
			ctx.JSON(http.StatusConflict, gin.H{"status": "Conflict", "message": err.Error()})
			return
		}
	}

	// If validity dates are updated
	validityDatesChanged := recordToUpdate != nil && (recordToUpdate.ValidFrom != updatedRecord.ValidFrom || recordToUpdate.ValidUntil != updatedRecord.ValidUntil)
	if !recordIsLatest && validityDatesChanged {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "Bad Request", "message": "Validity dates of not the latest PN record can not be updated"})
		return
	}

	err = db.UpdateDocument(ctx, recordId, updatedRecord)

	switch err {
	case nil:
		ctx.JSON(http.StatusOK, updatedRecord)
	case db_service.ErrNotFound:
		ctx.JSON(http.StatusNotFound,
			gin.H{
				"status":  "Not Found",
				"message": "Record with specified ID not found",
				"error":   err.Error(),
			},
		)
	default:
		ctx.JSON(http.StatusBadGateway,
			gin.H{
				"status":  "Bad Gateway",
				"message": "Failed to update record in database",
				"error":   err.Error(),
			})

	}

}
