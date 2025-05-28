package dnsstore

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/miekg/dns"

	"github.com/davidseybold/beacondns/internal/db/kvstore"
	"github.com/davidseybold/beacondns/internal/model"
)

var (
	ErrRRSetNotFound = errors.New("rrset not found")
)

type ResponsePolicyRuleMeta struct {
	PolicyID   uuid.UUID `msg:"policyId"`
	PolicyName string    `msg:"policyName"`
}

type ResponsePolicyRule struct {
	model.ResponsePolicyRule
	Priority uint                   `msg:"priority"`
	Meta     ResponsePolicyRuleMeta `msg:"meta"`
}

type DNSStore interface {
	ZoneStore
	ResponsePolicyStore
}

type ResponsePolicyStore interface {
	ResponsePolicyReader
	ResponsePolicyWriter
}

type ResponsePolicyWriter interface {
	PutResponsePolicyRule(ctx context.Context, rule *ResponsePolicyRule) error
}

type ResponsePolicyReader interface {
	SubscribeToResponsePolicyRuleEvents(ctx context.Context) (<-chan ResponsePolicyRuleEvent, error)
	GetAllResponsePolicyRules(ctx context.Context) ([]ResponsePolicyRule, error)
}

type ResponsePolicyRuleEventType string

const (
	ResponsePolicyRuleEventTypeCreate ResponsePolicyRuleEventType = "CREATE"
	ResponsePolicyRuleEventTypeDelete ResponsePolicyRuleEventType = "DELETE"
)

type ResponsePolicyRuleEvent struct {
	Rule *ResponsePolicyRule
	Type ResponsePolicyRuleEventType
}

type ZoneStore interface {
	ZoneReader
	ZoneWriter
}

type ZoneReader interface {
	GetRRSet(ctx context.Context, zone string, rrName string, rrType string) ([]dns.RR, error)
	SubscribeToZoneEvents(ctx context.Context) (<-chan ZoneEvent, error)
	GetAllZoneNames(ctx context.Context) ([]string, error)
}

type ZoneEventType string

const (
	ZoneEventTypeCreate ZoneEventType = "CREATE"
	ZoneEventTypeDelete ZoneEventType = "DELETE"
)

type ZoneEvent struct {
	Zone string
	Type ZoneEventType
}

type ZoneWriter interface {
	PutRRSet(ctx context.Context, zone string, rrName string, rrType string, rrset []dns.RR) error
	DeleteRRSet(ctx context.Context, zone string, rrName string, rrType string) error
	DeleteZone(ctx context.Context, zone string) error
	ZoneTxn(ctx context.Context, zone string) ZoneTransaction
}

type ZoneTransaction interface {
	CreateZoneMarker() ZoneTransaction
	PutRRSet(rrName string, rrType string, rrset []dns.RR) ZoneTransaction
	DeleteRRSet(rrName string, rrType string) ZoneTransaction
	Commit() error
}

type Store struct {
	kvstore kvstore.KVStore
}

type zoneTransaction struct {
	tx   kvstore.Transaction
	zone string
}

var _ DNSStore = (*Store)(nil)
var _ ZoneTransaction = (*zoneTransaction)(nil)

func New(kvstore kvstore.KVStore) *Store {
	return &Store{
		kvstore: kvstore,
	}
}

func (s *Store) ZoneTxn(ctx context.Context, zone string) ZoneTransaction {
	tx := s.kvstore.Txn(ctx)

	return &zoneTransaction{
		tx:   tx,
		zone: zone,
	}
}

func (s *Store) GetRRSet(ctx context.Context, zone string, rrName string, rrType string) ([]dns.RR, error) {
	key := createRecordKey(zone, rrName, rrType)
	val, err := s.kvstore.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if len(val) == 0 {
		return nil, ErrRRSetNotFound
	}

	rrset, err := unmarshalRRSet(val[0].Value)
	if err != nil {
		return nil, err
	}

	return rrset.RRs, nil
}

func (s *Store) PutRRSet(ctx context.Context, zone string, rrName string, rrType string, rrset []dns.RR) error {
	if len(rrset) == 0 {
		return nil
	}

	key := createRecordKey(zone, rrName, rrType)
	rs := &rrSet{
		RRs: rrset,
	}

	val, err := marshalRRSet(rs)
	if err != nil {
		return err
	}

	return s.kvstore.Put(ctx, key, val)
}

func (s *Store) DeleteRRSet(ctx context.Context, zone string, rrName string, rrType string) error {
	key := createRecordKey(zone, rrName, rrType)
	return s.kvstore.Delete(ctx, key)
}

func (s *Store) GetAllZoneNames(ctx context.Context) ([]string, error) {
	items, err := s.kvstore.Get(ctx, keyPrefixZones, kvstore.WithPrefix())
	if err != nil && errors.Is(err, kvstore.ErrNotFound) {
		return []string{}, nil
	} else if err != nil {
		return nil, err
	}

	zoneNames := make([]string, 0, len(items))
	for _, item := range items {
		zoneNames = append(zoneNames, string(item.Value))
	}

	return zoneNames, nil
}

