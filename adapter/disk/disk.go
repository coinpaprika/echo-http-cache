package disk

import (
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/labstack/gommon/log"
)

type (
	Adapter struct {
		capacity  int
		directory string
		debug     bool
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

	db, err := badger.Open(badger.DefaultOptions(a.directory).
		WithIndexCacheSize(100_000_000))
	if err != nil {
		return nil, err
	}
	a.db = db
	go a.gc()

	return a, nil
}

func WithDirectory(directory string) AdapterOptions {
	return func(a *Adapter) error {
		a.directory = directory
		return nil
	}
}

func WithDebug(debug bool) AdapterOptions {
	return func(a *Adapter) error {
		a.debug = debug
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
		if a.debug {
			log.Infof("[disk][get] key: %s, from cache: false", a.key(key))
		}

		return nil, false
	}

	if a.debug {
		log.Infof("[disk][get] key: %s, from cache: %t", a.key(key), len(response) > 0)
	}

	return response, len(response) > 0
}

func (a *Adapter) Set(key uint64, response []byte, expiration time.Time) error {
	if a.debug {
		log.Infof("[disk][set] key: %s, duration: %s", a.key(key), time.Until(expiration))
	}

	return a.db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry(a.key(key), response).WithTTL(time.Until(expiration))
		return txn.SetEntry(e)
	})
}

func (a *Adapter) Release(key uint64) error {
	if a.debug {
		log.Infof("[disk][delete] key: %s", a.key(key))
	}

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

func (a *Adapter) gc() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		if a.debug {
			log.Infof("[disk][gc] running gc")
		}
		_ = a.db.RunValueLogGC(0.5)
	}
}
