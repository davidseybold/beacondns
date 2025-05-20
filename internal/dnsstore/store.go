package dnsstore

import (
	"context"
	"errors"
	"fmt"

	"github.com/miekg/dns"

	"github.com/davidseybold/beacondns/internal/db/kvstore"
)

const (
	keyPrefixZones = "/zones"
)

var (
	ErrNotFound = errors.New("not found")
)

type DNSStore interface {
	GetRRSet(ctx context.Context, zone string, rrName string, rrType string) ([]dns.RR, error)
	PutRRSet(ctx context.Context, zone string, rrName string, rrType string, rrset []dns.RR) error
	DeleteRRSet(ctx context.Context, zone string, rrName string, rrType string) error
	GetZoneNames(ctx context.Context) ([]string, error)
	DeleteZone(ctx context.Context, zone string) error
	WatchForZoneChanges(ctx context.Context) (<-chan kvstore.Event, error)
	ZoneTxn(ctx context.Context, zone string) ZoneTransaction
}

type ZoneTransaction interface {
	CreateZoneMarker(ctx context.Context) ZoneTransaction
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
		return nil, ErrNotFound
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

func (s *Store) GetZoneNames(ctx context.Context) ([]string, error) {
	items, err := s.kvstore.Get(ctx, keyPrefixZones, kvstore.WithPrefix())
	if err != nil && errors.Is(err, kvstore.ErrNotFound) {
		return nil, nil
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
	tx.Delete(recordSetPrefix)

	return tx.Commit()
}

func (s *Store) WatchForZoneChanges(ctx context.Context) (<-chan kvstore.Event, error) {
	return s.kvstore.Watch(ctx, keyPrefixZones, kvstore.WithPrefix())
}

func (t *zoneTransaction) CreateZoneMarker(ctx context.Context) ZoneTransaction {
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

func createRecordKey(zoneName, rrName string, rrType string) string {
	return fmt.Sprintf("/zone/%s/recordset/%s/%s", zoneName, rrName, rrType)
}

func createZonesKey(zoneName string) string {
	return fmt.Sprintf("%s/%s", keyPrefixZones, zoneName)
}

func createZoneRecordSetPrefix(zoneName string) string {
	return fmt.Sprintf("/zone/%s/recordset", zoneName)
}
