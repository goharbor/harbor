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
	"github.com/goharbor/harbor/src/lib/log"
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
		err := c.Delete(ctx, c.opts.Key(key))
		log.Errorf("failed to delete cache in Contains() method when it's expired, error: %v", err)
		return false
	}

	return true
}

// Delete delete item from cache by key
func (c *Cache) Delete(_ context.Context, key string) error {
	c.storage.Delete(c.opts.Key(key))
	return nil
}

// Fetch retrieve the cached key value
func (c *Cache) Fetch(ctx context.Context, key string, value any) error {
	v, ok := c.storage.Load(c.opts.Key(key))
	if !ok {
		return cache.ErrNotFound
	}

	e := v.(*entry)
	if e.isExpirated() {
		err := c.Delete(ctx, c.opts.Key(key))
		if err != nil {
			log.Errorf("failed to delete cache in Fetch() method when it's expired, error: %v", err)
		}
		return cache.ErrNotFound
	}

	if err := c.opts.Codec.Decode(e.data, value); err != nil {
		return fmt.Errorf("failed to decode cached value to dest, key %s, error: %v", key, err)
	}

	return nil
}

// Ping ping the cache
func (c *Cache) Ping(_ context.Context) error {
	return nil
}

// Save cache the value by key
func (c *Cache) Save(_ context.Context, key string, value any, expiration ...time.Duration) error {
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

// Scan scans the keys matched by match string
func (c *Cache) Scan(_ context.Context, match string) (cache.Iterator, error) {
	var keys []string
	c.storage.Range(func(k, v any) bool {
		matched := true
		if match != "" {
			matched = strings.Contains(k.(string), match)
		}

		if matched {
			if v.(*entry).isExpirated() {
				c.storage.Delete(k)
			} else {
				keys = append(keys, strings.TrimPrefix(k.(string), c.opts.Prefix))
			}
		}
		return true
	})

	return &ScanIterator{keys: keys}, nil
}

// ScanIterator is a ScanIterator for memory cache
type ScanIterator struct {
	mu   sync.Mutex
	pos  int
	keys []string
}

// Next checks whether has the next element
func (i *ScanIterator) Next(_ context.Context) bool {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.pos++
	return i.pos <= len(i.keys)
}

// Val returns the key
func (i *ScanIterator) Val() string {
	i.mu.Lock()
	defer i.mu.Unlock()

	var val string
	if i.pos <= len(i.keys) {
		val = i.keys[i.pos-1]
	}

	return val
}

// New returns memory cache
func New(opts cache.Options) (cache.Cache, error) {
	return &Cache{opts: &opts}, nil
}

func init() {
	cache.Register(cache.Memory, New)
}
