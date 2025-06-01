package model

import (
	"github.com/google/uuid"
)

type DomainListInfo struct {
	ID          uuid.UUID   `json:"id"`
	Name        string      `json:"name"`
	DomainCount int         `json:"domainCount"`
	LinkedRules []uuid.UUID `json:"linkedRules"`
}

type DomainList struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Domains []string  `json:"domains"`
}

type FirewallRuleAction string

const (
	FirewallRuleActionAllow    FirewallRuleAction = "allow"
	FirewallRuleActionAlert    FirewallRuleAction = "alert"
	FirewallRuleActionOverride FirewallRuleAction = "override"
)

var ValidFirewallRuleActions = map[FirewallRuleAction]struct{}{
	FirewallRuleActionAllow:    {},
	FirewallRuleActionAlert:    {},
	FirewallRuleActionOverride: {},
}

type FirewallRuleBlockResponseType string

const (
	FirewallRuleBlockResponseTypeNXDOMAIN FirewallRuleBlockResponseType = "nxdomain"
	FirewallRuleBlockResponseTypeNODATA   FirewallRuleBlockResponseType = "nodata"
	FirewallRuleBlockResponseTypeOverride FirewallRuleBlockResponseType = "override"
)

var ValidFirewallRuleBlockResponseTypes = map[FirewallRuleBlockResponseType]struct{}{
	FirewallRuleBlockResponseTypeNXDOMAIN: {},
	FirewallRuleBlockResponseTypeNODATA:   {},
	FirewallRuleBlockResponseTypeOverride: {},
}

type FirewallRule struct {
	ID                uuid.UUID                      `json:"id"`
	Name              string                         `json:"name"`
	DomainListID      uuid.UUID                      `json:"domainListId"`
	Action            FirewallRuleAction             `json:"action"`
	BlockResponseType *FirewallRuleBlockResponseType `json:"blockResponseType"`
	BlockResponse     *ResourceRecordSet             `json:"blockResponse"`
	Priority          uint                           `json:"priority"`
}
