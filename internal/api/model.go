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
	Name            string           `json:"name"            binding:"required"`
	Type            string           `json:"type"            binding:"required"`
	TTL             uint32           `json:"ttl"             binding:"required"`
	ResourceRecords []ResourceRecord `json:"resourceRecords" binding:"required,min=1"`
}

type ResourceRecord struct {
	Value string `json:"value" binding:"required"`
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
	LocalData    []ResourceRecordSet `json:"localData,omitempty"`
}

type CreateResponsePolicyRequest struct {
	Name        string `json:"name"        binding:"required"`
	Description string `json:"description" binding:"required"`
	Priority    uint   `json:"priority"    binding:"required"`
}

type CreateResponsePolicyResponse struct {
	*ResponsePolicy
}

type CreateResponsePolicyRuleRequest struct {
	Name         string              `json:"name"         binding:"required"`
	TriggerType  string              `json:"triggerType"  binding:"required,responsePolicyRuleTriggerType"`
	TriggerValue string              `json:"triggerValue" binding:"required"`
	ActionType   string              `json:"actionType"   binding:"required,responsePolicyRuleActionType"`
	LocalData    []ResourceRecordSet `json:"localData"    binding:"responsePolicyRuleLocalData"`
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
	Name        string `json:"name"        binding:"required"`
	Description string `json:"description" binding:"required"`
	Priority    uint   `json:"priority"    binding:"required"`
}

type UpdateResponsePolicyResponse struct {
	ResponsePolicy
}

type UpdateResponsePolicyRuleRequest struct {
	Name         string              `json:"name"`
	TriggerType  string              `json:"triggerType"  binding:"required,responsePolicyRuleTriggerType"`
	TriggerValue string              `json:"triggerValue" binding:"required"`
	ActionType   string              `json:"actionType"   binding:"required,responsePolicyRuleActionType"`
	LocalData    []ResourceRecordSet `json:"localData"    binding:"responsePolicyRuleLocalData"`
}

type UpdateResponsePolicyRuleResponse struct {
	ResponsePolicyRule
}

type ToggleResponsePolicyRequest struct {
	Enabled bool `json:"enabled"`
}

type ToggleResponsePolicyResponse struct {
	ResponsePolicy
}
