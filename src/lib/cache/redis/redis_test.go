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
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/lib/cache"
)

type CacheTestSuite struct {
	suite.Suite
	cache cache.Cache
	ctx   context.Context
}

func (suite *CacheTestSuite) SetupSuite() {
	suite.cache, _ = cache.New("redis", cache.Expiration(time.Second*5))
	suite.ctx = context.TODO()
}

func (suite *CacheTestSuite) TestContains() {
	key := "contains"
	suite.False(suite.cache.Contains(suite.ctx, key))

	suite.cache.Save(suite.ctx, key, "value")
	suite.True(suite.cache.Contains(suite.ctx, key))

	suite.cache.Delete(suite.ctx, key)
	suite.False(suite.cache.Contains(suite.ctx, key))

	suite.cache.Save(suite.ctx, key, "value", time.Second*5)
	suite.True(suite.cache.Contains(suite.ctx, key))

	time.Sleep(time.Second * 8)
	suite.False(suite.cache.Contains(suite.ctx, key))
}

func (suite *CacheTestSuite) TestDelete() {
	key := "delete"

	suite.cache.Save(suite.ctx, key, "value")
	suite.True(suite.cache.Contains(suite.ctx, key))

	suite.cache.Delete(suite.ctx, key)
	suite.False(suite.cache.Contains(suite.ctx, key))
}

func (suite *CacheTestSuite) TestFetch() {
	key := "fetch"

	suite.cache.Save(suite.ctx, key, map[string]any{"name": "harbor", "version": "1.10"})

	mp := map[string]any{}
	suite.cache.Fetch(suite.ctx, key, &mp)
	suite.Len(mp, 2)
	suite.Equal("harbor", mp["name"])
	suite.Equal("1.10", mp["version"])

	var str string
	suite.Error(suite.cache.Fetch(suite.ctx, key, &str))
}

func (suite *CacheTestSuite) TestSave() {
	key := "save"

	{
		suite.cache.Save(suite.ctx, key, "hello, save")

		var value string
		suite.cache.Fetch(suite.ctx, key, &value)
		suite.Equal("hello, save", value)

		time.Sleep(time.Second * 8)

		value = ""
		suite.Error(suite.cache.Fetch(suite.ctx, key, &value))
		suite.Equal("", value)
	}

	{
		suite.cache.Save(suite.ctx, key, "hello, save", time.Second)

		time.Sleep(time.Second * 2)

		var value string
		suite.Error(suite.cache.Fetch(suite.ctx, key, &value))
		suite.Equal("", value)
	}
}

func (suite *CacheTestSuite) TestPing() {
	suite.NoError(suite.cache.Ping(suite.ctx))
}

func (suite *CacheTestSuite) TestScan() {
	seed := func(n int) {
		for i := 0; i < n; i++ {
			key := fmt.Sprintf("test-scan-%d", i)
			err := suite.cache.Save(suite.ctx, key, "")
			suite.NoError(err)
		}
	}
	clean := func(n int) {
		for i := 0; i < n; i++ {
			key := fmt.Sprintf("test-scan-%d", i)
			err := suite.cache.Delete(suite.ctx, key)
			suite.NoError(err)
		}
	}
	{
		// return all keys with test-scan-*
		expect := []string{"test-scan-0", "test-scan-1", "test-scan-2"}
		// seed data
		seed(3)
		// test scan
		iter, err := suite.cache.Scan(suite.ctx, "test-scan-*")
		suite.NoError(err)
		got := []string{}
		for iter.Next(suite.ctx) {
			got = append(got, iter.Val())
		}
		suite.ElementsMatch(expect, got)
		// clean up
		clean(3)
	}

	{
		// return matched keys with test-scan-1*
		expect := []string{"test-scan-1", "test-scan-10"}
		// seed data
		seed(11)
		// test scan
		iter, err := suite.cache.Scan(suite.ctx, "test-scan-1*")
		suite.NoError(err)
		got := []string{}
		for iter.Next(suite.ctx) {
			got = append(got, iter.Val())
		}
		suite.ElementsMatch(expect, got)
		// clean up
		clean(11)
	}
}

func TestCacheTestSuite(t *testing.T) {
	suite.Run(t, new(CacheTestSuite))
}

func BenchmarkCacheFetchParallel(b *testing.B) {
	key := "benchmark"
	cache, _ := cache.New("redis")
	cache.Save(context.TODO(), key, "hello, benchmark")

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var value string
			err := cache.Fetch(context.TODO(), key, &value)
			if err != nil {
				fmt.Printf("failed, error %v\n", err)
			}
		}
	})
}

func BenchmarkCacheSaveParallel(b *testing.B) {
	key := "benchmark"
	cache, _ := cache.New("redis")

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cache.Save(context.TODO(), key, "hello, benchmark")
		}
	})
}
