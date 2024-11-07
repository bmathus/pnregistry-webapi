package utils

import (
	"slices"
	"time"

	"github.com/bmathus/pnregistry-webapi/internal/pn_registry/models"
)

// Utility function witch filters out and return updated record from all patient's records
func FilterUpdatedAndLatest(patientRecords []models.Record, updatedRecordId string) ([]models.Record, *models.Record, bool) {

	var recordToUpdate *models.Record
	var latestRecord *models.Record

	filteredRecords := slices.DeleteFunc(patientRecords, func(r models.Record) bool {

		recordMatch := r.Id == updatedRecordId

		if recordMatch {
			recordToUpdate = &r
		}

		// Track the latest record by ValidUntil field
		if latestRecord == nil || time.Time(r.ValidUntil).After(time.Time(latestRecord.ValidUntil)) {
			latestRecord = &r
		}

		return recordMatch
	})

	isLatestRecordToUpdate := latestRecord != nil && recordToUpdate != nil && latestRecord.Id == recordToUpdate.Id

	return filteredRecords, recordToUpdate, isLatestRecordToUpdate

}
