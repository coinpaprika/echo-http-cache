# echo-http-cache
[![Check & test & build](https://github.com/coinpaprika/echo-http-cache/actions/workflows/main.yml/badge.svg)](https://github.com/coinpaprika/echo-http-cache/actions/workflows/main.yml)

This is a high performance Golang HTTP middleware for server-side application layer caching, ideal for REST APIs, using Echo framework.
It is simple, superfast, thread safe and gives the possibility to choose the adapter (memory, Redis).

## Getting Started

### Installation (Go Modules)
`go get github.com/coinpaprika/echo-http-cache`

### Usage
Full example is available at [example](./example/main.go) can be run by:

`go run ./example/main.go`

Example of use with the memory adapter:
```go
package main

import (
	"log"
	"net/http"
	"time"

	cache "github.com/coinpaprika/echo-http-cache"
	"github.com/coinpaprika/echo-http-cache/adapter/memory"
	"github.com/labstack/echo/v4"
)

func main() {
	memoryAdapter, err := memory.NewAdapter(
		// after reaching the capacity, items are not cached 
		// until the next cleaning goroutine makes the space
		// this is a protection against cache pollution attacks
		memory.WithCapacity(10_000),  
	) 
	if err != nil {
		log.Fatal(err)
	}

	cacheClient, err := cache.NewClient(
		cache.ClientWithAdapter(memoryAdapter),
		cache.ClientWithTTL(10*time.Minute),
		cache.ClientWithRefreshKey("opn"),
	)
	if err != nil {
		log.Fatal(err)
	}

	e := echo.New()
	e.Use(cacheClient.Middleware())
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
	e.Start(":8080")
}
```

Example of Client initialization with REDIS adapter:
```go
import (
    "github.com/coinpaprika/echo-http-cache"
    "github.com/coinpaprika/echo-http-cache/adapter/redis"
)

...

    ringOpt := &redis.RingOptions{
        Addrs: map[string]string{
            "server": ":6379",
        },
    }
    cacheClient := cache.NewClient(
        cache.ClientWithAdapter(redis.NewAdapter(ringOpt)),
        cache.ClientWithTTL(10 * time.Minute),
        cache.ClientWithRefreshKey("opn"),
    )
...
```

Example of Client initialization with disk based adapter using [diskv](https://github.com/peterbourgon/diskv):
```go
import (
    "github.com/coinpaprika/echo-http-cache"
    "github.com/coinpaprika/echo-http-cache/adapter/disk"
)

...
    cacheClient := cache.NewClient(
        // leave empty for default directory './cache'. Directory will be created if not exist.
        cache.ClientWithAdapter(disk.NewAdapter(disk.WithDirectory("./tmp/cache"), disk.WithMaxMemorySize(50_000_000))), 
        cache.ClientWithTTL(10 * time.Minute),
        cache.ClientWithRefreshKey("opn"),
    )
```

## Adapters selection guide
### `Memory`
- local environments
- production single & multi node environments
- short-lived objects < 3min
- cheap underlying operations' avg(exec time) < 300ms
- low number of entries: < 1M & < 1Gb in size
- memory safe (when used with `WithCapacity` option)

### `Disk`
- production single & multi node environments 
- short-lived to medium-lived objects < 12hr
- cheap underlying operations' avg(exec time) < 300ms
- always memory safe, disk space is used extensively
- some entries are cached in memory for performance - controlled by WithMaxMemorySize() settings, default 100Mb
- large number of entries > 1M & > 1 Gb in size (up to full size of a disk)

### `Redis`
- production multi node environments
- short-lived to long-lived objects > 10 min
- expensive underlying operations' avg(exec time) > 300ms, benefit from sharing across multi nodes
- large number of entries > 1M & >1 Gb in size (up to full size of a disk)

## License
echo-http-cache is released under the [MIT License](https://github.com/SporkHubr/echo-http-cache/blob/master/LICENSE).

## Forked from:
- [victorspringer/http-cache](https://github.com/victorspringer/http-cache)
- [SporkHubr/echo-http-cache](https://github.com/SporkHubr/echo-http-cache)
