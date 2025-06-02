package model

import (
	"time"

	"github.com/google/uuid"
)

type DomainListInfo struct {
	ID          uuid.UUID   `json:"id"`
	Name        string      `json:"name"`
	IsManaged   bool        `json:"isManaged"`
	SourceURL   *string     `json:"sourceUrl,omitempty"`
	DomainCount int         `json:"domainCount"`
	LinkedRules []uuid.UUID `json:"linkedRules"`
	LastUpdated *time.Time  `json:"lastUpdated,omitempty"`
}

type DomainList struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	IsManaged   bool       `json:"isManaged"`
	SourceURL   *string    `json:"sourceUrl,omitempty"`
	Domains     []string   `json:"domains"`
	LastUpdated *time.Time `json:"lastUpdated,omitempty"`
}

type FirewallRuleAction string

const (
	FirewallRuleActionAllow FirewallRuleAction = "ALLOW"
	FirewallRuleActionAlert FirewallRuleAction = "ALERT"
	FirewallRuleActionBlock FirewallRuleAction = "BLOCK"
)

var ValidFirewallRuleActions = map[FirewallRuleAction]struct{}{
	FirewallRuleActionAllow: {},
	FirewallRuleActionAlert: {},
	FirewallRuleActionBlock: {},
}

type FirewallRuleBlockResponseType string

const (
	FirewallRuleBlockResponseTypeNXDOMAIN FirewallRuleBlockResponseType = "NXDOMAIN"
	FirewallRuleBlockResponseTypeNODATA   FirewallRuleBlockResponseType = "NODATA"
	FirewallRuleBlockResponseTypeOverride FirewallRuleBlockResponseType = "OVERRIDE"
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
