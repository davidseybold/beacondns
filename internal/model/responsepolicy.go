package model

import (
	"github.com/google/uuid"
)

type ResponsePolicyRuleTriggerType string

const (
	ResponsePolicyRuleTriggerTypeQNAME    ResponsePolicyRuleTriggerType = "QNAME"
	ResponsePolicyRuleTriggerTypeIP       ResponsePolicyRuleTriggerType = "IP"
	ResponsePolicyRuleTriggerTypeClientIP ResponsePolicyRuleTriggerType = "CLIENT_IP"
	ResponsePolicyRuleTriggerTypeNSDNAME  ResponsePolicyRuleTriggerType = "NSDNAME"
	ResponsePolicyRuleTriggerTypeNSIP     ResponsePolicyRuleTriggerType = "NSIP"
)

var validResponsePolicyRuleTriggerTypes = map[ResponsePolicyRuleTriggerType]struct{}{
	ResponsePolicyRuleTriggerTypeQNAME:    {},
	ResponsePolicyRuleTriggerTypeIP:       {},
	ResponsePolicyRuleTriggerTypeClientIP: {},
	ResponsePolicyRuleTriggerTypeNSDNAME:  {},
	ResponsePolicyRuleTriggerTypeNSIP:     {},
}

func (r ResponsePolicyRuleTriggerType) String() string {
	return string(r)
}

func (r ResponsePolicyRuleTriggerType) IsValid() bool {
	if _, ok := validResponsePolicyRuleTriggerTypes[r]; !ok {
		return false
	}
	return true
}

type ResponsePolicyRuleActionType string

const (
	ResponsePolicyRuleActionTypeNXDOMAIN  ResponsePolicyRuleActionType = "NXDOMAIN"
	ResponsePolicyRuleActionTypeNODATA    ResponsePolicyRuleActionType = "NODATA"
	ResponsePolicyRuleActionTypePASSTHRU  ResponsePolicyRuleActionType = "PASSTHRU"
	ResponsePolicyRuleActionTypeLOCALDATA ResponsePolicyRuleActionType = "LOCALDATA"
)

var validResponsePolicyRuleActionTypes = map[ResponsePolicyRuleActionType]struct{}{
	ResponsePolicyRuleActionTypeNXDOMAIN:  {},
	ResponsePolicyRuleActionTypeNODATA:    {},
	ResponsePolicyRuleActionTypePASSTHRU:  {},
	ResponsePolicyRuleActionTypeLOCALDATA: {},
}

func (r ResponsePolicyRuleActionType) String() string {
	return string(r)
}

func (r ResponsePolicyRuleActionType) IsValid() bool {
	if _, ok := validResponsePolicyRuleActionTypes[r]; !ok {
		return false
	}
	return true
}

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
