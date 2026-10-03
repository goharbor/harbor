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
	"time"

	commonmodels "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/lib/retry"
	"github.com/goharbor/harbor/src/pkg/cached"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/project/models"
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
	project.Manager
	// Manager is the common interface for resource cache.
	cached.Manager
}

// Manager is the cached manager implemented by redis.
type Manager struct {
	*cached.BaseManager
	// delegator delegates the raw crud to DAO.
	delegator project.Manager
	// keyBuilder builds cache object key.
	keyBuilder *cached.ObjectKey
	// lifetime is the cache life time.
	lifetime time.Duration
}

// NewManager returns the redis cache manager.
func NewManager(m project.Manager) *Manager {
	return &Manager{
		BaseManager: cached.NewBaseManager(cached.ResourceTypeProject),
		delegator:   m,
		keyBuilder:  cached.NewObjectKey(cached.ResourceTypeProject),
		lifetime:    time.Duration(config.CacheExpireHours()) * time.Hour,
	}
}

func (m *Manager) Create(ctx context.Context, project *models.Project) (int64, error) {
	return m.delegator.Create(ctx, project)
}

func (m *Manager) Count(ctx context.Context, query *q.Query) (total int64, err error) {
	return m.delegator.Count(ctx, query)
}

func (m *Manager) List(ctx context.Context, query *q.Query) ([]*models.Project, error) {
	return m.delegator.List(ctx, query)
}

func (m *Manager) ListRoles(ctx context.Context, projectID int64, userID int, groupIDs ...int) ([]int, error) {
	return m.delegator.ListRoles(ctx, projectID, userID, groupIDs...)
}

func (m *Manager) ListAdminRolesOfUser(ctx context.Context, user commonmodels.User) ([]models.Member, error) {
	return m.delegator.ListAdminRolesOfUser(ctx, user)
}

func (m *Manager) Delete(ctx context.Context, id int64) error {
	p, err := m.Get(ctx, id)
	if err != nil {
		return err
	}

	// pass on delete operation
	if err := m.delegator.Delete(ctx, id); err != nil {
		return err
	}
	// Defer cache invalidation until after the enclosing transaction commits,
	// so Redis round-trips never hold the Postgres row locks open. When there
	// is no enclosing transaction, AfterCommit runs the hook synchronously.
	m.scheduleCleanUp(ctx, p)
	return nil
}

func (m *Manager) Get(ctx context.Context, idOrName any) (*models.Project, error) {
	var (
		key string
		err error
	)

	id, name, err := utils.ParseProjectIDOrName(idOrName)
	if err != nil {
		return nil, err
	}

	if id != 0 {
		key, err = m.keyBuilder.Format("id", id)
		if err != nil {
			return nil, err
		}
	}

	if name != "" {
		key, err = m.keyBuilder.Format("name", name)
		if err != nil {
			return nil, err
		}
	}

	p := &models.Project{}
	if err = m.CacheClient(ctx).Fetch(ctx, key, p); err == nil {
		return p, nil
	}

	log.Debugf("get project %v from cache error: %v, will query from database.", idOrName, err)

	p, err = m.delegator.Get(ctx, idOrName)
	if err != nil {
		return nil, err
	}

	if err = m.CacheClient(ctx).Save(ctx, key, p, m.lifetime); err != nil {
		// log error if save to cache failed
		log.Debugf("save project %s to cache error: %v", p.Name, err)
	}

	return p, nil
}

// scheduleCleanUp registers the cache invalidation for p to run after the
// enclosing transaction commits. The closure captures a value copy of the
// fields it needs so the hook is independent of the request context and any
// mutations to p after registration.
func (m *Manager) scheduleCleanUp(ctx context.Context, p *models.Project) {
	// Capture by value — the caller's project pointer may be reused or mutated.
	id := p.ProjectID
	name := p.Name
	orm.AfterCommit(ctx, func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), cleanupCacheTimeout)
		defer cancel()
		m.cleanUpKeys(cleanupCtx, id, name)
	})
}

// cleanUpKeys best-effort deletes the cache entries for a given project.
// Called after the enclosing transaction has committed with a detached
// context, so it neither holds row locks nor observes a canceled request.
func (m *Manager) cleanUpKeys(ctx context.Context, id int64, name string) {
	// clean index by id
	idIdx, err := m.keyBuilder.Format("id", id)
	if err != nil {
		log.Errorf("format project id key error: %v", err)
	} else {
		// retry to avoid dirty data; capped retry + ctx-aware so we cannot
		// block indefinitely on Redis hiccups.
		if err = retry.Retry(
			func() error { return m.CacheClient(ctx).Delete(ctx, idIdx) },
			retry.Context(ctx),
			retry.Timeout(cleanupCacheTimeout),
		); err != nil {
			log.Warningf("delete project cache key %s error: %v", idIdx, err)
		}
	}

	// clean index by name
	nameIdx, err := m.keyBuilder.Format("name", name)
	if err != nil {
		log.Errorf("format project name key error: %v", err)
	} else {
		if err = retry.Retry(
			func() error { return m.CacheClient(ctx).Delete(ctx, nameIdx) },
			retry.Context(ctx),
			retry.Timeout(cleanupCacheTimeout),
		); err != nil {
			log.Warningf("delete project cache key %s error: %v", nameIdx, err)
		}
	}
}
