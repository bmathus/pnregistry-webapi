package pn_registry

import (
	"slices"
	"time"
)

// Utility function witch filters out and return updated record from all patient's records
func filterUpdatedAndLatest(patientRecords []Record, updatedRecordId string) ([]Record, *Record, bool) {

	var recordToUpdate *Record
	var latestRecord *Record

	filteredRecords := slices.DeleteFunc(patientRecords, func(r Record) bool {

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
