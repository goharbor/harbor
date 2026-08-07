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
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/retry"
	"github.com/goharbor/harbor/src/pkg/cached"
	"github.com/goharbor/harbor/src/pkg/project/metadata"
	"github.com/goharbor/harbor/src/pkg/project/metadata/models"
)

// cleanupCacheTimeout bounds a single cache-eviction attempt after the
// enclosing transaction has committed. Cache invalidation is best-effort:
// any stale entry expires via the configured TTL, and the next read from
// the database repopulates it. We must not sit spinning in a retry loop
// for the full default timeout.
const cleanupCacheTimeout = 3 * time.Second

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
	// should cleanup cache when add metadata to project — deferred until the
	// enclosing transaction commits so Redis does not hold the tx open.
	m.scheduleCleanUp(ctx, projectID)
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
	// Defer cache invalidation until after the enclosing transaction commits,
	// so Redis round-trips never hold the Postgres row locks open. When there
	// is no enclosing transaction, AfterCommit runs the hook synchronously.
	m.scheduleCleanUp(ctx, projectID, meta...)
	return nil
}

func (m *Manager) Update(ctx context.Context, projectID int64, meta map[string]string) error {
	if err := m.delegator.Update(ctx, projectID, meta); err != nil {
		return err
	}
	// clean cache after commit: the scan + per-key deletes must not hold the
	// transaction open across Redis round-trips.
	prefix, err := m.keyBuilder.Format("projectID", projectID)
	if err != nil {
		return err
	}
	orm.AfterCommit(ctx, func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), cleanupCacheTimeout)
		defer cancel()
		// lookup all keys with projectID prefix
		iter, err := m.CacheClient(cleanupCtx).Scan(cleanupCtx, prefix)
		if err != nil {
			log.Warningf("scan project metadata cache keys with prefix %s error: %v", prefix, err)
			return
		}
		for iter.Next(cleanupCtx) {
			key := iter.Val()
			if err := retry.Retry(
				func() error { return m.CacheClient(cleanupCtx).Delete(cleanupCtx, key) },
				retry.Context(cleanupCtx),
				retry.Timeout(cleanupCacheTimeout),
			); err != nil {
				log.Warningf("delete project metadata cache key %s error: %v", key, err)
			}
		}
	})

	return nil
}

// scheduleCleanUp registers the cache invalidation for the project metadata to
// run after the enclosing transaction commits. The meta slice is copied so the
// hook is independent of the caller's backing array.
func (m *Manager) scheduleCleanUp(ctx context.Context, projectID int64, meta ...string) {
	metaCopy := append([]string(nil), meta...)
	orm.AfterCommit(ctx, func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), cleanupCacheTimeout)
		defer cancel()
		m.cleanUpKeys(cleanupCtx, projectID, metaCopy...)
	})
}

// cleanUpKeys best-effort deletes the project metadata cache entry. Called
// after the enclosing transaction has committed with a detached context, so it
// neither holds row locks nor observes a canceled request.
func (m *Manager) cleanUpKeys(ctx context.Context, projectID int64, meta ...string) {
	key, err := m.keyBuilder.Format("projectID", projectID, "meta", strings.Join(meta, ","))
	if err != nil {
		log.Errorf("format project metadata key error: %v", err)
		return
	}
	// retry to avoid dirty data; capped retry + ctx-aware so we cannot block
	// indefinitely on Redis hiccups.
	if err = retry.Retry(
		func() error { return m.CacheClient(ctx).Delete(ctx, key) },
		retry.Context(ctx),
		retry.Timeout(cleanupCacheTimeout),
	); err != nil {
		log.Warningf("delete project metadata cache key %s error: %v", key, err)
	}
}
