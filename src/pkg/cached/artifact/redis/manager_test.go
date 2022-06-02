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
	"time"

	"github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/pkg/artifact"
	testcache "github.com/goharbor/harbor/src/testing/lib/cache"
	"github.com/goharbor/harbor/src/testing/mock"
	testArt "github.com/goharbor/harbor/src/testing/pkg/artifact"
	"github.com/stretchr/testify/suite"
)

type managerTestSuite struct {
	suite.Suite
	cachedManager CachedManager
	artMgr        *testArt.Manager
	cache         *testcache.Cache
	ctx           context.Context
}

func (m *managerTestSuite) SetupTest() {
	m.artMgr = &testArt.Manager{}
	m.cache = &testcache.Cache{}
	m.cachedManager = NewManager(
		m.artMgr,
	)
	m.cachedManager.(*manager).client = func() cache.Cache { return m.cache }
	m.ctx = context.TODO()
}

func (m *managerTestSuite) TestGet() {
	// get from cache directly
	m.cache.On("Fetch", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	_, err := m.cachedManager.Get(m.ctx, 100)
	m.NoError(err, "should get from cache")
	m.artMgr.AssertNotCalled(m.T(), "Get", mock.Anything, mock.Anything)

	// not found in cache, read from dao
	m.cache.On("Fetch", mock.Anything, mock.Anything, mock.Anything).Return(cache.ErrNotFound).Once()
	m.cache.On("Save", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	m.artMgr.On("Get", mock.Anything, mock.Anything).Return(&artifact.Artifact{}, nil).Once()
	_, err = m.cachedManager.Get(m.ctx, 100)
	m.NoError(err, "should get from artMgr")
	m.artMgr.AssertCalled(m.T(), "Get", mock.Anything, mock.Anything)
}

func (m *managerTestSuite) TestGetByDigest() {
	// get from cache directly
	m.cache.On("Fetch", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	_, err := m.cachedManager.GetByDigest(m.ctx, "repo", "sha256:418fb88ec412e340cdbef913b8ca1bbe8f9e8dc705f9617414c1f2c8db980180")
	m.NoError(err, "should get from cache")
	m.artMgr.AssertNotCalled(m.T(), "Get", mock.Anything, mock.Anything)

	// not found in cache, read from dao
	m.cache.On("Fetch", mock.Anything, mock.Anything, mock.Anything).Return(cache.ErrNotFound).Once()
	m.cache.On("Save", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	m.artMgr.On("Get", mock.Anything, mock.Anything).Return(&artifact.Artifact{}, nil).Once()
	_, err = m.cachedManager.Get(m.ctx, 100)
	m.NoError(err, "should get from artMgr")
	m.artMgr.AssertCalled(m.T(), "Get", mock.Anything, mock.Anything)
}

func (m *managerTestSuite) TestDelete() {
	// delete from artMgr error
	errDelete := errors.New("delete failed")
	m.artMgr.On("Delete", mock.Anything, mock.Anything).Return(errDelete).Once()
	m.artMgr.On("DeleteReferences", mock.Anything, mock.Anything).Return(nil).Once()
	m.cache.On("Fetch", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	err := m.cachedManager.Delete(m.ctx, 100)
	m.ErrorIs(err, errDelete, "delete should error")
	m.cache.AssertNotCalled(m.T(), "Delete", mock.Anything, mock.Anything)

	// delete from artMgr success
	m.artMgr.On("Delete", mock.Anything, mock.Anything).Return(nil).Once()
	m.artMgr.On("DeleteReferences", mock.Anything, mock.Anything).Return(nil).Once()
	m.cache.On("Fetch", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	m.cache.On("Delete", mock.Anything, mock.Anything).Return(nil).Twice()
	err = m.cachedManager.Delete(m.ctx, 100)
	m.NoError(err, "delete should success")
	m.cache.AssertCalled(m.T(), "Delete", mock.Anything, mock.Anything)
}

func (m *managerTestSuite) TestUpdate() {
	// update from artMgr error
	errUpdate := errors.New("update failed")
	m.artMgr.On("Update", mock.Anything, mock.Anything).Return(errUpdate).Once()
	err := m.cachedManager.Update(m.ctx, &artifact.Artifact{})
	m.ErrorIs(err, errUpdate, "update should error")
	m.cache.AssertNotCalled(m.T(), "Delete", mock.Anything, mock.Anything)

	// update from artMgr success
	m.artMgr.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
	m.cache.On("Delete", mock.Anything, mock.Anything).Return(nil).Twice()
	err = m.cachedManager.Update(m.ctx, &artifact.Artifact{})
	m.NoError(err, "update should success")
	m.cache.AssertCalled(m.T(), "Delete", mock.Anything, mock.Anything)
}

func (m *managerTestSuite) TestUpdatePullTime() {
	// update pull time from artMgr error
	errUpdate := errors.New("update pull time failed")
	m.artMgr.On("UpdatePullTime", mock.Anything, mock.Anything, mock.Anything).Return(errUpdate).Once()
	m.cache.On("Fetch", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	err := m.cachedManager.UpdatePullTime(m.ctx, 100, time.Now())
	m.ErrorIs(err, errUpdate, "update pull time should error")
	m.cache.AssertNotCalled(m.T(), "Delete", mock.Anything, mock.Anything)

	// update pull time from artMgr success
	m.artMgr.On("UpdatePullTime", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	m.cache.On("Delete", mock.Anything, mock.Anything).Return(nil).Twice()
	err = m.cachedManager.UpdatePullTime(m.ctx, 100, time.Now())
	m.NoError(err, "update pull time should success")
	m.cache.AssertCalled(m.T(), "Delete", mock.Anything, mock.Anything)
}

func (m *managerTestSuite) TestCount() {
	m.artMgr.On("Count", mock.Anything, mock.Anything).Return(int64(1), nil)
	c, err := m.cachedManager.Count(m.ctx, nil)
	m.NoError(err)
	m.Equal(int64(1), c)
	m.artMgr.AssertCalled(m.T(), "Count", mock.Anything, mock.Anything)
}

func (m *managerTestSuite) TestList() {
	arts := []*artifact.Artifact{}
	m.artMgr.On("List", mock.Anything, mock.Anything).Return(arts, nil)
	as, err := m.cachedManager.List(m.ctx, nil)
	m.NoError(err)
	m.Equal(arts, as)
	m.artMgr.AssertCalled(m.T(), "List", mock.Anything, mock.Anything)
}

func (m *managerTestSuite) TestCreate() {
	m.artMgr.On("Create", mock.Anything, mock.Anything).Return(int64(1), nil)
	id, err := m.cachedManager.Create(m.ctx, nil)
	m.NoError(err)
	m.Equal(int64(1), id)
	m.artMgr.AssertCalled(m.T(), "Create", mock.Anything, mock.Anything)
}

func (m *managerTestSuite) TestListReferences() {
	refs := []*artifact.Reference{}
	m.artMgr.On("ListReferences", mock.Anything, mock.Anything).Return(refs, nil)
	rs, err := m.cachedManager.ListReferences(m.ctx, nil)
	m.NoError(err)
	m.Equal(refs, rs)
	m.artMgr.AssertCalled(m.T(), "ListReferences", mock.Anything, mock.Anything)
}

func (m *managerTestSuite) TestDeleteReference() {
	m.artMgr.On("DeleteReference", mock.Anything, mock.Anything).Return(nil)
	err := m.cachedManager.DeleteReference(m.ctx, 1)
	m.NoError(err)
	m.artMgr.AssertCalled(m.T(), "DeleteReference", mock.Anything, mock.Anything)
}

func (m *managerTestSuite) TestResourceType() {
	t := m.cachedManager.ResourceType(m.ctx)
	m.Equal("artifact", t)
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
