package pn_registry

import (
	"net/http"

	"github.com/bmathus/pnregistry-webapi/internal/db_service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Copy following section to separate file, uncomment, and implement accordingly
// CreateRecord - Saves new PN record into list of all PN records
func (this *implPnRegistryRecordsAPI) CreateRecord(ctx *gin.Context) {
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

	db, ok := value.(db_service.DbService[Record])
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

	record := Record{}
	err := ctx.BindJSON(&record)
	if err != nil {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"status":  "Bad Request",
				"message": "Invalid request body",
				"error":   err.Error(),
			})
		return
	}

	if record.Id == "" {
		record.Id = uuid.New().String()
	}

	err = db.CreateDocument(ctx, record.Id, &record)

	switch err {
	case nil:
		ctx.JSON(
			http.StatusCreated,
			record,
		)
	case db_service.ErrConflict:
		ctx.JSON(
			http.StatusConflict,
			gin.H{
				"status":  "Conflict",
				"message": "PN record already exists",
				"error":   err.Error(),
			},
		)
	default:
		ctx.JSON(
			http.StatusBadGateway,
			gin.H{
				"status":  "Bad Gateway",
				"message": "Failed to create PN record in database",
				"error":   err.Error(),
			},
		)
	}
}

// DeleteRecord - Deletes specific PN record
func (this *implPnRegistryRecordsAPI) DeleteRecord(ctx *gin.Context) {
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

	db, ok := value.(db_service.DbService[Record])
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

	err := db.DeleteDocument(ctx, recordId)

	switch err {
	case nil:
		ctx.AbortWithStatus(http.StatusNoContent)
	case db_service.ErrNotFound:
		ctx.JSON(
			http.StatusNotFound,
			gin.H{
				"status":  "Not Found",
				"message": "Record not found",
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

// GetRecord - Provides details about specific PN record
func (this *implPnRegistryRecordsAPI) GetRecord(ctx *gin.Context) {
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

	db, ok := value.(db_service.DbService[Record])
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
				"message": "Record not found",
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

// GetRecordAll - Provides list of all PN records
func (this *implPnRegistryRecordsAPI) GetRecordAll(ctx *gin.Context) {
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

	db, ok := value.(db_service.DbService[Record])
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

	records, err := db.FindDocuments(ctx, nil, nil)

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
				"message": "Failed to get all PN records from database",
				"error":   err.Error(),
			},
		)
	}
}

// UpdateRecord - Updates specific PN record
func (this *implPnRegistryRecordsAPI) UpdateRecord(ctx *gin.Context) {
	ctx.AbortWithStatus(http.StatusNotImplemented)
}
