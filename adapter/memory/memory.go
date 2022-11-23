package memory

import (
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
)

type (
	Adapter struct {
		cache    *cache.Cache
		capacity int
	}
	AdapterOptions func(a *Adapter) error
)

func NewAdapter(opts ...AdapterOptions) (*Adapter, error) {
	a := &Adapter{
		cache: cache.New(10*time.Minute, 30*time.Second),
	}
	for _, opt := range opts {
		if err := opt(a); err != nil {
			return nil, err
		}
	}
	return a, nil
}

func WithCapacity(capacity int) AdapterOptions {
	return func(a *Adapter) error {
		a.capacity = capacity
		return nil
	}
}

func (a *Adapter) Get(key uint64) ([]byte, bool) {
	if v, ok := a.cache.Get(a.key(key)); ok {
		return v.([]byte), true
	}
	return nil, false
}

func (a *Adapter) Set(key uint64, response []byte, expiration time.Time) error {
	if a.capacity > 0 && a.cache.ItemCount() >= a.capacity {
		// it's better to not cache an item than DDoS the server
		// we will wait for the cleanup goroutine to kick in within a few secs. and make a room for new entries
		return nil
	}

	a.cache.Set(a.key(key), response, time.Until(expiration))
	return nil
}

func (a *Adapter) Release(key uint64) error {
	a.cache.Delete(a.key(key))
	return nil
}

func (a *Adapter) key(key uint64) string {
	return fmt.Sprintf("%d", key)
}
