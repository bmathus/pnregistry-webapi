package handlers

import (
	"net/http"

	"github.com/bmathus/pnregistry-webapi/internal/db_service"
	"github.com/bmathus/pnregistry-webapi/internal/pn_registry/models"
	"github.com/gin-gonic/gin"
)

// GetRecordAll - Provides list of all PN records
func GetRecordAll(ctx *gin.Context) {
	value, exists := ctx.Get("db_service")
	if !exists {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"status":  "Internal Server Error",
				"message": "db not found",
				"error":   "db not found",
			})
		return
	}

	db, ok := value.(db_service.DbService[models.Record])
	if !ok {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"status":  "Internal Server Error",
				"message": "db_service context is not of type db_service.DbService",
				"error":   "cannot cast db_service context to db_service.DbService",
			})
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
