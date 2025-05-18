package kvstore

import (
	"github.com/dgraph-io/badger/v4"
)

type Action int

const (
	ActionPut Action = iota
	ActionDelete
)

type Change struct {
	Action Action
	Key    []byte
	Value  []byte
}

type Item struct {
	Key   []byte
	Value []byte
}

type KVStore interface {
	Get(key []byte) ([]byte, error)
	GetPrefix(prefix []byte) ([]Item, error)
	Put(key, value []byte) error
	Delete(key []byte) error
	DeletePrefix(prefix []byte) error
	BatchChange(changes []Change) error
	Close() error
}

type BadgerKVStore struct {
	db *badger.DB
}

func NewBadgerKVStore(path string) (*BadgerKVStore, error) {
	db, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		return nil, err
	}
	return &BadgerKVStore{db: db}, nil
}

func (b *BadgerKVStore) Close() error {
	return b.db.Close()
}

func (b *BadgerKVStore) Get(key []byte) ([]byte, error) {
	var value []byte
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		value, err = item.ValueCopy(nil)
		return err
	})
	return value, err
}

func (b *BadgerKVStore) Put(key, value []byte) error {
	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
}

func (b *BadgerKVStore) Delete(key []byte) error {
	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

func (b *BadgerKVStore) DeletePrefix(prefix []byte) error {
	return b.db.Update(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := txn.Delete(item.Key())
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (b *BadgerKVStore) BatchChange(changes []Change) error {
	return b.db.Update(func(txn *badger.Txn) error {
		for _, change := range changes {
			switch change.Action {
			case ActionPut:
				err := txn.Set(change.Key, change.Value)
				if err != nil {
					return err
				}
			case ActionDelete:
				err := txn.Delete(change.Key)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (b *BadgerKVStore) GetPrefix(prefix []byte) ([]Item, error) {
	var items []Item
	err := b.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := item.Key()
			value, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			items = append(items, Item{Key: key, Value: value})
		}
		return nil
	})
	return items, err
}
