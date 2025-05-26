package api

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type CreateZoneRequest struct {
	Name string `json:"name" binding:"required"`
}

type Zone struct {
	ID                     string `json:"id"`
	Name                   string `json:"name"`
	ResourceRecordSetCount int    `json:"resourceRecordSetCount"`
}

type ListZonesResponse struct {
	Zones []Zone `json:"zones"`
}

type ListResourceRecordSetsResponse struct {
	ResourceRecordSets []ResourceRecordSet `json:"resourceRecordSets"`
}

type UpsertResourceRecordSetRequest struct {
	ResourceRecordSet
}

type ResourceRecordSet struct {
	Name            string           `json:"name"`
	Type            string           `json:"type"`
	TTL             uint32           `json:"ttl"`
	ResourceRecords []ResourceRecord `json:"resourceRecords"`
}

type ResourceRecord struct {
	Value string `json:"value"`
}

type ResponsePolicy struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Priority    uint   `json:"priority"`
	Enabled     bool   `json:"enabled"`
}

type ResponsePolicyRule struct {
	ID           string              `json:"id"`
	Name         string              `json:"name"`
	TriggerType  string              `json:"triggerType"`
	TriggerValue string              `json:"triggerValue"`
	ActionType   string              `json:"actionType"`
	LocalData    []ResourceRecordSet `json:"localData"`
}

type CreateResponsePolicyRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Priority    uint   `json:"priority"`
	Enabled     bool   `json:"enabled"`
}

type CreateResponsePolicyResponse struct {
	ResponsePolicy
}

type CreateResponsePolicyRuleRequest struct {
	Name         string              `json:"name"`
	TriggerType  string              `json:"triggerType"`
	TriggerValue string              `json:"triggerValue"`
	ActionType   string              `json:"actionType"`
	LocalData    []ResourceRecordSet `json:"localData"`
}

type CreateResponsePolicyRuleResponse struct {
	ResponsePolicyRule
}

type ListResponsePoliciesResponse struct {
	ResponsePolicies []ResponsePolicy `json:"responsePolicies"`
}

type GetResponsePolicyResponse struct {
	ResponsePolicy
}

type ListResponsePolicyRulesResponse struct {
	ResponsePolicyRules []ResponsePolicyRule `json:"responsePolicyRules"`
}

type GetResponsePolicyRuleResponse struct {
	ResponsePolicyRule
}

type UpdateResponsePolicyRequest struct {
	Name         string              `json:"name"`
	TriggerType  string              `json:"triggerType"`
	TriggerValue string              `json:"triggerValue"`
	ActionType   string              `json:"actionType"`
	LocalData    []ResourceRecordSet `json:"localData"`
}

type UpdateResponsePolicyResponse struct {
	ResponsePolicy
}

type UpdateResponsePolicyRuleRequest struct {
	Name         string              `json:"name"`
	TriggerType  string              `json:"triggerType"`
	TriggerValue string              `json:"triggerValue"`
	ActionType   string              `json:"actionType"`
	LocalData    []ResourceRecordSet `json:"localData"`
}

type UpdateResponsePolicyRuleResponse struct {
	ResponsePolicyRule
}
