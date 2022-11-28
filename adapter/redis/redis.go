/*
MIT License

Copyright (c) 2018 Victor Springer

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package redis

import (
	"context"
	"time"

	cache "github.com/coinpaprika/echo-http-cache"
	redisCache "github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"github.com/labstack/gommon/log"
)

type (
	Adapter struct {
		store *redisCache.Cache
		debug bool
	}
	AdapterOptions func(a *Adapter)
	RingOptions    redis.RingOptions
)

// Get implements the cache Adapter interface Get method.
func (a *Adapter) Get(key uint64) ([]byte, bool) {
	var c []byte
	if err := a.store.Get(context.Background(), cache.KeyAsString(key), &c); err == nil {
		if a.debug {
			log.Infof("[redis][get] key: %s, from cache: true", cache.KeyAsString(key))
		}
		return c, true
	}

	if a.debug {
		log.Infof("[redis][get] key: %s, from cache: false", cache.KeyAsString(key))
	}
	return nil, false
}

// Set implements the cache Adapter interface Set method.
func (a *Adapter) Set(key uint64, response []byte, expiration time.Time) error {
	if a.debug {
		log.Infof("[redis][set] key: %s, duration: %s", cache.KeyAsString(key), time.Until(expiration))
	}

	ttl := time.Until(expiration)
	if ttl.Seconds() <= 1 {
		// REDIS TTL has to be > 1s
		// otherwise warning is generated: `2022/11/28 11:37:00 too short TTL for key="2bsunt0a1fan6": 808.586939ms`
		ttl = 1 * time.Second
	}

	return a.store.Set(&redisCache.Item{
		Key:   cache.KeyAsString(key),
		Value: response,
		TTL:   ttl,
	})
}

// Release implements the cache Adapter interface Release method.
func (a *Adapter) Release(key uint64) error {
	if a.debug {
		log.Infof("[redis][delete] key: %s", cache.KeyAsString(key))
	}

	return a.store.Delete(context.Background(), cache.KeyAsString(key))
}

// NewAdapter initializes Redis adapter.
func NewAdapter(opt *RingOptions, opts ...AdapterOptions) cache.Adapter {
	ropt := redis.RingOptions(*opt)
	adapter := &Adapter{
		store: redisCache.New(&redisCache.Options{
			Redis: redis.NewRing(&ropt),
		}),
		debug: false,
	}

	for _, opt := range opts {
		opt(adapter)
	}

	return adapter
}

func WithDebug(debug bool) AdapterOptions {
	return func(a *Adapter) {
		a.debug = debug
	}
}
