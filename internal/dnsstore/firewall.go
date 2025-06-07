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
	AddDomainsToFirewallRules(ctx context.Context, ruleIDs []uuid.UUID, domains []string) error
	RemoveDomainsFromFirewallRules(ctx context.Context, ruleIDs []uuid.UUID, domains []string) error
	RefreshDomainsForRules(ctx context.Context, ruleIDs []uuid.UUID, domains []string) error
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
	// Create batches of 100 domains to avoid overwhelming the transaction
	batchSize := 100
	domainBatches := createBatches(domains, batchSize)

	ruleKey := createFirewallRuleKey(rule.ID)
	ruleData, err := marshalFirewallRule(rule)
	if err != nil {
		return err
	}

	err = s.kvstore.Put(ctx, ruleKey, ruleData)
	if err != nil {
		return err
	}

	for _, domainBatch := range domainBatches {
		tx := s.kvstore.Txn(ctx)
		for _, domain := range domainBatch {
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

		err = tx.Commit()
		if err != nil {
			return err
		}
	}

	return nil
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
		rule, unmarshalErr := unmarshalFirewallRule(item.Value)
		if unmarshalErr != nil {
			return nil, unmarshalErr
		}
		rules = append(rules, *rule)
	}

	return rules, nil
}

func (s *Store) AddDomainsToFirewallRules(ctx context.Context, ruleIDs []uuid.UUID, domains []string) error {
	tx := s.kvstore.Txn(ctx)

	for _, ruleID := range ruleIDs {
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
	}

	return tx.Commit()
}

func (s *Store) RemoveDomainsFromFirewallRules(ctx context.Context, ruleIDs []uuid.UUID, domains []string) error {
	tx := s.kvstore.Txn(ctx)

	for _, ruleID := range ruleIDs {
		for _, domain := range domains {
			domainKey := createFirewallRuleMappingKey(ruleID, domain)
			tx.Delete(domainKey)
		}
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

func (s *Store) RefreshDomainsForRules(ctx context.Context, ruleIDs []uuid.UUID, domains []string) error {
	domainBatches := createBatches(domains, 100)

	for _, ruleID := range ruleIDs {
		// Delete all mappings for the rule
		err := s.kvstore.Delete(ctx, createFirewallRuleMappingPrefix(ruleID), kvstore.WithPrefix())
		if err != nil {
			return err
		}

		for _, domain := range domainBatches {
			tx := s.kvstore.Txn(ctx)
			for _, domain := range domain {
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

			err = tx.Commit()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func createBatches[T any](items []T, batchSize int) [][]T {
	batches := make([][]T, 0, (len(items)+batchSize-1)/batchSize)

	for i := 0; i < len(items); i += batchSize {
		end := min(i+batchSize, len(items))
		batches = append(batches, items[i:end])
	}

	for i := 0; i < len(items); i += batchSize {
		end := min(i+batchSize, len(items))
		batches = append(batches, items[i:end])
	}

	return batches
}
