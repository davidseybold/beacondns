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
	EventTypeRuleCreated              = "firewall.rule.create"
	EventTypeRuleDeleted              = "firewall.rule.delete"
	EventTypeRuleUpdated              = "firewall.rule.update"
)

type DomainListCreatedEvent struct {
	DomainListID uuid.UUID `json:"domainListId"`
	Domains      []string  `json:"domains"`
}

func NewDomainListCreatedEvent(domainListID uuid.UUID, domains []string) *model.Event {
	return model.NewEvent(EventTypeDomainListCreated, &DomainListCreatedEvent{
		DomainListID: domainListID,
		Domains:      domains,
	})
}

type DomainListDeletedEvent struct {
	DomainListID uuid.UUID `json:"domainListId"`
}

func NewDomainListDeletedEvent(domainListID uuid.UUID) *model.Event {
	return model.NewEvent(EventTypeDomainListDeleted, &DomainListDeletedEvent{
		DomainListID: domainListID,
	})
}

type DomainListDomainsAddedEvent struct {
	DomainListID uuid.UUID   `json:"domainListId"`
	Domains      []string    `json:"domains"`
	LinkedRules  []uuid.UUID `json:"linkedRules"`
}

func NewDomainListDomainsAddedEvent(domainListID uuid.UUID, domains []string, linkedRules []uuid.UUID) *model.Event {
	return model.NewEvent(EventTypeDomainListDomainsAdded, &DomainListDomainsAddedEvent{
		DomainListID: domainListID,
		Domains:      domains,
		LinkedRules:  linkedRules,
	})
}

type DomainListDomainsRemovedEvent struct {
	DomainListID uuid.UUID   `json:"domainListId"`
	Domains      []string    `json:"domains"`
	LinkedRules  []uuid.UUID `json:"linkedRules"`
}

func NewDomainListDomainsRemovedEvent(domainListID uuid.UUID, domains []string, linkedRules []uuid.UUID) *model.Event {
	return model.NewEvent(EventTypeDomainListDomainsRemoved, &DomainListDomainsRemovedEvent{
		DomainListID: domainListID,
		Domains:      domains,
		LinkedRules:  linkedRules,
	})
}

type RuleCreatedEvent struct {
	FirewallRule *model.FirewallRule `json:"firewallRule"`
	Domains      []string            `json:"domains"`
}

func NewRuleCreatedEvent(firewallRule *model.FirewallRule, domains []string) *model.Event {
	return model.NewEvent(EventTypeRuleCreated, &RuleCreatedEvent{
		FirewallRule: firewallRule,
		Domains:      domains,
	})
}

type RuleDeletedEvent struct {
	FirewallRule *model.FirewallRule `json:"firewallRule"`
}

func NewRuleDeletedEvent(firewallRule *model.FirewallRule) *model.Event {
	return model.NewEvent(EventTypeRuleDeleted, &RuleDeletedEvent{
		FirewallRule: firewallRule,
	})
}

type RuleUpdatedEvent struct {
	FirewallRule *model.FirewallRule `json:"firewallRule"`
}

func NewRuleUpdatedEvent(firewallRule *model.FirewallRule) *model.Event {
	return model.NewEvent(EventTypeRuleUpdated, &RuleUpdatedEvent{
		FirewallRule: firewallRule,
	})
}
