package dnsstore

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/miekg/dns"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/davidseybold/beacondns/internal/db/kvstore"
	"github.com/davidseybold/beacondns/internal/model"
)

type FirewallReader interface {
	GetFirewallRules(ctx context.Context, ids []uuid.UUID) ([]FirewallRule, error)
	GetFirewallRuleMappings(ctx context.Context) ([]FirewallRuleMapping, error)
	SubscribeToFirewallRuleEvents(ctx context.Context) (<-chan FirewallRuleMappingEvent, error)
}

type FirewallWriter interface {
	PutFirewallRule(ctx context.Context, rule *FirewallRule, domains []string) error
	DeleteFirewallRule(ctx context.Context, id uuid.UUID) error
	UpdateFirewallRule(ctx context.Context, rule *FirewallRule) error
	AddDomainsToFirewallRule(ctx context.Context, ruleID uuid.UUID, domains []string) error
	RemoveDomainsFromFirewallRule(ctx context.Context, ruleID uuid.UUID, domains []string) error
}

type FirewallRule struct {
	ID                uuid.UUID                            `msg:"id"`
	Action            model.FirewallRuleAction             `msg:"action"`
	BlockResponseType *model.FirewallRuleBlockResponseType `msg:"blockResponseType"`
	BlockResponse     []dns.RR                             `msg:"blockResponse"`
	Priority          uint                                 `msg:"priority"`
}

type FirewallRuleMapping struct {
	RuleID uuid.UUID `msg:"ruleId"`
	Domain string    `msg:"domain"`
}

type FirewallStore interface {
	FirewallReader
	FirewallWriter
}

type FirewallRuleMappingEventType string

const (
	FirewallRuleMappingEventTypeCreate FirewallRuleMappingEventType = "CREATE"
	FirewallRuleMappingEventTypeDelete FirewallRuleMappingEventType = "DELETE"
)

type FirewallRuleMappingEvent struct {
	Mapping *FirewallRuleMapping
	Type    FirewallRuleMappingEventType
}

func (s *Store) PutFirewallRule(ctx context.Context, rule *FirewallRule, domains []string) error {
	tx := s.kvstore.Txn(ctx)

	ruleKey := createFirewallRuleKey(rule.ID)
	ruleData, err := marshalFirewallRule(rule)
	if err != nil {
		return err
	}
	tx.Put(ruleKey, ruleData)

	for _, domain := range domains {
		domainKey := createFirewallRuleMappingKey(rule.ID, domain)
		mapping := &FirewallRuleMapping{
			RuleID: rule.ID,
			Domain: domain,
		}
		mappingData, marshalErr := msgpack.Marshal(mapping)
		if marshalErr != nil {
			return marshalErr
		}
		tx.Put(domainKey, mappingData)
	}

	return tx.Commit()
}

func (s *Store) DeleteFirewallRule(ctx context.Context, id uuid.UUID) error {
	tx := s.kvstore.Txn(ctx)

	ruleKey := createFirewallRuleKey(id)
	tx.Delete(ruleKey)

	tx.Delete(createFirewallRuleMappingPrefix(id))

	return tx.Commit()
}

func (s *Store) UpdateFirewallRule(ctx context.Context, rule *FirewallRule) error {
	ruleKey := createFirewallRuleKey(rule.ID)
	ruleData, err := marshalFirewallRule(rule)
	if err != nil {
		return err
	}

	return s.kvstore.Put(ctx, ruleKey, ruleData)
}

func (s *Store) GetFirewallRuleMappings(ctx context.Context) ([]FirewallRuleMapping, error) {
	m, err := s.kvstore.Get(ctx, keyPrefixFirewallRules, kvstore.WithPrefix())
	if err != nil && !errors.Is(err, kvstore.ErrNotFound) {
		return nil, err
	}

	rules := make([]FirewallRuleMapping, 0)
	for _, v := range m {
		rule := FirewallRuleMapping{}
		unmarshalErr := msgpack.Unmarshal(v.Value, &rule)
		if unmarshalErr != nil {
			return nil, unmarshalErr
		}
		rules = append(rules, rule)
	}

	return rules, nil
}

func (s *Store) GetFirewallRules(ctx context.Context, ids []uuid.UUID) ([]FirewallRule, error) {
	keys := make([]string, 0, len(ids))
	for _, id := range ids {
		keys = append(keys, createFirewallRuleKey(id))
	}

	m, err := s.kvstore.GetMany(ctx, keys)
	if err != nil {
		return nil, err
	}

	rules := make([]FirewallRule, 0, len(m))
	for _, item := range m {
		rule := FirewallRule{}
		unmarshalErr := msgpack.Unmarshal(item.Value, &rule)
		if unmarshalErr != nil {
			return nil, unmarshalErr
		}
		rules = append(rules, rule)
	}

	return rules, nil
}

func (s *Store) AddDomainsToFirewallRule(ctx context.Context, ruleID uuid.UUID, domains []string) error {
	tx := s.kvstore.Txn(ctx)

	for _, domain := range domains {
		domainKey := createFirewallRuleMappingKey(ruleID, domain)
		mapping := &FirewallRuleMapping{
			RuleID: ruleID,
			Domain: domain,
		}
		mappingData, marshalErr := msgpack.Marshal(mapping)
		if marshalErr != nil {
			return marshalErr
		}
		tx.Put(domainKey, mappingData)
	}

	return tx.Commit()
}

func (s *Store) RemoveDomainsFromFirewallRule(ctx context.Context, ruleID uuid.UUID, domains []string) error {
	tx := s.kvstore.Txn(ctx)

	for _, domain := range domains {
		domainKey := createFirewallRuleMappingKey(ruleID, domain)
		tx.Delete(domainKey)
	}

	return tx.Commit()
}

func (s *Store) SubscribeToFirewallRuleEvents(ctx context.Context) (<-chan FirewallRuleMappingEvent, error) {
	events := make(chan FirewallRuleMappingEvent, 100)

	kvstoreEvents, err := s.kvstore.Watch(ctx, keyPrefixFirewallRules, kvstore.WithPrefix())
	if err != nil {
		return nil, err
	}

	go func(kvEvents <-chan kvstore.Event, ruleEvents chan<- FirewallRuleMappingEvent) {
		defer close(ruleEvents)

		for {
			select {
			case <-ctx.Done():
				return
			case kvstoreEvent := <-kvEvents:
				var eventType FirewallRuleMappingEventType
				switch kvstoreEvent.Type {
				case kvstore.EventTypePut:
					eventType = FirewallRuleMappingEventTypeCreate
				case kvstore.EventTypeDelete:
					eventType = FirewallRuleMappingEventTypeDelete
				}

				mapping := &FirewallRuleMapping{}
				unmarshalErr := msgpack.Unmarshal(kvstoreEvent.Value, mapping)
				if unmarshalErr != nil {
					continue
				}

				events <- FirewallRuleMappingEvent{
					Mapping: mapping,
					Type:    eventType,
				}
			}
		}
	}(kvstoreEvents, events)

	return events, nil
}
