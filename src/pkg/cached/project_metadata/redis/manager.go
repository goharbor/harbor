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

	libcache "github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/retry"
	"github.com/goharbor/harbor/src/pkg/cached"
	"github.com/goharbor/harbor/src/pkg/project/metadata"
	"github.com/goharbor/harbor/src/pkg/project/metadata/models"
)

var _ CachedManager = &manager{}

// CachedManager is the interface combines raw resource manager and cached manager for better extension.
type CachedManager interface {
	// Manager is the raw resource manager.
	metadata.Manager
	// Manager is the common interface for resource cache.
	cached.Manager
}

// manager is the cached manager implemented by redis.
type manager struct {
	// delegator delegates the raw crud to DAO.
	delegator metadata.Manager
	// client returns the redis cache client.
	client func() libcache.Cache
	// keyBuilder builds cache object key.
	keyBuilder *cached.ObjectKey
	// lifetime is the cache life time.
	lifetime time.Duration
}

// NewManager returns the redis cache manager.
func NewManager(m metadata.Manager) *manager {
	return &manager{
		delegator:  m,
		client:     func() libcache.Cache { return libcache.Default() },
		keyBuilder: cached.NewObjectKey(cached.ResourceTypeProjectMeta),
		lifetime:   time.Duration(config.CacheExpireHours()) * time.Hour,
	}
}

func (m *manager) Add(ctx context.Context, projectID int64, meta map[string]string) error {
	return m.delegator.Add(ctx, projectID, meta)
}

func (m *manager) List(ctx context.Context, name string, value string) ([]*models.ProjectMetadata, error) {
	return m.delegator.List(ctx, name, value)
}

func (m *manager) Get(ctx context.Context, projectID int64, meta ...string) (map[string]string, error) {
	key, err := m.keyBuilder.Format("projectID", projectID, "meta", strings.Join(meta, ","))
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	if err = m.client().Fetch(ctx, key, &result); err == nil {
		return result, nil
	}

	log.Debugf("get project %d metadata from cache error: %v, will query from database.", projectID, err)

	result, err = m.delegator.Get(ctx, projectID, meta...)
	if err != nil {
		return nil, err
	}

	if err = m.client().Save(ctx, key, &result, m.lifetime); err != nil {
		// log error if save to cache failed
		log.Debugf("save project metadata %v to cache error: %v", result, err)
	}

	return result, nil
}

func (m *manager) Delete(ctx context.Context, projectID int64, meta ...string) error {
	// pass on delete operation
	if err := m.delegator.Delete(ctx, projectID, meta...); err != nil {
		return err
	}
	// clean cache
	m.cleanUp(ctx, projectID, meta...)
	return nil
}

func (m *manager) Update(ctx context.Context, projectID int64, meta map[string]string) error {
	if err := m.delegator.Update(ctx, projectID, meta); err != nil {
		return err
	}
	// clean cache
	prefix, err := m.keyBuilder.Format("projectID", projectID)
	if err != nil {
		return err
	}
	// lookup all keys with projectID prefix
	keys, err := m.client().Keys(ctx, prefix)
	if err != nil {
		return err
	}

	for _, key := range keys {
		if err = retry.Retry(func() error { return m.client().Delete(ctx, key) }); err != nil {
			log.Errorf("delete project metadata cache key %s error: %v", key, err)
		}
	}

	return nil
}

// cleanUp cleans up data in cache.
func (m *manager) cleanUp(ctx context.Context, projectID int64, meta ...string) {
	key, err := m.keyBuilder.Format("projectID", projectID, "meta", strings.Join(meta, ","))
	if err != nil {
		log.Errorf("format project metadata key error: %v", err)
	} else {
		// retry to avoid dirty data
		if err = retry.Retry(func() error { return m.client().Delete(ctx, key) }); err != nil {
			log.Errorf("delete project metadata cache key %s error: %v", key, err)
		}
	}
}

func (m *manager) ResourceType(ctx context.Context) string {
	return cached.ResourceTypeProjectMeta
}

func (m *manager) CountCache(ctx context.Context) (int64, error) {
	// prefix is resource type
	keys, err := m.client().Keys(ctx, m.ResourceType(ctx))
	if err != nil {
		return 0, err
	}

	return int64(len(keys)), nil
}

func (m *manager) DeleteCache(ctx context.Context, key string) error {
	return m.client().Delete(ctx, key)
}

func (m *manager) FlushAll(ctx context.Context) error {
	// prefix is resource type
	keys, err := m.client().Keys(ctx, m.ResourceType(ctx))
	if err != nil {
		return err
	}

	var errs errors.Errors
	for _, key := range keys {
		if err = m.client().Delete(ctx, key); err != nil {
			errs = append(errs, err)
		}
	}

	if errs.Len() > 0 {
		return errs
	}

	return nil
}
