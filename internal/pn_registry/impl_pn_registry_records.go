package pn_registry

import (
	"net/http"

	"github.com/bmathus/pnregistry-webapi/internal/db_service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateRecord - Saves new PN record into list of all PN records
func (this *implPnRegistryRecordsAPI) CreateRecord(ctx *gin.Context) {
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

	db, ok := value.(db_service.DbService[Record])
	if !ok {
		ctx.JSON(http.StatusInternalServerError,
			gin.H{
				"status":  "Internal Server Error",
				"message": "db_service context is not of type db_service.DbService",
				"error":   "cannot cast db_service context to db_service.DbService",
			})
		return
	}

	newRecord := Record{}

	// Field validation
	if err := ctx.ShouldBindJSON(&newRecord); err != nil {
		ctx.JSON(http.StatusBadRequest,
			gin.H{
				"status":  http.StatusBadRequest,
				"message": "Invalid request body",
				"error":   err.Error(),
			},
		)
		return
	}

	// Validate dates
	if newRecord.CheckUp != nil && newRecord.ValidFrom.After(*newRecord.CheckUp) {
		ctx.JSON(http.StatusBadRequest,
			gin.H{
				"status":  http.StatusBadRequest,
				"message": "Valid from is after CheckUp",
				"error":   "Valid from is after CheckUp",
			},
		)
		return
	}
	if newRecord.ValidFrom.After(newRecord.ValidUntil) {
		ctx.JSON(http.StatusBadRequest,
			gin.H{
				"status":  http.StatusBadRequest,
				"message": "Valid from is after valid until",
				"error":   "Valid from is after valid until",
			},
		)
		return
	}

	if newRecord.Id == "@new" {
		newRecord.Id = uuid.New().String()
	}

	existingRecords, err := db.FindDocuments(ctx, "patientId", newRecord.PatientId)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError,
			gin.H{
				"status":  "Internal Server Error",
				"message": "Failed to fetch existing records",
				"error":   err.Error(),
			},
		)
		return
	}

	//Full Name validation
	if newRecord.FullName == "" { // nemam meno
		if len(existingRecords) != 0 { // existuje pacient id
			newRecord.FullName = existingRecords[0].FullName
		} else {
			ctx.JSON(http.StatusNotFound,
				gin.H{
					"status":  http.StatusBadRequest,
					"message": "Patient records do not exist, provide Full Name",
					"error":   "Patient records do not exist, provide Full Name",
				},
			)
			return
		}
	}

	if (len(existingRecords) != 0) && newRecord.FullName != existingRecords[0].FullName {
		ctx.JSON(http.StatusBadRequest,
			gin.H{
				"status":  http.StatusBadRequest,
				"message": "Patient records already exist, Full Name does not correspond to patient ID",
				"error":   "Patient records already exist, Full Name does not correspond to patient ID",
			},
		)
		return
	}

	// Date overlap validation
	for _, record := range existingRecords {
		if !newRecord.ValidFrom.After(record.ValidUntil) {
			ctx.JSON(http.StatusBadRequest,
				gin.H{
					"status":  http.StatusBadRequest,
					"message": "Patient already has more up-to-date record, or records overlap",
					"error":   "Patient already has more up-to-date record, or records overlap",
				},
			)
			return
		}
	}

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
				"message": "Záznam nebol nájdený",
				"error":   err.Error(),
			},
		)
	default:
		ctx.JSON(
			http.StatusBadGateway,
			gin.H{
				"status":  "Bad Gateway",
				"message": "Nepodarilo sa vymazať záznam z databázy",
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
				"message": "Failed to get all PN records from database",
				"error":   err.Error(),
			},
		)
	}
}

// UpdateRecord - Updates specific PN record
func (this *implPnRegistryRecordsAPI) UpdateRecord(ctx *gin.Context) {
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

	db, ok := value.(db_service.DbService[Record])
	if !ok {
		ctx.JSON(http.StatusInternalServerError,
			gin.H{
				"status":  "Internal Server Error",
				"message": "db_service context is not of type db_service.DbService",
				"error":   "cannot cast db_service context to db_service.DbService",
			})
		return
	}

	updatedRecord := Record{}

	// Field validation
	if err := ctx.ShouldBindJSON(&updatedRecord); err != nil {
		ctx.JSON(http.StatusBadRequest,
			gin.H{
				"status":  http.StatusBadRequest,
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
				"status":  http.StatusBadRequest,
				"message": "Valid from is after CheckUp",
				"error":   "Valid from is after CheckUp",
			},
		)
		return
	}
	if updatedRecord.ValidFrom.After(updatedRecord.ValidUntil) {
		ctx.JSON(http.StatusBadRequest,
			gin.H{
				"status":  http.StatusBadRequest,
				"message": "Valid from is after valid until",
				"error":   "Valid from is after valid until",
			},
		)
		return
	}

	recordId := ctx.Param("recordId")

	// Ensure the ID in the URL matches the ID in the request body
	if updatedRecord.Id != recordId {
		ctx.JSON(http.StatusBadRequest,
			gin.H{
				"status":  http.StatusBadRequest,
				"message": "ID in URL does not match ID in body",
				"error":   "ID in URL does not match ID in body",
			},
		)
		return
	}

	patientRecords, err := db.FindDocuments(ctx, "patientId", updatedRecord.PatientId)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError,
			gin.H{
				"status":  "Internal Server Error",
				"message": "Failed to fetch existing records",
				"error":   err.Error(),
			},
		)
		return
	}

	// Full Name validation
	if updatedRecord.FullName == "" {
		if len(patientRecords) == 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Patient records do not exist, provide Full Name",
				"error":   "Patient records do not exist, provide Full Name",
			})
			return
		}
		// Set FullName from existing record
		updatedRecord.FullName = patientRecords[0].FullName
	}

	var recordToUpdate *Record
	recordIsLatest := false
	patientRecords, recordToUpdate, recordIsLatest = filterUpdatedAndLatest(patientRecords, updatedRecord.Id)

	// Check if FullName matches the existing records
	if len(patientRecords) != 0 && updatedRecord.FullName != patientRecords[0].FullName {
		// Allow update if there's only one existing record and the IDs match
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Patient records already exist, Full Name does not correspond to patient ID",
			"error":   "Patient records already exist, Full Name does not correspond to patient ID",
		})
		return
	}

	// Date validation
	if recordToUpdate == nil || (recordToUpdate != nil && recordIsLatest) {
		for _, record := range patientRecords {
			if !updatedRecord.ValidFrom.After(record.ValidUntil) {
				ctx.JSON(http.StatusBadRequest,
					gin.H{
						"status":  http.StatusBadRequest,
						"message": "Patient already has more up-to-date record, or records overlap",
						"error":   "Patient already has more up-to-date record, or records overlap",
					},
				)
				return
			}
		}
	}

	validityDatesChanged := recordToUpdate != nil && (recordToUpdate.ValidFrom != updatedRecord.ValidFrom || recordToUpdate.ValidUntil != updatedRecord.ValidUntil)

	if !recordIsLatest && validityDatesChanged {
		ctx.JSON(http.StatusBadRequest,
			gin.H{
				"status":  http.StatusBadRequest,
				"message": "Validity dates of older PN records can not be updated, there is more up to date PN record",
				"error":   "Validity dates of older PN records can not be updated, there is more up to date PN record",
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
				"message": "Record not found",
				"error":   err.Error(),
			},
		)
	default:
		ctx.JSON(http.StatusBadGateway,
			gin.H{
				"status":  "Bad Gateway",
				"message": "Failed to update PN record in database",
				"error":   err.Error(),
			})

	}

}
