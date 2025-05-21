package kvstore

import (
	"context"
	"errors"
	"fmt"
)

type KVStore interface {
	Get(ctx context.Context, key string, opts ...Option) ([]Item, error)
	Put(ctx context.Context, key string, value []byte, opts ...Option) error
	Delete(ctx context.Context, key string, opts ...Option) error
	Watch(ctx context.Context, key string, opts ...Option) (<-chan Event, error)
	Close() error
	Txn(ctx context.Context) Transaction
}

type Transaction interface {
	Put(key string, value []byte, opts ...Option) Transaction
	Delete(key string, opts ...Option) Transaction
	Commit() error
}

type options struct {
	Prefix bool
}

type Option func(*options)

func WithPrefix() Option {
	return func(o *options) {
		o.Prefix = true
	}
}

type Action int

const (
	ActionPut Action = iota
	ActionDelete
)

type Op struct {
	Action   Action
	Key      string
	Value    []byte
	IsPrefix bool
}

type Item struct {
	Key   string
	Value []byte
}

type EventType int

const (
	EventTypePut EventType = iota
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

func (s *Scope) Validate() error {
	if s.Namespace == "" {
		return fmt.Errorf("namespace is required")
	}

	return nil
}

func (s *Scope) String() string {
	if s.Namespace == "" {
		return fmt.Sprintf("/%s", s.Keyspace)
	}

	if s.Keyspace == "" {
		return fmt.Sprintf("/%s", s.Namespace)
	}

	return fmt.Sprintf("/%s/%s", s.Namespace, s.Keyspace)
}
