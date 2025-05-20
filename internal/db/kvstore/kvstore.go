package kvstore

import (
	"context"
	"errors"
	"fmt"
)

type KVStore interface {
	Get(ctx context.Context, key string) ([]byte, error)
	GetPrefix(ctx context.Context, prefix string) ([]Item, error)
	Put(ctx context.Context, key string, value []byte) error
	Delete(ctx context.Context, key string) error
	DeletePrefix(ctx context.Context, prefix string) error
	Txn(ctx context.Context, ops []Op) error
	Watch(ctx context.Context, key string) (<-chan Event, error)
	WatchPrefix(ctx context.Context, prefix string) (<-chan Event, error)
	Close() error
}

type Action int

const (
	ActionPut Action = iota
	ActionDelete
)

type Op struct {
	Action Action
	Key    string
	Value  []byte
}

type Item struct {
	Key   string
	Value []byte
}

type EventType int

const (
	EventTypeUnknown EventType = iota
	EventTypePut
	EventTypeDelete
)

type Event struct {
	Type  EventType
	Key   string
	Value []byte
}

var (
	ErrNotFound = errors.New("not found")
)

type Scope struct {
	Namespace string
	Keyspace  string
}

func (s *Scope) String() string {
	if s.Namespace == "" {
		return s.Keyspace
	}

	if s.Keyspace == "" {
		return s.Namespace
	}

	return fmt.Sprintf("%s/%s", s.Namespace, s.Keyspace)
}
