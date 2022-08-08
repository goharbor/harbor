package redis

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/errors"

	"github.com/go-redis/redis/v8"
)

const (
	// Separator is split redis cluster address
	Separator string = ","
	// RegexpPattern is checker redis cluster address is legal
	RegexpPattern string = "^[A-Za-z0-9+://@?&.,=_]+$"
)

func init() {
	cache.Register(cache.RedisCluster, NewRedisCluster)
}

// Check redisCluster whether implement cache.Cache interface
var _ cache.Cache = (*redisCluster)(nil)

type redisCluster struct {
	*redis.ClusterClient
	opts *cache.Options
}

func NewRedisCluster(opts cache.Options) (cache.Cache, error) {
	if opts.Address == "" {
		opts.Address = "redis+cluster://localhost:6379"
	}

	// check address
	if match, err := regexp.MatchString(RegexpPattern, opts.Address); err == nil && !match {
		return nil, errors.Errorf("redis cluster address validate error,cluster adders is %s", opts.Address)
	}

	if opts.Codec == nil {
		opts.Codec = cache.DefaultCodec()
	}

	var clusterClient *redis.ClusterClient

	u, err := url.Parse(opts.Address)
	if err != nil {
		return nil, err
	}

	clusterOpts, err := ParseClusterURL(u.String())
	if err != nil {
		return nil, err
	}
	clusterClient = redis.NewClusterClient(clusterOpts)

	return &redisCluster{
		ClusterClient: clusterClient,
		opts:          &opts,
	}, nil
}

// Contains returns true if key exists
func (c *redisCluster) Contains(ctx context.Context, key string) bool {
	val, err := c.ClusterClient.Exists(ctx, c.opts.Key(key)).Result()
	if err != nil {
		return false
	}

	return val == 1
}

// Delete delete item from cache by key
func (c *redisCluster) Delete(ctx context.Context, key string) error {
	return c.ClusterClient.Del(ctx, c.opts.Key(key)).Err()
}

// Fetch retrieve the cached key value
func (c *redisCluster) Fetch(ctx context.Context, key string, value interface{}) error {
	data, err := c.ClusterClient.Get(ctx, c.opts.Key(key)).Bytes()
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

// Ping the cache
func (c *redisCluster) Ping(ctx context.Context) error {
	return c.ClusterClient.Ping(ctx).Err()
}

// Save cache the value by key
func (c *redisCluster) Save(ctx context.Context, key string, value interface{}, expiration ...time.Duration) error {
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

	return c.ClusterClient.Set(ctx, c.opts.Key(key), data, exp).Err()
}

// Keys returns the key matched by prefixes.
func (c *redisCluster) Keys(ctx context.Context, prefixes ...string) ([]string, error) {
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
		cmd := c.ClusterClient.Keys(ctx, pattern)
		if err := cmd.Err(); err != nil {
			return nil, err
		}

		for _, k := range cmd.Val() {
			keys = append(keys, strings.TrimPrefix(k, c.opts.Prefix))
		}
	}

	return keys, nil
}
