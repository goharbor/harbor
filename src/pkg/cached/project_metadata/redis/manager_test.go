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
	"maps"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/lib/cache"
	_ "github.com/goharbor/harbor/src/lib/cache/memory"
	"github.com/goharbor/harbor/src/pkg/project/metadata"
	"github.com/goharbor/harbor/src/pkg/project/metadata/models"
	testcache "github.com/goharbor/harbor/src/testing/lib/cache"
	"github.com/goharbor/harbor/src/testing/mock"
	testProjectMeta "github.com/goharbor/harbor/src/testing/pkg/project/metadata"
)

type managerTestSuite struct {
	suite.Suite
	cachedManager  CachedManager
	projectMetaMgr *testProjectMeta.Manager
	cache          *testcache.Cache
	iterator       *testcache.Iterator
	ctx            context.Context
}

func (m *managerTestSuite) SetupTest() {
	m.projectMetaMgr = &testProjectMeta.Manager{}
	m.cache = &testcache.Cache{}
	m.iterator = &testcache.Iterator{}
	m.iterator.On("Err").Return(nil).Maybe()
	m.cachedManager = NewManager(m.projectMetaMgr)
	m.cachedManager.(*Manager).WithCacheClient(m.cache)
	m.ctx = context.TODO()
}

func (m *managerTestSuite) TestAdd() {
	m.cache.On("Save", mock.Anything, "project_metadata:projectID:1:version:", mock.Anything, mock.Anything).Return(nil).Once()
	m.cache.On("Scan", mock.Anything, "project_metadata:projectID:1:meta:").Return(m.iterator, nil).Once()
	m.iterator.On("Next", mock.Anything).Return(true).Once()
	m.iterator.On("Val").Return("project_metadata:projectID:1:meta:")
	m.iterator.On("Next", mock.Anything).Return(false).Once()
	m.cache.On("Delete", mock.Anything, mock.Anything).Return(nil).Once()
	m.projectMetaMgr.On("Add", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	err := m.cachedManager.Add(m.ctx, 1, map[string]string{})
	m.NoError(err)
}

func (m *managerTestSuite) TestList() {
	m.projectMetaMgr.On("List", mock.Anything, mock.Anything, mock.Anything).Return([]*models.ProjectMetadata{}, nil)
	ps, err := m.cachedManager.List(m.ctx, "", "")
	m.NoError(err)
	m.ElementsMatch([]*models.ProjectMetadata{}, ps)
}

func (m *managerTestSuite) TestGet() {
	versionKey := "project_metadata:projectID:100:version:"
	viewKey := "project_metadata:projectID:100:meta::version:v1"
	setVersion := func(args mock.Arguments) {
		*args.Get(2).(*string) = "v1"
	}

	// get from cache directly
	m.cache.On("Fetch", mock.Anything, versionKey, mock.Anything).Run(setVersion).Return(nil).Once()
	m.cache.On("Fetch", mock.Anything, viewKey, mock.Anything).Return(nil).Once()
	_, err := m.cachedManager.Get(m.ctx, 100)
	m.NoError(err, "should get from cache")
	m.projectMetaMgr.AssertNotCalled(m.T(), "Get", mock.Anything, mock.Anything)

	// not found in cache, read from dao
	m.cache.On("Fetch", mock.Anything, versionKey, mock.Anything).Run(setVersion).Return(nil).Once()
	m.cache.On("Fetch", mock.Anything, viewKey, mock.Anything).Return(cache.ErrNotFound).Once()
	m.projectMetaMgr.On("Get", mock.Anything, mock.Anything).Return(map[string]string{}, nil).Once()
	_, err = m.cachedManager.Get(m.ctx, 100)
	m.NoError(err, "should get from projectMetaMgr")
	m.projectMetaMgr.AssertCalled(m.T(), "Get", mock.Anything, mock.Anything)
}

func (m *managerTestSuite) TestDelete() {
	// delete from projectMgr error
	errDelete := errors.New("delete failed")
	m.projectMetaMgr.On("Delete", mock.Anything, mock.Anything).Return(errDelete).Once()
	err := m.cachedManager.Delete(m.ctx, 100)
	m.ErrorIs(err, errDelete, "delete should error")
	m.cache.AssertNotCalled(m.T(), "Delete", mock.Anything, mock.Anything)
	m.cache.AssertNotCalled(m.T(), "Scan", mock.Anything, mock.Anything)

	// delete from projectMgr success
	m.projectMetaMgr.On("Delete", mock.Anything, mock.Anything).Return(nil).Once()
	m.cache.On("Save", mock.Anything, "project_metadata:projectID:100:version:", mock.Anything, mock.Anything).Return(nil).Once()
	m.cache.On("Scan", mock.Anything, "project_metadata:projectID:100:meta:").Return(m.iterator, nil).Once()
	m.iterator.On("Next", mock.Anything).Return(true).Once()
	m.iterator.On("Val").Return("project_metadata:projectID:100:meta:")
	m.iterator.On("Next", mock.Anything).Return(false).Once()
	m.cache.On("Delete", mock.Anything, "project_metadata:projectID:100:meta:").Return(nil).Once()
	err = m.cachedManager.Delete(m.ctx, 100)
	m.NoError(err, "delete should success")
	m.cache.AssertCalled(m.T(), "Delete", mock.Anything, mock.Anything)
}

func (m *managerTestSuite) TestResourceType() {
	t := m.cachedManager.ResourceType(m.ctx)
	m.Equal("project_metadata", t)
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

func TestManager(t *testing.T) {
	suite.Run(t, &managerTestSuite{})
}

// metadataStateManager models persisted metadata independently of the cache.
type metadataStateManager struct {
	metadata.Manager
	values map[string]string
}

func (m *metadataStateManager) Get(_ context.Context, _ int64, names ...string) (map[string]string, error) {
	if len(names) == 0 {
		return maps.Clone(m.values), nil
	}
	result := map[string]string{}
	for _, name := range names {
		if value, ok := m.values[name]; ok {
			result[name] = value
		}
	}
	return result, nil
}

func (m *metadataStateManager) Add(_ context.Context, _ int64, values map[string]string) error {
	maps.Copy(m.values, values)
	return nil
}

func (m *metadataStateManager) Update(ctx context.Context, id int64, values map[string]string) error {
	return m.Add(ctx, id, values)
}

func (m *metadataStateManager) UpdateWithValidation(ctx context.Context, id int64, values map[string]string, validate metadata.Validator) error {
	prospective := maps.Clone(m.values)
	maps.Copy(prospective, values)
	if err := validate(prospective); err != nil {
		return err
	}
	return m.Update(ctx, id, values)
}

func (m *metadataStateManager) Delete(_ context.Context, _ int64, names ...string) error {
	if len(names) == 0 {
		clear(m.values)
	}
	for _, name := range names {
		delete(m.values, name)
	}
	return nil
}

func TestWritesInvalidateMetadataViews(t *testing.T) {
	for _, operation := range []string{"add", "update", "update-with-validation", "delete", "delete-all"} {
		t.Run(operation, func(t *testing.T) {
			ctx := context.Background()
			store := &metadataStateManager{values: map[string]string{"public": "true", "reuse_sys_cve_allowlist": "false"}}
			client, err := cache.New(cache.Memory)
			require.NoError(t, err)
			manager := NewManager(store)
			manager.WithCacheClient(client)
			views := [][]string{nil, {"public"}, {"public", "auto_scan"}, {"auto_scan", "public"}, {"reuse_sys_cve_allowlist"}}
			for _, names := range views {
				_, err := manager.Get(ctx, 1, names...)
				require.NoError(t, err)
			}
			iter, err := client.Scan(ctx, "project_metadata:projectID:1:meta:")
			require.NoError(t, err)
			var cachedViewKeys []string
			for iter.Next(ctx) {
				cachedViewKeys = append(cachedViewKeys, iter.Val())
			}
			require.NoError(t, iter.Err())
			require.Len(t, cachedViewKeys, len(views))
			for _, key := range cachedViewKeys {
				require.Contains(t, key, ":version:")
			}
			otherKey := "project_metadata:projectID:11:meta:"
			require.NoError(t, client.Save(ctx, otherKey, map[string]string{"public": "true"}))
			switch operation {
			case "add":
				err = manager.Add(ctx, 1, map[string]string{"auto_scan": "true"})
			case "update":
				err = manager.Update(ctx, 1, map[string]string{"public": "false"})
			case "update-with-validation":
				err = manager.UpdateWithValidation(ctx, 1, map[string]string{"public": "false"}, func(map[string]string) error { return nil })
			case "delete":
				err = manager.Delete(ctx, 1, "public")
			case "delete-all":
				err = manager.Delete(ctx, 1)
			}
			require.NoError(t, err)
			for _, key := range cachedViewKeys {
				require.False(t, client.Contains(ctx, key), "stale metadata view %s", key)
			}
			for _, names := range views {
				actual, err := manager.Get(ctx, 1, names...)
				require.NoError(t, err)
				expected, err := store.Get(ctx, 1, names...)
				require.NoError(t, err)
				require.Equal(t, expected, actual)
			}
			require.True(t, client.Contains(ctx, otherKey), "another project's metadata must remain cached")
		})
	}
}

func TestWriteReportsTerminalCacheScanFailure(t *testing.T) {
	client := &testcache.Cache{}
	iterator := &testcache.Iterator{}
	scanErr := errors.New("cache iterator failed")
	client.On("Save", mock.Anything, "project_metadata:projectID:1:version:", mock.Anything, mock.Anything).Return(nil).Once()
	client.On("Scan", mock.Anything, "project_metadata:projectID:1:meta:").Return(iterator, nil).Once()
	iterator.On("Next", mock.Anything).Return(false).Once()
	iterator.On("Err").Return(scanErr).Once()

	manager := NewManager(&metadataStateManager{values: map[string]string{}})
	manager.WithCacheClient(client)
	err := manager.Add(context.Background(), 1, map[string]string{"public": "true"})
	require.ErrorIs(t, err, scanErr)
	client.AssertExpectations(t)
}

func TestWriteReportsCacheScanFailure(t *testing.T) {
	for _, operation := range []string{"add", "update", "update-with-validation", "delete"} {
		t.Run(operation, func(t *testing.T) {
			client := &testcache.Cache{}
			scanErr := errors.New("cache scan failed")
			client.On("Save", mock.Anything, "project_metadata:projectID:1:version:", mock.Anything, mock.Anything).Return(nil).Once()
			client.On("Scan", mock.Anything, "project_metadata:projectID:1:meta:").Return(nil, scanErr).Once()
			manager := NewManager(&metadataStateManager{values: map[string]string{}})
			manager.WithCacheClient(client)
			var err error
			switch operation {
			case "add":
				err = manager.Add(context.Background(), 1, map[string]string{"public": "true"})
			case "update":
				err = manager.Update(context.Background(), 1, map[string]string{"public": "false"})
			case "update-with-validation":
				err = manager.UpdateWithValidation(context.Background(), 1, map[string]string{"public": "false"}, func(map[string]string) error { return nil })
			case "delete":
				err = manager.Delete(context.Background(), 1, "public")
			}
			require.ErrorIs(t, err, scanErr)
			client.AssertExpectations(t)
		})
	}
}
