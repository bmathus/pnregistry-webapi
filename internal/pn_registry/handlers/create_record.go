package handlers

import (
	"net/http"

	"github.com/bmathus/pnregistry-webapi/internal/db_service"
	"github.com/bmathus/pnregistry-webapi/internal/pn_registry/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CreateRecord(ctx *gin.Context) {
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

	newRecord := models.Record{}

	// Fields validation
	if err := ctx.ShouldBindJSON(&newRecord); err != nil {
		ctx.JSON(http.StatusBadRequest,
			gin.H{
				"status":  "Bad Request",
				"message": "Invalid request body",
				"error":   err.Error(),
			},
		)
		return
	}

	// Dates validation
	if newRecord.CheckUp != nil && newRecord.ValidFrom.After(*newRecord.CheckUp) {
		ctx.JSON(http.StatusBadRequest,
			gin.H{
				"status":  "Bad Request",
				"message": "'Check Up' date can only be on or after 'Valid from' date",
			},
		)
		return
	}
	if newRecord.ValidFrom.After(newRecord.ValidUntil) {
		ctx.JSON(http.StatusBadRequest,
			gin.H{
				"status":  "Bad Request",
				"message": "'Valid until' date can only be on or after 'Valid from' date",
			},
		)
		return
	}

	if newRecord.Id == "@new" {
		newRecord.Id = uuid.New().String()
	}

	// Fetching patient's records by patient ID to validate conflict with new record
	patientRecords, err := db.FindDocuments(ctx, "patientId", newRecord.PatientId)

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

	//Full Name validation
	if newRecord.FullName == "" { // is fullName is not provided
		if len(patientRecords) != 0 { // inherit fullname from existing records
			newRecord.FullName = patientRecords[0].FullName
		} else {
			ctx.JSON(http.StatusNotFound,
				gin.H{
					"status":  "Not Found",
					"message": "Patient's PN records not found, provide Full Name",
				},
			)
			return
		}
	}

	if (len(patientRecords) != 0) && newRecord.FullName != patientRecords[0].FullName {
		ctx.JSON(http.StatusConflict,
			gin.H{
				"status":  "Conflict",
				"message": "Full Name does not correspond to patient's ID (conflict with existing records)",
			},
		)
		return
	}

	// Date validity overlap validation
	for _, record := range patientRecords {
		if !newRecord.ValidFrom.After(record.ValidUntil) {
			ctx.JSON(http.StatusConflict,
				gin.H{
					"status":  "Conflict",
					"message": "Patient already has more up-to-date record or their validity overlap",
				},
			)
			return
		}
	}

	// Dreate new record in db
	err = db.CreateDocument(ctx, newRecord.Id, &newRecord)

	switch err {
	case nil:
		ctx.JSON(
			http.StatusCreated,
			newRecord,
		)
	case db_service.ErrConflict:
		ctx.JSON(
			http.StatusConflict,
			gin.H{
				"status":  "Conflict",
				"message": "Record already exists",
				"error":   err.Error(),
			},
		)
	default:
		ctx.JSON(
			http.StatusBadGateway,
			gin.H{
				"status":  "Bad Gateway",
				"message": "Failed to create record in database",
				"error":   err.Error(),
			},
		)
	}
}
