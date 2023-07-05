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
	"strings"
	"time"

	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/retry"
	"github.com/goharbor/harbor/src/pkg/cached"
	"github.com/goharbor/harbor/src/pkg/project/metadata"
	"github.com/goharbor/harbor/src/pkg/project/metadata/models"
)

var _ CachedManager = &Manager{}

// CachedManager is the interface combines raw resource Manager and cached Manager for better extension.
type CachedManager interface {
	// Manager is the raw resource Manager.
	metadata.Manager
	// Manager is the common interface for resource cache.
	cached.Manager
}

// Manager is the cached manager implemented by redis.
type Manager struct {
	*cached.BaseManager
	// delegator delegates the raw crud to DAO.
	delegator metadata.Manager
	// keyBuilder builds cache object key.
	keyBuilder *cached.ObjectKey
	// lifetime is the cache life time.
	lifetime time.Duration
}

// NewManager returns the redis cache manager.
func NewManager(m metadata.Manager) *Manager {
	return &Manager{
		BaseManager: cached.NewBaseManager(cached.ResourceTypeProjectMeta),
		delegator:   m,
		keyBuilder:  cached.NewObjectKey(cached.ResourceTypeProjectMeta),
		lifetime:    time.Duration(config.CacheExpireHours()) * time.Hour,
	}
}

func (m *Manager) Add(ctx context.Context, projectID int64, meta map[string]string) error {
	if err := m.delegator.Add(ctx, projectID, meta); err != nil {
		return err
	}
	// should cleanup cache when add metadata to project
	m.cleanUp(ctx, projectID)
	return nil
}

func (m *Manager) List(ctx context.Context, name string, value string) ([]*models.ProjectMetadata, error) {
	return m.delegator.List(ctx, name, value)
}

func (m *Manager) Get(ctx context.Context, projectID int64, meta ...string) (map[string]string, error) {
	key, err := m.keyBuilder.Format("projectID", projectID, "meta", strings.Join(meta, ","))
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	if err = m.CacheClient(ctx).Fetch(ctx, key, &result); err == nil {
		return result, nil
	}

	log.Debugf("get project %d metadata from cache error: %v, will query from database.", projectID, err)

	result, err = m.delegator.Get(ctx, projectID, meta...)
	if err != nil {
		return nil, err
	}
	// only cache when result has attributes
	if len(result) > 0 {
		if err = m.CacheClient(ctx).Save(ctx, key, &result, m.lifetime); err != nil {
			// log error if save to cache failed
			log.Debugf("save project metadata %v to cache error: %v", result, err)
		}
	}

	return result, nil
}

func (m *Manager) Delete(ctx context.Context, projectID int64, meta ...string) error {
	// pass on delete operation
	if err := m.delegator.Delete(ctx, projectID, meta...); err != nil {
		return err
	}
	// clean cache
	m.cleanUp(ctx, projectID, meta...)
	return nil
}

func (m *Manager) Update(ctx context.Context, projectID int64, meta map[string]string) error {
	if err := m.delegator.Update(ctx, projectID, meta); err != nil {
		return err
	}
	// clean cache
	prefix, err := m.keyBuilder.Format("projectID", projectID)
	if err != nil {
		return err
	}
	// lookup all keys with projectID prefix
	iter, err := m.CacheClient(ctx).Scan(ctx, prefix)
	if err != nil {
		return err
	}

	for iter.Next(ctx) {
		if err = retry.Retry(func() error { return m.CacheClient(ctx).Delete(ctx, iter.Val()) }); err != nil {
			log.Errorf("delete project metadata cache key %s error: %v", iter.Val(), err)
		}
	}

	return nil
}

// cleanUp cleans up data in cache.
func (m *Manager) cleanUp(ctx context.Context, projectID int64, meta ...string) {
	key, err := m.keyBuilder.Format("projectID", projectID, "meta", strings.Join(meta, ","))
	if err != nil {
		log.Errorf("format project metadata key error: %v", err)
	} else {
		// retry to avoid dirty data
		if err = retry.Retry(func() error { return m.CacheClient(ctx).Delete(ctx, key) }); err != nil {
			log.Errorf("delete project metadata cache key %s error: %v", key, err)
		}
	}
}
