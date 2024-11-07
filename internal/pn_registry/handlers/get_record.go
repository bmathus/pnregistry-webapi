package handlers

import (
	"net/http"

	"github.com/bmathus/pnregistry-webapi/internal/db_service"
	"github.com/bmathus/pnregistry-webapi/internal/pn_registry/models"
	"github.com/gin-gonic/gin"
)

// GetRecord - Provides details about specific PN record
func GetRecord(ctx *gin.Context) {
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

	recordId := ctx.Param("recordId")

	if recordId == "" {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"status":  http.StatusBadRequest,
				"message": "Record ID is required",
			})
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
