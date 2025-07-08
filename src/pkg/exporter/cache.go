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

package exporter

import (
	"sync"
	"time"
)

var c *cache

const defaultCacheCleanInterval = 10

type cachedValue struct {
	Value      any
	Expiration int64
}

type cache struct {
	CacheDuration int64
	store         map[string]cachedValue
	*sync.RWMutex
}

// CacheGet get a value from cache
func CacheGet(key string) (value any, ok bool) {
	c.RLock()
	v, ok := c.store[key]
	c.RUnlock()
	if !ok {
		return nil, false
	}
	if time.Now().Unix() > v.Expiration {
		c.Lock()
		delete(c.store, key)
		c.Unlock()
		return nil, false
	}
	return v.Value, true
}

// CachePut put a value to cache with key
func CachePut(key, value any) {
	c.Lock()
	defer c.Unlock()
	c.store[key.(string)] = cachedValue{
		Value:      value,
		Expiration: time.Now().Unix() + c.CacheDuration,
	}
}

// CacheDelete delete a key from cache
func CacheDelete(key string) {
	c.Lock()
	defer c.Unlock()
	delete(c.store, key)
}

// StartCacheCleaner start a cache clean job
func StartCacheCleaner() {
	now := time.Now().UnixNano()
	c.Lock()
	defer c.Unlock()
	for k, v := range c.store {
		if v.Expiration < now {
			delete(c.store, k)
		}
	}
}

// CacheEnabled returns if the cache in exporter enabled
func CacheEnabled() bool {
	return c != nil
}

// CacheInit add cache to exporter
func CacheInit(opt *Opt) {
	c = &cache{
		CacheDuration: opt.CacheDuration,
		store:         make(map[string]cachedValue),
		RWMutex:       &sync.RWMutex{},
	}
	go func() {
		var cacheCleanInterval int64
		if opt.CacheCleanInterval > 0 {
			cacheCleanInterval = opt.CacheCleanInterval
		} else {
			cacheCleanInterval = defaultCacheCleanInterval
		}
		ticker := time.NewTicker(time.Duration(cacheCleanInterval) * time.Second)
		for range ticker.C {
			StartCacheCleaner()
		}
	}()
}
