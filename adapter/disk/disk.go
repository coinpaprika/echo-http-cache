package disk

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/labstack/gommon/bytes"
	"github.com/labstack/gommon/log"
	"github.com/patrickmn/go-cache"
	"github.com/peterbourgon/diskv"
)

type (
	Adapter struct {
		directory     string
		debug         bool
		db            *diskv.Diskv
		maxMemorySize uint64

		expirationCache *cache.Cache
	}

	AdapterOptions func(a *Adapter) error
)

func NewAdapter(opts ...AdapterOptions) (*Adapter, error) {
	a := &Adapter{
		directory:       "./cache",
		maxMemorySize:   100 * bytes.MiB,
		expirationCache: cache.New(10*time.Minute, 30*time.Second),
	}

	for _, opt := range opts {
		if err := opt(a); err != nil {
			return nil, err
		}
	}

	a.db = diskv.New(diskv.Options{
		BasePath: a.directory,
		Transform: func(s string) []string {
			// "abcdef" -> ./cache/a/abcdef
			return []string{
				s[0:1],
				s[1:],
			}
		},
		CacheSizeMax: a.maxMemorySize,
	})

	a.expirationCache.OnEvicted(a.evict)
	a.runCleaner()

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

func WithMaxMemorySize(size uint64) AdapterOptions {
	return func(a *Adapter) error {
		a.maxMemorySize = size
		return nil
	}
}

func (a *Adapter) Get(key uint64) ([]byte, bool) {
	response, err := a.db.Read(a.key(key))
	if err != nil {
		if a.debug {
			log.Infof("[disk][get] key: %s, from cache: %t", a.key(key), len(response) > 0)
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

	// diskv doesn't have TTL, so we need to emulate it
	// on expirationCache eviction we will remove diskv entry
	// see Adapter.evict method for more details
	a.expirationCache.Set(a.key(key), struct{}{}, time.Until(expiration))
	return a.db.Write(a.key(key), response)
}

func (a *Adapter) Release(key uint64) error {
	if a.debug {
		log.Infof("[disk][delete] key: %s", a.key(key))
	}

	err := a.db.Erase(a.key(key))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (a *Adapter) key(key uint64) string {
	return fmt.Sprintf("%d", key)
}

func (a *Adapter) evict(k string, _ any) {
	if a.debug {
		log.Infof("[disk][expired] key: %s", k)
	}

	key, err := strconv.ParseUint(k, 0, 64)
	if err != nil {
		log.Error(err)
		return
	}

	if err := a.Release(key); err != nil {
		log.Error(err)
	}
}

func (a *Adapter) runCleaner() {
	// just in case remove all cache not to pollute the disk
	go func() {
		ticker := time.NewTicker(time.Duration(20+rand.Int31n(4)) * time.Hour) // #nosec
		defer ticker.Stop()

		for range ticker.C {
			if a.debug {
				log.Infof("[disk][gc] removing all cache")
			}
			if err := a.db.EraseAll(); err != nil {
				log.Error(err)
			}
		}
	}()
}
