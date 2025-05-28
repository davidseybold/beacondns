package api

import (
	"strings"

	"github.com/davidseybold/beacondns/internal/model"
)

func convertAPIResourceRecordSetToModel(recordSet *ResourceRecordSet) *model.ResourceRecordSet {
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

func convertModelResponsePolicyToAPI(policy *model.ResponsePolicy) *ResponsePolicy {
	return &ResponsePolicy{
		ID:          policy.ID.String(),
		Name:        policy.Name,
		Description: policy.Description,
		Priority:    policy.Priority,
		Enabled:     policy.Enabled,
	}
}

func convertModelResponsePoliciesToAPI(policies []model.ResponsePolicy) []ResponsePolicy {
	var apiPolicies []ResponsePolicy
	for _, policy := range policies {
		apiPolicies = append(apiPolicies, *convertModelResponsePolicyToAPI(&policy))
	}
	return apiPolicies
}

func convertAPIResourceRecordSetsToModel(records []ResourceRecordSet) []model.ResourceRecordSet {
	var modelRecords []model.ResourceRecordSet
	for _, record := range records {
		modelRecords = append(modelRecords, *convertAPIResourceRecordSetToModel(&record))
	}
	return modelRecords
}

func convertModelResponsePolicyRuleToAPI(rule *model.ResponsePolicyRule) *ResponsePolicyRule {
	return &ResponsePolicyRule{
		Name:         rule.Name,
		TriggerType:  string(rule.TriggerType),
		TriggerValue: rule.TriggerValue,
		ActionType:   string(rule.ActionType),
		LocalData:    convertModelResourceRecordSetsToAPI(rule.LocalData),
	}
}

func convertModelResourceRecordSetsToAPI(records []model.ResourceRecordSet) []ResourceRecordSet {
	var apiRecords []ResourceRecordSet
	for _, record := range records {
		apiRecords = append(apiRecords, *convertModelResourceRecordSetToAPI(&record))
	}
	return apiRecords
}

func convertModelResponsePolicyRulesToAPI(rules []model.ResponsePolicyRule) []ResponsePolicyRule {
	var apiRules []ResponsePolicyRule
	for _, rule := range rules {
		apiRules = append(apiRules, *convertModelResponsePolicyRuleToAPI(&rule))
	}
	return apiRules
}
