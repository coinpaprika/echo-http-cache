package redis

import (
	"os"
	"reflect"
	"testing"
	"time"

	cache "github.com/coinpaprika/echo-http-cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type RedisTestSuite struct {
	suite.Suite
	adapter cache.Adapter
}

func TestRedisTestSuite(t *testing.T) {
	suite.Run(t, new(RedisTestSuite))
}

func (suite *RedisTestSuite) SetupTest() {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "server"
	}

	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}

	suite.T().Logf("Using REDIS host: %s port: %s", host, port)

	suite.adapter = NewAdapter(&RingOptions{
		Addrs: map[string]string{
			host: ":" + port,
		},
	})
}

func (suite *RedisTestSuite) Test() {
	testsSet := []struct {
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
	for _, tt := range testsSet {
		suite.T().Run(tt.name, func(t *testing.T) {
			suite.T().Logf("setting key: %d", tt.key)

			err := suite.adapter.Set(tt.key, tt.response, time.Now().Add(1*time.Minute))
			assert.NoError(t, err)
		})
	}

	testsGet := []struct {
		name string
		key  uint64
		want []byte
		ok   bool
	}{
		{
			"returns right response 1",
			1,
			[]byte("value 1"),
			true,
		},
		{
			"returns right response 2",
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
	for _, tt := range testsGet {
		suite.T().Run(tt.name, func(t *testing.T) {
			suite.T().Logf("getting key: %d", tt.key)

			b, ok := suite.adapter.Get(tt.key)
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

	testsRelease := []struct {
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
	for _, tt := range testsRelease {
		suite.T().Run(tt.name, func(t *testing.T) {
			suite.T().Logf("releasing key: %d", tt.key)

			err := suite.adapter.Release(tt.key)
			assert.NoError(t, err)
			if _, ok := suite.adapter.Get(tt.key); ok {
				t.Errorf("memory.Release() error; key %v should not be found", tt.key)
			}
		})
	}
}
