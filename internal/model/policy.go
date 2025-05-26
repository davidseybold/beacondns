package model

import "github.com/google/uuid"

type ResponsePolicyRuleTriggerType string

const (
	ResponsePolicyRuleTriggerTypeQNAME    ResponsePolicyRuleTriggerType = "QNAME"
	ResponsePolicyRuleTriggerTypeIP       ResponsePolicyRuleTriggerType = "IP"
	ResponsePolicyRuleTriggerTypeClientIP ResponsePolicyRuleTriggerType = "CLIENT_IP"
	ResponsePolicyRuleTriggerTypeNSDNAME  ResponsePolicyRuleTriggerType = "NSDNAME"
	ResponsePolicyRuleTriggerTypeNSIP     ResponsePolicyRuleTriggerType = "NSIP"
)

type ResponsePolicyRuleActionType string

const (
	ResponsePolicyRuleActionTypeNXDOMAIN  ResponsePolicyRuleActionType = "NXDOMAIN"
	ResponsePolicyRuleActionTypeNODATA    ResponsePolicyRuleActionType = "NODATA"
	ResponsePolicyRuleActionTypePASSTHRU  ResponsePolicyRuleActionType = "PASSTHRU"
	ResponsePolicyRuleActionTypeLOCALDATA ResponsePolicyRuleActionType = "LOCALDATA"
)

type ResponsePolicy struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Priority    uint      `json:"priority"`
	Enabled     bool      `json:"enabled"`
}

type ResponsePolicyRule struct {
	ID           uuid.UUID                     `json:"id"`
	Name         string                        `json:"name"`
	TriggerType  ResponsePolicyRuleTriggerType `json:"triggerType"`
	TriggerValue string                        `json:"triggerValue"`
	ActionType   ResponsePolicyRuleActionType  `json:"actionType"`
	LocalData    []ResourceRecordSet           `json:"localData"`
}
