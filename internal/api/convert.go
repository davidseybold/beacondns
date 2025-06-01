package api

import (
	"strings"

	"github.com/davidseybold/beacondns/internal/model"
)

func convertAPIResourceRecordSetToModel(recordSet *ResourceRecordSet) *model.ResourceRecordSet {
	if recordSet == nil {
		return nil
	}

	return &model.ResourceRecordSet{
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

func convertModelResourceRecordSetToAPI(rrSet *model.ResourceRecordSet) *ResourceRecordSet {
	if rrSet == nil {
		return nil
	}

	return &ResourceRecordSet{
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

func convertModelFirewallRuleToAPI(rule *model.FirewallRule) *FirewallRule {
	var blockResponseType *string
	if rule.BlockResponseType != nil {
		t := string(*rule.BlockResponseType)
		blockResponseType = &t
	}

	return &FirewallRule{
		ID:                rule.ID.String(),
		DomainListID:      rule.DomainListID.String(),
		Action:            string(rule.Action),
		BlockResponseType: blockResponseType,
		BlockResponse:     convertModelResourceRecordSetToAPI(rule.BlockResponse),
		Priority:          rule.Priority,
	}
}
