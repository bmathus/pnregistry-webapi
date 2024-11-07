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
	value, exists := ctx.Get("db_service")
	if !exists {
		ctx.JSON(http.StatusInternalServerError,
			gin.H{
				"status":  "Internal Server Error",
				"message": "db not found",
				"error":   "db not found",
			})
		return
	}

	db, ok := value.(db_service.DbService[models.Record])
	if !ok {
		ctx.JSON(http.StatusInternalServerError,
			gin.H{
				"status":  "Internal Server Error",
				"message": "db_service context is not of type db_service.DbService",
				"error":   "cannot cast db_service context to db_service.DbService",
			})
		return
	}

	updatedRecord := models.Record{}

	// Fields validation
	if err := ctx.ShouldBindJSON(&updatedRecord); err != nil {
		ctx.JSON(http.StatusBadRequest,
			gin.H{
				"status":  "Bad Request",
				"message": "Invalid request body",
				"error":   err.Error(),
			},
		)
		return
	}

	// Validate dates
	if updatedRecord.CheckUp != nil && updatedRecord.ValidFrom.After(*updatedRecord.CheckUp) {
		ctx.JSON(http.StatusBadRequest,
			gin.H{
				"status":  "Bad Request",
				"message": "'Check Up' date can only be on or after 'Valid from' date",
			},
		)
		return
	}
	if updatedRecord.ValidFrom.After(updatedRecord.ValidUntil) {
		ctx.JSON(http.StatusBadRequest,
			gin.H{
				"status":  "Bad Request",
				"message": "'Valid until' date can only be on or after 'Valid from' date",
			},
		)
		return
	}

	recordId := ctx.Param("recordId")

	// Ensure the ID in the URL matches the ID in the request body
	if updatedRecord.Id != recordId {
		ctx.JSON(http.StatusBadRequest,
			gin.H{
				"status":  "Bad Request",
				"message": "Record ID in URL does not match ID in request body",
			},
		)
		return
	}

	// Fetching patient's records by patient ID to validate conflict with updated record
	patientRecords, err := db.FindDocuments(ctx, "patientId", updatedRecord.PatientId)

	if err != nil {
		ctx.JSON(http.StatusBadGateway,
			gin.H{
				"status":  "Bad Gateway",
				"message": "Failed to fetch existing records",
				"error":   err.Error(),
			},
		)
		return
	}

	// Full Name validation
	if updatedRecord.FullName == "" { // if fullname not provided
		if len(patientRecords) == 0 { // and no existing records
			ctx.JSON(http.StatusNotFound,
				gin.H{
					"status":  "Not Found",
					"message": "Patient's PN records not found, create new record (with Full Name)",
				},
			)
			return
		}

		// when records are available, set FullName from them
		updatedRecord.FullName = patientRecords[0].FullName
	}

	var recordToUpdate *models.Record //record we are updating but from db
	recordIsLatest := false           //if updated record is latest for patient

	// Filter out updated record from all patient's records
	patientRecords, recordToUpdate, recordIsLatest = utils.FilterUpdatedAndLatest(patientRecords, updatedRecord.Id)

	// Check if FullName matches the existing records
	if len(patientRecords) != 0 && updatedRecord.FullName != patientRecords[0].FullName {
		// allow fullname update if there's only one existing record and the IDs match
		ctx.JSON(http.StatusConflict, gin.H{
			"status":  "Conflict",
			"message": "Cannot update Full Name for this patient's ID (conflict with existing records)",
		})
		return
	}

	// Date validity overlap validation - if record changed patient or patient is the same and record its latest
	if recordToUpdate == nil || (recordToUpdate != nil && recordIsLatest) {
		for _, record := range patientRecords {
			if !updatedRecord.ValidFrom.After(record.ValidUntil) {
				ctx.JSON(http.StatusConflict,
					gin.H{
						"status":  "Conflict",
						"message": "Patient already has more up-to-date record or their validity overlap",
					},
				)
				return
			}
		}
	}

	// If validity dates are updated
	validityDatesChanged := recordToUpdate != nil && (recordToUpdate.ValidFrom != updatedRecord.ValidFrom || recordToUpdate.ValidUntil != updatedRecord.ValidUntil)

	if !recordIsLatest && validityDatesChanged {
		ctx.JSON(http.StatusBadRequest,
			gin.H{
				"status":  "Bad Request",
				"message": "Validity dates of not the latest PN record can not be updated",
			},
		)
		return
	}

	err = db.UpdateDocument(ctx, recordId, &updatedRecord)

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
