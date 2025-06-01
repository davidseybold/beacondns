package dnsstore

import "github.com/davidseybold/beacondns/internal/db/kvstore"

type DNSStore interface {
	ZoneStore
	FirewallStore
}

var _ DNSStore = (*Store)(nil)

type Store struct {
	kvstore kvstore.KVStore
}

func New(kvstore kvstore.KVStore) *Store {
	return &Store{
		kvstore: kvstore,
	}
}
