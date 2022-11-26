package main

import (
	"log"
	"net/http"
	"time"

	cache "github.com/coinpaprika/echo-http-cache"
	"github.com/coinpaprika/echo-http-cache/adapter/disk"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/bytes"
)

func main() {
	var adapter cache.Adapter
	var err error
	adapter, err = disk.NewAdapter(
		disk.WithDebug(true),
		disk.WithMaxMemorySize(20*bytes.MiB),
	)

	if err != nil {
		log.Fatal(err)
	}

	cacheClient, err := cache.NewClient(
		cache.ClientWithAdapter(adapter),
		cache.ClientWithTTL(10*time.Second),
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

	if err := e.Start(":8080"); err != nil {
		log.Fatal(err)
	}
}
