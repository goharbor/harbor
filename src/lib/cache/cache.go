// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cache

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/retry"
)

const (
	// Memory the cache name of memory
	Memory = "memory"
	// Redis the cache name of redis
	Redis = "redis"
	// RedisSentinel the cache name of redis sentinel
	RedisSentinel = "redis+sentinel"
)

var (
	// ErrNotFound error returns when the key value not found in the cache
	ErrNotFound = errors.New("key not found")
)

// Iterator returns the ScanIterator
type Iterator interface {
	Next(ctx context.Context) bool
	Val() string
}

// Cache cache interface
type Cache interface {
	// Contains returns true if key exists
	Contains(ctx context.Context, key string) bool

	// Delete delete item from cache by key
	Delete(ctx context.Context, key string) error

	// Fetch retrieve the cached key value
	Fetch(ctx context.Context, key string, value interface{}) error

	// Ping ping the cache
	Ping(ctx context.Context) error

	// Save cache the value by key
	Save(ctx context.Context, key string, value interface{}, expiration ...time.Duration) error

	// Scan scans the keys matched by match string
	// NOTICE: memory cache does not support use wildcard, compared by strings.Contains
	Scan(ctx context.Context, match string) (Iterator, error)
}

var (
	factories   = map[string]func(opts Options) (Cache, error){}
	factoriesMu sync.RWMutex
)

// Register register cache factory for type
func Register(typ string, factory func(opts Options) (Cache, error)) {
	factoriesMu.Lock()
	defer factoriesMu.Unlock()

	factories[typ] = factory
}

// New returns cache from addr
func New(typ string, opt ...Option) (Cache, error) {
	opts := newOptions(opt...)
	opts.Codec = codec // use the default codec for the cache

	factoriesMu.Lock()
	defer factoriesMu.Unlock()

	factory, ok := factories[typ]
	if !ok {
		return nil, fmt.Errorf("cache type %s not support", typ)
	}

	return factory(opts)
}

var (
	cache Cache
)

// Initialize initialize the default cache from the addr
func Initialize(typ, addr string) error {
	c, err := New(typ, Address(addr), Prefix("cache:"))
	if err != nil {
		return err
	}

	redactedAddr := addr
	if u, err := url.Parse(addr); err == nil {
		redactedAddr = u.Redacted()
	}

	options := []retry.Option{
		retry.InitialInterval(time.Millisecond * 500),
		retry.MaxInterval(time.Second * 10),
		retry.Timeout(time.Minute),
		retry.Callback(func(err error, sleep time.Duration) {
			log.Errorf("failed to ping %s, retry after %s : %v", redactedAddr, sleep, err)
		}),
	}

	if err := retry.Retry(func() error { return c.Ping(context.TODO()) }, options...); err != nil {
		return err
	}

	cache = c

	return nil
}

// Default returns the default cache
func Default() Cache {
	return cache
}

var (
	// cacheLayer is the global cache layer cache instance.
	cacheLayer Cache
	// cacheLayerOnce is the once condition for initializing instance.
	cacheLayerOnce sync.Once
)

// LayerCache is the global cache instance for cache layer.
func LayerCache() Cache {
	// parse the redis url for cache layer, use the default cache if not specify
	redisCacheURL := os.Getenv("_REDIS_URL_CACHE_LAYER")
	if redisCacheURL == "" {
		if cache != nil {
			return cache
		}
		// use the core url if cache layer url not found
		redisCacheURL = os.Getenv("_REDIS_URL_CORE")
	}

	u, err := url.Parse(redisCacheURL)
	if err != nil {
		log.Fatal("failed to parse the redis url for cache layer, bad _REDIS_URL_CACHE_LAYER")
	}

	cacheLayerOnce.Do(func() {
		cacheLayer, err = New(u.Scheme, Address(redisCacheURL), Prefix("cache:"))
		if err != nil {
			log.Fatalf("failed to initialize cache for cache layer, err: %v", err)
		}
	})

	return cacheLayer
}
