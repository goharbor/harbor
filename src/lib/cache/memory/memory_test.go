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
	"testing"
	"time"

	"github.com/goharbor/harbor/src/lib/cache"
	"github.com/stretchr/testify/suite"
)

type CacheTestSuite struct {
	suite.Suite
	cache cache.Cache
	ctx   context.Context
}

func (suite *CacheTestSuite) SetupSuite() {
	suite.cache, _ = cache.New("memory", cache.Expiration(time.Second*5))
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

	suite.cache.Save(suite.ctx, key, map[string]interface{}{"name": "harbor", "version": "1.10"})

	mp := map[string]interface{}{}
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

func (suite *CacheTestSuite) TestKeys() {
	key1 := "p1"
	key2 := "p2"

	var err error
	err = suite.cache.Save(suite.ctx, key1, "hello, p1")
	suite.Nil(err)
	err = suite.cache.Save(suite.ctx, key2, "hello, p2")
	suite.Nil(err)

	// should match all
	keys, err := suite.cache.Keys(suite.ctx, "p")
	suite.Nil(err)
	suite.ElementsMatch([]string{"p1", "p2"}, keys)
	// only get p1
	keys, err = suite.cache.Keys(suite.ctx, key1)
	suite.Nil(err)
	suite.Equal([]string{"p1"}, keys)
	// only get p2
	keys, err = suite.cache.Keys(suite.ctx, key2)
	suite.Nil(err)
	suite.Equal([]string{"p2"}, keys)
}

func TestCacheTestSuite(t *testing.T) {
	suite.Run(t, new(CacheTestSuite))
}

func BenchmarkCacheFetchParallel(b *testing.B) {
	key := "benchmark"
	cache, _ := cache.New("memory")
	cache.Save(context.TODO(), key, "hello, benchmark")

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var value string
			cache.Fetch(context.TODO(), key, &value)
		}
	})
}

func BenchmarkCacheSaveParallel(b *testing.B) {
	key := "benchmark"
	cache, _ := cache.New("memory")

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cache.Save(context.TODO(), key, "hello, benchmark")
		}
	})
}
