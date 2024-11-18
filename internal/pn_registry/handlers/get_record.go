package handlers

import (
	"net/http"

	"github.com/bmathus/pnregistry-webapi/internal/db_service"
	"github.com/gin-gonic/gin"
)

// GetRecord - Provides details about specific PN record
func GetRecord(ctx *gin.Context) {
	db, message, err := db_service.GetDatabaseService(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "Internal Server Error", "message": message, "error": err.Error()})
		return
	}

	recordId := ctx.Param("recordId")
	if recordId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "Record ID is required"})
		return
	}

	record, err := db.FindDocument(ctx, recordId)

	switch err {
	case nil:
		ctx.JSON(
			http.StatusOK,
			record,
		)
	case db_service.ErrNotFound:
		ctx.JSON(
			http.StatusNotFound,
			gin.H{
				"status":  "Not Found",
				"message": "Record with specified ID not found",
				"error":   err.Error(),
			},
		)
	default:
		ctx.JSON(
			http.StatusBadGateway,
			gin.H{
				"status":  "Bad Gateway",
				"message": "Failed to load record from database",
				"error":   err.Error(),
			})
	}
}
