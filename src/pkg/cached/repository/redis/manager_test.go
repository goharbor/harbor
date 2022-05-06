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
	"github.com/goharbor/harbor/src/pkg/repository/model"
	testcache "github.com/goharbor/harbor/src/testing/lib/cache"
	"github.com/goharbor/harbor/src/testing/mock"
	testRepo "github.com/goharbor/harbor/src/testing/pkg/repository"
	"github.com/stretchr/testify/suite"
)

type managerTestSuite struct {
	suite.Suite
	cachedManager CachedManager
	repoMgr       *testRepo.Manager
	cache         *testcache.Cache
	ctx           context.Context
}

func (m *managerTestSuite) SetupTest() {
	m.repoMgr = &testRepo.Manager{}
	m.cache = &testcache.Cache{}
	m.cachedManager = NewManager(
		m.repoMgr,
	)
	m.cachedManager.(*manager).client = func() cache.Cache { return m.cache }
	m.ctx = context.TODO()
}

func (m *managerTestSuite) TestGet() {
	// get from cache directly
	m.cache.On("Fetch", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	_, err := m.cachedManager.Get(m.ctx, 100)
	m.NoError(err, "should get from cache")
	m.repoMgr.AssertNotCalled(m.T(), "Get", mock.Anything, mock.Anything)

	// not found in cache, read from dao
	m.cache.On("Fetch", mock.Anything, mock.Anything, mock.Anything).Return(cache.ErrNotFound).Once()
	m.cache.On("Save", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	m.repoMgr.On("Get", mock.Anything, mock.Anything).Return(&model.RepoRecord{}, nil).Once()
	_, err = m.cachedManager.Get(m.ctx, 100)
	m.NoError(err, "should get from repoMgr")
	m.repoMgr.AssertCalled(m.T(), "Get", mock.Anything, mock.Anything)
}

func (m *managerTestSuite) TestGetByName() {
	// get from cache directly
	m.cache.On("Fetch", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	_, err := m.cachedManager.GetByName(m.ctx, "repo")
	m.NoError(err, "should get from cache")
	m.repoMgr.AssertNotCalled(m.T(), "Get", mock.Anything, mock.Anything)

	// not found in cache, read from dao
	m.cache.On("Fetch", mock.Anything, mock.Anything, mock.Anything).Return(cache.ErrNotFound).Once()
	m.cache.On("Save", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	m.repoMgr.On("GetByName", mock.Anything, mock.Anything).Return(&model.RepoRecord{}, nil).Once()
	_, err = m.cachedManager.GetByName(m.ctx, "repo")
	m.NoError(err, "should get from repoMgr")
	m.repoMgr.AssertCalled(m.T(), "GetByName", mock.Anything, mock.Anything)
}

func (m *managerTestSuite) TestDelete() {
	// delete from repoMgr error
	errDelete := errors.New("delete failed")
	m.repoMgr.On("Delete", mock.Anything, mock.Anything).Return(errDelete).Once()
	m.cache.On("Fetch", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	err := m.cachedManager.Delete(m.ctx, 100)
	m.ErrorIs(err, errDelete, "delete should error")
	m.cache.AssertNotCalled(m.T(), "Delete", mock.Anything, mock.Anything)

	// delete from repoMgr success
	m.repoMgr.On("Delete", mock.Anything, mock.Anything).Return(nil).Once()
	m.cache.On("Fetch", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	m.cache.On("Delete", mock.Anything, mock.Anything).Return(nil).Twice()
	err = m.cachedManager.Delete(m.ctx, 100)
	m.NoError(err, "delete should success")
	m.cache.AssertCalled(m.T(), "Delete", mock.Anything, mock.Anything)
}

func (m *managerTestSuite) TestUpdate() {
	// update from repoMgr error
	errUpdate := errors.New("update failed")
	m.repoMgr.On("Update", mock.Anything, mock.Anything).Return(errUpdate).Once()
	err := m.cachedManager.Update(m.ctx, &model.RepoRecord{})
	m.ErrorIs(err, errUpdate, "update should error")
	m.cache.AssertNotCalled(m.T(), "Delete", mock.Anything, mock.Anything)

	// update from repoMgr success
	m.repoMgr.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
	m.cache.On("Delete", mock.Anything, mock.Anything).Return(nil).Twice()
	err = m.cachedManager.Update(m.ctx, &model.RepoRecord{})
	m.NoError(err, "update should success")
	m.cache.AssertCalled(m.T(), "Delete", mock.Anything, mock.Anything)
}

func (m *managerTestSuite) TestAddPullCount() {
	// update pull count from repoMgr error
	errUpdate := errors.New("update pull count failed")
	m.repoMgr.On("AddPullCount", mock.Anything, mock.Anything, mock.Anything).Return(errUpdate).Once()
	m.cache.On("Fetch", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	err := m.cachedManager.AddPullCount(m.ctx, 100, 1)
	m.ErrorIs(err, errUpdate, "update pull count should error")
	m.cache.AssertNotCalled(m.T(), "Delete", mock.Anything, mock.Anything)

	// update pull count from repoMgr success
	m.repoMgr.On("AddPullCount", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	m.cache.On("Delete", mock.Anything, mock.Anything).Return(nil).Twice()
	err = m.cachedManager.AddPullCount(m.ctx, 100, 1)
	m.NoError(err, "update pull count should success")
	m.cache.AssertCalled(m.T(), "Delete", mock.Anything, mock.Anything)
}

func (m *managerTestSuite) TestCount() {
	m.repoMgr.On("Count", mock.Anything, mock.Anything).Return(int64(1), nil)
	c, err := m.cachedManager.Count(m.ctx, nil)
	m.NoError(err)
	m.Equal(int64(1), c)
	m.repoMgr.AssertCalled(m.T(), "Count", mock.Anything, mock.Anything)
}

func (m *managerTestSuite) TestList() {
	repos := []*model.RepoRecord{}
	m.repoMgr.On("List", mock.Anything, mock.Anything).Return(repos, nil)
	as, err := m.cachedManager.List(m.ctx, nil)
	m.NoError(err)
	m.Equal(repos, as)
	m.repoMgr.AssertCalled(m.T(), "List", mock.Anything, mock.Anything)
}

func (m *managerTestSuite) TestCreate() {
	m.repoMgr.On("Create", mock.Anything, mock.Anything).Return(int64(1), nil)
	id, err := m.cachedManager.Create(m.ctx, nil)
	m.NoError(err)
	m.Equal(int64(1), id)
	m.repoMgr.AssertCalled(m.T(), "Create", mock.Anything, mock.Anything)
}

func (m *managerTestSuite) TestNonEmptyRepos() {
	repos := []*model.RepoRecord{}
	m.repoMgr.On("NonEmptyRepos", mock.Anything, mock.Anything).Return(repos, nil)
	rs, err := m.cachedManager.NonEmptyRepos(m.ctx)
	m.NoError(err)
	m.Equal(repos, rs)
	m.repoMgr.AssertCalled(m.T(), "NonEmptyRepos", mock.Anything, mock.Anything)
}
func (m *managerTestSuite) TestResourceType() {
	t := m.cachedManager.ResourceType(m.ctx)
	m.Equal("repository", t)
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
