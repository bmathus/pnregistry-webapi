package handlers

import (
	"net/http"

	"github.com/bmathus/pnregistry-webapi/internal/db_service"
	"github.com/bmathus/pnregistry-webapi/internal/pn_registry/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CreateRecord(ctx *gin.Context) {
	db, message, err := utils.GetDatabaseService(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "Internal Server Error", "message": message, "error": err.Error()})
		return
	}

	newRecord, err := utils.RequestBodyValidator(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "Bad Request", "message": "Invalid request body", "error": err.Error()})
		return
	}

	if newRecord.Id == "@new" {
		newRecord.Id = uuid.New().String()
	}

	// Fetching patient's records by patient ID to validate conflict with new record
	patientRecords, err := db.FindDocuments(ctx, "patientId", newRecord.PatientId)
	if err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "Bad Gateway", "message": "Failed to fetch existing records", "error": err.Error()})
		return
	}

	// Full name validation and set
	status, err := utils.FullNameSetterAndValidator(newRecord, patientRecords)
	if err != nil {
		ctx.JSON(status, gin.H{"status": http.StatusText(status), "message": err.Error()})
		return
	}

	// Date validity overlap validation
	err = utils.OverlapValidator(newRecord, patientRecords)
	if err != nil {
		ctx.JSON(http.StatusConflict, gin.H{"status": "Conflict", "message": err.Error()})
		return
	}

	// Dreate new record in db
	err = db.CreateDocument(ctx, newRecord.Id, newRecord)

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
