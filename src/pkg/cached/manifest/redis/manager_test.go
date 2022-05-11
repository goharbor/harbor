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

	"github.com/goharbor/harbor/src/lib/cache"
	testcache "github.com/goharbor/harbor/src/testing/lib/cache"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/stretchr/testify/suite"
)

type managerTestSuite struct {
	suite.Suite
	cachedManager CachedManager
	cache         *testcache.Cache
	ctx           context.Context

	digest      string
	manifestKey string
}

func (m *managerTestSuite) SetupTest() {
	m.cache = &testcache.Cache{}
	m.cachedManager = NewManager()
	m.cachedManager.(*manager).client = func() cache.Cache { return m.cache }
	m.ctx = context.TODO()

	m.digest = "sha256:52f431d980baa76878329b68ddb69cb124c25efa6e206d8b0bd797a828f0528e"
	m.manifestKey = fmt.Sprintf("manifest:digest:%s", m.digest)
}

func (m *managerTestSuite) TestSave() {
	m.cache.On("Save", mock.Anything, m.manifestKey, mock.Anything, mock.Anything).Return(nil).Once()
	err := m.cachedManager.Save(m.ctx, m.digest, []byte{})
	m.NoError(err)
}

func (m *managerTestSuite) TestGet() {
	m.cache.On("Fetch", mock.Anything, m.manifestKey, mock.Anything).Return(nil).Once()
	_, err := m.cachedManager.Get(m.ctx, m.digest)
	m.NoError(err)
}

func (m *managerTestSuite) TestDelete() {
	m.cache.On("Delete", mock.Anything, m.manifestKey).Return(nil).Once()
	err := m.cachedManager.Delete(m.ctx, m.digest)
	m.NoError(err)
}

func (m *managerTestSuite) TestResourceType() {
	t := m.cachedManager.ResourceType(m.ctx)
	m.Equal("manifest", t)
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
