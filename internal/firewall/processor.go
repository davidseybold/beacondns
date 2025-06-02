package firewall

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	bdns "github.com/davidseybold/beacondns/internal/dns"
	"github.com/davidseybold/beacondns/internal/dnsstore"
	"github.com/davidseybold/beacondns/internal/model"
	"github.com/davidseybold/beacondns/internal/repository"
)

type EventProcessor struct {
	repository repository.TransactorRegistry
	store      dnsstore.FirewallWriter
	logger     *slog.Logger
}

type EventProcessorDeps struct {
	Repository repository.TransactorRegistry
	DNSStore   dnsstore.FirewallWriter
	Logger     *slog.Logger
}

func (d *EventProcessorDeps) Validate() error {
	if d.Repository == nil {
		return errors.New("repository is required")
	}

	if d.DNSStore == nil {
		return errors.New("dns store is required")
	}

	if d.Logger == nil {
		return errors.New("logger is required")
	}

	return nil
}

func NewEventProcessor(deps *EventProcessorDeps) (*EventProcessor, error) {
	if err := deps.Validate(); err != nil {
		return nil, err
	}

	return &EventProcessor{
		repository: deps.Repository,
		store:      deps.DNSStore,
		logger:     deps.Logger,
	}, nil
}

func (p *EventProcessor) Events() []string {
	return []string{
		EventTypeDomainListDomainsAdded,
		EventTypeDomainListDomainsRemoved,
		EventTypeRuleCreated,
		EventTypeRuleDeleted,
		EventTypeRuleUpdated,
		EventTypeDomainListDeleted,
		EventTypeDomainListCreated,
		EventTypeDomainListRefreshed,
	}
}

func (p *EventProcessor) ProcessEvent(ctx context.Context, event *model.Event) error {
	switch event.Type {
	case EventTypeDomainListDomainsAdded:
		return p.processDomainListDomainsAddedEvent(ctx, event)
	case EventTypeDomainListDomainsRemoved:
		return p.processDomainListDomainsRemovedEvent(ctx, event)
	case EventTypeRuleCreated:
		return p.processFirewallRuleCreatedEvent(ctx, event)
	case EventTypeRuleDeleted:
		return p.processFirewallRuleDeletedEvent(ctx, event)
	case EventTypeRuleUpdated:
		return p.processFirewallRuleUpdatedEvent(ctx, event)
	case EventTypeDomainListDeleted:
		return nil
	case EventTypeDomainListCreated:
		return nil
	case EventTypeDomainListRefreshed:
		return p.processDomainListRefreshedEvent(ctx, event)
	default:
		return fmt.Errorf("unknown event type: %s", event.Type)
	}
}

func (p *EventProcessor) processDomainListDomainsAddedEvent(ctx context.Context, event *model.Event) error {
	var domainListDomainsAddedEvent DomainListDomainsAddedEvent
	if err := json.Unmarshal(event.Payload, &domainListDomainsAddedEvent); err != nil {
		return err
	}

	rules, err := p.repository.GetFirewallRepository().
		GetFirewallRulesByDomainListID(ctx, domainListDomainsAddedEvent.ID)
	if err != nil {
		return err
	}

	ruleIDs := make([]uuid.UUID, 0, len(rules))

	for _, rule := range rules {
		ruleIDs = append(ruleIDs, rule.ID)
	}

	return p.store.AddDomainsToFirewallRules(
		ctx,
		ruleIDs,
		domainListDomainsAddedEvent.Domains,
	)
}

func (p *EventProcessor) processDomainListDomainsRemovedEvent(ctx context.Context, event *model.Event) error {
	var domainListDomainsRemovedEvent DomainListDomainsRemovedEvent
	if err := json.Unmarshal(event.Payload, &domainListDomainsRemovedEvent); err != nil {
		return err
	}

	rules, err := p.repository.GetFirewallRepository().
		GetFirewallRulesByDomainListID(ctx, domainListDomainsRemovedEvent.ID)
	if err != nil {
		return err
	}

	ruleIDs := make([]uuid.UUID, 0, len(rules))
	for _, rule := range rules {
		ruleIDs = append(ruleIDs, rule.ID)
	}

	return p.store.RemoveDomainsFromFirewallRules(
		ctx,
		ruleIDs,
		domainListDomainsRemovedEvent.Domains,
	)
}

func (p *EventProcessor) processFirewallRuleCreatedEvent(ctx context.Context, event *model.Event) error {
	var firewallRuleCreatedEvent RuleCreatedEvent
	if err := json.Unmarshal(event.Payload, &firewallRuleCreatedEvent); err != nil {
		return err
	}

	blockResponse, err := bdns.ParseRRs(firewallRuleCreatedEvent.BlockResponse)
	if err != nil {
		return err
	}

	domains, err := p.repository.GetFirewallRepository().
		GetDomainListDomains(ctx, firewallRuleCreatedEvent.DomainListID)
	if err != nil {
		return err
	}

	rule := &dnsstore.FirewallRule{
		ID:                firewallRuleCreatedEvent.ID,
		Action:            firewallRuleCreatedEvent.Action,
		Priority:          firewallRuleCreatedEvent.Priority,
		BlockResponseType: firewallRuleCreatedEvent.BlockResponseType,
		BlockResponse:     blockResponse,
	}

	return p.store.PutFirewallRule(ctx, rule, domains)
}

func (p *EventProcessor) processFirewallRuleDeletedEvent(ctx context.Context, event *model.Event) error {
	var firewallRuleDeletedEvent RuleDeletedEvent
	if err := json.Unmarshal(event.Payload, &firewallRuleDeletedEvent); err != nil {
		return err
	}

	return p.store.DeleteFirewallRule(ctx, firewallRuleDeletedEvent.ID)
}

func (p *EventProcessor) processFirewallRuleUpdatedEvent(ctx context.Context, event *model.Event) error {
	var firewallRuleUpdatedEvent RuleUpdatedEvent
	if err := json.Unmarshal(event.Payload, &firewallRuleUpdatedEvent); err != nil {
		return err
	}

	blockResponse, err := bdns.ParseRRs(firewallRuleUpdatedEvent.BlockResponse)
	if err != nil {
		return err
	}

	rule := &dnsstore.FirewallRule{
		ID:                firewallRuleUpdatedEvent.ID,
		Action:            firewallRuleUpdatedEvent.Action,
		Priority:          firewallRuleUpdatedEvent.Priority,
		BlockResponseType: firewallRuleUpdatedEvent.BlockResponseType,
		BlockResponse:     blockResponse,
	}

	return p.store.UpdateFirewallRule(ctx, rule)
}

func (p *EventProcessor) processDomainListRefreshedEvent(ctx context.Context, event *model.Event) error {
	var domainListRefreshedEvent DomainListRefreshedEvent
	if err := json.Unmarshal(event.Payload, &domainListRefreshedEvent); err != nil {
		return err
	}

	domains, err := p.repository.GetFirewallRepository().GetDomainListDomains(ctx, domainListRefreshedEvent.ID)
	if err != nil {
		return err
	}

	rules, err := p.repository.GetFirewallRepository().GetFirewallRulesByDomainListID(ctx, domainListRefreshedEvent.ID)
	if err != nil {
		return err
	}

	ruleIDs := make([]uuid.UUID, 0, len(rules))
	for _, rule := range rules {
		ruleIDs = append(ruleIDs, rule.ID)
	}

	return p.store.RefreshDomainsForRules(ctx, ruleIDs, domains)
}
