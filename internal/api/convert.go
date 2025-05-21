package api

import "github.com/davidseybold/beacondns/internal/model"

func convertAPIChangeToModel(change Change) model.ResourceRecordSetChange {
	return model.ResourceRecordSetChange{
		Action:            model.RRSetChangeAction(change.Action),
		ResourceRecordSet: convertAPIResourceRecordSetToModel(change.ResourceRecordSet),
	}
}

func convertAPIResourceRecordSetToModel(recordSet ResourceRecordSet) model.ResourceRecordSet {
	return model.ResourceRecordSet{
		Name:            recordSet.Name,
		Type:            model.RRType(recordSet.Type),
		TTL:             recordSet.TTL,
		ResourceRecords: convertAPIResourceRecordsToModel(recordSet.ResourceRecords),
	}
}

func convertAPIResourceRecordsToModel(records []ResourceRecord) []model.ResourceRecord {
	var modelRecords []model.ResourceRecord
	for _, record := range records {
		modelRecords = append(modelRecords, model.ResourceRecord{
			Value: record.Value,
		})
	}
	return modelRecords
}
