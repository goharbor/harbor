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
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/errors"
)

var _ cache.Cache = (*Cache)(nil)

// Cache redis cache
type Cache struct {
	*redis.Client
	opts *cache.Options
}

// Contains returns true if key exists
func (c *Cache) Contains(ctx context.Context, key string) bool {
	val, err := c.Client.Exists(ctx, c.opts.Key(key)).Result()
	if err != nil {
		return false
	}

	return val == 1
}

// Delete delete item from cache by key
func (c *Cache) Delete(ctx context.Context, key string) error {
	return c.Client.Del(ctx, c.opts.Key(key)).Err()
}

// Fetch retrieve the cached key value
func (c *Cache) Fetch(ctx context.Context, key string, value interface{}) error {
	data, err := c.Client.Get(ctx, c.opts.Key(key)).Bytes()
	if err != nil {
		// convert internal or Timeout error to be ErrNotFound
		// so that the caller can continue working without breaking
		// return cache.ErrNotFound
		return fmt.Errorf("%w:%v", cache.ErrNotFound, err)
	}

	if err := c.opts.Codec.Decode(data, value); err != nil {
		return errors.Errorf("failed to decode cached value to dest, key %s, error: %v", key, err)
	}

	return nil
}

// Ping ping the cache
func (c *Cache) Ping(ctx context.Context) error {
	return c.Client.Ping(ctx).Err()
}

// Save cache the value by key
func (c *Cache) Save(ctx context.Context, key string, value interface{}, expiration ...time.Duration) error {
	data, err := c.opts.Codec.Encode(value)
	if err != nil {
		return errors.Errorf("failed to encode value, key %s, error: %v", key, err)
	}

	var exp time.Duration
	if len(expiration) > 0 {
		exp = expiration[0]
	} else if c.opts.Expiration > 0 {
		exp = c.opts.Expiration
	}

	return c.Client.Set(ctx, c.opts.Key(key), data, exp).Err()
}

// Keys returns the key matched by prefixes.
func (c *Cache) Keys(ctx context.Context, prefixes ...string) ([]string, error) {
	patterns := make([]string, 0, len(prefixes))
	if len(prefixes) == 0 {
		patterns = append(patterns, "*")
	} else {
		for _, p := range prefixes {
			patterns = append(patterns, c.opts.Key(p)+"*")
		}
	}

	keys := make([]string, 0)
	for _, pattern := range patterns {
		cmd := c.Client.Keys(ctx, pattern)
		if err := cmd.Err(); err != nil {
			return nil, err
		}

		keys = append(keys, cmd.Val()...)
	}

	return keys, nil
}

// New returns redis cache
func New(opts cache.Options) (cache.Cache, error) {
	if opts.Address == "" {
		opts.Address = "redis://localhost:6379/0"
	}

	if opts.Codec == nil {
		opts.Codec = cache.DefaultCodec()
	}

	u, err := url.Parse(opts.Address)
	if err != nil {
		return nil, err
	}

	// For compatibility, should convert idle_timeout_seconds to idle_timeout.
	values, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return nil, err
	}

	if t := values.Get("idle_timeout_seconds"); t != "" {
		values.Del("idle_timeout_seconds")
		values.Set("idle_timeout", t)
		u.RawQuery = values.Encode()
	}

	var client *redis.Client

	// default options in go-redis, also support change by provide parameters
	// from redis url.
	// DEFAULT VALUES
	/*
		OPTION  	       |  		QUERY		  |	  DEFAULT
		----------------------------------------------------------------------
		DialTimeout        | dial_timeout         | 5 * time.Second
		PoolSize           | pool_size            | 10 * runtime.GOMAXPROCS(0)
		ReadTimeout        | read_timeout         | 3 * time.Second
		WriteTimeout       | write_timeout        | ReadTimeout
		PoolTimeout        | pool_timeout         | ReadTimeout + time.Second
		IdleTimeout        | idle_timeout         | 5 * time.Minute
		IdleCheckFrequency | idle_check_frequency | time.Minute
		MaxRetries         | max_retries          | 3
		MinRetryBackoff    | min_retry_backoff    | 8 * time.Millisecond
		MaxRetryBackoff    | max_retry_backoff    | 512 * time.Millisecond
	*/

	switch u.Scheme {
	case cache.Redis:
		rdbOpts, err := redis.ParseURL(u.String())
		if err != nil {
			return nil, err
		}

		client = redis.NewClient(rdbOpts)
	case cache.RedisSentinel:
		failoverOpts, err := ParseSentinelURL(u.String())
		if err != nil {
			return nil, err
		}

		client = redis.NewFailoverClient(failoverOpts)
	default:
		return nil, errors.Errorf("redis: invalid URL scheme: %s", u.Scheme)
	}

	return &Cache{opts: &opts, Client: client}, nil
}

func init() {
	cache.Register(cache.Redis, New)
	cache.Register(cache.RedisSentinel, New)
}
