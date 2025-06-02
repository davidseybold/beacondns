package firewall

import (
	"github.com/google/uuid"

	"github.com/davidseybold/beacondns/internal/model"
)

const (
	EventTypeDomainListCreated        = "firewall.domainlist.create"
	EventTypeDomainListDeleted        = "firewall.domainlist.delete"
	EventTypeDomainListDomainsAdded   = "firewall.domainlist.addDomains"
	EventTypeDomainListDomainsRemoved = "firewall.domainlist.removeDomains"
	EventTypeDomainListRefreshed      = "firewall.domainlist.refresh"
	EventTypeRuleCreated              = "firewall.rule.create"
	EventTypeRuleDeleted              = "firewall.rule.delete"
	EventTypeRuleUpdated              = "firewall.rule.update"
)

type DomainListCreatedEvent struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	IsManaged   bool      `json:"isManaged"`
	SourceURL   *string   `json:"sourceUrl,omitempty"`
	DomainCount int       `json:"domainCount"`
}

func NewDomainListCreatedEvent(domainList *model.DomainListInfo) *model.Event {
	return model.NewEvent(EventTypeDomainListCreated, &DomainListCreatedEvent{
		ID:          domainList.ID,
		Name:        domainList.Name,
		IsManaged:   domainList.IsManaged,
		SourceURL:   domainList.SourceURL,
		DomainCount: domainList.DomainCount,
	})
}

type DomainListDeletedEvent struct {
	ID uuid.UUID `json:"id"`
}

func NewDomainListDeletedEvent(id uuid.UUID) *model.Event {
	return model.NewEvent(EventTypeDomainListDeleted, &DomainListDeletedEvent{
		ID: id,
	})
}

type DomainListDomainsAddedEvent struct {
	ID      uuid.UUID `json:"id"`
	Domains []string  `json:"domains"`
}

func NewDomainListDomainsAddedEvent(id uuid.UUID, domains []string) *model.Event {
	return model.NewEvent(EventTypeDomainListDomainsAdded, &DomainListDomainsAddedEvent{
		ID:      id,
		Domains: domains,
	})
}

type DomainListDomainsRemovedEvent struct {
	ID      uuid.UUID `json:"id"`
	Domains []string  `json:"domains"`
}

func NewDomainListDomainsRemovedEvent(id uuid.UUID, domains []string) *model.Event {
	return model.NewEvent(EventTypeDomainListDomainsRemoved, &DomainListDomainsRemovedEvent{
		ID:      id,
		Domains: domains,
	})
}

type DomainListRefreshedEvent struct {
	ID uuid.UUID `json:"id"`
}

func NewDomainListRefreshedEvent(id uuid.UUID) *model.Event {
	return model.NewEvent(EventTypeDomainListRefreshed, &DomainListRefreshedEvent{
		ID: id,
	})
}

type RuleCreatedEvent struct {
	ID                uuid.UUID                            `json:"id"`
	Name              string                               `json:"name"`
	DomainListID      uuid.UUID                            `json:"domainListId"`
	Action            model.FirewallRuleAction             `json:"action"`
	BlockResponseType *model.FirewallRuleBlockResponseType `json:"blockResponseType,omitempty"`
	BlockResponse     *model.ResourceRecordSet             `json:"blockResponse,omitempty"`
	Priority          uint                                 `json:"priority"`
}

func NewRuleCreatedEvent(firewallRule *model.FirewallRule) *model.Event {
	return model.NewEvent(EventTypeRuleCreated, &RuleCreatedEvent{
		ID:                firewallRule.ID,
		Name:              firewallRule.Name,
		DomainListID:      firewallRule.DomainListID,
		Action:            firewallRule.Action,
		BlockResponseType: firewallRule.BlockResponseType,
		BlockResponse:     firewallRule.BlockResponse,
		Priority:          firewallRule.Priority,
	})
}

type RuleDeletedEvent struct {
	ID uuid.UUID `json:"id"`
}

func NewRuleDeletedEvent(id uuid.UUID) *model.Event {
	return model.NewEvent(EventTypeRuleDeleted, &RuleDeletedEvent{
		ID: id,
	})
}

type RuleUpdatedEvent struct {
	ID                uuid.UUID                            `json:"id"`
	Name              string                               `json:"name"`
	DomainListID      uuid.UUID                            `json:"domainListId"`
	Action            model.FirewallRuleAction             `json:"action"`
	BlockResponseType *model.FirewallRuleBlockResponseType `json:"blockResponseType,omitempty"`
	BlockResponse     *model.ResourceRecordSet             `json:"blockResponse,omitempty"`
	Priority          uint                                 `json:"priority"`
}

func NewRuleUpdatedEvent(firewallRule *model.FirewallRule) *model.Event {
	return model.NewEvent(EventTypeRuleUpdated, &RuleUpdatedEvent{
		ID:                firewallRule.ID,
		Name:              firewallRule.Name,
		DomainListID:      firewallRule.DomainListID,
		Action:            firewallRule.Action,
		BlockResponseType: firewallRule.BlockResponseType,
		BlockResponse:     firewallRule.BlockResponse,
		Priority:          firewallRule.Priority,
	})
}
