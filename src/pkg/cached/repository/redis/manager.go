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

	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/lib/retry"
	"github.com/goharbor/harbor/src/pkg/cached"
	"github.com/goharbor/harbor/src/pkg/repository"
	"github.com/goharbor/harbor/src/pkg/repository/model"
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
	repository.Manager
	// Manager is the common interface for resource cache.
	cached.Manager
}

// Manager is the cached manager implemented by redis.
type Manager struct {
	*cached.BaseManager
	// delegator delegates the raw crud to DAO.
	delegator repository.Manager
	// keyBuilder builds cache object key.
	keyBuilder *cached.ObjectKey
	// lifetime is the cache life time.
	lifetime time.Duration
}

// NewManager returns the redis cache manager.
func NewManager(m repository.Manager) *Manager {
	return &Manager{
		BaseManager: cached.NewBaseManager(cached.ResourceTypeRepository),
		delegator:   m,
		keyBuilder:  cached.NewObjectKey(cached.ResourceTypeRepository),
		lifetime:    time.Duration(config.CacheExpireHours()) * time.Hour,
	}
}

func (m *Manager) Count(ctx context.Context, query *q.Query) (int64, error) {
	return m.delegator.Count(ctx, query)
}

func (m *Manager) List(ctx context.Context, query *q.Query) ([]*model.RepoRecord, error) {
	return m.delegator.List(ctx, query)
}

func (m *Manager) Create(ctx context.Context, repo *model.RepoRecord) (int64, error) {
	return m.delegator.Create(ctx, repo)
}

func (m *Manager) NonEmptyRepos(ctx context.Context) ([]*model.RepoRecord, error) {
	return m.delegator.NonEmptyRepos(ctx)
}

func (m *Manager) Get(ctx context.Context, id int64) (*model.RepoRecord, error) {
	key, err := m.keyBuilder.Format("id", id)
	if err != nil {
		return nil, err
	}

	repo := &model.RepoRecord{}
	if err = m.CacheClient(ctx).Fetch(ctx, key, repo); err == nil {
		return repo, nil
	}

	log.Debugf("get repository %d from cache error: %v, will query from database.", id, err)

	repo, err = m.delegator.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if err = m.CacheClient(ctx).Save(ctx, key, repo, m.lifetime); err != nil {
		// log error if save to cache failed
		log.Debugf("save repository %s to cache error: %v", repo.Name, err)
	}

	return repo, nil
}

func (m *Manager) GetByName(ctx context.Context, name string) (*model.RepoRecord, error) {
	key, err := m.keyBuilder.Format("name", name)
	if err != nil {
		return nil, err
	}

	repo := &model.RepoRecord{}
	if err = m.CacheClient(ctx).Fetch(ctx, key, repo); err == nil {
		return repo, nil
	}

	repo, err = m.delegator.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}

	if err = m.CacheClient(ctx).Save(ctx, key, repo, m.lifetime); err != nil {
		// log error if save to cache failed
		log.Debugf("save repository %s to cache error: %v", repo.Name, err)
	}

	return repo, nil
}

func (m *Manager) Delete(ctx context.Context, id int64) error {
	repo, err := m.Get(ctx, id)
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
	m.scheduleCleanUp(ctx, repo)
	return nil
}

func (m *Manager) Update(ctx context.Context, repo *model.RepoRecord, props ...string) error {
	// pass on update operation
	if err := m.delegator.Update(ctx, repo, props...); err != nil {
		return err
	}
	// Defer cache invalidation until after the enclosing transaction commits,
	// so Redis round-trips never hold the Postgres row locks open. When there
	// is no enclosing transaction, AfterCommit runs the hook synchronously.
	m.scheduleCleanUp(ctx, repo)
	return nil
}

func (m *Manager) AddPullCount(ctx context.Context, id int64, count uint64) error {
	repo, err := m.Get(ctx, id)
	if err != nil {
		return err
	}
	// pass on update operation
	if err = m.delegator.AddPullCount(ctx, id, count); err != nil {
		return err
	}
	// refresh cache
	m.refreshCache(ctx, repo)
	return nil
}

func (m *Manager) Touch(ctx context.Context, id int64) error {
	repo, err := m.Get(ctx, id)
	if err != nil {
		return err
	}
	if err = m.delegator.Touch(ctx, id); err != nil {
		return err
	}
	m.scheduleCleanUp(ctx, repo)
	return nil
}

// scheduleCleanUp registers the cache invalidation for repo to run after the
// enclosing transaction commits. The closure captures a value copy of the
// fields it needs so the hook is independent of the request context and any
// mutations to repo after registration.
func (m *Manager) scheduleCleanUp(ctx context.Context, repo *model.RepoRecord) {
	// Capture by value — the caller's repo pointer may be reused or mutated.
	id := repo.RepositoryID
	name := repo.Name
	orm.AfterCommit(ctx, func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), cleanupCacheTimeout)
		defer cancel()
		m.cleanUpKeys(cleanupCtx, id, name)
	})
}

// cleanUpKeys best-effort deletes the cache entries for a given repository.
// Called after the enclosing transaction has committed with a detached
// context, so it neither holds row locks nor observes a canceled request.
func (m *Manager) cleanUpKeys(ctx context.Context, id int64, name string) {
	// clean index by id
	idIdx, err := m.keyBuilder.Format("id", id)
	if err != nil {
		log.Errorf("format repository id key error: %v", err)
	} else {
		// retry to avoid dirty data; capped retry + ctx-aware so we cannot
		// block indefinitely on Redis hiccups.
		if err = retry.Retry(
			func() error { return m.CacheClient(ctx).Delete(ctx, idIdx) },
			retry.Context(ctx),
			retry.Timeout(cleanupCacheTimeout),
		); err != nil {
			log.Warningf("delete repository cache key %s error: %v", idIdx, err)
		}
	}

	// clean index by name
	nameIdx, err := m.keyBuilder.Format("name", name)
	if err != nil {
		log.Errorf("format repository name key error: %v", err)
	} else {
		if err = retry.Retry(
			func() error { return m.CacheClient(ctx).Delete(ctx, nameIdx) },
			retry.Context(ctx),
			retry.Timeout(cleanupCacheTimeout),
		); err != nil {
			log.Warningf("delete repository cache key %s error: %v", nameIdx, err)
		}
	}
}

// refreshCache refreshes cache.
func (m *Manager) refreshCache(ctx context.Context, repo *model.RepoRecord) {
	// refreshCache used for AddPullCount, because we have a background goroutine to
	// update per repo's pull_count in period time, in that case, we don't want to lose
	// cache every fixed interval, so prefer to use refreshCache instead of cleanUp.
	// no need to consider lock because we only have one goroutine do this work one by one.

	// refreshCache includes 2 steps:
	//   1. cleanUp keys synchronously off a detached context (AddPullCount is
	//      driven by a background goroutine, not inside orm.WithTransaction)
	//   2. re-get
	cleanupCtx, cancel := context.WithTimeout(context.Background(), cleanupCacheTimeout)
	defer cancel()
	m.cleanUpKeys(cleanupCtx, repo.RepositoryID, repo.Name)

	var err error
	// re-get by id
	_, err = m.Get(ctx, repo.RepositoryID)
	if err != nil {
		log.Errorf("refresh cache by repository id %d error: %v", repo.RepositoryID, err)
	}
	// re-get by name
	_, err = m.GetByName(ctx, repo.Name)
	if err != nil {
		log.Errorf("refresh cache by repository name %s error: %v", repo.Name, err)
	}
}
