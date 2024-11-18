package handlers

import (
	"net/http"

	"github.com/bmathus/pnregistry-webapi/internal/db_service"
	"github.com/gin-gonic/gin"
)

// DeleteRecord - Deletes specific PN record
func DeleteRecord(ctx *gin.Context) {
	db, message, err := db_service.GetDatabaseService(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "Internal Server Error", "message": message, "error": err.Error()})
		return
	}

	recordId := ctx.Param("recordId")

	err = db.DeleteDocument(ctx, recordId)

	switch err {
	case nil:
		ctx.AbortWithStatus(http.StatusNoContent)
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
				"message": "Failed to delete record from database",
				"error":   err.Error(),
			})
	}
}
