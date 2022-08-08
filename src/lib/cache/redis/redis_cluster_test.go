package redis

import (
	"context"
	"fmt"

	"strings"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/errors"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/suite"
)

type RedisClusterTestSuite struct {
	suite.Suite
	cache cache.Cache
	ctx   context.Context
	c     *miniredis.Miniredis
}

func (suite *RedisClusterTestSuite) SetupSuite() {
	suite.c = miniredis.RunT(suite.T())
	addr := suite.c.Addr()
	suite.T().Logf("redis cluster addr is %s", addr)
	var err error
	suite.cache, err = cache.New(cache.RedisCluster, cache.Expiration(time.Second*5),
		cache.Address(fmt.Sprintf("redis+cluster://%s", addr)))
	suite.NoError(err, "RedisCluster New error")
	suite.ctx = context.TODO()
	err = suite.cache.Ping(suite.ctx)
	suite.NoError(err)
}

func (suite *RedisClusterTestSuite) TestNewRedisCluster() {
	tests := []struct {
		name string
		opt  cache.Options
		err  error
	}{
		{
			name: "address_is_success",
			opt: cache.Options{
				Address: "redis+cluster://localhost:12345,localhost:123",
			},
			err: nil,
		},
		{
			name: "address_separator_is_not_legal",
			opt: cache.Options{
				Address: "redis+cluster://localhost:12345;localhost:123",
			},
			err: errors.Errorf("redis cluster address validate error,cluster adders is %s", "localhost:12345;localhost:123"),
		},
		{
			name: "address_not_default_param",
			opt: cache.Options{
				Address: "redis+cluster://anonymous:password@host1:26379,host2:26379?idle_timeout_seconds=30&max_retries=10&min_retry_backoff=1&max_retry_backoff=10&dial_timeout=30&read_timeout=5&write_timeout=5&pool_fifo=true&pool_size=1000&min_idle_conns=100&max_conn_age=10&pool_timeout=10",
			},
			err: nil,
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			_, err := NewRedisCluster(tt.opt)
			if tt.err != nil {
				suite.Error(err)
			} else {
				suite.NoError(err)
			}
		})
	}

}

func (suite *RedisClusterTestSuite) TestContains() {
	type tests struct {
		name   string
		key    string
		before func(tt *tests) error
		after  func(tt *tests) error
		want   bool
	}
	tts := []tests{
		{
			name:   "key_is_not_contains",
			key:    "TestContains1",
			before: nil,
			after:  nil,
			want:   false,
		},
		{
			name: "key_is_containers",
			key:  "TestContains1",
			before: func(tt *tests) error {
				return suite.cache.Save(suite.ctx, tt.key, 1)
			},
			after: func(tt *tests) error {
				return suite.cache.Delete(suite.ctx, tt.key)
			},
			want: true,
		},
	}
	for _, tt := range tts {
		suite.T().Run(tt.name, func(t *testing.T) {
			if tt.before != nil {
				if err := tt.before(&tt); err != nil {
					t.Fatal(err)
				}
			}
			contains := suite.cache.Contains(suite.ctx, tt.key)
			suite.Equal(tt.want, contains)
			if tt.after != nil {
				if err := tt.after(&tt); err != nil {
					t.Fatal(err)
				}
			}
		})
	}

}

func (suite *RedisClusterTestSuite) TestFetch() {
	type tests struct {
		name   string
		key    string
		values string
		before func(tt *tests) error
		after  func(tt *tests) error
		want   error
	}
	tts := []tests{
		{
			name:   "fetch_not_content",
			key:    "TestFetch1",
			values: "1",
			before: nil,
			after:  nil,
			want:   fmt.Errorf("%w", cache.ErrNotFound),
		},
		{
			name:   "fetch_success",
			key:    "TestFetch2",
			values: "2",
			before: func(tt *tests) error {
				return suite.cache.Save(suite.ctx, tt.key, tt.values)
			},
			after: func(tt *tests) error {
				return suite.cache.Delete(suite.ctx, tt.key)
			},
			want: nil,
		},
	}
	for _, tt := range tts {
		suite.T().Run(tt.name, func(t *testing.T) {
			if tt.before != nil {
				if err := tt.before(&tt); err != nil {
					t.Fatal(err)
				}

			}
			err := suite.cache.Fetch(suite.ctx, tt.key, tt.values)
			if tt.want != nil {
				if !strings.Contains(err.Error(), tt.want.Error()) {
					t.Fatal(err)
				}
			}
			if tt.after != nil {
				if err := tt.after(&tt); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func (suite *RedisClusterTestSuite) TestKeys() {
	type tests struct {
		name     string
		prefixes []string
		before   func(tt *tests) error
		after    func(tt *tests) error
		want     []string
	}
	tts := []tests{
		{
			name:     "keys_is_nil",
			prefixes: []string{"TestKeys"},
			before:   nil,
			after:    nil,
			want:     nil,
		},
		{
			name:     "keys_is_succes",
			prefixes: []string{"TestKeys", "a"},
			before: func(tt *tests) error {
				for i := 0; i < 3; i++ {
					suite.cache.Save(suite.ctx, fmt.Sprintf("TestKeys%d", i), i)
				}
				suite.cache.Save(suite.ctx, "a1", 1)
				suite.cache.Save(suite.ctx, "ab", 1)
				suite.cache.Save(suite.ctx, "c", 1)
				return nil
			},
			after: func(tt *tests) error {
				keys, err := suite.cache.Keys(suite.ctx)
				if err != nil {
					return err
				}
				for _, key := range keys {
					err = suite.cache.Delete(suite.ctx, key)
					if err != nil {
						return err
					}
				}
				return nil
			},
			want: []string{"TestKeys0", "TestKeys1", "TestKeys2", "a1", "ab"},
		},
	}
	for _, tt := range tts {
		suite.T().Run(tt.name, func(t *testing.T) {
			if tt.before != nil {
				if err := tt.before(&tt); err != nil {
					t.Fatal(err)
				}
			}
			keys, err := suite.cache.Keys(suite.ctx, tt.prefixes...)
			if err != nil {
				t.Fatal(err)
			}
			if tt.want != nil {
				t.Logf("%+v", keys)
				t.Logf("%+v", tt.want)
				for i, v := range tt.want {
					if keys[i] != v {
						t.Fatal()
					}
				}
			}
			if tt.after != nil {
				if err = tt.after(&tt); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestRedisClusterTestSuite(t *testing.T) {
	suite.Run(t, new(RedisClusterTestSuite))
}
