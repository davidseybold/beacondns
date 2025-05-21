package kvstore

import (
	"context"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/namespace"
)

const (
	etcdDialTimeout = 5 * time.Second
	etcdWatchBuffer = 100
)

type EtcdClient struct {
	etcdClient *clientv3.Client
}

type etcdTransaction struct {
	tx  clientv3.Txn
	ops []clientv3.Op
}

var _ KVStore = (*EtcdClient)(nil)
var _ Transaction = (*etcdTransaction)(nil)

func NewEtcdClient(endpoints []string, scope Scope) (*EtcdClient, error) {
	if err := scope.Validate(); err != nil {
		return nil, err
	}

	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: etcdDialTimeout,
	})
	if err != nil {
		return nil, err
	}

	etcdClient.KV = namespace.NewKV(etcdClient.KV, scope.String())
	etcdClient.Watcher = namespace.NewWatcher(etcdClient.Watcher, scope.String())
	etcdClient.Lease = namespace.NewLease(etcdClient.Lease, scope.String())

	return &EtcdClient{etcdClient: etcdClient}, nil
}

func (e *EtcdClient) Get(ctx context.Context, key string, opts ...Option) ([]Item, error) {
	etcdOpts := applyOptions(opts...)

	resp, err := e.etcdClient.Get(ctx, key, etcdOpts...)
	if err != nil {
		return nil, err
	}

	if resp.Count == 0 || len(resp.Kvs) == 0 {
		return nil, ErrNotFound
	}

	items := make([]Item, 0, resp.Count)
	for _, kv := range resp.Kvs {
		items = append(items, Item{Key: string(kv.Key), Value: kv.Value})
	}

	return items, nil
}

func (e *EtcdClient) Put(ctx context.Context, key string, value []byte, opts ...Option) error {
	etcdOpts := applyOptions(opts...)

	_, err := e.etcdClient.Put(ctx, key, string(value), etcdOpts...)
	return err
}

func (e *EtcdClient) Delete(ctx context.Context, key string, opts ...Option) error {
	etcdOpts := applyOptions(opts...)

	_, err := e.etcdClient.Delete(ctx, key, etcdOpts...)
	return err
}

func (e *EtcdClient) Watch(ctx context.Context, key string, opts ...Option) (<-chan Event, error) {
	etcdOpts := applyOptions(opts...)
	rch := e.etcdClient.Watch(ctx, key, etcdOpts...)

	ch := make(chan Event, etcdWatchBuffer)

	go watchLoop(ctx, rch, ch)

	return ch, nil
}

func watchLoop(ctx context.Context, in <-chan clientv3.WatchResponse, out chan<- Event) {
	defer close(out)

	for {
		select {
		case <-ctx.Done():
			return
		case resp, ok := <-in:
			if !ok {
				return
			}
			processWatchResponse(ctx, resp, out)
		}
	}
}

func processWatchResponse(ctx context.Context, resp clientv3.WatchResponse, out chan<- Event) {
	if resp.Err() != nil {
		// TODO: log error
		return
	}

	for _, ev := range resp.Events {
		event := convertEtcdEvent(ev)
		select {
		case out <- event:
		case <-ctx.Done():
			return
		}
	}
}

func convertEtcdEvent(ev *clientv3.Event) Event {
	event := Event{
		Key:   string(ev.Kv.Key),
		Value: ev.Kv.Value,
	}

	switch ev.Type {
	case clientv3.EventTypePut:
		event.Type = EventTypePut
	case clientv3.EventTypeDelete:
		event.Type = EventTypeDelete
	}

	return event
}

func (e *EtcdClient) Close() error {
	return e.etcdClient.Close()
}

func (e *EtcdClient) Txn(ctx context.Context) Transaction {
	return &etcdTransaction{
		tx:  e.etcdClient.Txn(ctx),
		ops: make([]clientv3.Op, 0),
	}
}

func (t *etcdTransaction) Put(key string, value []byte, opts ...Option) Transaction {
	etcdOpts := applyOptions(opts...)

	t.ops = append(t.ops, clientv3.OpPut(key, string(value), etcdOpts...))
	return t
}

func (t *etcdTransaction) Delete(key string, opts ...Option) Transaction {
	etcdOpts := applyOptions(opts...)

	t.ops = append(t.ops, clientv3.OpDelete(key, etcdOpts...))
	return t
}

func (t *etcdTransaction) Commit() error {
	if len(t.ops) == 0 {
		return nil
	}

	_, err := t.tx.Then(t.ops...).Commit()
	return err
}

func applyOptions(opts ...Option) []clientv3.OpOption {
	options := &options{}
	for _, opt := range opts {
		opt(options)
	}

	etcdOpts := make([]clientv3.OpOption, 0)
	if options.Prefix {
		etcdOpts = append(etcdOpts, clientv3.WithPrefix())
	}

	return etcdOpts
}
