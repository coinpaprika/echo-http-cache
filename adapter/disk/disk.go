package disk

import (
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v3"
)

type (
	Adapter struct {
		capacity  int
		directory string
		db        *badger.DB
	}

	AdapterOptions func(a *Adapter) error
)

func NewAdapter(opts ...AdapterOptions) (*Adapter, error) {
	a := &Adapter{
		directory: "./badger",
	}

	for _, opt := range opts {
		if err := opt(a); err != nil {
			return nil, err
		}
	}

	badgerOpts := badger.DefaultOptions(a.directory)
	badgerOpts.IndexCacheSize = 100_000_000 // 100 Mb
	db, err := badger.Open(badgerOpts)
	if err != nil {
		return nil, err
	}
	a.db = db

	return a, nil
}

func WithDirectory(directory string) AdapterOptions {
	return func(a *Adapter) error {
		a.directory = directory
		return nil
	}
}

func (a *Adapter) Get(key uint64) ([]byte, bool) {
	var response []byte

	err := a.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(a.key(key))
		if err != nil {
			return err
		}

		b, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		response = b
		return nil
	})

	if err != nil {
		return nil, false
	}

	return response, len(response) > 0
}

func (a *Adapter) Set(key uint64, response []byte, expiration time.Time) error {
	return a.db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry(a.key(key), response).WithTTL(time.Until(expiration))
		return txn.SetEntry(e)
	})
}

func (a *Adapter) Release(key uint64) error {
	txn := a.db.NewTransaction(true)
	defer txn.Discard()

	err := txn.Delete(a.key(key))
	if err != nil {
		return err
	}

	return txn.Commit()
}

func (a *Adapter) key(key uint64) []byte {
	return []byte(fmt.Sprintf("%d", key))
}