func (s *Store) DeleteZone(ctx context.Context, zone string) error {
	zoneKey := createZonesKey(zone)
	recordSetPrefix := createZoneRecordSetPrefix(zone)

	tx := s.kvstore.Txn(ctx)

	tx.Delete(zoneKey)
	tx.Delete(recordSetPrefix, kvstore.WithPrefix())

	return tx.Commit()
}

func (s *Store) SubscribeToZoneEvents(ctx context.Context) (<-chan ZoneEvent, error) {
	events := make(chan ZoneEvent, 100)

	kvstoreEvents, err := s.kvstore.Watch(ctx, keyPrefixZones, kvstore.WithPrefix())
	if err != nil {
		return nil, err
	}

	go func(kvEvents <-chan kvstore.Event, zoneEvents chan<- ZoneEvent) {
		defer close(zoneEvents)

		for {
			select {
			case <-ctx.Done():
				return
			case kvstoreEvent := <-kvEvents:
				var eventType ZoneEventType
				switch kvstoreEvent.Type {
				case kvstore.EventTypePut:
					eventType = ZoneEventTypeCreate
				case kvstore.EventTypeDelete:
					eventType = ZoneEventTypeDelete
				}

				events <- ZoneEvent{
					Zone: string(kvstoreEvent.Value),
					Type: eventType,
				}
			}
		}
	}(kvstoreEvents, events)

	return events, nil
}

func (t *zoneTransaction) CreateZoneMarker() ZoneTransaction {
	t.tx.Put(createZonesKey(t.zone), []byte(t.zone))
	return t
}

func (t *zoneTransaction) PutRRSet(rrName string, rrType string, rrset []dns.RR) ZoneTransaction {
	if len(rrset) == 0 {
		return t
	}

	key := createRecordKey(t.zone, rrName, rrType)
	rs := &rrSet{
		RRs: rrset,
	}

	val, err := marshalRRSet(rs)
	if err != nil {
		return t
	}

	t.tx.Put(key, val)
	return t
}

func (t *zoneTransaction) DeleteRRSet(rrName string, rrType string) ZoneTransaction {
	key := createRecordKey(t.zone, rrName, rrType)
	t.tx.Delete(key)
	return t
}

func (t *zoneTransaction) Commit() error {
	return t.tx.Commit()
}

func (s *Store) PutResponsePolicyRule(ctx context.Context, rule *ResponsePolicyRule) error {
	key := createResponsePolicyRuleKey(rule)
	val, err := marshalResponsePolicyRule(rule)
	if err != nil {
		return err
	}

	return s.kvstore.Put(ctx, key, val)
}

func (s *Store) GetAllResponsePolicyRules(ctx context.Context) ([]ResponsePolicyRule, error) {
	items, err := s.kvstore.Get(ctx, keyPrefixResponsePolicy, kvstore.WithPrefix())
	if err != nil {
		return nil, err
	}

	rules := make([]ResponsePolicyRule, 0, len(items))
	for _, item := range items {
		rule, unmarshalErr := unmarshalResponsePolicyRule(item.Value)
		if unmarshalErr != nil {
			return nil, unmarshalErr
		}
		rules = append(rules, *rule)
	}

	return rules, nil
}

func (s *Store) DeleteResponsePolicyRule(ctx context.Context, rule *ResponsePolicyRule) error {
	key := createResponsePolicyRuleKey(rule)
	return s.kvstore.Delete(ctx, key)
}

func (s *Store) SubscribeToResponsePolicyRuleEvents(ctx context.Context) (<-chan ResponsePolicyRuleEvent, error) {
	events := make(chan ResponsePolicyRuleEvent, 100)

	kvstoreEvents, err := s.kvstore.Watch(ctx, keyPrefixResponsePolicy, kvstore.WithPrefix())
	if err != nil {
		return nil, err
	}

	go func(kvEvents <-chan kvstore.Event, responsePolicyRuleEvents chan<- ResponsePolicyRuleEvent) {
		defer close(responsePolicyRuleEvents)

		for {
			select {
			case <-ctx.Done():
				return
			case kvstoreEvent := <-kvEvents:
				var eventType ResponsePolicyRuleEventType
				switch kvstoreEvent.Type {
				case kvstore.EventTypePut:
					eventType = ResponsePolicyRuleEventTypeCreate
				case kvstore.EventTypeDelete:
					eventType = ResponsePolicyRuleEventTypeDelete
				}

				rule, unmarshalErr := unmarshalResponsePolicyRule(kvstoreEvent.Value)
				if unmarshalErr != nil {
					continue
				}

				events <- ResponsePolicyRuleEvent{
					Rule: rule,
					Type: eventType,
				}
			}
		}
	}(kvstoreEvents, events)

	return events, nil
}
