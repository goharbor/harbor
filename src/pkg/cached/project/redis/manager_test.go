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
	"errors"
	"testing"

	"github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/project/models"
	testcache "github.com/goharbor/harbor/src/testing/lib/cache"
	"github.com/goharbor/harbor/src/testing/mock"
	testProject "github.com/goharbor/harbor/src/testing/pkg/project"

	"github.com/stretchr/testify/suite"
)

type managerTestSuite struct {
	suite.Suite
	cachedManager CachedManager
	projectMgr    *testProject.Manager
	cache         *testcache.Cache
	ctx           context.Context
}

func (m *managerTestSuite) SetupTest() {
	m.projectMgr = &testProject.Manager{}
	m.cache = &testcache.Cache{}
	m.cachedManager = NewManager(
		m.projectMgr,
	)
	m.cachedManager.(*manager).client = func() cache.Cache { return m.cache }
	m.ctx = context.TODO()
}

func (m *managerTestSuite) TestCreate() {
	m.projectMgr.On("Create", mock.Anything, mock.Anything).Return(int64(1), nil)
	id, err := m.cachedManager.Create(m.ctx, &models.Project{})
	m.NoError(err)
	m.Equal(int64(1), id)
}

func (m *managerTestSuite) TestCount() {
	m.projectMgr.On("Count", mock.Anything, mock.Anything).Return(int64(100), nil)
	c, err := m.cachedManager.Count(m.ctx, q.New(q.KeyWords{}))
	m.NoError(err)
	m.Equal(int64(100), c)
}

func (m *managerTestSuite) TestList() {
	m.projectMgr.On("List", mock.Anything, mock.Anything).Return([]*models.Project{}, nil)
	ps, err := m.cachedManager.List(m.ctx, q.New(q.KeyWords{}))
	m.NoError(err)
	m.ElementsMatch([]*models.Project{}, ps)
}

func (m *managerTestSuite) TestListRoles() {
	m.projectMgr.On("ListRoles", mock.Anything, mock.Anything, mock.Anything).Return([]int{1}, nil)
	rs, err := m.cachedManager.ListRoles(m.ctx, 1, 1)
	m.NoError(err)
	m.Equal([]int{1}, rs)
}

func (m *managerTestSuite) TestGet() {
	// get from cache directly
	m.cache.On("Fetch", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	_, err := m.cachedManager.Get(m.ctx, 100)
	m.NoError(err, "should get from cache")
	m.projectMgr.AssertNotCalled(m.T(), "Get", mock.Anything, mock.Anything)

	// not found in cache, read from dao
	m.cache.On("Fetch", mock.Anything, mock.Anything, mock.Anything).Return(cache.ErrNotFound).Once()
	m.cache.On("Save", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	m.projectMgr.On("Get", mock.Anything, mock.Anything).Return(&models.Project{}, nil).Once()
	_, err = m.cachedManager.Get(m.ctx, 100)
	m.NoError(err, "should get from projectMgr")
	m.projectMgr.AssertCalled(m.T(), "Get", mock.Anything, mock.Anything)
}

func (m *managerTestSuite) TestDelete() {
	// delete from projectMgr error
	errDelete := errors.New("delete failed")
	m.projectMgr.On("Delete", mock.Anything, mock.Anything).Return(errDelete).Once()
	m.cache.On("Fetch", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	err := m.cachedManager.Delete(m.ctx, 100)
	m.ErrorIs(err, errDelete, "delete should error")
	m.cache.AssertNotCalled(m.T(), "Delete", mock.Anything, mock.Anything)

	// delete from projectMgr success
	m.projectMgr.On("Delete", mock.Anything, mock.Anything).Return(nil).Once()
	m.cache.On("Fetch", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	m.cache.On("Delete", mock.Anything, mock.Anything).Return(nil).Twice()
	err = m.cachedManager.Delete(m.ctx, 100)
	m.NoError(err, "delete should success")
	m.cache.AssertCalled(m.T(), "Delete", mock.Anything, mock.Anything)
}

func (m *managerTestSuite) TestResourceType() {
	t := m.cachedManager.ResourceType(m.ctx)
	m.Equal("project", t)
}

func (m *managerTestSuite) TestCountCache() {
	m.cache.On("Keys", mock.Anything, mock.Anything).Return([]string{"1"}, nil).Once()
	c, err := m.cachedManager.CountCache(m.ctx)
	m.NoError(err)
	m.Equal(int64(1), c)
}

func (m *managerTestSuite) TestDeleteCache() {
	m.cache.On("Delete", mock.Anything, mock.Anything).Return(nil).Once()
	err := m.cachedManager.DeleteCache(m.ctx, "key")
	m.NoError(err)
}

func (m *managerTestSuite) TestFlushAll() {
	m.cache.On("Keys", mock.Anything, mock.Anything).Return([]string{"1"}, nil).Once()
	m.cache.On("Delete", mock.Anything, mock.Anything).Return(nil).Once()
	err := m.cachedManager.FlushAll(m.ctx)
	m.NoError(err)
}

func TestManager(t *testing.T) {
	suite.Run(t, &managerTestSuite{})
}
