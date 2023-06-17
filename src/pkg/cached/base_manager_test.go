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

package cached

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/lib/orm"
	testcache "github.com/goharbor/harbor/src/testing/lib/cache"
	"github.com/goharbor/harbor/src/testing/mock"
)

var testResourceType = "resource-test"

type testCache struct {
	*testcache.Cache
	iterator *testcache.Iterator
}

func (tc *testCache) Save(ctx context.Context, key string, value interface{}, expiration ...time.Duration) error {
	if orm.HasCommittedKey(ctx) {
		return nil
	}

	return tc.Cache.Save(ctx, key, value, expiration...)
}

type baseManagerTestSuite struct {
	suite.Suite
	cache *testCache
	mgr   *BaseManager
}

func (m *baseManagerTestSuite) SetupTest() {
	m.cache = &testCache{Cache: &testcache.Cache{}, iterator: &testcache.Iterator{}}
	m.mgr = NewBaseManager(testResourceType).WithCacheClient(m.cache)
}

func (m *baseManagerTestSuite) TestSave() {
	// normal ctx, should call cache.Save
	m.cache.On("Save", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	ctx := context.TODO()
	err := m.mgr.CacheClient(ctx).Save(ctx, "key", "value")
	m.NoError(err)
	m.cache.AssertCalled(m.T(), "Save", mock.Anything, mock.Anything, mock.Anything)

	// ctx in transaction, should skip call cache.Save
	m.cache.On("Save", mock.Anything, mock.Anything, mock.Anything).Panic("should not be called")
	ctx = context.WithValue(ctx, orm.CommittedKey{}, true)
	err = m.mgr.CacheClient(ctx).Save(ctx, "key", "value")
	m.NoError(err)
	m.cache.AssertNumberOfCalls(m.T(), "Save", 1)
}

func (m *baseManagerTestSuite) TestResourceType() {
	m.Equal(testResourceType, m.mgr.ResourceType(context.TODO()))
}

func (m *baseManagerTestSuite) TestCountCache() {
	m.cache.iterator.On("Next", mock.Anything).Return(false).Once()
	m.cache.On("Scan", mock.Anything, mock.Anything).Return(m.cache.iterator, nil).Once()
	c, err := m.mgr.CountCache(context.TODO())
	m.NoError(err)
	m.Equal(int64(0), c)
}

func (m *baseManagerTestSuite) TestDeleteCache() {
	m.cache.On("Delete", mock.Anything, "k1").Return(nil).Once()
	err := m.mgr.DeleteCache(context.TODO(), "k1")
	m.NoError(err)
}

func (m *baseManagerTestSuite) TestFlushAll() {
	m.cache.iterator.On("Next", mock.Anything).Return(false).Once()
	m.cache.On("Scan", mock.Anything, mock.Anything).Return(m.cache.iterator, nil).Once()
	err := m.mgr.FlushAll(context.TODO())
	m.NoError(err)
}

func TestBaseManager(t *testing.T) {
	suite.Run(t, &baseManagerTestSuite{})
}
