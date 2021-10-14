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

package redis

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/goharbor/harbor/src/lib/cache"
	libredis "github.com/goharbor/harbor/src/lib/redis"
	"github.com/gomodule/redigo/redis"
)

var _ cache.Cache = (*Cache)(nil)

// Cache redis cache
type Cache struct {
	opts *cache.Options
	pool *redis.Pool
}

// Contains returns true if key exists
func (c *Cache) Contains(key string) bool {
	reply, err := redis.Int(c.do("EXISTS", c.opts.Key(key)))
	if err != nil {
		return false
	}

	return reply == 1
}

// Delete delete item from cache by key
func (c *Cache) Delete(key string) error {
	_, err := c.do("DEL", c.opts.Key(key))
	return err
}

// Fetch retrieve the cached key value
func (c *Cache) Fetch(key string, value interface{}) error {
	data, err := redis.Bytes(c.do("GET", c.opts.Key(key)))
	if err != nil {
		// convert internal or Timeout error to be ErrNotFound
		// so that the caller can continue working without breaking
		return cache.ErrNotFound
	}

	if err := c.opts.Codec.Decode(data, value); err != nil {
		return fmt.Errorf("failed to decode cached value to dest, key %s, error: %v", key, err)
	}

	return nil
}

// Ping ping the cache
func (c *Cache) Ping() error {
	_, err := c.do("PING")
	return err
}

// Save cache the value by key
func (c *Cache) Save(key string, value interface{}, expiration ...time.Duration) error {
	data, err := c.opts.Codec.Encode(value)
	if err != nil {
		return fmt.Errorf("failed to encode value, key %s, error: %v", key, err)
	}

	args := []interface{}{c.opts.Key(key), data}

	var exp time.Duration
	if len(expiration) > 0 {
		exp = expiration[0]
	} else if c.opts.Expiration > 0 {
		exp = c.opts.Expiration
	}

	if exp > 0 {
		args = append(args, "EX", int64(exp/time.Second))
	}

	_, err = c.do("SET", args...)
	return err
}

func (c *Cache) do(commandName string, args ...interface{}) (reply interface{}, err error) {
	conn := c.pool.Get()
	defer conn.Close()

	return conn.Do(commandName, args...)
}

// New returns redis cache
func New(opts cache.Options) (cache.Cache, error) {
	if opts.Address == "" {
		opts.Address = "redis://localhost:6379/0"
	}

	name := fmt.Sprintf("%x", sha256.Sum256([]byte(opts.Address)))

	param := &libredis.PoolParam{
		PoolMaxIdle:           100,
		PoolMaxActive:         1000,
		PoolIdleTimeout:       10 * time.Minute,
		DialConnectionTimeout: time.Second,
		DialReadTimeout:       time.Second * 2,
		DialWriteTimeout:      time.Second * 5,
	}

	pool, err := libredis.GetRedisPool(name, opts.Address, param)
	if err != nil {
		return nil, err
	}

	return &Cache{opts: &opts, pool: pool}, nil
}

func init() {
	cache.Register(cache.Redis, New)
	cache.Register(cache.RedisSentinel, New)
}
