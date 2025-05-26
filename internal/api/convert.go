package api

import (
	"strings"

	"github.com/davidseybold/beacondns/internal/model"
)

func convertAPIResourceRecordSetToModel(recordSet ResourceRecordSet) model.ResourceRecordSet {
	return model.ResourceRecordSet{
		Name:            recordSet.Name,
		Type:            model.RRType(strings.ToUpper(recordSet.Type)),
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

func convertModelResourceRecordSetToAPI(rrSet model.ResourceRecordSet) ResourceRecordSet {
	return ResourceRecordSet{
		Name:            rrSet.Name,
		Type:            strings.ToUpper(string(rrSet.Type)),
		TTL:             rrSet.TTL,
		ResourceRecords: convertModelResourceRecordsToAPI(rrSet.ResourceRecords),
	}
}

func convertModelResourceRecordsToAPI(records []model.ResourceRecord) []ResourceRecord {
	var apiRecords []ResourceRecord
	for _, record := range records {
		apiRecords = append(apiRecords, ResourceRecord{
			Value: record.Value,
		})
	}
	return apiRecords
}

func convertModelResponsePolicyToAPI(policy model.ResponsePolicy) ResponsePolicy {
	return ResponsePolicy{
		ID:          policy.ID.String(),
		Name:        policy.Name,
		Description: policy.Description,
		Priority:    policy.Priority,
		Enabled:     policy.Enabled,
	}
}
