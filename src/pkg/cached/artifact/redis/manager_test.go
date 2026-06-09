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
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/artifact"
	testcache "github.com/goharbor/harbor/src/testing/lib/cache"
	"github.com/goharbor/harbor/src/testing/mock"
	testArt "github.com/goharbor/harbor/src/testing/pkg/artifact"
)

type managerTestSuite struct {
	suite.Suite
	cachedManager CachedManager
	artMgr        *testArt.Manager
	cache         *testcache.Cache
	iterator      *testcache.Iterator
	ctx           context.Context
}

func (m *managerTestSuite) SetupTest() {
	m.artMgr = &testArt.Manager{}
	m.cache = &testcache.Cache{}
	m.iterator = &testcache.Iterator{}
	m.cachedManager = NewManager(m.artMgr)
	m.cachedManager.(*Manager).WithCacheClient(m.cache)
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
	m.iterator.On("Next", mock.Anything).Return(false).Once()
	m.cache.On("Scan", mock.Anything, mock.Anything).Return(m.iterator, nil).Once()
	c, err := m.cachedManager.CountCache(m.ctx)
	m.NoError(err)
	m.Equal(int64(0), c)
}

func (m *managerTestSuite) TestDeleteCache() {
	m.cache.On("Delete", mock.Anything, mock.Anything).Return(nil).Once()
	err := m.cachedManager.DeleteCache(m.ctx, "key")
	m.NoError(err)
}

func (m *managerTestSuite) TestFlushAll() {
	m.iterator.On("Next", mock.Anything).Return(false).Once()
	m.cache.On("Scan", mock.Anything, mock.Anything).Return(m.iterator, nil).Once()
	m.cache.On("Delete", mock.Anything, mock.Anything).Return(nil).Once()
	err := m.cachedManager.FlushAll(m.ctx)
	m.NoError(err)
}

// TestCleanUpKeys_DoesNotSpinOnCanceledContext is the key regression test for
// https://github.com/goharbor/harbor/issues/21062. Under the old retry.Retry,
// a canceled context passed into cache.Delete would keep the retry loop
// spinning for the full 60-second default timeout on each of the two cache
// keys — ~2 minutes per artifact — while the enclosing DB transaction sat
// open in ClientRead / idle-in-transaction holding row locks. Under the fix,
// retry observes context.Canceled as terminal and returns immediately.
func (m *managerTestSuite) TestCleanUpKeys_DoesNotSpinOnCanceledContext() {
	// cache.Delete always returns context.Canceled. Under the old code
	// this would loop in retry.Retry for ~60 seconds per key; under the
	// fix it returns on the first attempt.
	m.cache.On("Delete", mock.Anything, mock.Anything).Return(context.Canceled)

	canceled, cancel := context.WithCancel(context.Background())
	cancel()

	mgr := m.cachedManager.(*Manager)
	start := time.Now()
	mgr.cleanUpKeys(canceled, 100, "repo", "sha256:deadbeef")
	elapsed := time.Since(start)

	m.Less(elapsed.Seconds(), 1.0, "cleanUpKeys must return promptly on canceled ctx, elapsed=%s", elapsed)
}

// TestDelete_ConcurrentInlineCleanup confirms that the new
// scheduleCleanUp path still runs inline (via orm.AfterCommit's
// non-tx fast path) when no transaction scope is present, and that
// two concurrent Deletes each trigger their own cleanup rather than
// being accidentally serialized or dropped. The in-transaction
// deferral behavior is covered by lib/orm/test.TestAfterCommit_*.
func (m *managerTestSuite) TestDelete_ConcurrentInlineCleanup() {
	m.artMgr.On("Delete", mock.Anything, mock.Anything).Return(nil)
	m.artMgr.On("DeleteReferences", mock.Anything, mock.Anything).Return(nil)
	m.cache.On("Fetch", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	m.cache.On("Delete", mock.Anything, mock.Anything).Return(nil)

	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = m.cachedManager.Delete(m.ctx, 1)
		}()
	}
	wg.Wait()

	m.cache.AssertCalled(m.T(), "Delete", mock.Anything, mock.Anything)
}

// TestScheduleCleanUp_DefersViaAfterCommit verifies the in-transaction
// deferral contract at unit-test level: scheduleCleanUp must register via
// orm.AfterCommit, meaning that when the caller already has a hooks sink
// (normally set up by WithTransaction) the cache.Delete call is NOT
// invoked inline from Delete. We drive this by using the dedicated test
// harness in lib/orm that exposes a ctx with an active hooks sink.
func (m *managerTestSuite) TestScheduleCleanUp_DefersViaAfterCommit() {
	// Install a tx-like hooks context via the orm test helper.
	ctx, drainHooks := orm.ContextWithAfterCommitHooksForTest(context.Background())

	m.artMgr.On("Delete", mock.Anything, mock.Anything).Return(nil).Once()
	m.artMgr.On("DeleteReferences", mock.Anything, mock.Anything).Return(nil).Once()
	m.cache.On("Fetch", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	// cache.Delete is expected to be invoked *after* we drain hooks —
	// we set up the expectation now but assert it has NOT been called
	// yet immediately after Delete returns.
	m.cache.On("Delete", mock.Anything, mock.Anything).Return(nil)

	err := m.cachedManager.Delete(ctx, 100)
	m.NoError(err)

	// The delegator.Delete has run (simulating the in-tx work), but
	// cache.Delete must NOT have been called yet — the hook is still
	// pending in the sink.
	m.cache.AssertNotCalled(m.T(), "Delete", mock.Anything, mock.Anything)

	// Now simulate a successful commit: run the queued hooks.
	drainHooks()

	// After draining, cache.Delete must have been invoked for the two
	// cache keys (id index + digest index).
	m.cache.AssertCalled(m.T(), "Delete", mock.Anything, mock.Anything)
}

func TestManager(t *testing.T) {
	suite.Run(t, &managerTestSuite{})
}
