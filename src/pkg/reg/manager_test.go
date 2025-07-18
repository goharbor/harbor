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

package reg

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/pkg/reg/dao"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	testcache "github.com/goharbor/harbor/src/testing/lib/cache"
	"github.com/goharbor/harbor/src/testing/mock"
	testingdao "github.com/goharbor/harbor/src/testing/pkg/reg/dao"
)

type managerTestSuite struct {
	suite.Suite
	mgr   *manager
	dao   *testingdao.DAO
	cache *testcache.Cache
	ctx   context.Context
}

func (m *managerTestSuite) SetupTest() {
	m.dao = &testingdao.DAO{}
	m.cache = &testcache.Cache{}
	m.ctx = context.TODO()
	m.mgr = &manager{
		dao:          m.dao,
		cacheEnabled: true,
		cacheExpire:  time.Hour,
	}
	m.mgr.setCacheClient(m.cache)
}

func (m *managerTestSuite) TestCount() {
	mock.OnAnything(m.dao, "Count").Return(int64(1), nil)
	n, err := m.mgr.Count(nil, nil)
	m.Require().Nil(err)
	m.Equal(int64(1), n)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestList() {
	mock.OnAnything(m.dao, "List").Return([]*dao.Registry{
		{
			ID: 1,
		},
	}, nil)
	registries, err := m.mgr.List(nil, nil)
	m.Require().Nil(err)
	m.Require().Equal(1, len(registries))
	m.Equal(int64(1), registries[0].ID)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestGet() {
	// Test local registry (id=0) - should not use cache
	registry, err := m.mgr.Get(m.ctx, 0)
	m.Require().Nil(err)
	m.Equal("Local", registry.Name)
	m.cache.AssertNotCalled(m.T(), "Fetch")
	m.dao.AssertNotCalled(m.T(), "Get")
}

func (m *managerTestSuite) TestCreate() {
	mock.OnAnything(m.dao, "Create").Return(int64(1), nil)
	_, err := m.mgr.Create(nil, &model.Registry{})
	m.Require().Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestDelete() {
	mock.OnAnything(m.dao, "Delete").Return(nil)
	m.cache.On("Delete", mock.Anything, "registry:id:1").Return(nil)
	err := m.mgr.Delete(m.ctx, 1)
	m.Require().Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestUpdate() {
	mock.OnAnything(m.dao, "Update").Return(nil)
	m.cache.On("Delete", mock.Anything, "registry:id:0").Return(nil)
	err := m.mgr.Update(m.ctx, &model.Registry{})
	m.Require().Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestGetWithCache() {
	// Test cache hit - should return from cache directly
	m.cache.On("Fetch", mock.Anything, "registry:id:1", mock.Anything).Return(nil).Once().Run(func(args mock.Arguments) {
		registry := args.Get(2).(*model.Registry)
		registry.ID = 1
		registry.Name = "cached-registry"
	})

	registry, err := m.mgr.Get(m.ctx, 1)
	m.Require().Nil(err)
	m.Equal(int64(1), registry.ID)
	m.Equal("cached-registry", registry.Name)
	m.dao.AssertNotCalled(m.T(), "Get")
}

func (m *managerTestSuite) TestGetWithCacheMiss() {
	// Test cache miss - should query database and save to cache
	m.cache.On("Fetch", mock.Anything, "registry:id:1", mock.Anything).Return(cache.ErrNotFound).Once()
	m.cache.On("Save", mock.Anything, "registry:id:1", mock.Anything, mock.Anything).Return(nil).Once()

	mock.OnAnything(m.dao, "Get").Return(&dao.Registry{
		ID:   1,
		Name: "test-registry",
	}, nil).Once()

	registry, err := m.mgr.Get(m.ctx, 1)
	m.Require().Nil(err)
	m.Equal(int64(1), registry.ID)
	m.Equal("test-registry", registry.Name)
	m.dao.AssertCalled(m.T(), "Get", mock.Anything, int64(1))
	m.cache.AssertCalled(m.T(), "Save", mock.Anything, "registry:id:1", mock.Anything, mock.Anything)
}

func (m *managerTestSuite) TestGetCacheDisabled() {
	// Test with cache disabled
	m.mgr.cacheEnabled = false

	mock.OnAnything(m.dao, "Get").Return(&dao.Registry{
		ID:   1,
		Name: "test-registry",
	}, nil).Once()

	registry, err := m.mgr.Get(m.ctx, 1)
	m.Require().Nil(err)
	m.Equal(int64(1), registry.ID)
	m.Equal("test-registry", registry.Name)
	m.dao.AssertCalled(m.T(), "Get", mock.Anything, int64(1))
	m.cache.AssertNotCalled(m.T(), "Fetch")
	m.cache.AssertNotCalled(m.T(), "Save")
}

func (m *managerTestSuite) TestUpdateWithCacheInvalidation() {
	// Test update with cache invalidation
	mock.OnAnything(m.dao, "Update").Return(nil).Once()
	m.cache.On("Delete", mock.Anything, "registry:id:1").Return(nil).Once()

	registry := &model.Registry{
		ID:   1,
		Name: "updated-registry",
	}

	err := m.mgr.Update(m.ctx, registry)
	m.Require().Nil(err)
	m.dao.AssertCalled(m.T(), "Update", mock.Anything, mock.Anything)
	m.cache.AssertCalled(m.T(), "Delete", mock.Anything, "registry:id:1")
}

func (m *managerTestSuite) TestUpdateWithDaoError() {
	// Test update with dao error - should not invalidate cache
	updateErr := errors.New("update failed")
	mock.OnAnything(m.dao, "Update").Return(updateErr).Once()

	registry := &model.Registry{
		ID:   1,
		Name: "updated-registry",
	}

	err := m.mgr.Update(m.ctx, registry)
	m.Require().Equal(updateErr, err)
	m.dao.AssertCalled(m.T(), "Update", mock.Anything, mock.Anything)
	m.cache.AssertNotCalled(m.T(), "Delete")
}

func (m *managerTestSuite) TestDeleteWithCacheInvalidation() {
	// Test delete with cache invalidation
	mock.OnAnything(m.dao, "Delete").Return(nil).Once()
	m.cache.On("Delete", mock.Anything, "registry:id:1").Return(nil).Once()

	err := m.mgr.Delete(m.ctx, 1)
	m.Require().Nil(err)
	m.dao.AssertCalled(m.T(), "Delete", mock.Anything, int64(1))
	m.cache.AssertCalled(m.T(), "Delete", mock.Anything, "registry:id:1")
}

func (m *managerTestSuite) TestDeleteWithDaoError() {
	// Test delete with dao error - should not invalidate cache
	deleteErr := errors.New("delete failed")
	mock.OnAnything(m.dao, "Delete").Return(deleteErr).Once()

	err := m.mgr.Delete(m.ctx, 1)
	m.Require().Equal(deleteErr, err)
	m.dao.AssertCalled(m.T(), "Delete", mock.Anything, int64(1))
	m.cache.AssertNotCalled(m.T(), "Delete")
}

func (m *managerTestSuite) TestCacheKeyFormat() {
	// Test that cache keys are formatted correctly
	m.cache.On("Fetch", mock.Anything, "registry:id:123", mock.Anything).Return(cache.ErrNotFound).Once()
	m.cache.On("Save", mock.Anything, "registry:id:123", mock.Anything, mock.Anything).Return(nil).Once()

	mock.OnAnything(m.dao, "Get").Return(&dao.Registry{
		ID:   123,
		Name: "test-registry",
	}, nil).Once()

	_, err := m.mgr.Get(m.ctx, 123)
	m.Require().Nil(err)
	m.cache.AssertCalled(m.T(), "Fetch", mock.Anything, "registry:id:123", mock.Anything)
	m.cache.AssertCalled(m.T(), "Save", mock.Anything, "registry:id:123", mock.Anything, mock.Anything)
}

func TestManager(t *testing.T) {
	suite.Run(t, &managerTestSuite{})
}
