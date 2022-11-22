package memory

import (
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
)

type (
	Adapter struct {
		cache *cache.Cache
	}
	AdapterOptions func(a *Adapter) error
)

func NewAdapter(opts ...AdapterOptions) (*Adapter, error) {
	a := &Adapter{
		cache: cache.New(10*time.Minute, 1*time.Minute),
	}
	for _, opt := range opts {
		if err := opt(a); err != nil {
			return nil, err
		}
	}
	return a, nil
}

func WithDefaultExpiration(expiration time.Duration) AdapterOptions {
	return func(a *Adapter) error {
		a.cache = cache.New(expiration, 1*time.Minute)
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
