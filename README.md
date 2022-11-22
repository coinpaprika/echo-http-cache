# echo-http-cache
[![Check & test & build](https://github.com/coinpaprika/echo-http-cache/actions/workflows/main.yml/badge.svg)](https://github.com/coinpaprika/echo-http-cache/actions/workflows/main.yml)

This is a high performance Golang HTTP middleware for server-side application layer caching, ideal for REST APIs, using Echo framework.
It is simple, superfast, thread safe and gives the possibility to choose the adapter (memory, Redis).

## Getting Started

### Installation (Go Modules)
`go get github.com/coinpaprika/echo-http-cache`

### Usage
This is an example of use with the memory adapter:

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
	memoryAdapter, err := memory.NewAdapter()
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

Example of Client initialization with Redis adapter:
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

## License
echo-http-cache is released under the [MIT License](https://github.com/SporkHubr/echo-http-cache/blob/master/LICENSE).

## Forked from:
- [victorspringer/http-cache](https://github.com/victorspringer/http-cache)
- [SporkHubr/echo-http-cache](https://github.com/SporkHubr/echo-http-cache)