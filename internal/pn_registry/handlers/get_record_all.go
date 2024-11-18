package handlers

import (
	"net/http"

	"github.com/bmathus/pnregistry-webapi/internal/db_service"
	"github.com/gin-gonic/gin"
)

// GetRecordAll - Provides list of all PN records
func GetRecordAll(ctx *gin.Context) {
	db, message, err := db_service.GetDatabaseService(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "Internal Server Error", "message": message, "error": err.Error()})
		return
	}

	records, err := db.FindDocuments(ctx, "", nil)

	switch err {
	case nil:
		ctx.JSON(
			http.StatusOK,
			records,
		)
	default:
		ctx.JSON(
			http.StatusBadGateway,
			gin.H{
				"status":  "Bad Gateway",
				"message": "Failed to load all records from database",
				"error":   err.Error(),
			},
		)
	}
}
