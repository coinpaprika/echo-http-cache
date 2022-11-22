package redis

import (
	"os"
	"reflect"
	"testing"
	"time"

	cache "github.com/coinpaprika/echo-http-cache"
	"github.com/stretchr/testify/suite"
)

var a cache.Adapter

type RedisTestSuite struct {
	suite.Suite
}

func TestRedisTestSuite(t *testing.T) {
	suite.Run(t, new(RedisTestSuite))
}

func (suite *RedisTestSuite) Test1Set() {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "server"
	}

	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}

	suite.T().Logf("Using REDIS host: %s port: %s", host, port)

	a = NewAdapter(&RingOptions{
		Addrs: map[string]string{
			host: ":" + port,
		},
	})

	tests := []struct {
		name     string
		key      uint64
		response []byte
	}{
		{
			"sets a response cache",
			1,
			cache.Response{
				Value:      []byte("value 1"),
				Expiration: time.Now().Add(1 * time.Minute),
			}.Bytes(),
		},
		{
			"sets a response cache",
			2,
			cache.Response{
				Value:      []byte("value 2"),
				Expiration: time.Now().Add(1 * time.Minute),
			}.Bytes(),
		},
		{
			"sets a response cache",
			3,
			cache.Response{
				Value:      []byte("value 3"),
				Expiration: time.Now().Add(1 * time.Minute),
			}.Bytes(),
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			a.Set(tt.key, tt.response, time.Now().Add(1*time.Minute))
		})
	}
}

func (suite *RedisTestSuite) Test2Get() {
	tests := []struct {
		name string
		key  uint64
		want []byte
		ok   bool
	}{
		{
			"returns right response",
			1,
			[]byte("value 1"),
			true,
		},
		{
			"returns right response",
			2,
			[]byte("value 2"),
			true,
		},
		{
			"key does not exist",
			4,
			nil,
			false,
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			b, ok := a.Get(tt.key)
			if ok != tt.ok {
				t.Errorf("memory.Get() ok = %v, tt.ok %v", ok, tt.ok)
				return
			}
			got := cache.BytesToResponse(b).Value
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("memory.Get() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}

func (suite *RedisTestSuite) Test3Release() {
	tests := []struct {
		name string
		key  uint64
	}{
		{
			"removes cached response from store",
			1,
		},
		{
			"removes cached response from store",
			2,
		},
		{
			"removes cached response from store",
			3,
		},
		{
			"key does not exist",
			4,
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			a.Release(tt.key)
			if _, ok := a.Get(tt.key); ok {
				t.Errorf("memory.Release() error; key %v should not be found", tt.key)
			}
		})
	}
}
