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
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
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
	return m.invalidateAfterCommit(ctx, projectID)
}

func (m *Manager) List(ctx context.Context, name string, value string) ([]*models.ProjectMetadata, error) {
	return m.delegator.List(ctx, name, value)
}

func (m *Manager) Get(ctx context.Context, projectID int64, meta ...string) (map[string]string, error) {
	version, err := m.cacheVersion(ctx, projectID)
	if err != nil {
		log.Debugf("get project %d metadata cache version error: %v, will query from database without caching.", projectID, err)
		return m.delegator.Get(ctx, projectID, meta...)
	}
	key, err := m.keyBuilder.Format("projectID", projectID, "meta", strings.Join(meta, ","), "version", version)
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
	// Cache under the version read before the database query. A write that
	// commits while this query is in flight rotates the version, making a late
	// save of the old result unreachable to subsequent reads.
	if len(result) > 0 {
		if err = m.CacheClient(ctx).Save(ctx, key, &result, m.lifetime); err != nil {
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
	return m.invalidateAfterCommit(ctx, projectID)
}

func (m *Manager) Update(ctx context.Context, projectID int64, meta map[string]string) error {
	if err := m.delegator.Update(ctx, projectID, meta); err != nil {
		return err
	}
	return m.invalidateAfterCommit(ctx, projectID)
}

// UpdateWithValidation atomically validates and updates metadata, then invalidates its cache.
func (m *Manager) UpdateWithValidation(ctx context.Context, projectID int64, meta map[string]string, validate metadata.Validator) error {
	if err := m.delegator.UpdateWithValidation(ctx, projectID, meta, validate); err != nil {
		return err
	}
	return m.invalidateAfterCommit(ctx, projectID)
}

func (m *Manager) invalidateAfterCommit(ctx context.Context, projectID int64) error {
	return orm.AfterCommit(ctx, func() error {
		// A committed request context suppresses cache writes. The callback is
		// already outside the transaction, so use a standalone context here.
		return m.cleanUpAll(context.Background(), projectID)
	})
}

func (m *Manager) cacheVersion(ctx context.Context, projectID int64) (string, error) {
	key, err := m.keyBuilder.Format("projectID", projectID, "version", "")
	if err != nil {
		return "", err
	}
	var version string
	if err := m.CacheClient(ctx).Fetch(ctx, key, &version); err == nil && version != "" {
		return version, nil
	} else if err != nil && !errors.Is(err, cache.ErrNotFound) {
		return "", err
	}

	candidate := uuid.NewString()
	if err := m.CacheClient(ctx).Save(ctx, key, candidate, m.lifetime); err != nil {
		return "", err
	}
	if err := m.CacheClient(ctx).Fetch(ctx, key, &version); err != nil {
		return "", err
	}
	if version == "" {
		version = candidate
	}
	return version, nil
}

// cleanUpAll invalidates every metadata view after the database write commits.
// Get caches the full map, individual keys, and combinations of keys separately,
// so adding or deleting one entry must also invalidate views that include it.
func (m *Manager) cleanUpAll(ctx context.Context, projectID int64) error {
	// Rotate the project version before deleting old views. A read already in
	// flight may still save under the old version, but new reads cannot reach it.
	versionKey, err := m.keyBuilder.Format("projectID", projectID, "version", "")
	if err != nil {
		return err
	}
	var cleanupErr error
	if err := retry.Retry(func() error {
		return m.CacheClient(ctx).Save(ctx, versionKey, uuid.NewString(), m.lifetime)
	}); err != nil {
		cleanupErr = errors.Join(cleanupErr, err)
	}

	// Include the metadata delimiter so project 1 does not also match project 11.
	prefix, err := m.keyBuilder.Format("projectID", projectID, "meta", "")
	if err != nil {
		return errors.Join(cleanupErr, err)
	}
	iter, err := m.CacheClient(ctx).Scan(ctx, prefix)
	if err != nil {
		return errors.Join(cleanupErr, err)
	}

	// Attempt every deletion even after a failure, but report incomplete cleanup
	// to the caller instead of treating the write as fully synchronized.
	for iter.Next(ctx) {
		if err = retry.Retry(func() error { return m.CacheClient(ctx).Delete(ctx, iter.Val()) }); err != nil {
			log.Errorf("delete project metadata cache key %s error: %v", iter.Val(), err)
			cleanupErr = errors.Join(cleanupErr, err)
		}
	}

	return errors.Join(cleanupErr, iter.Err())
}
