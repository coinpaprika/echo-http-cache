# echo-http-cache
[![Build Status](https://travis-ci.org/victorspringer/http-cache.svg?branch=master)](https://travis-ci.org/victorspringer/http-cache) [![Coverage Status](https://coveralls.io/repos/github/victorspringer/http-cache/badge.svg?branch=master)](https://coveralls.io/github/victorspringer/http-cache?branch=master) [![](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat)](https://godoc.org/github.com/SporkHubr/echo-http-cache)

This is a high performance Golang HTTP middleware for server-side application layer caching, ideal for REST APIs, using Echo framework.

It is simple, super fast, thread safe and gives the possibility to choose the adapter (memory, Redis, DynamoDB etc).

The memory adapter minimizes GC overhead to near zero and supports some options of caching algorithms (LRU, MRU, LFU, MFU). This way, it is able to store plenty of gigabytes of responses, keeping great performance and being free of leaks.

## Getting Started

### Installation (Go Modules)
`go get github.com/SporkHubr/echo-http-cache`

### Usage
This is an example of use with the memory adapter:

```go
package main

import (
    "fmt"
    "net/http"
    "os"
    "time"
    
    "github.com/SporkHubr/echo-http-cache"
    "github.com/SporkHubr/echo-http-cache/adapter/memory"
    "github.com/labstack/echo/v4"
)

func example(c echo.Context) {
   c.String(http.StatusOk, "Ok")
}

func main() {
    memcached, err := memory.NewAdapter(
        memory.AdapterWithAlgorithm(memory.LRU),
        memory.AdapterWithCapacity(10000000),
    )
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    cacheClient, err := cache.NewClient(
        cache.ClientWithAdapter(memcached),
        cache.ClientWithTTL(10 * time.Minute),
        cache.ClientWithRefreshKey("opn"),
    )
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    router := echo.New()
    router.Use(cacheClient.Middleware())
    router.GET("/", example)
    e.Start(":8080")
}
```

Example of Client initialization with Redis adapter:
```go
import (
    "github.com/SporkHubr/echo-http-cache"
    "github.com/SporkHubr/echo-http-cache/adapter/redis"
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

## Godoc Reference
- [echo-http-cache](https://pkg.go.dev/github.com/SporkHubr/echo-http-cache)
- [Memory adapter](https://pkg.go.dev/github.com/SporkHubr/echo-http-cache/adapter/memory)
- [Redis adapter](https://pkg.go.dev/github.com/SporkHubr/echo-http-cache/adapter/redis)

## License
echo-http-cache is released under the [MIT License](https://github.com/SporkHubr/echo-http-cache/blob/master/LICENSE).

## Forked from:
- [victorspringer/http-cache](https://github.com/victorspringer/http-cache)
- [SporkHubr/echo-http-cache](https://github.com/SporkHubr/echo-http-cache)