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

package memory

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/goharbor/harbor/src/lib/cache"
)

type entry struct {
	data        []byte
	expiratedAt int64
}

func (e *entry) isExpirated() bool {
	return e.expiratedAt < time.Now().UnixNano()
}

var _ cache.Cache = (*Cache)(nil)

// Cache memory cache
type Cache struct {
	opts    *cache.Options
	storage sync.Map
}

// Contains returns true if key exists
func (c *Cache) Contains(ctx context.Context, key string) bool {
	e, ok := c.storage.Load(c.opts.Key(key))
	if !ok {
		return false
	}

	if e.(*entry).isExpirated() {
		c.Delete(ctx, c.opts.Key(key))
		return false
	}

	return true
}

// Delete delete item from cache by key
func (c *Cache) Delete(ctx context.Context, key string) error {
	c.storage.Delete(c.opts.Key(key))
	return nil
}

// Fetch retrieve the cached key value
func (c *Cache) Fetch(ctx context.Context, key string, value interface{}) error {
	v, ok := c.storage.Load(c.opts.Key(key))
	if !ok {
		return cache.ErrNotFound
	}

	e := v.(*entry)
	if e.isExpirated() {
		c.Delete(ctx, c.opts.Key(key))
		return cache.ErrNotFound
	}

	if err := c.opts.Codec.Decode(e.data, value); err != nil {
		return fmt.Errorf("failed to decode cached value to dest, key %s, error: %v", key, err)
	}

	return nil
}

// Ping ping the cache
func (c *Cache) Ping(ctx context.Context) error {
	return nil
}

// Save cache the value by key
func (c *Cache) Save(ctx context.Context, key string, value interface{}, expiration ...time.Duration) error {
	data, err := c.opts.Codec.Encode(value)
	if err != nil {
		return fmt.Errorf("failed to encode value, key %s, error: %v", key, err)
	}

	var expiratedAt int64
	if len(expiration) > 0 {
		expiratedAt = time.Now().Add(expiration[0]).UnixNano()
	} else if c.opts.Expiration > 0 {
		expiratedAt = time.Now().Add(c.opts.Expiration).UnixNano()
	} else {
		expiratedAt = math.MaxInt64
	}

	c.storage.Store(c.opts.Key(key), &entry{
		data:        data,
		expiratedAt: expiratedAt,
	})

	return nil
}

// Keys returns the key matched by prefixes.
func (c *Cache) Keys(ctx context.Context, prefixes ...string) ([]string, error) {
	// if no prefix, means match all keys.
	matchAll := len(prefixes) == 0
	// range map to get all keys
	keys := make([]string, 0)
	c.storage.Range(func(k, v interface{}) bool {
		ks := k.(string)
		if matchAll {
			keys = append(keys, ks)
		} else {
			for _, p := range prefixes {
				if strings.HasPrefix(ks, c.opts.Key(p)) {
					keys = append(keys, ks)
				}
			}
		}
		return true
	})

	return keys, nil
}

// New returns memory cache
func New(opts cache.Options) (cache.Cache, error) {
	return &Cache{opts: &opts}, nil
}

func init() {
	cache.Register(cache.Memory, New)
}
