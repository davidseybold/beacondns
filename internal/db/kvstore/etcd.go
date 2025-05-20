package kvstore

import (
	"context"
	"fmt"
	"time"

	etcdpb "go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/namespace"
)

type EtcdClient struct {
	etcdClient *clientv3.Client
}

var _ KVStore = (*EtcdClient)(nil)

func NewEtcdClient(endpoints []string, scope Scope) (*EtcdClient, error) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	etcdClient.KV = namespace.NewKV(etcdClient.KV, scope.String())
	etcdClient.Watcher = namespace.NewWatcher(etcdClient.Watcher, scope.String())
	etcdClient.Lease = namespace.NewLease(etcdClient.Lease, scope.String())

	return &EtcdClient{etcdClient: etcdClient}, nil
}

func (e *EtcdClient) Get(ctx context.Context, key string) ([]byte, error) {
	resp, err := e.etcdClient.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if resp.Count == 0 || len(resp.Kvs) == 0 {
		return nil, ErrNotFound
	}

	return resp.Kvs[0].Value, nil
}

func (e *EtcdClient) GetPrefix(ctx context.Context, prefix string) ([]Item, error) {
	resp, err := e.etcdClient.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	items := make([]Item, 0, resp.Count)
	for _, kv := range resp.Kvs {
		items = append(items, Item{Key: string(kv.Key), Value: kv.Value})
	}

	return items, nil
}

func (e *EtcdClient) Put(ctx context.Context, key string, value []byte) error {
	_, err := e.etcdClient.Put(ctx, key, string(value))
	return err
}

func (e *EtcdClient) Delete(ctx context.Context, key string) error {
	_, err := e.etcdClient.Delete(ctx, key)
	return err
}

func (e *EtcdClient) DeletePrefix(ctx context.Context, prefix string) error {
	_, err := e.etcdClient.Delete(ctx, prefix, clientv3.WithPrefix())
	return err
}

func (e *EtcdClient) Txn(ctx context.Context, ops []Op) error {
	fmt.Printf("Txn: %v\n", ops)
	etcdOps := make([]clientv3.Op, 0, len(ops))
	for _, o := range ops {
		switch o.Action {
		case ActionPut:
			etcdOps = append(etcdOps, clientv3.OpPut(o.Key, string(o.Value)))
		case ActionDelete:
			etcdOps = append(etcdOps, clientv3.OpDelete(o.Key))
		}
	}

	_, err := e.etcdClient.Txn(ctx).Then(etcdOps...).Commit()

	return err
}

func (e *EtcdClient) Watch(ctx context.Context, key string) (<-chan Event, error) {
	rch := e.etcdClient.Watch(ctx, key)

	ch := make(chan Event)

	go func(in <-chan clientv3.WatchResponse, out chan<- Event) {
		for resp := range in {
			for _, ev := range resp.Events {
				out <- Event{Key: string(ev.Kv.Key), Value: ev.Kv.Value}
			}
		}
	}(rch, ch)

	return ch, nil
}

func (e *EtcdClient) WatchPrefix(ctx context.Context, prefix string) (<-chan Event, error) {
	rch := e.etcdClient.Watch(ctx, prefix, clientv3.WithPrefix())

	ch := make(chan Event)

	go func(in <-chan clientv3.WatchResponse, out chan<- Event) {
		for resp := range in {
			for _, ev := range resp.Events {
				out <- Event{Key: string(ev.Kv.Key), Value: ev.Kv.Value, Type: convertEtcdEventTypeToEventType(ev.Type)}
			}
		}
	}(rch, ch)

	return ch, nil
}

func (e *EtcdClient) Close() error {
	return e.etcdClient.Close()
}

var etcdEventTypeToEventType = map[etcdpb.Event_EventType]EventType{
	etcdpb.PUT:    EventTypePut,
	etcdpb.DELETE: EventTypeDelete,
}

func convertEtcdEventTypeToEventType(etcdEventType etcdpb.Event_EventType) EventType {
	et, ok := etcdEventTypeToEventType[etcdEventType]
	if !ok {
		return EventTypeUnknown
	}

	return et
}
