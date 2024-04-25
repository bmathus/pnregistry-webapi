package pn_registry

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Copy following section to separate file, uncomment, and implement accordingly
// CreateRecord - Saves new PN record into list of all PN records
func (this *implPnRegistryRecordsAPI) CreateRecord(ctx *gin.Context) {
	ctx.AbortWithStatus(http.StatusNotImplemented)
}

// DeleteRecord - Deletes specific PN record
func (this *implPnRegistryRecordsAPI) DeleteRecord(ctx *gin.Context) {
	ctx.AbortWithStatus(http.StatusNotImplemented)
}

// GetRecord - Provides details about specific PN record
func (this *implPnRegistryRecordsAPI) GetRecord(ctx *gin.Context) {
	ctx.AbortWithStatus(http.StatusNotImplemented)
}

// GetRecordAll - Provides list of all PN records
func (this *implPnRegistryRecordsAPI) GetRecordAll(ctx *gin.Context) {
	ctx.AbortWithStatus(http.StatusNotImplemented)
}

// UpdateRecord - Updates specific PN record
func (this *implPnRegistryRecordsAPI) UpdateRecord(ctx *gin.Context) {
	ctx.AbortWithStatus(http.StatusNotImplemented)
}
